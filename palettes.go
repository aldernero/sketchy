package sketchy

import (
	"log"
	"os"
	"slices"

	"github.com/aldernero/gaul"
	"github.com/aldernero/palettedb"
)

// initPaletteDB populates the Builtins palette dropdowns and loads the initial
// selections into DiscretePalette/SinePalette. Stored palettes come from the
// palettedb database at PaletteDBPath (or the palettedb default location);
// palettedb's compiled-in palettes (viridis, plasma, …) are always listed, so
// the dropdowns work even when no database file exists.
func (s *Sketch) initPaletteDB() {
	if s.paletteDB != nil { // Init() may run more than once
		if err := s.paletteDB.Close(); err != nil {
			log.Printf("sketchy: close palette db: %v", err)
		}
		s.paletteDB = nil
	}
	if s.DiscretePalette.NumStops() == 0 {
		s.DiscretePalette = gaul.NewGradientFromNamed([]string{"black", "white"})
	}
	if s.SinePalette == (gaul.SinePalette{}) {
		// Inigo Quilez's classic rainbow cosine palette.
		s.SinePalette = gaul.NewSinePalette(gaul.Vec3{X: 1, Y: 1, Z: 1}, gaul.Vec3{X: 0, Y: 0.33, Z: 0.67})
	}

	var discrete, sine []string
	if pdb, ok := s.openPaletteDB(); ok {
		var err error
		if discrete, err = pdb.ListNames("discrete"); err == nil {
			sine, err = pdb.ListNames("sine")
		}
		if err != nil {
			log.Printf("sketchy: list palettes: %v", err)
			discrete, sine = nil, nil
			if cerr := pdb.Close(); cerr != nil {
				log.Printf("sketchy: close palette db: %v", cerr)
			}
		} else {
			s.paletteDB = pdb
		}
	}
	s.discretePaletteNames = appendBuiltinPaletteNames(discrete, "discrete")
	s.sinePaletteNames = appendBuiltinPaletteNames(sine, "sine")
	s.builtinDiscretePaletteIdx = 0
	s.builtinSinePaletteIdx = 0
	s.applyDiscretePaletteSelection()
	s.applySinePaletteSelection()
}

// openPaletteDB opens the database at PaletteDBPath (or the palettedb default
// location). A missing file is not an error — built-ins still populate the
// dropdowns — but it is logged, as is any open failure.
func (s *Sketch) openPaletteDB() (*palettedb.DB, bool) {
	path := s.PaletteDBPath
	if path == "" {
		p, err := palettedb.DefaultPath()
		if err != nil {
			log.Printf("sketchy: palettedb default path: %v", err)
			return nil, false
		}
		path = p
	}
	if fi, err := os.Stat(path); err != nil || !fi.Mode().IsRegular() {
		log.Printf("sketchy: no palette db found at %s; palette dropdowns list built-ins only", path)
		return nil, false
	}
	pdb, err := palettedb.Open(path)
	if err != nil {
		log.Printf("sketchy: open palette db %s: %v", path, err)
		return nil, false
	}
	return pdb, true
}

// appendBuiltinPaletteNames extends the stored-palette names with palettedb's
// compiled-in palettes of the same type (viridis, plasma, …). Stored palettes
// come first and shadow same-named built-ins, matching the by-name loaders.
func appendBuiltinPaletteNames(stored []string, paletteType string) []string {
	names := stored
	for _, n := range palettedb.BuiltinNames(paletteType) {
		if !slices.Contains(stored, n) {
			names = append(names, n)
		}
	}
	return names
}

// SelectedDiscretePalette returns the name selected in the Builtins discrete
// palette dropdown ("" before Init).
func (s *Sketch) SelectedDiscretePalette() string {
	if s.builtinDiscretePaletteIdx < 0 || s.builtinDiscretePaletteIdx >= len(s.discretePaletteNames) {
		return ""
	}
	return s.discretePaletteNames[s.builtinDiscretePaletteIdx]
}

// SelectedSinePalette returns the name selected in the Builtins sine palette
// dropdown ("" before Init).
func (s *Sketch) SelectedSinePalette() string {
	if s.builtinSinePaletteIdx < 0 || s.builtinSinePaletteIdx >= len(s.sinePaletteNames) {
		return ""
	}
	return s.sinePaletteNames[s.builtinSinePaletteIdx]
}

// applyDiscretePaletteSelection loads the selected discrete palette into
// DiscretePalette and invalidates the sketch.
func (s *Sketch) applyDiscretePaletteSelection() {
	name := s.SelectedDiscretePalette()
	if name == "" {
		return
	}
	var g gaul.Gradient
	var err error
	if s.paletteDB != nil {
		g, err = s.paletteDB.LoadDiscreteByName(name)
	} else {
		g, err = palettedb.BuiltinDiscreteByName(name)
	}
	if err != nil {
		log.Printf("sketchy: load discrete palette %q: %v", name, err)
		return
	}
	s.DiscretePalette = g
	s.DidControlsChange = true
	s.dirty = true
}

// applySinePaletteSelection loads the selected sine palette into SinePalette
// and invalidates the sketch.
func (s *Sketch) applySinePaletteSelection() {
	name := s.SelectedSinePalette()
	if name == "" {
		return
	}
	var sp gaul.SinePalette
	var err error
	if s.paletteDB != nil {
		sp, err = s.paletteDB.LoadSineByName(name)
	} else {
		sp, err = palettedb.BuiltinSineByName(name)
	}
	if err != nil {
		log.Printf("sketchy: load sine palette %q: %v", name, err)
		return
	}
	s.SinePalette = sp
	s.DidControlsChange = true
	s.dirty = true
}

// selectPaletteByName moves a Builtins palette dropdown to name and loads it;
// used when applying snapshots. Returns false if name is not in the list.
func (s *Sketch) selectPaletteByName(names []string, idx *int, name string, apply func()) bool {
	for i, n := range names {
		if n == name {
			*idx = i
			apply()
			return true
		}
	}
	return false
}
