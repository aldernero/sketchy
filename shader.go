package sketchy

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/draw"
)

// shaderTimeTPS is the nominal ticks-per-second used to derive the Time
// builtin uniform (Time = Tick/60). The shader template pins
// ebiten.SetTPS(60) so Time advances in real seconds and recordings are
// deterministic.
const shaderTimeTPS = 60.0

// shaderReloadPollTicks is how often the shader file's mtime is checked.
const shaderReloadPollTicks = 30

// pingPongImageSlot is the Fragment source-image slot (imageSrc0At) reserved
// for the ping-pong state buffer when StatePath is set. Static
// //sketchy:image directives may not claim this slot in that case.
const pingPongImageSlot = 0

// IsShaderSketch reports whether this sketch renders with a Kage shader
// (Config.ShaderPath or Config.ShaderSrc) instead of a CPU Drawer.
func (s *Sketch) IsShaderSketch() bool {
	return s.ShaderPath != "" || len(s.ShaderSrc) > 0
}

// initShader loads, parses, and compiles the shader (and, if StatePath is
// set, the state/simulation shader and its ping-pong buffers) at Init. Any
// failure here is fatal: a shader sketch has nothing else to render.
func (s *Sketch) initShader() {
	if s.StatePath != "" && !s.IsShaderSketch() {
		log.Fatal("sketchy: StatePath requires ShaderPath (or ShaderSrc) for the display pass")
	}
	if !s.IsShaderSketch() {
		return
	}
	if s.DisableClearBetweenFrames {
		log.Fatal("sketchy: DisableClearBetweenFrames is not supported in shader mode")
	}

	// The live display always renders 1:1 with the sketch's logical size
	// (s.offscreen is allocated at exactly SketchWidth x SketchHeight and
	// the shader draws straight into it), so Preview Mode never applies in
	// shader mode and is forced off and hidden from the Builtins panel.
	// Export Scale does apply: it controls the resolution of PNG saves
	// only (see CaptureShaderImage), independent of what's on screen. See
	// docs/shaders.md.
	s.PreviewMode = false

	src, mtime, err := s.loadShaderSource(s.ShaderPath, s.ShaderSrc)
	if err != nil {
		log.Fatalf("sketchy: %v", err)
	}
	if err := s.applyShaderSource(src); err != nil {
		log.Fatalf("sketchy: %v", err)
	}
	s.shaderMtime = mtime
	s.shaderErr = ""
	s.shaderStatus = ""

	var stateSrc []byte
	if s.StatePath != "" {
		stateSrc, mtime, err = s.loadShaderSource(s.StatePath, nil)
		if err != nil {
			log.Fatalf("sketchy: %v", err)
		}
		if err := s.applyStateShaderSource(stateSrc); err != nil {
			log.Fatalf("sketchy: %v", err)
		}
		s.stateMtime = mtime

		w, h := int(s.SketchWidth), int(s.SketchHeight)
		s.pingFront = ebiten.NewImage(w, h)
		s.pingBack = ebiten.NewImage(w, h)
	}

	if err := s.loadShaderImages(src, stateSrc); err != nil {
		log.Fatalf("sketchy: %v", err)
	}
}

// loadShaderSource reads path (if non-empty), returning its content and
// mtime; otherwise it falls back to embedded, with a zero mtime (no live
// reload for embedded source).
func (s *Sketch) loadShaderSource(path string, embedded []byte) ([]byte, time.Time, error) {
	if path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("shader file: %w", err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("shader file: %w", err)
		}
		return b, info.ModTime(), nil
	}
	return embedded, time.Time{}, nil
}

// shaderSourceName is the name a shader source reports in compile errors.
// Embedded sources have no path, so they get a stable placeholder.
func shaderSourceName(path string) string {
	if path == "" {
		return "shader.kage"
	}
	return path
}

// shaderDep is an imported library file the shader was built from, watched
// alongside the shader itself so editing a library live-reloads the sketch.
type shaderDep struct {
	mtime time.Time
	path  string
}

func statShaderDeps(paths []string) []shaderDep {
	deps := make([]shaderDep, 0, len(paths))
	for _, p := range paths {
		var mtime time.Time
		if info, err := os.Stat(p); err == nil {
			mtime = info.ModTime()
		}
		deps = append(deps, shaderDep{path: p, mtime: mtime})
	}
	return deps
}

// shaderDepsChanged reports whether any imported library has been touched. A
// dep that has become unreadable counts as changed, so the reload runs and
// surfaces the real error instead of silently rendering stale output.
func shaderDepsChanged(deps []shaderDep) bool {
	for _, d := range deps {
		info, err := os.Stat(d.path)
		if err != nil || info.ModTime().After(d.mtime) {
			return true
		}
	}
	return false
}

// applyShaderSource parses and compiles src as the display shader, replacing
// the active shader and uniform list on success, and recomputes
// shaderAnimates/shaderUsesMouse over both the display and (if present)
// state uniform lists. It does not touch controls; callers decide whether a
// control rebuild is needed.
//
// Uniform and image directives are read from the original source, not the
// import-resolved one: libraries may not declare uniforms (resolveShaderImports
// rejects package-level vars), so only the sketch's own source can contribute
// controls.
func (s *Sketch) applyShaderSource(src []byte) error {
	uniforms, err := parseShaderUniforms(src)
	if err != nil {
		return err
	}
	merged, deps, err := resolveShaderImports(src, shaderSourceName(s.ShaderPath), s.workDir)
	if err != nil {
		return fmt.Errorf("resolving shader imports: %w", err)
	}
	shader, err := ebiten.NewShader(merged)
	if err != nil {
		return fmt.Errorf("compiling shader: %w", err)
	}
	s.shader = shader
	s.shaderDeps = statShaderDeps(deps)
	s.setShaderUniforms(uniforms)
	warnUndirectedUniforms(uniforms)
	return nil
}

// applyStateShaderSource parses and compiles src as the state (simulation)
// shader. Mirrors applyShaderSource for the second pass.
func (s *Sketch) applyStateShaderSource(src []byte) error {
	uniforms, err := parseShaderUniforms(src)
	if err != nil {
		return err
	}
	merged, deps, err := resolveShaderImports(src, shaderSourceName(s.StatePath), s.workDir)
	if err != nil {
		return fmt.Errorf("resolving state shader imports: %w", err)
	}
	shader, err := ebiten.NewShader(merged)
	if err != nil {
		return fmt.Errorf("compiling state shader: %w", err)
	}
	s.stateShader = shader
	s.stateDeps = statShaderDeps(deps)
	s.stateUniforms = uniforms
	s.recomputeShaderTraits()
	warnUndirectedUniforms(uniforms)
	return nil
}

func warnUndirectedUniforms(uniforms []shaderUniform) {
	for _, u := range uniforms {
		if u.Directive == nil && !isBuiltinUniform(u) {
			fmt.Printf("sketchy: shader uniform %s (%s) has no //sketchy: directive; it will be zero (use //sketchy:none to silence)\n", u.Name, u.Kind)
		}
	}
}

// loadShaderImages parses //sketchy:image directives from the display
// source and (if present) the state source, validates slot assignment
// (slot 0 is reserved for the ping-pong buffer when StatePath is set), and
// loads+resizes each bound image to exactly SketchWidth x SketchHeight —
// Ebitengine requires every DrawRectShaderOptions.Images entry to match the
// draw target's size exactly, and the live display always renders at the
// sketch's native logical size (see initShader). A supersampled export
// (CaptureShaderImage) upscales a copy of these on the fly rather than
// reloading at a different size. Builds into a local array and only
// commits on full success, so a bad directive never leaves s.shaderImages
// partially updated.
func (s *Sketch) loadShaderImages(displaySrc, stateSrc []byte) error {
	dirs, err := parseShaderImageDirectives(displaySrc)
	if err != nil {
		return fmt.Errorf("display shader: %w", err)
	}
	if len(stateSrc) > 0 {
		stateDirs, err := parseShaderImageDirectives(stateSrc)
		if err != nil {
			return fmt.Errorf("state shader: %w", err)
		}
		dirs = append(dirs, stateDirs...)
	}

	seen := map[int]bool{}
	var imgs [4]*ebiten.Image
	for _, d := range dirs {
		if s.StatePath != "" && d.Slot == pingPongImageSlot {
			return fmt.Errorf("//sketchy:image slot %d is reserved for the ping-pong state buffer (StatePath is set); use slot=1-3", pingPongImageSlot)
		}
		if seen[d.Slot] {
			return fmt.Errorf("//sketchy:image slot %d is bound more than once", d.Slot)
		}
		seen[d.Slot] = true
		img, err := s.loadShaderImage(d.Path)
		if err != nil {
			return fmt.Errorf("//sketchy:image %s: %w", d.Path, err)
		}
		imgs[d.Slot] = img
	}

	// Dispose the images being replaced; nothing else references them.
	for _, old := range s.shaderImages {
		if old != nil {
			old.Dispose()
		}
	}
	s.shaderImages = imgs
	s.imageDirectives = dirs
	return nil
}

// loadShaderImage decodes the image at path (relative to the sketch working
// directory unless absolute) and resizes it to exactly SketchWidth x
// SketchHeight using a high-quality CPU resize, then uploads it as a
// *ebiten.Image ready to bind to a Fragment source-image slot.
func (s *Sketch) loadShaderImage(path string) (*ebiten.Image, error) {
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(s.workDir, full)
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	w, h := int(s.SketchWidth), int(s.SketchHeight)
	resized := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(resized, resized.Bounds(), src, src.Bounds(), draw.Over, nil)
	return ebiten.NewImageFromImage(resized), nil
}

// setShaderUniforms stores the display shader's uniform list and
// recomputes the traits derived from it and the state shader's list (if
// any): auto-animation when Time/Tick is declared (or StatePath is set —
// a ping-pong shader is definitionally continuous), and mouse-driven
// redraws.
func (s *Sketch) setShaderUniforms(uniforms []shaderUniform) {
	s.shaderUniforms = uniforms
	s.recomputeShaderTraits()
}

// recomputeShaderTraits derives shaderAnimates/shaderUsesMouse from the
// combined display + state uniform lists. Safe to call independently of
// which list last changed since it always reads current field values.
func (s *Sketch) recomputeShaderTraits() {
	s.shaderAnimates = s.StatePath != ""
	s.shaderUsesMouse = false
	for _, u := range s.allShaderUniforms() {
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

// allShaderUniforms returns the display and state shaders' uniforms
// combined, for control registration and trait derivation.
func (s *Sketch) allShaderUniforms() []shaderUniform {
	if len(s.stateUniforms) == 0 {
		return s.shaderUniforms
	}
	out := make([]shaderUniform, 0, len(s.shaderUniforms)+len(s.stateUniforms))
	out = append(out, s.shaderUniforms...)
	out = append(out, s.stateUniforms...)
	return out
}

// registerShaderControls creates panel controls from both shaders'
// //sketchy: directives, in declaration order (display shader first, then
// state shader). Called from rebuildControls after the user's BuildUI so
// user controls list first. A uniform declared with the same folder+name in
// both files resolves to one shared control feeding both passes.
func (s *Sketch) registerShaderControls(ui *UI) {
	for _, u := range s.allShaderUniforms() {
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

// buildUniforms assembles the uniform map for the display pass at the given
// render-target pixel size.
func (s *Sketch) buildUniforms(w, h int) map[string]any {
	return s.buildUniformsFor(s.shaderUniforms, w, h)
}

// buildUniformsFor assembles a uniform map for one shader draw (display or
// state pass) at the given render-target pixel size: directive-mapped
// control values, then builtins declared by that shader, then
// ExtraUniforms (which wins on conflicts, applied identically to both
// passes).
func (s *Sketch) buildUniformsFor(uniforms []shaderUniform, w, h int) map[string]any {
	m := make(map[string]any, len(uniforms)+1)
	for _, u := range uniforms {
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
		case "Substep":
			// Index of the current state-pass iteration within this tick
			// (0 when Steps=1). Declared //sketchy:none on the state shader.
			m[u.Name] = s.stateSubstep
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

// currentShaderImages returns the source-image array for the next draw
// call: static //sketchy:image bindings, with the ping-pong buffer's
// current state substituted at pingPongImageSlot when StatePath is set.
func (s *Sketch) currentShaderImages() [4]*ebiten.Image {
	imgs := s.shaderImages
	if s.stateShader != nil {
		imgs[pingPongImageSlot] = s.pingFront
	}
	return imgs
}

// ClearState clears the ping-pong state buffers to transparent black.
// Pair with a one-tick Reset uniform (and Tick==0 seeding in the state
// shader) so the next advanceState reseeds instead of evolving stale
// contents — used by examples/reaction_diffusion's Reset button.
func (s *Sketch) ClearState() {
	if s.pingFront != nil {
		s.pingFront.Clear()
	}
	if s.pingBack != nil {
		s.pingBack.Clear()
	}
}

// stateSteps returns how many times to run the state pass this tick.
// A state-shader int slider named Steps (any folder) selects the count;
// otherwise a single pass. Clamped to at least 1.
func (s *Sketch) stateSteps() int {
	for _, u := range s.stateUniforms {
		if u.Name != "Steps" || u.Kind != ukInt || u.Directive == nil || u.Directive.Control != "slider" {
			continue
		}
		key := controlMapKey(u.Directive.Folder, u.controlName())
		if i, ok := s.intSliderControlMap[key]; ok {
			if v := s.IntSliders[i].Val; v >= 1 {
				return v
			}
		}
		return 1
	}
	return 1
}

// advanceState runs the state (simulation) pass once: it reads the current
// ping-pong buffer (pingFront) via imageSrc0At, writes the next one
// (pingBack), then swaps them so pingFront is always "the current state"
// for both the next tick's state-pass read and the display pass that
// follows within the same tick. Called from updateShader (possibly several
// times per tick when Steps > 1); never from a capture/export path, so
// saving an image never perturbs the simulation.
func (s *Sketch) advanceState() {
	if s.stateShader == nil {
		return
	}
	w, h := int(s.SketchWidth), int(s.SketchHeight)
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Blend = ebiten.BlendCopy
	opts.Uniforms = s.buildUniformsFor(s.stateUniforms, w, h)
	opts.Images = s.shaderImages
	opts.Images[pingPongImageSlot] = s.pingFront
	s.pingBack.Clear()
	s.pingBack.DrawRectShader(w, h, s.stateShader, opts)
	s.pingFront, s.pingBack = s.pingBack, s.pingFront
}

// renderShaderFrame draws the display pass over all of dst, reading
// whichever source images (including the ping-pong buffer, if any) are
// currently bound. Pure read of current state — never advances the
// simulation, so it is safe to call from both the live display path and
// CaptureShaderImage without side effects.
func (s *Sketch) renderShaderFrame(dst *ebiten.Image) {
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	dst.Clear()
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Uniforms = s.buildUniforms(w, h)
	opts.Images = s.currentShaderImages()
	dst.DrawRectShader(w, h, s.shader, opts)
}

// CaptureShaderImage renders the display pass at the Builtins Export Scale
// resolution (RasterDPI/DefaultDPI x SketchWidth x SketchHeight — the live
// display always stays native 1:1, see initShader) and reads the result
// back to the CPU — the building block for custom export flows (pair with
// [Sketch.EnqueueSavePixels]). Must be called on the ebiten thread
// (Updater/Drawer callbacks are fine); returns nil for non-shader sketches.
// Pixels are premultiplied RGBA, identical in layout to the CPU raster
// path.
//
// Any bound //sketchy:image source or ping-pong state buffer is native
// resolution (SketchWidth x SketchHeight) regardless of scale — Ebitengine
// requires every DrawRectShaderOptions.Images entry to match the draw
// target's size exactly, so at scale != 1 they are upscaled (GPU, linear
// filter) into scratch images sized to match, used for this one capture,
// and disposed; the originals and the live simulation are untouched.
func (s *Sketch) CaptureShaderImage() *image.RGBA {
	if !s.IsShaderSketch() {
		return nil
	}
	scale := s.RasterDPI / DefaultDPI
	if scale <= 0 {
		scale = 1
	}
	w := int(s.SketchWidth*scale + 0.5)
	h := int(s.SketchHeight*scale + 0.5)
	if s.shaderTarget == nil || s.shaderTarget.Bounds().Dx() != w || s.shaderTarget.Bounds().Dy() != h {
		s.shaderTarget = ebiten.NewImage(w, h)
	}

	nativeW, nativeH := int(s.SketchWidth), int(s.SketchHeight)
	if w == nativeW && h == nativeH {
		s.renderShaderFrame(s.shaderTarget)
	} else {
		scaledImages, dispose := scaleShaderImages(s.currentShaderImages(), nativeW, nativeH, w, h)
		defer dispose()
		s.shaderTarget.Clear()
		opts := &ebiten.DrawRectShaderOptions{}
		opts.Uniforms = s.buildUniforms(w, h)
		opts.Images = scaledImages
		s.shaderTarget.DrawRectShader(w, h, s.shader, opts)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	s.shaderTarget.ReadPixels(img.Pix)
	return img
}

// scaleShaderImages upscales each non-nil entry of imgs from
// (nativeW, nativeH) to (w, h) with GPU linear filtering, for use as
// DrawRectShaderOptions.Images against a (w, h) draw target (Ebitengine
// requires an exact size match). The returned dispose func frees the
// scratch images; it is always safe to call.
func scaleShaderImages(imgs [4]*ebiten.Image, nativeW, nativeH, w, h int) (scaled [4]*ebiten.Image, dispose func()) {
	var made []*ebiten.Image
	sx := float64(w) / float64(nativeW)
	sy := float64(h) / float64(nativeH)
	for i, img := range imgs {
		if img == nil {
			continue
		}
		dst := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		op.GeoM.Scale(sx, sy)
		dst.DrawImage(img, op)
		scaled[i] = dst
		made = append(made, dst)
	}
	return scaled, func() {
		for _, img := range made {
			img.Dispose()
		}
	}
}

// updateShader runs once per tick from Update: live reload polling, the
// state-pass advance(s) (if StatePath is set), and automatic dirtying for
// animated / mouse-driven shaders. When the state shader declares a Steps
// int slider, the simulation pass runs that many times per tick (with
// Substep = 0..Steps-1) so slow GPU sims like Gray-Scott can keep up.
func (s *Sketch) updateShader() {
	if !s.IsShaderSketch() {
		return
	}
	s.checkShaderReload()
	steps := s.stateSteps()
	for i := 0; i < steps; i++ {
		s.stateSubstep = i
		s.advanceState()
	}
	s.stateSubstep = 0
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

// checkShaderReload polls the shader file(s)' mtime and hot-swaps on
// change. Both the display and (if set) state shader are recompiled
// together and committed only if both succeed, since a bad edit to either
// must never leave the pair (and their shared image-slot assignment) out of
// sync — the last good pair keeps rendering, with the error surfaced in the
// Builtins panel and on stdout, until both files compile again. Ping-pong
// buffer contents are never reset by a reload.
func (s *Sketch) checkShaderReload() {
	if s.ShaderPath == "" || s.Tick%shaderReloadPollTicks != 0 {
		return
	}
	displayInfo, err := os.Stat(s.ShaderPath)
	if err != nil {
		return
	}
	var stateInfo os.FileInfo
	if s.StatePath != "" {
		stateInfo, err = os.Stat(s.StatePath)
		if err != nil {
			return
		}
	}
	displayChanged := displayInfo.ModTime().After(s.shaderMtime)
	stateChanged := s.StatePath != "" && stateInfo.ModTime().After(s.stateMtime)
	// Editing an imported library must reload too, even though neither shader
	// file itself was touched.
	depsChanged := shaderDepsChanged(s.shaderDeps) || shaderDepsChanged(s.stateDeps)
	if !displayChanged && !stateChanged && !depsChanged {
		return
	}

	displaySrc, err := os.ReadFile(s.ShaderPath)
	if err != nil {
		s.reportShaderReloadErr(fmt.Sprintf("Shader reload: %v", err))
		return
	}
	var stateSrc []byte
	if s.StatePath != "" {
		stateSrc, err = os.ReadFile(s.StatePath)
		if err != nil {
			s.reportShaderReloadErr(fmt.Sprintf("Shader reload: %v", err))
			return
		}
	}

	newUniforms, err := parseShaderUniforms(displaySrc)
	if err != nil {
		s.reportShaderReloadErr(fmt.Sprintf("Shader reload failed (keeping last good shader): %v", err))
		return
	}
	displayMerged, displayDeps, err := resolveShaderImports(displaySrc, shaderSourceName(s.ShaderPath), s.workDir)
	if err != nil {
		s.reportShaderReloadErr(fmt.Sprintf("Shader reload failed (keeping last good shader): %v", err))
		return
	}
	newShader, err := ebiten.NewShader(displayMerged)
	if err != nil {
		s.reportShaderReloadErr(fmt.Sprintf("Shader reload failed (keeping last good shader): %v", err))
		return
	}
	var newStateUniforms []shaderUniform
	var newStateShader *ebiten.Shader
	var stateDeps []string
	if s.StatePath != "" {
		newStateUniforms, err = parseShaderUniforms(stateSrc)
		if err != nil {
			s.reportShaderReloadErr(fmt.Sprintf("State shader reload failed (keeping last good shader): %v", err))
			return
		}
		var stateMerged []byte
		stateMerged, stateDeps, err = resolveShaderImports(stateSrc, shaderSourceName(s.StatePath), s.workDir)
		if err != nil {
			s.reportShaderReloadErr(fmt.Sprintf("State shader reload failed (keeping last good shader): %v", err))
			return
		}
		newStateShader, err = ebiten.NewShader(stateMerged)
		if err != nil {
			s.reportShaderReloadErr(fmt.Sprintf("State shader reload failed (keeping last good shader): %v", err))
			return
		}
	}
	if err := s.loadShaderImages(displaySrc, stateSrc); err != nil {
		s.reportShaderReloadErr(fmt.Sprintf("Shader reload failed (keeping last good shader): %v", err))
		return
	}

	s.shader = newShader
	s.shaderMtime = displayInfo.ModTime()
	s.shaderDeps = statShaderDeps(displayDeps)
	warnUndirectedUniforms(newUniforms)
	if s.StatePath != "" {
		s.stateShader = newStateShader
		s.stateUniforms = newStateUniforms
		s.stateMtime = stateInfo.ModTime()
		s.stateDeps = statShaderDeps(stateDeps)
		warnUndirectedUniforms(newStateUniforms)
	}
	s.setShaderUniforms(newUniforms) // also recomputes traits over both lists

	s.rebuildControlsPreservingValues()
	s.shaderErr = ""
	s.shaderStatus = "Shader reloaded " + time.Now().Format("15:04:05")
	fmt.Println(s.shaderStatus)
	s.MarkDirty()
}

func (s *Sketch) reportShaderReloadErr(msg string) {
	s.shaderErr = msg
	fmt.Println(s.shaderErr)
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
