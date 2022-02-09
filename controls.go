package genart

import (
	"image/color"
	"math"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	sliderHeight              = 10.0
	sliderHPadding            = 5.0
	sliderVPadding            = 2.0
	sliderMouseWheelThreshold = 0.5
)

var sliderOutline color.Color = color.Black
var sliderBackground color.Color = color.White
var sliderFill color.Color = color.Gray{Y: 128}

type Slider struct {
	Name   string
	Pos    Point
	Width  float64
	Height float64
	MinVal float64
	MaxVal float64
	Val    float64
	Incr   float64
}

func (s *Slider) GetPercentage() float64 {
	return Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *Slider) GetRect(ctx *gg.Context) Rect {
	x := s.Pos.X - sliderHPadding
	y := s.Pos.Y - sliderVPadding - ctx.FontHeight()
	w := s.Width + 2*sliderHPadding
	h := s.Height + ctx.FontHeight() + sliderVPadding
	return Rect{X: x, Y: y, W: w, H: h}
}

func (s *Slider) AutoHeight(ctx *gg.Context) {
	s.Height = ctx.FontHeight()
}

func (s *Slider) IsInside(x float64, y float64) bool {
	return x >= s.Pos.X &&
		x <= s.Pos.X+s.Width && y >= s.Pos.Y && y <= s.Pos.Y+sliderHeight
}

func (s *Slider) Update(x float64) {
	totalIncr := math.Round((s.MaxVal - s.MinVal) / s.Incr)
	pct := Map(s.Pos.X, s.Pos.X+s.Width, 0, 1, x)
	s.Val = s.MinVal + pct*totalIncr*s.Incr
}

func (s *Slider) CheckAndUpdate() error {
	x, y := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if s.IsInside(float64(x), float64(y)) {
			s.Update(float64(x))
		}
	} else {
		if s.IsInside(float64(x), float64(y)) {
			_, dy := ebiten.Wheel()
			if math.Abs(dy) > sliderMouseWheelThreshold {
				if dy < 0 {
					s.Val -= s.Incr
				} else {
					s.Val += s.Incr
				}
			}
		}
	}
	return nil
}

func (s *Slider) Draw(ctx *gg.Context) {
	ctx.SetColor(sliderBackground)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width, sliderHeight)
	ctx.Fill()
	ctx.SetColor(sliderFill)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width*s.GetPercentage(), sliderHeight)
	ctx.Fill()
	ctx.SetColor(sliderOutline)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width, sliderHeight)
	ctx.Stroke()
	digits := 0
	if s.Incr < 1 {
		digits = int(math.Ceil(math.Abs(math.Log10(s.Incr))))
	}
	ctx.DrawStringWrapped(s.Name, s.Pos.X, s.Pos.Y-ctx.FontHeight()-sliderVPadding, 0, 0, s.Width, 1, gg.AlignLeft)
	ctx.DrawStringWrapped(
		strconv.FormatFloat(s.Val, 'f', digits, 64),
		s.Pos.X, s.Pos.Y-ctx.FontHeight()-sliderVPadding,
		0, 0, s.Width, 1, gg.AlignRight)
}

func NewIntStepSlider(name string, minVal int, maxVal int) Slider {
	s := Slider{
		Name:   name,
		Pos:    Point{},
		Width:  0,
		MinVal: float64(minVal),
		MaxVal: float64(maxVal),
		Val:    float64(minVal),
		Incr:   1,
	}
	return s
}

func NewRadiansSlider(name string, steps int) Slider {
	s := Slider{
		Name:   name,
		Pos:    Point{},
		Width:  0,
		MinVal: 0,
		MaxVal: Tau,
		Val:    0,
		Incr:   Tau / float64(steps),
	}
	return s
}
