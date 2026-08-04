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

// The two cite provenances must never be counted as one. Measured on the 2026-08-04 smoke: the
// board reported 10 "citations checked" where the truth was 3 blue-authored + 7 red-verified —
// 43% inflation on RED's own audit-volume tile, which debate.js tells red to copy into its
// envelope. The inflation grows with every use of the citation axis.
func TestCiteProvenanceIsDiscriminated(t *testing.T) {
	authored := Event{Type: "cite", Payload: NewPayload().Set("label", "c-abc").Set("url", "https://x").Set("title", "T")}
	verified := Event{Type: "cite", Payload: NewPayload().Set("claim", "c").Set("reference", "https://y").Set("confidence", "high")}
	other := Event{Type: "finding", Payload: NewPayload().Set("label", "L1-F1")}

	if !IsAuthoredCite(authored) || IsVerifiedCite(authored) {
		t.Error("a blue cite (carries label) must read as AUTHORED, not verified")
	}
	if !IsVerifiedCite(verified) || IsAuthoredCite(verified) {
		t.Error("a red lens cite (no label) must read as VERIFIED, not authored")
	}
	// A finding also carries a label — the discriminator must key on TYPE too, or every finding
	// would count as an authored citation.
	if IsAuthoredCite(other) || IsVerifiedCite(other) {
		t.Error("a non-cite event must be neither")
	}
}
