package sketchy

import (
	"encoding/json"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	DefaultTitle              = "Sketch"
	DefaultPrefix             = "sketch"
	DefaultBackgroundColor    = "#1e1e1e"
	DefaultOutlineColor       = "#ffdb00"
	DefaultSketchOutlineColor = ""
	ControlAreaMargin         = 1.0
	MmPerPx                   = 0.26458333
	DefaultDPI                = 96.0
)

type SketchUpdater func(s *Sketch)
type SketchDrawer func(s *Sketch, ctx *canvas.Context)

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
	RasterDPI                 float64       `json:"RasterDPI"`
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
	isSavingSVG               bool
	isSavingScreen            bool
	needToClear               bool
	Tick                      int64              `json:"-"`
	ControlCanvas             *canvas.Canvas     `json:"-"`
	SketchCanvas              *canvas.Canvas     `json:"-"`
	FontFamily                *canvas.FontFamily `json:"-"`
	FontFace                  *canvas.FontFace   `json:"-"`
}

func (s *Sketch) Width() float64 {
	return s.SketchWidth * MmPerPx
}

func (s *Sketch) Height() float64 {
	return s.SketchHeight * MmPerPx
}

func (s *Sketch) FullWidth() float64 {
	return (s.ControlWidth + s.SketchWidth) * MmPerPx
}

func (s *Sketch) controlWidth() float64 {
	return s.ControlWidth * MmPerPx
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
	if s.RasterDPI <= 0 {
		s.RasterDPI = DefaultDPI
	}
	if s.RandomSeed == 0 {
		s.RandomSeed = time.Now().UnixNano()
	}
	s.FontFamily = canvas.NewFontFamily("DejaVu Sans")
	if err := s.FontFamily.LoadLocalFont("DejaVuSans", canvas.FontRegular); err != nil {
		panic(err)
	}
	s.buildMaps()
	s.parseColors()
	s.FontFace = s.FontFamily.Face(14.0, color.White, canvas.FontRegular, canvas.FontNormal)
	s.Rand = NewRng(s.RandomSeed)
	s.ControlCanvas = canvas.New(s.controlWidth(), s.Height())
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	ctx := canvas.NewContext(s.ControlCanvas)
	s.PlaceControls(s.controlWidth(), s.Height(), ctx)
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
	if inpututil.IsKeyJustReleased(ebiten.KeyP) {
		s.isSavingPNG = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		s.isSavingSVG = true
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
	if inpututil.IsKeyJustReleased(ebiten.KeySlash) {
		s.RandomSeed = rand.Int63()
		s.Rand.SetSeed(s.RandomSeed)
		controlsChanged = true
		fmt.Println("RandomSeed changed: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySpace) {
		s.DumpState()
	}
	for i := range s.Sliders {
		didChange, err := s.Sliders[i].CheckAndUpdate(s.ControlCanvas)
		if err != nil {
			panic(err)
		}
		if didChange {
			controlsChanged = true
		}
	}
	for i := range s.Toggles {
		didChange, err := s.Toggles[i].CheckAndUpdate(s.ControlCanvas)
		if err != nil {
			panic(err)
		}
		if didChange {
			controlsChanged = true
		}
	}
	s.DidControlsChange = controlsChanged
}

func (s *Sketch) RandomizeSliders() {
	for i := range s.Sliders {
		s.Sliders[i].Randomize()
	}
}

func (s *Sketch) RandomizeSlider(name string) {
	i, ok := s.sliderControlMap[name]
	if !ok {
		log.Fatalf("%s not a valid control name", name)
	}
	s.Sliders[i].Randomize()
}

func (s *Sketch) PlaceControls(_ float64, _ float64, ctx *canvas.Context) {
	var lastRect Rect
	for i := range s.Sliders {
		if s.Sliders[i].Height == 0 {
			s.Sliders[i].Height = SliderHeight
		}
		s.Sliders[i].Width = ctx.Width() - 2*SliderHPadding
		rect := s.Sliders[i].GetRect()
		s.Sliders[i].Pos = Point{
			X: SliderHPadding,
			Y: ctx.Height() - (float64(i)*rect.H + s.Sliders[i].Height + 2*SliderVPadding) - 2*SliderVPadding,
		}
		lastRect = s.Sliders[i].GetRect()
	}
	startY := lastRect.Y - 2*lastRect.H - 2*SliderVPadding
	for i := range s.Toggles {
		if s.Toggles[i].Height == 0 {
			if s.Toggles[i].IsButton {
				s.Toggles[i].Height = ButtonHeight
			} else {
				s.Toggles[i].Height = ToggleHeight
			}
		}
		s.Toggles[i].Width = s.controlWidth() - 2*ToggleHPadding
		rect := s.Toggles[i].GetRect()
		s.Toggles[i].Pos = Point{
			X: ToggleHPadding,
			Y: startY + float64(i)*rect.H + s.Toggles[i].Height + ToggleVPadding,
		}
	}
}

func (s *Sketch) DrawControls(ctx *canvas.Context) {
	ctx.SetFillColor(s.controlColorConfig.Background)
	ctx.SetStrokeColor(s.controlColorConfig.Background)
	ctx.DrawPath(0, 0, canvas.Rectangle(s.controlWidth(), s.Height()))
	ctx.SetStrokeColor(s.controlColorConfig.Outline)
	ctx.DrawPath(
		0.5*ControlAreaMargin,
		0.5*ControlAreaMargin,
		canvas.Rectangle(ctx.Width()-ControlAreaMargin, ctx.Height()-ControlAreaMargin),
	)
	ctx.Stroke()
	for i := range s.Sliders {
		s.Sliders[i].Draw(ctx)
	}
	for i := range s.Toggles {
		s.Toggles[i].Draw(ctx)
	}
}

func (s *Sketch) Layout(
	int,
	int,
) (int, int) {
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
	s.ControlCanvas.Reset()
	s.SketchCanvas.Reset()
	cc := canvas.NewContext(s.ControlCanvas)
	cc.SetStrokeWidth(0.3)
	s.DrawControls(cc)
	img := rasterizer.Draw(s.ControlCanvas, canvas.DefaultResolution, canvas.DefaultColorSpace)
	screen.DrawImage(ebiten.NewImageFromImage(img), nil)
	ctx := canvas.NewContext(s.SketchCanvas)
	if !s.DisableClearBetweenFrames || s.needToClear {
		ctx.SetFillColor(s.sketchColorConfig.Background)
		ctx.SetStrokeColor(color.Transparent)
		ctx.DrawPath(0, 0, canvas.Rectangle(ctx.Width(), ctx.Height()))
		ctx.Close()
		s.needToClear = false
	}
	s.Drawer(s, ctx)
	if s.isSavingPNG {
		fname := s.Prefix + "_" + GetTimestampString() + ".png"
		err := renderers.Write(fname, s.SketchCanvas, canvas.DPI(s.RasterDPI))
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingPNG = false
	}
	if s.isSavingSVG {
		fname := s.Prefix + "_" + GetTimestampString() + ".svg"
		err := renderers.Write(fname, s.SketchCanvas)
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingSVG = false
	}
	if s.isSavingScreen {
		fname := s.Prefix + "_" + GetTimestampString() + ".png"
		sketchImage := screen.SubImage(s.getSketchImageRect())
		f, err := os.Create(fname)
		if err != nil {
			log.Fatal("error while trying to create screenshot file", err)
		}
		if err := png.Encode(f, sketchImage); err != nil {
			err := f.Close()
			if err != nil {
				panic(err)
			}
			log.Fatal("error while trying to encode screenshot image", err)
		}
		if err := f.Close(); err != nil {
			log.Fatal("error while trying to close screenshot file", err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingScreen = false
	}
	img = rasterizer.Draw(s.SketchCanvas, canvas.DefaultResolution, canvas.DefaultColorSpace)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.ControlWidth, 0)
	screen.DrawImage(ebiten.NewImageFromImage(img), op)
}

// Converts window coordinates (pixels, upper left origin) to canvas coordinates (mm, lower left origin)
func (s *Sketch) CanvasCoords(x, y float64) Point {
	return Point{X: MmPerPx * (x - s.ControlWidth), Y: MmPerPx * (s.SketchHeight - y)}
}

// Converts window coordinates to sketch coordinates
func (s *Sketch) SketchCoords(x, y float64) Point {
	return Point{X: MmPerPx * (x - s.ControlWidth), Y: MmPerPx * (s.SketchHeight - y)}
}

// Coordinates are in pixels, useful when checkin if mouse clicks are in the sketch area
func (s *Sketch) PointInSketchArea(x, y float64) bool {
	return x > s.ControlWidth && x <= s.ControlWidth+s.SketchWidth && y >= 0 && y <= s.SketchHeight
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

func (s *Sketch) RandomWidth() float64 {
	return rand.Float64() * s.Width()
}

func (s *Sketch) RandomHeight() float64 {
	return rand.Float64() * s.Height()
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
		s.Sliders[i].SetFont(s.FontFamily)
	}
	for i := range s.Toggles {
		s.Toggles[i].parseColors()
		s.Toggles[i].SetFont(s.FontFamily)
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
