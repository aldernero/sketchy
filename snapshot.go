package sketchy

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const snapshotSchemaVersion = 2

// snapshotPayload is stored in sqlite control_json.
// Schema 1 had only "sliders" (float). Schema 2 adds "int_sliders" for IntSlider values.
type snapshotPayload struct {
	Schema      int                `json:"_schema"`
	Sliders     map[string]float64 `json:"sliders,omitempty"`
	IntSliders  map[string]int     `json:"int_sliders,omitempty"`
	Toggles     map[string]bool    `json:"toggles"`
	Colors      map[string]string  `json:"colors"`
	Dropdowns   map[string]int     `json:"dropdowns"`
}

func controlMapKey(folder, name string) string {
	if folder == "" {
		return name
	}
	return folder + "/" + name
}

func (s *Sketch) serializeControlState() ([]byte, error) {
	p := snapshotPayload{
		Schema:    snapshotSchemaVersion,
		Sliders:   make(map[string]float64),
		Toggles:   make(map[string]bool),
		Colors:    make(map[string]string),
		Dropdowns: make(map[string]int),
	}
	for i := range s.FloatSliders {
		k := controlMapKey(s.FloatSliders[i].Folder, s.FloatSliders[i].Name)
		p.Sliders[k] = s.FloatSliders[i].Val
	}
	if len(s.IntSliders) > 0 {
		p.IntSliders = make(map[string]int)
		for i := range s.IntSliders {
			k := controlMapKey(s.IntSliders[i].Folder, s.IntSliders[i].Name)
			p.IntSliders[k] = s.IntSliders[i].Val
		}
	}
	for i := range s.Toggles {
		k := controlMapKey(s.Toggles[i].Folder, s.Toggles[i].Name)
		p.Toggles[k] = s.Toggles[i].Checked
	}
	for i := range s.ColorPickers {
		if i == s.builtinColorBGIdx || i == s.builtinColorFGIdx {
			continue
		}
		k := controlMapKey(s.ColorPickers[i].Folder, s.ColorPickers[i].Name)
		p.Colors[k] = s.ColorPickers[i].GetHex()
	}
	for i := range s.Dropdowns {
		k := controlMapKey(s.Dropdowns[i].Folder, s.Dropdowns[i].Name)
		p.Dropdowns[k] = s.Dropdowns[i].Index
	}
	return json.Marshal(p)
}

func (s *Sketch) applyControlStateJSON(data []byte) ([]string, error) {
	var p snapshotPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	var missing []string
	for k, v := range p.Sliders {
		f, n := splitControlKey(k)
		if err := s.setFloatQuiet(f, n, v); err != nil {
			// Schema v1 stored all values as floats; int sliders accept rounded values.
			if err2 := s.setIntQuiet(f, n, int(math.Round(v))); err2 == nil {
				continue
			}
			missing = append(missing, k)
		}
	}
	if p.IntSliders != nil {
		for k, v := range p.IntSliders {
			f, n := splitControlKey(k)
			if err := s.setIntQuiet(f, n, v); err != nil {
				missing = append(missing, k)
			}
		}
	}
	for k, v := range p.Toggles {
		f, n := splitControlKey(k)
		if err := s.setBoolQuiet(f, n, v); err != nil {
			missing = append(missing, k)
		}
	}
	for k, v := range p.Colors {
		f, n := splitControlKey(k)
		if err := s.setColorQuiet(f, n, v); err != nil {
			missing = append(missing, k)
		}
	}
	for k, v := range p.Dropdowns {
		f, n := splitControlKey(k)
		if err := s.setDropdownQuiet(f, n, v); err != nil {
			missing = append(missing, k)
		}
	}
	s.syncControlLastState()
	s.syncBuiltinDefaultsFromColorPickers()
	s.DidControlsChange = true
	s.dirty = true
	return missing, nil
}

// builtinSnapshotPayload is stored in sqlite snapshots.builtin_json.
type builtinSnapshotPayload struct {
	DefaultBackground    string  `json:"default_background"`
	DefaultForeground    string  `json:"default_foreground"`
	DefaultStrokeWidthMM float64 `json:"default_stroke_width_mm"`
	RandomSeed           int64   `json:"random_seed"`
}

func (s *Sketch) serializeBuiltinState() ([]byte, error) {
	p := builtinSnapshotPayload{
		DefaultBackground:    colorToRGBHex(s.DefaultBackground),
		DefaultForeground:    colorToRGBHex(s.DefaultForeground),
		DefaultStrokeWidthMM: s.DefaultStrokeWidth,
		RandomSeed:           s.RandomSeed,
	}
	return json.Marshal(p)
}

func (s *Sketch) replaceBuiltinColorPicker(idx int, hex string) {
	if idx < 0 || idx >= len(s.ColorPickers) {
		return
	}
	cur := &s.ColorPickers[idx]
	cp := NewColorPicker(cur.Name, hex)
	cp.Folder = cur.Folder
	s.ColorPickers[idx] = cp
}

func (s *Sketch) applyBuiltinStateJSON(data []byte) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	var p builtinSnapshotPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	const minW, maxW = 0.05, 3.0
	if p.DefaultBackground != "" {
		s.DefaultBackground = stringToColor(p.DefaultBackground)
		s.replaceBuiltinColorPicker(s.builtinColorBGIdx, p.DefaultBackground)
	}
	if p.DefaultForeground != "" {
		s.DefaultForeground = stringToColor(p.DefaultForeground)
		s.replaceBuiltinColorPicker(s.builtinColorFGIdx, p.DefaultForeground)
	}
	s.DefaultStrokeWidth = clampFloat(p.DefaultStrokeWidthMM, minW, maxW)
	s.setRandomSeed(p.RandomSeed)
	s.syncControlLastState()
	return nil
}

func (s *Sketch) syncBuiltinDefaultsFromColorPickers() {
	if s.builtinColorBGIdx >= 0 && s.builtinColorBGIdx < len(s.ColorPickers) {
		s.DefaultBackground = s.ColorPickers[s.builtinColorBGIdx].GetColor()
	}
	if s.builtinColorFGIdx >= 0 && s.builtinColorFGIdx < len(s.ColorPickers) {
		s.DefaultForeground = s.ColorPickers[s.builtinColorFGIdx].GetColor()
	}
}

func splitControlKey(k string) (folder, name string) {
	for i := range k {
		if k[i] == '/' {
			return k[:i], k[i+1:]
		}
	}
	return "", k
}

func (s *Sketch) snapshotKeysPresentInJSON(data []byte) (missing []string, err error) {
	var p snapshotPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	check := func(k string) {
		f, n := splitControlKey(k)
		if !s.hasControl(f, n) {
			missing = append(missing, k)
		}
	}
	for k := range p.Sliders {
		check(k)
	}
	if p.IntSliders != nil {
		for k := range p.IntSliders {
			check(k)
		}
	}
	for k := range p.Toggles {
		check(k)
	}
	for k := range p.Colors {
		check(k)
	}
	for k := range p.Dropdowns {
		check(k)
	}
	return missing, nil
}

func (s *Sketch) hasControl(folder, name string) bool {
	k := controlMapKey(folder, name)
	if _, ok := s.floatSliderControlMap[k]; ok {
		return true
	}
	if _, ok := s.intSliderControlMap[k]; ok {
		return true
	}
	if _, ok := s.toggleControlMap[k]; ok {
		return true
	}
	if _, ok := s.colorPickerControlMap[k]; ok {
		return true
	}
	_, ok := s.dropdownControlMap[k]
	return ok
}

func (s *Sketch) setFloatQuiet(folder, name string, v float64) error {
	k := controlMapKey(folder, name)
	i, ok := s.floatSliderControlMap[k]
	if !ok {
		return fmt.Errorf("no float slider %q", k)
	}
	s.FloatSliders[i].Val = v
	s.FloatSliders[i].lastVal = v
	return nil
}

func (s *Sketch) setIntQuiet(folder, name string, v int) error {
	k := controlMapKey(folder, name)
	i, ok := s.intSliderControlMap[k]
	if !ok {
		return fmt.Errorf("no int slider %q", k)
	}
	s.IntSliders[i].Val = v
	s.IntSliders[i].lastVal = v
	return nil
}

func (s *Sketch) setBoolQuiet(folder, name string, v bool) error {
	k := controlMapKey(folder, name)
	i, ok := s.toggleControlMap[k]
	if !ok {
		return fmt.Errorf("no toggle %q", k)
	}
	s.Toggles[i].Checked = v
	s.Toggles[i].lastVal = v
	return nil
}

func (s *Sketch) setColorQuiet(folder, name string, hex string) error {
	k := controlMapKey(folder, name)
	i, ok := s.colorPickerControlMap[k]
	if !ok {
		return fmt.Errorf("no color %q", k)
	}
	fld := s.ColorPickers[i].Folder
	cp := NewColorPicker(s.ColorPickers[i].Name, hex)
	cp.Folder = fld
	s.ColorPickers[i] = cp
	return nil
}

func (s *Sketch) setDropdownQuiet(folder, name string, idx int) error {
	k := controlMapKey(folder, name)
	i, ok := s.dropdownControlMap[k]
	if !ok {
		return fmt.Errorf("no dropdown %q", k)
	}
	if idx < 0 || idx >= len(s.Dropdowns[i].Options) {
		idx = 0
	}
	s.Dropdowns[i].Index = idx
	s.Dropdowns[i].lastIdx = idx
	return nil
}

func (s *Sketch) syncControlLastState() {
	for i := range s.FloatSliders {
		s.FloatSliders[i].lastVal = s.FloatSliders[i].Val
	}
	for i := range s.IntSliders {
		s.IntSliders[i].lastVal = s.IntSliders[i].Val
	}
	for i := range s.Toggles {
		s.Toggles[i].lastVal = s.Toggles[i].Checked
	}
	for i := range s.ColorPickers {
		s.ColorPickers[i].lastColor = s.ColorPickers[i].Color
	}
	for i := range s.Dropdowns {
		s.Dropdowns[i].lastIdx = s.Dropdowns[i].Index
	}
}
