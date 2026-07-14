package sketchy

import (
	"image"
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
		s.rasterize(canvas.DPI(DefaultDPI), true)
	}
}

// BenchmarkRasterizeFullFrameImage measures the "every pixel updates" case:
// the Drawer blits one sketch-sized image (as the noise example does), which
// takes the fastImageRasterizer 1:1 path.
func BenchmarkRasterizeFullFrameImage(b *testing.B) {
	s := New(Config{SketchWidth: 1080, SketchHeight: 1080})
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	img := image.NewRGBA(image.Rect(0, 0, 1080, 1080))
	for i := range img.Pix {
		img.Pix[i] = uint8(i)
	}
	ctx := canvas.NewContext(s.SketchCanvas)
	ctx.DrawImage(0, 0, img, canvas.Resolution(s.SketchWidth/s.Width()))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.rasterize(canvas.DPI(DefaultDPI), true)
	}
}
