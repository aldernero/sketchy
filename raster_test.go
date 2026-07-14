package sketchy

import (
	"image/color"
	"testing"

	"github.com/tdewolff/canvas"
)

func rasterTestSketch() *Sketch {
	s := New(Config{SketchWidth: 100, SketchHeight: 100})
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	return s
}

func fillCircleAt(c *canvas.Canvas, x, y float64, col color.RGBA) {
	ctx := canvas.NewContext(c)
	ctx.SetFillColor(col)
	ctx.DrawPath(x, y, canvas.Circle(5*MmPerPx))
}

// TestRasterizeAccumulates covers the DisableClearBetweenFrames contract:
// with clearBuf false the previous raster stays under the new frame, and
// with clearBuf true it is wiped.
func TestRasterizeAccumulates(t *testing.T) {
	s := rasterTestSketch()
	red := color.RGBA{R: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}

	fillCircleAt(s.SketchCanvas, 25*MmPerPx, 50*MmPerPx, red)
	s.rasterize(canvas.DPI(DefaultDPI), true)

	s.SketchCanvas.Reset()
	fillCircleAt(s.SketchCanvas, 75*MmPerPx, 50*MmPerPx, blue)
	img := s.rasterize(canvas.DPI(DefaultDPI), false)

	if r, _, _, a := img.At(25, 50).RGBA(); r == 0 || a == 0 {
		t.Error("accumulated frame lost the previous red circle")
	}
	if _, _, b, _ := img.At(75, 50).RGBA(); b == 0 {
		t.Error("accumulated frame missing the new blue circle")
	}

	s.SketchCanvas.Reset()
	fillCircleAt(s.SketchCanvas, 75*MmPerPx, 50*MmPerPx, blue)
	img = s.rasterize(canvas.DPI(DefaultDPI), true)
	if _, _, _, a := img.At(25, 50).RGBA(); a != 0 {
		t.Error("cleared frame still contains the previous red circle")
	}
}

// TestRasterizePreviewSize checks that PreviewMode's half DPI halves the
// raster (Draw scales it back up to the logical sketch size).
func TestRasterizePreviewSize(t *testing.T) {
	s := rasterTestSketch()
	img := s.rasterize(canvas.DPI(DefaultPreviewDPI), true)
	if img.Bounds().Dx() != 50 || img.Bounds().Dy() != 50 {
		t.Errorf("preview raster = %v, want 50x50", img.Bounds())
	}
	dx, dy := s.displaySizeF()
	if dx != 100 || dy != 100 {
		t.Errorf("displaySizeF = %v,%v, want 100,100 regardless of raster DPI", dx, dy)
	}
}
