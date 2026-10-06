package claimcount

import (
	"reflect"
	"testing"
)

// A finding-marker is a DIFFERENT invisible anchor class than a citation, so it never
// affects the claim count (only citation anchors are claims). This pins that a report
// peppered with finding markers counts only its cited claims.
func TestFindingMarkersDoNotAffectCount(t *testing.T) {
	base := "Water is wet<!--cite:c-1-->.\n\nThe sky is blue."
	withMarker := "Water is wet<!--cite:c-1--><!--fx:f-abc-->.\n\nThe sky is blue.<!--fx:f-9f2-->"
	if got := Count(base); got != 1 {
		t.Fatalf("base Count = %d, want 1 (the cited claim)", got)
	}
	if got := Count(withMarker); got != 1 {
		t.Errorf("Count with finding-markers = %d, want 1 UNCHANGED (a finding anchor is not a claim)", got)
	}
}

// ProtectedAnchorIDs is the PRESENT set: the distinct ids of every anchor, walking the kinds
// table in its order (finding, citation, proof) and each kind first-seen. A footnote reference is
// not an anchor, and a report with none yields none.
func TestProtectedAnchorIDsWalksTheKindsTable(t *testing.T) {
	for _, c := range []struct {
		name, md string
		want     []string
	}{
		{"findings, distinct and first-seen", "One.<!--fx:f-a--> Two.<!--fx:f-b--> Three.<!--fx:f-a-->\n\nFour.[^L1]", []string{"f-a", "f-b"}},
		{"citations, distinct and first-seen", "One.<!--cite:c-a--> Two.<!--cite:c-b--> Three.<!--cite:c-a-->", []string{"c-a", "c-b"}},
		{"kinds in table order, not text order", "A.<!--proof:p-1--> B<!--cite:c-1--><!--fx:f-1-->.", []string{"f-1", "c-1", "p-1"}},
		{"none", "no markers here.[^L1]", nil},
	} {
		if got := ProtectedAnchorIDs(c.md); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: ProtectedAnchorIDs = %v, want %v", c.name, got, c.want)
		}
	}
}
