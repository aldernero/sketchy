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
	Name      string
	Kind      uniformKind
	Directive *uniformDirective // nil = no directive
}

// uniformDirective is a parsed, validated //sketchy: comment.
type uniformDirective struct {
	Control string // "slider" | "checkbox" | "color" | "dropdown" | "none"

	Min, Max, Default, Step float64 // slider (and checkbox Default 0/1)
	Digits                  int     // slider decimal digits (-1 = derive from step)
	DefaultHex              string  // color
	Options                 []string
	DefaultIdx              int // dropdown
	Folder, Label           string

	// seenKeys records which keys the directive spelled out, so validation
	// can apply kind-aware defaults only for omitted ones.
	seenKeys map[string]bool
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
	fields := strings.Fields(text)
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
