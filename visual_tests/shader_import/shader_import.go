package main

import (
	"flag"
	"log"
	"os"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	var seed int64
	flag.Int64Var(&seed, "s", 0, "Random number generator seed (0 = auto)")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "Shader Import",
		SketchWidth:  800,
		SketchHeight: 800,
		ShaderPath:   "fragment.kage",
	})
	s.RandomSeed = seed
	// Headless self-test hook (vshot): SHADER_IMPORT_PNG=<rel.png> saves a
	// PNG at native resolution.
	if pngOut := os.Getenv("SHADER_IMPORT_PNG"); pngOut != "" {
		s.Updater = func(s *sketchy.Sketch) {
			if s.Tick == 10 {
				s.EnqueueSavePixels(pngOut, s.CaptureShaderImage(), false)
			}
		}
	}
	s.Init()

	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
