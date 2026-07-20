package sketchy

import (
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestShaderTemplateCompiles guards the shipped shader template against
// rot: its uniforms must parse (with directives) and the Kage source must
// compile.
func TestShaderTemplateCompiles(t *testing.T) {
	src, err := os.ReadFile("cmd/sketchy/template_shader/fragment.kage")
	if err != nil {
		t.Fatal(err)
	}
	us, err := parseShaderUniforms(src)
	if err != nil {
		t.Fatal(err)
	}
	directives := 0
	hasTime := false
	for _, u := range us {
		if u.Directive != nil {
			directives++
		}
		if u.Name == "Time" && isBuiltinUniform(u) {
			hasTime = true
		}
	}
	if directives < 4 || !hasTime {
		t.Fatalf("template should demo directives and the Time builtin (got %d directives, Time=%v)", directives, hasTime)
	}
	if _, err := ebiten.NewShader(src); err != nil {
		t.Fatalf("template shader does not compile: %v", err)
	}
}

const testKageSrc = `//kage:unit pixels

package main

var (
	Zoom    float //sketchy:slider min=0.1 max=100 default=0.4 step=0.05 folder=Fractal
	MaxIter int   //sketchy:slider min=16 max=1024 default=256 step=16
	ColorA  vec3  //sketchy:color default=#cc3311
	Tint    vec4  //sketchy:color
	Invert  float //sketchy:checkbox default=true label=Inverted
	Mode    int   //sketchy:dropdown options=Normal|Smooth|Bands default=1
	Aux     vec2  //sketchy:none
	Plain   float
	Time    float
	Mouse   vec2
)

func Fragment(dstPos vec4) vec4 {
	return vec4(ColorA, 1.0)
}
`

func TestParseShaderUniforms(t *testing.T) {
	us, err := parseShaderUniforms([]byte(testKageSrc))
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]shaderUniform{}
	for _, u := range us {
		byName[u.Name] = u
	}
	if len(us) != 10 {
		t.Fatalf("parsed %d uniforms, want 10", len(us))
	}

	zoom := byName["Zoom"]
	if zoom.Kind != ukFloat || zoom.Directive == nil || zoom.Directive.Control != "slider" {
		t.Fatalf("Zoom parsed wrong: %+v", zoom)
	}
	d := zoom.Directive
	if d.Min != 0.1 || d.Max != 100 || d.Default != 0.4 || d.Step != 0.05 || d.Folder != "Fractal" {
		t.Fatalf("Zoom directive fields wrong: %+v", d)
	}

	mi := byName["MaxIter"]
	if mi.Kind != ukInt || mi.Directive.Default != 256 || mi.Directive.Step != 16 {
		t.Fatalf("MaxIter wrong: %+v", mi.Directive)
	}

	ca := byName["ColorA"]
	if ca.Kind != ukVec3 || ca.Directive.Control != "color" || ca.Directive.DefaultHex != "#cc3311" {
		t.Fatalf("ColorA wrong: %+v", ca.Directive)
	}
	if byName["Tint"].Directive.DefaultHex != "#ffffff" {
		t.Fatalf("Tint should default to white, got %q", byName["Tint"].Directive.DefaultHex)
	}

	inv := byName["Invert"]
	if inv.Directive.Control != "checkbox" || inv.Directive.Default != 1 || inv.controlName() != "Inverted" {
		t.Fatalf("Invert wrong: %+v", inv.Directive)
	}

	mode := byName["Mode"]
	if mode.Directive.Control != "dropdown" || len(mode.Directive.Options) != 3 || mode.Directive.DefaultIdx != 1 {
		t.Fatalf("Mode wrong: %+v", mode.Directive)
	}

	if byName["Aux"].Directive.Control != "none" {
		t.Fatal("Aux should be control=none")
	}
	if byName["Plain"].Directive != nil {
		t.Fatal("Plain should have no directive")
	}
	if !isBuiltinUniform(byName["Time"]) || !isBuiltinUniform(byName["Mouse"]) {
		t.Fatal("Time and Mouse should be builtins")
	}
	if isBuiltinUniform(byName["Plain"]) || isBuiltinUniform(byName["Zoom"]) {
		t.Fatal("Plain/Zoom should not be builtins")
	}

	// Slider defaults for omitted keys.
	src := "package main\nvar X float //sketchy:slider\n"
	us2, err := parseShaderUniforms([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	d2 := us2[0].Directive
	if d2.Min != 0 || d2.Max != 1 || d2.Default != 0 || d2.Step != 0.01 {
		t.Fatalf("float slider defaults wrong: %+v", d2)
	}
	src = "package main\nvar N int //sketchy:slider\n"
	us3, _ := parseShaderUniforms([]byte(src))
	if d3 := us3[0].Directive; d3.Max != 10 || d3.Step != 1 {
		t.Fatalf("int slider defaults wrong: %+v", d3)
	}
}

// newTestShaderSketch builds a headless shader-mode sketch: uniforms parsed
// and controls registered, but no ebiten shader compiled (no GPU needed).
func newTestShaderSketch(t *testing.T, src string) *Sketch {
	t.Helper()
	s := newTestSketch(200, 100, nil)
	s.ShaderSrc = []byte(src)
	us, err := parseShaderUniforms([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	s.setShaderUniforms(us)
	s.rebuildControls()
	return s
}

const buildUniformsSrc = `package main

var (
	Zoom    float //sketchy:slider min=0.1 max=100 default=0.4 folder=Fractal
	MaxIter int   //sketchy:slider min=16 max=1024 default=256 step=16
	ColorA  vec3  //sketchy:color default=#ff0000
	Tint    vec4  //sketchy:color default=#00ff00
	Invert  float //sketchy:checkbox default=true
	Flip    int   //sketchy:checkbox
	Mode    int   //sketchy:dropdown options=Normal|Smooth|Bands default=1
	Aux     vec2  //sketchy:none
	Plain   float
	Time    float
	Tick    int
	Resolution vec2
	Seed    float
)
`

func TestRegisterShaderControls(t *testing.T) {
	s := newTestShaderSketch(t, buildUniformsSrc)
	if got := len(s.FloatSliders); got != 1 {
		t.Fatalf("FloatSliders = %d, want 1", got)
	}
	z := s.FloatSliders[0]
	if z.Name != "Zoom" || z.Folder != "Fractal" || z.MinVal != 0.1 || z.MaxVal != 100 || z.Val != 0.4 {
		t.Fatalf("Zoom slider wrong: %+v", z)
	}
	if got := len(s.IntSliders); got != 1 || s.IntSliders[0].Val != 256 || s.IntSliders[0].Incr != 16 {
		t.Fatalf("IntSliders wrong: %+v", s.IntSliders)
	}
	if got := len(s.Toggles); got != 2 || !s.Toggles[0].Checked || s.Toggles[1].Checked {
		t.Fatalf("Toggles wrong: %+v", s.Toggles)
	}
	// ColorA, Tint + the two builtin BG/FG pickers.
	if got := len(s.ColorPickers); got != 4 {
		t.Fatalf("ColorPickers = %d, want 4", got)
	}
	if hex := s.ColorPickers[0].GetHex(); hex != "#FF0000" {
		t.Fatalf("ColorA hex = %s", hex)
	}
	if got := len(s.Dropdowns); got != 1 || s.Dropdowns[0].Index != 1 || len(s.Dropdowns[0].Options) != 3 {
		t.Fatalf("Dropdowns wrong: %+v", s.Dropdowns)
	}
	// Plain (no directive), Aux (none), and builtins get no controls; maps
	// must resolve the generated ones.
	if _, ok := s.floatSliderControlMap[controlMapKey("Fractal", "Zoom")]; !ok {
		t.Fatal("Zoom missing from float slider map")
	}
}

func TestBuildUniforms(t *testing.T) {
	s := newTestShaderSketch(t, buildUniformsSrc)
	s.Tick = 120
	s.RandomSeed = 7

	// Nudge some control values through the maps.
	s.FloatSliders[s.floatSliderControlMap[controlMapKey("Fractal", "Zoom")]].Val = 2.5
	s.IntSliders[s.intSliderControlMap["MaxIter"]].Val = 512
	s.Dropdowns[s.dropdownControlMap["Mode"]].Index = 2

	m := s.buildUniforms(200, 100)

	if v, ok := m["Zoom"].(float64); !ok || v != 2.5 {
		t.Fatalf("Zoom = %v (%T)", m["Zoom"], m["Zoom"])
	}
	if v, ok := m["MaxIter"].(int); !ok || v != 512 {
		t.Fatalf("MaxIter = %v (%T)", m["MaxIter"], m["MaxIter"])
	}
	ca, ok := m["ColorA"].([]float32)
	if !ok || len(ca) != 3 || ca[0] != 1 || ca[1] != 0 || ca[2] != 0 {
		t.Fatalf("ColorA = %v", m["ColorA"])
	}
	tint, ok := m["Tint"].([]float32)
	if !ok || len(tint) != 4 || tint[1] != 1 || tint[3] != 1 {
		t.Fatalf("Tint = %v", m["Tint"])
	}
	if v, ok := m["Invert"].(float64); !ok || v != 1 {
		t.Fatalf("Invert = %v (%T)", m["Invert"], m["Invert"])
	}
	if v, ok := m["Flip"].(int); !ok || v != 0 {
		t.Fatalf("Flip = %v (%T)", m["Flip"], m["Flip"])
	}
	if v, ok := m["Mode"].(int); !ok || v != 2 {
		t.Fatalf("Mode = %v", m["Mode"])
	}
	if v, ok := m["Time"].(float64); !ok || v != 2.0 {
		t.Fatalf("Time = %v, want 2.0", m["Time"])
	}
	if v, ok := m["Tick"].(int); !ok || v != 120 {
		t.Fatalf("Tick = %v", m["Tick"])
	}
	res, ok := m["Resolution"].([]float32)
	if !ok || res[0] != 200 || res[1] != 100 {
		t.Fatalf("Resolution = %v", m["Resolution"])
	}
	if v, ok := m["Seed"].(float64); !ok || v != 7 {
		t.Fatalf("Seed = %v", m["Seed"])
	}
	if _, present := m["Plain"]; present {
		t.Fatal("Plain (no directive) should not be passed")
	}
	if _, present := m["Aux"]; present {
		t.Fatal("Aux (//sketchy:none) should not be passed")
	}
	if !s.shaderAnimates {
		t.Fatal("Time/Tick declared should make the shader animate")
	}

	// ExtraUniforms merges last and wins.
	s.ExtraUniforms = func(_ *Sketch) map[string]any {
		return map[string]any{"Zoom": 9.9, "Aux": []float32{1, 2}}
	}
	m = s.buildUniforms(200, 100)
	if m["Zoom"] != 9.9 {
		t.Fatalf("ExtraUniforms should override Zoom, got %v", m["Zoom"])
	}
	if _, present := m["Aux"]; !present {
		t.Fatal("ExtraUniforms should supply Aux")
	}
}

func TestShaderReloadPreservesValues(t *testing.T) {
	const srcA = `package main
var (
	Zoom float //sketchy:slider min=0 max=100 default=10
	Iter int   //sketchy:slider min=1 max=64 default=8
	Base vec3  //sketchy:color default=#000000
	Gone float //sketchy:slider
)
`
	const srcB = `package main
var (
	Zoom float //sketchy:slider min=0 max=20 default=10
	Iter int   //sketchy:slider min=1 max=64 default=8
	Base vec3  //sketchy:color default=#000000
	Glow float //sketchy:slider min=0 max=1 default=0.5
)
`
	s := newTestShaderSketch(t, srcA)
	s.FloatSliders[s.floatSliderControlMap["Zoom"]].Val = 42
	s.IntSliders[s.intSliderControlMap["Iter"]].Val = 33
	i := s.colorPickerControlMap["Base"]
	restored := NewColorPicker("Base", "#123456")
	s.ColorPickers[i] = restored

	us, err := parseShaderUniforms([]byte(srcB))
	if err != nil {
		t.Fatal(err)
	}
	s.setShaderUniforms(us)
	s.rebuildControlsPreservingValues()

	// Zoom survives but clamps to the new max.
	if v := s.FloatSliders[s.floatSliderControlMap["Zoom"]].Val; v != 20 {
		t.Fatalf("Zoom after reload = %g, want 20 (clamped)", v)
	}
	if v := s.IntSliders[s.intSliderControlMap["Iter"]].Val; v != 33 {
		t.Fatalf("Iter after reload = %d, want 33", v)
	}
	if hex := s.ColorPickers[s.colorPickerControlMap["Base"]].GetHex(); hex != "#123456" {
		t.Fatalf("Base after reload = %s, want #123456", hex)
	}
	// New control gets its default; removed control is gone.
	if v := s.FloatSliders[s.floatSliderControlMap["Glow"]].Val; v != 0.5 {
		t.Fatalf("Glow default = %g, want 0.5", v)
	}
	if _, ok := s.floatSliderControlMap["Gone"]; ok {
		t.Fatal("Gone should have been removed")
	}
	// Folder plan and maps stay consistent.
	if len(s.uiPlan) != 4 {
		t.Fatalf("uiPlan has %d entries, want 4", len(s.uiPlan))
	}
}

func TestParseShaderUniformErrors(t *testing.T) {
	cases := []struct {
		name, src, wantErr string
	}{
		{"multi-name directive", "package main\nvar A, B float //sketchy:slider\n", "one uniform per line"},
		{"dropdown without options", "package main\nvar M int //sketchy:dropdown\n", "requires options"},
		{"color on float", "package main\nvar C float //sketchy:color\n", "requires a vec3 or vec4"},
		{"unknown control", "package main\nvar X float //sketchy:knob\n", "unknown //sketchy: control"},
		{"unknown key", "package main\nvar X float //sketchy:slider speed=2\n", "unknown directive key"},
		{"min >= max", "package main\nvar X float //sketchy:slider min=5 max=1\n", "must be < max"},
		{"default outside range", "package main\nvar X float //sketchy:slider min=0 max=1 default=2\n", "outside"},
		{"checkbox bad default", "package main\nvar X float //sketchy:checkbox default=3\n", "must be 0, 1"},
		{"dropdown on float", "package main\nvar X float //sketchy:dropdown options=A|B\n", "requires an int"},
		{"malformed token", "package main\nvar X float //sketchy:slider min\n", "key=value"},
		{"not go syntax", "this is not kage\n", "parsing shader"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseShaderUniforms([]byte(tc.src))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tc.wantErr)
			}
		})
	}
}
