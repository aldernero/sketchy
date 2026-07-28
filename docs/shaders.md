# Shader sketches

Sketchy can render a sketch entirely on the GPU with an Ebitengine
[Kage](https://ebitengine.org/en/documents/shader.html) fragment shader
instead of the CPU canvas. The killer feature: **the shader file is the
single source of truth for its controls.** Declare a uniform with a
`//sketchy:` directive and sketchy creates the control-panel control,
passes its value back as a uniform every frame, live-reloads the shader
when you save the file, persists the control in snapshots, and records it
all to video — no Go plumbing per uniform.

# Quickstart

```shell
sketchy init shader myshader
sketchy run myshader
```

The generated project contains `main.go` (which you will rarely touch) and
`fragment.kage` (where everything happens). The template is an animated
palette demo — move the sliders, then edit `fragment.kage` while it runs
and watch it hot-reload.

A shader sketch is a normal `sketchy.Config` with one extra field:

```go
s := sketchy.New(sketchy.Config{
    Title:       "Shader Sketch",
    SketchWidth: 1080, SketchHeight: 1080,
    ShaderPath:  "fragment.kage", // enables shader mode + live reload
})
```

`Config.ShaderSrc []byte` accepts embedded source instead (e.g. via
`//go:embed`) when you want a single self-contained binary; there is no
live reload in that case. `Drawer` is unused in shader mode; `Updater` is
optional.

# The live display is always 1:1

The on-screen shader sketch always renders at exactly
`SketchWidth x SketchHeight` — the Builtins **Preview Mode** checkbox (a
raster-resolution knob for CPU sketches) doesn't apply and is hidden from
the panel in shader mode; a fragment shader is never expected to benefit
from a cheaper preview raster the way a CPU sketch's vector recording does.
Keeping the live display at 1:1 is also what makes source images (below)
and the ping-pong state buffer (below) simple to reason about while
running — every image sketchy binds to the shader on the live path is
guaranteed to be the same size as the render target.

**Export Scale** *does* apply in shader mode: it controls the resolution of
PNG saves (Save Image / Snapshot dialogs) only, independent of what's on
screen. At a scale above 1x, any bound `//sketchy:image` source or the
ping-pong state buffer is upscaled (GPU, linear filter) to match the export
size for that one capture — the originals, and a running simulation, are
untouched.

# Directives: uniforms become controls

Declare uniforms in the shader's top-level `var` block (one per line) with
a trailing directive comment:

```go
var (
    Zoom    float //sketchy:slider min=0.1 max=100 default=0.4 step=0.05 folder=Fractal
    MaxIter int   //sketchy:slider min=16 max=1024 default=256 step=16
    ColorA  vec3  //sketchy:color default=#cc3311
    Invert  float //sketchy:checkbox label=Inverted
    Mode    int   //sketchy:dropdown options=Waves|Rings default=0
    Aux     vec2  //sketchy:none
)
```

| Control | Uniform types | Keys (defaults) | Uniform value |
|---------|---------------|-----------------|---------------|
| `slider` | `float`, `int` | `min` (0), `max` (1 float / 10 int), `default` (min), `step` (0.01 / 1), `digits`, `folder`, `label` | the slider value |
| `checkbox` | `float`, `int` | `default` (0; accepts `true`/`false`), `folder`, `label` | 0 or 1 |
| `color` | `vec3`, `vec4` | `default` (`#ffffff`), `folder`, `label` | RGB(A) normalized 0–1 (vec4 alpha is 1) |
| `dropdown` | `int` | `options=A\|B\|C` (required), `default` (index), `folder`, `label` | selected index |
| `none` | any | — | not passed; supply via `ExtraUniforms` |

`folder=` groups the control under a collapsible header; `label=` changes
the panel text without changing the uniform name (quote values that contain
spaces: `label="Use sine palette"`). A uniform with no directive (and no
builtin match, below) is passed as zero and noted once on stdout — mark it
`//sketchy:none` to silence the note.

# Builtin uniforms

Declare any of these (name **and** type must match) and sketchy supplies
the value automatically — no directive needed:

| Declaration | Value |
|-------------|-------|
| `Time float` | seconds since start (`Tick / 60`; the template pins `ebiten.SetTPS(60)`) |
| `Tick int` | raw tick count |
| `Resolution vec2` | render-target size in pixels (`imageDstSize()` works too) |
| `Mouse vec2` | cursor position in canvas coordinates |
| `Seed float` | the sketch's random seed (changes with ↑/↓//) |
| `Substep int` | when the state shader has a `Steps` int slider, index `0..Steps-1` of the current tick's simulation passes (for dither / multi-step feedback) |

Declaring `Time` or `Tick` makes the sketch redraw every tick (animated);
declaring `Mouse` makes it redraw when the cursor moves. Without any of
these, a shader sketch redraws only when a control changes — same dirty
model as CPU sketches.

# Live reload

With `ShaderPath`, sketchy polls the file's mtime (every 30 ticks) and
recompiles on change:

- **Success**: the shader hot-swaps; controls are rebuilt from the new
  directives with current values preserved by name (values clamp into any
  new slider range). A "Shader reloaded" line appears in the Builtins
  panel.
- **Failure** (Kage compile error, bad directive): the last good shader
  keeps rendering and the error is shown in the Builtins panel and on
  stdout until the next successful save.

This makes the edit loop shadertoy-fast: leave the sketch running, edit
`fragment.kage`, save, look.

Imported libraries (below) are watched too, so editing a shared SDF or
palette file reloads every sketch you have open against it.

# Importing a shader library

Kage itself has no imports — Ebitengine rejects any `import` declaration —
so sketchy resolves them before handing the source to the compiler. Write a
library as an ordinary `.kage` file with a package clause and no `Fragment`
entry point:

```go
// lib/sdf.kage
package sdf

const Tau = 6.283185307179586

func Circle(p vec2, r float) float {
	return length(p) - r
}
```

and import it by package name, qualifying references as you would in Go:

```go
//kage:unit pixels

package main

import "sdf"

func Fragment(dstPos vec4) vec4 {
	d := sdf.Circle(dstPos.xy, 40.0*sdf.Tau)
	return vec4(vec3(step(0.0, -d)), 1.0)
}
```

An import path resolves to `<path>.kage`, searched in this order — the
first hit wins, so a sketch-local copy shadows the shared one and can be
iterated on before being promoted:

1. each entry of `$SKETCHY_KAGE_PATH` (colon-separated), if set
2. the sketch directory
3. the sketch's `lib/` subdirectory
4. `~/.config/sketchy/kage/`

Libraries may import other libraries. Import cycles, a package clause that
disagrees with the import path, and references to names a library doesn't
declare are all reported with the offending file and line.

Rules worth knowing:

- **A library cannot declare package-level `var`.** Every one of those
  would become a uniform (see below), and only the sketch's own source
  contributes controls, so sketchy rejects it outright rather than letting
  it fail later as a confusing `undefined`.
- **Import aliases and dot-imports are not supported.** One package per
  name.
- Two libraries may define the same function name; each package's
  declarations are namespaced during resolution.
- Only the sketch's own file gets `//kage:unit`; libraries must not repeat
  it.
- Unused library functions cost nothing — Ebitengine drops functions that
  the entry point can't reach before generating GLSL/HLSL/MSL, so a large
  library only costs a little compile time.

Compile errors report the file you actually wrote, with the original line
and column, whether the mistake is in the sketch or in a library:

```
/home/you/.config/sketchy/kage/sdf.kage:12:9: unexpected identifier: lenth
```

`visual_tests/shader_import/` is a worked example: two libraries, one of
which imports the other.

# Constants and globals

Kage has no package-level `var` other than uniforms, but package-level
**`const` does work**, which covers the usual reason for wanting a global:

```go
const Pi = 3.141592653589793
const Tau = 6.283185307179586
```

Constants are scalar only (`float`, `int`, `bool`) — `const c = vec3(1,0,0)`
is not valid. For composite constants use a zero-argument function, which
costs nothing when unused:

```go
func Red() vec3 { return vec3(1.0, 0.0, 0.0) }
```

# Computed uniforms from Go

For uniform types with no natural control (`vec2`, matrices) or values
computed per frame, set `ExtraUniforms`. It is merged last, so it can also
override any control or builtin:

```go
s.ExtraUniforms = func(s *sketchy.Sketch) map[string]any {
    return map[string]any{
        "Aux":  []float32{float32(x), float32(y)},
        "Beat": math.Sin(float64(s.Tick) / 30.0),
    }
}
```

Floats pass as `float64`, ints as `int`, vectors as `[]float32`.

**Momentary buttons**: there's no `//sketchy:` directive for a one-shot
button (only `slider`/`checkbox`/`color`/`dropdown`/`none`). For a
"trigger this once" control — a reset, a re-randomize — register a real
button in `BuildUI` and forward a one-tick pulse through `ExtraUniforms`,
edge-detected against the toggle's previous state so holding it "checked"
doesn't fire every tick:

```go
var pulse, wasChecked bool

s.BuildUI = func(_ *sketchy.Sketch, ui *sketchy.UI) {
    ui.Button("Reset")
}
s.Updater = func(s *sketchy.Sketch) {
    checked := s.Toggle("Reset")
    pulse = checked && !wasChecked
    wasChecked = checked
    if checked {
        s.SetBool("", "Reset", false) // consume the click
    }
}
s.ExtraUniforms = func(_ *sketchy.Sketch) map[string]any {
    v := 0.0
    if pulse {
        v = 1.0
    }
    return map[string]any{"Reset": v}
}
```

Declare the shader side as `Reset float //sketchy:none` and branch on
`Reset > 0.5`. See `examples/reaction_diffusion` for a complete example
(paired with `StatePath` to reset a running simulation).

# Source images: //sketchy:image

A shader can read a photo or texture via a source-image slot
(`imageSrc0At`.. `imageSrc3At`, Ebitengine's mechanism for binding up to 4
images to a `Fragment` call). Kage has no sampler/image type to declare a
uniform for, so instead of a `var` line, bind an image with a standalone
directive comment anywhere in the file:

```go
//sketchy:image path=photo.jpg slot=0

func Fragment(dstPos vec4, srcPos vec2) vec4 {
    c := imageSrc0At(srcPos)
    gray := dot(c.rgb, vec3(0.299, 0.587, 0.114))
    return vec4(vec3(gray), c.a)
}
```

- `path=` resolves relative to the sketch's working directory (same as
  `Config.Images`), or use an absolute path.
- `slot=` is optional; omitted slots are assigned in appearance order
  starting from 0. Explicit `slot=` lets you control which one a shader
  uses without relying on directive order.
- The image is decoded once and resized (high-quality CPU resize) to
  exactly `SketchWidth x SketchHeight` before upload — Ebitengine requires
  every bound source image to match the render target's size exactly, and
  the live display is always 1:1 (above), so this happens once, not per
  frame. A PNG save at Export Scale above 1x upscales a copy on the GPU
  just for that capture (above); the native-resolution original is
  unaffected.
- Using `imageSrcNAt` in `Fragment` requires the Kage function signature
  `func Fragment(dstPos vec4, srcPos vec2) vec4` (Ebitengine's requirement,
  not sketchy's — the plain `func Fragment(dstPos vec4) vec4` form doesn't
  receive `srcPos`).
- A shader with a `StatePath` (below) has slot 0 reserved for the ping-pong
  buffer; a `//sketchy:image` directive may not claim slot 0 in that case
  (use `slot=1`-`3`).
- Editing `path=`/`slot=` and saving live-reloads the bound image like any
  other directive change.

# Ping-pong: a state (simulation) pass

For feedback effects that need to remember the previous frame —
reaction-diffusion, flame-fractal-style accumulation, trails — set
`Config.StatePath` alongside `Config.ShaderPath`:

```go
s := sketchy.New(sketchy.Config{
    ShaderPath: "fragment.kage", // display pass (unchanged role)
    StatePath:  "state.kage",    // simulation pass (new)
})
```

Each tick, sketchy runs `state.kage` first: it reads the current state
buffer via `imageSrc0At` and returns the next state, which becomes the
buffer `fragment.kage` (and the next tick's `state.kage`) reads at slot 0.
Both files are ordinary Kage programs — `state.kage` needs the same
`func Fragment(dstPos vec4, srcPos vec2) vec4` signature as any shader
reading a source image, and can declare its own `//sketchy:` uniforms and
directives, merged into the same control panel as `fragment.kage`'s (a
uniform with the same folder+name in both files shares one control).

**Seeding**: buffers start blank. Branch on the existing `Tick` builtin to
initialize:

```go
func Fragment(dstPos vec4, srcPos vec2) vec4 {
    if Tick == 0 {
        return seedPattern(dstPos) // initial condition
    }
    prev := imageSrc0At(srcPos)
    return update(prev)
}
```

**Notes**:

- A `StatePath` shader always animates (redraws every tick), regardless of
  whether it declares `Time`/`Tick` — a ping-pong simulation is
  definitionally continuous.
- An int slider named **`Steps`** on the state shader makes sketchy run that
  many simulation passes per tick (with builtin `Substep` counting them).
  Useful when one pass per frame is too slow (e.g. Gray-Scott on an 8-bit
  buffer). `Sketch.ClearState()` clears both ping-pong images — pair with a
  one-tick `Reset` uniform to reseed.
- The buffer pair is allocated once at exactly `SketchWidth x SketchHeight`
  (the live display is always 1:1) and persists for the sketch's lifetime.
- Live-reloading either file does **not** reset the buffers — tune
  `state.kage`'s equation live and watch the existing pattern respond. The
  display and state shaders are recompiled and committed together (both
  must compile successfully) so a broken edit never leaves the pair, or
  their shared image-slot assignment, out of sync.
- Saving an image or snapshot never advances the simulation — capture only
  re-runs the display pass against whatever the buffer currently holds.
- Snapshots restore control values only, not buffer contents — reloading a
  snapshot doesn't rewind or fast-forward a running simulation.
- See `examples/reaction_diffusion` for a complete Gray-Scott example.

# Saving and recording

- **PNG** works from the Save Image / Snapshot dialogs, at the Builtins
  **Export Scale** dropdown's resolution (see above) — the live display
  stays native 1:1 regardless of the selected scale.
- **SVG is unavailable** (a fragment shader has no vector form); the
  checkbox disappears in shader mode.
- **Video recording** ([recording.md](recording.md)) works exactly as for
  CPU sketches — including perfect loops: if your shader is periodic in
  `Time` with period `P` seconds, arm a Loop recording with `N = 60 * P`
  frames. Frames are read back from the GPU each tick; like the CPU path,
  encoder backpressure slows the preview but never the file. A `StatePath`
  simulation keeps evolving during recording exactly as it does live.
- **Snapshots** store directive-generated controls like any other control
  and restore them by name.

# Limitations

- Preview Mode doesn't apply in shader mode; the live display is always
  1:1 (above). The Builtins **Export Scale** dropdown affects PNG saves
  only, not the live display; video recording has its own independent
  **Rec scale** ([recording.md](recording.md)).
- `DisableClearBetweenFrames` is not supported in shader mode (fatal at
  Init). Use `StatePath` for accumulation/feedback effects instead.
- `DefaultBackground`/`DefaultForeground`/stroke settings do not affect
  shader output — the fragment shader owns every pixel.
- Controls renamed in the shader lose their snapshot/live values (matching
  is by name).
- Imported libraries cannot declare uniforms or `//sketchy:image`
  directives; only the sketch's own shader file contributes controls and
  source images.
- A `StatePath` simulation's buffer contents are not part of snapshots
  (only control values are).
