package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE LENS WAS NEVER TOLD ABOUT AN UNVERIFIED CITATION.
//
// citedClaimsWithoutVerify joined on `cite_key`, falling back to `key`. A `cite` event carries
// neither — its keys are access_date, claim, label, location, sha256, title, url — so the lookup
// always missed and the affordance returned empty on every board in the tool's lifetime. "Nothing
// outstanding" and "the join key is not on the event" are the same bytes.
//
// These use the payload shapes the verbs actually write, so a rename on either side fails here
// rather than going quiet.

// THE ANCHOR AND THE CLAIM ARE THE JOIN, and the earlier conversion dropped both — the helpers
// took them as parameters and built bodies that used neither, so every fixture cited the same
// anonymous source and the join under test could not be exercised.
func citeEvent(t *testing.T, anchor, claim string) *Event {
	t.Helper()
	return recordtest.Event(t, "blue-synthesize", 0, &recordpb.Cite{
		Label:    proto.String(anchor),
		Text:     proto.String(claim),
		Url:      proto.String("https://x"),
		Sha256:   proto.String("abc"),
		Location: proto.String("§1"),
	})
}

func verifyEvent(t *testing.T, anchor string) *Event {
	t.Helper()
	return recordtest.Event(t, "red-lens-r1-L1", 0, &recordpb.Verify{
		Anchor:     proto.String(anchor),
		Claim:      proto.String("c"),
		Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
		Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
		Text:       proto.String("read at the leaf"),
	})
}

func TestAnUnverifiedCitationIsAfforded(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{citeEvent(t, "c-a08c9764", "the floor is 30 days")}}
	got := citedClaimsWithoutVerify(b)
	if len(got) != 1 || got[0] != "c-a08c9764" {
		t.Fatalf("an unverified citation afforded %v — the join key is not reaching the event", got)
	}
}

func TestAVerifiedCitationStopsBeingAfforded(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
		citeEvent(t, "c-a08c9764", "the floor is 30 days"),
		verifyEvent(t, "c-a08c9764"),
	}}
	if got := citedClaimsWithoutVerify(b); len(got) != 0 {
		t.Errorf("the affordance survived its own discharge: %v", got)
	}
}

// AN INDEPENDENT VERIFY DISCHARGES NOTHING, and that is the rule holding rather than a gap in it:
// it is a check against a source blue never cited, so it carries no anchor and must not silence a
// citation nobody looked at.
func TestAnIndependentVerifyDoesNotDischargeACitation(t *testing.T) {
	indep := recordtest.Event(t, "red-lens-r1-L1", 0, &recordpb.Verify{Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)})
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{citeEvent(t, "c-a08c9764", "x"), indep}}
	if got := citedClaimsWithoutVerify(b); len(got) != 1 {
		t.Errorf("an independent verify silenced an uninspected citation: %v", got)
	}
}

// The affordance must reach the seat, not merely exist — the surrounding derivation is what a
// lens actually reads.
func TestTheLensSeesTheAffordance(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{citeEvent(t, "c-a08c9764", "x")}}
	// Asked of the SITTING, not of availableOf: "reaches the seat" is a claim about the one list
	// a seat reads, and this test passed for as long as the affordance existed on a surface the
	// seat's completion check could not see.
	got := SittingOf(b, "lens", "red-lens-r1-L1").Open
	if !mentions(got, "c-a08c9764") {
		t.Errorf("the lens is not shown the unverified citation on its work list: %v", hows(got))
	}
}
