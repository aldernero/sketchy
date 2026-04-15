package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Voronoi", func() {
		ui.ColorPicker("start", "#3d5a80")
		ui.ColorPicker("end", "#ee6c4d")
		ui.IntSlider("stops", 1, 50, 5, 1)
		ui.IntSlider("seed points", 20, 200, 20, 1)
	})
}

// Simulation state (package-level for Updater/Drawer).
var (
	sites []gaul.Point
	vels  []gaul.Point
	fills []color.Color
	cells []gaul.Curve
)

func resetSimulation(s *sketchy.Sketch) {
	s.Rand.SetSeed(s.RandomSeed)

	n := s.GetInt("Voronoi", "seed points")
	stops := s.GetInt("Voronoi", "stops")
	w, h := s.Width(), s.Height()
	bounds := gaul.Rect{X: 0, Y: 0, W: w, H: h}

	sg := gaul.SimpleGradient{
		StartColor: s.GetColor("Voronoi", "start"),
		EndColor:   s.GetColor("Voronoi", "end"),
	}
	palette := make([]color.Color, stops)
	for i := 0; i < stops; i++ {
		t := 0.0
		if stops > 1 {
			t = float64(i) / float64(stops-1)
		}
		palette[i] = sg.Color(t)
	}

	sites = make([]gaul.Point, n)
	vels = make([]gaul.Point, n)
	fills = make([]color.Color, n)
	for i := 0; i < n; i++ {
		sites[i] = gaul.Point{
			X: s.Rand.Prng.Float64() * w,
			Y: s.Rand.Prng.Float64() * h,
		}
		ang := gaul.Tau * s.Rand.Prng.Float64()
		spd := 10 * s.Rand.Prng.Float64()
		vels[i] = gaul.Point{X: math.Cos(ang) * spd, Y: math.Sin(ang) * spd}
		idx := int(s.Rand.Prng.Uint64n(uint64(stops)))
		fills[i] = palette[idx]
	}

	var err error
	cells, err = gaul.VoronoiCells(bounds, sites)
	if err != nil {
		log.Printf("VoronoiCells: %v", err)
		cells = nil
	}
}

func stepSimulation(s *sketchy.Sketch) {
	w, h := s.Width(), s.Height()
	bounds := gaul.Rect{X: 0, Y: 0, W: w, H: h}

	dt := 1.0 / 60.0
	if t := ebiten.ActualTPS(); t > 0 {
		dt = 1.0 / t
	}

	for i := range sites {
		p := &sites[i]
		v := &vels[i]
		p.X += v.X * dt
		p.Y += v.Y * dt
		if p.X < 0 {
			p.X = 0
			v.X = -v.X
		} else if p.X > w {
			p.X = w
			v.X = -v.X
		}
		if p.Y < 0 {
			p.Y = 0
			v.Y = -v.Y
		} else if p.Y > h {
			p.Y = h
			v.Y = -v.Y
		}
	}

	var err error
	cells, err = gaul.VoronoiCells(bounds, sites)
	if err != nil {
		log.Printf("VoronoiCells: %v", err)
	}
}

func update(s *sketchy.Sketch) {
	if s.DidControlsChange {
		resetSimulation(s)
	}
	stepSimulation(s)
	s.MarkDirty()
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	for i := range cells {
		if len(cells[i].Points) == 0 {
			continue
		}
		c.SetFillColor(fills[i])
		c.SetStrokeColor(s.DefaultForeground)
		c.SetStrokeWidth(s.DefaultStrokeWidth)
		cells[i].Draw(c)
	}

	c.SetFillColor(color.White)
	c.SetStrokeColor(color.Transparent)
	c.SetStrokeWidth(0)
	for _, p := range sites {
		c.DrawPath(p.X, p.Y, canvas.Circle(2))
		c.Fill()
	}
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:                 "Voronoi (gaul + Sketchy)",
		SketchWidth:           1080,
		SketchHeight:          1080,
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
	resetSimulation(s)
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
