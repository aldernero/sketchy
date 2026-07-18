package sketchy

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuiltinSnapshotExportScaleRoundTrip(t *testing.T) {
	s := New(Config{SketchWidth: 100, SketchHeight: 100})
	s.RasterDPI = 4 * DefaultDPI
	s.syncExportScaleIdxFromDPI()

	data, err := s.serializeBuiltinState()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"export_scale":4`) {
		t.Fatalf("serialized builtin state missing export scale: %s", data)
	}

	// Loading into a fresh sketch restores RasterDPI and the dropdown index.
	s2 := New(Config{SketchWidth: 100, SketchHeight: 100})
	s2.RasterDPI = DefaultDPI
	if err := s2.applyBuiltinStateJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.RasterDPI != 4*DefaultDPI {
		t.Fatalf("RasterDPI = %v, want %v", s2.RasterDPI, 4*DefaultDPI)
	}
	if got := exportScaleFactors[s2.builtinExportScaleIdx]; got != 4 {
		t.Fatalf("export scale dropdown = %vx, want 4x", got)
	}
}

func TestBuiltinSnapshotWithoutExportScaleKeepsDPI(t *testing.T) {
	// Older snapshots have no export_scale; loading one must not touch
	// RasterDPI.
	old := builtinSnapshotPayload{
		DefaultBackground:    "#000000",
		DefaultForeground:    "#ffffff",
		DefaultStrokeWidthMM: 1,
		RandomSeed:           42,
	}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	s := New(Config{SketchWidth: 100, SketchHeight: 100})
	s.RasterDPI = 2 * DefaultDPI
	s.syncExportScaleIdxFromDPI()
	if err := s.applyBuiltinStateJSON(data); err != nil {
		t.Fatal(err)
	}
	if s.RasterDPI != 2*DefaultDPI {
		t.Fatalf("RasterDPI = %v, want %v (unchanged)", s.RasterDPI, 2*DefaultDPI)
	}
}

func TestSyncExportScaleIdxFromDPI(t *testing.T) {
	s := New(Config{SketchWidth: 100, SketchHeight: 100})
	cases := []struct {
		dpi  float64
		want float64
	}{
		{96, 1},
		{192, 2},
		{288, 3},
		{384, 4},
		{768, 8},
		{100, 1}, // non-preset snaps to nearest for display
		{500, 4},
	}
	for _, tc := range cases {
		s.RasterDPI = tc.dpi
		s.syncExportScaleIdxFromDPI()
		if got := exportScaleFactors[s.builtinExportScaleIdx]; got != tc.want {
			t.Errorf("dpi %v -> %vx, want %vx", tc.dpi, got, tc.want)
		}
	}
	// A non-preset DPI is preserved until the user picks a preset.
	s.RasterDPI = 100
	s.syncExportScaleIdxFromDPI()
	if s.RasterDPI != 100 {
		t.Fatalf("RasterDPI = %v, want 100 (sync must not modify it)", s.RasterDPI)
	}
}
