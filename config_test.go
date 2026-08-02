package sketchy

import "testing"

func TestAppSizeDefaultsToSketchSize(t *testing.T) {
	s := New(Config{SketchWidth: 800, SketchHeight: 600})
	if w, h := s.AppSize(); w != 800 || h != 600 {
		t.Fatalf("AppSize() = %v, %v, want 800, 600", w, h)
	}
}

func TestAppSizeOverridesSketchSize(t *testing.T) {
	s := New(Config{SketchWidth: 800, SketchHeight: 600, AppWidth: 1400, AppHeight: 700})
	if w, h := s.AppSize(); w != 1400 || h != 700 {
		t.Fatalf("AppSize() = %v, %v, want 1400, 700", w, h)
	}
	if s.SketchWidth != 800 || s.SketchHeight != 600 {
		t.Fatalf("sketch size = %v x %v, want 800 x 600", s.SketchWidth, s.SketchHeight)
	}
	w, h := s.WindowSize()
	if w != 1400 || h != 700 {
		t.Fatalf("WindowSize() = %d, %d, want 1400, 700", w, h)
	}
}

// App dimensions set directly on the Sketch (after New) still take effect, and
// unset ones fall back to the sketch size even without Init.
func TestAppSizePartialOverride(t *testing.T) {
	s := New(Config{SketchWidth: 800, SketchHeight: 600})
	s.AppWidth = 1200
	if w, h := s.AppSize(); w != 1200 || h != 600 {
		t.Fatalf("AppSize() = %v, %v, want 1200, 600", w, h)
	}
}

func TestWindowSizeClampsToMinimum(t *testing.T) {
	s := New(Config{SketchWidth: 800, SketchHeight: 600, AppWidth: 100, AppHeight: 100})
	w, h := s.WindowSize()
	if w != MinWindowWidth || h != MinWindowHeight {
		t.Fatalf("WindowSize() = %d, %d, want %d, %d", w, h, MinWindowWidth, MinWindowHeight)
	}
}

func TestLayoutFallsBackToAppSize(t *testing.T) {
	s := New(Config{SketchWidth: 800, SketchHeight: 600, AppWidth: 1400, AppHeight: 700})
	w, h := s.Layout(0, 0)
	if w != 1400 || h != 700 {
		t.Fatalf("Layout(0, 0) = %d, %d, want 1400, 700", w, h)
	}
}
