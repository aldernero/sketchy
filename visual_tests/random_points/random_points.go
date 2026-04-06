package main

import (
	"flag"
	"fmt"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Points", func() {
		ui.IntSlider("points", 0, 10000, 1000, 100)
		ui.FloatSlider("threshold", 0, 0.75, 0.5, 0.01)
		ui.IntSlider("octaves", 1, 10, 1, 1)
		ui.FloatSlider("persistence", 0, 2, 0.9, 0.01)
		ui.FloatSlider("lacunarity", 0, 10, 2, 0.1)
		ui.FloatSlider("xscale", 0, 0.1, 0.005, 0.0001)
		ui.FloatSlider("yscale", 0, 0.1, 0.005, 0.0001)
	})
	ui.Checkbox("OpenSimplex", false)
}

func update(s *sketchy.Sketch) {
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(s.GetInt("Points", "octaves"))
	s.Rand.SetNoisePersistence(s.GetFloat("Points", "persistence"))
	s.Rand.SetNoiseLacunarity(s.GetFloat("Points", "lacunarity"))
	s.Rand.SetNoiseScaleX(s.GetFloat("Points", "xscale"))
	s.Rand.SetNoiseScaleY(s.GetFloat("Points", "yscale"))
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	c.SetStrokeColor(color.White)
	c.SetFillColor(color.Transparent)
	var points []gaul.Point
	num := s.GetInt("Points", "points")
	if s.Toggle("OpenSimplex") {
		points = s.Rand.NoisyRandomPoints(num, s.GetFloat("Points", "threshold"), s.CanvasRect())
	} else {
		points = s.Rand.UniformRandomPoints(num, s.CanvasRect())
	}
	if len(points) < num {
		fmt.Println("Points: ", len(points))
	}
	for _, p := range points {
		c.DrawPath(p.X, p.Y, canvas.Circle(0.05))
	}
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:        "Random Points",
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
