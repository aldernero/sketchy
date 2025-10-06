package sketchy

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"sync"
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
	// Performance tuning constants
	DefaultPreviewDPI = 48.0 // Lower DPI for preview mode
	SaveChannelBuffer = 10   // Buffer size for async save requests
)

type (
	SketchUpdater func(s *Sketch)
	SketchDrawer  func(s *Sketch, ctx *canvas.Context)
)

// SaveRequest represents an async save operation
type SaveRequest struct {
	Filename string
	Format   string // "png" or "svg"
	DPI      float64
}

type Sketch struct {
	Title                     string        `json:"Title"`
	Prefix                    string        `json:"Prefix"`
	SketchWidth               float64       `json:"SketchWidth"`
	SketchHeight              float64       `json:"SketchHeight"`
	ControlWidth              int           `json:"ControlWidth"`
	ControlHeight             int           `json:"ControlHeight"`
	SliderTextWidth           int           `json:"SliderTextWidth"`
	CheckboxColumns           int           `json:"CheckboxColumns"`
	ButtonColumns             int           `json:"ButtonColumns"`
	ControlBackgroundColor    string        `json:"ControlBackgroundColor"`
	ControlOutlineColor       string        `json:"ControlOutlineColor"`
	SketchBackgroundColor     string        `json:"SketchBackgroundColor"`
	SketchOutlineColor        string        `json:"SketchOutlineColor"`
	DisableClearBetweenFrames bool          `json:"DisableClearBetweenFrames"`
	ShowFPS                   bool          `json:"ShowFPS"`
	RasterDPI                 float64       `json:"RasterDPI"`
	PreviewMode               bool          `json:"PreviewMode"` // Use lower DPI for preview
	RandomSeed                int64         `json:"RandomSeed"`
	Sliders                   []Slider      `json:"Sliders"`
	Toggles                   []Toggle      `json:"Toggles"`
	ColorPickers              []ColorPicker `json:"ColorPickers"`
	Updater                   SketchUpdater `json:"-"`
	Drawer                    SketchDrawer  `json:"-"`
	DidControlsChange         bool          `json:"-"`
	DidSlidersChange          bool          `json:"-"`
	DidTogglesChange          bool          `json:"-"`
	DidColorPickersChange     bool          `json:"-"`
	Rand                      gaul.Rng      `json:"-"`
	sliderControlMap          map[string]int
	toggleControlMap          map[string]int
	colorPickerControlMap     map[string]int
	isSavingPNG               bool
	isSavingSVG               bool
	needToClear               bool
	Tick                      int64          `json:"-"`
	SketchCanvas              *canvas.Canvas `json:"-"`
	ui                        debugui.DebugUI
	showDebugUI               bool `json:"-"`

	// Performance optimization fields
	offscreen    *ebiten.Image    `json:"-"`
	cachedRGBA   *image.RGBA      `json:"-"`
	ctx          *canvas.Context  `json:"-"`
	dirty        bool             `json:"-"`
	saveRequests chan SaveRequest `json:"-"`
	saveMutex    sync.Mutex       `json:"-"`
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
	if s.ControlWidth == 0 {
		s.ControlWidth = DefaultControlWindowWidth
	}
	if s.ControlHeight == 0 {
		s.ControlHeight = DefaultControlWindowHeight
	}
	if s.SliderTextWidth == 0 {
		s.SliderTextWidth = DefaultSliderTextWidth
	}
	if s.CheckboxColumns == 0 {
		s.CheckboxColumns = DefaultCheckboxColumns
	}
	if s.ButtonColumns == 0 {
		s.ButtonColumns = DefaultButtonColumns
	}
	s.buildMaps()
	s.Rand = gaul.NewRng(s.RandomSeed)
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	s.needToClear = true
	s.showDebugUI = true
	s.dirty = true // Mark as dirty for initial render

	// Initialize performance optimization fields
	s.offscreen = ebiten.NewImage(int(s.SketchWidth), int(s.SketchHeight))
	s.ctx = canvas.NewContext(s.SketchCanvas)
	s.saveRequests = make(chan SaveRequest, SaveChannelBuffer)

	// Start background save worker
	go s.saveWorker()

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

func (s *Sketch) ColorPicker(name string) color.Color {
	i, ok := s.colorPickerControlMap[name]
	if !ok {
		log.Fatalf("%s not a valid color picker name", name)
	}
	return s.ColorPickers[i].c
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
		s.incrementRandomSeed()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		s.decrementRandomSeed()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySlash) {
		s.randomizeRandomSeed()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		s.DumpState()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySpace) {
		s.showDebugUI = !s.showDebugUI
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
	// check if the color pickers have changed
	for i := range s.ColorPickers {
		// Update the state
		s.ColorPickers[i].UpdateState()
		if s.ColorPickers[i].DidJustChange {
			s.DidColorPickersChange = true
		}
	}
	// check if the controls have changed
	if s.DidSlidersChange || s.DidTogglesChange || s.DidColorPickersChange {
		s.DidControlsChange = true
		s.dirty = true // Mark for re-render when controls change
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
	if s.showDebugUI {
		_, err := s.ui.Update(func(ctx *debugui.Context) error {
			s.controlWindow(ctx)
			return nil
		})
		if err != nil {
			return err
		}
	}
	s.UpdateControls()
	s.Updater(s)
	s.Tick++

	// Mark dirty if this is an animated sketch (Updater does work)
	// This is a simple heuristic - sketches that need per-frame updates should set dirty=true
	// in their Updater function when they actually change content
	return nil
}

func (s *Sketch) Clear() {
	s.needToClear = true
	s.dirty = true
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	// Only re-render if dirty
	if s.dirty {
		s.SketchCanvas.Reset()
		s.ctx = canvas.NewContext(s.SketchCanvas)

		if !s.DisableClearBetweenFrames || s.needToClear {
			s.ctx.SetFillColor(color.Black)
			s.ctx.SetStrokeColor(color.Transparent)
			s.ctx.DrawPath(0, 0, canvas.Rectangle(s.ctx.Width(), s.ctx.Height()))
			s.needToClear = false
		}

		s.Drawer(s, s.ctx)

		// Determine DPI based on preview mode
		dpi := s.RasterDPI
		if s.PreviewMode {
			dpi = DefaultPreviewDPI
		}

		// Rasterize to cached image
		rasterizedImg := rasterizer.Draw(s.SketchCanvas, canvas.DPI(dpi), canvas.DefaultColorSpace)

		// Convert to RGBA and update offscreen buffer
		bounds := rasterizedImg.Bounds()
		if s.cachedRGBA == nil || s.cachedRGBA.Bounds() != bounds {
			s.cachedRGBA = image.NewRGBA(bounds)
		}

		// Convert image to RGBA format
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				s.cachedRGBA.Set(x, y, rasterizedImg.At(x, y))
			}
		}

		// Update offscreen buffer
		s.offscreen.WritePixels(s.cachedRGBA.Pix)
		s.dirty = false
	}

	// Always draw the cached offscreen buffer
	screen.DrawImage(s.offscreen, nil)

	// Handle async save requests
	if s.isSavingPNG {
		fname := s.Prefix + "_" + gaul.GetTimestampString() + ".png"
		select {
		case s.saveRequests <- SaveRequest{Filename: fname, Format: "png", DPI: s.RasterDPI}:
			fmt.Println("Queued PNG save: ", fname)
		default:
			fmt.Println("Save queue full, skipping PNG save")
		}
		s.isSavingPNG = false
	}
	if s.isSavingSVG {
		fname := s.Prefix + "_" + gaul.GetTimestampString() + ".svg"
		select {
		case s.saveRequests <- SaveRequest{Filename: fname, Format: "svg", DPI: s.RasterDPI}:
			fmt.Println("Queued SVG save: ", fname)
		default:
			fmt.Println("Save queue full, skipping SVG save")
		}
		s.isSavingSVG = false
	}

	if s.ShowFPS {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
	}

	if s.showDebugUI {
		s.ui.Draw(screen)
	}

	s.DidControlsChange = false
	s.DidSlidersChange = false
	s.DidTogglesChange = false
	s.DidColorPickersChange = false
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
	for i := range s.ColorPickers {
		fmt.Printf("%s: %s\n", s.ColorPickers[i].Name, s.ColorPickers[i].Color)
	}
	fmt.Println("RandomSeed: ", s.RandomSeed)
}

func (s *Sketch) RandomWidth() float64 {
	return rand.Float64() * s.Width()
}

func (s *Sketch) RandomHeight() float64 {
	return rand.Float64() * s.Height()
}

func (s *Sketch) IsMouseOverControlPanel() bool {
	state, _ := s.ui.Update(func(ctx *debugui.Context) error { return nil })
	return state != 0
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
	s.colorPickerControlMap = make(map[string]int)
	for i := range s.ColorPickers {
		s.ColorPickers[i] = NewColorPicker(s.ColorPickers[i].Name, s.ColorPickers[i].Color)
		s.colorPickerControlMap[s.ColorPickers[i].Name] = i
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

func (s *Sketch) decrementRandomSeed() {
	s.RandomSeed--
	s.Rand.SetSeed(s.RandomSeed)
	s.DidControlsChange = true
	s.dirty = true
	fmt.Println("RandomSeed decremented: ", s.RandomSeed)
}

func (s *Sketch) incrementRandomSeed() {
	s.RandomSeed++
	s.Rand.SetSeed(s.RandomSeed)
	s.DidControlsChange = true
	s.dirty = true
	fmt.Println("RandomSeed incremented: ", s.RandomSeed)
}

func (s *Sketch) randomizeRandomSeed() {
	s.RandomSeed = rand.Int63()
	s.Rand.SetSeed(s.RandomSeed)
	s.DidControlsChange = true
	s.dirty = true
	fmt.Println("RandomSeed changed: ", s.RandomSeed)
}

// MarkDirty marks the sketch for re-rendering (useful for animated sketches)
func (s *Sketch) MarkDirty() {
	s.dirty = true
}

// saveWorker processes save requests in the background
func (s *Sketch) saveWorker() {
	for req := range s.saveRequests {
		s.saveMutex.Lock()
		var err error
		switch req.Format {
		case "png":
			err = renderers.Write(req.Filename, s.SketchCanvas, canvas.DPI(req.DPI))
		case "svg":
			err = renderers.Write(req.Filename, s.SketchCanvas)
		}
		s.saveMutex.Unlock()

		if err != nil {
			fmt.Printf("Error saving %s: %v\n", req.Filename, err)
		} else {
			fmt.Println("Saved ", req.Filename)
		}
	}
}
