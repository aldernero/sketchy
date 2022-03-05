package sketchy

import (
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
	"image/color"
	"regexp"
	"strings"
)

type ColorType int

const (
	BackgroundColorType = iota
	OutlineColorType
	TextColorType
	FillColorType
)

type ColorConfig struct {
	Background color.Color
	Outline    color.Color
	Text       color.Color
	Fill       color.Color
}

func (cc *ColorConfig) Set(hexString string, colorType ColorType, defaultString string) {
	colorHexString := defaultString
	if hexString != "" {
		colorHexString = hexString
	}
	c := StringToColor(colorHexString)
	switch colorType {
	case BackgroundColorType:
		cc.Background = c
	case OutlineColorType:
		cc.Outline = c
	case TextColorType:
		cc.Text = c
	case FillColorType:
		cc.Fill = c
	}
}

func StringToColor(colorString string) color.Color {
	if colorString == "" {
		return color.Transparent
	}
	re := regexp.MustCompile("#[0-9a-f]{6}")
	name := strings.ToLower(colorString)
	if re.MatchString(name) {
		c, err := colorful.Hex(name)
		if err != nil {
			panic(err)
		}
		return c
	}
	return NamedColor(name)
}

func NamedColor(name string) color.Color {
	val, ok := colornames.Map[strings.ToLower(name)]
	if !ok {
		panic("invalid color name")
	}
	return val
}

type SimpleGradient struct {
	startColor color.Color
	endColor   color.Color
}

func NewSimpleGradientFromNamed(c1, c2 string) SimpleGradient {
	gradient := SimpleGradient{
		startColor: NamedColor(c1),
		endColor:   NamedColor(c2),
	}
	return gradient
}

func (sg *SimpleGradient) Color(percentage float64) color.Color {
	val := Clamp(0, 1, percentage)
	c1, _ := colorful.MakeColor(sg.startColor)
	c2, _ := colorful.MakeColor(sg.endColor)
	return c1.BlendHcl(c2, val)
}
