package sketchy

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aldernero/sketchy/internal/sketchdb"
)

// newTextBoxSketch builds a sketch with text-box controls without going
// through Init, which would open a database and allocate GPU images.
func newTextBoxSketch(t *testing.T, build func(ui *UI)) *Sketch {
	t.Helper()
	s := New(Config{SketchWidth: 100, SketchHeight: 100})
	s.BuildUI = func(_ *Sketch, ui *UI) { build(ui) }
	s.rebuildControls()
	return s
}

func TestTextBoxCommitValidates(t *testing.T) {
	// A validator that upper-cases what it accepts, and rejects anything with
	// a digit in it.
	s := newTextBoxSketch(t, func(ui *UI) {
		ui.TextBox("name", "alpha", func(v string) (string, bool) {
			if strings.ContainsAny(v, "0123456789") {
				return "", false
			}
			return strings.ToUpper(v), true
		})
	})
	tb := &s.TextBoxes[0]

	// Accepted input is normalized into Val and echoed back to the field.
	tb.textBuf = "  beta "
	commitTextBoxText(tb)
	if tb.Val != "BETA" {
		t.Fatalf("Val = %q, want %q", tb.Val, "BETA")
	}
	if tb.textBuf != "BETA" {
		t.Fatalf("textBuf = %q, want the normalized value", tb.textBuf)
	}

	// Rejected input reverts the field and leaves the value alone.
	tb.textBuf = "gamma9"
	commitTextBoxText(tb)
	if tb.Val != "BETA" {
		t.Fatalf("Val = %q after a rejected commit, want it unchanged", tb.Val)
	}
	if tb.textBuf != "BETA" {
		t.Fatalf("textBuf = %q after a rejected commit, want it reverted", tb.textBuf)
	}
}

func TestTextBoxNilValidatorAcceptsAnything(t *testing.T) {
	s := newTextBoxSketch(t, func(ui *UI) { ui.TextBox("free", "", nil) })
	tb := &s.TextBoxes[0]
	tb.textBuf = "  whatever  "
	commitTextBoxText(tb)
	if tb.Val != "whatever" {
		t.Fatalf("Val = %q, want the trimmed input", tb.Val)
	}
}

func TestTextBoxGetSetAndChangeFlag(t *testing.T) {
	s := newTextBoxSketch(t, func(ui *UI) {
		ui.Folder("Location", func() { ui.TextBox("re", "-0.5", nil) })
	})

	if got := s.GetText("Location", "re"); got != "-0.5" {
		t.Fatalf("GetText = %q, want %q", got, "-0.5")
	}

	s.SetText("Location", "re", "-0.75")
	if got := s.GetText("Location", "re"); got != "-0.75" {
		t.Fatalf("GetText after SetText = %q, want %q", got, "-0.75")
	}
	// SetText also refreshes the edit buffer when nothing has focus, so the
	// panel shows the new value rather than the old one.
	if s.TextBoxes[0].textBuf != "-0.75" {
		t.Fatalf("textBuf = %q, want it synced from Val", s.TextBoxes[0].textBuf)
	}

	s.TextBoxes[0].UpdateState()
	if !s.TextBoxes[0].DidJustChange {
		t.Fatal("DidJustChange = false after SetText, want true")
	}
	s.TextBoxes[0].UpdateState()
	if s.TextBoxes[0].DidJustChange {
		t.Fatal("DidJustChange stayed true on a second UpdateState with no change")
	}
}

func TestTextBoxSyncSkippedWhileFocused(t *testing.T) {
	// A value arriving from code (mouse-driven navigation, say) must not
	// rewrite the buffer under the caret of someone typing into it.
	tb := NewTextBox("re", "-0.5", nil)
	tb.textBuf = "-0.6" // half-typed
	tb.Val = "-0.7"     // changed from outside

	tb.maybeSyncTextBufFromVal(true)
	if tb.textBuf != "-0.6" {
		t.Fatalf("textBuf = %q while focused, want the in-progress edit kept", tb.textBuf)
	}

	tb.maybeSyncTextBufFromVal(false)
	if tb.textBuf != "-0.7" {
		t.Fatalf("textBuf = %q once unfocused, want it synced from Val", tb.textBuf)
	}
}

func TestSnapshotTextBoxRoundTrip(t *testing.T) {
	build := func(ui *UI) {
		ui.Folder("Location", func() {
			ui.TextBox("re", "-0.5", nil)
			ui.TextBox("im", "0", nil)
		})
		ui.TextBox("note", "root-level", nil)
	}
	s := newTextBoxSketch(t, build)
	s.SetText("Location", "re", "-0.743643887037158704752191506114774")
	s.SetText("Location", "im", "0.131825904205311970493132056385139")
	s.SetText("", "note", "seahorse valley")

	data, err := s.serializeControlState()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"texts"`) {
		t.Fatalf("serialized control state has no texts key: %s", data)
	}

	s2 := newTextBoxSketch(t, build)
	missing, err := s2.applyControlStateJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing keys = %v, want none", missing)
	}
	for _, tc := range []struct{ folder, name, want string }{
		{"Location", "re", "-0.743643887037158704752191506114774"},
		{"Location", "im", "0.131825904205311970493132056385139"},
		{"", "note", "seahorse valley"},
	} {
		if got := s2.GetText(tc.folder, tc.name); got != tc.want {
			t.Errorf("GetText(%q, %q) = %q, want %q", tc.folder, tc.name, got, tc.want)
		}
	}
}

func TestSnapshotTextBoxRestoreSkipsValidator(t *testing.T) {
	// A snapshot holds a value the sketch accepted when it was taken. Loading
	// must not re-validate it, or a validator that has since been tightened
	// silently drops the saved location.
	s := newTextBoxSketch(t, func(ui *UI) {
		ui.TextBox("re", "0", func(string) (string, bool) { return "", false })
	})
	if err := s.setTextQuiet("", "re", "-0.75"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetText("", "re"); got != "-0.75" {
		t.Fatalf("GetText = %q, want the restored value", got)
	}
}

func TestSnapshotWithoutTextsStillLoads(t *testing.T) {
	// Schema 2 payloads have no texts key; loading one must leave text boxes
	// at their current values rather than erroring or clearing them.
	old := snapshotPayload{
		Schema:    2,
		Sliders:   map[string]float64{"gain": 0.5},
		Toggles:   map[string]bool{},
		Colors:    map[string]string{},
		Dropdowns: map[string]int{},
	}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	s := newTextBoxSketch(t, func(ui *UI) {
		ui.FloatSlider("gain", 0, 1, 0.1, 0.01)
		ui.TextBox("re", "-0.5", nil)
	})
	missing, err := s.applyControlStateJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing keys = %v, want none", missing)
	}
	if got := s.GetText("", "re"); got != "-0.5" {
		t.Fatalf("GetText = %q, want it untouched by a schema 2 snapshot", got)
	}
	if got := s.GetFloat("", "gain"); got != 0.5 {
		t.Fatalf("GetFloat = %v, want 0.5", got)
	}
}

// TestSnapshotTextBoxThroughDatabase closes the loop the JSON tests leave open:
// that a text box's value survives an actual sqlite round trip, which is what
// the Take Snapshot / Load Snapshot buttons do. The value here is a 155-digit
// coordinate, the case the control exists for — no numeric control could carry
// it, and it must come back to the digit.
func TestSnapshotTextBoxThroughDatabase(t *testing.T) {
	const deepRe = "-0.39430667921715491483844973529121181791621990356559393421504774409510447733094181971854989245264624081075266568297541965005389755403911066750342977279988891"

	db, err := sketchdb.Open(filepath.Join(t.TempDir(), "sketch.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	build := func(ui *UI) {
		ui.Folder("Location", func() {
			ui.TextBox("re", "-0.5", nil)
			ui.TextBox("span", "2.5", nil)
		})
	}
	s := newTextBoxSketch(t, build)
	s.db = db
	s.SetText("Location", "re", deepRe)
	s.SetText("Location", "span", "1.2219745453998419e-150")

	controlJSON, err := s.SerializeControlState()
	if err != nil {
		t.Fatal(err)
	}
	builtinJSON, err := s.serializeBuiltinState()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.dbInsertSnapshot("deep", "a very deep zoom", string(controlJSON), string(builtinJSON), nil, nil); err != nil {
		t.Fatal(err)
	}

	// A fresh sketch, as if the process had been restarted.
	s2 := newTextBoxSketch(t, build)
	s2.db = db
	row := s2.dbGetSnapshot("deep")
	if row == nil {
		t.Fatal("snapshot not found in the database")
	}
	missing, err := s2.ApplyControlState([]byte(row.ControlJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing controls after load: %v", missing)
	}
	if got := s2.GetText("Location", "re"); got != deepRe {
		t.Errorf("centre did not survive the database round trip:\n got %s\nwant %s", got, deepRe)
	}
	if got := s2.GetText("Location", "span"); got != "1.2219745453998419e-150" {
		t.Errorf("span = %q, want it unchanged", got)
	}
}

func TestSnapshotReportsUnknownTextKeys(t *testing.T) {
	s := newTextBoxSketch(t, func(ui *UI) { ui.TextBox("re", "0", nil) })
	data, err := json.Marshal(snapshotPayload{
		Schema:    snapshotSchemaVersion,
		Toggles:   map[string]bool{},
		Colors:    map[string]string{},
		Dropdowns: map[string]int{},
		Texts:     map[string]string{"re": "1", "Gone/away": "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	missing, err := s.applyControlStateJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 || missing[0] != "Gone/away" {
		t.Fatalf("missing = %v, want [Gone/away]", missing)
	}
}
