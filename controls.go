package sketchy

import (
	"math"
	"math/rand"
	"strconv"

	"github.com/aldernero/gaul"
)

const (
	DefaultControlWindowWidth  = 250
	DefaultControlWindowHeight = 500
	DefaultControlWindowX      = 25
	DefaultControlWindowY      = 25
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
