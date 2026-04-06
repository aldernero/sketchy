package main

import (
	"flag"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Colors", func() {
		ui.ColorPicker("fill", "#3d5a80")
		ui.ColorPicker("stroke", "#ee6c4d")
	})
	ui.Folder("Shape", func() {
		ui.Dropdown("figure", []string{"Circle", "Square", "Triangle"}, 0)
		ui.FloatSlider("size", 20, 220, 100, 1)
		ui.FloatSlider("stroke width", 0.5, 12, 2.5, 0.5)
	})
}

func update(*sketchy.Sketch) {}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	fill := s.GetColor("Colors", "fill")
	stroke := s.GetColor("Colors", "stroke")
	c.SetFillColor(fill)
	c.SetStrokeColor(stroke)
	c.SetStrokeWidth(s.GetFloat("Shape", "stroke width"))

	cx := c.Width() / 2
	cy := c.Height() / 2
	r := s.GetFloat("Shape", "size")

	switch s.GetDropdownIndex("Shape", "figure") {
	case 0:
		c.DrawPath(cx, cy, canvas.Circle(r))
	case 1:
		c.DrawPath(cx-r, cy-r, canvas.Rectangle(2*r, 2*r))
	case 2:
		c.DrawPath(cx, cy, canvas.RegularPolygon(3, r, true))
	}
	c.FillStroke()
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:                 "Color pickers & dropdown",
		SketchWidth:           720,
		SketchHeight:          720,
		SketchBackgroundColor: "#1e1e1e",
		SketchOutlineColor:    "#1e1e1e",
		ControlOutlineColor:   "#ffdb00",
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
