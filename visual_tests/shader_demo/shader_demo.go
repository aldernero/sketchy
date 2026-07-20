package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	var seed int64
	flag.Int64Var(&seed, "s", 0, "Random number generator seed (0 = auto)")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "Shader Demo",
		SketchWidth:  800,
		SketchHeight: 800,
		ShaderPath:   "fragment.kage",
	})
	s.RandomSeed = seed
	// Headless self-test hooks (vshot): SHADER_DEMO_RECORD=<path.webm>
	// records a 60-frame clip; SHADER_DEMO_PNG=<rel.png> saves a 2x PNG.
	recOut := os.Getenv("SHADER_DEMO_RECORD")
	pngOut := os.Getenv("SHADER_DEMO_PNG")
	if recOut != "" || pngOut != "" {
		s.Updater = func(s *sketchy.Sketch) {
			switch s.Tick {
			case 5:
				if recOut == "" {
					return
				}
				err := s.StartRecording(sketchy.RecordingOptions{
					Format:    sketchy.RecordWebM,
					NumFrames: 60,
					OutPath:   recOut,
				})
				if err != nil {
					log.Fatal(err)
				}
			case 10:
				if pngOut != "" {
					s.EnqueueSavePixels(pngOut, s.CaptureShaderImage(2), false)
				}
			case 70: // frames end at tick 64; block until ffmpeg flushes
				if recOut != "" {
					if err := s.FinishRecording(60 * time.Second); err != nil {
						log.Fatal(err)
					}
				}
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
