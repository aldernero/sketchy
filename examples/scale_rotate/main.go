package main

import (
	"flag"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const radius = 80.0

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Not needed for this example
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	N := int(s.Slider("N"))
	sides := int(s.Slider("sides"))
	rotate := s.Slider("rotate")
	scale := s.Slider("scale")
	c.SetStrokeColor(color.CMYK{
		C: 200,
		M: 0,
		Y: 0,
		K: 0,
	})
	c.SetStrokeWidth(0.5)
	for i := 0; i < N; i++ {
		x := c.Width() / 2
		y := c.Height() / 2
		r := math.Pow(scale, float64(i)) * radius
		c.Push()
		c.Translate(x, y)
		c.Rotate(float64(i) * rotate)
		path := canvas.RegularPolygon(sides, r, true)
		c.DrawPath(0, 0, path)
		c.Pop()
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
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
