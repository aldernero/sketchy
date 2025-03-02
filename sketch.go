package sketchy

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aldernero/gaul"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

const (
	DefaultTitle  = "Sketch"
	DefaultPrefix = "sketch"
	MmPerPx       = 0.26458333
	DefaultDPI    = 96.0
)

type (
	SketchUpdater func(s *Sketch)
	SketchDrawer  func(s *Sketch, ctx *canvas.Context)
)

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
	ShowFPS                   bool          `json:"ShowFPS"`
	RasterDPI                 float64       `json:"RasterDPI"`
	RandomSeed                int64         `json:"RandomSeed"`
	Sliders                   []Slider      `json:"Sliders"`
	Toggles                   []Toggle      `json:"Toggles"`
	Updater                   SketchUpdater `json:"-"`
	Drawer                    SketchDrawer  `json:"-"`
	DidControlsChange         bool          `json:"-"`
	DidSlidersChange          bool          `json:"-"`
	DidTogglesChange          bool          `json:"-"`
	Rand                      gaul.Rng      `json:"-"`
	sliderControlMap          map[string]int
	toggleControlMap          map[string]int
	controlColorConfig        gaul.ColorConfig
	sketchColorConfig         gaul.ColorConfig
	isSavingPNG               bool
	isSavingSVG               bool
	needToClear               bool
	Tick                      int64              `json:"-"`
	SketchCanvas              *canvas.Canvas     `json:"-"`
	FontFamily                *canvas.FontFamily `json:"-"`
	FontFace                  *canvas.FontFace   `json:"-"`
	ui                        *debugui.DebugUI
}

func (s *Sketch) Width() float64 {
	return s.SketchWidth * MmPerPx
}

func (s *Sketch) Height() float64 {
	return s.SketchHeight * MmPerPx
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
	s.ui = debugui.New()
	s.FontFamily = canvas.NewFontFamily("dejavu")
	if err := s.FontFamily.LoadSystemFont("DejaVu Sans", canvas.FontRegular); err != nil {
		panic(err)
	}
	s.buildMaps()
	s.FontFace = s.FontFamily.Face(14.0, color.White, canvas.FontRegular, canvas.FontNormal)
	s.Rand = gaul.NewRng(s.RandomSeed)
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
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
	i, ok := s.toggleControlMap[name]
	if !ok {
		log.Fatalf("%s not a valid control name", name)
	}
	return s.Toggles[i].Checked
}

func (s *Sketch) UpdateControls() {
	if inpututil.IsKeyJustReleased(ebiten.KeyP) {
		s.isSavingPNG = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		s.isSavingSVG = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyC) {
		s.saveConfig()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		s.RandomSeed++
		s.Rand.SetSeed(s.RandomSeed)
		s.DidControlsChange = true
		fmt.Println("RandomSeed incremented: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		s.RandomSeed--
		s.Rand.SetSeed(s.RandomSeed)
		s.DidControlsChange = true
		fmt.Println("RandomSeed decremented: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySlash) {
		s.RandomSeed = rand.Int63()
		s.Rand.SetSeed(s.RandomSeed)
		s.DidControlsChange = true
		fmt.Println("RandomSeed changed: ", s.RandomSeed)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySpace) {
		s.DumpState()
	}
	// check if the values of the sliders have changed
	for i := range s.Sliders {
		s.Sliders[i].UpdateState()
		if s.Sliders[i].DidJustChange {
			s.DidSlidersChange = true
		}
	}
	// check if the values of the toggles have changed
	for i := range s.Toggles {
		s.Toggles[i].UpdateState()
		if s.Toggles[i].DidJustChange {
			s.DidTogglesChange = true
		}
	}
	// check if the controls have changed
	if s.DidSlidersChange || s.DidTogglesChange {
		s.DidControlsChange = true
	}
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

func (s *Sketch) Layout(
	int,
	int,
) (int, int) {
	return int(s.SketchWidth), int(s.SketchHeight)
}

func (s *Sketch) Update() error {
	s.UpdateControls()
	s.Updater(s)
	s.ui.Update(func(ctx *debugui.Context) {
		s.controlWindow(ctx)
	})
	s.Tick++
	return nil
}

func (s *Sketch) Clear() {
	s.needToClear = true
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	s.SketchCanvas.Reset()
	ctx := canvas.NewContext(s.SketchCanvas)
	if !s.DisableClearBetweenFrames || s.needToClear {
		ctx.SetFillColor(color.Black)
		ctx.SetStrokeColor(color.Transparent)
		ctx.DrawPath(0, 0, canvas.Rectangle(ctx.Width(), ctx.Height()))
		ctx.Close()
		s.needToClear = false
	}
	s.Drawer(s, ctx)
	if s.ShowFPS {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
	}
	if s.isSavingPNG {
		fname := s.Prefix + "_" + gaul.GetTimestampString() + ".png"
		err := renderers.Write(fname, s.SketchCanvas, canvas.DPI(s.RasterDPI))
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingPNG = false
	}
	if s.isSavingSVG {
		fname := s.Prefix + "_" + gaul.GetTimestampString() + ".svg"
		err := renderers.Write(fname, s.SketchCanvas)
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved ", fname)
		s.isSavingSVG = false
	}
	img := rasterizer.Draw(s.SketchCanvas, canvas.DefaultResolution, canvas.DefaultColorSpace)
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(ebiten.NewImageFromImage(img), op)
	s.ui.Draw(screen)
	s.DidControlsChange = false
	s.DidSlidersChange = false
	s.DidTogglesChange = false
}

// CanvasCoords converts window coordinates (pixels, upper left origin) to canvas coordinates (mm, lower left origin)
func (s *Sketch) CanvasCoords(x, y float64) gaul.Point {
	return gaul.Point{X: MmPerPx * x, Y: MmPerPx * (s.SketchHeight - y)}
}

// SketchCoords converts canvas coordinates (mm, lower left origin) to sketch coordinates (pixels, upper left origin)
// this ignores the control area
func (s *Sketch) SketchCoords(x, y float64) gaul.Point {
	return gaul.Point{X: x / MmPerPx, Y: s.SketchHeight - y/MmPerPx}
}

// PointInSketchArea calculates coordinates in pixels, useful when checkin if mouse clicks are in the sketch area
func (s *Sketch) PointInSketchArea(x, y float64) bool {
	return x > 0 && x <= s.SketchWidth && y >= 0 && y <= s.SketchHeight
}

func (s *Sketch) CanvasRect() gaul.Rect {
	return gaul.Rect{X: 0, Y: 0, W: s.Width(), H: s.Height()}
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
		s.Sliders[i].lastVal = s.Sliders[i].Val
		s.Sliders[i].CalcDigits()
		s.sliderControlMap[s.Sliders[i].Name] = i
	}
	s.toggleControlMap = make(map[string]int)
	for i := range s.Toggles {
		s.Toggles[i].lastVal = s.Toggles[i].Checked
		s.toggleControlMap[s.Toggles[i].Name] = i
	}
}

func (s *Sketch) saveConfig() {
	configJson, _ := json.MarshalIndent(s, "", "    ")
	fname := s.Prefix + "_config_" + gaul.GetTimestampString() + ".json"
	err := os.WriteFile(fname, configJson, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Saved config ", fname)
}

func (s *Sketch) getSketchImageRect() image.Rectangle {
	right := int(s.SketchWidth)
	bottom := int(s.SketchHeight)
	return image.Rect(0, 0, right, bottom)
}

func (s *Sketch) controlWindow(ctx *debugui.Context) {
	ctx.Window("Controls", image.Rect(DefaultControlWindowX, DefaultControlWindowY, DefaultControlWindowWidth, DefaultControlWindowHeight), func(res debugui.Response, layout debugui.Layout) {
		// window info
		if ctx.Header("Sliders", true) != 0 {
			for i := range s.Sliders {
				ctx.Label(s.Sliders[i].Name)
				ctx.Slider(&s.Sliders[i].Val, s.Sliders[i].MinVal, s.Sliders[i].MaxVal, s.Sliders[i].Incr, s.Sliders[i].digits)
			}
		}
		if ctx.Header("Toggles", true) != 0 {
			for i := range s.Toggles {
				if s.Toggles[i].IsButton {
					if ctx.Button(s.Toggles[i].Name) != 0 {
						s.Toggles[i].Checked = !s.Toggles[i].Checked
					}
				} else {
					ctx.Checkbox(s.Toggles[i].Name, &s.Toggles[i].Checked)
				}
			}
		}
	})
}
