package main

import (
	"flag"
	"github.com/aldernero/gaul"
	"github.com/tdewolff/canvas"
	"log"
	"math"
	"os"
	"runtime/pprof"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

var triangles []gaul.Triangle
var curves1 []gaul.Curve
var curves2 []gaul.Curve
var curves3 []gaul.Curve
var presses int
var rng gaul.LFSRLarge

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Bezier", func() {
		ui.IntSlider("num", 1, 400, 200, 1)
		ui.FloatSlider("radius", 0, 1, 0.16, 0.01)
		ui.FloatSlider("xoffset", 0, 1, 0.30, 0.01)
		ui.FloatSlider("yoffset", 0, 1, 0.45, 0.01)
	})
	ui.Checkbox("show triangles", false)
	ui.Checkbox("symmetric", false)
	ui.Button("next")
}

func genCurves(num int, triangle gaul.Triangle) ([]gaul.Curve, []gaul.Curve, []gaul.Curve) {
	AB := gaul.Line{P: triangle.A, Q: triangle.B}
	BC := gaul.Line{P: triangle.B, Q: triangle.C}
	CA := gaul.Line{P: triangle.C, Q: triangle.A}
	c1 := make([]gaul.Curve, num)
	c2 := make([]gaul.Curve, num)
	c3 := make([]gaul.Curve, num)
	for i := 0; i < num; i++ {
		p := rng.Float64()
		q := rng.Float64()
		side1 := rng.Uint64n(3)
		side2 := rng.Float64()
		switch side1 {
		case 0: // AB
			if side2 < 0.5 { // BC
				P := AB.Lerp(p)
				Q := BC.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.B)
				c2[i] = gaul.CubicBezier(P, Q, triangle.A, triangle.C)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.B, triangle.A, triangle.C)
			} else { // CA
				P := AB.Lerp(p)
				Q := CA.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.A)
				c2[i] = gaul.CubicBezier(P, Q, triangle.B, triangle.C)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.A, triangle.B, triangle.C)
			}
		case 1: // BC
			if side2 < 0.5 { // CA
				P := BC.Lerp(p)
				Q := CA.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.C)
				c2[i] = gaul.CubicBezier(P, Q, triangle.B, triangle.A)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.C, triangle.B, triangle.A)
			} else { // AB
				P := BC.Lerp(p)
				Q := AB.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.B)
				c2[i] = gaul.CubicBezier(P, Q, triangle.C, triangle.A)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.B, triangle.C, triangle.A)
			}
		case 2: // CA
			if side2 < 0.5 { // AB
				P := CA.Lerp(p)
				Q := AB.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.A)
				c2[i] = gaul.CubicBezier(P, Q, triangle.C, triangle.B)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.A, triangle.C, triangle.B)
			} else { // BC
				P := CA.Lerp(p)
				Q := BC.Lerp(q)
				c1[i] = gaul.QuadBezier(P, Q, triangle.C)
				c2[i] = gaul.CubicBezier(P, Q, triangle.A, triangle.B)
				c3[i] = gaul.QuarticBezier(P, Q, triangle.C, triangle.A, triangle.B)

			}
		}
	}
	return c1, c2, c3
}

func setup(s *sketchy.Sketch) {
	// Setup logic goes here
	num := s.GetInt("Bezier", "num")
	radius := s.GetFloat("Bezier", "radius")
	W := s.Width()
	H := s.Height()
	xoffset := s.GetFloat("Bezier", "xoffset") * W
	yoffset := s.GetFloat("Bezier", "yoffset") * H
	R := radius * W
	rng = gaul.NewLFSRLargeWithSeed(uint64(s.RandomSeed))
	curves1 = []gaul.Curve{}
	curves2 = []gaul.Curve{}
	curves3 = []gaul.Curve{}
	angle1 := gaul.Pi / 2
	angle2 := -gaul.Pi / 2
	triangle2 := gaul.Triangle{
		A: gaul.Point{
			X: W/2 + R*math.Cos(angle2-gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle2-gaul.Tau/3),
		},
		B: gaul.Point{
			X: W/2 + R*math.Cos(angle2),
			Y: yoffset + R*math.Sin(angle2),
		},
		C: gaul.Point{
			X: W/2 + R*math.Cos(angle2+gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle2+gaul.Tau/3),
		},
	}
	triangle1 := gaul.Triangle{
		A: gaul.Point{
			X: W/2 - xoffset + R*math.Cos(angle1-gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle1-gaul.Tau/3),
		},
		B: gaul.Point{
			X: W/2 - xoffset + R*math.Cos(angle1),
			Y: yoffset + R*math.Sin(angle1),
		},
		C: gaul.Point{
			X: W/2 - xoffset + R*math.Cos(angle1+gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle1+gaul.Tau/3),
		},
	}
	triangle3 := gaul.Triangle{
		A: gaul.Point{
			X: W/2 + xoffset + R*math.Cos(angle1-gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle1-gaul.Tau/3),
		},
		B: gaul.Point{
			X: W/2 + xoffset + R*math.Cos(angle1),
			Y: yoffset + R*math.Sin(angle1),
		},
		C: gaul.Point{
			X: W/2 + xoffset + R*math.Cos(angle1+gaul.Tau/3),
			Y: yoffset + R*math.Sin(angle1+gaul.Tau/3),
		},
	}
	triangles = []gaul.Triangle{triangle1, triangle2, triangle3}
	curves1, _, _ = genCurves(num, triangle1)
	_, curves2, _ = genCurves(num, triangle2)
	_, _, curves3 = genCurves(num, triangle3)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	if s.Toggle("next") {
		presses++
		return
	}
	if s.DidControlsChange {
		setup(s)
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetFillColor(canvas.Transparent)
	c.SetStrokeColor(canvas.White)
	c.SetStrokeWidth(0.25)
	if s.Toggle("show triangles") {
		for _, triangle := range triangles {
			triangle.Draw(c)
		}
	}
	for _, curve := range curves1 {
		curve.Draw(c)
	}

	for _, curve := range curves2 {
		curve.Draw(c)
	}
	for _, curve := range curves3 {
		curve.Draw(c)
	}
}

func main() {
	var prefix string
	var randomSeed int64
	var cpuprofile = flag.String("pprof", "", "Collect CPU profile")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	s := sketchy.New(sketchy.Config{
		Title:                 "Bezier Sketch",
		SketchWidth:           1920,
		SketchHeight:          1080,
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
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
