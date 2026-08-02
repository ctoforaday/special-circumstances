package claimcount

import (
	"reflect"
	"testing"
)

// The load-bearing Count-safety pin for slice 1b: a finding-marker "[^f-…]" is routed
// to Anchors, never Labels, so Count is UNCHANGED by markers — the monotonicity the
// retire-vs-drop detector rests on.
func TestFindingAnchorsDoNotAffectCount(t *testing.T) {
	base := "Water is wet.[^L1]\n\nThe sky is blue."
	withMarker := "Water is wet.[^L1][^f-abc]\n\nThe sky is blue.[^f-xyz]"

	if got := Count(base); got != 1 {
		t.Fatalf("base Count = %d, want 1 (the [^L1] claim)", got)
	}
	if got := Count(withMarker); got != 1 {
		t.Errorf("Count with finding-markers = %d, want 1 UNCHANGED — a marker must never enter Labels", got)
	}
	// The claim index (Labels) is likewise unchanged: still one [^L1] claim occurrence.
	if got := occCount(Index(withMarker)); got != 1 {
		t.Errorf("Index occurrences = %d, want 1 (markers are not claims)", got)
	}
}

// A label beginning "f" but NOT "f-" (e.g. "food") is a CLAIM label, not a marker —
// it stays in Labels and counts. Guards the exact over-match the bare-`f` predicate
// would have caused.
func TestNonHyphenFLabelStaysAClaim(t *testing.T) {
	if got := Count("A tasty note.[^food]"); got != 1 {
		t.Errorf("[^food] Count = %d, want 1 — 'food' begins 'f' but not 'f-', so it is a claim", got)
	}
	if ids := FindingAnchorIDs("A tasty note.[^food]"); len(ids) != 0 {
		t.Errorf("[^food] wrongly routed to anchors: %v", ids)
	}
}

// FindingAnchorIDs is the detector's PRESENT set: the distinct f- ids in the report.
func TestFindingAnchorIDs(t *testing.T) {
	md := "One.[^f-a] Two.[^f-b] Three.[^f-a]\n\nFour.[^L1]"
	got := FindingAnchorIDs(md)
	want := []string{"f-a", "f-b"} // distinct, first-seen; the L1 claim is not an anchor
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindingAnchorIDs = %v, want %v", got, want)
	}
}
