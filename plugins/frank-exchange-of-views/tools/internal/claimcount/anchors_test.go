package claimcount

import (
	"reflect"
	"testing"
)

// A finding-marker is a DIFFERENT invisible anchor class than a citation, so it never
// affects the claim count (only citation anchors are claims). This pins that a report
// peppered with finding markers counts only its cited claims.
func TestFindingMarkersDoNotAffectCount(t *testing.T) {
	base := "Water is wet<!--cite:C-00000001-->.\n\nThe sky is blue."
	withMarker := "Water is wet<!--cite:C-00000001--><!--fx:F-00000abc-->.\n\nThe sky is blue.<!--fx:F-000009f2-->"
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
		{"findings, distinct and first-seen", "One.<!--fx:F-0000000a--> Two.<!--fx:F-0000000b--> Three.<!--fx:F-0000000a-->\n\nFour.[^L1]", []string{"F-0000000a", "F-0000000b"}},
		{"citations, distinct and first-seen", "One.<!--cite:C-0000000a--> Two.<!--cite:C-0000000b--> Three.<!--cite:C-0000000a-->", []string{"C-0000000a", "C-0000000b"}},
		{"kinds in table order, not text order", "A.<!--proof:P-00000001--> B<!--cite:C-00000001--><!--fx:F-00000001-->.", []string{"F-00000001", "C-00000001", "P-00000001"}},
		{"none", "no markers here.[^L1]", nil},
	} {
		if got := ProtectedAnchorIDs(c.md); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: ProtectedAnchorIDs = %v, want %v", c.name, got, c.want)
		}
	}
}
