package sketchy

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aldernero/gaul/render"
)

// memSink collects frame sizes (buffers are recycled, so only the first
// frame is copied for pixel checks).
type memSink struct {
	mu     sync.Mutex
	sizes  []int
	first  []byte
	closed bool
}

func (m *memSink) WriteFrame(buf []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sizes = append(m.sizes, len(buf))
	if m.first == nil {
		m.first = slices.Clone(buf)
	}
	return nil
}

func (m *memSink) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *memSink) frameCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sizes)
}

// startMemRecording wires a memSink recording at the given scale.
func startMemRecording(t *testing.T, s *Sketch, opts RecordingOptions) *memSink {
	t.Helper()
	if opts.Scale <= 0 {
		opts.Scale = 1
	}
	if opts.FPS == 0 {
		opts.FPS = recordDefaultFPS
	}
	sink := &memSink{}
	w := int(s.SketchWidth*opts.Scale + 0.5)
	h := int(s.SketchHeight*opts.Scale + 0.5)
	if err := s.startRecordingWithSink(opts, sink, w, h, ""); err != nil {
		t.Fatal(err)
	}
	return sink
}

// tickOnce mimics the Update-loop ordering: updateRecording, then Tick++.
func tickOnce(s *Sketch, dirty bool) {
	if dirty {
		s.dirty = true
	}
	s.updateRecording()
	s.Tick++
}

// waitForFrames waits until the encoder goroutine has consumed n frames.
func waitForFrames(t *testing.T, sink *memSink, n int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for sink.frameCount() < n {
		if time.Now().After(deadline) {
			t.Fatalf("sink has %d frames, want %d", sink.frameCount(), n)
		}
		time.Sleep(time.Millisecond)
	}
}

// waitFinalized polls the finalize handshake like the real Update loop.
func waitFinalized(t *testing.T, s *Sketch) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for s.vrec != nil {
		if time.Now().After(deadline) {
			t.Fatal("recording did not finalize in time")
		}
		s.updateRecording()
		time.Sleep(time.Millisecond)
	}
}

func TestFFmpegArgs(t *testing.T) {
	cases := []struct {
		format  RecordingFormat
		ext     string
		encoder string
	}{
		{RecordWebM, ".webm", "libvpx-vp9"},
		{RecordMP4, ".mp4", "libx264"},
		{RecordWebP, ".webp", "libwebp"},
		{RecordFFV1, ".mkv", "ffv1"},
	}
	for _, tc := range cases {
		out := "out" + tc.ext
		if videoFormatExt[tc.format] != tc.ext {
			t.Errorf("format %d ext = %q, want %q", tc.format, videoFormatExt[tc.format], tc.ext)
		}
		args := ffmpegArgs(tc.format, 640, 480, 30, out, []string{"-crf", "99"})
		for _, want := range []string{"-video_size", "640x480", "-framerate", "30", tc.encoder} {
			if !slices.Contains(args, want) {
				t.Errorf("%s args missing %q: %v", tc.ext, want, args)
			}
		}
		// Output path last, ExtraArgs just before it (last-wins overrides).
		if args[len(args)-1] != out {
			t.Errorf("%s output path not last: %v", tc.ext, args)
		}
		if args[len(args)-3] != "-crf" || args[len(args)-2] != "99" {
			t.Errorf("%s ExtraArgs not immediately before output: %v", tc.ext, args)
		}
	}

	// Odd dimensions get a pad filter only for MP4 (yuv420p needs even).
	odd := ffmpegArgs(RecordMP4, 641, 480, 30, "o.mp4", nil)
	if !slices.Contains(odd, "pad=ceil(iw/2)*2:ceil(ih/2)*2") {
		t.Errorf("odd MP4 args missing pad filter: %v", odd)
	}
	even := ffmpegArgs(RecordMP4, 640, 480, 30, "o.mp4", nil)
	if slices.Contains(even, "-vf") {
		t.Errorf("even MP4 args should not filter: %v", even)
	}
	oddWebM := ffmpegArgs(RecordWebM, 641, 480, 30, "o.webm", nil)
	if slices.Contains(oddWebM, "-vf") {
		t.Errorf("WebM should not pad odd dims: %v", oddWebM)
	}
}

func TestRecordingStateMachine(t *testing.T) {
	t.Run("manual", func(t *testing.T) {
		s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {
			c.SetFillColor(color.White)
			c.DrawCircle(50, 50, 10)
			c.Fill()
		})
		sink := startMemRecording(t, s, RecordingOptions{})
		if !s.IsRecording() {
			t.Fatal("should be recording after manual start")
		}
		for i := 0; i < 7; i++ {
			tickOnce(s, i%2 == 0) // static (non-dirty) ticks still emit frames
		}
		s.StopRecording()
		waitFinalized(t, s)
		if got := sink.frameCount(); got != 7 {
			t.Fatalf("frames = %d, want 7 (duplicates for static ticks)", got)
		}
		if sink.sizes[0] != 100*100*4 {
			t.Fatalf("frame size = %d, want %d", sink.sizes[0], 100*100*4)
		}
		if !sink.closed {
			t.Fatal("sink not closed")
		}
	})

	t.Run("timed stops at NumFrames", func(t *testing.T) {
		s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {})
		sink := startMemRecording(t, s, RecordingOptions{NumFrames: 5})
		for i := 0; i < 20; i++ {
			tickOnce(s, true)
		}
		waitFinalized(t, s)
		if got := sink.frameCount(); got != 5 {
			t.Fatalf("frames = %d, want exactly 5", got)
		}
	})

	t.Run("armed loop starts on modulus", func(t *testing.T) {
		s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {})
		s.Tick = 3
		sink := startMemRecording(t, s, RecordingOptions{StartModulus: 5, NumFrames: 5})
		if s.IsRecording() {
			t.Fatal("armed recording should not be active yet")
		}
		var startTick int64 = -1
		for i := 0; i < 20; i++ {
			before := s.IsRecording()
			tickOnce(s, true)
			if !before && s.vrec != nil && s.vrec.frames > 0 && startTick < 0 {
				startTick = s.Tick - 1 // tick of the first captured frame
			}
		}
		waitFinalized(t, s)
		if startTick != 5 {
			t.Fatalf("first frame captured at tick %d, want 5", startTick)
		}
		if got := sink.frameCount(); got != 5 {
			t.Fatalf("frames = %d, want 5 (ticks 5..9)", got)
		}
	})

	t.Run("disarm before start captures nothing", func(t *testing.T) {
		s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {})
		s.Tick = 1
		sink := startMemRecording(t, s, RecordingOptions{StartModulus: 100, NumFrames: 100})
		tickOnce(s, true)
		s.StopRecording() // disarm
		waitFinalized(t, s)
		if got := sink.frameCount(); got != 0 {
			t.Fatalf("frames = %d, want 0 after disarm", got)
		}
	})

	t.Run("second start rejected", func(t *testing.T) {
		s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {})
		startMemRecording(t, s, RecordingOptions{})
		other := &memSink{}
		if err := s.startRecordingWithSink(RecordingOptions{}, other, 100, 100, ""); err == nil {
			t.Fatal("second concurrent recording should be rejected")
		}
		if !other.closed {
			t.Fatal("rejected sink should be closed")
		}
		s.StopRecording()
		waitFinalized(t, s)
	})
}

func TestRecordingCaptureScale(t *testing.T) {
	s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {
		c.SetFillColor(color.RGBA{255, 0, 0, 255})
		c.SetStrokeColor(nil)
		c.DrawCircle(50, 50, 25)
		c.Fill()
	})
	// Preview mode halves the display raster; recording must be immune.
	s.PreviewMode = true
	sink := startMemRecording(t, s, RecordingOptions{Scale: 2})
	tickOnce(s, true)
	s.StopRecording()
	waitFinalized(t, s)
	if sink.sizes[0] != 200*200*4 {
		t.Fatalf("frame size = %d, want %d (200x200 RGBA)", sink.sizes[0], 200*200*4)
	}
	// Center pixel of the scaled frame is the red circle.
	center := (100*200 + 100) * 4
	if sink.first[center] != 255 || sink.first[center+1] != 0 {
		t.Fatalf("center pixel = %v, want red", sink.first[center:center+4])
	}
}

func TestRecordingAccumulationPath(t *testing.T) {
	s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {
		c.SetFillColor(color.White)
		c.DrawCircle(50, 50, 10)
		c.Fill()
	})
	s.DisableClearBetweenFrames = true
	sink := startMemRecording(t, s, RecordingOptions{})
	tickOnce(s, true)
	waitForFrames(t, sink, 1)
	if sink.sizes[0] != 100*100*4 {
		t.Fatalf("accumulation capture: first frame size %d, want %d", sink.sizes[0], 100*100*4)
	}
	// Changing the display raster size mid-recording aborts with an error.
	s.PreviewMode = true
	tickOnce(s, true)
	waitFinalized(t, s)
	if !strings.Contains(s.recStatus, "failed") {
		t.Fatalf("expected failure status after raster resize, got %q", s.recStatus)
	}
}

func TestRecordingEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	for _, tc := range []struct {
		format RecordingFormat
		name   string
		size   float64
	}{
		{RecordWebM, "out.webm", 100},
		{RecordMP4, "out.mp4", 100},
		{RecordMP4, "odd.mp4", 101}, // exercises the yuv420p pad filter
		{RecordWebP, "out.webp", 100},
		{RecordFFV1, "out.mkv", 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			phase := 0.0
			s := newTestSketch(tc.size, tc.size, func(_ *Sketch, c *render.Context) {
				c.SetFillColor(color.RGBA{0, 200, 255, 255})
				c.DrawCircle(50+20*phase, 50, 10)
				c.Fill()
			})
			out := filepath.Join(dir, tc.name)
			err := s.StartRecording(RecordingOptions{Format: tc.format, FPS: 30, OutPath: out})
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 10; i++ {
				phase = float64(i) / 10
				tickOnce(s, true)
			}
			s.StopRecording()
			waitFinalized(t, s)
			if strings.Contains(s.recStatus, "failed") {
				t.Fatal(s.recStatus)
			}
			info, err := os.Stat(out)
			if err != nil {
				t.Fatal(err)
			}
			if info.Size() == 0 {
				t.Fatal("output file is empty")
			}
			fmt.Printf("%s: %d bytes\n", tc.name, info.Size())
		})
	}
}
