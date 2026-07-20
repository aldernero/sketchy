package sketchy

import (
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// shaderTimeTPS is the nominal ticks-per-second used to derive the Time
// builtin uniform (Time = Tick/60). The shader template pins
// ebiten.SetTPS(60) so Time advances in real seconds and recordings are
// deterministic.
const shaderTimeTPS = 60.0

// shaderReloadPollTicks is how often the shader file's mtime is checked.
const shaderReloadPollTicks = 30

// IsShaderSketch reports whether this sketch renders with a Kage shader
// (Config.ShaderPath or Config.ShaderSrc) instead of a CPU Drawer.
func (s *Sketch) IsShaderSketch() bool {
	return s.ShaderPath != "" || len(s.ShaderSrc) > 0
}

// initShader loads, parses, and compiles the shader at Init. Any failure
// here is fatal: a shader sketch has nothing else to render.
func (s *Sketch) initShader() {
	if !s.IsShaderSketch() {
		return
	}
	if s.DisableClearBetweenFrames {
		log.Fatal("sketchy: DisableClearBetweenFrames is not supported in shader mode")
	}
	src, mtime, err := s.loadShaderSource()
	if err != nil {
		log.Fatalf("sketchy: %v", err)
	}
	if err := s.applyShaderSource(src); err != nil {
		log.Fatalf("sketchy: %v", err)
	}
	s.shaderMtime = mtime
	s.shaderErr = ""
	s.shaderStatus = ""
}

func (s *Sketch) loadShaderSource() ([]byte, time.Time, error) {
	if s.ShaderPath != "" {
		info, err := os.Stat(s.ShaderPath)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("shader file: %w", err)
		}
		b, err := os.ReadFile(s.ShaderPath)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("shader file: %w", err)
		}
		return b, info.ModTime(), nil
	}
	return s.ShaderSrc, time.Time{}, nil
}

// applyShaderSource parses and compiles src, replacing the active shader
// and uniform list on success. It does not touch controls; callers decide
// whether a control rebuild is needed.
func (s *Sketch) applyShaderSource(src []byte) error {
	uniforms, err := parseShaderUniforms(src)
	if err != nil {
		return err
	}
	shader, err := ebiten.NewShader(src)
	if err != nil {
		return fmt.Errorf("compiling shader: %w", err)
	}
	s.shader = shader
	s.setShaderUniforms(uniforms)
	for _, u := range uniforms {
		if u.Directive == nil && !isBuiltinUniform(u) {
			fmt.Printf("sketchy: shader uniform %s (%s) has no //sketchy: directive; it will be zero (use //sketchy:none to silence)\n", u.Name, u.Kind)
		}
	}
	return nil
}

// setShaderUniforms stores the uniform list and the traits derived from it
// (auto-animation when Time/Tick is declared, mouse-driven redraws).
func (s *Sketch) setShaderUniforms(uniforms []shaderUniform) {
	s.shaderUniforms = uniforms
	s.shaderAnimates = false
	s.shaderUsesMouse = false
	for _, u := range uniforms {
		if !isBuiltinUniform(u) {
			continue
		}
		switch u.Name {
		case "Time", "Tick":
			s.shaderAnimates = true
		case "Mouse":
			s.shaderUsesMouse = true
		}
	}
}

// registerShaderControls creates panel controls from the shader's
// //sketchy: directives, in declaration order. Called from rebuildControls
// after the user's BuildUI so user controls list first.
func (s *Sketch) registerShaderControls(ui *UI) {
	for _, u := range s.shaderUniforms {
		d := u.Directive
		if d == nil || d.Control == "none" {
			continue
		}
		name := u.controlName()
		ui.Folder(d.Folder, func() {
			switch d.Control {
			case "slider":
				if u.Kind == ukInt {
					ui.IntSlider(name, int(d.Min), int(d.Max), int(d.Default), int(d.Step))
				} else {
					ui.FloatSliderDecimals(name, d.Min, d.Max, d.Default, d.Step, d.Digits)
				}
			case "checkbox":
				ui.Checkbox(name, d.Default != 0)
			case "color":
				ui.ColorPicker(name, d.DefaultHex)
			case "dropdown":
				ui.Dropdown(name, d.Options, d.DefaultIdx)
			}
		})
	}
}

// buildUniforms assembles the uniform map for one shader draw at the given
// render-target pixel size: directive-mapped control values, then builtins
// declared by the shader, then ExtraUniforms (which wins on conflicts).
func (s *Sketch) buildUniforms(w, h int) map[string]any {
	m := make(map[string]any, len(s.shaderUniforms)+1)
	for _, u := range s.shaderUniforms {
		if u.Directive != nil && u.Directive.Control != "none" {
			if v, ok := s.uniformControlValue(u); ok {
				m[u.Name] = v
			}
			continue
		}
		if !isBuiltinUniform(u) {
			continue
		}
		switch u.Name {
		case "Time":
			m[u.Name] = float64(s.Tick) / shaderTimeTPS
		case "Tick":
			m[u.Name] = int(s.Tick)
		case "Resolution":
			m[u.Name] = []float32{float32(w), float32(h)}
		case "Mouse":
			p := s.CanvasCoords(cursorPositionF())
			m[u.Name] = []float32{float32(p.X), float32(p.Y)}
		case "Seed":
			m[u.Name] = float64(s.RandomSeed)
		}
	}
	if s.ExtraUniforms != nil {
		for k, v := range s.ExtraUniforms(s) {
			m[k] = v
		}
	}
	return m
}

func cursorPositionF() (float64, float64) {
	x, y := ebiten.CursorPosition()
	return float64(x), float64(y)
}

// uniformControlValue reads the panel control backing a directive uniform.
func (s *Sketch) uniformControlValue(u shaderUniform) (any, bool) {
	d := u.Directive
	key := controlMapKey(d.Folder, u.controlName())
	switch d.Control {
	case "slider":
		if u.Kind == ukInt {
			if i, ok := s.intSliderControlMap[key]; ok {
				return s.IntSliders[i].Val, true
			}
		} else if i, ok := s.floatSliderControlMap[key]; ok {
			return s.FloatSliders[i].Val, true
		}
	case "checkbox":
		if i, ok := s.toggleControlMap[key]; ok {
			v := 0
			if s.Toggles[i].Checked {
				v = 1
			}
			if u.Kind == ukFloat {
				return float64(v), true
			}
			return v, true
		}
	case "color":
		if i, ok := s.colorPickerControlMap[key]; ok {
			r, g, b, _ := s.ColorPickers[i].GetColor().RGBA()
			rgb := []float32{float32(r) / 65535, float32(g) / 65535, float32(b) / 65535}
			if u.Kind == ukVec4 {
				return append(rgb, 1), true
			}
			return rgb, true
		}
	case "dropdown":
		if i, ok := s.dropdownControlMap[key]; ok {
			return s.Dropdowns[i].Index, true
		}
	}
	return nil, false
}

// renderShaderFrame draws the shader with the current uniforms over all of
// dst. Runs on the ebiten thread.
func (s *Sketch) renderShaderFrame(dst *ebiten.Image) {
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	dst.Clear()
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Uniforms = s.buildUniforms(w, h)
	dst.DrawRectShader(w, h, s.shader, opts)
}

// CaptureShaderImage renders the shader at scale (1 = one pixel per sketch
// pixel) and reads the result back to the CPU — the building block for
// custom export flows (pair with [Sketch.EnqueueSavePixels]). Must be
// called on the ebiten thread (Updater/Drawer callbacks are fine); returns
// nil for non-shader sketches. Pixels are premultiplied RGBA, identical in
// layout to the CPU raster path.
func (s *Sketch) CaptureShaderImage(scale float64) *image.RGBA {
	if !s.IsShaderSketch() {
		return nil
	}
	if scale <= 0 {
		scale = 1
	}
	w := int(s.SketchWidth*scale + 0.5)
	h := int(s.SketchHeight*scale + 0.5)
	if s.shaderTarget == nil || s.shaderTarget.Bounds().Dx() != w || s.shaderTarget.Bounds().Dy() != h {
		s.shaderTarget = ebiten.NewImage(w, h)
	}
	s.renderShaderFrame(s.shaderTarget)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	s.shaderTarget.ReadPixels(img.Pix)
	return img
}

// updateShader runs once per tick from Update: live reload polling and
// automatic dirtying for animated / mouse-driven shaders.
func (s *Sketch) updateShader() {
	if !s.IsShaderSketch() {
		return
	}
	s.checkShaderReload()
	if s.shaderAnimates {
		s.dirty = true
	}
	if s.shaderUsesMouse {
		cx, cy := ebiten.CursorPosition()
		if cx != s.lastCursorX || cy != s.lastCursorY {
			s.lastCursorX, s.lastCursorY = cx, cy
			s.dirty = true
		}
	}
}

// checkShaderReload polls the shader file's mtime and hot-swaps the shader
// on change. Compile/parse errors keep the last good shader and surface in
// the Builtins panel; on success, controls are rebuilt from the new
// directives with current values preserved by name.
func (s *Sketch) checkShaderReload() {
	if s.ShaderPath == "" || s.Tick%shaderReloadPollTicks != 0 {
		return
	}
	info, err := os.Stat(s.ShaderPath)
	if err != nil || !info.ModTime().After(s.shaderMtime) {
		return
	}
	s.shaderMtime = info.ModTime()
	src, err := os.ReadFile(s.ShaderPath)
	if err != nil {
		s.shaderErr = fmt.Sprintf("Shader reload: %v", err)
		fmt.Println(s.shaderErr)
		return
	}
	if err := s.applyShaderSource(src); err != nil {
		s.shaderErr = fmt.Sprintf("Shader reload failed (keeping last good shader): %v", err)
		fmt.Println(s.shaderErr)
		return
	}
	s.rebuildControlsPreservingValues()
	s.shaderErr = ""
	s.shaderStatus = "Shader reloaded " + time.Now().Format("15:04:05")
	fmt.Println(s.shaderStatus)
	s.MarkDirty()
}

// rebuildControlsPreservingValues re-registers all controls (user BuildUI +
// shader directives + builtins) and restores the current value of every
// control whose folder/name survives, clamped to any new slider range.
func (s *Sketch) rebuildControlsPreservingValues() {
	floats := make(map[string]float64)
	ints := make(map[string]int)
	bools := make(map[string]bool)
	colors := make(map[string]string)
	drops := make(map[string]int)
	for _, c := range s.FloatSliders {
		floats[controlMapKey(c.Folder, c.Name)] = c.Val
	}
	for _, c := range s.IntSliders {
		ints[controlMapKey(c.Folder, c.Name)] = c.Val
	}
	for _, c := range s.Toggles {
		bools[controlMapKey(c.Folder, c.Name)] = c.Checked
	}
	for _, c := range s.ColorPickers {
		colors[controlMapKey(c.Folder, c.Name)] = c.GetHex()
	}
	for _, c := range s.Dropdowns {
		drops[controlMapKey(c.Folder, c.Name)] = c.Index
	}

	s.rebuildControls()

	for i := range s.FloatSliders {
		c := &s.FloatSliders[i]
		if v, ok := floats[controlMapKey(c.Folder, c.Name)]; ok {
			c.Val = clampFloat(v, c.MinVal, c.MaxVal)
		}
	}
	for i := range s.IntSliders {
		c := &s.IntSliders[i]
		if v, ok := ints[controlMapKey(c.Folder, c.Name)]; ok {
			c.Val = clampInt(v, c.MinVal, c.MaxVal)
		}
	}
	for i := range s.Toggles {
		c := &s.Toggles[i]
		if v, ok := bools[controlMapKey(c.Folder, c.Name)]; ok {
			c.Checked = v
		}
	}
	for i := range s.ColorPickers {
		c := &s.ColorPickers[i]
		if hex, ok := colors[controlMapKey(c.Folder, c.Name)]; ok {
			restored := NewColorPicker(c.Name, hex)
			restored.Folder = c.Folder
			s.ColorPickers[i] = restored
		}
	}
	for i := range s.Dropdowns {
		c := &s.Dropdowns[i]
		if v, ok := drops[controlMapKey(c.Folder, c.Name)]; ok && v >= 0 && v < len(c.Options) {
			c.Index = v
		}
	}

	// Modals hold control indices; a rebuild invalidates them.
	s.colorModalIdx = -1
	s.sliderRangeModalOpen = false
}
