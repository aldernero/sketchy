# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.1] - 2026-07-28

### Changed

- **gaul upgraded to v0.4.0** ([release notes](https://github.com/aldernero/gaul/releases/tag/v0.4.0)), a correctness release that fixes 13 functions which silently returned wrong results. No sketchy source changes were needed and nothing sketchy calls changed signature. All 21 bundled examples and visual tests render byte-identically before and after the upgrade, apart from three whose output is already nondeterministic run to run (`examples/reaction_diffusion`, `examples/shader_photo`, `visual_tests/nearest_neighbor`) — for those the cross-version difference is smaller than the same-version run-to-run variance, so none of it is attributable to gaul.

  Your own sketches may still render differently, because the bundled examples happen not to exercise the changed paths. Specifically:

  - **Seeded output changed.** `Rand.Gaussian`, `Rand.UniformRandomPoints`, and `Rand.NoisyRandomPoints` previously drew from the global `math/rand`, so `Config.RandomSeed` did not control them at all and the same seed gave different results on every run. They now draw from the seeded PRNG and are genuinely reproducible. `Rand.Prng.Uint64n` also rejects the biased tail of its range instead of returning a plain `Next() % n`.
  - **Fractal noise weights octaves.** `Rand.Noise*D` never multiplied each octave by its amplitude, so `SetNoisePersistence` had no effect. Sketches using the default single octave are unaffected — which is why the examples did not move — but sketches calling `SetNoiseOctaves(n)` with `n > 1` will look different, and correct.
  - **`Noise4D`'s w axis works.** `wscale` defaulted to zero, pinning the w term at 0. gaul adds `SetNoiseScaleW` / `SetNoiseOffsetW`.
  - **Geometry fixes.** `Point.Rotate`, `Rect.Intersects` (returned false for *every* overlapping pair), `Circle.Boundary`, `Curve.Copy` (forced `Closed = true`), `Curve.Centroid` and `Polygon.Centroid` (sign-flipped for one winding direction), `QuadTree.QueryCircle` (searched a quarter of the circle), and `KDTree.Size` all returned wrong answers before. `Line.Angle` now returns a full-quadrant `Atan2` result, which shifts `PerpendicularAt`. Any sketch written to compensate for the old behavior needs revisiting.
  - **`Affine2D.SetRotation` / `SetShearFactor`** now assign their component instead of multiplying into it; `SetRotation` previously produced a scale by `cos(angle)` and never a rotation.
  - **`Curve.Lerp` / `Curve.LineAt`** clamp an out-of-range percentage instead of calling `log.Fatalf` and terminating the process.

  Nearest-neighbor queries are now both correct and much faster: `PointHeap.Pop` corrupted the heap, so `KDTree.NearestNeighbors` and `QuadTree.NearestNeighbors` returned the wrong neighbors, and neither tree pruned its search. They are roughly 20x and 11x faster respectively.

## [0.7.0] - 2026-07-27

### Added

- **Shader libraries via `import`**: a shader file can now `import "sdf"` and call `sdf.Circle(...)`, so SDFs, palettes and other helpers live in one place instead of being copy-pasted into every sketch (`shader_import.go`). Kage has no import mechanism — Ebitengine rejects any import declaration — so sketchy resolves them into a single flat source before compiling. Import paths resolve to `<path>.kage`, searched in `$SKETCHY_KAGE_PATH`, then the sketch directory, then its `lib/`, then `~/.config/sketchy/kage/`; a sketch-local copy shadows the shared one. Libraries may import other libraries, and two libraries may define the same name. Imported files are watched for live reload alongside the shader itself. The transform splices bytes rather than re-printing the AST and emits `/*line*/` directives, so a compile error reports the file, line and column you actually wrote, in the sketch or in a library — an improvement on the bare `34:12:` with no filename that Ebitengine reports today. Libraries may not declare package-level `var` (every one would become a uniform); that is rejected with a pointer to `const` or a zero-argument function. See `docs/shaders.md` and `visual_tests/shader_import`. Anticipates [ebiten#3439](https://github.com/hajimehoshi/ebiten/issues/3439).
- **Source images in shader sketches**: bind a photo/texture to a `Fragment` source-image slot with a standalone `//sketchy:image path=... [slot=N]` directive comment — no Go plumbing (`shader_parse.go`, `shader.go`). The image is decoded once and resized to exactly `SketchWidth x SketchHeight` (Ebitengine requires an exact size match with the render target). Slots default to appearance order (0-3); editing `path=`/`slot=` live-reloads like any other directive.
- **Ping-pong / state pass**: `Config.StatePath` adds a second Kage shader — the simulation pass — alongside `Config.ShaderPath` (now the display pass), for feedback effects that need to remember the previous frame (reaction-diffusion, flame-fractal-style accumulation, trails). Each tick, the state shader reads the current buffer via `imageSrc0At` (slot 0, reserved) and writes the next one; the display shader then reads the updated buffer. Both files' `//sketchy:` directives merge into one control panel. Seeding uses the existing `Tick` builtin (`if Tick == 0 { ... }`); live-reloading either file preserves buffer contents; the display and state shaders recompile and commit together so a broken edit never leaves the pair (or their shared slot assignment) out of sync. Capturing an image never advances the simulation. New example: `examples/reaction_diffusion` (Gray-Scott per Karl Sims: A=1/B=0 with a noisy central B seed, DA=1/DB=0.5/dt=1, periodic wrap, write-dither so 8-bit ping-pong buffers keep evolving; `Speed` timestep, `Steps` multi-pass-per-tick, `GridSize` virtual grid, `BuildUI`/`ClearState`/`ExtraUniforms` Reset button; see `docs/shaders.md`).
- **State-pass helpers**: builtin uniform `Substep`; state-shader int slider `Steps` runs that many simulation passes per tick; `Sketch.ClearState()` clears the ping-pong buffers.
- **Example**: `examples/shader_photo` — Sobel edge detection and Hue/Saturation/Lightness shift sliders demonstrating `//sketchy:image`.
- **Export Scale in shader mode**: the Builtins **Export Scale** dropdown is shown for shader sketches and now controls the resolution of PNG saves (Save Image / Snapshot dialogs) — `Sketch.CaptureShaderImage` renders at `RasterDPI/DefaultDPI x SketchWidth x SketchHeight`. Any bound `//sketchy:image` source or ping-pong state buffer is upscaled (GPU, linear filter) into scratch images sized to match for that one capture; the live simulation and originals are untouched.

### Changed

- **Breaking**: shader sketches always render their live display at exactly `SketchWidth x SketchHeight` — the Builtins **Preview Mode** checkbox is hidden in shader mode (a raster-resolution knob that doesn't apply once source images and ping-pong buffers need a size guarantee on the live path). **Export Scale** remains available and now applies to PNG saves (above).
- **Breaking**: `Sketch.CaptureShaderImage` no longer takes a `scale float64` parameter; it now reads scale from `Sketch.RasterDPI` (the Builtins Export Scale dropdown), consistent with the CPU-sketch export path.

## [0.6.0] - 2026-07-19

### Changed

- **Breaking (CLI)**: `sketchy init` now requires the project type: `sketchy init <sketch|shader> <name>`. The former `sketchy init <name>` form is rejected with a usage hint (`cmd/sketchy/sketchy.go`).

### Added

- **Shader sketches**: render a sketch with an Ebitengine Kage fragment shader (`Config.ShaderPath`/`ShaderSrc`; `shader.go`, `shader_parse.go`). `//sketchy:` directive comments on the shader's uniforms auto-generate control-panel controls (slider, checkbox, color, dropdown) that are passed back as uniforms each frame; builtin uniforms `Time`/`Tick`/`Resolution`/`Mouse`/`Seed` are supplied when declared. The shader file live-reloads on save (errors keep the last good shader and show in the Builtins panel; control values survive reloads by name). PNG saves and video recording work via GPU readback at any export scale (SVG is unavailable in shader mode). New CLI form `sketchy init [sketch|shader] <name>` with a `template_shader` project (`fragment.kage` demo); bare `sketchy init <name>` unchanged. `ExtraUniforms` hook for computed/vec2/matrix uniforms. New guide: `docs/shaders.md`; demo: `visual_tests/shader_demo`.
- **`Sketch.FinishRecording(timeout)`**: blocks until the current video recording is fully written — for scripted sketches that exit right after recording (`video.go`).
- **Video recording**: record animations straight to WebM (VP9), MP4 (H.264), animated WebP, or lossless FFV1 by piping raw frames to a user-installed ffmpeg (`video.go`, `video_ui.go`). Builtins panel rows (format, FPS, record scale, mode) plus a **Ctrl+R** hotkey; manual, fixed-frame-count, and armed perfect-loop modes (capture starts at `Tick % N == 0` and stops after exactly N frames). Scriptable via `StartRecording` / `StopRecording` / `ArmLoopRecording`. Encoding backpressure slows the live preview instead of dropping frames, so output is always frame-perfect; recording renders independently of Preview mode. New guide: `docs/recording.md`.

## [0.3.0] - 2026-05-25

### Added

- **Named image assets**: `sketchy.Config.Images` (`ImageAsset` name + path), loaded at `Init`; `Image`, `DrawImage`, `DrawNamedImage`, `DrawImageAt`, `DrawNamedImageAt`, and `RegisterImage` for runtime bitmaps (`images.go`).
- **`DisableFastStroke`** on `Config` to opt out of the new default fast stroke path in tdewolff/canvas.
- **Example**: `examples/photo_stripes` — horizontal/vertical strip shifts with uniform, Gaussian, alternating, and cumulative modes (cylindrical wrap).
- **Example**: `examples/voronoi` — Voronoi diagram simulation.

### Changed

- **Default stroke rendering**: `Init` enables `canvas.FastStroke` unless `DisableFastStroke` is set (better performance for generative strokes).
- **Examples and visual tests**: Stroke color and width now come from Builtins **Default foreground** and **Default stroke width** (removed duplicate per-sketch thickness controls where applicable).
- **Example**: `examples/noise` uses `RegisterImage` and `DrawNamedImage` for the generated noise bitmap.
- **Template**: Commented `Images` / draw helpers in `cmd/sketchy/template/main.go`.
- **Dependencies**:
  - **gaul**, **tdewolff/canvas** (and font/minify/parse), **modernc.org/sqlite**, **golang.org/x/image**, and assorted indirects updated.
- **Example**: `examples/voronoi` updated for `gaul.VoronoiWithRect` (replaces removed `VoronoiCells`).

## [0.2.0] - 2026-04-05

### Added

- **Code-first configuration**: `sketchy.New(sketchy.Config{...})` and `config.go` so sketch dimensions, control panel size, colors, FPS toggle, random seed, preview mode, raster DPI, clear-between-frames, and default canvas background/foreground/stroke width are set in Go instead of `sketch.json`.
- **Control panel UI overhaul**: New layout and interaction model (`controls_ui.go`, `ui_builder.go`, `ui_plan.go`, `ui_theme.go`) built on updated debug UI usage.
- **Themes**: Built-in `themes/light.json` and `themes/dark.json`, with UI to switch appearance and keep controls readable against the sketch letterbox.
- **Color workflows**: Color picker / modal flow (`color_modal.go`), numeric range editing (`slider_range_modal.go`), and text-backed sliders (`slider_text.go`) for richer control bindings.
- **Persistence**: `internal/sketchdb` using SQLite (`modernc.org/sqlite`) for sketch metadata and a history of saves; snapshot support (`snapshot.go`) with sync helpers (`save_sync.go`).
- **Example**: `examples/styled_shape` demonstrating styled drawing.
- **Template**: `cmd/sketchy/template` updated for the new API; template `.gitignore` for local artifacts.

### Changed

- **`sketch.go`**: Large refactor to align runtime, UI, input (including improved mouse handling with controls), and persistence with the new configuration and panel architecture.
- **`controls.go`**: Reworked to cooperate with the new UI layer and interaction fixes.
- **Examples**: Every example now uses `sketchy.Config` in code; all per-example `sketch.json` files removed.
- **Visual tests**: Same migration to code-only config; `sketch.json` removed from each test sketch.
- **`cmd/sketchy/sketchy.go`**: CLI / embedded template path adjustments for the new layout.
- **Documentation**: `README.md`, `docs/getting-started.md`, `docs/builtin-goodies.md`, and `docs/sketch-configuration.md` updated for the new setup; screenshots refreshed under `assets/images/` (including a noise example image).
- **Dependencies**:
  - Go toolchain: **1.23.2 → 1.26.1** (`go.mod`).
  - **Ebitengine**: `ebiten/v2` **v2.8.6 → v2.9.9**.
  - **Debug UI**: `github.com/ebitengine/debugui` replaced by **`github.com/aldernero/debugui`** (fork/module line in `go.mod`).
  - **gaul**, **tdewolff/canvas**, **go-colorful** bumped to current pseudo-versions.
  - **New direct dependency**: `modernc.org/sqlite` (and related `modernc.org/*` indirects) for embedded SQLite.
- **CI**: `.github/workflows/release.yaml` Go version updated to match release builds.
- **`.gitignore`**: Additional ignores for generated or local files.

### Fixed

- Noise example issues and related README inaccuracies (prior to the large UI/config migration).
- Assorted control-panel and theme bugs; mouse interaction improvements with the new UI.

### Removed

- **`sketch.json`** from the `sketchy new` template, all `examples/*`, and all `visual_tests/*` — configuration previously expressed in JSON now lives in `sketchy.Config`.

### Performance

- General performance improvements in the library and hot paths (alongside the dependency updates).

---

## [0.1.0]

Initial published release (tag `v0.1.0`). Use `git log v0.1.0` for commit-level history before this changelog existed.

[0.7.0]: https://github.com/aldernero/sketchy/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/aldernero/sketchy/compare/v0.5.0...v0.6.0
[0.3.0]: https://github.com/aldernero/sketchy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/aldernero/sketchy/compare/v0.1.0...v0.2.0
