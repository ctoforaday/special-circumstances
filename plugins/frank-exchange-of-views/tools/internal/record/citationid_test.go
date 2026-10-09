package record

import (
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// idLetters is the letter each kind's id begins with, written out here and nowhere read from the
// table that spells it: a letter that moves in the table fails this file by name.
var idLetters = map[string]string{
	"finding": "F", "citation": "C", "proof": "P", "gap": "G", "avenue": "Q", "motion": "M",
}

// Every kind's id is the one shape: its letter, a hyphen and exactly eight lower-case hex.
func TestNewID_Shape(t *testing.T) {
	for kind, letter := range idLetters {
		shape := regexp.MustCompile(`^` + letter + `-[0-9a-f]{8}$`)
		for i := 0; i < 64; i++ {
			if id := NewID(kind); !shape.MatchString(id) {
				t.Fatalf("NewID(%q) = %q, want match %s", kind, id, shape)
			}
		}
	}
}

func TestNewID_Unguessable(t *testing.T) {
	// Distinct across a batch — the id is random, not a sequence a seat could continue
	// (the L6-F8..F16 failure id.go documents). Collision at 4 bytes is ~1 in 4e9
	// per pair; 1000 draws is a comfortable non-flaky margin.
	for kind := range idLetters {
		seen := map[string]bool{}
		for i := 0; i < 1000; i++ {
			id := NewID(kind)
			if seen[id] {
				t.Fatalf("NewID(%q) collided on %q within 1000 draws", kind, id)
			}
			seen[id] = true
		}
	}
}

func TestNewID_TheLetterIsTheKind(t *testing.T) {
	// No two kinds may ever be confused: each has a letter no other kind holds.
	kindOf := map[string]string{}
	for kind, letter := range idLetters {
		if other, dup := kindOf[letter]; dup {
			t.Fatalf("%s and %s share the letter %s", other, kind, letter)
		}
		kindOf[letter] = kind
		if got := NewID(kind)[:2]; got != letter+"-" {
			t.Errorf("%s id prefix = %q, want %q", kind, got, letter+"-")
		}
	}
	// A kind the record does not mint is the caller's defect, never an id nothing reads.
	defer func() {
		if recover() == nil {
			t.Error(`NewID("label") minted an id for a kind that does not exist`)
		}
	}()
	NewID("label")
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
	authored := recordtest.Event(t, "", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), Label: proto.String("c-abc"), Url: proto.String("https://x"), Title: proto.String("T")})
	verified := recordtest.Event(t, "", &recordpb.Verify{Claim: proto.String("c"), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH)})

	if authored.Type == verified.Type {
		t.Fatal("blue authoring a citation and red verifying one must not share an event type — that sharing is what made the count inflatable")
	}
	// The discriminator must not be recoverable from a payload field: a blue cite MISSING its
	// label must still be a blue cite, which is exactly the case the old heuristic got wrong.
	unlabelled := recordtest.Event(t, "", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), Url: proto.String("https://x")})
	if unlabelled.GetType() != recordpb.EventType_EVENT_TYPE_CITE {
		t.Error("a cite without a label is still a cite — provenance is the type, not a field's emptiness")
	}
}
