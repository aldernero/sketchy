package sketchy

import (
	"math"
	"math/rand"
	"strconv"

	"github.com/aldernero/gaul"
)

const (
	SliderHeight   = 4.0
	SliderHPadding = 2.0
	SliderVPadding = 2.0
	ToggleHeight   = 5.5
	ToggleHPadding = 3.0
	ToggleVPadding = 2.0
	ButtonHeight   = 5.5
)

type Slider struct {
	Name          string  `json:"Name"`
	MinVal        float64 `json:"MinVal"`
	MaxVal        float64 `json:"MaxVal"`
	Val           float64 `json:"Val"`
	Incr          float64 `json:"Incr"`
	DidJustChange bool    `json:"-"`
}

type Toggle struct {
	Name          string     `json:"Name"`
	Pos           gaul.Point `json:"-"`
	Width         float64    `json:"Width"`
	Height        float64    `json:"Height"`
	Checked       bool       `json:"Checked"`
	IsButton      bool       `json:"IsButton"`
	OutlineColor  string     `json:"OutlineColor"`
	DidJustChange bool       `json:"-"`
}

func (s *Slider) GetPercentage() float64 {
	return gaul.Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *Slider) Randomize() {
	val := gaul.Map(0, 1, s.MinVal, s.MaxVal, rand.Float64())
	s.Val = val
}

func (s *Slider) StringVal() string {
	digits := 0
	if s.Incr < 1 {
		digits = int(math.Ceil(math.Abs(math.Log10(s.Incr))))
	}
	return strconv.FormatFloat(s.Val, 'f', digits, 64)
}
