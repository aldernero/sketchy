package main

import (
	"flag"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

var cg sketchy.SimpleGradient

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	left := 0.05 * s.SketchWidth
	right := 0.95 * s.SketchWidth
	N := 100
	x := sketchy.Linspace(left, right, N, true)
	dx := 1.10 * (right - left) / float64(N)
	dy := 20.0
	c.SetLineCapButt()
	c.SetLineWidth(0)
	for _, i := range x {
		p := sketchy.Map(0.05*s.SketchWidth, 0.95*s.SketchWidth, 0, 1, i)
		c.SetColor(cg.Color(p))
		c.DrawRectangle(i, 20, dx, dy)
		c.Fill()
	}
	c.SetColor(cg.Color(s.Slider("percentage")))
	c.DrawCircle(s.SketchWidth/2, s.SketchHeight/2, 100)
	c.Fill()
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
	cg = sketchy.NewSimpleGradientFromNamed("cyan", "magenta")
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
