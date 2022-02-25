package sketchy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	DefaultTitle           = "Sketch"
	DefaultPrefix          = "sketch"
	DefaultBackgroundColor = "#1e1e1e"
	DefaultOutlineColor    = "#ffdb00"
)

type SketchUpdater func(s *Sketch)
type SketchDrawer func(s *Sketch, c *gg.Context)

type Sketch struct {
	Title                     string         `json:"Title"`
	Prefix                    string         `json:"Prefix"`
	SketchWidth               float64        `json:"SketchWidth"`
	SketchHeight              float64        `json:"SketchHeight"`
	ControlWidth              float64        `json:"ControlWidth"`
	ControlBackgroundColor    string         `json:"ControlBackgroundColor"`
	ControlOutlineColor       string         `json:"ControlOutlineColor"`
	SketchBackgroundColor     string         `json:"SketchBackgroundColor"`
	SketchOutlineColor        string         `json:"SketchOutlineColor"`
	DisableClearBetweenFrames bool           `json:"DisableClearBetweenFrames"`
	RandomSeed                int64          `json:"RandomSeed"`
	Controls                  []Slider       `json:"Controls"`
	Updater                   SketchUpdater  `json:"-"`
	Drawer                    SketchDrawer   `json:"-"`
	DidControlsChange         bool           `json:"-"`
	Rand                      Rng            `json:"-"`
	controlMap                map[string]int `json:"-"`
	controlColorConfig        ColorConfig    `json:"-"`
	sketchColorConfig         ColorConfig    `json:"-"`
	isSavingPNG               bool           `json:"-"`
	needToClear               bool           `json:"-"`
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
	if s.Title == "" {
		s.Title = DefaultTitle
	}
	if s.Prefix == "" {
		s.Prefix = DefaultPrefix
	}
	s.buildMaps()
	s.parseColors()
	s.Rand = NewRng(s.RandomSeed)
	W := int(s.ControlWidth + s.SketchWidth)
	H := int(s.SketchHeight)
	ctx := gg.NewContext(W, H)
	s.PlaceControls(s.ControlWidth, s.SketchHeight, ctx)
	s.needToClear = true
	if s.DisableClearBetweenFrames {
		ebiten.SetScreenClearedEveryFrame(false)
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
	controlsChanged := false
	if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		s.isSavingPNG = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyC) {
		s.saveConfig()
	}
	for i := range s.Controls {
		didChange, err := s.Controls[i].CheckAndUpdate()
		if err != nil {
			panic(err)
		}
		if didChange {
			controlsChanged = true
		}
	}
	s.DidControlsChange = controlsChanged
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
	ctx.SetColor(s.controlColorConfig.Background)
	ctx.DrawRectangle(0, 0, s.ControlWidth, s.SketchHeight)
	ctx.Fill()
	ctx.SetColor(s.controlColorConfig.Outline)
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

func (s *Sketch) Clear() {
	s.needToClear = true
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	W := int(s.ControlWidth + s.SketchWidth)
	H := int(s.SketchHeight)
	cc := gg.NewContext(W, H)
	s.DrawControls(cc)
	if !s.DisableClearBetweenFrames || s.needToClear {
		cc.SetColor(s.sketchColorConfig.Background)
		cc.DrawRectangle(s.ControlWidth, 0, s.SketchWidth, s.SketchHeight)
		cc.Fill()
		s.needToClear = false
	}
	cc.SetColor(s.sketchColorConfig.Outline)
	cc.DrawRectangle(s.ControlWidth+2.5, 2.5, s.SketchWidth-5, s.SketchHeight-5)
	cc.Stroke()
	screen.DrawImage(ebiten.NewImageFromImage(cc.Image()), nil)
	ctx := gg.NewContext(int(s.SketchWidth), H)
	ctx.Push()
	s.Drawer(s, ctx)
	ctx.Pop()
	if s.isSavingPNG {
		fname := s.Prefix + "_" + GetTimestampString() + ".png"
		ctx.SavePNG(fname)
		fmt.Println("Saved ", fname)
		s.isSavingPNG = false
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.ControlWidth, 0)
	screen.DrawImage(ebiten.NewImageFromImage(ctx.Image()), op)
}

func (s *Sketch) SketchCoords(x, y float64) Point {
	return Point{X: float64(x) - s.ControlWidth, Y: float64(y)}
}

func (s *Sketch) PointInSketchArea(x, y float64) bool {
	return x > s.ControlWidth && x <= (s.ControlWidth+s.SketchWidth) && y >= 0 && y <= s.SketchHeight
}

func (s *Sketch) buildMaps() {
	s.controlMap = make(map[string]int)
	for i := range s.Controls {
		s.controlMap[s.Controls[i].Name] = i
	}
}

func (s *Sketch) parseColors() {
	s.controlColorConfig.Set(s.ControlBackgroundColor, BackgroundColorType, DefaultBackgroundColor)
	s.controlColorConfig.Set(s.ControlOutlineColor, OutlineColorType, DefaultOutlineColor)
	s.sketchColorConfig.Set(s.SketchBackgroundColor, BackgroundColorType, DefaultBackgroundColor)
	s.sketchColorConfig.Set(s.SketchOutlineColor, OutlineColorType, DefaultOutlineColor)
	for i := range s.Controls {
		s.Controls[i].parseColors()
	}
}

func (s *Sketch) saveConfig() {
	configJson, _ := json.MarshalIndent(s, "", "    ")
	fname := s.Prefix + "_config_" + GetTimestampString() + ".json"
	err := ioutil.WriteFile(fname, configJson, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Saved config ", fname)
}
