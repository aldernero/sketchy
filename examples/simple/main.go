package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Shape", func() {
		ui.FloatSlider("radius", 0, 80, 40, 0.5)
		ui.FloatSlider("thickness", 0, 10, 2, 0.1)
	})
}

func update(s *sketchy.Sketch) {}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	c.SetStrokeWidth(s.GetFloat("Shape", "thickness"))
	c.SetStrokeColor(color.White)
	circle := canvas.Circle(s.GetFloat("Shape", "radius"))
	c.DrawPath(c.Width()/2, c.Height()/2, circle)
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
