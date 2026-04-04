package main

import (
	"flag"
	"image"
	"log"
	"runtime"
	"sync"

	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var tick int64

type pixel struct {
	x, y       int
	r, g, b, a uint8
}

type result []pixel

var img *image.RGBA
var pxPerMm float64

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Noise", func() {
		ui.IntSlider("octaves", 1, 10, 2, 1)
		ui.FloatSlider("persistence", 0, 2, 0.23, 0.01)
		ui.FloatSlider("lacunarity", 0, 10, 0.3, 0.1)
		ui.FloatSlider("xscale", 0, 0.1, 0.0056, 0.0001)
		ui.FloatSlider("yscale", 0, 0.1, 0.0035, 0.0001)
		ui.IntSlider("xoffset", -1000, 1000, 0, 1)
		ui.IntSlider("yoffset", -1000, 1000, 0, 1)
	})
	ui.Checkbox("animate", false)
	ui.Checkbox("monochrome", false)
	ui.Button("reset")
}

func calcNoise(s *sketchy.Sketch, mono bool, cs []pixel, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	res := make(result, len(cs))
	for i, cell := range cs {
		noise := s.Rand.Noise3D(float64(cell.x), float64(cell.y), float64(tick))
		if !mono {
			hue := gaul.Map(-1, 1, 0, 360, noise)
			c := colorful.Hsl(hue, 0.5, 0.5)
			cell.r, cell.g, cell.b = c.RGB255()
			cell.a = 255
		} else {
			gray := gaul.Map(-1, 1, 0, 255, noise)
			grayVal := uint8(gray)
			cell.r, cell.g, cell.b, cell.a = grayVal, grayVal, grayVal, 255
		}
		res[i] = cell
	}
	results <- res
}

func setup(s *sketchy.Sketch) {
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(s.GetInt("Noise", "octaves"))
	s.Rand.SetNoisePersistence(s.GetFloat("Noise", "persistence"))
	s.Rand.SetNoiseLacunarity(s.GetFloat("Noise", "lacunarity"))
	s.Rand.SetNoiseScaleX(s.GetFloat("Noise", "xscale"))
	s.Rand.SetNoiseScaleY(s.GetFloat("Noise", "yscale"))
	s.Rand.SetNoiseOffsetX(float64(s.GetInt("Noise", "xoffset")))
	s.Rand.SetNoiseOffsetY(float64(s.GetInt("Noise", "yoffset")))
	s.Rand.SetNoiseScaleZ(0.005)

	W := int(s.SketchWidth)
	H := int(s.SketchHeight)
	if img == nil || img.Bounds().Dx() != W || img.Bounds().Dy() != H {
		img = image.NewRGBA(image.Rect(0, 0, W, H))
	}

	pixels := make([]pixel, W*H)
	for i := 0; i < W; i++ {
		for j := 0; j < H; j++ {
			pixels[i*H+j] = pixel{x: i, y: j}
		}
	}

	numWorkers := runtime.NumCPU()
	results := make(chan result, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	mono := s.Toggle("monochrome")
	for i := 0; i < numWorkers; i++ {
		cs := pixels[i*len(pixels)/numWorkers : (i+1)*len(pixels)/numWorkers]
		go calcNoise(s, mono, cs, results, &wg)
	}
	wg.Wait()

	stride := img.Stride
	for i := 0; i < numWorkers; i++ {
		r := <-results
		for _, p := range r {
			idx := p.y*stride + p.x*4
			img.Pix[idx] = p.r
			img.Pix[idx+1] = p.g
			img.Pix[idx+2] = p.b
			img.Pix[idx+3] = p.a
		}
	}
	close(results)
}

func update(s *sketchy.Sketch) {
	if s.DidControlsChange {
		if s.Toggle("reset") {
			tick = 0
		}
		setup(s)
		s.MarkDirty()
		return
	}
	if s.Toggle("animate") {
		setup(s)
		tick++
		s.MarkDirty()
		return
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	c.DrawImage(0, 0, img, canvas.Resolution(pxPerMm))
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "OpenSimplex Noise Example",
		SketchWidth:  1080,
		SketchHeight: 768,
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	pxPerMm = s.SketchWidth / s.Width()
	setup(s)
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
