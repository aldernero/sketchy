package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var hasShownMinMax bool = false

func update(s *sketchy.Sketch) {
	s.Rand.SetSeed(int64(s.Var("seed")))
	s.Rand.SetNoiseOctaves(int(s.Var("octaves")))
	s.Rand.SetNoisePersistence(s.Var("persistence"))
	s.Rand.SetNoiseLacunarity(s.Var("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Var("xscale"))
	s.Rand.SetNoiseScaleY(s.Var("yscale"))
	s.Rand.SetNoiseOffsetX(s.Var("xoffset"))
	s.Rand.SetNoiseOffsetY(s.Var("yoffset"))
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	cellSize := s.Var("cellSize")
	minNoise := 100.0
	maxNoise := -100.0
	for x := 0.0; x < s.SketchWidth; x += cellSize {
		for y := 0.0; y < s.SketchHeight; y += cellSize {
			noise := s.Rand.Noise2D(x, y)
			if noise > maxNoise {
				maxNoise = noise
			}
			if noise < minNoise {
				minNoise = noise
			}
			hue := sketchy.Map(-1, 1, 0, 360, noise)
			cellColor := colorful.Hsl(hue, 0.5, 0.5)
			c.SetColor(cellColor)
			c.DrawRectangle(x, y, cellSize, cellSize)
			c.Fill()
		}
	}
	if !hasShownMinMax {
		fmt.Println(minNoise, maxNoise)
		hasShownMinMax = true
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&configFile, "p", "sketch", "Output file prefix")
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
