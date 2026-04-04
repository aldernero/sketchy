package sketchy

// UI registers controls inside BuildUI. Call Folder() to group controls under a collapsible header.
type UI struct {
	s      *Sketch
	folder string
}

// Folder runs fn with the given header title as the current folder. Use "" for controls at the root (no extra header).
func (u *UI) Folder(title string, fn func()) {
	prev := u.folder
	u.folder = title
	fn()
	u.folder = prev
}

// FloatSlider adds a float slider and value text field in the current folder.
func (u *UI) FloatSlider(name string, min, max, val, incr float64) {
	u.FloatSliderDecimals(name, min, max, val, incr, -1)
}

// FloatSliderDecimals is like FloatSlider but sets the number of fraction digits shown in the value text
// when decimals >= 0. Use decimals < 0 to derive fraction digits from incr (same precision as the slider thumb).
func (u *UI) FloatSliderDecimals(name string, min, max, val, incr float64, decimals int) {
	sl := NewFloatSliderWithDecimals(name, min, max, val, incr, decimals)
	sl.Folder = u.folder
	u.s.FloatSliders = append(u.s.FloatSliders, sl)
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryFloatSlider, Index: len(u.s.FloatSliders) - 1, Folder: u.folder})
}

// IntSlider adds an integer stepped slider and value text field in the current folder.
func (u *UI) IntSlider(name string, min, max, val, incr int) {
	sl := NewIntSlider(name, min, max, val, incr)
	sl.Folder = u.folder
	u.s.IntSliders = append(u.s.IntSliders, sl)
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryIntSlider, Index: len(u.s.IntSliders) - 1, Folder: u.folder})
}

// Checkbox adds a checkbox in the current folder.
func (u *UI) Checkbox(name string, checked bool) {
	u.s.Toggles = append(u.s.Toggles, Toggle{
		Folder:   u.folder,
		Name:     name,
		Checked:  checked,
		IsButton: false,
	})
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryToggle, Index: len(u.s.Toggles) - 1, Folder: u.folder})
}

// Button adds a momentary button (exposes Checked toggling on click, same as before).
func (u *UI) Button(name string) {
	u.s.Toggles = append(u.s.Toggles, Toggle{
		Folder:   u.folder,
		Name:     name,
		IsButton: true,
	})
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryToggle, Index: len(u.s.Toggles) - 1, Folder: u.folder})
}

// ColorPicker adds a color control in the current folder (initial color as hex or HTML color name).
func (u *UI) ColorPicker(name, initial string) {
	cp := NewColorPicker(name, initial)
	cp.Folder = u.folder
	u.s.ColorPickers = append(u.s.ColorPickers, cp)
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryColor, Index: len(u.s.ColorPickers) - 1, Folder: u.folder})
}

// Dropdown adds a dropdown in the current folder. options must be non-empty.
func (u *UI) Dropdown(name string, options []string, selectedIndex int) {
	if len(options) == 0 {
		panic("sketchy: Dropdown options must be non-empty")
	}
	if selectedIndex < 0 || selectedIndex >= len(options) {
		selectedIndex = 0
	}
	u.s.Dropdowns = append(u.s.Dropdowns, Dropdown{
		Folder:  u.folder,
		Name:    name,
		Options: append([]string(nil), options...),
		Index:   selectedIndex,
	})
	u.s.uiPlan = append(u.s.uiPlan, controlEntry{Kind: entryDropdown, Index: len(u.s.Dropdowns) - 1, Folder: u.folder})
}
