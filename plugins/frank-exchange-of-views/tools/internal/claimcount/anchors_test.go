package claimcount

import (
	"reflect"
	"testing"
)

// A gap marker is a DIFFERENT invisible anchor class than a citation, so it never
// affects the claim count (only citation anchors are claims). This pins that a report
// peppered with gap markers counts only its cited claims.
func TestGapMarkersDoNotAffectCount(t *testing.T) {
	base := "Water is wet<!--cite:C-00000001-->.\n\nThe sky is blue."
	withMarker := "Water is wet<!--cite:C-00000001--><!--gap:G-00000abc-->.\n\nThe sky is blue.<!--gap:G-000009f2-->"
	if got := Count(base); got != 1 {
		t.Fatalf("base Count = %d, want 1 (the cited claim)", got)
	}
	if got := Count(withMarker); got != 1 {
		t.Errorf("Count with gap markers = %d, want 1 UNCHANGED (a gap anchor is not a claim)", got)
	}
}

// ProtectedAnchorIDs is the PRESENT set: the distinct ids of every anchor, walking the kinds
// table in its order (citation, computation, gap) and each kind first-seen. A footnote reference is
// not an anchor, and a report with none yields none.
func TestProtectedAnchorIDsWalksTheKindsTable(t *testing.T) {
	for _, c := range []struct {
		name, md string
		want     []string
	}{
		{"gaps, distinct and first-seen", "One.<!--gap:G-0000000a--> Two.<!--gap:G-0000000b--> Three.<!--gap:G-0000000a-->\n\nFour.[^L1]", []string{"G-0000000a", "G-0000000b"}},
		{"citations, distinct and first-seen", "One.<!--cite:C-0000000a--> Two.<!--cite:C-0000000b--> Three.<!--cite:C-0000000a-->", []string{"C-0000000a", "C-0000000b"}},
		{"kinds in table order, not text order", "A.<!--computation:K-00000001--> B<!--cite:C-00000001--><!--gap:G-00000001-->.", []string{"C-00000001", "K-00000001", "G-00000001"}},
		{"none", "no markers here.[^L1]", nil},
	} {
		if got := ProtectedAnchorIDs(c.md); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: ProtectedAnchorIDs = %v, want %v", c.name, got, c.want)
		}
	}
}
