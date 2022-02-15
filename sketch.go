package sketchy

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type SketchUpdater func(s *Sketch)
type SketchDrawer func(s *Sketch, c *gg.Context)

type Sketch struct {
	SketchWidth  float64
	SketchHeight float64
	ControlWidth float64
	Controls     []Slider
	Updater      SketchUpdater
	Drawer       SketchDrawer
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

func (s *Sketch) UpdateControls() {
	for i := range s.Controls {
		s.Controls[i].CheckAndUpdate()
	}
}

func (s *Sketch) PlaceControls(w float64, h float64, ctx *gg.Context) {
	for i := range s.Controls {
		s.Controls[i].AutoHeight(ctx)
		s.Controls[i].Width = s.ControlWidth - 2*SliderHPadding
		s.Controls[i].Pos = Point{
			X: SliderHPadding,
			Y: 2*float64(i)*s.Controls[i].Height + s.Controls[i].Height,
		}
	}
}

func (s *Sketch) DrawControls(ctx *gg.Context) {
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
	w := int(s.ControlWidth)
	W := int(s.ControlWidth + s.SketchWidth)
	H := int(s.SketchHeight)
	screen.Fill(color.White)
	cc := gg.NewContext(w, H)
	cc.DrawRectangle(0, 0, s.ControlWidth, s.SketchHeight)
	cc.SetColor(color.White)
	cc.Fill()
	s.DrawControls(cc)
	screen.DrawImage(ebiten.NewImageFromImage(cc.Image()), nil)
	ctx := gg.NewContext(W, H)
	ctx.Push()
	ctx.Translate(s.ControlWidth, 0)
	s.Drawer(s, ctx)
	ctx.Pop()
	screen.DrawImage(ebiten.NewImageFromImage(ctx.Image()), nil)
}
