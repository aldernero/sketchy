package sketchy

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/aldernero/gaul/render"
)

// renderPNGToFile replays the current frame's recording into a fresh raster
// at the given DPI (96 = one raster pixel per sketch pixel) and writes it as
// PNG. It holds saveMutex so it never observes a half-rebuilt frame.
func (s *Sketch) renderPNGToFile(full string, dpi float64) error {
	if s.usesGPUCanvas() {
		// The recorder is empty for a GPU-rendered sketch; replaying it would
		// write a blank image. Capture the frame instead.
		return fmt.Errorf("sketchy: GPU-rendered sketches have no vector recording to save; use EnqueueSavePixels with CaptureGPUImage")
	}
	scale := dpi / DefaultDPI
	if scale <= 0 {
		scale = 1
	}
	w := int(s.SketchWidth*scale + 0.5)
	h := int(s.SketchHeight*scale + 0.5)

	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()
	ras := render.NewRaster(w, h)
	ras.SetScale(scale)
	s.recorder.Replay(ras)
	return ras.SavePNG(full)
}

// renderSVGToFile replays the current frame's recording into an SVG document.
func (s *Sketch) renderSVGToFile(full string) error {
	if s.usesGPUCanvas() {
		return fmt.Errorf("sketchy: GPU-rendered sketches have no vector representation to save as SVG")
	}
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()
	svg := render.NewSVG(s.SketchWidth, s.SketchHeight)
	s.recorder.Replay(svg)
	return svg.Save(full)
}

func writePNG(full string, s *Sketch) error {
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return s.renderPNGToFile(full, s.RasterDPI)
}

func writeSVG(full string, s *Sketch) error {
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return s.renderSVGToFile(full)
}

// writePixelsPNG encodes an already-captured frame (e.g. a shader sketch's
// GPU readback) as PNG.
func writePixelsPNG(full string, img *image.RGBA) error {
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// writeSnapshotPNG writes the current frame for the snapshot dialog,
// dispatching between the CPU recorder replay and the GPU capture.
// Must run on the ebiten thread for GPU-rendered sketches.
func (s *Sketch) writeSnapshotPNG(full string) error {
	if s.usesGPUCanvas() {
		return writePixelsPNG(full, s.CaptureGPUImage())
	}
	return writePNG(full, s)
}
