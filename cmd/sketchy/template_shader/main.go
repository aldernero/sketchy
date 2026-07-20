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
)

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
}

func update(s *sketchy.Sketch) {
	// Per-tick logic goes here (optional for shader sketches — controls
	// declared in fragment.kage drive the shader without any code here)
}

func main() {
	var prefix string
	var randomSeed int64
	var paletteDBPath string
	var cpuprofile = flag.String("pprof", "", "Collect CPU profile")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed (0 = auto)")
	flag.StringVar(&paletteDBPath, "palettedb", "", "Path to palettedb database (default ~/.config/palettedb/palettedb.db)")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
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
		Title:                  "Shader Sketch",
		SketchWidth:            1080,
		SketchHeight:           1080,
		ControlBackgroundColor: "#1e1e1e",
		ControlOutlineColor:    "#ffdb00",
		// The Kage fragment shader that renders this sketch. Its //sketchy:
		// directives create the control panel; edits live-reload while running.
		ShaderPath: "fragment.kage",
	})
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.PaletteDBPath = paletteDBPath
	s.Updater = update
	// s.ExtraUniforms = func(s *sketchy.Sketch) map[string]any {
	// 	return map[string]any{"MyVec2": []float32{1, 2}} // computed uniforms
	// }
	s.Init()
	setup(s)

	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	// Fixed 60 TPS keeps the Time uniform in real seconds and makes video
	// recordings deterministic.
	ebiten.SetTPS(60)
	if iconImages != nil {
		ebiten.SetWindowIcon(iconImages)
	}
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
