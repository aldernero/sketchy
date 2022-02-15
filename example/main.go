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
	lissa.Px += s.Controls[2].Val
	lissa.Py = sketchy.Deg2Rad(s.Controls[1].Val)
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	origin := sketchy.Point{X: s.SketchWidth / 2, Y: s.SketchHeight / 2}
	curve := sketchy.GenLissajous(lissa, 1000, origin, s.Controls[0].Val)
	c.SetColor(color.Black)
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
	ctx := gg.NewContext(int(s.ControlWidth), int(s.SketchHeight))
	s.PlaceControls(s.ControlWidth, s.SketchHeight, ctx)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy Sketch")
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
