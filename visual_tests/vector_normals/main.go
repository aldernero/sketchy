package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *gg.Context) {
	// Drawing code goes here
	var points []sketchy.Point
	ls := sketchy.Linspace(0, sketchy.Tau, int(s.Slider("num_points"))+1, true)
	for _, i := range ls {
		x := 0.5*s.SketchWidth + 200*math.Cos(i)
		y := 0.5*s.SketchHeight + 200*math.Sin(i)
		points = append(points, sketchy.Point{X: x, Y: y})
	}
	curve := sketchy.Curve{Points: points, Closed: true}
	c.SetLineCapButt()
	c.SetLineWidth(2)
	c.SetColor(color.White)
	curve.Draw(c)
	c.Stroke()
	c.SetColor(sketchy.StringToColor("magenta"))
	for i := 0; i < len(points)-1; i++ {
		p := points[i]
		q := points[i+1]
		l := sketchy.Line{P: p, Q: q}
		m := l.Midpoint()
		vec := sketchy.Vec2{X: q.X - p.X, Y: q.Y - p.Y}
		norm := vec.UnitNormal()
		norm = norm.Scale(s.Slider("scale"))
		c.DrawLine(m.X, m.Y, m.X+norm.X, m.Y+norm.Y)
		c.Stroke()
	}
}

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "sketch", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Prefix = prefix
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
