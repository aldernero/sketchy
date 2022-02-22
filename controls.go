package sketchy

import (
	"math"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	SliderHeight              = 15.0
	SliderHPadding            = 7.0
	SliderVPadding            = 7.0
	SliderMouseWheelThreshold = 0.5
	SliderBackgroundColor     = "#1e1e1e"
	SliderOutlineColor        = "#ffdb00"
	SliderFillColor           = "#ffdb00"
	SliderTextColor           = "#ffffff"
)

type Slider struct {
	Name            string      `json:"Name"`
	Pos             Point       `json:"-"`
	Width           float64     `json:"Width"`
	Height          float64     `json:"Height"`
	MinVal          float64     `json:"MinVal"`
	MaxVal          float64     `json:"MaxVal"`
	Val             float64     `json:"Val"`
	Incr            float64     `json:"Incr"`
	OutlineColor    string      `json:"OutlineColor"`
	BackgroundColor string      `json:"BackgroundColor"`
	FillColor       string      `json:"FillColor"`
	TextColor       string      `json:"TextColor"`
	colors          ColorConfig `json:"-"`
	DidJustChange   bool        `json:"-"`
}

func (s *Slider) GetPercentage() float64 {
	return Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *Slider) GetRect(ctx *gg.Context) Rect {
	x := s.Pos.X - SliderHPadding
	y := s.Pos.Y - SliderVPadding - ctx.FontHeight()
	w := s.Width + 2*SliderHPadding
	h := s.Height + ctx.FontHeight() + SliderVPadding
	return Rect{X: x, Y: y, W: w, H: h}
}

func (s *Slider) AutoHeight(ctx *gg.Context) {
	s.Height = ctx.FontHeight()
}

func (s *Slider) IsInside(x float64, y float64) bool {
	return x >= s.Pos.X &&
		x <= s.Pos.X+s.Width && y >= s.Pos.Y && y <= s.Pos.Y+SliderHeight
}

func (s *Slider) Update(x float64) {
	totalIncr := math.Round((s.MaxVal - s.MinVal) / s.Incr)
	pct := Map(s.Pos.X, s.Pos.X+s.Width, 0, 1, x)
	s.Val = s.MinVal + pct*totalIncr*s.Incr
}

func (s *Slider) CheckAndUpdate() (bool, error) {
	didChange := false
	x, y := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if s.IsInside(float64(x), float64(y)) {
			s.Update(float64(x))
			didChange = true
		}
	} else {
		if s.IsInside(float64(x), float64(y)) {
			_, dy := ebiten.Wheel()
			if math.Abs(dy) > SliderMouseWheelThreshold {
				if dy < 0 {
					s.Val -= s.Incr
				} else {
					s.Val += s.Incr
				}
				s.Val = Clamp(s.MinVal, s.MaxVal, s.Val)
				didChange = true
			}
		}
	}
	s.DidJustChange = didChange
	return didChange, nil
}

func (s *Slider) Draw(ctx *gg.Context) {
	ctx.SetColor(s.colors.Background)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width, SliderHeight)
	ctx.Fill()
	ctx.SetColor(s.colors.Fill)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width*s.GetPercentage(), SliderHeight)
	ctx.Fill()
	ctx.SetColor(s.colors.Outline)
	ctx.DrawRectangle(s.Pos.X, s.Pos.Y, s.Width, SliderHeight)
	ctx.Stroke()
	digits := 0
	if s.Incr < 1 {
		digits = int(math.Ceil(math.Abs(math.Log10(s.Incr))))
	}
	ctx.SetColor(s.colors.Text)
	ctx.DrawStringWrapped(s.Name, s.Pos.X, s.Pos.Y-ctx.FontHeight()-SliderVPadding, 0, 0, s.Width, 1, gg.AlignLeft)
	ctx.DrawStringWrapped(
		strconv.FormatFloat(s.Val, 'f', digits, 64),
		s.Pos.X, s.Pos.Y-ctx.FontHeight()-SliderVPadding,
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

func (s *Slider) parseColors() {
	s.colors.Set(s.BackgroundColor, BackgroundColorType, SliderBackgroundColor)
	s.colors.Set(s.OutlineColor, OutlineColorType, SliderOutlineColor)
	s.colors.Set(s.TextColor, TextColorType, SliderTextColor)
	s.colors.Set(s.FillColor, FillColorType, SliderFillColor)
}
