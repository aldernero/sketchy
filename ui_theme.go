package sketchy

import (
	_ "embed"
	"log"

	"github.com/aldernero/debugui"
)

//go:embed themes/dark.json
var debugUIDarkThemeJSON []byte

//go:embed themes/light.json
var debugUILightThemeJSON []byte

// debugUIThemeLabels matches [Sketch.debugUIThemeIndex] for the Builtins theme dropdown.
var debugUIThemeLabels = []string{
	"Dark",
	"Light",
}

func (s *Sketch) applyDebugUITheme() {
	i := s.debugUIThemeIndex
	if i < 0 || i >= len(debugUIThemeLabels) {
		s.debugUIThemeIndex = 0
		i = 0
	}
	var raw []byte
	switch i {
	case 0:
		raw = debugUIDarkThemeJSON
	case 1:
		raw = debugUILightThemeJSON
	default:
		s.ui.SetStyle(nil)
		return
	}
	st, err := debugui.ParseStyleJSON(raw)
	if err != nil {
		log.Printf("sketchy: parse UI theme: %v", err)
		s.ui.SetStyle(nil)
		return
	}
	s.ui.SetStyle(&st)
}
