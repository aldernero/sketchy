package sketchy

import (
	"math"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	CheckboxHeight            = 15.0
	CheckboxHPadding          = 12.0
	CheckboxVPadding          = 7.0
	CheckboxBackgroundColor   = "#1e1e1e"
	CheckboxOutlineColor      = "#ffdb00"
	CheckboxFillColor         = "#ffdb00"
	CheckboxTextColor         = "#ffffff"
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

type Checkbox struct {
	Name            string      `json:"Name"`
	Pos             Point       `json:"-"`
	Width           float64     `json:"Width"`
	Height          float64     `json:"Height"`
	Checked         bool        `json:"Checked"`
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

func (c *Checkbox) GetRect(ctx *gg.Context) Rect {
	x := c.Pos.X - CheckboxHPadding
	y := c.Pos.Y - CheckboxVPadding
	w := c.Width + 2*CheckboxHPadding
	h := c.Height + CheckboxVPadding
	return Rect{X: x, Y: y, W: w, H: h}
}

func (c *Checkbox) AutoHeight(ctx *gg.Context) {
	c.Height = ctx.FontHeight()
}

func (c *Checkbox) IsInside(x float64, y float64) bool {
	return x >= c.Pos.X &&
		x <= c.Pos.X+c.Height && y >= c.Pos.Y && y <= c.Pos.Y+c.Height
}

func (c *Checkbox) Update() {
	c.Checked = !c.Checked
}

func (c *Checkbox) CheckAndUpdate() (bool, error) {
	didChange := false
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if c.IsInside(float64(x), float64(y)) {
			c.Update()
			didChange = true
		}
	}
	c.DidJustChange = didChange
	return didChange, nil
}

func (c *Checkbox) Draw(ctx *gg.Context) {
	ctx.SetLineCapButt()
	ctx.SetLineWidth(1.5)
	ctx.SetColor(c.colors.Background)
	ctx.DrawRectangle(c.Pos.X, c.Pos.Y, c.Height, c.Height)
	ctx.Fill()
	if c.Checked {
		ctx.SetColor(c.colors.Fill)
		ctx.DrawLine(c.Pos.X, c.Pos.Y, c.Pos.X+c.Height, c.Pos.Y+c.Height)
		ctx.DrawLine(c.Pos.X, c.Pos.Y+c.Height, c.Pos.X+c.Height, c.Pos.Y)
		ctx.Stroke()
	}
	ctx.SetColor(c.colors.Outline)
	ctx.DrawRectangle(c.Pos.X, c.Pos.Y, c.Height, c.Height)
	ctx.Stroke()
	ctx.SetColor(c.colors.Text)
	ctx.DrawStringWrapped(c.Name, c.Pos.X, c.Pos.Y, 0, 0, c.Width, 1, gg.AlignRight)
}

func (s *Slider) parseColors() {
	s.colors.Set(s.BackgroundColor, BackgroundColorType, SliderBackgroundColor)
	s.colors.Set(s.OutlineColor, OutlineColorType, SliderOutlineColor)
	s.colors.Set(s.TextColor, TextColorType, SliderTextColor)
	s.colors.Set(s.FillColor, FillColorType, SliderFillColor)
}

func (c *Checkbox) parseColors() {
	c.colors.Set(c.BackgroundColor, BackgroundColorType, CheckboxBackgroundColor)
	c.colors.Set(c.OutlineColor, OutlineColorType, CheckboxOutlineColor)
	c.colors.Set(c.TextColor, TextColorType, CheckboxTextColor)
	c.colors.Set(c.FillColor, FillColorType, CheckboxFillColor)
}
