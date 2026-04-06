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
	SketchBackgroundColor     string
	SketchOutlineColor        string
	DisableClearBetweenFrames bool
	ShowFPS                   bool
	RasterDPI                 float64
	PreviewMode               bool
	RandomSeed                int64
	// DefaultBackground is the canvas clear color before Drawer runs; nil means black at Init.
	DefaultBackground color.Color
	// DefaultForeground is the initial stroke (and default pen) color for the canvas context; nil means white at Init.
	DefaultForeground color.Color
	// DefaultStrokeWidth is the initial stroke width in millimeters; 0 means 0.5 at Init.
	DefaultStrokeWidth float64
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
		ShowFPS:                   cfg.ShowFPS,
		RasterDPI:                 cfg.RasterDPI,
		PreviewMode:               cfg.PreviewMode,
		RandomSeed:                cfg.RandomSeed,
		DefaultBackground:         cfg.DefaultBackground,
		DefaultForeground:         cfg.DefaultForeground,
		DefaultStrokeWidth:        cfg.DefaultStrokeWidth,
	}
	if s.SketchWidth <= 0 {
		s.SketchWidth = 1080
	}
	if s.SketchHeight <= 0 {
		s.SketchHeight = 1080
	}
	return s
}
