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
	Name   string  `json:"Name"`
	MinVal float64 `json:"MinVal"`
	MaxVal float64 `json:"MaxVal"`
	Val    float64 `json:"Val"`
	Incr   float64 `json:"Incr"`
	digits int
}

type Toggle struct {
	Name         string     `json:"Name"`
	Pos          gaul.Point `json:"-"`
	Width        float64    `json:"Width"`
	Height       float64    `json:"Height"`
	Checked      bool       `json:"Checked"`
	IsButton     bool       `json:"IsButton"`
	OutlineColor string     `json:"OutlineColor"`
}

func NewSlider(name string, min, max, val, incr float64) Slider {
	digits := 0
	if incr < 1 {
		digits = int(math.Ceil(math.Abs(math.Log10(incr))))
	}
	return Slider{
		Name:   name,
		MinVal: min,
		MaxVal: max,
		Val:    val,
		Incr:   incr,
		digits: digits,
	}
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
