package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

var cg1 sketchy.SimpleGradient
var cg2 sketchy.Gradient

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
	c.SetLineWidth(1)
	for _, i := range x {
		p := sketchy.Map(0.05*s.SketchWidth, 0.95*s.SketchWidth, 0, 1, i)
		c.SetColor(cg1.Color(p))
		c.DrawRectangle(i, 20, dx, dy)
		c.Fill()
		c.SetColor(cg2.Color(p))
		c.DrawRectangle(i, 120, dx, dy)
		c.Fill()
	}
	p := s.Slider("percentage")
	xPos := sketchy.Map(0, 1, 0.05*s.SketchWidth, 0.95*s.SketchWidth, p)
	c.SetColor(color.White)
	c.DrawLine(xPos, 42, xPos, 58)
	c.Stroke()
	c.DrawRectangle(xPos-10, 60, 20, 20)
	c.SetColor(cg1.Color(p))
	c.Fill()
	c.SetColor(color.White)
	c.DrawLine(xPos, 142, xPos, 158)
	c.Stroke()
	c.DrawRectangle(xPos-10, 160, 20, 20)
	c.SetColor(cg2.Color(p))
	c.Fill()
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
	cg1 = sketchy.NewSimpleGradientFromNamed("cyan", "magenta")
	cg2 = sketchy.NewGradientFromNamed([]string{"blue", "green", "yellow", "red"})
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
