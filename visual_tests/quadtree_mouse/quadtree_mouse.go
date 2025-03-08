package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
)

var (
	qt            *gaul.QuadTree
	nearestPoints []gaul.IndexPoint
	count         int
)

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.Toggle("Clear") {
		qt.Clear()
		count = 0
	}
	nearestPoints = []gaul.IndexPoint{}
	if !s.IsMouseOverControlPanel() && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if s.PointInSketchArea(float64(x), float64(y)) {
			p := s.CanvasCoords(float64(x), float64(y))
			qt.Insert(p.ToIndexPoint(count))
			count++
		}
	}
	x, y := ebiten.CursorPosition()
	if s.PointInSketchArea(float64(x), float64(y)) {
		p := s.CanvasCoords(float64(x), float64(y))
		nearestPoints = qt.NearestNeighbors(p.ToIndexPoint(-1), int(s.Slider("Closest Neighbors")))
	}

}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(color.White)
	c.SetFillColor(color.Transparent)
	c.SetStrokeCapper(canvas.ButtCap)
	c.SetStrokeWidth(s.Slider("Line Thickness"))
	pointSize := s.Slider("Point Size")
	if s.Toggle("Show Points") {
		qt.DrawWithPoints(pointSize, c)
	} else {
		qt.Draw(c)
	}
	queryRect := gaul.Rect{
		X: 0.4 * c.Width(),
		Y: 0.4 * c.Height(),
		W: 0.2 * c.Width(),
		H: 0.2 * c.Height(),
	}
	foundPoints := qt.Query(queryRect)
	c.SetStrokeColor(canvas.Blue)
	c.DrawPath(queryRect.X, queryRect.Y, canvas.Rectangle(queryRect.W, queryRect.H))
	for _, p := range foundPoints {
		p.Draw(pointSize, c)
	}
	c.SetStrokeColor(canvas.Magenta)
	if len(nearestPoints) > 0 {
		for _, p := range nearestPoints {
			p.Draw(pointSize, c)
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
	qt = gaul.NewQuadTree(gaul.Rect{
		X: 0,
		Y: 0,
		W: s.Width(),
		H: s.Height(),
	})
	ebiten.SetWindowSize(int(s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
