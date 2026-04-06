package sketchy

import (
	"os"
	"path/filepath"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

func writePNG(full string, s *Sketch) error {
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()
	return renderers.Write(full, s.SketchCanvas, canvas.DPI(s.RasterDPI))
}

func writeSVG(full string, s *Sketch) error {
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()
	return renderers.Write(full, s.SketchCanvas)
}
