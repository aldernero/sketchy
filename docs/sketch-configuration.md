This guide covers how to configure a sketch. Configuration is **code-first**:
everything is set on the [`sketchy.Config`](../config.go) struct (or on the
[`Sketch`](../sketch.go) before `Init()`), and controls are registered in Go
with the [`UI`](../ui_builder.go) builder. There is no JSON configuration
file.

# The Config struct

A minimal sketch configures size and colors and hands off three functions:

```go
s := sketchy.New(sketchy.Config{
	Title:        "My Sketch",
	SketchWidth:  1080,
	SketchHeight: 1080,
})
s.BuildUI = buildUI // register controls
s.Updater = update  // per-tick logic
s.Drawer = draw     // drawing, receives *render.Context
s.Init()
```

All canvas units are **pixels** (origin top-left, y down); the drawing surface
is exactly `SketchWidth` × `SketchHeight`.

## Config fields

| Field                     | Type        | Default     | Description |
|---------------------------|-------------|-------------|-------------|
| Title                     | string      | "Sketch"    | window/sketch title |
| Prefix                    | string      | "sketch"    | filename prefix for saves and snapshots |
| SketchWidth               | float64     | 1080        | sketch area width in pixels |
| SketchHeight              | float64     | 1080        | sketch area height in pixels |
| ControlWidth              | int         | 330         | control panel width in pixels |
| ControlHeight             | int         | 500         | control panel height in pixels |
| ControlBackgroundColor    | string      | "#1e1e1e"   | control panel background color |
| ControlOutlineColor       | string      | "#ffdb00"   | control panel outline color |
| SketchBackgroundColor     | string      | (unused)    | reserved; window margins follow the Builtins UI theme (dark/light grey) |
| SketchOutlineColor        | string      | —           | sketch area outline color |
| DefaultBackground         | color.Color | black       | canvas clear color before each draw |
| DefaultForeground         | color.Color | white       | initial stroke color before `Drawer` runs |
| DefaultStrokeWidth        | float64     | 1           | initial stroke width in pixels |
| DisableClearBetweenFrames | bool        | false       | keep previous frames' raster so strokes accumulate on screen (display-only; saves render just the current frame). `Sketch.Clear()` wipes to `DefaultBackground` |
| ShowFPS                   | bool        | false       | overlay the frame rate |
| RasterDPI                 | float64     | 96          | raster resolution; 96 = one raster pixel per sketch pixel, higher values supersample redraws and PNG saves (also settable from the Builtins **Export scale** dropdown, where 1×–8× maps to 96–768) |
| PreviewMode               | bool        | false       | render the display at half resolution for ~4× faster redraws; saves are unaffected (also a Builtins checkbox) |
| RandomSeed                | int64       | 0 (auto)    | seed for the builtin PRNG; 0 seeds from the clock at `Init` |
| PaletteDBPath             | string      | ""          | [palettedb](https://github.com/aldernero/palettedb) database for the Builtins palette dropdowns; empty means `~/.config/palettedb/palettedb.db` |
| Images                    | []ImageAsset| (none)      | image files loaded at `Init`; draw with `DrawNamedImage` |
| ShaderPath                | string      | ""          | enables [shader mode](shaders.md): the file is compiled as a Kage fragment shader whose `//sketchy:` directives auto-create controls; live-reloaded on change. `Drawer` is unused |
| ShaderSrc                 | []byte      | (none)      | embedded Kage source instead of a file (no live reload); `ShaderPath` wins when both are set |
| DisableFastStroke         | bool        | false       | no-op kept for compatibility (the old tdewolff/canvas FastStroke workaround is gone) |

Each [`ImageAsset`](../images.go) has `Name` (the key used with
`Image`/`DrawNamedImage`) and `Path` (relative to the sketch directory or
absolute).

## Fields set on the Sketch after New

Between `sketchy.New` and `s.Init()` you assign the sketch's behavior and can
override anything from `Config`:

- **`BuildUI func(*Sketch, *UI)`** — registers controls (sliders, checkboxes,
  buttons, color pickers, dropdowns, folders). See the
  [Getting Started](getting-started.md) guide.
- **`Updater func(*Sketch)`** — runs every tick; use it for animation and to
  react to `DidControlsChange` and friends.
- **`Drawer func(*Sketch, *render.Context)`** — draws the frame with a
  [gaul render context](https://pkg.go.dev/github.com/aldernero/gaul/render).
  Only re-run when the sketch is dirty (control changes, pointer presses in
  the sketch, or an explicit `MarkDirty()`).

# Referencing configuration in code

Sketch fields are available as `s.<FieldName>` inside `update` and `draw`
(e.g. `s.Width()`, `s.Height()`, `s.CanvasRect()`). Control values are read
with `s.GetFloat(folder, name)` / `s.GetInt` / `s.GetBool` / `s.GetColor` /
`s.GetDropdownIndex`, or the root-folder shorthands `s.Slider(name)`,
`s.Int(name)`, `s.Toggle(name)`, `s.ColorPicker(name)`.

# Command line parameters

The project template (`sketchy init`) wires up these flags:

- `-p <prefix>` — filename prefix for saves (overrides `Prefix`)
- `-s <seed>` — seed for the builtin random number generator (0 = auto)
- `-palettedb <path>` — palettedb database for the Builtins palette dropdowns
- `-pprof <file>` — collect a CPU profile

They are ordinary `flag` definitions in your `main.go`, so add or remove
flags as you see fit.
