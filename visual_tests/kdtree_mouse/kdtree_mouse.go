package main

import (
	"flag"
	"log"
	"math"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tdewolff/canvas"
	"image/color"
)

var (
	kdtree         *gaul.KDTree
	nearestPoints  []gaul.IndexPoint
	count          int
	prevNeighborFP uint64
)

func neighborFingerprint(pts []gaul.IndexPoint) uint64 {
	var h uint64
	for i, p := range pts {
		h ^= uint64(p.Index)*uint64(i*31+1) ^ math.Float64bits(p.X) ^ (math.Float64bits(p.Y) << 1)
		h *= 0x9e3779b97f4a7c15
	}
	return h
}

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Display", func() {
		ui.FloatSlider("Line Thickness", 0.05, 2, 0.3, 0.05)
		ui.FloatSlider("Point Size", 0, 5, 0.5, 0.1)
		ui.IntSlider("Closest Neighbors", 0, 10, 2, 1)
	})
	ui.Checkbox("Show Points", true)
	ui.Button("Clear")
}

func update(s *sketchy.Sketch) {
	if s.DidTogglesChange && s.Toggle("Clear") {
		kdtree.Clear()
		count = 0
		s.SetBool("", "Clear", false)
		s.MarkDirty()
	}
	nearestPoints = []gaul.IndexPoint{}
	if ok, wx, wy := s.PrimaryPointerPressInSketch(); ok {
		p := s.CanvasCoords(wx, wy)
		kdtree.Insert(p.ToIndexPoint(count))
		count++
	}
	x, y := ebiten.CursorPosition()
	if s.PointInSketchArea(float64(x), float64(y)) {
		p := s.CanvasCoords(float64(x), float64(y))
		nearestPoints = kdtree.NearestNeighbors(p.ToIndexPoint(-1), s.GetInt("Display", "Closest Neighbors"))
	}
	if fp := neighborFingerprint(nearestPoints); fp != prevNeighborFP {
		prevNeighborFP = fp
		s.MarkDirty()
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(color.White)
	c.SetFillColor(color.Transparent)
	c.SetStrokeCapper(canvas.ButtCap)
	c.SetStrokeWidth(s.GetFloat("Display", "Line Thickness"))
	pointSize := s.GetFloat("Display", "Point Size")
	if s.Toggle("Show Points") {
		kdtree.DrawWithPoints(pointSize, c)
	} else {
		kdtree.Draw(c)
	}
	queryRect := gaul.Rect{
		X: 0.4 * c.Width(),
		Y: 0.4 * c.Height(),
		W: 0.2 * c.Width(),
		H: 0.2 * c.Height(),
	}
	foundPoints := kdtree.Query(queryRect)
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
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:                 "KDTree Interaction Test",
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
	w := s.Width()
	h := s.Height()
	rect := gaul.Rect{
		X: 0,
		Y: 0,
		W: w,
		H: h,
	}
	kdtree = gaul.NewKDTree(rect)
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
