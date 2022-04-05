package main

import (
	"flag"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var qt *sketchy.QuadTree

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.Toggle("Clear") {
		qt.Clear()
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if s.PointInSketchArea(float64(x), float64(y)) {
			p := s.CanvasCoords(float64(x), float64(y))
			qt.Insert(p)
		}
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(color.White)
	c.SetFillColor(color.Transparent)
	c.SetStrokeCapper(canvas.ButtCap)
	c.SetStrokeWidth(s.Slider("Line Thickness"))
	if s.Toggle("Show Points") {
		qt.DrawWithPoints(s.Slider("Point Size"), c)
	} else {
		qt.Draw(c)
	}
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
	qt = sketchy.NewQuadTree(sketchy.Rect{
		X: 0,
		Y: 0,
		W: s.Width(),
		H: s.Height(),
	})
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
