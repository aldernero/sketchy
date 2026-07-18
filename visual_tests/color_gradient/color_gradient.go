package main

import (
	"flag"
	"log"

	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var cg1 gaul.SimpleGradient
var cg2 gaul.Gradient

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Gradient", func() {
		ui.FloatSlider("percentage", 0, 1, 0.5, 0.01)
		ui.ColorPicker("start", "cyan")
		ui.ColorPicker("end", "magenta")
	})
}

func setup(s *sketchy.Sketch) {
	cg1 = gaul.SimpleGradient{
		StartColor: s.GetColor("Gradient", "start"),
		EndColor:   s.GetColor("Gradient", "end"),
	}
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidColorPickersChange || s.DidSlidersChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *render.Context) {
	// Drawing code goes here
	W := c.Width()
	H := c.Height()
	_ = H
	left := 0.05 * W
	right := 0.95 * W
	N := 100
	x := gaul.Linspace(left, right, N, true)
	dx := 1.10 * (right - left) / float64(N)
	dy := 19.0
	c.SetLineCap(render.ButtCap)
	c.SetLineJoin(render.MiterJoin)
	for _, i := range x {
		p := gaul.Map(0.05*W, 0.95*W, 0, 1, i)
		c.SetFillColor(cg1.Color(p))
		c.SetStrokeColor(cg1.Color(p))
		c.DrawRectangle(i, 8, dx, dy)
		c.FillStroke()
		c.SetFillColor(cg2.Color(p))
		c.SetStrokeColor(cg2.Color(p))
		c.DrawRectangle(i, 95, dx, dy)
		c.FillStroke()
	}
	p := s.GetFloat("Gradient", "percentage")
	xPos := gaul.Map(0, 1, 0.05*W, 0.95*W, p)
	c.SetFillColor(cg1.Color(p))
	c.SetStrokeColor(cg1.Color(p))
	c.DrawRectangle(xPos-9.5, 46, 19, 19)
	c.FillStroke()
	c.SetFillColor(cg2.Color(p))
	c.SetStrokeColor(cg2.Color(p))
	c.DrawRectangle(xPos-9.5, 132, 19, 19)
	c.FillStroke()
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:        "Color gradient",
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
	cg2 = gaul.NewGradientFromNamed([]string{"blue", "green", "yellow", "red"})
	setup(s)
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
