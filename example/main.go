package main

import (
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func (s *sketchy.Sketch) Update() error {
	s.UpdateControls()
	// Custom update code goes here
	return nil
}

func (s *sketchy.Sketch) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	// Custom draw code goes here
}

func main() {
	s, err := sketchy.NewSketchFromFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy Sketch")
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
