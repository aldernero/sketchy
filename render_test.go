package sketchy

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aldernero/gaul/render"
)

// newTestSketch builds a sketch with just enough state for the render
// pipeline, without Init (which opens a database, spawns the save worker,
// and touches ebiten).
func newTestSketch(w, h float64, drawer SketchDrawer) *Sketch {
	s := New(Config{SketchWidth: w, SketchHeight: h})
	s.DefaultBackground = color.Black
	s.DefaultForeground = color.White
	s.DefaultStrokeWidth = 1
	s.RasterDPI = DefaultDPI
	s.recorder = render.NewRecorder(w, h)
	s.needToClear = true
	s.Drawer = drawer
	return s
}

func TestRenderFrameAndSaves(t *testing.T) {
	s := newTestSketch(200, 200, func(_ *Sketch, c *render.Context) {
		c.SetFillColor(color.RGBA{255, 0, 0, 255})
		c.SetStrokeColor(nil)
		c.DrawCircle(100, 100, 50)
		c.Fill()
	})

	img := s.renderFrame()
	if got := img.Bounds().Dx(); got != 200 {
		t.Fatalf("raster width = %d, want 200", got)
	}
	// Center is red, background is black.
	if c := img.RGBAAt(100, 100); c.R != 255 || c.G != 0 {
		t.Fatalf("center pixel = %v, want red", c)
	}
	if c := img.RGBAAt(10, 10); c.R != 0 || c.A != 255 {
		t.Fatalf("corner pixel = %v, want opaque black", c)
	}

	dir := t.TempDir()

	// PNG save at 2x DPI doubles the output resolution.
	pngPath := filepath.Join(dir, "out.png")
	if err := s.renderPNGToFile(pngPath, 2*DefaultDPI); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := png.DecodeConfig(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 400 || cfg.Height != 400 {
		t.Fatalf("saved PNG is %dx%d, want 400x400", cfg.Width, cfg.Height)
	}

	// SVG save contains the recorded shapes with logical coordinates.
	svgPath := filepath.Join(dir, "out.svg")
	if err := s.renderSVGToFile(svgPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	if !strings.Contains(doc, `viewBox="0 0 200 200"`) {
		t.Fatalf("SVG missing expected viewBox: %s", doc[:min(len(doc), 200)])
	}
	if !strings.Contains(doc, `fill="#ff0000"`) {
		t.Fatal("SVG missing red fill")
	}
}

func TestRenderFramePreviewModeHalvesRaster(t *testing.T) {
	s := newTestSketch(200, 200, func(_ *Sketch, c *render.Context) {
		c.SetFillColor(color.White)
		c.DrawRectangle(0, 0, 200, 200)
		c.Fill()
	})
	s.PreviewMode = true
	img := s.renderFrame()
	if got := img.Bounds().Dx(); got != 100 {
		t.Fatalf("preview raster width = %d, want 100", got)
	}
	// Logical coordinates still cover the full surface.
	if c := img.RGBAAt(99, 99); c.R != 255 {
		t.Fatalf("far corner = %v, want white", c)
	}
}

func TestRenderFrameAccumulation(t *testing.T) {
	tick := 0
	s := newTestSketch(100, 100, func(_ *Sketch, c *render.Context) {
		c.SetFillColor(color.White)
		x := 25.0
		if tick > 0 {
			x = 75.0
		}
		c.DrawCircle(x, 50, 10)
		c.Fill()
	})
	s.DisableClearBetweenFrames = true

	img := s.renderFrame()
	if img.RGBAAt(25, 50).R != 255 {
		t.Fatal("first frame circle missing")
	}
	tick = 1
	img = s.renderFrame()
	// Display accumulates both circles...
	if img.RGBAAt(25, 50).R != 255 || img.RGBAAt(75, 50).R != 255 {
		t.Fatal("accumulated display should show both circles")
	}
	// ...but a save renders only the current frame.
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "frame.png")
	if err := s.renderPNGToFile(pngPath, DefaultDPI); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	r1, _, _, _ := saved.At(25, 50).RGBA()
	r2, _, _, _ := saved.At(75, 50).RGBA()
	if r1 != 0 {
		t.Fatal("saved frame should not include the previous frame's circle")
	}
	if r2 == 0 {
		t.Fatal("saved frame missing current frame's circle")
	}
}
