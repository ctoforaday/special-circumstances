package goldentest

import (
	"strings"
	"testing"
)

// firstDivergence replaced a full dump of both artifacts. The dump is what made
// "regenerate and move on" cheaper than reading, so these tests hold the property that
// made it worth replacing: the message must say WHERE they part, not restate both sides.

func TestFirstDivergenceNamesTheLineAndShowsBothSides(t *testing.T) {
	want := "L1\nL2\nL3\nL4\nL5\nL6\ncharlie\nL8\nL9\n"
	got := "L1\nL2\nL3\nL4\nL5\nL6\nCHARLIE\nL8\nL9\n"
	out := firstDivergence(want, got)

	if !strings.Contains(out, "line 7") {
		t.Errorf("must name the diverging line number, got:\n%s", out)
	}
	if !strings.Contains(out, "- golden") || !strings.Contains(out, "+ got") {
		t.Errorf("must show both sides of the diverging line, got:\n%s", out)
	}
	if !strings.Contains(out, "charlie") || !strings.Contains(out, "CHARLIE") {
		t.Errorf("must include the actual differing content, got:\n%s", out)
	}
	// The point of the change: an artifact far longer than the window is NOT reprinted.
	// Two lines of context either side is context; everything else is the dump this replaced.
	// Divergence at line 7 with a ±2 window prints lines 5–9, so L5–L9 are context, by
	// design. L1–L4 are the part a dump would have reprinted for nothing.
	for _, far := range []string{"L1", "L2", "L3", "L4"} {
		if strings.Contains(out, far+"\n") || strings.Contains(out, "| "+far) {
			t.Errorf("%s is outside the ±2 window and must not be printed; the whole artifact "+
				"must not be reprinted:\n%s", far, out)
		}
	}
}

// A pure length difference has no differing line to point at. Reporting "no difference"
// there would be a failure message that denies the failure.
//
// Note the inputs carry NO trailing newline: with one, the split yields a final empty
// element and the artifacts diverge on it, so this branch would not be reached. That is a
// real property of the function, and writing the test the other way round is how I first
// got it wrong.
func TestFirstDivergenceHandlesAPureLengthDifference(t *testing.T) {
	out := firstDivergence("alpha\nbravo", "alpha\nbravo\ncharlie")
	if !strings.Contains(out, "length differs") {
		t.Errorf("a prefix-identical artifact of different length must say so, got:\n%s", out)
	}
	if !strings.Contains(out, "identical for") {
		t.Errorf("must state how far they agreed, got:\n%s", out)
	}
}

// Divergence on the very first line must not index out of bounds.
func TestFirstDivergenceAtLineOne(t *testing.T) {
	out := firstDivergence("alpha\n", "beta\n")
	if !strings.Contains(out, "line 1") {
		t.Errorf("got:\n%s", out)
	}
}

// The counts are the reader's sanity check that they are looking at the whole artifact.
func TestFirstDivergenceReportsBothLengths(t *testing.T) {
	out := firstDivergence("a\nb\nc\n", "a\nX\nc\n")
	if !strings.Contains(out, "line(s) vs") {
		t.Errorf("must report both artifact lengths, got:\n%s", out)
	}
}
