package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A CORROBORATION THAT PLACED A MARKER IS A SOURCE (gblock's ruling, 2026-09-11). In B9 red-lens-
// evidence's labelled verify put <!--cite:c-d49a32c4--> in the report; the evidence view listed it
// only under independent[], so the same lens minted G4 calling its own marker a citation with no
// source, while the assembler wove it into a working footnote.
func TestEvidence_ACorroborationThatPlacedAMarkerIsASource(t *testing.T) {
	b := evidenceBoard(
		evidenceEvent(t, 1, "red-lens-evidence", &recordpb.Verify{Claim: proto.String("the textbook example"),
			Label: proto.String("c-d49a32c4"), Url: proto.String("https://w/fp"), Title: proto.String("Fermat pseudoprime")}),
		evidenceEvent(t, 1, "red-lens-evidence", &recordpb.Verify{Claim: proto.String("a check with no marker"),
			Url: proto.String("https://w/other"), Title: proto.String("Other")}),
	)
	got := EvidenceJSONOf(b)

	var src *EvidenceSourceJSON
	for i := range got.Sources {
		if got.Sources[i].Anchor == "c-d49a32c4" {
			src = &got.Sources[i]
		}
	}
	if src == nil {
		t.Fatalf("the corroboration that placed c-d49a32c4 is not a source: %+v", got.Sources)
	}
	if src.CorroboratedBy != "red-lens-evidence" || src.Location != "the textbook example" || src.URL != "https://w/fp" {
		t.Errorf("the corroboration source = %+v, want it marked as red-lens-evidence's, at its claim, on its url", *src)
	}
	if len(src.Verified) != 1 {
		t.Errorf("a corroboration is red's own reading of the source and must count as checked; verified = %+v", src.Verified)
	}
	if len(got.Sources) != 1 {
		t.Errorf("a verify that placed no marker became a source: %+v", got.Sources)
	}
	if len(got.Independent) != 2 {
		t.Errorf("the checks themselves stay under independent[]: %+v", got.Independent)
	}
}
