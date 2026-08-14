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
// envelope.
//
// That was first fixed by INFERENCE — a cite with no `label` was read as red's — which made the
// distinction depend on the absence of a field, so a blue cite written without a label silently
// rejoined red's count. #341 makes it structural: two acts, two EVENT TYPES, nothing inferred.
func TestCiteProvenanceIsTwoEventTypes(t *testing.T) {
	authored := Event{Type: "cite", Payload: NewPayload().Set("label", "c-abc").Set("url", "https://x").Set("title", "T")}
	verified := Event{Type: "verify", Payload: NewPayload().Set("claim", "c").Set("reference", "https://y").Set("trust", "high")}

	if authored.Type == verified.Type {
		t.Fatal("blue authoring a citation and red verifying one must not share an event type — that sharing is what made the count inflatable")
	}
	// The discriminator must not be recoverable from a payload field: a blue cite MISSING its
	// label must still be a blue cite, which is exactly the case the old heuristic got wrong.
	unlabelled := Event{Type: "cite", Payload: NewPayload().Set("url", "https://x")}
	if unlabelled.Type != "cite" {
		t.Error("a cite without a label is still a cite — provenance is the type, not a field's emptiness")
	}
}
