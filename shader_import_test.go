package sketchy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// writeLib creates dir/name.kage under root and returns root.
func writeLib(t *testing.T, root, name, src string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name)+".kage")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

const sdfLib = `package sdf

const Tau = 6.283185307179586

func Circle(p vec2, r float) float {
	return length(p) - r
}
`

func TestResolveShaderImportsNoImports(t *testing.T) {
	src := []byte(`//kage:unit pixels

package main

func Fragment(dstPos vec4) vec4 {
	return vec4(1)
}
`)
	got, deps, err := resolveShaderImports(src, "fragment.kage", t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(got) != string(src) {
		t.Errorf("source without imports should pass through unchanged, got:\n%s", got)
	}
	if len(deps) != 0 {
		t.Errorf("want no deps, got %v", deps)
	}
}

func TestResolveShaderImportsCompiles(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	d := sdf.Circle(dstPos.xy, 40.0)
	return vec4(vec3(step(0.0, -d))*sdf.Tau, 1.0)
}
`)
	merged, deps, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(deps) != 1 || !strings.HasSuffix(deps[0], filepath.Join("lib", "sdf.kage")) {
		t.Errorf("want lib/sdf.kage as the only dep, got %v", deps)
	}
	if _, err := ebiten.NewShader(merged); err != nil {
		t.Fatalf("merged source failed to compile: %v\n--- merged ---\n%s", err, merged)
	}
	if strings.Contains(string(merged), "sdf.Circle") {
		t.Error("qualified reference was not flattened")
	}
	if !strings.Contains(string(merged), mangle("sdf", "Circle")) {
		t.Error("mangled name missing from merged source")
	}
}

// The point of splicing rather than re-printing: every line of the sketch must
// still be at the line the author sees in their editor.
func TestResolveShaderImportsPreservesMainPositions(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(nope(sdf.Tau))
}
`)
	merged, _, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	_, err = ebiten.NewShader(merged)
	if err == nil {
		t.Fatal("expected a compile error for the undefined identifier")
	}
	t.Logf("error: %v", err)
	if !strings.Contains(err.Error(), "fragment.kage:8:") {
		t.Errorf("want a position in fragment.kage line 8, got: %v", err)
	}
}

func TestResolveShaderImportsMapsLibraryPositions(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", `package sdf

func Circle(p vec2, r float) float {
	return alsoNope(p) - r
}
`)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0))
}
`)
	merged, _, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	_, err = ebiten.NewShader(merged)
	if err == nil {
		t.Fatal("expected a compile error from the library")
	}
	t.Logf("error: %v", err)
	want := filepath.Join(dir, "lib", "sdf.kage") + ":4:"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("want a position at %s, got: %v", want, err)
	}
}

func TestResolveShaderImportsSearchPathOrder(t *testing.T) {
	global := t.TempDir()
	writeLib(t, global, "sdf", `package sdf

func Circle(p vec2, r float) float {
	return 1.0
}
`)
	local := t.TempDir()
	writeLib(t, local, "lib/sdf", sdfLib) // shadows the global copy
	t.Setenv("SKETCHY_KAGE_PATH", "")

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0) * sdf.Tau)
}
`)
	// sdf.Tau exists only in the local copy, so resolving it proves which won.
	merged, deps, err := resolveShaderImports(src, "fragment.kage", local)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.HasPrefix(deps[0], local) {
		t.Errorf("local library should shadow the global one, got %v", deps)
	}
	if _, err := ebiten.NewShader(merged); err != nil {
		t.Fatalf("merged source failed to compile: %v", err)
	}
}

func TestResolveShaderImportsEnvSearchPath(t *testing.T) {
	shared := t.TempDir()
	writeLib(t, shared, "sdf", sdfLib)
	t.Setenv("SKETCHY_KAGE_PATH", shared)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0))
}
`)
	_, deps, err := resolveShaderImports(src, "fragment.kage", t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(deps) != 1 || !strings.HasPrefix(deps[0], shared) {
		t.Errorf("want the library from SKETCHY_KAGE_PATH, got %v", deps)
	}
}

func TestResolveShaderImportsTransitive(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/consts", `package consts

const Tau = 6.283185307179586
`)
	writeLib(t, dir, "lib/sdf", `package sdf

import "consts"

func Ring(p vec2) float {
	return length(p) - consts.Tau
}
`)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Ring(dstPos.xy))
}
`)
	merged, deps, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(deps) != 2 {
		t.Errorf("want both libraries as deps, got %v", deps)
	}
	if _, err := ebiten.NewShader(merged); err != nil {
		t.Fatalf("merged source failed to compile: %v\n--- merged ---\n%s", err, merged)
	}
}

// Two libraries defining the same name must coexist; that is what the
// per-package prefix buys over a flat concatenation.
func TestResolveShaderImportsNameCollision(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/a", `package a

func Shape(p vec2) float {
	return length(p) - 1.0
}
`)
	writeLib(t, dir, "lib/b", `package b

func Shape(p vec2) float {
	return abs(p.x) - 1.0
}
`)

	src := []byte(`//kage:unit pixels

package main

import (
	"a"
	"b"
)

func Fragment(dstPos vec4) vec4 {
	return vec4(a.Shape(dstPos.xy) + b.Shape(dstPos.xy))
}
`)
	merged, _, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if _, err := ebiten.NewShader(merged); err != nil {
		t.Fatalf("merged source failed to compile: %v\n--- merged ---\n%s", err, merged)
	}
}

// A swizzle must survive untouched even when a local variable is named like a
// package member access would look.
func TestResolveShaderImportsLeavesSwizzlesAlone(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)

	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	p := dstPos.xy
	q := p.yx
	return vec4(sdf.Circle(q, 1.0), p.x, q.y, 1.0)
}
`)
	merged, _, err := resolveShaderImports(src, "fragment.kage", dir)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	for _, swizzle := range []string{"dstPos.xy", "p.yx", "p.x", "q.y"} {
		if !strings.Contains(string(merged), swizzle) {
			t.Errorf("swizzle %q was rewritten", swizzle)
		}
	}
	if _, err := ebiten.NewShader(merged); err != nil {
		t.Fatalf("merged source failed to compile: %v", err)
	}
}

// Editing a library must live-reload the sketch even though neither shader
// file itself was touched.
func TestShaderImportsTrackedForReload(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)

	s := newTestSketch(200, 100, nil)
	s.workDir = dir
	src := []byte(`//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0))
}
`)
	if err := s.applyShaderSource(src); err != nil {
		t.Fatalf("applyShaderSource: %v", err)
	}
	if len(s.shaderDeps) != 1 {
		t.Fatalf("want the library tracked as a dep, got %v", s.shaderDeps)
	}
	if shaderDepsChanged(s.shaderDeps) {
		t.Fatal("deps should be unchanged immediately after loading")
	}

	lib := s.shaderDeps[0].path
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(lib, future, future); err != nil {
		t.Fatal(err)
	}
	if !shaderDepsChanged(s.shaderDeps) {
		t.Error("touching an imported library should mark the deps changed")
	}

	// A library that disappears counts as changed, so the reload runs and
	// reports the real error rather than rendering stale output forever.
	if err := os.Remove(lib); err != nil {
		t.Fatal(err)
	}
	if !shaderDepsChanged(s.shaderDeps) {
		t.Error("a removed library should mark the deps changed")
	}
}

// The full reload path: editing only a library, leaving the shader file
// untouched, must recompile and swap in the new shader.
func TestCheckShaderReloadOnLibraryEdit(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)
	shaderPath := filepath.Join(dir, "fragment.kage")
	src := `//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0))
}
`
	if err := os.WriteFile(shaderPath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	s := newTestSketch(200, 100, nil)
	s.workDir = dir
	s.ShaderPath = shaderPath
	if err := s.applyShaderSource([]byte(src)); err != nil {
		t.Fatalf("applyShaderSource: %v", err)
	}
	info, err := os.Stat(shaderPath)
	if err != nil {
		t.Fatal(err)
	}
	s.shaderMtime = info.ModTime()
	before := s.shader

	// A poll tick with nothing touched must not reload.
	s.Tick = shaderReloadPollTicks
	s.checkShaderReload()
	if s.shaderStatus != "" {
		t.Fatalf("reloaded without any change: %q", s.shaderStatus)
	}

	// Edit only the library.
	writeLib(t, dir, "lib/sdf", strings.Replace(sdfLib, "length(p) - r", "length(p) - r*2.0", 1))
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(dir, "lib", "sdf.kage"), future, future); err != nil {
		t.Fatal(err)
	}

	s.Tick = 2 * shaderReloadPollTicks
	s.checkShaderReload()
	if s.shaderErr != "" {
		t.Fatalf("reload reported an error: %s", s.shaderErr)
	}
	if !strings.HasPrefix(s.shaderStatus, "Shader reloaded") {
		t.Fatalf("want a reload, got status %q", s.shaderStatus)
	}
	if s.shader == before {
		t.Error("shader was not replaced")
	}
}

// A broken library must leave the last good shader rendering, with the error
// surfaced rather than the sketch going blank.
func TestCheckShaderReloadKeepsLastGoodShader(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)
	shaderPath := filepath.Join(dir, "fragment.kage")
	src := `//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	return vec4(sdf.Circle(dstPos.xy, 1.0))
}
`
	if err := os.WriteFile(shaderPath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	s := newTestSketch(200, 100, nil)
	s.workDir = dir
	s.ShaderPath = shaderPath
	if err := s.applyShaderSource([]byte(src)); err != nil {
		t.Fatalf("applyShaderSource: %v", err)
	}
	info, err := os.Stat(shaderPath)
	if err != nil {
		t.Fatal(err)
	}
	s.shaderMtime = info.ModTime()
	good := s.shader

	writeLib(t, dir, "lib/sdf", strings.Replace(sdfLib, "length(p)", "lenth(p)", 1))
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(dir, "lib", "sdf.kage"), future, future); err != nil {
		t.Fatal(err)
	}

	s.Tick = shaderReloadPollTicks
	s.checkShaderReload()
	if s.shaderErr == "" {
		t.Fatal("want an error from the broken library")
	}
	if !strings.Contains(s.shaderErr, "sdf.kage:") {
		t.Errorf("error should point at the library file, got: %s", s.shaderErr)
	}
	if s.shader != good {
		t.Error("last good shader should still be active after a failed reload")
	}
}

// A sketch with no imports must not gain any reload deps, so the common case
// is completely unaffected.
func TestShaderWithoutImportsHasNoDeps(t *testing.T) {
	s := newTestSketch(200, 100, nil)
	s.workDir = t.TempDir()
	err := s.applyShaderSource([]byte(`//kage:unit pixels

package main

func Fragment(dstPos vec4) vec4 {
	return vec4(1)
}
`))
	if err != nil {
		t.Fatalf("applyShaderSource: %v", err)
	}
	if len(s.shaderDeps) != 0 {
		t.Errorf("want no deps, got %v", s.shaderDeps)
	}
}

func TestResolveShaderImportsErrors(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib/sdf", sdfLib)
	writeLib(t, dir, "lib/withvar", `package withvar

var Amount float

func Scale(p vec2) vec2 {
	return p * Amount
}
`)
	writeLib(t, dir, "lib/wrongpkg", `package different

func F() float {
	return 1.0
}
`)
	writeLib(t, dir, "lib/cyc1", `package cyc1

import "cyc2"

func A() float { return cyc2.B() }
`)
	writeLib(t, dir, "lib/cyc2", `package cyc2

import "cyc1"

func B() float { return cyc1.A() }
`)

	const header = "//kage:unit pixels\n\npackage main\n\n"

	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "missing import lists the paths tried",
			src:  header + "import \"nope\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(nope.X) }\n",
			want: "not found; tried:",
		},
		{
			name: "package-level var in a library is rejected",
			src:  header + "import \"withvar\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(withvar.Scale(dstPos.xy), 0, 1) }\n",
			want: "cannot declare package-level variables",
		},
		{
			name: "package name must match the import path",
			src:  header + "import \"wrongpkg\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(wrongpkg.F()) }\n",
			want: "declares package different",
		},
		{
			name: "import cycle is reported with the chain",
			src:  header + "import \"cyc1\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(cyc1.A()) }\n",
			want: "import cycle:",
		},
		{
			name: "undefined member of an imported package",
			src:  header + "import \"sdf\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(sdf.Square(dstPos.xy)) }\n",
			want: `Square is not declared in imported package "sdf"`,
		},
		{
			name: "import alias is rejected",
			src:  header + "import s \"sdf\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(s.Circle(dstPos.xy, 1.0)) }\n",
			want: "import aliases are not supported",
		},
		{
			name: "local declaration shadowing a package name",
			src:  header + "import \"sdf\"\n\nfunc Fragment(dstPos vec4) vec4 {\n\tsdf := dstPos.xy\n\treturn vec4(sdf, 0, 1)\n}\n",
			want: "shadows the imported package",
		},
		{
			name: "parent-relative import path",
			src:  header + "import \"../escape\"\n\nfunc Fragment(dstPos vec4) vec4 { return vec4(1) }\n",
			want: `must not contain ".." elements`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := resolveShaderImports([]byte(tt.src), "fragment.kage", dir)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("want error containing %q, got: %v", tt.want, err)
			}
		})
	}
}
