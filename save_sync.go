package sketchy

import (
	"os"
	"path/filepath"

	"github.com/aldernero/gaul/render"
)

// renderPNGToFile replays the current frame's recording into a fresh raster
// at the given DPI (96 = one raster pixel per sketch pixel) and writes it as
// PNG. It holds saveMutex so it never observes a half-rebuilt frame.
func (s *Sketch) renderPNGToFile(full string, dpi float64) error {
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
