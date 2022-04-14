package main

import (
	"flag"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"strconv"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var qt *sketchy.QuadTree

var nearestPoints []sketchy.Point

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.Toggle("Clear") {
		qt.Clear()
	}
	nearestPoints = []sketchy.Point{}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if s.PointInSketchArea(float64(x), float64(y)) {
			p := s.CanvasCoords(float64(x), float64(y))
			qt.Insert(p)
		}
	} else {
		x, y := ebiten.CursorPosition()
		if s.PointInSketchArea(float64(x), float64(y)) {
			p := s.CanvasCoords(float64(x), float64(y))
			nearestPoints = qt.NearestNeighbors(p, int(s.Slider("Closest Neighbors")))
		}
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
	queryRect := sketchy.Rect{
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
	ff := s.FontFamily.Face(14, canvas.Red, canvas.FontRegular, canvas.FontNormal)
	textBox := canvas.NewTextBox(ff, strconv.FormatInt(int64(len(foundPoints)), 10), 100, 20, canvas.Left, canvas.Bottom, 0, 0)
	c.DrawText(0.1*c.Width(), 0.95*c.Height(), textBox)
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
	qt = sketchy.NewQuadTree(sketchy.Rect{
		X: 0,
		Y: 0,
		W: s.Width(),
		H: s.Height(),
	})
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizable(false)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
