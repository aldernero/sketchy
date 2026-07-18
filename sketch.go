package sketchy

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aldernero/debugui"
	"github.com/aldernero/gaul"
	"github.com/aldernero/gaul/render"
	"github.com/aldernero/palettedb"
	"github.com/aldernero/sketchy/internal/sketchdb"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	DefaultTitle  = "Sketch"
	DefaultPrefix = "sketch"
	// DefaultDPI is the raster resolution at which one raster pixel matches
	// one logical sketch pixel. RasterDPI/DefaultDPI is the supersampling
	// factor.
	DefaultDPI = 96.0
	// Performance tuning constants
	DefaultPreviewDPI = 48.0 // Lower DPI for preview mode
	SaveChannelBuffer = 10   // Buffer size for async save requests
	scrollWheelSpeed  = 32.0
	// Default ebiten window size is sketch dimensions, clamped to at least this.
	MinWindowWidth  = 640
	MinWindowHeight = 480
)

type (
	SketchUpdater func(s *Sketch)
	SketchDrawer  func(s *Sketch, ctx *render.Context)
)

// SaveRequest is an async save operation (relative path under workDir).
type SaveRequest struct {
	RelPath  string // e.g. saves/png/foo.png
	Format   string // "png" or "svg"
	DPI      float64
	RecordDB bool
}

type Sketch struct {
	Title                  string
	Prefix                 string
	SketchWidth            float64
	SketchHeight           float64
	ControlWidth           int
	ControlHeight          int
	ControlBackgroundColor string
	ControlOutlineColor    string
	// SketchBackgroundColor is unused for window filling; letterbox margins follow the Builtins UI theme.
	// Kept for API compatibility; may be used again if margins become configurable.
	SketchBackgroundColor string
	SketchOutlineColor    string
	// DefaultBackground is the canvas clear color before each draw (image/color, default black).
	DefaultBackground color.Color
	// DefaultForeground is the initial stroke color for the canvas context before Drawer (default white).
	DefaultForeground color.Color
	// DefaultStrokeWidth is the initial stroke width in pixels (default 1).
	DefaultStrokeWidth float64
	// PaletteDBPath locates the palettedb SQLite database for the Builtins
	// palette dropdowns; empty means the palettedb default
	// (~/.config/palettedb/palettedb.db). Set before Init().
	PaletteDBPath string
	// DiscretePalette holds the discrete palette selected in the Builtins
	// panel (default black→white until a palette is loaded).
	DiscretePalette gaul.Gradient
	// SinePalette holds the sine palette selected in the Builtins panel
	// (default rainbow cosine palette until a palette is loaded).
	SinePalette gaul.SinePalette
	// DisableClearBetweenFrames keeps the previous frame's raster under each
	// new frame so strokes accumulate on screen; Clear() wipes to
	// DefaultBackground. Accumulation is display-only: image saves render
	// just the current frame.
	DisableClearBetweenFrames bool
	// DisableFastStroke is a no-op kept for API compatibility; the old
	// tdewolff/canvas FastStroke workaround is gone with the gaul renderer.
	DisableFastStroke bool
	ShowFPS           bool
	// RasterDPI sets raster resolution (default 96, where one canvas pixel
	// matches one logical sketch pixel). The sketch is always displayed at
	// SketchWidth x SketchHeight; higher DPI affects raster/save detail only.
	RasterDPI float64
	// PreviewMode rasterizes at DefaultPreviewDPI and scales up to the sketch
	// size on screen: same layout, lower detail, ~4x faster raster.
	PreviewMode  bool
	RandomSeed   int64
	imageAssets  []ImageAsset
	images       map[string]image.Image
	FloatSliders []FloatSlider
	IntSliders   []IntSlider
	Toggles      []Toggle
	ColorPickers []ColorPicker
	Dropdowns    []Dropdown
	uiPlan       []controlEntry
	uiFolders    uiFolderPlan

	// BuildUI registers controls; set before Init().
	BuildUI func(s *Sketch, ui *UI)

	Updater               SketchUpdater
	Drawer                SketchDrawer
	DidControlsChange     bool
	DidSlidersChange      bool
	DidTogglesChange      bool
	DidColorPickersChange bool
	DidDropdownsChange    bool
	Rand                  gaul.Rng
	floatSliderControlMap map[string]int
	intSliderControlMap   map[string]int
	toggleControlMap      map[string]int
	colorPickerControlMap map[string]int
	dropdownControlMap    map[string]int
	needToClear           bool
	Tick                  int64
	ui                    debugui.DebugUI
	showDebugUI           bool
	uiCaptureState        debugui.InputCapturingState

	offscreen *ebiten.Image
	// rasterBuf is the reused CPU-side raster target for the per-frame
	// display render in [Sketch.Draw].
	rasterBuf *image.RGBA
	// recorder keeps the current frame's draw operations so saves can
	// replay the exact displayed frame into a PNG raster (at any DPI) or an
	// SVG without re-running Drawer.
	recorder     *render.Recorder
	ctx          *render.Context
	dirty        bool
	saveRequests chan SaveRequest
	saveMutex    sync.Mutex

	workDir string
	db      *sketchdb.DB

	viewportW, viewportH int
	scrollX, scrollY     float64

	dlgSaveImageOpen   bool
	dlgSaveImagePrefix string
	dlgSavePNG         bool
	dlgSaveSVG         bool

	dlgSnapshotOpen        bool
	dlgSnapshotName        string
	dlgSnapshotDescription string
	dlgSnapshotPNG         bool
	dlgSnapshotSVG         bool

	dlgLoadOpen       bool
	dlgLoadNames      []string
	dlgLoadSelected   string
	dlgLoadMissing    []string
	dlgLoadPreviewRow *sketchdb.SnapshotRow

	// Indices of Builtins-only ColorPickers (Folder "_builtins"), not in uiPlan.
	builtinColorBGIdx int
	builtinColorFGIdx int

	// Builtins palette dropdowns (palettedb); paletteDB is nil when no palette db was found.
	paletteDB                 *palettedb.DB
	discretePaletteNames      []string
	sinePaletteNames          []string
	builtinDiscretePaletteIdx int
	builtinSinePaletteIdx     int

	colorModalIdx int // >= 0 => editing ColorPickers[idx]

	modalH, modalS, modalV float64
	modalR, modalG, modalB int
	modalHexBuf            string
	modalErr               string

	sliderRangeModalOpen  bool
	sliderRangeModalFloat bool // true = FloatSliders[idx], false = IntSliders[idx]
	sliderRangeModalIdx   int
	sliderRangeModalErr   string
	sliderRangeEditMinF   float64
	sliderRangeEditMaxF   float64
	sliderRangeEditIncrF  float64
	sliderRangeEditMinI   int
	sliderRangeEditMaxI   int
	sliderRangeEditIncrI  int

	// builtinSeedInt mirrors RandomSeed for the Builtins NumberField (debugui uses *int).
	builtinSeedInt int

	// builtinExportScaleIdx mirrors RasterDPI for the Builtins "Export scale"
	// dropdown (index into exportScaleFactors).
	builtinExportScaleIdx int

	// debugUIThemeIndex selects the control-panel style (Builtins dropdown); 0 = themes/dark.json, 1 = themes/light.json.
	debugUIThemeIndex int

	// Primary mouse edge (see refreshPrimaryMouseEdge): avoids relying on inpututil JustPressed tick matching.
	sketchPrimaryMouseDown     bool
	sketchPrimaryMouseJustDown bool
}

// Width is the drawing surface width in pixels (same as SketchWidth).
// Canvas coordinates are pixels: origin top-left, x right, y down.
func (s *Sketch) Width() float64 {
	return s.SketchWidth
}

// Height is the drawing surface height in pixels (same as SketchHeight).
func (s *Sketch) Height() float64 {
	return s.SketchHeight
}

// WindowSize returns outer window dimensions for ebiten: the sketch size in pixels,
// with width and height each at least MinWindowWidth and MinWindowHeight.
func (s *Sketch) WindowSize() (w, h int) {
	w = int(s.SketchWidth)
	h = int(s.SketchHeight)
	if w < MinWindowWidth {
		w = MinWindowWidth
	}
	if h < MinWindowHeight {
		h = MinWindowHeight
	}
	return w, h
}

func (s *Sketch) Init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	s.workDir = wd
	if err := validateImageAssets(s.imageAssets); err != nil {
		log.Fatalf("sketchy: %v", err)
	}
	s.loadImages()

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
	if s.ControlBackgroundColor == "" {
		s.ControlBackgroundColor = "#1e1e1e"
	}
	if s.ControlOutlineColor == "" {
		s.ControlOutlineColor = "#ffdb00"
	}
	if s.DefaultBackground == nil {
		s.DefaultBackground = color.Black
	}
	if s.DefaultForeground == nil {
		s.DefaultForeground = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	if s.DefaultStrokeWidth <= 0 {
		s.DefaultStrokeWidth = 1
	}

	s.FloatSliders = nil
	s.IntSliders = nil
	s.Toggles = nil
	s.ColorPickers = nil
	s.Dropdowns = nil
	s.uiPlan = nil
	if s.BuildUI != nil {
		ui := &UI{s: s}
		s.BuildUI(s, ui)
	}
	s.builtinColorBGIdx = len(s.ColorPickers)
	s.ColorPickers = append(s.ColorPickers, NewColorPicker("Default background", colorToRGBHex(s.DefaultBackground)))
	s.ColorPickers[s.builtinColorBGIdx].Folder = "_builtins"
	s.builtinColorFGIdx = len(s.ColorPickers)
	s.ColorPickers = append(s.ColorPickers, NewColorPicker("Default foreground", colorToRGBHex(s.DefaultForeground)))
	s.ColorPickers[s.builtinColorFGIdx].Folder = "_builtins"
	s.buildMaps()
	s.uiFolders = buildFolderPlan(s.uiPlan)

	s.initPaletteDB()

	dbPath := filepath.Join(s.workDir, "sketch.db")
	if s.db != nil { // Init() may run more than once
		if cerr := s.db.Close(); cerr != nil {
			log.Printf("sketchy: close %s: %v", dbPath, cerr)
		}
		s.db = nil
	}
	if db, derr := sketchdb.Open(dbPath); derr != nil {
		log.Printf("sketchy: could not open %s: %v", dbPath, derr)
	} else {
		s.db = db
		if merr := db.InitMetadata(s.Title, s.workDir); merr != nil {
			log.Printf("sketchy: metadata init: %v", merr)
		}
	}

	s.Rand = gaul.NewRng(s.RandomSeed)
	s.builtinSeedInt = int(s.RandomSeed)
	s.syncExportScaleIdxFromDPI()
	s.recorder = render.NewRecorder(s.SketchWidth, s.SketchHeight)
	s.needToClear = true
	s.showDebugUI = true
	s.dirty = true
	s.colorModalIdx = -1

	s.offscreen = ebiten.NewImage(int(s.SketchWidth), int(s.SketchHeight))
	s.ctx = render.NewContext(s.recorder)
	if s.saveRequests == nil { // keep a single worker across repeated Init()
		s.saveRequests = make(chan SaveRequest, SaveChannelBuffer)
		go s.saveWorker()
	}

	if os.Getenv("EBITEN_SCREENSHOT_KEY") == "" {
		if err := os.Setenv("EBITEN_SCREENSHOT_KEY", "escape"); err != nil {
			log.Fatal("error while setting ebiten screenshot key: ", err)
		}
	}

	s.applyDebugUITheme()
}

// GetFloat returns a float slider value in folder (use "" for root).
func (s *Sketch) GetFloat(folder, name string) float64 {
	k := controlMapKey(folder, name)
	i, ok := s.floatSliderControlMap[k]
	if !ok {
		log.Fatalf("%q is not a float slider", k)
	}
	return s.FloatSliders[i].Val
}

// SetFloat sets a float slider value.
func (s *Sketch) SetFloat(folder, name string, v float64) {
	k := controlMapKey(folder, name)
	i, ok := s.floatSliderControlMap[k]
	if !ok {
		log.Fatalf("%q is not a float slider", k)
	}
	s.FloatSliders[i].Val = v
}

// GetInt returns an int slider value in folder (use "" for root).
func (s *Sketch) GetInt(folder, name string) int {
	k := controlMapKey(folder, name)
	i, ok := s.intSliderControlMap[k]
	if !ok {
		log.Fatalf("%q is not an int slider", k)
	}
	return s.IntSliders[i].Val
}

// SetInt sets an int slider value (clamped to min/max on next UI sync; immediate assign here).
func (s *Sketch) SetInt(folder, name string, v int) {
	k := controlMapKey(folder, name)
	i, ok := s.intSliderControlMap[k]
	if !ok {
		log.Fatalf("%q is not an int slider", k)
	}
	s.IntSliders[i].Val = v
}

// Slider returns a float slider in the root folder. Prefer GetFloat("", name).
func (s *Sketch) Slider(name string) float64 {
	return s.GetFloat("", name)
}

// Int returns an int slider in the root folder. Prefer GetInt("", name).
func (s *Sketch) Int(name string) int {
	return s.GetInt("", name)
}

// GetBool returns checkbox state (or button pulse state).
func (s *Sketch) GetBool(folder, name string) bool {
	k := controlMapKey(folder, name)
	i, ok := s.toggleControlMap[k]
	if !ok {
		log.Fatalf("%q is not a toggle", k)
	}
	return s.Toggles[i].Checked
}

// SetBool sets checkbox state.
func (s *Sketch) SetBool(folder, name string, v bool) {
	k := controlMapKey(folder, name)
	i, ok := s.toggleControlMap[k]
	if !ok {
		log.Fatalf("%q is not a toggle", k)
	}
	s.Toggles[i].Checked = v
	s.Toggles[i].lastVal = v
}

// Toggle returns root-folder checkbox/button state.
func (s *Sketch) Toggle(name string) bool {
	return s.GetBool("", name)
}

// GetColor returns a color picker value.
func (s *Sketch) GetColor(folder, name string) color.Color {
	k := controlMapKey(folder, name)
	i, ok := s.colorPickerControlMap[k]
	if !ok {
		log.Fatalf("%q is not a color picker", k)
	}
	return s.ColorPickers[i].GetColor()
}

// ColorPicker returns root-folder color.
func (s *Sketch) ColorPicker(name string) color.Color {
	return s.GetColor("", name)
}

// GetDropdownIndex returns selected index for a dropdown.
func (s *Sketch) GetDropdownIndex(folder, name string) int {
	k := controlMapKey(folder, name)
	i, ok := s.dropdownControlMap[k]
	if !ok {
		log.Fatalf("%q is not a dropdown", k)
	}
	return s.Dropdowns[i].Index
}

// Dropdown is shorthand for GetDropdownIndex("", name).
func (s *Sketch) Dropdown(name string) int {
	return s.GetDropdownIndex("", name)
}

// SelectedDropdown returns the selected string for a root-folder dropdown.
func (s *Sketch) SelectedDropdown(name string) string {
	k := controlMapKey("", name)
	i, ok := s.dropdownControlMap[k]
	if !ok {
		log.Fatalf("%q is not a dropdown", k)
	}
	return s.Dropdowns[i].Selected()
}

func (s *Sketch) buildMaps() {
	s.floatSliderControlMap = make(map[string]int)
	s.intSliderControlMap = make(map[string]int)
	for i := range s.FloatSliders {
		s.FloatSliders[i].lastVal = s.FloatSliders[i].Val
		s.FloatSliders[i].CalcDigits()
		k := controlMapKey(s.FloatSliders[i].Folder, s.FloatSliders[i].Name)
		if _, dup := s.floatSliderControlMap[k]; dup {
			log.Fatalf("duplicate float slider key %q", k)
		}
		s.floatSliderControlMap[k] = i
	}
	for i := range s.IntSliders {
		s.IntSliders[i].lastVal = s.IntSliders[i].Val
		k := controlMapKey(s.IntSliders[i].Folder, s.IntSliders[i].Name)
		if _, dup := s.intSliderControlMap[k]; dup {
			log.Fatalf("duplicate int slider key %q", k)
		}
		if _, clash := s.floatSliderControlMap[k]; clash {
			log.Fatalf("control name %q is both float and int slider", k)
		}
		s.intSliderControlMap[k] = i
	}
	s.toggleControlMap = make(map[string]int)
	for i := range s.Toggles {
		s.Toggles[i].lastVal = s.Toggles[i].Checked
		k := controlMapKey(s.Toggles[i].Folder, s.Toggles[i].Name)
		if _, dup := s.toggleControlMap[k]; dup {
			log.Fatalf("duplicate toggle key %q", k)
		}
		s.toggleControlMap[k] = i
	}
	s.colorPickerControlMap = make(map[string]int)
	for i := range s.ColorPickers {
		folder := s.ColorPickers[i].Folder
		cp := NewColorPicker(s.ColorPickers[i].Name, s.ColorPickers[i].Color)
		cp.Folder = folder
		s.ColorPickers[i] = cp
		k := controlMapKey(s.ColorPickers[i].Folder, s.ColorPickers[i].Name)
		if _, dup := s.colorPickerControlMap[k]; dup {
			log.Fatalf("duplicate color picker key %q", k)
		}
		s.colorPickerControlMap[k] = i
	}
	s.dropdownControlMap = make(map[string]int)
	for i := range s.Dropdowns {
		s.Dropdowns[i].lastIdx = s.Dropdowns[i].Index
		k := controlMapKey(s.Dropdowns[i].Folder, s.Dropdowns[i].Name)
		if _, dup := s.dropdownControlMap[k]; dup {
			log.Fatalf("duplicate dropdown key %q", k)
		}
		s.dropdownControlMap[k] = i
	}
}

func (s *Sketch) UpdateControls() {
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		s.incrementRandomSeed()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		s.decrementRandomSeed()
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySlash) {
		s.randomizeRandomSeed()
	}
	// Ctrl+Space toggles the panel so plain Space still works in debugui text fields.
	ctrlDown := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if ctrlDown && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.showDebugUI = !s.showDebugUI
	}

	if s.showDebugUI && s.uiCaptureState == 0 {
		_, dy := ebiten.Wheel()
		if dy != 0 {
			sw, sh := s.displaySizeF()
			vw := float64(s.viewportW)
			vh := float64(s.viewportH)
			if sw > vw || sh > vh {
				s.scrollY -= dy * scrollWheelSpeed
				s.clampScroll()
			}
		}
	}

	for i := range s.FloatSliders {
		s.FloatSliders[i].UpdateState()
		if s.FloatSliders[i].DidJustChange {
			s.DidSlidersChange = true
		}
	}
	for i := range s.IntSliders {
		s.IntSliders[i].UpdateState()
		if s.IntSliders[i].DidJustChange {
			s.DidSlidersChange = true
		}
	}
	for i := range s.Toggles {
		s.Toggles[i].UpdateState()
		if s.Toggles[i].DidJustChange {
			s.DidTogglesChange = true
		}
	}
	for i := range s.ColorPickers {
		s.ColorPickers[i].UpdateState()
		if s.ColorPickers[i].DidJustChange {
			switch i {
			case s.builtinColorBGIdx:
				s.DefaultBackground = s.ColorPickers[i].GetColor()
			case s.builtinColorFGIdx:
				s.DefaultForeground = s.ColorPickers[i].GetColor()
			}
			s.DidColorPickersChange = true
		}
	}
	for i := range s.Dropdowns {
		s.Dropdowns[i].UpdateState()
		if s.Dropdowns[i].DidJustChange {
			s.DidDropdownsChange = true
		}
	}
	if s.DidSlidersChange || s.DidTogglesChange || s.DidColorPickersChange || s.DidDropdownsChange {
		s.DidControlsChange = true
		s.dirty = true
	}
}

// viewPad is half the empty margin when the sketch is smaller than the viewport (may be negative).
func (s *Sketch) viewPadX() float64 {
	dx, _ := s.displaySizeF()
	return (float64(s.viewportW) - dx) / 2
}

func (s *Sketch) viewPadY() float64 {
	_, dy := s.displaySizeF()
	return (float64(s.viewportH) - dy) / 2
}

func (s *Sketch) clampScroll() {
	sw, sh := s.displaySizeF()
	vw := float64(s.viewportW)
	vh := float64(s.viewportH)
	padX := s.viewPadX()
	padY := s.viewPadY()
	if sw <= vw {
		s.scrollX = 0
	} else {
		s.scrollX = clampFloat(s.scrollX, padX, (sw-vw)/2)
	}
	if sh <= vh {
		s.scrollY = 0
	} else {
		s.scrollY = clampFloat(s.scrollY, padY, (sh-vh)/2)
	}
}

func clampFloat(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func (s *Sketch) RandomizeSliders() {
	for i := range s.FloatSliders {
		s.FloatSliders[i].Randomize()
	}
	for i := range s.IntSliders {
		s.IntSliders[i].Randomize()
	}
}

func (s *Sketch) RandomizeSlider(name string) {
	s.RandomizeSliderIn("", name)
}

func (s *Sketch) RandomizeSliderIn(folder, name string) {
	k := controlMapKey(folder, name)
	if i, ok := s.floatSliderControlMap[k]; ok {
		s.FloatSliders[i].Randomize()
		return
	}
	if i, ok := s.intSliderControlMap[k]; ok {
		s.IntSliders[i].Randomize()
		return
	}
	log.Fatalf("%q is not a slider", k)
}

func (s *Sketch) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 {
		outsideWidth = int(s.SketchWidth)
	}
	if outsideHeight <= 0 {
		outsideHeight = int(s.SketchHeight)
	}
	s.viewportW, s.viewportH = outsideWidth, outsideHeight
	s.clampScroll()
	return outsideWidth, outsideHeight
}

func (s *Sketch) Update() error {
	s.refreshPrimaryMouseEdge()
	if s.showDebugUI {
		var err error
		s.uiCaptureState, err = s.ui.Update(func(ctx *debugui.Context) error {
			s.controlWindow(ctx)
			if s.colorModalIdx >= 0 {
				s.drawColorModal(ctx)
			}
			if s.sliderRangeModalOpen {
				s.drawSliderRangeModal(ctx)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		s.uiCaptureState = 0
	}
	s.UpdateControls()
	s.Updater(s)
	if ok, _, _ := s.PrimaryPointerPressInSketch(); ok {
		s.MarkDirty()
	}
	s.Tick++
	return nil
}

func (s *Sketch) Clear() {
	s.needToClear = true
	s.dirty = true
}

// letterboxMarginRGBA fills the window behind the letterboxed sketch bitmap. It follows
// the Builtins UI theme (dark vs light) so margins stay visually separate from the canvas
// ([Sketch.DefaultBackground] inside the rasterized sketch area).
func (s *Sketch) letterboxMarginRGBA() color.RGBA {
	switch s.debugUIThemeIndex {
	case 1: // Light (themes/light.json)
		return color.RGBA{R: 0xe8, G: 0xe8, B: 0xe8, A: 0xff}
	default: // Dark (themes/dark.json) and unknown indices
		return color.RGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff}
	}
}

func colorToRGBHex(c color.Color) string {
	if c == nil {
		return "#000000"
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func (s *Sketch) Draw(screen *ebiten.Image) {
	if s.dirty {
		img := s.renderFrame()

		rw, rh := img.Bounds().Dx(), img.Bounds().Dy()
		if s.offscreen == nil || s.offscreen.Bounds().Dx() != rw || s.offscreen.Bounds().Dy() != rh {
			s.offscreen = ebiten.NewImage(rw, rh)
		}
		s.offscreen.WritePixels(img.Pix)
		s.dirty = false
	}

	screen.Fill(s.letterboxMarginRGBA())
	op := &ebiten.DrawImageOptions{}
	// The raster may be smaller (PreviewMode) or larger (RasterDPI > 96) than
	// the logical sketch size; always present at displaySizeF.
	dw, dh := s.displaySizeF()
	rb := s.offscreen.Bounds()
	sx := dw / float64(rb.Dx())
	sy := dh / float64(rb.Dy())
	if sx != 1 || sy != 1 {
		op.Filter = ebiten.FilterLinear
		op.GeoM.Scale(sx, sy)
	}
	op.GeoM.Translate(s.viewPadX()-s.scrollX, s.viewPadY()-s.scrollY)
	screen.DrawImage(s.offscreen, op)

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
	s.DidDropdownsChange = false
}

// renderFrame rebuilds the current frame: it re-records the drawing (for
// saves) and rasterizes it into the reused display buffer, whose
// premultiplied Pix layout feeds ebiten's WritePixels without a per-pixel
// copy. saveWorker replays the recorder on its own goroutine under
// saveMutex; renderFrame holds it while rebuilding so a queued PNG/SVG save
// never observes a half-rebuilt recording.
func (s *Sketch) renderFrame() *image.RGBA {
	s.saveMutex.Lock()
	defer s.saveMutex.Unlock()
	s.recorder.Reset()

	dpi := s.RasterDPI
	if s.PreviewMode {
		dpi = DefaultPreviewDPI
	}
	// The raster may be smaller (PreviewMode) or larger (RasterDPI > 96)
	// than the logical sketch size; the pixel scale keeps drawing
	// coordinates logical either way.
	scale := dpi / DefaultDPI
	w := int(s.SketchWidth*scale + 0.5)
	h := int(s.SketchHeight*scale + 0.5)

	// With DisableClearBetweenFrames, the raster buffer keeps its previous
	// contents (until Clear() requests a wipe) so strokes accumulate
	// across frames; the recording itself only ever holds the current frame.
	clearFrame := !s.DisableClearBetweenFrames || s.needToClear
	if s.rasterBuf == nil || s.rasterBuf.Bounds().Dx() != w || s.rasterBuf.Bounds().Dy() != h {
		s.rasterBuf = image.NewRGBA(image.Rect(0, 0, w, h))
	} else if clearFrame {
		clear(s.rasterBuf.Pix)
	}
	ras := render.NewRasterFromImage(s.rasterBuf)
	ras.SetScale(scale)
	s.ctx = render.NewContext(s.recorder, ras)

	if clearFrame {
		s.ctx.Clear(s.DefaultBackground)
		s.needToClear = false
	}

	s.ctx.SetStrokeColor(s.DefaultForeground)
	s.ctx.SetStrokeWidth(s.DefaultStrokeWidth)
	s.Drawer(s, s.ctx)
	return s.rasterBuf
}

// CanvasCoords maps window coordinates to canvas coordinates. Canvas
// coordinates are logical sketch pixels (origin top-left, y down), so this
// only undoes the viewport padding and scroll offset.
func (s *Sketch) CanvasCoords(x, y float64) gaul.Point {
	sx, sy := s.WindowToSketchPixels(x, y)
	return gaul.Point{X: sx, Y: sy}
}

// SketchCoords maps logical sketch pixel coordinates to canvas coordinates.
// Since the canvas is addressed in logical sketch pixels, this is the
// identity; it remains for compatibility with pre-gaul-render sketches.
func (s *Sketch) SketchCoords(sx, sy float64) gaul.Point {
	return gaul.Point{X: sx, Y: sy}
}

func (s *Sketch) PointInSketchArea(x, y float64) bool {
	dx, dy := s.displaySizeF()
	if dx <= 0 || dy <= 0 {
		return false
	}
	xs, ys := s.WindowToSketchPixels(x, y)
	// Half-open interval matches image pixel indices [0, Dx); same quad as Draw(screen).DrawImage(s.offscreen, …).
	return xs >= 0 && xs < dx && ys >= 0 && ys < dy
}

// WindowToSketchPixels maps game-surface coordinates into logical sketch
// pixels (the sketch is drawn at Translate(viewPad - scroll) after scaling to
// displaySizeF), i.e. the inverse of the Draw path for s.offscreen.
func (s *Sketch) WindowToSketchPixels(wx, wy float64) (sx, sy float64) {
	sx = wx - s.viewPadX() + s.scrollX
	sy = wy - s.viewPadY() + s.scrollY
	return sx, sy
}

// displaySizeF is the on-screen size of the sketch in logical pixels. The
// sketch is always presented at SketchWidth x SketchHeight; RasterDPI and
// PreviewMode change only the resolution of the underlying raster, which
// Draw scales to this size.
func (s *Sketch) displaySizeF() (dx, dy float64) {
	return s.SketchWidth, s.SketchHeight
}

func (s *Sketch) refreshPrimaryMouseEdge() {
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	s.sketchPrimaryMouseJustDown = down && !s.sketchPrimaryMouseDown
	s.sketchPrimaryMouseDown = down
}

// ControlPanelScreenRect returns the screen-space bounds of the control panel window
// (must match [Sketch.controlWindow]).
func (s *Sketch) ControlPanelScreenRect() image.Rectangle {
	return image.Rect(DefaultControlWindowX, DefaultControlWindowY,
		DefaultControlWindowX+s.ControlWidth, DefaultControlWindowY+s.ControlHeight)
}

// PrimaryPointerPressInSketch reports whether the left mouse button or a newly pressed
// touch just went down over the sketch (not on the control panel's default rectangle).
// It returns the window coordinates of that press. For the mouse, Sketch uses an edge
// detector ([Sketch.refreshPrimaryMouseEdge]) instead of [inpututil.IsMouseButtonJustPressed],
// because the latter only holds when the press timestamp matches the current UI tick,
// which can fail depending on platform and frame timing. For touch, use the returned
// coordinates — not [ebiten.CursorPosition], which is unset on many mobile builds.
// If this is true for the current frame, [Sketch.Update] marks the sketch dirty after
// [Sketch.Updater] returns, so the next [Sketch.Draw] re-rasterizes without [Sketch.MarkDirty].
func (s *Sketch) PrimaryPointerPressInSketch() (ok bool, wx, wy float64) {
	for _, tid := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(tid)
		fx, fy := float64(x), float64(y)
		if s.pressInSketchIgnoringPanel(fx, fy) {
			return true, fx, fy
		}
	}
	if s.sketchPrimaryMouseJustDown {
		x, y := ebiten.CursorPosition()
		fx, fy := float64(x), float64(y)
		if s.pressInSketchIgnoringPanel(fx, fy) {
			return true, fx, fy
		}
	}
	return false, 0, 0
}

// PrimaryPressInSketch is like [Sketch.PrimaryPointerPressInSketch] but only reports whether
// a qualifying press occurred.
func (s *Sketch) PrimaryPressInSketch() bool {
	ok, _, _ := s.PrimaryPointerPressInSketch()
	return ok
}

func (s *Sketch) pressInSketchIgnoringPanel(px, py float64) bool {
	if !s.PointInSketchArea(px, py) {
		return false
	}
	if s.showDebugUI && image.Pt(int(px), int(py)).In(s.ControlPanelScreenRect()) {
		return false
	}
	return true
}

func (s *Sketch) CanvasRect() gaul.Rect {
	return gaul.Rect{X: 0, Y: 0, W: s.Width(), H: s.Height()}
}

func (s *Sketch) RandomWidth() float64 {
	return rand.Float64() * s.Width()
}

func (s *Sketch) RandomHeight() float64 {
	return rand.Float64() * s.Height()
}

// IsMouseOverControlPanel reports whether the cursor is over the control panel's
// default screen rectangle or any debug UI hover surface (dialogs, dropdowns, a dragged panel).
func (s *Sketch) IsMouseOverControlPanel() bool {
	if !s.showDebugUI {
		return false
	}
	x, y := ebiten.CursorPosition()
	pt := image.Pt(x, y)
	if pt.In(s.ControlPanelScreenRect()) {
		return true
	}
	return s.uiCaptureState&debugui.InputCapturingStateHover != 0
}

func (s *Sketch) setRandomSeed(v int64) {
	s.RandomSeed = v
	s.builtinSeedInt = int(v)
	s.Rand.SetSeed(s.RandomSeed)
	s.DidControlsChange = true
	s.dirty = true
}

func (s *Sketch) decrementRandomSeed() {
	s.setRandomSeed(s.RandomSeed - 1)
	fmt.Println("RandomSeed decremented: ", s.RandomSeed)
}

func (s *Sketch) incrementRandomSeed() {
	s.setRandomSeed(s.RandomSeed + 1)
	fmt.Println("RandomSeed incremented: ", s.RandomSeed)
}

func (s *Sketch) randomizeRandomSeed() {
	s.setRandomSeed(rand.Int63())
	fmt.Println("RandomSeed changed: ", s.RandomSeed)
}

// MarkDirty schedules a full sketch raster pass on the next Draw.
// Sketchy also sets dirty when a primary pointer press lands in the sketch
// (same rules as [Sketch.PrimaryPointerPressInSketch]). Call MarkDirty when
// the picture changes without such a press (keyboard, timers, other mouse
// buttons, or state updates that bypass control-driven invalidation).
func (s *Sketch) MarkDirty() {
	s.dirty = true
}

func (s *Sketch) EnqueueSave(relPath, format string, dpi float64, recordDB bool) {
	select {
	case s.saveRequests <- SaveRequest{RelPath: relPath, Format: format, DPI: dpi, RecordDB: recordDB}:
		fmt.Println("Queued save:", relPath)
	default:
		fmt.Println("Save queue full, skipping save")
	}
}

func (s *Sketch) saveWorker() {
	for req := range s.saveRequests {
		full := filepath.Join(s.workDir, filepath.FromSlash(req.RelPath))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			fmt.Printf("Error mkdir %s: %v\n", filepath.Dir(full), err)
			continue
		}
		var err error
		switch req.Format {
		case "png":
			err = s.renderPNGToFile(full, req.DPI)
		case "svg":
			err = s.renderSVGToFile(full)
		default:
			err = fmt.Errorf("unknown format %q", req.Format)
		}

		if err != nil {
			fmt.Printf("Error saving %s: %v\n", full, err)
			continue
		}
		fmt.Println("Saved ", full)
		if req.RecordDB && s.db != nil {
			if _, err := s.db.InsertSave(req.RelPath, req.Format); err != nil {
				fmt.Printf("sketch.db insert save: %v\n", err)
			}
		}
	}
}

func (s *Sketch) dbListSnapshots() []string {
	if s.db == nil {
		return nil
	}
	names, err := s.db.ListSnapshotNames()
	if err != nil {
		fmt.Printf("list snapshots: %v\n", err)
		return nil
	}
	return names
}

func (s *Sketch) dbGetSnapshot(name string) *sketchdb.SnapshotRow {
	if s.db == nil {
		return nil
	}
	row, err := s.db.GetSnapshotByName(name)
	if err != nil {
		fmt.Printf("get snapshot: %v\n", err)
		return nil
	}
	return row
}

func (s *Sketch) dbInsertSnapshot(name, description, controlJSON, builtinJSON string, pngID, svgID *int64) error {
	if s.db == nil {
		return fmt.Errorf("no database")
	}
	return s.db.InsertSnapshot(name, description, controlJSON, builtinJSON, pngID, svgID)
}
