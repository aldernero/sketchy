package sketchy

import (
	"testing"

	"github.com/tdewolff/canvas"
)

// benchSketch builds a Sketch with enough state for the raster pipeline
// (no ebiten graphics context needed).
func benchSketch(w, h float64) *Sketch {
	s := New(Config{SketchWidth: w, SketchHeight: h})
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	ctx := canvas.NewContext(s.SketchCanvas)
	ctx.SetFillColor(canvas.Mediumseagreen)
	ctx.DrawPath(0, 0, canvas.Rectangle(ctx.Width(), ctx.Height()))
	for i := 0; i < 40; i++ {
		ctx.SetFillColor(canvas.Orangered)
		ctx.DrawPath(float64(i*7)*MmPerPx, float64(i*5)*MmPerPx, canvas.Circle(20*MmPerPx))
	}
	return s
}

// BenchmarkRasterize measures the per-dirty-frame CPU raster cost for a
// 1080x1080 sketch (canvas render into the reused buffer, ready for
// ebiten WritePixels).
func BenchmarkRasterize(b *testing.B) {
	s := benchSketch(1080, 1080)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.rasterize(canvas.DPI(DefaultDPI))
	}
}
