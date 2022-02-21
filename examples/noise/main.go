package main

import (
	"flag"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var tick int64

func update(s *sketchy.Sketch) {
	s.Rand.SetSeed(int64(s.Var("seed")))
	s.Rand.SetNoiseOctaves(int(s.Var("octaves")))
	s.Rand.SetNoisePersistence(s.Var("persistence"))
	s.Rand.SetNoiseLacunarity(s.Var("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Var("xscale"))
	s.Rand.SetNoiseScaleY(s.Var("yscale"))
	s.Rand.SetNoiseOffsetX(s.Var("xoffset"))
	s.Rand.SetNoiseOffsetY(s.Var("yoffset"))
	s.Rand.SetNoiseScaleZ(0.005)
	s.Rand.SetNoiseOffsetZ(float64(tick))
	tick++
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	cellSize := s.Var("cellSize")
	c.SetLineWidth(0)
	for x := 0.0; x < s.SketchWidth; x += cellSize {
		for y := 0.0; y < s.SketchHeight; y += cellSize {
			noise := s.Rand.Noise3D(x, y, 0)
			hue := sketchy.Map(0, 1, 0, 360, noise)
			cellColor := colorful.Hsl(hue, 0.5, 0.5)
			c.SetColor(cellColor)
			c.DrawRectangle(x, y, cellSize, cellSize)
			c.Fill()
		}
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
