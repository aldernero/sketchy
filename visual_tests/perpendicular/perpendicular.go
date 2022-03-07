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

var line sketchy.Line
var curve1, curve2, curve3 sketchy.Curve

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	c.SetLineCapButt()
	c.SetColor(color.White)
	line.Draw(c)
	curve1.Draw(c)
	curve2.Draw(c)
	curve3.Draw(c)
	c.Stroke()
	c.SetColor(color.CMYK{M: 255})
	percs := sketchy.Linspace(0, 1, int(s.Slider("num_lines")), true)
	for _, p := range percs {
		pb := line.PerpendicularAt(p, 20)
		pb.Draw(c)
		pb = curve1.PerpendicularAt(p, 20)
		pb.Draw(c)
		pb = curve2.PerpendicularAt(p, 20)
		pb.Draw(c)
		pb = curve3.PerpendicularAt(p, 20)
		pb.Draw(c)
	}
	c.Stroke()
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
	// setup lines and curves
	w := s.SketchWidth
	h := s.SketchHeight
	line = sketchy.Line{
		P: sketchy.Point{X: 50, Y: 50},
		Q: sketchy.Point{X: w - 50, Y: 50},
	}
	curve1.Points = []sketchy.Point{
		{X: w/2 - 200, Y: h/2 - 250},
		{X: w/2 - 100, Y: h/2 - 250},
		{X: w / 2, Y: h/2 - 250},
		{X: w/2 + 100, Y: h/2 - 250},
		{X: w/2 + 200, Y: h/2 - 250},
	}
	curve2.Points = []sketchy.Point{
		{X: w/2 - 100, Y: h/2 - 100},
		{X: w/2 + 100, Y: h/2 - 100},
		{X: w/2 + 100, Y: h/2 + 100},
		{X: w/2 - 100, Y: h/2 + 100},
	}
	curve2.Closed = true
	angles := sketchy.Linspace(0, sketchy.Tau, 360, false)
	radius := 100.0
	for _, a := range angles {
		p := sketchy.Point{
			X: radius*math.Cos(a) + w/2,
			Y: radius*math.Sin(a) + h/2 + 250,
		}
		curve3.Points = append(curve3.Points, p)
	}
	curve3.Closed = true
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}