package main

import (
	"flag"
	gaul "github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var lissa = gaul.Lissajous{Nx: 3, Ny: 2}

func update(s *sketchy.Sketch) {
	lissa.Nx = int(s.Slider("nx"))
	lissa.Ny = int(s.Slider("ny"))
	lissa.Px += s.Slider("phaseChange")
	lissa.Py = gaul.Deg2Rad(s.Slider("yphase"))
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	radius := s.Slider("radius")
	origin := gaul.Point{X: c.Width() / 2, Y: c.Height() / 2}
	curve := gaul.GenLissajous(lissa, 1000, origin, radius)
	c.SetStrokeColor(color.CMYK{C: 200})
	c.SetStrokeWidth(1)
	c.MoveTo(curve.Points[0].X, curve.Points[0].Y)
	for _, p := range curve.Points {
		c.LineTo(p.X, p.Y)
	}
	c.LineTo(curve.Points[0].X, curve.Points[0].Y)
	c.Stroke()
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
	s.ShowFPS = true
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
