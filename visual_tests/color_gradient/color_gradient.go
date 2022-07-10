package main

import (
	"flag"
	gaul "github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var cg1 sketchy.SimpleGradient
var cg2 sketchy.Gradient

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	W := c.Width()
	H := c.Height()
	left := 0.05 * W
	right := 0.95 * W
	N := 100
	x := gaul.Linspace(left, right, N, true)
	dx := 1.10 * (right - left) / float64(N)
	dy := 5.0
	c.SetStrokeWidth(1)
	c.SetStrokeCapper(canvas.ButtCap)
	c.SetStrokeJoiner(canvas.MiterJoin)
	for _, i := range x {
		p := gaul.Map(0.05*W, 0.95*W, 0, 1, i)
		c.SetFillColor(cg1.Color(p))
		c.SetStrokeColor(cg1.Color(p))
		c.DrawPath(i, H-7, canvas.Rectangle(dx, dy))
		c.SetFillColor(cg2.Color(p))
		c.SetStrokeColor(cg2.Color(p))
		c.DrawPath(i, H-30, canvas.Rectangle(dx, dy))
	}
	p := s.Slider("percentage")
	xPos := gaul.Map(0, 1, 0.05*W, 0.95*W, p)
	c.SetFillColor(cg1.Color(p))
	c.SetStrokeColor(cg1.Color(p))
	c.DrawPath(xPos-2.5, H-17, canvas.Rectangle(5, 5))
	c.SetFillColor(cg2.Color(p))
	c.SetStrokeColor(cg2.Color(p))
	c.DrawPath(xPos-2.5, H-40, canvas.Rectangle(5, 5))
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
	cg1 = sketchy.NewSimpleGradientFromNamed("cyan", "magenta")
	cg2 = sketchy.NewGradientFromNamed([]string{"blue", "green", "yellow", "red"})
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
