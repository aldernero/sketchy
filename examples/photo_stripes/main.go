package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	sourceImageName = "cloud"
	outputImageName = "output"
)

var (
	out         *image.RGBA
	stripShifts []int
)

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Strips", func() {
		ui.Dropdown("direction", []string{"horizontal", "vertical"}, 0)
		ui.Dropdown("type", []string{"uniform random", "gaussian", "alternating", "cumulative"}, 0)
		ui.IntSlider("strip size", 1, 200, 24, 1)
		ui.IntSlider("max shift", 0, 250, 60, 1)
	})
	ui.Button("reshuffle")
}

func update(s *sketchy.Sketch) {
	if s.DidTogglesChange && s.Toggle("reshuffle") {
		s.SetBool("", "reshuffle", false)
		setup(s)
		s.MarkDirty()
		return
	}
	if s.DidControlsChange || s.DidSlidersChange || s.DidDropdownsChange {
		setup(s)
		s.MarkDirty()
	}
}

func draw(s *sketchy.Sketch, c *render.Context) {
	s.DrawNamedImage(c, outputImageName)
}

func setup(s *sketchy.Sketch) {
	src := s.Image(sourceImageName)
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if out == nil || out.Bounds().Dx() != w || out.Bounds().Dy() != h {
		out = image.NewRGBA(image.Rect(0, 0, w, h))
	}
	for i := range out.Pix {
		out.Pix[i] = 0
	}

	stripSize := s.GetInt("Strips", "strip size")
	maxShift := s.GetInt("Strips", "max shift")
	if stripSize < 1 {
		stripSize = 1
	}

	horizontal := s.GetDropdownIndex("Strips", "direction") == 0
	shiftType := s.GetDropdownIndex("Strips", "type")

	span := h
	if !horizontal {
		span = w
	}
	numStrips := (span + stripSize - 1) / stripSize
	stripShifts = computeStripShifts(s, numStrips, maxShift, shiftType)

	if horizontal {
		for y := 0; y < h; y++ {
			strip := y / stripSize
			if strip >= len(stripShifts) {
				strip = len(stripShifts) - 1
			}
			shift := stripShifts[strip]
			for x := 0; x < w; x++ {
				xd := wrap(x+shift, w)
				set(out, b, xd, y, src.At(x, y))
			}
		}
	} else {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				strip := x / stripSize
				if strip >= len(stripShifts) {
					strip = len(stripShifts) - 1
				}
				shift := stripShifts[strip]
				yd := wrap(y+shift, h)
				set(out, b, x, yd, src.At(x, y))
			}
		}
	}
	s.RegisterImage(outputImageName, out)
}

const (
	shiftUniform = iota
	shiftGaussian
	shiftAlternating
	shiftCumulative
)

func computeStripShifts(s *sketchy.Sketch, numStrips, maxShift, shiftType int) []int {
	shifts := make([]int, numStrips)
	switch shiftType {
	case shiftGaussian:
		stdev := float64(maxShift) / 3
		if stdev <= 0 {
			stdev = 1
		}
		for i := range shifts {
			g := s.Rand.Gaussian(0, stdev)
			shifts[i] = clampShift(int(math.Round(g)), maxShift)
		}
	case shiftAlternating:
		for i := range shifts {
			if i%2 == 0 {
				shifts[i] = maxShift
			} else {
				shifts[i] = -maxShift
			}
		}
	case shiftCumulative:
		for i := range shifts {
			shifts[i] = i * maxShift
		}
	default: // shiftUniform
		for i := range shifts {
			u := s.Rand.Prng.Float64()
			shifts[i] = int(gaul.Map(0, 1, float64(-maxShift), float64(maxShift), u) + 0.5)
		}
	}
	return shifts
}

func clampShift(v, maxShift int) int {
	if v > maxShift {
		return maxShift
	}
	if v < -maxShift {
		return -maxShift
	}
	return v
}

// wrap maps a coordinate as if the image wraps on a cylinder along that axis.
func wrap(v, size int) int {
	v %= size
	if v < 0 {
		v += size
	}
	return v
}

func set(out *image.RGBA, b image.Rectangle, x, y int, c color.Color) {
	if image.Pt(x, y).In(b) {
		out.Set(x, y, c)
	}
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()

	s := sketchy.New(sketchy.Config{
		Title:        "Photo Strips",
		SketchWidth:  533,
		SketchHeight: 800,
		Images: []sketchy.ImageAsset{
			{Name: sourceImageName, Path: "cloud.png"},
		},
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
