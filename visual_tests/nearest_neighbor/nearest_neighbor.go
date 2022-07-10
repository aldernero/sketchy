package main

import (
	"flag"
	"fmt"
	gaul "github.com/aldernero/gaul"
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

func setup(s *sketchy.Sketch) {
	points = []gaul.Point{}
	rect := s.CanvasRect()
	tree = gaul.NewKDTree(rect)
	currentPoint = gaul.IndexPoint{
		Index: -1,
		Point: gaul.Point{X: 0, Y: 0},
	}
	// Get a sample of random points, add them to the tree
	numPoints := int(s.Slider("numPoints"))
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
	setup(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
