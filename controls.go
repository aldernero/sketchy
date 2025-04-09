package sketchy

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/aldernero/gaul"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
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

type ColorPicker struct {
	Name          string `json:"Name"`
	Color         string `json:"Color"`
	DidJustChange bool   `json:"-"`
	lastColor     string
	r             int
	g             int
	b             int
	c             color.Color
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

func NewColorPicker(name, color string) ColorPicker {
	clr := stringToColor(color)
	r, g, b, _ := clr.RGBA()
	hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	c := ColorPicker{
		Name:  name,
		Color: hex,
	}
	c.lastColor = color
	c.c = clr
	c.r, c.g, c.b = int(r), int(g), int(b)
	return c
}

func (c *ColorPicker) GetColor() color.Color {
	return color.RGBA{byte(c.r), byte(c.g), byte(c.b), 255}
}

func (c *ColorPicker) GetHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
}

func (c *ColorPicker) UpdateState() {
	// Update hex string from RGB components
	newColor := fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
	if newColor != c.Color {
		c.Color = newColor
		c.DidJustChange = true
		// Update the color.Color value
		clr, err := colorful.Hex(newColor)
		if err != nil {
			panic(err)
		}
		c.c = clr
	} else {
		c.DidJustChange = false
	}
	c.lastColor = c.Color
}

func calcDigits(val float64) int {
	if val < 1 {
		return int(math.Ceil(math.Abs(math.Log10(val))))
	}
	return 0
}

func (s *Sketch) controlWindow(ctx *debugui.Context) {
	ctx.Window("Control Panel", image.Rect(DefaultControlWindowX, DefaultControlWindowY, s.ControlWidth, s.ControlHeight), func(layout debugui.ContainerLayout) {
		ctx.Header("Builtins", true, func() {
			ctx.TreeNode("Random Seed", func() {
				ctx.Text(fmt.Sprintf("Seed: %d", s.RandomSeed))
				ctx.SetGridLayout([]int{40, 40, 40}, nil)
				ctx.Button("Decr").On(func() { s.decrementRandomSeed() })
				ctx.Button("Incr").On(func() { s.incrementRandomSeed() })
				ctx.Button("Rand").On(func() { s.randomizeRandomSeed() })
			})
			ctx.TreeNode("Save Options", func() {
				ctx.SetGridLayout([]int{80, 80}, nil)
				ctx.Button("Save as PNG").On(func() { s.isSavingPNG = true })
				ctx.Button("Save as SVG").On(func() { s.isSavingSVG = true })
				ctx.Button("Save Config").On(func() { s.saveConfig() })
				ctx.Button("Dump State").On(func() { s.DumpState() })
			})
		})
		ctx.Header("Sliders", true, func() {
			ctx.Button("Randomize (unless checked)").On(func() {
				for i := range s.Sliders {
					if !s.Sliders[i].dontRandomize {
						s.Sliders[i].Randomize()
					}
				}
			})
			ctx.SetGridLayout([]int{20, s.SliderTextWidth, -1}, nil)
			for i := range s.Sliders {
				ctx.IDScope(fmt.Sprintf("sliders%d", i), func() {
					ctx.Checkbox(&s.Sliders[i].dontRandomize, "")
					ctx.Text(s.Sliders[i].Name)
					ctx.SliderF(&s.Sliders[i].Val, s.Sliders[i].MinVal, s.Sliders[i].MaxVal, s.Sliders[i].Incr, s.Sliders[i].digits)
				})
			}
		})
		ctx.Header("Checkboxes", true, func() {
			ctx.SetGridLayout(getRowIntSlice(s.CheckboxColumns), nil)
			for i := range s.Toggles {
				ctx.IDScope(fmt.Sprintf("checkboxes%d", i), func() {
					if !s.Toggles[i].IsButton {
						ctx.Checkbox(&s.Toggles[i].Checked, s.Toggles[i].Name)
					}
				})
			}
		})
		ctx.Header("Buttons", true, func() {
			ctx.SetGridLayout(getRowIntSlice(s.ButtonColumns), nil)
			for i := range s.Toggles {
				ctx.IDScope(fmt.Sprintf("buttons%d", i), func() {
					if s.Toggles[i].IsButton {
						ctx.Button(s.Toggles[i].Name).On(func() {
							s.Toggles[i].Checked = !s.Toggles[i].Checked
						})
					}
				})
			}
		})
		ctx.Header("Color Pickers", true, func() {
			for i := range s.ColorPickers {
				ctx.IDScope(fmt.Sprintf("pickers%d", i), func() {
					ctx.Text(s.ColorPickers[i].Name)
					ctx.SetGridLayout([]int{-3, -1}, []int{54})
					ctx.GridCell(func(bounds image.Rectangle) {
						ctx.SetGridLayout([]int{-1, -3}, nil)
						ctx.Text("R")
						ctx.Slider(&s.ColorPickers[i].r, 0, 255, 1)
						ctx.Text("G")
						ctx.Slider(&s.ColorPickers[i].g, 0, 255, 1)
						ctx.Text("B")
						ctx.Slider(&s.ColorPickers[i].b, 0, 255, 1)
					})
					ctx.GridCell(func(bounds image.Rectangle) {
						ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
							scale := ctx.Scale()
							vector.DrawFilledRect(
								screen,
								float32(bounds.Min.X*scale),
								float32(bounds.Min.Y*scale),
								float32(bounds.Dx()*scale),
								float32(bounds.Dy()*scale),
								color.RGBA{byte(s.ColorPickers[i].r), byte(s.ColorPickers[i].g), byte(s.ColorPickers[i].b), 255},
								false)
							txt := s.ColorPickers[i].GetHex()
							op := &text.DrawOptions{}
							op.GeoM.Translate(float64((bounds.Min.X+bounds.Max.X)/2), float64((bounds.Min.Y+bounds.Max.Y)/2))
							op.GeoM.Scale(float64(scale), float64(scale))
							op.PrimaryAlign = text.AlignCenter
							op.SecondaryAlign = text.AlignCenter
							debugui.DrawText(screen, txt, op)
						})
					})
				})
			}
		})
	})
}

func getRowIntSlice(col int) []int {
	result := make([]int, col)
	for i := range result {
		result[i] = -1
	}
	return result
}

func stringToColor(s string) color.Color {
	hexColorRegex := regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	if hexColorRegex.MatchString(s) {
		clr, err := colorful.Hex(s)
		if err != nil {
			panic(err)
		}
		return clr
	}
	clr, ok := colornames.Map[strings.ToLower(s)]
	if !ok {
		panic(fmt.Sprintf("Color %s not found", s))
	}
	return clr
}
