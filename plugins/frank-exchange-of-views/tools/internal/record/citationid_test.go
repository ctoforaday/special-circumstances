package record

import (
	"regexp"
	"testing"
)

var citIDShape = regexp.MustCompile(`^c-[0-9a-f]{8}$`)

func TestNewCitationID_Shape(t *testing.T) {
	id := NewCitationID()
	if !citIDShape.MatchString(id) {
		t.Fatalf("NewCitationID() = %q, want match %s", id, citIDShape)
	}
}

func TestNewCitationID_Unguessable(t *testing.T) {
	// Distinct across a batch — the id is random, not a sequence a seat could continue
	// (the L6-F8..F16 failure findingid.go documents). Collision at 4 bytes is ~1 in 4e9
	// per pair; 1000 draws is a comfortable non-flaky margin.
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := NewCitationID()
		if seen[id] {
			t.Fatalf("NewCitationID() collided on %q within 1000 draws", id)
		}
		seen[id] = true
	}
}

func TestCitationID_DistinctFromFindingID(t *testing.T) {
	// The two anchor classes must never be confused: c- vs f- prefix.
	if got := NewCitationID()[:2]; got != "c-" {
		t.Fatalf("citation id prefix = %q, want %q", got, "c-")
	}
	if got := NewFindingID()[:2]; got != "f-" {
		t.Fatalf("finding id prefix = %q, want %q", got, "f-")
	}
}
