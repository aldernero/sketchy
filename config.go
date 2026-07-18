package sketchy

import "image/color"

// Config holds sketch options set from code (no JSON).
type Config struct {
	Title                  string
	Prefix                 string
	SketchWidth            float64
	SketchHeight           float64
	ControlWidth           int
	ControlHeight          int
	ControlBackgroundColor string
	ControlOutlineColor    string
	// SketchBackgroundColor is currently unused at runtime (letterbox uses Builtins dark/light theme).
	SketchBackgroundColor string
	SketchOutlineColor    string
	// DisableClearBetweenFrames keeps previous frames' raster under each new
	// frame so strokes accumulate on screen (display-only; saves render just
	// the current frame). Sketch.Clear() wipes to DefaultBackground.
	DisableClearBetweenFrames bool
	// DisableFastStroke is a no-op kept for compatibility; the old
	// tdewolff/canvas FastStroke workaround is gone with the gaul renderer.
	DisableFastStroke bool
	ShowFPS           bool
	// RasterDPI sets raster resolution (default 96 = one canvas pixel per
	// logical sketch pixel); display size is unaffected, saves gain detail.
	RasterDPI float64
	// PreviewMode rasterizes at half detail and scales up on screen for
	// ~4x faster frames while iterating.
	PreviewMode bool
	RandomSeed  int64
	// DefaultBackground is the canvas clear color before Drawer runs; nil means black at Init.
	DefaultBackground color.Color
	// DefaultForeground is the initial stroke (and default pen) color for the canvas context; nil means white at Init.
	DefaultForeground color.Color
	// DefaultStrokeWidth is the initial stroke width in pixels; 0 means 1 at Init.
	DefaultStrokeWidth float64
	// PaletteDBPath locates the palettedb SQLite database for the Builtins palette
	// dropdowns; empty means the palettedb default (~/.config/palettedb/palettedb.db).
	PaletteDBPath string
	// Images lists files to load at Init; use Image/DrawNamedImage in Drawer by Name.
	Images []ImageAsset
}

// New returns an uninitialized sketch. Set BuildUI, Updater, and Drawer, then call Init().
func New(cfg Config) *Sketch {
	s := &Sketch{
		Title:                     cfg.Title,
		Prefix:                    cfg.Prefix,
		SketchWidth:               cfg.SketchWidth,
		SketchHeight:              cfg.SketchHeight,
		ControlWidth:              cfg.ControlWidth,
		ControlHeight:             cfg.ControlHeight,
		ControlBackgroundColor:    cfg.ControlBackgroundColor,
		ControlOutlineColor:       cfg.ControlOutlineColor,
		SketchBackgroundColor:     cfg.SketchBackgroundColor,
		SketchOutlineColor:        cfg.SketchOutlineColor,
		DisableClearBetweenFrames: cfg.DisableClearBetweenFrames,
		DisableFastStroke:         cfg.DisableFastStroke,
		ShowFPS:                   cfg.ShowFPS,
		RasterDPI:                 cfg.RasterDPI,
		PreviewMode:               cfg.PreviewMode,
		RandomSeed:                cfg.RandomSeed,
		DefaultBackground:         cfg.DefaultBackground,
		DefaultForeground:         cfg.DefaultForeground,
		DefaultStrokeWidth:        cfg.DefaultStrokeWidth,
		PaletteDBPath:             cfg.PaletteDBPath,
		imageAssets:               append([]ImageAsset(nil), cfg.Images...),
	}
	if s.SketchWidth <= 0 {
		s.SketchWidth = 1080
	}
	if s.SketchHeight <= 0 {
		s.SketchHeight = 1080
	}
	return s
}
