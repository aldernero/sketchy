package sketchy

import (
	"encoding/json"
	"log"
	"os"

	"github.com/fogleman/gg"
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

func (g *Sketch) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(g.ControlWidth + g.SketchWidth), int(g.SketchHeight)
}
