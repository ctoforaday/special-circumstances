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

// CitationAnchorIDs is the citation-axis twin: the distinct cite ids in the
// <!--cite:...--> comment tokens, first-seen order. It must ignore finding anchors.
func TestCitationAnchorIDs(t *testing.T) {
	md := "One.<!--cite:c-a--> Two.<!--cite:c-b--> Three.<!--cite:c-a-->\n\nFour.<!--fx:f-9-->"
	got := CitationAnchorIDs(md)
	want := []string{"c-a", "c-b"} // distinct, first-seen; the finding anchor is not a citation
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CitationAnchorIDs = %v, want %v", got, want)
	}
	if ids := CitationAnchorIDs("no citations here.<!--fx:f-1-->"); len(ids) != 0 {
		t.Errorf("a report with no citation anchors yielded %v", ids)
	}
}

// The two anchor classes are orthogonal: each extractor sees only its own prefix even
// when both appear in one report.
func TestAnchorClassesAreOrthogonal(t *testing.T) {
	md := "Claim.<!--cite:c-1--><!--fx:f-1-->"
	if got := CitationAnchorIDs(md); !reflect.DeepEqual(got, []string{"c-1"}) {
		t.Errorf("CitationAnchorIDs = %v, want [c-1]", got)
	}
	if got := FindingAnchorIDs(md); !reflect.DeepEqual(got, []string{"f-1"}) {
		t.Errorf("FindingAnchorIDs = %v, want [f-1]", got)
	}
}
