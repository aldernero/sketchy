package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const radius = 300.0

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Not needed for this example
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	N := int(s.Var("N"))
	sides := int(s.Var("sides"))
	rotate := sketchy.Deg2Rad(s.Var("rotate"))
	scale := s.Var("scale")
	c.SetColor(color.CMYK{
		C: 200,
		M: 0,
		Y: 0,
		K: 0,
	})
	for i := 0; i < N; i++ {
		x := s.SketchWidth / 2
		y := s.SketchHeight / 2
		r := math.Pow(scale, float64(i)) * radius
		c.DrawRegularPolygon(sides, x, y, r, float64(i)*rotate)
		c.Stroke()
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
