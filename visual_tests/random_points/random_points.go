package main

import (
	"flag"
	"fmt"
	gaul "github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func update(s *sketchy.Sketch) {
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(int(s.Slider("octaves")))
	s.Rand.SetNoisePersistence(s.Slider("persistence"))
	s.Rand.SetNoiseLacunarity(s.Slider("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Slider("xscale"))
	s.Rand.SetNoiseScaleY(s.Slider("yscale"))
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	c.SetStrokeColor(color.White)
	c.SetFillColor(color.Transparent)
	var points []gaul.Point
	num := int(s.Slider("points"))
	if s.Toggle("OpenSimplex") {
		points = s.Rand.NoisyRandomPoints(num, s.Slider("threshold"), s.CanvasRect())
	} else {
		points = s.Rand.UniformRandomPoints(num, s.CanvasRect())
	}
	if len(points) < num {
		fmt.Println("Points: ", len(points))
	}
	for _, p := range points {
		c.DrawPath(p.X, p.Y, canvas.Circle(0.05))
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
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
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
