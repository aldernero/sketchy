package main

import (
	"flag"
	"image"
	"log"
	"runtime"
	"sync"

	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var tick int64

const noiseImageName = "noise"

var img *image.RGBA

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

// calcNoiseRows fills rows [y0, y1) of img with noise, writing straight into
// the pixel buffer in row-major order.
func calcNoiseRows(s *sketchy.Sketch, mono bool, w, y0, y1 int, wg *sync.WaitGroup) {
	defer wg.Done()
	for y := y0; y < y1; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w*4]
		for x := 0; x < w; x++ {
			noise := s.Rand.Noise3D(float64(x), float64(y), float64(tick))
			var r, g, b uint8
			if !mono {
				hue := gaul.Map(-1, 1, 0, 360, noise)
				r, g, b = colorful.Hsl(hue, 0.5, 0.5).RGB255()
			} else {
				gray := uint8(gaul.Map(-1, 1, 0, 255, noise))
				r, g, b = gray, gray, gray
			}
			i := x * 4
			row[i], row[i+1], row[i+2], row[i+3] = r, g, b, 255
		}
	}
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

	numWorkers := runtime.NumCPU()
	mono := s.Toggle("monochrome")
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go calcNoiseRows(s, mono, W, i*H/numWorkers, (i+1)*H/numWorkers, &wg)
	}
	wg.Wait()
	s.RegisterImage(noiseImageName, img)
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

func draw(s *sketchy.Sketch, c *render.Context) {
	s.DrawNamedImage(c, noiseImageName)
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
