package blue

import "testing"

// The whole design (#709) rests on ONE property: replaying the diff-stack over the frozen base
// reproduces exactly the bytes the file used to hold. If that holds, the file is derivable and can
// be deleted; if it drifts by a byte, deleting the file loses data. So the fidelity test compares
// Render against planEdit — the SAME transform blue edit writes to the file — op for op.
func TestRenderReproducesTheEditPathByteForByte(t *testing.T) {
	base := "The cost is stable and the volume grows steadily."
	steps := []Op{
		{Old: "stable", New: "steady."}, // ends in "." abutting nothing — exercises seam tidy
		{Old: "the volume grows steadily.", New: "demand climbs."},
		{Old: "cost", New: "price"},
	}

	// Apply the steps through blue edit's own pure core, accumulating the running report the way
	// the file would have.
	viaEdit := base
	for i, s := range steps {
		next, err := planEdit(viaEdit, s.Old, s.New)
		if err != nil {
			t.Fatalf("planEdit step %d (%q→%q): %v", i, s.Old, s.New, err)
		}
		viaEdit = next
	}

	viaRender, err := Render(base, steps)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if viaRender != viaEdit {
		t.Errorf("Render drifted from the edit path:\n  edit:   %q\n  render: %q", viaEdit, viaRender)
	}
}

func TestRenderEmptyStackIsTheFrozenBase(t *testing.T) {
	base := "An untouched base with a finding-marker.<!--fx:f-abcdef01-->"
	got, err := Render(base, nil)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got != base {
		t.Errorf("empty stack changed the base:\n  base:   %q\n  render: %q", base, got)
	}
}

func TestRenderTidiesTheSplicedSeam(t *testing.T) {
	// "stable" → "steady." lands a period against the base's period; the seam tidy collapses "..".
	got, err := Render("The value is stable.", []Op{{Old: "stable", New: "steady."}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if want := "The value is steady."; got != want {
		t.Errorf("seam not tidied on replay:\n  want %q\n  got  %q", want, got)
	}
}

// A stack that no longer describes a real sequence of edits over this base is a RECORD INTEGRITY
// failure, and it must be loud — a silently skipped op would render a report that never existed.
func TestRenderFailsLoudOnAnOpItCannotLocate(t *testing.T) {
	_, err := Render("The base text.", []Op{{Old: "not present in the base", New: "x"}})
	if err == nil {
		t.Fatal("Render silently skipped an op whose old span is absent — a replay integrity failure must be loud")
	}
}
