package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

const radius = 80.0

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Pattern", func() {
		ui.IntSlider("N", 1, 100, 30, 1)
		ui.IntSlider("sides", 1, 100, 7, 1)
		ui.FloatSlider("rotate", -180, 180, -60, 0.5)
		ui.FloatSlider("scale", 0, 1, 0.9, 0.01)
	})
}

func update(s *sketchy.Sketch) {}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	N := s.GetInt("Pattern", "N")
	sides := s.GetInt("Pattern", "sides")
	rotate := s.GetFloat("Pattern", "rotate")
	scale := s.GetFloat("Pattern", "scale")
	c.SetStrokeColor(color.CMYK{
		C: 200,
		M: 0,
		Y: 0,
		K: 0,
	})
	c.SetStrokeWidth(0.5)
	for i := 0; i < N; i++ {
		x := c.Width() / 2
		y := c.Height() / 2
		r := math.Pow(scale, float64(i)) * radius
		c.Push()
		c.Translate(x, y)
		c.Rotate(float64(i) * rotate)
		path := canvas.RegularPolygon(sides, r, true)
		c.DrawPath(0, 0, path)
		c.Pop()
	}
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "Scale & Rotate",
		Prefix:       "scale_rotate",
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
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
