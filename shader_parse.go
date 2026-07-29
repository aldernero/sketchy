package sketchy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// uniformKind is the Kage type of a shader uniform, as far as control
// mapping is concerned.
type uniformKind int

const (
	ukFloat uniformKind = iota
	ukInt
	ukVec2
	ukVec3
	ukVec4
	ukOther // mat2/mat3/mat4, arrays, … — usable only via ExtraUniforms
)

var uniformKindNames = map[string]uniformKind{
	"float": ukFloat,
	"int":   ukInt,
	"vec2":  ukVec2,
	"vec3":  ukVec3,
	"vec4":  ukVec4,
}

func (k uniformKind) String() string {
	for name, kind := range uniformKindNames {
		if kind == k {
			return name
		}
	}
	return "other"
}

// shaderUniform is one uniform declared in the Kage source's top-level var
// block, with its optional //sketchy: directive.
type shaderUniform struct {
	Directive *uniformDirective // nil = no directive
	Name      string
	Kind      uniformKind
}

// uniformDirective is a parsed, validated //sketchy: comment.
type uniformDirective struct {
	// seenKeys records which keys the directive spelled out, so validation
	// can apply kind-aware defaults only for omitted ones.
	seenKeys      map[string]bool
	Control       string // "slider" | "checkbox" | "color" | "dropdown" | "none"
	DefaultHex    string // color
	Folder, Label string
	Options       []string

	Min, Max, Default, Step float64 // slider (and checkbox Default 0/1)
	Digits                  int     // slider decimal digits (-1 = derive from step)
	DefaultIdx              int     // dropdown
}

// controlName is the panel display name (Label override or uniform name).
func (u *shaderUniform) controlName() string {
	if u.Directive != nil && u.Directive.Label != "" {
		return u.Directive.Label
	}
	return u.Name
}

// builtinUniformKinds is the reserved set auto-provided by sketchy when
// declared in the shader (matched by name and type, no directive needed).
var builtinUniformKinds = map[string]uniformKind{
	"Time":       ukFloat, // seconds = Tick/60; declaring it auto-animates
	"Tick":       ukInt,   // raw tick count; also auto-animates
	"Resolution": ukVec2,  // render-target pixel size
	"Mouse":      ukVec2,  // cursor in canvas coordinates
	"Seed":       ukFloat, // RandomSeed
	"Substep":    ukInt,   // 0..Steps-1 within a tick's state-pass loop
}

func isBuiltinUniform(u shaderUniform) bool {
	k, ok := builtinUniformKinds[u.Name]
	return ok && k == u.Kind
}

var hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// parseShaderUniforms extracts the top-level var declarations from Kage
// source (which is valid Go syntax) along with their //sketchy: directives.
func parseShaderUniforms(src []byte) ([]shaderUniform, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "shader.kage", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing shader: %w", err)
	}
	var out []shaderUniform
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			kind := ukOther
			if id, ok := vs.Type.(*ast.Ident); ok {
				if k, known := uniformKindNames[id.Name]; known {
					kind = k
				}
			}
			directive, err := directiveFromComment(vs.Comment)
			if err != nil {
				return nil, fmt.Errorf("uniform %s: %w", vs.Names[0].Name, err)
			}
			if directive != nil && len(vs.Names) > 1 {
				return nil, fmt.Errorf("//sketchy: directive on multi-name declaration %q — declare one uniform per line", vs.Names[0].Name)
			}
			for _, name := range vs.Names {
				u := shaderUniform{Name: name.Name, Kind: kind, Directive: directive}
				if directive != nil {
					if err := validateDirective(&u); err != nil {
						return nil, fmt.Errorf("uniform %s: %w", u.Name, err)
					}
				}
				out = append(out, u)
			}
		}
	}
	return out, nil
}

// directiveFromComment parses a trailing //sketchy:<control> key=value …
// comment. Returns nil when the comment group has no sketchy directive.
func directiveFromComment(cg *ast.CommentGroup) (*uniformDirective, error) {
	if cg == nil {
		return nil, nil
	}
	for _, c := range cg.List {
		text := strings.TrimPrefix(c.Text, "//")
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, "sketchy:") {
			continue
		}
		return parseDirective(strings.TrimPrefix(text, "sketchy:"))
	}
	return nil, nil
}

func parseDirective(text string) (*uniformDirective, error) {
	fields, err := splitDirectiveFields(text)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty //sketchy: directive")
	}
	d := &uniformDirective{Control: fields[0], Digits: -1}
	switch d.Control {
	case "slider", "checkbox", "color", "dropdown", "none":
	default:
		return nil, fmt.Errorf("unknown //sketchy: control %q (want slider, checkbox, color, dropdown, or none)", d.Control)
	}

	seen := map[string]bool{}
	for _, kv := range fields[1:] {
		key, val, ok := strings.Cut(kv, "=")
		if !ok || val == "" {
			return nil, fmt.Errorf("malformed directive token %q (want key=value)", kv)
		}
		// Strip optional quotes around values (label="Use sine palette").
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate directive key %q", key)
		}
		seen[key] = true
		var err error
		switch key {
		case "min":
			d.Min, err = strconv.ParseFloat(val, 64)
		case "max":
			d.Max, err = strconv.ParseFloat(val, 64)
		case "step":
			d.Step, err = strconv.ParseFloat(val, 64)
		case "digits":
			d.Digits, err = strconv.Atoi(val)
		case "default":
			err = parseDirectiveDefault(d, val)
		case "options":
			d.Options = strings.Split(val, "|")
		case "folder":
			d.Folder = val
		case "label":
			d.Label = val
		default:
			return nil, fmt.Errorf("unknown directive key %q", key)
		}
		if err != nil {
			return nil, fmt.Errorf("directive key %s=%q: %w", key, val, err)
		}
	}
	// Remember which keys were given so validation can fill kind-aware defaults.
	d.seenKeys = seen
	return d, nil
}

// splitDirectiveFields splits a //sketchy: body on whitespace, keeping
// double-quoted values intact so label="Use sine palette" is one token.
func splitDirectiveFields(text string) ([]string, error) {
	var fields []string
	i := 0
	for i < len(text) {
		for i < len(text) && isDirectiveSpace(text[i]) {
			i++
		}
		if i >= len(text) {
			break
		}
		start := i
		for i < len(text) && !isDirectiveSpace(text[i]) {
			if text[i] == '=' && i+1 < len(text) && text[i+1] == '"' {
				i += 2 // skip ="
				for i < len(text) && text[i] != '"' {
					i++
				}
				if i >= len(text) {
					return nil, fmt.Errorf("unclosed quote in directive")
				}
				i++ // closing "
				break
			}
			i++
		}
		fields = append(fields, text[start:i])
	}
	return fields, nil
}

func isDirectiveSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// parseDirectiveDefault handles the polymorphic default= key: a number for
// sliders, 0/1/true/false for checkboxes, #hex for colors, an index for
// dropdowns. Stored in all candidate fields; validateDirective picks.
func parseDirectiveDefault(d *uniformDirective, val string) error {
	if strings.HasPrefix(val, "#") {
		d.DefaultHex = val
		return nil
	}
	switch val {
	case "true":
		d.Default = 1
		return nil
	case "false":
		d.Default = 0
		return nil
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fmt.Errorf("not a number, bool, or #hex color")
	}
	d.Default = f
	d.DefaultIdx = int(f)
	return nil
}

// shaderImageDirective is a parsed, validated standalone //sketchy:image
// comment: a source image bound to a Fragment source-image slot
// (imageSrc0At.. imageSrc3At), independent of the uniform var block since
// Kage has no sampler/image type to declare a uniform for.
type shaderImageDirective struct {
	Path string
	Slot int // 0-3
}

// parseShaderImageDirectives scans every comment in the Kage source (not
// just those attached to var decls) for standalone "//sketchy:image
// path=... [slot=N]" directives. Slots default to the next unused slot in
// appearance order when slot= is omitted; duplicate slots are an error.
func parseShaderImageDirectives(src []byte) ([]shaderImageDirective, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "shader.kage", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing shader: %w", err)
	}
	var out []shaderImageDirective
	seen := map[int]bool{}
	next := 0
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if !strings.HasPrefix(text, "sketchy:image") {
				continue
			}
			d, err := parseImageDirective(strings.TrimSpace(strings.TrimPrefix(text, "sketchy:image")))
			if err != nil {
				return nil, err
			}
			if d.Slot < 0 {
				for seen[next] {
					next++
				}
				d.Slot = next
			}
			if seen[d.Slot] {
				return nil, fmt.Errorf("//sketchy:image slot %d is bound more than once", d.Slot)
			}
			seen[d.Slot] = true
			next = d.Slot + 1
			out = append(out, *d)
		}
	}
	return out, nil
}

func parseImageDirective(text string) (*shaderImageDirective, error) {
	d := &shaderImageDirective{Slot: -1}
	for _, kv := range strings.Fields(text) {
		key, val, ok := strings.Cut(kv, "=")
		if !ok || val == "" {
			return nil, fmt.Errorf("malformed //sketchy:image token %q (want key=value)", kv)
		}
		switch key {
		case "path":
			d.Path = val
		case "slot":
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 || n > 3 {
				return nil, fmt.Errorf("//sketchy:image slot must be 0-3, got %q", val)
			}
			d.Slot = n
		default:
			return nil, fmt.Errorf("unknown //sketchy:image key %q", key)
		}
	}
	if d.Path == "" {
		return nil, fmt.Errorf("//sketchy:image requires path=")
	}
	return d, nil
}

// validateDirective checks control/type compatibility and fills defaults.
func validateDirective(u *shaderUniform) error {
	d := u.Directive
	has := func(k string) bool { return d.seenKeys[k] }
	switch d.Control {
	case "none":
		return nil
	case "slider":
		if u.Kind != ukFloat && u.Kind != ukInt {
			return fmt.Errorf("slider requires a float or int uniform, got %s", u.Kind)
		}
		if !has("max") {
			if u.Kind == ukInt {
				d.Max = 10
			} else {
				d.Max = 1
			}
		}
		if !has("min") {
			d.Min = 0
		}
		if d.Min >= d.Max {
			return fmt.Errorf("slider min (%g) must be < max (%g)", d.Min, d.Max)
		}
		if !has("step") {
			if u.Kind == ukInt {
				d.Step = 1
			} else {
				d.Step = 0.01
			}
		}
		if !has("default") {
			d.Default = d.Min
		}
		if d.Default < d.Min || d.Default > d.Max {
			return fmt.Errorf("slider default (%g) outside [%g, %g]", d.Default, d.Min, d.Max)
		}
	case "checkbox":
		if u.Kind != ukFloat && u.Kind != ukInt {
			return fmt.Errorf("checkbox requires a float or int uniform, got %s", u.Kind)
		}
		if d.Default != 0 && d.Default != 1 {
			return fmt.Errorf("checkbox default must be 0, 1, true, or false")
		}
	case "color":
		if u.Kind != ukVec3 && u.Kind != ukVec4 {
			return fmt.Errorf("color requires a vec3 or vec4 uniform, got %s", u.Kind)
		}
		if d.DefaultHex == "" {
			d.DefaultHex = "#ffffff"
		} else if !hexColorPattern.MatchString(d.DefaultHex) {
			return fmt.Errorf("color default %q is not a #rgb or #rrggbb hex color", d.DefaultHex)
		}
	case "dropdown":
		if u.Kind != ukInt {
			return fmt.Errorf("dropdown requires an int uniform, got %s", u.Kind)
		}
		if len(d.Options) == 0 {
			return fmt.Errorf("dropdown requires options=A|B|C")
		}
		if d.DefaultIdx < 0 || d.DefaultIdx >= len(d.Options) {
			return fmt.Errorf("dropdown default index %d outside options (%d)", d.DefaultIdx, len(d.Options))
		}
	}
	return nil
}
