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
the panel text without changing the uniform name. A uniform with no
directive (and no builtin match, below) is passed as zero and noted once on
stdout — mark it `//sketchy:none` to silence the note.

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

# Saving and recording

- **PNG** works from the Save Image / Snapshot dialogs at the Builtins
  **Export scale** — the shader re-renders natively at the scaled
  resolution, so supersampled exports are essentially free.
- **SVG is unavailable** (a fragment shader has no vector form); the
  checkbox disappears in shader mode.
- **Video recording** ([recording.md](recording.md)) works exactly as for
  CPU sketches — including perfect loops: if your shader is periodic in
  `Time` with period `P` seconds, arm a Loop recording with `N = 60 * P`
  frames. Frames are read back from the GPU each tick; like the CPU path,
  encoder backpressure slows the preview but never the file.
- **Snapshots** store directive-generated controls like any other control
  and restore them by name.

# Limitations

- `PreviewMode` is ignored (shaders render at display resolution; they are
  already fast).
- `DisableClearBetweenFrames` is not supported in shader mode (fatal at
  Init). Accumulation effects belong inside the shader.
- `DefaultBackground`/`DefaultForeground`/stroke settings do not affect
  shader output — the fragment shader owns every pixel.
- Controls renamed in the shader lose their snapshot/live values (matching
  is by name).
