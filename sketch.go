package sketchy

import (
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

const (
	DefaultTitle              = "Sketch"
	DefaultPrefix             = "sketch"
	DefaultBackgroundColor    = "#1e1e1e"
	DefaultOutlineColor       = "#ffdb00"
	DefaultSketchOutlineColor = ""
)

type SketchUpdater func(s *Sketch)
type SketchDrawer func(s *Sketch, c *gg.Context)

type Sketch struct {
	Title                     string        `json:"Title"`
	Prefix                    string        `json:"Prefix"`
	SketchWidth               float64       `json:"SketchWidth"`
	SketchHeight              float64       `json:"SketchHeight"`
	ControlWidth              float64       `json:"ControlWidth"`
	ControlBackgroundColor    string        `json:"ControlBackgroundColor"`
	ControlOutlineColor       string        `json:"ControlOutlineColor"`
	SketchBackgroundColor     string        `json:"SketchBackgroundColor"`
	SketchOutlineColor        string        `json:"SketchOutlineColor"`
	DisableClearBetweenFrames bool          `json:"DisableClearBetweenFrames"`
	RandomSeed                int64         `json:"RandomSeed"`
	Sliders                   []Slider      `json:"Sliders"`
	Toggles                   []Toggle      `json:"Toggles"`
	Updater                   SketchUpdater `json:"-"`
	Drawer                    SketchDrawer  `json:"-"`
	DidControlsChange         bool          `json:"-"`
	Rand                      Rng           `json:"-"`
	sliderControlMap          map[string]int
	ToggleControlMap          map[string]int `json:"-"`
	controlColorConfig        ColorConfig
	sketchColorConfig         ColorConfig
	isSavingPNG               bool
	isSavingScreen            bool
	needToClear               bool
	Tick                      int64 `json:"-"`
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
	err := os.Setenv("EBITEN_SCREENSHOT_KEY", "escape")
	if err != nil {
		log.Fatal("error while setting ebiten screenshot key: ", err)
	}
}

func (s *Sketch) Slider(name string) float64 {
	i, ok := s.sliderControlMap[name]
	if !ok {
		log.Fatalf("%s not a valid control name", name)
	}
	return s.Sliders[i].Val
}

func (s *Sketch) Toggle(name string) bool {
	i, ok := s.ToggleControlMap[name]
	if !ok {
		log.Fatalf("%s not a valid control name", name)
	}
	return s.Toggles[i].Checked
}

func (s *Sketch) UpdateControls() {
	controlsChanged := false
	if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		s.isSavingPNG = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyQ) {
		s.isSavingScreen = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyC) {
		s.saveConfig()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		s.RandomSeed++
		s.Rand.SetSeed(s.RandomSeed)
		controlsChanged = true
		fmt.Println("RandomSeed incremented: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		s.RandomSeed--
		s.Rand.SetSeed(s.RandomSeed)
		controlsChanged = true
		fmt.Println("RandomSeed decremented: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyNumpadDecimal) {
		s.RandomSeed = rand.Int63()
		s.Rand.SetSeed(s.RandomSeed)
		controlsChanged = true
		fmt.Println("RandomSeed changed: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySpace) {
		s.DumpState()
	}
	for i := range s.Sliders {
		didChange, err := s.Sliders[i].CheckAndUpdate()
		if err != nil {
			panic(err)
		}
		if didChange {
			controlsChanged = true
		}
	}
	for i := range s.Toggles {
		didChange, err := s.Toggles[i].CheckAndUpdate()
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
	var lastRect Rect
	for i := range s.Sliders {
		if s.Sliders[i].Height == 0 {
			s.Sliders[i].Height = SliderHeight
		}
		s.Sliders[i].Width = s.ControlWidth - 2*SliderHPadding
		rect := s.Sliders[i].GetRect(ctx)
		s.Sliders[i].Pos = Point{
			X: SliderHPadding,
			Y: float64(i)*rect.H + s.Sliders[i].Height + SliderVPadding,
		}
		lastRect = s.Sliders[i].GetRect(ctx)
	}
	startY := lastRect.Y + lastRect.H
	for i := range s.Toggles {
		if s.Toggles[i].Height == 0 {
			if s.Toggles[i].IsButton {
				s.Toggles[i].Height = ButtonHeight
			} else {
				s.Toggles[i].Height = ToggleHeight
			}
		}
		s.Toggles[i].Width = s.ControlWidth - 2*ToggleHPadding
		rect := s.Toggles[i].GetRect(ctx)
		s.Toggles[i].Pos = Point{
			X: ToggleHPadding,
			Y: startY + float64(i)*rect.H + s.Toggles[i].Height + ToggleVPadding,
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
	for i := range s.Sliders {
		s.Sliders[i].Draw(ctx)
	}
	for i := range s.Toggles {
		s.Toggles[i].Draw(ctx)
	}
}

func (s *Sketch) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(s.ControlWidth + s.SketchWidth), int(s.SketchHeight)
}

func (s *Sketch) Update() error {
	s.UpdateControls()
	s.Updater(s)
	s.Tick++
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
	if s.sketchColorConfig.Outline != color.Transparent {
		cc.SetColor(s.sketchColorConfig.Outline)
		cc.DrawRectangle(s.ControlWidth+2.5, 2.5, s.SketchWidth-5, s.SketchHeight-5)
		cc.Stroke()
	}
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
	if s.isSavingScreen {
		fname := s.Prefix + "_" + GetTimestampString() + ".png"
		sketchImage := screen.SubImage(s.getSketchImageRect())
		f, err := os.Create(fname)
		if err != nil {
			log.Fatal("error while trying to create screenshot file", err)
		}
		if err := png.Encode(f, sketchImage); err != nil {
			f.Close()
			log.Fatal("error while trying to encode screenshot image", err)
		}
		if err := f.Close(); err != nil {
			log.Fatal("error while trying to close screenshot file", err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingScreen = false
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.ControlWidth, 0)
	screen.DrawImage(ebiten.NewImageFromImage(ctx.Image()), op)
}

func (s *Sketch) SketchCoords(x, y float64) Point {
	return Point{X: x - s.ControlWidth, Y: y}
}

func (s *Sketch) PointInSketchArea(x, y float64) bool {
	return x > s.ControlWidth && x <= (s.ControlWidth+s.SketchWidth) && y >= 0 && y <= s.SketchHeight
}

func (s *Sketch) DumpState() {
	for i := range s.Sliders {
		fmt.Printf("%s: %s\n", s.Sliders[i].Name, s.Sliders[i].StringVal())
	}
	for i := range s.Toggles {
		if !s.Toggles[i].IsButton {
			fmt.Printf("%s: %t\n", s.Toggles[i].Name, s.Toggles[i].Checked)
		}
	}
	fmt.Println("RandomSeed: ", s.RandomSeed)
}

func (s *Sketch) buildMaps() {
	s.sliderControlMap = make(map[string]int)
	for i := range s.Sliders {
		s.sliderControlMap[s.Sliders[i].Name] = i
	}
	s.ToggleControlMap = make(map[string]int)
	for i := range s.Toggles {
		s.ToggleControlMap[s.Toggles[i].Name] = i
	}
}

func (s *Sketch) parseColors() {
	s.controlColorConfig.Set(s.ControlBackgroundColor, BackgroundColorType, DefaultBackgroundColor)
	s.controlColorConfig.Set(s.ControlOutlineColor, OutlineColorType, DefaultOutlineColor)
	s.sketchColorConfig.Set(s.SketchBackgroundColor, BackgroundColorType, DefaultBackgroundColor)
	s.sketchColorConfig.Set(s.SketchOutlineColor, OutlineColorType, DefaultSketchOutlineColor)
	for i := range s.Sliders {
		s.Sliders[i].parseColors()
	}
	for i := range s.Toggles {
		s.Toggles[i].parseColors()
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

func (s *Sketch) getSketchImageRect() image.Rectangle {
	left := int(s.ControlWidth)
	top := 0
	right := left + int(s.SketchWidth)
	bottom := int(s.SketchHeight)
	return image.Rect(left, top, right, bottom)
}
