package sketchy

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	DefaultBackgroundColor = "#1e1e1e"
	DefaultOutlineColor    = "#ffdb00"
	SliderBackgroundColor  = "#1e1e1e"
	SliderOutlineColor     = "#ffdb00"
	SliderFillColor        = "#ffdb00"
	SliderTextColor        = "#ffffff"
)

type SketchUpdater func(s *Sketch)
type SketchDrawer func(s *Sketch, c *gg.Context)

type Sketch struct {
	SketchWidth            float64
	SketchHeight           float64
	ControlWidth           float64
	ControlBackgroundColor color.Color
	ControlOutlineColor    color.Color
	SketchBackgroundColor  color.Color
	SketchOutlineColor     color.Color
	Controls               []Slider
	Updater                SketchUpdater
	Drawer                 SketchDrawer
	controlMap             map[string]int
}

func NewSketchFromFile(fname string) (*Sketch, error) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	var sketch Sketch
	parser := json.NewDecoder(jsonFile)
	if err = parser.Decode(&sketch); err != nil {
		log.Fatal(err)
	}
	return &sketch, nil
}

func (s *Sketch) Init() {
	s.buildMap()
	ctx := gg.NewContext(int(s.ControlWidth), int(s.SketchHeight))
	s.PlaceControls(s.ControlWidth, s.SketchHeight, ctx)
	for i := range s.Controls {
		if s.Controls[i].BackgroundColor == nil {
			c, err := colorful.Hex(SliderBackgroundColor)
			if err != nil {
				panic(err)
			}
			s.Controls[i].BackgroundColor = c
		}
		if s.Controls[i].OutlineColor == nil {
			c, err := colorful.Hex(SliderOutlineColor)
			if err != nil {
				panic(err)
			}
			s.Controls[i].OutlineColor = c
		}
		if s.Controls[i].FillColor == nil {
			c, err := colorful.Hex(SliderFillColor)
			if err != nil {
				panic(err)
			}
			s.Controls[i].FillColor = c
		}
		if s.Controls[i].TextColor == nil {
			c, err := colorful.Hex(SliderTextColor)
			if err != nil {
				panic(err)
			}
			s.Controls[i].TextColor = c
		}
	}
	if s.ControlBackgroundColor == nil {
		c, err := colorful.Hex(DefaultBackgroundColor)
		if err != nil {
			panic(err)
		}
		s.ControlBackgroundColor = c
	}
	if s.ControlOutlineColor == nil {
		c, err := colorful.Hex(DefaultOutlineColor)
		if err != nil {
			panic(err)
		}
		s.ControlOutlineColor = c
	}
	if s.SketchBackgroundColor == nil {
		c, err := colorful.Hex(DefaultBackgroundColor)
		if err != nil {
			panic(err)
		}
		s.SketchBackgroundColor = c
	}
	if s.SketchOutlineColor == nil {
		c, err := colorful.Hex(DefaultOutlineColor)
		if err != nil {
			panic(err)
		}
		s.SketchOutlineColor = c
	}
}

func (s *Sketch) Var(name string) float64 {
	i, ok := s.controlMap[name]
	if !ok {
		log.Fatalf("%s not a valid control name", name)
	}
	return s.Controls[i].Val
}

func (s *Sketch) UpdateControls() {
	for i := range s.Controls {
		s.Controls[i].CheckAndUpdate()
	}
}

func (s *Sketch) PlaceControls(w float64, h float64, ctx *gg.Context) {
	for i := range s.Controls {
		if s.Controls[i].Height == 0 {
			s.Controls[i].Height = SliderHeight
		}
		s.Controls[i].Width = s.ControlWidth - 2*SliderHPadding
		rect := s.Controls[i].GetRect(ctx)
		s.Controls[i].Pos = Point{
			X: SliderHPadding,
			Y: float64(i)*rect.H + s.Controls[i].Height + SliderVPadding,
		}
	}
}

func (s *Sketch) DrawControls(ctx *gg.Context) {
	ctx.SetColor(s.ControlBackgroundColor)
	ctx.DrawRectangle(0, 0, s.ControlWidth, s.SketchHeight)
	ctx.Fill()
	ctx.SetColor(s.ControlOutlineColor)
	ctx.DrawRectangle(2.5, 2.5, s.ControlWidth-5, s.SketchHeight-5)
	ctx.Stroke()
	for i := range s.Controls {
		s.Controls[i].Draw(ctx)
	}
}

func (s *Sketch) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(s.ControlWidth + s.SketchWidth), int(s.SketchHeight)
}

func (s *Sketch) Update() error {
	s.UpdateControls()
	s.Updater(s)
	return nil
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	W := int(s.ControlWidth + s.SketchWidth)
	H := int(s.SketchHeight)
	//screen.Fill(color.White)
	cc := gg.NewContext(W, H)
	cc.DrawRectangle(0, 0, s.ControlWidth, s.SketchHeight)
	cc.SetColor(color.White)
	cc.Fill()
	s.DrawControls(cc)
	cc.SetColor(s.SketchBackgroundColor)
	cc.DrawRectangle(s.ControlWidth, 0, s.SketchWidth, s.SketchHeight)
	cc.Fill()
	cc.SetColor(s.SketchOutlineColor)
	cc.DrawRectangle(s.ControlWidth+2.5, 2.5, s.SketchWidth-5, s.SketchHeight-5)
	cc.Stroke()
	screen.DrawImage(ebiten.NewImageFromImage(cc.Image()), nil)
	ctx := gg.NewContext(W, H)
	ctx.Push()
	ctx.Translate(s.ControlWidth, 0)
	s.Drawer(s, ctx)
	ctx.Pop()
	screen.DrawImage(ebiten.NewImageFromImage(ctx.Image()), nil)
}

func (s *Sketch) buildMap() {
	s.controlMap = make(map[string]int)
	for i := range s.Controls {
		s.controlMap[s.Controls[i].Name] = i
	}
}
