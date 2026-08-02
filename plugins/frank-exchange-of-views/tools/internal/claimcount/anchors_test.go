package claimcount

import (
	"reflect"
	"testing"
)

// A finding-marker is an invisible HTML comment, not a footnote, so it never affects
// Count. This pins that a report peppered with markers counts only its real claims.
func TestFindingMarkersDoNotAffectCount(t *testing.T) {
	base := "Water is wet.[^L1]\n\nThe sky is blue."
	withMarker := "Water is wet.[^L1]<!--fx:f-abc-->\n\nThe sky is blue.<!--fx:f-9f2-->"
	if got := Count(base); got != 1 {
		t.Fatalf("base Count = %d, want 1 (the [^L1] claim)", got)
	}
	if got := Count(withMarker); got != 1 {
		t.Errorf("Count with finding-markers = %d, want 1 UNCHANGED (a comment is not a claim)", got)
	}
}

// FindingAnchorIDs is the detector's PRESENT set: the distinct finding ids in the
// <!--fx:...--> comment tokens, first-seen order.
func TestFindingAnchorIDs(t *testing.T) {
	md := "One.<!--fx:f-a--> Two.<!--fx:f-b--> Three.<!--fx:f-a-->\n\nFour.[^L1]"
	got := FindingAnchorIDs(md)
	want := []string{"f-a", "f-b"} // distinct, first-seen; the [^L1] claim is not a marker
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindingAnchorIDs = %v, want %v", got, want)
	}
	if ids := FindingAnchorIDs("no markers here.[^L1]"); len(ids) != 0 {
		t.Errorf("a report with no markers yielded %v", ids)
	}
}
