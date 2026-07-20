package sketchy

import (
	"fmt"
	"strings"

	"github.com/aldernero/debugui"
)

// Recording modes for the Builtins "Rec mode" dropdown.
const (
	recModeManual = iota // record until stopped
	recModeFrames        // record a fixed number of frames
	recModeLoop          // arm: start at Tick % N == 0, capture N frames
)

var (
	recFormatLabels = []string{"WebM (VP9)", "MP4 (H.264)", "WebP (anim)", "FFV1 (MKV)"}
	recModeLabels   = []string{"Manual", "Frames", "Loop"}
)

// recordingOptionsFromPanel builds options from the Builtins Recording rows
// (mode-specific frame counts are applied by the caller).
func (s *Sketch) recordingOptionsFromPanel() RecordingOptions {
	return RecordingOptions{
		Format: RecordingFormat(s.recFormatIdx),
		FPS:    s.recFPS,
		Scale:  exportScaleFactors[s.recScaleIdx],
	}
}

// startRecordingFromPanel starts (or arms, in Loop mode) a recording with
// the current Builtins Recording settings.
func (s *Sketch) startRecordingFromPanel() {
	opts := s.recordingOptionsFromPanel()
	switch s.recModeIdx {
	case recModeFrames:
		opts.NumFrames = int64(s.recFrames)
	case recModeLoop:
		opts.StartModulus = int64(s.recModulus)
		opts.NumFrames = int64(s.recModulus)
	}
	if err := s.StartRecording(opts); err != nil {
		s.recStatus = "Recording error: " + err.Error()
		fmt.Println(s.recStatus)
	}
}

// toggleRecordingHotkey backs both Ctrl+R and the panel button: idle starts
// (or arms) with the panel settings; armed disarms; active stops.
func (s *Sketch) toggleRecordingHotkey() {
	if s.vrec != nil {
		s.StopRecording()
		return
	}
	s.startRecordingFromPanel()
}

// recButtonLabel is stable per state so the debugui widget ID doesn't
// change while the button is being pressed.
func (s *Sketch) recButtonLabel() string {
	if s.vrec == nil {
		if s.recModeIdx == recModeLoop {
			return "Arm Loop (Ctrl+R)"
		}
		return "Start Recording (Ctrl+R)"
	}
	switch s.vrec.state {
	case recArmed:
		return "Disarm (Ctrl+R)"
	case recActive:
		return "Stop Recording (Ctrl+R)"
	default:
		return "Finalizing…"
	}
}

func (s *Sketch) drawBuiltinRecordingRows(ctx *debugui.Context) {
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1}, nil)
	ctx.Text("Rec format")
	ctx.IDScope("recFormat", func() {
		ctx.Dropdown(&s.recFormatIdx, recFormatLabels)
	})
	ctx.Text("Rec FPS")
	ctx.IDScope("recFPS", func() {
		ctx.NumberField(&s.recFPS, 1).On(func() {
			s.recFPS = clampInt(s.recFPS, recordMinFPS, recordMaxFPS)
		})
	})
	ctx.Text("Rec scale")
	ctx.IDScope("recScale", func() {
		ctx.Dropdown(&s.recScaleIdx, exportScaleLabels)
	})
	ctx.Text("Rec mode")
	ctx.IDScope("recMode", func() {
		ctx.Dropdown(&s.recModeIdx, recModeLabels)
	})
	switch s.recModeIdx {
	case recModeFrames:
		ctx.Text("Frames")
		ctx.IDScope("recFrames", func() {
			ctx.NumberField(&s.recFrames, 1).On(func() {
				if s.recFrames < 1 {
					s.recFrames = 1
				}
			})
		})
	case recModeLoop:
		ctx.Text("Loop N")
		ctx.IDScope("recModulus", func() {
			ctx.NumberField(&s.recModulus, 1).On(func() {
				if s.recModulus < 1 {
					s.recModulus = 1
				}
			})
		})
	}

	ctx.SetGridLayout([]int{-1}, nil)
	if n := s.panelRecFrameCount(); n > 0 {
		ctx.Text(fmt.Sprintf("%d frames = %.1fs @ %d fps", n, float64(n)/float64(s.recFPS), s.recFPS))
	}
	ctx.IDScope("recButton", func() {
		ctx.Button(s.recButtonLabel()).On(func() {
			s.toggleRecordingHotkey()
		})
	})
	if s.vrec != nil && s.vrec.state == recActive {
		ctx.Text(fmt.Sprintf("REC %05d (%.1fs)", s.vrec.frames, float64(s.vrec.frames)/float64(s.vrec.opts.FPS)))
	} else if s.vrec != nil && s.vrec.state == recArmed {
		ctx.Text(fmt.Sprintf("Armed: starts at tick %% %d == 0", s.vrec.opts.StartModulus))
	} else if s.recStatus != "" {
		for _, line := range strings.Split(s.recStatus, "\n") {
			ctx.Text(line)
		}
	}
}

// panelRecFrameCount is the frame count the current panel mode implies, or
// 0 for manual mode.
func (s *Sketch) panelRecFrameCount() int {
	switch s.recModeIdx {
	case recModeFrames:
		return s.recFrames
	case recModeLoop:
		return s.recModulus
	default:
		return 0
	}
}
