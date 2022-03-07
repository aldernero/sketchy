package main

import (
	"flag"
	"log"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

var synthwavePalette sketchy.Gradient

func drawLines(l sketchy.Line, n int, p float64, ctx *gg.Context) {
	if n == 0 {
		return
	}
	left := l.P
	middle := l.Midpoint()
	right := l.Q
	L := l.Length()
	pb := l.PerpendicularBisector(p * L)
	colorPercent := sketchy.Map(2, 100, 0, 1, L)
	if n%2 == 0 {
		colorPercent = 1 - colorPercent
	}
	ctx.SetColor(synthwavePalette.Color(colorPercent))
	pb.Draw(ctx)
	ctx.Stroke()
	drawLines(sketchy.Line{P: middle, Q: pb.P}, n-1, p, ctx)
	drawLines(sketchy.Line{P: middle, Q: pb.Q}, n-1, p, ctx)
	drawLines(sketchy.Line{P: left, Q: middle}, n-1, p, ctx)
	drawLines(sketchy.Line{P: middle, Q: right}, n-1, p, ctx)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	c.SetLineCapButt()
	c.SetLineWidth(1.5)
	line := sketchy.Line{
		P: sketchy.Point{X: 20, Y: s.SketchHeight / 2},
		Q: sketchy.Point{X: s.SketchWidth - 20, Y: s.SketchHeight / 2},
	}
	line.Draw(c)
	drawLines(line, int(s.Slider("depth")), s.Slider("persistence"), c)
	c.Stroke()
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	synthwavePalette = sketchy.NewGradientFromNamed([]string{"#1bbbd9", "#f900a4"})
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
