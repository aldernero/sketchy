package main

import (
	"flag"
	"log"

	"github.com/aldernero/gaul/render"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Shape", func() {
		ui.FloatSlider("radius", 0, 300, 150, 1)
	})
}

func update(s *sketchy.Sketch) {}

func draw(s *sketchy.Sketch, c *render.Context) {
	c.DrawCircle(c.Width()/2, c.Height()/2, s.GetFloat("Shape", "radius"))
	c.FillStroke()
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:                 "Simple Example",
		SketchWidth:           800,
		SketchHeight:          800,
		SketchBackgroundColor: "#1e1e1e",
		SketchOutlineColor:    "#ffdb00",
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
