package main

import (
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	s, err := sketchy.NewSketchFromFile("config.json")
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
