package main

import (
	"flag"
	"fmt"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var points []gaul.Point
var tree *gaul.KDTree
var currentPoint gaul.IndexPoint

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.IntSlider("numPoints", 1, 10000, 100, 1)
	ui.Checkbox("Show Points", false)
}

func setup(s *sketchy.Sketch) {
	points = []gaul.Point{}
	rect := s.CanvasRect()
	tree = gaul.NewKDTree(rect)
	currentPoint = gaul.IndexPoint{
		Index: -1,
		Point: gaul.Point{X: 0, Y: 0},
	}
	// Get a sample of random points, add them to the tree
	numPoints := s.GetInt("", "numPoints")
	radius := 0.25 * rect.W
	for i := 0; i < numPoints; i++ {
		r := rand.Float64() * radius
		theta := rand.Float64() * gaul.Tau
		point := gaul.IndexPoint{
			Index: i,
			Point: gaul.Point{
				X: r*math.Cos(theta) + rect.W/2,
				Y: r*math.Sin(theta) + rect.H/2,
			},
		}
		tree.Insert(point)
	}
	for len(points) < numPoints {
		neighbor := tree.NearestNeighbors(currentPoint, 1)
		if len(neighbor) == 0 {
			break
		}
		points = append(points, neighbor[0].Point)
		q := tree.UpdateIndex(neighbor[0], -1)
		if q == nil {
			fmt.Println("something went wrong")
		}
		currentPoint = neighbor[0]
	}
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	if s.Toggle("Show Points") {
		c.SetFillColor(color.Transparent)
		c.SetStrokeColor(canvas.Blue)
		c.SetStrokeWidth(0.3)
		for _, p := range points {
			p.Draw(1, c)
		}
	}
	c.SetStrokeColor(color.White)
	c.SetStrokeWidth(0.3)
	curve := gaul.Curve{
		Points: points,
		Closed: false,
	}
	curve.Draw(c)
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:                 "Nearest neighbor walk",
		SketchWidth:           800,
		SketchHeight:          800,
		SketchBackgroundColor: "#1e1e1e",
		ControlOutlineColor:   "#ffdb00",
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
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
