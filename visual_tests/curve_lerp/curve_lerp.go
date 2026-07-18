package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var line gaul.Line
var curve1, curve2, curve3 gaul.Curve

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.IntSlider("num_points", 1, 100, 10, 1)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *render.Context) {
	// Drawing code goes here
	c.SetFillColor(color.Transparent)
	line.Draw(c)
	for _, p := range curve1.Points {
		c.DrawCircle(p.X, p.Y, 38)
		c.FillStroke()
	}
	curve1.Draw(c)
	for _, p := range curve2.Points {
		c.DrawCircle(p.X, p.Y, 38)
		c.FillStroke()
	}
	curve2.Draw(c)
	//for _, p := range curve3.Points {
	//	c.DrawCircle(p.X, p.Y, 4)
	//}
	curve3.Draw(c)
	c.Stroke()
	percs := gaul.Linspace(0, 1, s.GetInt("", "num_points"), true)
	for _, p := range percs {
		point := line.Lerp(p)
		c.DrawCircle(point.X, point.Y, 11)
		c.FillStroke()
		point = curve1.Lerp(p)
		c.DrawCircle(point.X, point.Y, 11)
		c.FillStroke()
		point = curve2.Lerp(p)
		c.DrawCircle(point.X, point.Y, 11)
		c.FillStroke()
		point = curve3.Lerp(p)
		c.DrawCircle(point.X, point.Y, 11)
		c.FillStroke()
	}
	c.Fill()
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:        "Curve lerp",
		SketchWidth:  800,
		SketchHeight: 800,
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	// Setup lines and curves
	w := s.Width()
	h := s.Height()
	line = gaul.Line{
		P: gaul.Point{X: 75, Y: 75},
		Q: gaul.Point{X: w - 75, Y: 75},
	}
	curve1.Points = []gaul.Point{
		{X: w/2 - 189, Y: h/2 - 227},
		{X: w/2 - 94, Y: h/2 - 227},
		{X: w / 2, Y: h/2 - 227},
		{X: w/2 + 94, Y: h/2 - 227},
		{X: w/2 + 189, Y: h/2 - 227},
	}
	curve2.Points = []gaul.Point{
		{X: w/2 - 94, Y: h/2 - 94},
		{X: w/2 + 94, Y: h/2 - 94},
		{X: w/2 + 94, Y: h/2 + 94},
		{X: w/2 - 94, Y: h/2 + 94},
	}
	curve2.Closed = true
	angles := gaul.Linspace(0, gaul.Tau, 60, false)
	radius := 94.0
	for _, a := range angles {
		p := gaul.Point{
			X: radius*math.Cos(a) + w/2,
			Y: radius*math.Sin(a) + h/2 + 227,
		}
		curve3.Points = append(curve3.Points, p)
	}
	curve3.Closed = true
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
