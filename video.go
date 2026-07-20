package sketchy

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// RecordingFormat selects the container/codec for video recording. All
// formats are encoded by a user-installed ffmpeg binary; sketchy pipes raw
// RGBA frames to it over stdin.
type RecordingFormat int

const (
	// RecordWebM is VP9 in WebM: royalty-free with excellent
	// quality-per-bit. The default.
	RecordWebM RecordingFormat = iota
	// RecordMP4 is H.264 in MP4: maximum playback compatibility.
	RecordMP4
	// RecordWebP is an infinitely-looping animated WebP, the format GitHub
	// renders inline in Markdown.
	RecordWebP
	// RecordFFV1 is lossless FFV1 in Matroska: a perfect archival master
	// for later transcoding.
	RecordFFV1
)

// RecordingOptions configures a recording started with
// [Sketch.StartRecording]. One video frame is captured per sketch tick.
type RecordingOptions struct {
	Format RecordingFormat
	// FPS is the playback frame rate written into the file (1-240,
	// default 60). It does not resample: one tick is always one frame.
	FPS int
	// Scale is the record raster scale, like the Builtins Export scale
	// (1 = one raster pixel per sketch pixel). Default 1.
	Scale float64
	// NumFrames stops the recording after exactly this many frames;
	// 0 means record until StopRecording.
	NumFrames int64
	// StartModulus arms the recording instead of starting it: capture
	// begins at the next tick where Tick % StartModulus == 0. Combined
	// with NumFrames == StartModulus this captures a perfect loop.
	StartModulus int64
	// ExtraArgs are appended after the per-format defaults and before the
	// output path, so ffmpeg's last-option-wins lets them override
	// anything (e.g. "-crf", "30", or "-lossless", "1" for VP9).
	ExtraArgs []string
	// OutPath overrides the output file path. Default:
	// saves/video/<Prefix>_<timestamp>.<ext> under the working directory.
	OutPath string
}

const (
	recordMinFPS     = 1
	recordMaxFPS     = 240
	recordDefaultFPS = 60
	// recordFrameChanCap bounds in-flight frames; sends block when the
	// encoder falls behind, slowing the live loop instead of dropping
	// frames (tick-driven animation keeps the output frame-perfect).
	recordFrameChanCap = 4
)

// videoEncoderArgs holds the quality-first per-format encoder defaults.
// RecordingOptions.ExtraArgs can override any of them.
var videoEncoderArgs = map[RecordingFormat][]string{
	RecordWebM: {"-c:v", "libvpx-vp9", "-pix_fmt", "yuv420p", "-crf", "24", "-b:v", "0",
		"-deadline", "good", "-cpu-used", "2", "-row-mt", "1"},
	RecordMP4: {"-c:v", "libx264", "-pix_fmt", "yuv420p", "-crf", "17", "-preset", "slow",
		"-movflags", "+faststart"},
	RecordWebP: {"-c:v", "libwebp", "-quality", "90", "-compression_level", "6", "-loop", "0"},
	RecordFFV1: {"-c:v", "ffv1", "-level", "3", "-coder", "1", "-context", "1", "-g", "1",
		"-slices", "4", "-slicecrc", "1", "-pix_fmt", "bgra"},
}

var videoFormatExt = map[RecordingFormat]string{
	RecordWebM: ".webm",
	RecordMP4:  ".mp4",
	RecordWebP: ".webp",
	RecordFFV1: ".mkv",
}

// ffmpegArgs builds the full ffmpeg argument list for one recording: raw
// RGBA frames on stdin, encoded to out.
func ffmpegArgs(f RecordingFormat, w, h, fps int, out string, extra []string) []string {
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "rawvideo", "-pixel_format", "rgba",
		"-video_size", fmt.Sprintf("%dx%d", w, h),
		"-framerate", strconv.Itoa(fps),
		"-i", "-",
	}
	args = append(args, videoEncoderArgs[f]...)
	// yuv420p subsampling needs even dimensions; pad one black row/column
	// rather than cropping content.
	if f == RecordMP4 && (w%2 == 1 || h%2 == 1) {
		args = append(args, "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2")
	}
	args = append(args, extra...)
	return append(args, out)
}

// frameSink consumes raw RGBA frames. The ffmpeg pipe implements it; tests
// substitute an in-memory sink.
type frameSink interface {
	WriteFrame(buf []byte) error
	Close() error
}

type ffmpegSink struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr bytes.Buffer
}

func newFFmpegSink(ffmpegPath string, args []string) (*ffmpegSink, error) {
	sink := &ffmpegSink{cmd: exec.Command(ffmpegPath, args...)}
	stdin, err := sink.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	sink.stdin = stdin
	sink.cmd.Stderr = &sink.stderr
	if err := sink.cmd.Start(); err != nil {
		return nil, err
	}
	return sink, nil
}

func (fs *ffmpegSink) WriteFrame(buf []byte) error {
	_, err := fs.stdin.Write(buf)
	return err
}

func (fs *ffmpegSink) Close() error {
	cerr := fs.stdin.Close()
	if werr := fs.cmd.Wait(); werr != nil {
		if msg := strings.TrimSpace(fs.stderr.String()); msg != "" {
			return fmt.Errorf("ffmpeg: %v: %s", werr, tailLines(msg, 10))
		}
		return fmt.Errorf("ffmpeg: %v", werr)
	}
	return cerr
}

func tailLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

type recState int

const (
	recArmed recState = iota
	recActive
	recFinalizing
)

// videoRecorder is the live state of one recording. It exists only between
// StartRecording and the finalize handshake; s.vrec == nil means idle.
// State transitions happen on the ebiten thread; the encoder goroutine only
// stores encErr and sends the final result on doneCh.
type videoRecorder struct {
	state recState
	opts  RecordingOptions
	// w, h are fixed at start; ffmpeg's frame size cannot change mid-stream.
	w, h       int
	frames     int64
	frameCh    chan []byte
	doneCh     chan error
	encErr     atomic.Value  // error: encoder failed; Update auto-stops next tick
	free       chan []byte   // recycled frame buffers
	raster     *image.RGBA   // reused replay target for the record-scale path
	gpuTarget  *ebiten.Image // reused record-size shader render target
	sink       frameSink
	outPath    string // full path, for the status message
	captureErr error  // capture-side abort reason, reported at finish
}

func (r *videoRecorder) getBuf() []byte {
	n := r.w * r.h * 4
	select {
	case b := <-r.free:
		if cap(b) >= n {
			return b[:n]
		}
	default:
	}
	return make([]byte, n)
}

// StartRecording begins (or arms, when StartModulus > 0) a video recording.
// It fails if a recording is already in progress or ffmpeg is not installed.
func (s *Sketch) StartRecording(opts RecordingOptions) error {
	if _, ok := videoFormatExt[opts.Format]; !ok {
		return fmt.Errorf("unknown recording format %d", opts.Format)
	}
	if opts.FPS == 0 {
		opts.FPS = recordDefaultFPS
	}
	opts.FPS = clampInt(opts.FPS, recordMinFPS, recordMaxFPS)
	if opts.Scale <= 0 {
		opts.Scale = 1
	}
	if opts.StartModulus < 0 || opts.NumFrames < 0 {
		return fmt.Errorf("StartModulus and NumFrames must be >= 0")
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH — install ffmpeg to record video")
	}

	full := opts.OutPath
	if full == "" {
		name := s.Prefix + "_" + gaul.GetTimestampString() + videoFormatExt[opts.Format]
		full = filepath.Join(s.workDir, "saves", "video", name)
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}

	w := int(s.SketchWidth*opts.Scale + 0.5)
	h := int(s.SketchHeight*opts.Scale + 0.5)
	sink, err := newFFmpegSink(ffmpegPath, ffmpegArgs(opts.Format, w, h, opts.FPS, full, opts.ExtraArgs))
	if err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}
	return s.startRecordingWithSink(opts, sink, w, h, full)
}

// startRecordingWithSink is the sink-injectable core of StartRecording
// (tests use an in-memory sink).
func (s *Sketch) startRecordingWithSink(opts RecordingOptions, sink frameSink, w, h int, outPath string) error {
	if s.vrec != nil {
		_ = sink.Close()
		return fmt.Errorf("a recording is already in progress")
	}
	r := &videoRecorder{
		state:   recActive,
		opts:    opts,
		w:       w,
		h:       h,
		frameCh: make(chan []byte, recordFrameChanCap),
		doneCh:  make(chan error, 1),
		free:    make(chan []byte, recordFrameChanCap+2),
		sink:    sink,
		outPath: outPath,
	}
	if opts.StartModulus > 0 {
		r.state = recArmed
	}
	go func() {
		var err error
		// Keep draining after an error so the capture side never blocks.
		for buf := range r.frameCh {
			if err == nil {
				if werr := r.sink.WriteFrame(buf); werr != nil {
					err = werr
					r.encErr.Store(werr)
				}
			}
			select {
			case r.free <- buf:
			default:
			}
		}
		if cerr := r.sink.Close(); err == nil {
			err = cerr
		}
		r.doneCh <- err
	}()
	s.vrec = r
	s.recStatus = ""
	return nil
}

// ArmLoopRecording arms a perfect-loop recording with the current panel
// format/fps/scale: capture starts at the next tick where Tick % n == 0 and
// stops after exactly n frames, so the last frame precedes a repeat of the
// first.
func (s *Sketch) ArmLoopRecording(n int64) error {
	if n <= 0 {
		return fmt.Errorf("loop length must be positive, got %d", n)
	}
	opts := s.recordingOptionsFromPanel()
	opts.StartModulus = n
	opts.NumFrames = n
	return s.StartRecording(opts)
}

// StopRecording ends (or disarms) the current recording. The file is
// finalized asynchronously; the result appears in the Builtins panel status
// row and on stdout. No-op when nothing is recording.
func (s *Sketch) StopRecording() {
	r := s.vrec
	if r == nil || r.state == recFinalizing {
		return
	}
	r.state = recFinalizing
	close(r.frameCh)
}

// FinishRecording stops the current recording (if still running) and
// blocks until the file is fully written, or the timeout elapses. The
// non-blocking path — StopRecording plus the per-tick finalize poll — is
// right for interactive use; this is for scripted sketches that must not
// exit while ffmpeg is still flushing. Safe to call from Updater.
func (s *Sketch) FinishRecording(timeout time.Duration) error {
	r := s.vrec
	if r == nil {
		return nil
	}
	if r.state != recFinalizing {
		s.StopRecording()
	}
	select {
	case err := <-r.doneCh:
		s.finishRecording(err)
		return err
	case <-time.After(timeout):
		return fmt.Errorf("recording did not finalize within %v", timeout)
	}
}

// IsRecording reports whether frames are being captured (not armed, not
// finalizing).
func (s *Sketch) IsRecording() bool {
	return s.vrec != nil && s.vrec.state == recActive
}

// RecordingFrameCount returns the number of frames captured so far, or 0
// when idle.
func (s *Sketch) RecordingFrameCount() int64 {
	if s.vrec == nil {
		return 0
	}
	return s.vrec.frames
}

// abortRecording stops with a capture-side error (e.g. raster size changed
// mid-recording on the accumulation path).
func (s *Sketch) abortRecording(err error) {
	if s.vrec == nil {
		return
	}
	s.vrec.captureErr = err
	s.StopRecording()
}

// updateRecording runs once per tick from Update, before Tick increments:
// it finalizes finished recordings, fires armed ones, and captures the
// current frame.
func (s *Sketch) updateRecording() {
	r := s.vrec
	if r == nil {
		return
	}
	if r.state == recFinalizing {
		select {
		case err := <-r.doneCh:
			s.finishRecording(err)
		default:
		}
		return
	}
	if err, ok := r.encErr.Load().(error); ok && err != nil {
		s.StopRecording()
		return
	}
	if r.state == recArmed {
		if s.Tick%r.opts.StartModulus != 0 {
			return
		}
		r.state = recActive
	}

	// Render at most once per tick: Drawer may consume s.Rand, so rendering
	// here and again in Draw would change the animation vs. non-recording
	// runs. rasterUploadPending tells Draw to upload without re-rendering
	// (shader mode renders straight into the offscreen, nothing to upload).
	if s.dirty {
		if s.IsShaderSketch() {
			s.renderShaderFrame(s.offscreen)
		} else {
			s.renderFrame()
			s.rasterUploadPending = true
		}
		s.dirty = false
	}
	buf := s.captureVideoFrame()
	if buf == nil {
		return
	}
	r.frameCh <- buf
	r.frames++
	if n := r.opts.NumFrames; n > 0 && r.frames >= n {
		s.StopRecording()
	}
}

// captureVideoFrame copies the current frame's pixels into a recycled
// buffer at the recording's fixed size. Non-dirty ticks reproduce the
// previous frame (duplicate, not skip) so ticks and frames stay 1:1.
func (s *Sketch) captureVideoFrame() []byte {
	r := s.vrec
	if s.IsShaderSketch() {
		// GPU readback (premultiplied RGBA — the layout the ffmpeg pipe
		// expects). At scale 1 the display offscreen already holds the
		// frame; other scales re-render into a reused record-size target.
		buf := r.getBuf()
		if s.offscreen != nil && s.offscreen.Bounds().Dx() == r.w && s.offscreen.Bounds().Dy() == r.h {
			s.offscreen.ReadPixels(buf)
			return buf
		}
		if r.gpuTarget == nil || r.gpuTarget.Bounds().Dx() != r.w || r.gpuTarget.Bounds().Dy() != r.h {
			r.gpuTarget = ebiten.NewImage(r.w, r.h)
		}
		s.renderShaderFrame(r.gpuTarget)
		r.gpuTarget.ReadPixels(buf)
		return buf
	}
	if s.DisableClearBetweenFrames {
		// Accumulated strokes only exist in the display raster; the
		// recording holds just the current frame, so replay would lose
		// them. Copy the display buffer as-is.
		rb := s.rasterBuf
		if rb == nil || rb.Bounds().Dx() != r.w || rb.Bounds().Dy() != r.h {
			s.abortRecording(fmt.Errorf(
				"display raster size changed mid-recording (Preview mode or Export scale toggled?); with DisableClearBetweenFrames the recording captures the display raster, which must stay %dx%d", r.w, r.h))
			return nil
		}
		buf := r.getBuf()
		copy(buf, rb.Pix)
		return buf
	}

	// Fast path: the display raster already matches the record size and is
	// full quality (not preview).
	if rb := s.rasterBuf; rb != nil && !s.PreviewMode &&
		rb.Bounds().Dx() == r.w && rb.Bounds().Dy() == r.h {
		buf := r.getBuf()
		copy(buf, rb.Pix)
		return buf
	}

	// Replay the frame recording at record scale, independent of the
	// display DPI — Preview mode never leaks into the video.
	if r.raster == nil {
		r.raster = image.NewRGBA(image.Rect(0, 0, r.w, r.h))
	}
	s.saveMutex.Lock()
	clear(r.raster.Pix)
	ras := render.NewRasterFromImage(r.raster)
	ras.SetScale(r.opts.Scale)
	s.recorder.Replay(ras)
	s.saveMutex.Unlock()
	buf := r.getBuf()
	copy(buf, r.raster.Pix)
	return buf
}

// finishRecording completes the finalize handshake on the ebiten thread.
func (s *Sketch) finishRecording(encErr error) {
	r := s.vrec
	s.vrec = nil
	err := r.captureErr
	if err == nil {
		err = encErr
	}
	switch {
	case err != nil:
		s.recStatus = fmt.Sprintf("Recording failed: %v", err)
	case r.frames == 0:
		// Disarmed (or stopped) before any frame: drop the empty file.
		_ = os.Remove(r.outPath)
		s.recStatus = "Recording canceled (no frames captured)"
	default:
		s.recStatus = fmt.Sprintf("Saved %s (%d frames)", r.outPath, r.frames)
	}
	fmt.Println(s.recStatus)
}
