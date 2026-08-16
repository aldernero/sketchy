package sketchy

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/aldernero/gaul"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

const (
	DefaultControlWindowWidth  = 330
	DefaultControlWindowHeight = 500
	DefaultControlWindowX      = 25
	DefaultControlWindowY      = 25
	DefaultSliderTextWidth     = 100
	// ControlLabelColumnWidth: name column for sliders, colors, dropdowns (~20 chars at 6px/glyph + debugui padding).
	ControlLabelColumnWidth = 136
	DefaultCheckboxColumns  = 1
	DefaultButtonColumns    = 1
)

// FloatSlider is a continuous control backed by debugui SliderF and a text field for the value.
type FloatSlider struct {
	Folder  string
	Name    string
	textBuf string
	MinVal  float64
	MaxVal  float64
	Val     float64
	Incr    float64
	lastVal float64
	digits  int // precision for SliderF display (from Incr)
	// TextDecimals is fraction digits for the value text when >= 0; when < 0, use digits derived from Incr.
	TextDecimals  int
	textSyncVal   float64
	DidJustChange bool
	textSyncOK    bool
}

// IntSlider is a discrete stepped control backed by debugui Slider and a text field for the value.
type IntSlider struct {
	Folder        string
	Name          string
	textBuf       string
	MinVal        int
	MaxVal        int
	Val           int
	Incr          int
	lastVal       int
	textSyncVal   int
	DidJustChange bool
	textSyncOK    bool
}

type Toggle struct {
	Folder        string
	Name          string
	Pos           gaul.Point
	Width         float64
	Height        float64
	Checked       bool
	IsButton      bool
	DidJustChange bool
	lastVal       bool
}

type ColorPicker struct {
	c             color.Color
	Folder        string
	Name          string
	Color         string
	lastColor     string
	r             int
	g             int
	b             int
	DidJustChange bool
}

// TextBox is a free-form string control: a full-width debugui text field whose
// value is committed on Enter or focus loss. It exists for values no numeric
// control can hold — arbitrary-precision coordinates, identifiers, expressions
// — and it persists in snapshots like any other control.
//
// Val is the committed value and the only thing to read; textBuf is the live
// edit buffer and holds whatever the user has typed so far.
type TextBox struct {
	Folder string
	Name   string
	// Val is the last committed value, after validation.
	Val string
	// Validate normalizes a typed value before it is committed. Returning
	// ok = false rejects the input and reverts the field to Val, the same way
	// an unparseable slider value reverts. A nil Validate accepts anything.
	Validate func(string) (string, bool)
	// Multiline lays the field out on its own row under the label, for values
	// too long to read in the value column.
	Multiline     bool
	DidJustChange bool

	textBuf string
	lastVal string
	// textSyncOK guards the initial sync of textBuf from Val.
	textSyncOK bool
}

func NewTextBox(name, initial string, validate func(string) (string, bool)) TextBox {
	t := TextBox{
		Name:     name,
		Val:      initial,
		Validate: validate,
	}
	t.lastVal = initial
	t.syncTextBufFromVal()
	return t
}

func (t *TextBox) UpdateState() {
	t.DidJustChange = t.Val != t.lastVal
	t.lastVal = t.Val
}

// Label is a read-only status row. Value is called each time the panel is
// drawn, so a Label can report live state that no control owns.
type Label struct {
	Folder string
	Name   string
	Value  func() string
}

// Dropdown is a string option list control.
type Dropdown struct {
	Folder        string
	Name          string
	Options       []string
	Index         int
	DidJustChange bool
	lastIdx       int
}

func (d *Dropdown) Selected() string {
	if len(d.Options) == 0 || d.Index < 0 || d.Index >= len(d.Options) {
		return ""
	}
	return d.Options[d.Index]
}

func (d *Dropdown) UpdateState() {
	if d.Index != d.lastIdx {
		d.DidJustChange = true
	} else {
		d.DidJustChange = false
	}
	d.lastIdx = d.Index
}

func NewFloatSlider(name string, min, max, val, incr float64) FloatSlider {
	return NewFloatSliderWithDecimals(name, min, max, val, incr, -1)
}

// NewFloatSliderWithDecimals is like NewFloatSlider but sets TextDecimals for the value text field.
// Pass textDecimals < 0 to derive fraction digits from incr (same as SliderF precision).
func NewFloatSliderWithDecimals(name string, min, max, val, incr float64, textDecimals int) FloatSlider {
	s := FloatSlider{
		Name:         name,
		MinVal:       min,
		MaxVal:       max,
		Val:          val,
		Incr:         incr,
		Folder:       "",
		TextDecimals: textDecimals,
	}
	s.lastVal = val
	s.CalcDigits()
	return s
}

func (s *FloatSlider) GetPercentage() float64 {
	return gaul.Map(s.MinVal, s.MaxVal, 0, 1, s.Val)
}

func (s *FloatSlider) Randomize() {
	s.Val = gaul.Map(0, 1, s.MinVal, s.MaxVal, rand.Float64())
}

func (s *FloatSlider) StringVal() string {
	return strconv.FormatFloat(s.Val, 'f', s.digits, 64)
}

func (s *FloatSlider) CalcDigits() {
	s.digits = calcDigits(s.Incr)
}

func (s *FloatSlider) UpdateState() {
	if s.Val != s.lastVal {
		s.DidJustChange = true
	} else {
		s.DidJustChange = false
	}
	s.lastVal = s.Val
}

func NewIntSlider(name string, min, max, val, incr int) IntSlider {
	if incr <= 0 {
		incr = 1
	}
	s := IntSlider{
		Name:   name,
		MinVal: min,
		MaxVal: max,
		Val:    val,
		Incr:   incr,
		Folder: "",
	}
	s.lastVal = val
	return s
}

func (s *IntSlider) Randomize() {
	if s.MinVal > s.MaxVal || s.Incr <= 0 {
		return
	}
	n := (s.MaxVal - s.MinVal) / s.Incr
	if n < 0 {
		return
	}
	s.Val = s.MinVal + rand.Intn(n+1)*s.Incr
	if s.Val > s.MaxVal {
		s.Val = s.MaxVal
	}
}

func (s *IntSlider) UpdateState() {
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
	r16, g16, b16, _ := clr.RGBA()
	r8, g8, b8 := int(r16>>8), int(g16>>8), int(b16>>8)
	hex := fmt.Sprintf("#%02X%02X%02X", r8, g8, b8)
	c := ColorPicker{
		Name:   name,
		Color:  hex,
		Folder: "",
	}
	c.lastColor = color
	c.c = clr
	c.r, c.g, c.b = r8, g8, b8
	return c
}

func (c *ColorPicker) GetColor() color.Color {
	return color.RGBA{byte(c.r), byte(c.g), byte(c.b), 255}
}

func (c *ColorPicker) GetHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
}

// syncFromRGB updates Color and c from r,g,b (8-bit).
func (c *ColorPicker) syncFromRGB() {
	c.Color = fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
	cl, err := colorful.Hex(c.Color)
	if err == nil {
		c.c = cl
	}
}

func (c *ColorPicker) UpdateState() {
	newColor := fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
	if newColor != c.Color {
		c.Color = newColor
		c.DidJustChange = true
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
