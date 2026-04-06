package main

import (
	"flag"
	"log"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

var lissa = gaul.Lissajous{Nx: 3, Ny: 2}

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Curve", func() {
		ui.IntSlider("nx", 1, 100, 3, 1)
		ui.IntSlider("ny", 1, 100, 2, 1)
		ui.FloatSlider("radius", 0, 100, 50, 0.5)
		ui.IntSlider("yphase", 0, 360, 180, 1)
		ui.FloatSlider("phaseChange", -1, 1, 0.01, 0.01)
	})
}

func update(s *sketchy.Sketch) {
	lissa.Nx = s.GetInt("Curve", "nx")
	lissa.Ny = s.GetInt("Curve", "ny")
	pc := s.GetFloat("Curve", "phaseChange")
	lissa.Px += pc
	lissa.Py = gaul.Deg2Rad(float64(s.GetInt("Curve", "yphase")))
	// Px advances every Update; Drawer only runs when dirty. Without a per-frame
	// MarkDirty while animating, redraws (clicks, controls) would show a large
	// phase jump vs the last cached frame.
	if pc != 0 {
		s.MarkDirty()
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	radius := s.GetFloat("Curve", "radius")
	origin := gaul.Point{X: c.Width() / 2, Y: c.Height() / 2}
	curve := gaul.GenLissajous(lissa, 1000, origin, radius)
	c.SetStrokeColor(s.DefaultForeground)
	c.SetStrokeWidth(s.DefaultStrokeWidth)
	c.MoveTo(curve.Points[0].X, curve.Points[0].Y)
	for _, p := range curve.Points {
		c.LineTo(p.X, p.Y)
	}
	c.LineTo(curve.Points[0].X, curve.Points[0].Y)
	c.Stroke()
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "Lissajous Curve Example",
		SketchWidth:  1080,
		SketchHeight: 768,
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	s.ShowFPS = true
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
