package main

import (
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

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
	for x := 0.0; x < s.SketchWidth; x += cellSize {
		for y := 0.0; y < s.SketchHeight; y += cellSize {
			noise := s.Rand.Noise2D(x, y)
			hue := sketchy.Map(0, 1, 0, 360, noise)
			cellColor := colorful.Hsl(hue, 0.5, 0.5)
			c.SetColor(cellColor)
			c.DrawRectangle(x, y, cellSize, cellSize)
			c.Fill()
		}
	}
}

func main() {
	s, err := sketchy.NewSketchFromFile("config.json")
	s.Updater = update
	s.Drawer = draw
	if err != nil {
		log.Fatal(err)
	}
	s.Init()
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy Noise Example")
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
