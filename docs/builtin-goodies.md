# Builtin Goodies

Some of Sketchy's best features are builtin and designed to make iterating on
your designs quick and easy. This document covers the **Builtins** header in
the control panel and the other conveniences every sketch gets for free.

# The Builtins panel

The Builtins header is fixed by Sketchy (it is not part of your `BuildUI`
controls):

- **Seed** — integer field plus a **Rand** button; mirrors `RandomSeed`.
- **Default background / Default foreground** — color pickers for the canvas
  clear color and the initial stroke color before your `Drawer` runs.
- **Default stroke width** — pixels, clamped text field.
- **Export scale** — preset multiplier (1×–8×) over the raster resolution.
  The display always presents at `SketchWidth` × `SketchHeight`; higher
  scales supersample redraws and PNG saves (a 2048px sketch at 4× saves an
  8192px PNG). Persisted in snapshots.
- **Preview mode** — renders the display at half resolution for ~4× faster
  redraws while iterating; saves are unaffected. Not persisted in snapshots.
- **Discrete palette / Sine palette** — dropdowns listing
  [palettedb](https://github.com/aldernero/palettedb) palettes for use in
  your `Drawer` via `s.DiscretePalette` / `s.SinePalette`.
- **Recording** (Rec format / FPS / scale / mode) — records animations to
  WebM, MP4, animated WebP, or lossless FFV1 via ffmpeg, with manual,
  fixed-length, and perfect-loop modes. **Ctrl+R** starts/stops. See
  [Recording video](recording.md).
- **Save Image… / Take Snapshot… / Load Snapshot…** — dialogs described
  below.
- **UI theme** — Dark or Light control-panel style; the letterbox margin
  around the sketch follows it.

The panel is hidden from rasterized output and toggles with **Ctrl+Space**
(plain **Space** is reserved for typing in the panel's text fields).

# Saving designs as images

**Save Image…** writes the current frame under `saves/png/` and/or
`saves/svg/` in the sketch working directory, named
`<prefix>_<timestamp>.{png,svg}`. Because every frame is recorded as it is
drawn, saves replay that recording — the file matches the display exactly,
even for sketches that use randomness:

- **PNG** renders at the Builtins **Export scale** (via `RasterDPI`), so one
  click exports print-resolution rasters.
- **SVG** is true vector output with real stroked bezier paths in pixel
  coordinates — ready for pen-plotter toolchains (vpype, axidraw, …).

With `DisableClearBetweenFrames`, accumulation is display-only: saves render
just the current frame's recording.

Ebitengine's screenshot key is also wired up: **Esc** saves the entire window
(panel included) as `screenshot_<timestamp>.png` — useful for blog posts and
bug reports. Sketchy sets `EBITEN_SCREENSHOT_KEY=escape` at `Init` unless you
have set it yourself.

# Snapshots

**Take Snapshot…** stores the complete state of the sketch in a SQLite
database (`sketch.db`, created on first run in the sketch directory):

- every user control (sliders, toggles, color pickers, dropdowns), and
- the Builtins state (default colors, stroke width, seed, export scale,
  selected palettes),

optionally along with a PNG and/or SVG render of the frame (checkboxes in the
dialog; the files go under `saves/` and are linked from the snapshot row).

**Load Snapshot…** lists saved snapshots and restores one, including the
random seed — so a snapshot reproduces the exact design, not just the
settings. Controls that no longer exist in the sketch are reported and
skipped.

# Random number generator (with noise)

The sketch struct has a builtin random number generator `s.Rand`
(a [gaul.Rng](https://pkg.go.dev/github.com/aldernero/gaul#Rng)), seeded from
`RandomSeed` / the `-s` flag, or from the clock when unset. Beyond uniform
and Gaussian values it wraps a simplex-noise source
([peterhellberg/gfx](https://github.com/peterhellberg/gfx)) with fractal
octaves and offsets — invaluable for organic-looking generative work:

```
SetSeed(seed int64)
Gaussian(mean, stdev float64) float64
Noise1D(x) / Noise2D(x, y) / Noise3D(x, y, z) / Noise4D(x, y, z, w)
SetNoiseOctaves(n) / SetNoisePersistence(p) / SetNoiseLacunarity(l)
SetNoiseScaleX/Y/Z(scale) and SetNoiseOffsetX/Y/Z(offset)
UniformRandomPoints(num, rect) / NoisyRandomPoints(num, threshold, rect)
```

See [gaul's random.go](https://github.com/aldernero/gaul/blob/main/random.go)
for the full API.

# Keyboard shortcuts

| Key | Action |
|-----|--------|
| **↑** / **↓** | Increment / decrement the random seed |
| **/** | Randomize the seed |
| **Ctrl+Space** | Show / hide the control panel |
| **Ctrl+R** | Start/arm or stop/disarm a [video recording](recording.md) |
| **Esc** | Ebitengine window screenshot |

The seed keys re-render immediately, which makes stepping through seed
variations of a design fast — a workflow inspired by
[vsketch](https://github.com/abey79/vsketch).
