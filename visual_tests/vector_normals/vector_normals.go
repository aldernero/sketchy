package main

import (
	"flag"
	"fmt"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func update(s *sketchy.Sketch) {
	// Update logic goes here
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	var points []gaul.Point
	ls := gaul.Linspace(0, gaul.Tau, int(s.Slider("num_points"))+1, true)
	for _, i := range ls {
		x := 0.5*c.Width() + 50*math.Cos(i)
		y := 0.5*c.Height() + 50*math.Sin(i)
		points = append(points, gaul.Point{X: x, Y: y})
	}
	curve := gaul.Curve{Points: points, Closed: true}
	c.SetStrokeColor(color.White)
	curve.Draw(c)
	c.SetStrokeColor(sketchy.StringToColor("magenta"))
	c.SetStrokeWidth(0.3)
	for i := 0; i < len(points)-1; i++ {
		p := points[i]
		q := points[i+1]
		l := gaul.Line{P: p, Q: q}
		m := l.Midpoint()
		vec := gaul.Vec2{X: q.X - p.X, Y: q.Y - p.Y}
		norm := vec.UnitNormal()
		norm = norm.Scale(s.Slider("scale"))
		if s.Tick == 1 {
			fmt.Println(m.X, m.Y, norm.X, norm.Y)
		}
		c.MoveTo(m.X, m.Y)
		c.LineTo(m.X+norm.X, m.Y+norm.Y)
		c.Stroke()
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
