package main

import (
	"flag"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

var lissa sketchy.Lissajous = sketchy.Lissajous{Nx: 3, Ny: 2}

func update(s *sketchy.Sketch) {
	lissa.Nx = int(s.Slider("nx"))
	lissa.Ny = int(s.Slider("ny"))
	lissa.Px += s.Slider("phaseChange")
	lissa.Py = sketchy.Deg2Rad(s.Slider("yphase"))
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	radius := s.Slider("radius")
	origin := sketchy.Point{X: s.SketchWidth / 2, Y: s.SketchHeight / 2}
	curve := sketchy.GenLissajous(lissa, 1000, origin, radius)
	c.SetColor(color.CMYK{C: 200})
	c.SetLineWidth(3)
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
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
