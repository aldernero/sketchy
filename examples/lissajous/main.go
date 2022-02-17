package main

import (
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

var lissa sketchy.Lissajous = sketchy.Lissajous{Nx: 3, Ny: 2}

func update(s *sketchy.Sketch) {
	lissa.Nx = int(s.Var("nx"))
	lissa.Ny = int(s.Var("ny"))
	lissa.Px += s.Var("phaseChange")
	lissa.Py = sketchy.Deg2Rad(s.Var("yphase"))
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	radius := s.Var("radius")
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
	s, err := sketchy.NewSketchFromFile("config.json")
	s.Updater = update
	s.Drawer = draw
	if err != nil {
		log.Fatal(err)
	}
	s.Init()
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy Sketch")
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
