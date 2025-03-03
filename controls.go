package sketchy

import (
	"fmt"
	"github.com/aldernero/gaul"
	"github.com/ebitengine/debugui"
	"image"
	"math"
	"math/rand"
	"strconv"
)

const (
	DefaultControlWindowWidth  = 250
	DefaultControlWindowHeight = 500
	DefaultControlWindowX      = 25
	DefaultControlWindowY      = 25
	DefaultSliderTextWidth     = 100
	DefaultCheckboxColumns     = 1
	DefaultButtonColumns       = 1
)

type Slider struct {
	Name          string  `json:"Name"`
	MinVal        float64 `json:"MinVal"`
	MaxVal        float64 `json:"MaxVal"`
	Val           float64 `json:"Val"`
	Incr          float64 `json:"Incr"`
	DidJustChange bool    `json:"-"`
	lastVal       float64
	digits        int
	dontRandomize bool
}

type Toggle struct {
	Name          string     `json:"Name"`
	Pos           gaul.Point `json:"-"`
	Width         float64    `json:"Width"`
	Height        float64    `json:"Height"`
	Checked       bool       `json:"Checked"`
	IsButton      bool       `json:"IsButton"`
	DidJustChange bool       `json:"-"`
	lastVal       bool
}

func NewSlider(name string, min, max, val, incr float64) Slider {
	s := Slider{
		Name:   name,
		MinVal: min,
		MaxVal: max,
		Val:    val,
		Incr:   incr,
	}
	s.lastVal = val
	s.CalcDigits()
	return s
}

func (s *Slider) GetPercentage() float64 {
	return gaul.Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *Slider) Randomize() {
	val := gaul.Map(0, 1, s.MinVal, s.MaxVal, rand.Float64())
	s.Val = val
}

func (s *Slider) StringVal() string {
	return strconv.FormatFloat(s.Val, 'f', s.digits, 64)
}

func (s *Slider) CalcDigits() {
	s.digits = calcDigits(s.Incr)
}

func (s *Slider) UpdateState() {
	if s.Val != s.lastVal {
		s.DidJustChange = true
	} else {
		s.DidJustChange = false
	}
	s.lastVal = s.Val
}

func (t *Toggle) UpdateState() {
	if t.Checked != t.lastVal {
		t.DidJustChange = true
	} else {
		t.DidJustChange = false
	}
	t.lastVal = t.Checked
}

func calcDigits(val float64) int {
	if val < 1 {
		return int(math.Ceil(math.Abs(math.Log10(val))))
	}
	return 0
}

func (s *Sketch) controlWindow(ctx *debugui.Context) {
	ctx.Window("Control Panel", image.Rect(DefaultControlWindowX, DefaultControlWindowY, s.ControlWidth, s.ControlHeight), func(res debugui.Response, layout debugui.Layout) {
		if ctx.Header("Builtins", true) != 0 {
			ctx.TreeNode("Random Seed", func(res debugui.Response) {
				ctx.Label(fmt.Sprintf("Seed: %d", s.RandomSeed))
				ctx.SetLayoutRow([]int{40, 40, 40}, 0)
				if ctx.Button("Decr") != 0 {
					s.RandomSeed--
				}
				if ctx.Button("Incr") != 0 {
					s.RandomSeed++
				}
				if ctx.Button("Rand") != 0 {
					s.RandomSeed = rand.Int63()
				}
			})
			ctx.TreeNode("Save Options", func(res debugui.Response) {
				ctx.SetLayoutRow([]int{80, 80}, 0)
				if ctx.Button("Save as PNG") != 0 {
					s.isSavingPNG = true
				}
				if ctx.Button("Save as SVG") != 0 {
					s.isSavingSVG = true
				}
				if ctx.Button("Save Config") != 0 {
					s.saveConfig()
				}
				if ctx.Button("Dump State") != 0 {
					s.DumpState()
				}
			})
		}
		if ctx.Header("Sliders", true) != 0 {
			if ctx.Button("Randomize (unless checked)") != 0 {
				for i := range s.Sliders {
					if !s.Sliders[i].dontRandomize {
						s.Sliders[i].Randomize()
					}
				}
			}
			ctx.SetLayoutRow([]int{20, s.SliderTextWidth, -1}, 0)
			for i := range s.Sliders {
				ctx.Checkbox("", &s.Sliders[i].dontRandomize)
				ctx.Label(s.Sliders[i].Name)
				ctx.Slider(&s.Sliders[i].Val, s.Sliders[i].MinVal, s.Sliders[i].MaxVal, s.Sliders[i].Incr, s.Sliders[i].digits)
			}
		}
		if ctx.Header("Checkboxes", true) != 0 {
			ctx.SetLayoutRow(getRowIntSlice(s.CheckboxColumns), 0)
			for i := range s.Toggles {
				if !s.Toggles[i].IsButton {
					ctx.Checkbox(s.Toggles[i].Name, &s.Toggles[i].Checked)
				}
			}
		}
		if ctx.Header("Buttons", true) != 0 {
			ctx.SetLayoutRow(getRowIntSlice(s.ButtonColumns), 0)
			for i := range s.Toggles {
				if s.Toggles[i].IsButton {
					if ctx.Button(s.Toggles[i].Name) != 0 {
						s.Toggles[i].Checked = !s.Toggles[i].Checked
					}
				}
			}
		}
	})
}

func getRowIntSlice(col int) []int {
	result := make([]int, col)
	for i := range result {
		result[i] = -1
	}
	return result
}
