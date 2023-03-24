package main

import (
	"flag"
	"github.com/tdewolff/canvas"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	//if s.DidControlsChange {
	//	setup(s)
	//}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	var cpuprofile = flag.String("pprof", "", "Collect CPU profile")
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
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
	setup(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
