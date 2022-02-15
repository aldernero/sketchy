package sketchy

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type Sketch struct {
	SketchWidth  float64
	SketchHeight float64
	ControlWidth float64
	Controls     []Slider
}

func NewSketchFromFile(fname string) (Sketch, error) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	var sketch Sketch
	parser := json.NewDecoder(jsonFile)
	if err = parser.Decode(&sketch); err != nil {
		log.Fatal(err)
	}
	return sketch, nil
}

func (s *Sketch) UpdateControls() {
	for _, c := range s.Controls {
		c.CheckAndUpdate()
	}
}

func (s *Sketch) PlaceControls(w float64, h float64, ctx *gg.Context) {
	for i, c := range s.Controls {
		c.AutoHeight(ctx)
		c.Width = s.ControlWidth - 2*sliderHPadding
		c.Pos = Point{
			X: sliderHPadding,
			Y: 2*float64(i)*c.Height + c.Height,
		}
	}
}

func (s *Sketch) DrawControls(ctx *gg.Context) {
	for _, c := range s.Controls {
		c.Draw(ctx)
	}
}

func (s *Sketch) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(s.ControlWidth + s.SketchWidth), int(s.SketchHeight)
}

func (s *Sketch) Update() error {
	s.UpdateControls()
	// Custom update code goes here
	return nil
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	// Control context
	cc := gg.NewContext(int(s.ControlWidth), int(s.SketchHeight))
	cc.DrawRectangle(0, 0, s.ControlWidth, s.SketchHeight)
	cc.SetColor(color.White)
	cc.Fill()
	s.DrawControls(cc)
	// Custom draw code goes here
}
