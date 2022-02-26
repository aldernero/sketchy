package sketchy

import (
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
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
	var c color.Color
	if hexString != "" {
		colorHexString = hexString
	}
	if colorHexString == "" {
		c = color.Transparent
	} else {
		d, err := colorful.Hex(colorHexString)
		if err != nil {
			panic(err)
		}
		c = d
	}
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
