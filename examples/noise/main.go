package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var tick int64

func update(s *sketchy.Sketch) {
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(int(s.Slider("octaves")))
	s.Rand.SetNoisePersistence(s.Slider("persistence"))
	s.Rand.SetNoiseLacunarity(s.Slider("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Slider("xscale"))
	s.Rand.SetNoiseScaleY(s.Slider("yscale"))
	s.Rand.SetNoiseOffsetX(s.Slider("xoffset"))
	s.Rand.SetNoiseOffsetY(s.Slider("yoffset"))
	s.Rand.SetNoiseScaleZ(0.005)
	s.Rand.SetNoiseOffsetZ(3 * float64(tick))
	if s.Toggle("animate") {
		tick++
	}
	if s.Toggle("reset") {
		tick = 0
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	cellSize := s.Slider("cellSize")
	//c.SetStrokeWidth(1)
	for x := 0.0; x < s.SketchWidth; x += cellSize {
		for y := 0.0; y < s.SketchHeight; y += cellSize {
			noise := s.Rand.Noise3D(x, y, 0)
			if !s.Toggle("monochrome") {
				hue := gaul.Map(0, 1, 0, 360, noise)
				cellColor := colorful.Hsl(hue, 0.5, 0.5)
				c.SetFillColor(cellColor)
				c.SetStrokeColor(cellColor)
			} else {
				gray := gaul.Map(0, 1, 0, 255, noise)
				cellColor := color.Gray{Y: uint8(gray)}
				c.SetFillColor(cellColor)
				c.SetStrokeColor(cellColor)
			}
			c.DrawPath(x, y, canvas.Rectangle(cellSize, cellSize))
		}
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
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
