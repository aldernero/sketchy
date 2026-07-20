# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.6.0]: https://github.com/aldernero/sketchy/compare/v0.5.0...v0.6.0
[0.3.0]: https://github.com/aldernero/sketchy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/aldernero/sketchy/compare/v0.1.0...v0.2.0
