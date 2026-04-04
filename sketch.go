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

	"github.com/aldernero/sketchy/internal/sketchdb"
	"github.com/aldernero/debugui"
	"github.com/aldernero/gaul"
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
	scrollWheelSpeed  = 32.0
	// Default ebiten window size is sketch dimensions, clamped to at least this.
	MinWindowWidth  = 640
	MinWindowHeight = 480
)

type (
	SketchUpdater func(s *Sketch)
	SketchDrawer  func(s *Sketch, ctx *canvas.Context)
)

// SaveRequest is an async save operation (relative path under workDir).
type SaveRequest struct {
	RelPath  string // e.g. saves/png/foo.png
	Format   string // "png" or "svg"
	DPI      float64
	RecordDB bool
}

type Sketch struct {
	Title                     string
	Prefix                    string
	SketchWidth               float64
	SketchHeight              float64
	ControlWidth              int
	ControlHeight             int
	ControlBackgroundColor    string
	ControlOutlineColor       string
	SketchBackgroundColor     string
	SketchOutlineColor        string
	// DefaultBackground is the canvas clear color before each draw (image/color, default black).
	DefaultBackground color.Color
	// DefaultForeground is the initial stroke color for the canvas context before Drawer (default white).
	DefaultForeground color.Color
	// DefaultStrokeWidth is the initial stroke width in millimeters (default 0.5).
	DefaultStrokeWidth float64
	DisableClearBetweenFrames bool
	ShowFPS                   bool
	RasterDPI                 float64
	PreviewMode               bool
	RandomSeed                int64
	FloatSliders              []FloatSlider
	IntSliders                []IntSlider
	Toggles                   []Toggle
	ColorPickers              []ColorPicker
	Dropdowns                 []Dropdown
	uiPlan                    []controlEntry

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
	SketchCanvas          *canvas.Canvas
	ui                    debugui.DebugUI
	showDebugUI           bool
	uiCaptureState        debugui.InputCapturingState

	offscreen    *ebiten.Image
	cachedRGBA   *image.RGBA
	ctx          *canvas.Context
	dirty        bool
	saveRequests chan SaveRequest
	saveMutex    sync.Mutex

	workDir string
	db      *sketchdb.DB

	viewportW, viewportH int
	scrollX, scrollY       float64

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

	colorModalIdx int // >= 0 => editing ColorPickers[idx]
	// colorModalUpdateSketchBG: default-background modal only; if OK with checked, set SketchBackgroundColor.
	colorModalUpdateSketchBG bool

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
}

func (s *Sketch) Width() float64 {
	return s.SketchWidth * MmPerPx
}

func (s *Sketch) Height() float64 {
	return s.SketchHeight * MmPerPx
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
	if s.SketchBackgroundColor == "" {
		s.SketchBackgroundColor = "#1e1e1e"
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
		s.DefaultStrokeWidth = 0.5
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

	dbPath := filepath.Join(s.workDir, "sketch.db")
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
	s.SketchCanvas = canvas.New(s.Width(), s.Height())
	s.needToClear = true
	s.showDebugUI = true
	s.dirty = true
	s.colorModalIdx = -1

	s.offscreen = ebiten.NewImage(int(s.SketchWidth), int(s.SketchHeight))
	s.ctx = canvas.NewContext(s.SketchCanvas)
	s.saveRequests = make(chan SaveRequest, SaveChannelBuffer)

	go s.saveWorker()

	if s.DisableClearBetweenFrames {
		ebiten.SetScreenClearedEveryFrame(false)
	}
	if err := os.Setenv("EBITEN_SCREENSHOT_KEY", "escape"); err != nil {
		log.Fatal("error while setting ebiten screenshot key: ", err)
	}
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
			sw := float64(s.offscreen.Bounds().Dx())
			sh := float64(s.offscreen.Bounds().Dy())
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
	return (float64(s.viewportW) - float64(s.offscreen.Bounds().Dx())) / 2
}

func (s *Sketch) viewPadY() float64 {
	return (float64(s.viewportH) - float64(s.offscreen.Bounds().Dy())) / 2
}

func (s *Sketch) clampScroll() {
	sw := float64(s.offscreen.Bounds().Dx())
	sh := float64(s.offscreen.Bounds().Dy())
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
	s.Tick++
	return nil
}

func (s *Sketch) Clear() {
	s.needToClear = true
	s.dirty = true
}

func (s *Sketch) sketchBackgroundRGBA() color.RGBA {
	c := stringToColor(s.SketchBackgroundColor)
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
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
		s.SketchCanvas.Reset()
		s.ctx = canvas.NewContext(s.SketchCanvas)

		if !s.DisableClearBetweenFrames || s.needToClear {
			s.ctx.SetFillColor(s.DefaultBackground)
			s.ctx.SetStrokeColor(color.Transparent)
			s.ctx.DrawPath(0, 0, canvas.Rectangle(s.ctx.Width(), s.ctx.Height()))
			s.needToClear = false
		}

		s.ctx.SetStrokeColor(s.DefaultForeground)
		s.ctx.SetStrokeWidth(s.DefaultStrokeWidth)
		s.Drawer(s, s.ctx)

		dpi := s.RasterDPI
		if s.PreviewMode {
			dpi = DefaultPreviewDPI
		}

		rasterizedImg := rasterizer.Draw(s.SketchCanvas, canvas.DPI(dpi), canvas.DefaultColorSpace)

		bounds := rasterizedImg.Bounds()
		if s.cachedRGBA == nil || s.cachedRGBA.Bounds() != bounds {
			s.cachedRGBA = image.NewRGBA(bounds)
		}

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				s.cachedRGBA.Set(x, y, rasterizedImg.At(x, y))
			}
		}

		s.offscreen.WritePixels(s.cachedRGBA.Pix)
		s.dirty = false
	}

	screen.Fill(s.sketchBackgroundRGBA())
	op := &ebiten.DrawImageOptions{}
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

func (s *Sketch) CanvasCoords(x, y float64) gaul.Point {
	xs := x - s.viewPadX() + s.scrollX
	ys := y - s.viewPadY() + s.scrollY
	return gaul.Point{X: MmPerPx * xs, Y: MmPerPx * (s.SketchHeight - ys)}
}

func (s *Sketch) SketchCoords(x, y float64) gaul.Point {
	return gaul.Point{X: x / MmPerPx, Y: s.SketchHeight - y/MmPerPx}
}

func (s *Sketch) PointInSketchArea(x, y float64) bool {
	xs := x - s.viewPadX() + s.scrollX
	ys := y - s.viewPadY() + s.scrollY
	return xs >= 0 && xs <= s.SketchWidth && ys >= 0 && ys <= s.SketchHeight
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

func (s *Sketch) IsMouseOverControlPanel() bool {
	return s.uiCaptureState != 0
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
		s.saveMutex.Lock()
		var err error
		switch req.Format {
		case "png":
			err = renderers.Write(full, s.SketchCanvas, canvas.DPI(req.DPI))
		case "svg":
			err = renderers.Write(full, s.SketchCanvas)
		default:
			err = fmt.Errorf("unknown format %q", req.Format)
		}
		s.saveMutex.Unlock()

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
