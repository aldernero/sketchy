package sketchy

import (
	"slices"
	"testing"
)

func TestAppendBuiltinPaletteNames(t *testing.T) {
	stored := []string{"my-grad", "viridis"}
	names := appendBuiltinPaletteNames(stored, "discrete")

	if names[0] != "my-grad" || names[1] != "viridis" {
		t.Errorf("stored palettes should come first, got %v", names[:2])
	}
	if n := len(slices.DeleteFunc(slices.Clone(names), func(s string) bool { return s != "viridis" })); n != 1 {
		t.Errorf("viridis appears %d times, want 1 (stored shadows built-in)", n)
	}
	if !slices.Contains(names, "plasma") {
		t.Errorf("built-in plasma missing from %v", names)
	}
	if slices.Contains(names, "warm-sunset") {
		t.Errorf("sine built-in warm-sunset should not be in discrete names %v", names)
	}

	sines := appendBuiltinPaletteNames(nil, "sine")
	if !slices.Contains(sines, "warm-sunset") {
		t.Errorf("built-in warm-sunset missing from sine names %v", sines)
	}
}
