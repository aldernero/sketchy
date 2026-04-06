# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.2.0]: https://github.com/aldernero/sketchy/compare/v0.1.0...v0.2.0
