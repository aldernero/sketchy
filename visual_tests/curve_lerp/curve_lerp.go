package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var line gaul.Line
var curve1, curve2, curve3 gaul.Curve

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
	percs := gaul.Linspace(0, 1, int(s.Slider("num_points")), true)
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
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	// Setup lines and curves
	w := s.SketchCanvas.W
	h := s.SketchCanvas.H
	line = gaul.Line{
		P: gaul.Point{X: 20, Y: 20},
		Q: gaul.Point{X: w - 20, Y: 20},
	}
	curve1.Points = []gaul.Point{
		{X: w/2 - 50, Y: h/2 - 60},
		{X: w/2 - 25, Y: h/2 - 60},
		{X: w / 2, Y: h/2 - 60},
		{X: w/2 + 25, Y: h/2 - 60},
		{X: w/2 + 50, Y: h/2 - 60},
	}
	curve2.Points = []gaul.Point{
		{X: w/2 - 25, Y: h/2 - 25},
		{X: w/2 + 25, Y: h/2 - 25},
		{X: w/2 + 25, Y: h/2 + 25},
		{X: w/2 - 25, Y: h/2 + 25},
	}
	curve2.Closed = true
	angles := gaul.Linspace(0, gaul.Tau, 60, false)
	radius := 25.0
	for _, a := range angles {
		p := gaul.Point{
			X: radius*math.Cos(a) + w/2,
			Y: radius*math.Sin(a) + h/2 + 60,
		}
		curve3.Points = append(curve3.Points, p)
	}
	curve3.Closed = true
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
