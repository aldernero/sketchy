package sketchy

import (
	"github.com/aldernero/gaul"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"image/color"
	"math"
	"math/rand"
	"strconv"
)

const (
	SliderHeight              = 4.0
	SliderHPadding            = 2.0
	SliderVPadding            = 2.0
	SliderMouseWheelThreshold = 0.5
	SliderBackgroundColor     = "#1e1e1e"
	SliderOutlineColor        = "#ffdb00"
	SliderFillColor           = "#ffdb00"
	SliderTextColor           = "#ffffff"
	SliderGradientStart       = "cyan"
	SliderGradientEnd         = "magenta"
	ToggleHeight              = 5.5
	ToggleHPadding            = 3.0
	ToggleVPadding            = 2.0
	ToggleBackgroundColor     = "#1e1e1e"
	ToggleOutlineColor        = "#ffdb00"
	ToggleFillColor           = "#ffdb00"
	ToggleTextColor           = "#ffffff"
	ButtonHeight              = 5.5
	TextHeight                = 5.5
	FontSize                  = 10
)

type Slider struct {
	Name               string     `json:"Name"`
	Pos                gaul.Point `json:"-"`
	Width              float64    `json:"Width"`
	Height             float64    `json:"Height"`
	MinVal             float64    `json:"MinVal"`
	MaxVal             float64    `json:"MaxVal"`
	Val                float64    `json:"Val"`
	Incr               float64    `json:"Incr"`
	OutlineColor       string     `json:"OutlineColor"`
	BackgroundColor    string     `json:"BackgroundColor"`
	FillColor          string     `json:"FillColor"`
	UseGradientFill    bool       `json:"UseGradientFill"`
	GradientStartColor string     `json:"GradientStartColor"`
	GradientEndColor   string     `json:"GradientEndColor"`
	TextColor          string     `json:"TextColor"`
	DrawRect           bool       `json:"DrawRect"`
	colors             gaul.ColorConfig
	DidJustChange      bool `json:"-"`
	fontFamily         *canvas.FontFamily
	fontFace           *canvas.FontFace
}

type Toggle struct {
	Name            string     `json:"Name"`
	Pos             gaul.Point `json:"-"`
	Width           float64    `json:"Width"`
	Height          float64    `json:"Height"`
	Checked         bool       `json:"Checked"`
	IsButton        bool       `json:"IsButton"`
	OutlineColor    string     `json:"OutlineColor"`
	BackgroundColor string     `json:"BackgroundColor"`
	FillColor       string     `json:"FillColor"`
	TextColor       string     `json:"TextColor"`
	colors          gaul.ColorConfig
	DidJustChange   bool `json:"-"`
	wasPressed      bool
	fontFamily      *canvas.FontFamily
	fontFace        *canvas.FontFace
}

func (s *Slider) GetPercentage() float64 {
	return gaul.Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *Slider) GetRect() gaul.Rect {
	x := s.Pos.X - SliderHPadding
	y := s.Pos.Y - SliderVPadding
	w := s.Width + 2*SliderHPadding
	h := s.Height + TextHeight + 2*SliderVPadding
	return gaul.Rect{X: x, Y: y, W: w, H: h}
}

func (s *Slider) IsInside(x float64, y float64) bool {
	return x >= s.Pos.X &&
		x <= s.Pos.X+s.Width && y >= s.Pos.Y && y <= s.Pos.Y+SliderHeight
}

func (s *Slider) Update(x float64) {
	totalIncr := (s.MaxVal - s.MinVal) / s.Incr
	pct := gaul.Map(s.Pos.X, s.Pos.X+s.Width, 0, 1, x)
	s.Val = s.MinVal + math.Round(pct*totalIncr)*s.Incr
}

func (s *Slider) CheckAndUpdate(ctx *canvas.Canvas) (bool, error) {
	didChange := false
	x, y := ebiten.CursorPosition()
	sx := MmPerPx * float64(x)
	sy := ctx.H - MmPerPx*float64(y)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if s.IsInside(sx, sy) {
			s.Update(sx)
			didChange = true
		}
	} else {
		if s.IsInside(sx, sy) {
			_, dy := ebiten.Wheel()
			if math.Abs(dy) > SliderMouseWheelThreshold {
				if dy < 0 {
					s.Val -= s.Incr
				} else {
					s.Val += s.Incr
				}
				s.Val = gaul.Clamp(s.MinVal, s.MaxVal, s.Val)
				didChange = true
			}
		}
	}
	s.DidJustChange = didChange
	return didChange, nil
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

func (s *Slider) Draw(ctx *canvas.Context) {
	ctx.SetStrokeWidth(0.3)
	ctx.SetFillColor(s.colors.Background)
	ctx.SetStrokeColor(s.colors.Outline)
	ctx.DrawPath(s.Pos.X, s.Pos.Y, canvas.Rectangle(s.Width, SliderHeight))
	fillColor := s.colors.Fill
	if s.UseGradientFill {
		fillColor = s.colors.Gradient.Color(s.GetPercentage())
	}
	ctx.SetFillColor(fillColor)
	ctx.SetStrokeColor(color.Transparent)
	ctx.DrawPath(s.Pos.X, s.Pos.Y, canvas.Rectangle(s.GetPercentage()*s.Width, SliderHeight))
	ctx.SetStrokeColor(s.colors.Outline)
	ff := s.fontFamily.Face(FontSize, s.colors.Text, canvas.FontRegular, canvas.FontNormal)
	titleText := canvas.NewTextBox(
		ff,
		s.Name,
		s.Width,
		TextHeight,
		canvas.Left,
		canvas.Top,
		0,
		0,
	)
	valText := canvas.NewTextBox(
		ff,
		s.StringVal(),
		s.Width,
		TextHeight,
		canvas.Right,
		canvas.Top,
		0,
		0,
	)
	ctx.DrawText(s.Pos.X, s.Pos.Y+SliderHeight+TextHeight, titleText)
	ctx.DrawText(s.Pos.X, s.Pos.Y+SliderHeight+TextHeight, valText)
	if s.DrawRect {
		rect := s.GetRect()
		ctx.SetFillColor(color.Transparent)
		ctx.SetStrokeColor(color.CMYK{C: 200})
		ctx.SetStrokeWidth(0.3)
		path := canvas.Rectangle(rect.W, rect.H)
		ctx.DrawPath(rect.X, rect.Y, path)
	}
}

func (t *Toggle) GetRect() gaul.Rect {
	x := t.Pos.X - ToggleHPadding
	y := t.Pos.Y + ToggleVPadding
	w := t.Width + 2*ToggleHPadding
	h := t.Height + ToggleVPadding
	return gaul.Rect{X: x, Y: y, W: w, H: h}
}

func (t *Toggle) IsInside(x float64, y float64) bool {
	right := t.Pos.X + t.Height
	if t.IsButton {
		right = t.Pos.X + t.Width
	}
	return x >= t.Pos.X &&
		x <= t.Pos.X+right && y >= t.Pos.Y && y <= t.Pos.Y+t.Height
}

func (t *Toggle) Update() {
	t.Checked = !t.Checked
	if t.Checked {
		t.wasPressed = true
	}
}

func (t *Toggle) CheckAndUpdate(ctx *canvas.Canvas) (bool, error) {
	didChange := false
	x, y := ebiten.CursorPosition()
	sx := MmPerPx * float64(x)
	sy := ctx.H - MmPerPx*float64(y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if t.IsInside(sx, sy) {
			t.Update()
			didChange = true
		}
	}
	t.DidJustChange = didChange
	if t.IsButton && !didChange && t.wasPressed {
		t.wasPressed = false
		t.Checked = false
	}
	return didChange, nil
}

func (t *Toggle) Draw(ctx *canvas.Context) {
	ctx.SetStrokeWidth(0.4)
	ctx.SetFillColor(t.colors.Background)
	ctx.SetStrokeColor(t.colors.Outline)
	if t.IsButton {
		ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Rectangle(t.Width, t.Height))
	} else {
		ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Rectangle(t.Height, t.Height))
	}
	if t.Checked {
		if t.IsButton {
			ctx.SetFillColor(t.colors.Fill)
			ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Rectangle(t.Width, t.Height))
		} else {
			ctx.SetStrokeColor(t.colors.Fill)
			ctx.MoveTo(t.Pos.X, t.Pos.Y)
			ctx.LineTo(t.Pos.X+t.Height, t.Pos.Y+t.Height)
			//ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Line(t.Pos.X+t.Height, t.Pos.Y+t.Height))
			ctx.MoveTo(t.Pos.X, t.Pos.Y+t.Height)
			ctx.LineTo(t.Pos.X+t.Height, t.Pos.Y)
			//ctx.DrawPath(t.Pos.X, t.Pos.Y+t.Height, canvas.Line(t.Pos.X+t.Height, t.Pos.Y))
		}
	}
	ctx.SetStrokeColor(t.colors.Outline)
	if t.IsButton {
		ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Rectangle(t.Width, t.Height))
	} else {
		ctx.DrawPath(t.Pos.X, t.Pos.Y, canvas.Rectangle(t.Height, t.Height))
	}
	ctx.Stroke()
	ctx.SetStrokeColor(t.colors.Text)
	ff := t.fontFamily.Face(FontSize, t.colors.Text, canvas.FontRegular, canvas.FontNormal)
	if t.IsButton {
		toggleText := canvas.NewTextBox(
			ff,
			t.Name,
			t.Width,
			ToggleHeight,
			canvas.Center,
			canvas.Center,
			0,
			0,
		)
		ctx.DrawText(t.Pos.X, t.Pos.Y+TextHeight, toggleText)
	} else {
		toggleText := canvas.NewTextBox(
			ff,
			t.Name,
			t.Width,
			ToggleHeight,
			canvas.Right,
			canvas.Center,
			0,
			0,
		)
		ctx.DrawText(t.Pos.X, t.Pos.Y+TextHeight, toggleText)
	}
}

func (s *Slider) SetFont(ff *canvas.FontFamily) {
	s.fontFamily = ff
	s.fontFace = s.fontFamily.Face(14, color.White, canvas.FontRegular, canvas.FontNormal)
}

func (t *Toggle) SetFont(ff *canvas.FontFamily) {
	t.fontFamily = ff
	t.fontFace = t.fontFamily.Face(14, color.White, canvas.FontRegular, canvas.FontNormal)
}

func (s *Slider) parseColors() {
	s.colors.Set(s.BackgroundColor, gaul.BackgroundColorType, SliderBackgroundColor)
	s.colors.Set(s.OutlineColor, gaul.OutlineColorType, SliderOutlineColor)
	s.colors.Set(s.TextColor, gaul.TextColorType, SliderTextColor)
	s.colors.Set(s.FillColor, gaul.FillColorType, SliderFillColor)
	if s.UseGradientFill {
		c1 := SliderGradientStart
		c2 := SliderGradientEnd
		if s.GradientStartColor != "" {
			c1 = s.GradientStartColor
		}
		if s.GradientEndColor != "" {
			c2 = s.GradientEndColor
		}
		s.colors.Gradient = gaul.NewSimpleGradientFromNamed(c1, c2)
	}
}

func (t *Toggle) parseColors() {
	t.colors.Set(t.BackgroundColor, gaul.BackgroundColorType, ToggleBackgroundColor)
	t.colors.Set(t.OutlineColor, gaul.OutlineColorType, ToggleOutlineColor)
	t.colors.Set(t.TextColor, gaul.TextColorType, ToggleTextColor)
	t.colors.Set(t.FillColor, gaul.FillColorType, ToggleFillColor)
}
