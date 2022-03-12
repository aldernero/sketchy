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

var line sketchy.Line
var curve1, curve2, curve3 sketchy.Curve

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetFillColor(color.Transparent)
	c.SetStrokeColor(color.White)
	line.Draw(c)
	for _, p := range curve1.Points {
		c.DrawPath(p.X, p.Y, canvas.Circle(10))
	}
	curve1.Draw(c)
	for _, p := range curve2.Points {
		c.DrawPath(p.X, p.Y, canvas.Circle(10))
	}
	curve2.Draw(c)
	//for _, p := range curve3.Points {
	//	c.DrawCircle(p.X, p.Y, 4)
	//}
	curve3.Draw(c)
	c.Stroke()
	c.SetStrokeColor(color.CMYK{M: 255})
	percs := sketchy.Linspace(0, 1, int(s.Slider("num_points")), true)
	for _, p := range percs {
		point := line.Lerp(p)
		c.DrawPath(point.X, point.Y, canvas.Circle(3))
		point = curve1.Lerp(p)
		c.DrawPath(point.X, point.Y, canvas.Circle(3))
		point = curve2.Lerp(p)
		c.DrawPath(point.X, point.Y, canvas.Circle(3))
		point = curve3.Lerp(p)
		c.DrawPath(point.X, point.Y, canvas.Circle(3))
	}
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
	// Setup lines and curves
	w := s.SketchCanvas.W
	h := s.SketchCanvas.H
	line = sketchy.Line{
		P: sketchy.Point{X: 20, Y: 20},
		Q: sketchy.Point{X: w - 20, Y: 20},
	}
	curve1.Points = []sketchy.Point{
		{X: w/2 - 50, Y: h/2 - 60},
		{X: w/2 - 25, Y: h/2 - 60},
		{X: w / 2, Y: h/2 - 60},
		{X: w/2 + 25, Y: h/2 - 60},
		{X: w/2 + 50, Y: h/2 - 60},
	}
	curve2.Points = []sketchy.Point{
		{X: w/2 - 25, Y: h/2 - 25},
		{X: w/2 + 25, Y: h/2 - 25},
		{X: w/2 + 25, Y: h/2 + 25},
		{X: w/2 - 25, Y: h/2 + 25},
	}
	curve2.Closed = true
	angles := sketchy.Linspace(0, sketchy.Tau, 60, false)
	radius := 25.0
	for _, a := range angles {
		p := sketchy.Point{
			X: radius*math.Cos(a) + w/2,
			Y: radius*math.Sin(a) + h/2 + 60,
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
