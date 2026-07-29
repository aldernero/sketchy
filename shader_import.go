package sketchy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Kage has no import mechanism of its own — internal/shader/shader.go rejects
// any import decl with "import is forbidden" — so sketchy resolves imports
// itself and hands ebiten.NewShader a single flat source. See
// hajimehoshi/ebiten#3439 for the upstream feature this anticipates.
//
// The transform is a byte splice, never a re-print: every removal blanks its
// bytes to spaces (keeping newlines) and every rename is length-agnostic but
// stays on one line, so line numbers never move. The merged source then
// carries /*line file:n:1*/ directives, which go/parser honours, so compile
// errors report positions in the file the author actually wrote — including
// the column, which the //line form would drop.
//
// One constraint from ebiten: //kage:unit must stay alone on its own line.
// reUnit (internal/graphics/shader.go) is anchored, and a missed match makes
// ebiten fall back to texels mode, where convertToPixels re-prints the AST and
// discards every comment — silently destroying the line directives. So the
// main file's directive goes at the head of the line *after* the unit
// directive, never on it.

// shaderImportPrefix namespaces every identifier lifted out of a library, so
// two libraries can define the same function name without colliding. The
// leading __ matches the convention ebiten uses for its own generated bridge
// identifiers (__vertex, __texelAt), which are likewise flattened into the
// user's namespace.
const shaderImportPrefix = "__sk_"

// shaderLibExt is the file extension every importable library carries.
const shaderLibExt = ".kage"

// kageUnitLine matches a lone //kage:unit directive line, mirroring reUnit in
// ebiten's internal/graphics so we agree with it on which line is the unit
// directive.
var kageUnitLine = regexp.MustCompile(`^[ \t\r]*//kage:unit\s+[^ \t\r\n]+[ \t\r]*$`)

// shaderSearchPaths returns the directories searched for an imported library,
// in order: $SKETCHY_KAGE_PATH entries first (if set), then the sketch's own
// directory, then its lib/ subdirectory, then the user-wide library. A
// sketch-local file therefore shadows a global one, so a library can be
// forked and iterated on inside one sketch before being promoted.
func shaderSearchPaths(workDir string) []string {
	var dirs []string
	if env := os.Getenv("SKETCHY_KAGE_PATH"); env != "" {
		for _, dir := range filepath.SplitList(env) {
			if dir != "" {
				dirs = append(dirs, dir)
			}
		}
	}
	dirs = append(dirs, workDir, filepath.Join(workDir, "lib"))
	if cfg, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(cfg, "sketchy", "kage"))
	}
	return dirs
}

// kageFile is one parsed Kage source, retained with the FileSet that produced
// it so byte offsets can be recovered from token positions.
type kageFile struct {
	name string // path as it should appear in error messages
	src  []byte
	fset *token.FileSet
	ast  *ast.File
	// imports maps the local package name to its import path, in source order.
	imports map[string]string
	order   []string
}

func parseKageFile(name string, src []byte) (*kageFile, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}
	kf := &kageFile{name: name, src: src, fset: fset, ast: f, imports: map[string]string{}}
	for _, spec := range f.Imports {
		if spec.Name != nil {
			return nil, fmt.Errorf("%s: import aliases are not supported (%s)",
				kf.pos(spec.Pos()), spec.Name.Name)
		}
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: malformed import path %s", kf.pos(spec.Pos()), spec.Path.Value)
		}
		if err := validateImportPath(path); err != nil {
			return nil, fmt.Errorf("%s: %w", kf.pos(spec.Pos()), err)
		}
		name := path[strings.LastIndex(path, "/")+1:]
		if prev, dup := kf.imports[name]; dup {
			return nil, fmt.Errorf("%s: imports %q and %q both bind the name %q",
				kf.pos(spec.Pos()), prev, path, name)
		}
		kf.imports[name] = path
		kf.order = append(kf.order, path)
	}
	return kf, nil
}

func (f *kageFile) pos(p token.Pos) string { return f.fset.Position(p).String() }

func (f *kageFile) offset(p token.Pos) int { return f.fset.Position(p).Offset }

func validateImportPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty import path")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
		return fmt.Errorf("import path %q must be relative", path)
	}
	for _, elem := range strings.Split(path, "/") {
		if elem == "" || elem == "." || elem == ".." {
			return fmt.Errorf("import path %q must not contain %q elements", path, elem)
		}
	}
	return nil
}

// edit is a byte-range replacement in a source file.
type edit struct {
	text       string
	start, end int
}

// blankEdit replaces a range with spaces, preserving line breaks so that
// removing a declaration never shifts any line number.
func blankEdit(src []byte, start, end int) edit {
	var b strings.Builder
	for _, c := range src[start:end] {
		if c == '\n' || c == '\r' {
			b.WriteByte(c)
		} else {
			b.WriteByte(' ')
		}
	}
	return edit{start: start, end: end, text: b.String()}
}

func applyEdits(src []byte, edits []edit) []byte {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var out []byte
	prev := 0
	for _, e := range edits {
		if e.start < prev {
			continue // overlapping edit; the earlier one wins
		}
		out = append(out, src[prev:e.start]...)
		out = append(out, e.text...)
		prev = e.end
	}
	return append(out, src[prev:]...)
}

// mangle is the flattened name an imported declaration takes in the merged
// source. The import path is sanitised rather than indexed so that names
// remain legible if they ever surface in a raw ebiten error.
func mangle(importPath, name string) string {
	return shaderImportPrefix + strings.NewReplacer("/", "_", "-", "_", ".", "_").Replace(importPath) + "_" + name
}

// resolvedLib is one library after rewriting, ready to be appended.
type resolvedLib struct {
	importPath string
	file       string // filesystem path, for reload watching and error messages
	src        []byte // rewritten, line-preserving
}

type importResolver struct {
	libs map[string]*resolvedLib
	// declared records each library's package-scope names, so a qualified
	// reference can be validated against the library that must define it.
	declared map[string]map[string]bool
	// mangled guards against two libraries flattening to the same name.
	mangled  map[string]string
	dirs     []string
	order    []string // resolution order, dependencies first
	visiting []string // cycle-detection stack
}

// resolveShaderImports rewrites src into a single flat Kage source with every
// imported library inlined, and returns the merged source plus the filesystem
// paths of the libraries used, so the caller can watch them for live reload.
// Sources with no imports are returned unchanged, so a sketch that never
// imports anything is completely unaffected.
func resolveShaderImports(src []byte, name, workDir string) ([]byte, []string, error) {
	main, err := parseKageFile(name, src)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing shader: %w", err)
	}
	if len(main.order) == 0 {
		return src, nil, nil
	}

	r := &importResolver{
		dirs:     shaderSearchPaths(workDir),
		libs:     map[string]*resolvedLib{},
		declared: map[string]map[string]bool{},
		mangled:  map[string]string{},
	}
	for _, path := range main.order {
		if err := r.resolve(path); err != nil {
			return nil, nil, err
		}
	}

	mainSrc, err := rewriteKageFile(main, "", r.declared)
	if err != nil {
		return nil, nil, err
	}

	var buf []byte
	buf = append(buf, insertMainLineDirective(mainSrc, name)...)
	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	deps := make([]string, 0, len(r.order))
	for _, path := range r.order {
		lib := r.libs[path]
		deps = append(deps, lib.file)
		// The directive binds the byte immediately after it, so the library's
		// first byte is reported as line 1, column 1 of its own file.
		buf = append(buf, "\n/*line "...)
		buf = append(buf, lib.file...)
		buf = append(buf, ":1:1*/"...)
		buf = append(buf, lib.src...)
		if len(lib.src) > 0 && lib.src[len(lib.src)-1] != '\n' {
			buf = append(buf, '\n')
		}
	}
	return buf, deps, nil
}

func (r *importResolver) resolve(path string) error {
	if _, done := r.libs[path]; done {
		return nil
	}
	for i, v := range r.visiting {
		if v == path {
			return fmt.Errorf("import cycle: %s", strings.Join(append(append([]string{}, r.visiting[i:]...), path), " -> "))
		}
	}

	file, err := r.find(path)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading imported shader %q: %w", path, err)
	}
	lib, err := parseKageFile(file, src)
	if err != nil {
		return fmt.Errorf("parsing imported shader: %w", err)
	}

	want := path[strings.LastIndex(path, "/")+1:]
	if lib.ast.Name.Name != want {
		return fmt.Errorf("%s: import %q expects package %s but the file declares package %s",
			lib.pos(lib.ast.Name.Pos()), path, want, lib.ast.Name.Name)
	}

	r.visiting = append(r.visiting, path)
	for _, dep := range lib.order {
		if err := r.resolve(dep); err != nil {
			return err
		}
	}
	r.visiting = r.visiting[:len(r.visiting)-1]

	rewritten, err := rewriteKageFile(lib, path, r.declared)
	if err != nil {
		return err
	}
	names := map[string]bool{}
	for _, name := range topLevelNames(lib.ast) {
		names[name] = true
		m := mangle(path, name)
		if prev, dup := r.mangled[m]; dup {
			return fmt.Errorf("imports %q and %q both flatten %s to %s", prev, path, name, m)
		}
		r.mangled[m] = path
	}
	r.declared[path] = names

	r.libs[path] = &resolvedLib{importPath: path, file: file, src: rewritten}
	r.order = append(r.order, path)
	return nil
}

// find locates the file backing an import path, reporting every directory
// tried when it misses — the search path is configurable, so a bare "not
// found" would be unhelpfully ambiguous.
func (r *importResolver) find(path string) (string, error) {
	rel := filepath.FromSlash(path) + shaderLibExt
	var tried []string
	for _, dir := range r.dirs {
		candidate := filepath.Join(dir, rel)
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
		tried = append(tried, candidate)
	}
	return "", fmt.Errorf("imported shader %q not found; tried:\n  %s", path, strings.Join(tried, "\n  "))
}

// topLevelNames lists the package-scope declarations a library exports into
// the merged namespace. Package-level var is deliberately absent: it is
// rejected outright by rewriteKageFile.
func topLevelNames(f *ast.File) []string {
	var names []string
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil && d.Name.Name != "_" {
				names = append(names, d.Name.Name)
			}
		case *ast.GenDecl:
			if d.Tok != token.CONST && d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name != nil && s.Name.Name != "_" {
						names = append(names, s.Name.Name)
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name != "_" {
							names = append(names, n.Name)
						}
					}
				}
			}
		}
	}
	return names
}

// rewriteKageFile blanks a file's package clause, imports and //kage:unit
// directive, rewrites qualified references to imported packages, and — when
// ownPath is non-empty, i.e. the file is a library — prefixes its own
// top-level declarations. declared gives the package-scope names of every
// import, so a reference to something a library does not define is caught
// here rather than surfacing as a mangled "undefined" from ebiten. The result
// has exactly the same line structure as the input.
func rewriteKageFile(f *kageFile, ownPath string, declared map[string]map[string]bool) ([]byte, error) {
	isLib := ownPath != ""
	var edits []edit

	if isLib {
		// The package clause would be a second one in the merged source.
		edits = append(edits, blankEdit(f.src, f.offset(f.ast.Package), f.offset(f.ast.Name.End())))
		// ebiten allows at most one //kage:unit per shader.
		for _, group := range f.ast.Comments {
			for _, c := range group.List {
				if kageUnitLine.MatchString(c.Text) {
					edits = append(edits, blankEdit(f.src, f.offset(c.Pos()), f.offset(c.End())))
				}
			}
		}
	}

	for _, decl := range f.ast.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		switch d.Tok {
		case token.IMPORT:
			edits = append(edits, blankEdit(f.src, f.offset(d.Pos()), f.offset(d.End())))
		case token.VAR:
			if isLib {
				// Every package-scope var in Kage becomes a uniform
				// (internal/shader/shader.go), and only the main source's
				// uniforms are bound to controls. Silently dropping these —
				// as the upstream prototype does — turns a mistake into a
				// baffling "undefined" at the use site, so reject it here.
				return nil, fmt.Errorf("%s: a shader library cannot declare package-level variables (every one would become a uniform); use a const, or a zero-argument function for composite values",
					f.pos(d.Pos()))
			}
		}
	}

	renames := map[string]string{}
	if isLib {
		for _, name := range topLevelNames(f.ast) {
			renames[name] = mangle(ownPath, name)
		}
	}

	// Guard against a local declaration shadowing an imported package name:
	// `sdf.Circle` and a variable named sdf cannot coexist, and resolving it
	// silently either way would be worse than saying so.
	for _, decl := range f.ast.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		var shadow error
		ast.Inspect(fn, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok || id.Obj == nil || shadow != nil {
				return true
			}
			if _, isPkg := f.imports[id.Name]; isPkg && id.Obj.Pos() == id.Pos() {
				shadow = fmt.Errorf("%s: %q shadows the imported package of the same name", f.pos(id.Pos()), id.Name)
			}
			return true
		})
		if shadow != nil {
			return nil, shadow
		}
	}

	var walkErr error
	ast.Inspect(f.ast, func(n ast.Node) bool {
		if walkErr != nil {
			return false
		}
		switch e := n.(type) {
		case *ast.SelectorExpr:
			// Kage reads every selector as a swizzle (internal/shader/expr.go),
			// so a package-qualified name must be flattened before it gets
			// there. A package name never resolves to a declaration, hence
			// Obj == nil.
			id, ok := e.X.(*ast.Ident)
			if !ok || id.Obj != nil {
				return true
			}
			path, ok := f.imports[id.Name]
			if !ok {
				return true
			}
			if names, known := declared[path]; known && !names[e.Sel.Name] {
				walkErr = fmt.Errorf("%s: %s is not declared in imported package %q", f.pos(e.Sel.Pos()), e.Sel.Name, path)
				return false
			}
			edits = append(edits, edit{start: f.offset(e.X.Pos()), end: f.offset(e.Sel.End()), text: mangle(path, e.Sel.Name)})
			return false // do not descend; X is a package name, not an expression
		case *ast.Ident:
			if to, ok := renames[e.Name]; ok && isTopLevelRef(f, e) {
				edits = append(edits, edit{start: f.offset(e.Pos()), end: f.offset(e.End()), text: to})
			}
		}
		return true
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return applyEdits(f.src, edits), nil
}

// isTopLevelRef reports whether an identifier refers to a package-scope
// declaration of this file rather than to something local that shadows it.
// go/parser's object resolution links every reference to its declaration, and
// package-scope objects are the ones recorded in the file's own scope.
func isTopLevelRef(f *kageFile, id *ast.Ident) bool {
	if id.Obj == nil {
		// Unresolved: a builtin, a uniform, or a forward reference to a
		// package-scope declaration. Only the last can be in renames.
		return true
	}
	obj := f.ast.Scope.Lookup(id.Name)
	return obj != nil && obj == id.Obj
}

// insertMainLineDirective puts a /*line*/ directive at the head of the line
// following the //kage:unit directive, so positions in the sketch's own source
// keep their original file, line and column. It must not land on the unit line
// itself: ebiten's matcher is anchored, and a miss silently reroutes the
// source through convertToPixels, which re-prints the AST and drops all
// comments — including this directive.
func insertMainLineDirective(src []byte, name string) []byte {
	lines := strings.SplitAfter(string(src), "\n")
	insertAt := 0 // no unit directive: cover the file from its first line
	for i, line := range lines {
		if kageUnitLine.MatchString(strings.TrimRight(line, "\n")) {
			insertAt = i + 1
			break
		}
	}
	if insertAt >= len(lines) {
		return src
	}
	directive := fmt.Sprintf("/*line %s:%d:1*/", name, insertAt+1)
	var b strings.Builder
	for i, line := range lines {
		if i == insertAt {
			b.WriteString(directive)
		}
		b.WriteString(line)
	}
	return []byte(b.String())
}
