package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Controls", func() {
		ui.IntSlider("control1", 1, 100, 10, 1)
		ui.FloatSlider("control2", 0, 2, 0.9, 0.01)
	})
	ui.Checkbox("checkbox", false)
	ui.Button("button")
	ui.ColorPicker("accent", "#f3b709")
}

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
}

func main() {
	var prefix string
	var randomSeed int64
	var cpuprofile = flag.String("pprof", "", "Collect CPU profile")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed (0 = auto)")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	var iconImages []image.Image
	fd, err := os.Open("icon.png")
	if err == nil {
		img, err := png.Decode(fd)
		if err == nil {
			iconImages = append(iconImages, img)
		}
	}

	s := sketchy.New(sketchy.Config{
		Title:                  "Sketch",
		SketchWidth:            1080,
		SketchHeight:           1080,
		SketchBackgroundColor:  "#1e1e1e",
		ControlBackgroundColor: "#1e1e1e",
		ControlOutlineColor:    "#ffdb00",
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	setup(s)

	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if iconImages != nil {
		ebiten.SetWindowIcon(iconImages)
	}
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
