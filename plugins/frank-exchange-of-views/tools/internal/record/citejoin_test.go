package record

import "testing"

// THE LENS WAS NEVER TOLD ABOUT AN UNVERIFIED CITATION.
//
// citedClaimsWithoutVerify joined on `cite_key`, falling back to `key`. A `cite` event carries
// neither — its keys are access_date, claim, label, location, sha256, title, url — so the lookup
// always missed and the affordance returned empty on every board in the tool's lifetime. "Nothing
// outstanding" and "the join key is not on the event" are the same bytes.
//
// These use the payload shapes the verbs actually write, so a rename on either side fails here
// rather than going quiet.

func citeEvent(anchor, claim string) Event {
	return Event{SeatID: "blue-synthesize", Type: "cite", Payload: NewPayload().
		Set("label", anchor).Set("claim", claim).Set("url", "https://x").
		Set("title", "T").Set("sha256", "abc").Set("location", "§1")}
}

func verifyEvent(anchor string) Event {
	return Event{SeatID: "red-lens-r1-L1", Type: "verify", Payload: NewPayload().
		Set("anchor", anchor).Set("claim", "c").Set("outcome", "supports").
		Set("confidence", "high").Set("reference", "https://x")}
}

func TestAnUnverifiedCitationIsAfforded(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{citeEvent("c-a08c9764", "the floor is 30 days")}}
	got := citedClaimsWithoutVerify(b)
	if len(got) != 1 || got[0] != "c-a08c9764" {
		t.Fatalf("an unverified citation afforded %v — the join key is not reaching the event", got)
	}
}

func TestAVerifiedCitationStopsBeingAfforded(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
		citeEvent("c-a08c9764", "the floor is 30 days"),
		verifyEvent("c-a08c9764"),
	}}
	if got := citedClaimsWithoutVerify(b); len(got) != 0 {
		t.Errorf("the affordance survived its own discharge: %v", got)
	}
}

// AN INDEPENDENT VERIFY DISCHARGES NOTHING, and that is the rule holding rather than a gap in it:
// it is a check against a source blue never cited, so it carries no anchor and must not silence a
// citation nobody looked at.
func TestAnIndependentVerifyDoesNotDischargeACitation(t *testing.T) {
	indep := Event{SeatID: "red-lens-r1-L1", Type: "verify", Payload: NewPayload().
		Set("claim", "c").Set("outcome", "supports").Set("reference", "https://other")}
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{citeEvent("c-a08c9764", "x"), indep}}
	if got := citedClaimsWithoutVerify(b); len(got) != 1 {
		t.Errorf("an independent verify silenced an uninspected citation: %v", got)
	}
}

// The affordance must reach the seat, not merely exist — the surrounding derivation is what a
// lens actually reads.
func TestTheLensSeesTheAffordance(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{citeEvent("c-a08c9764", "x")}}
	// Asked of the SITTING, not of AvailableOf: "reaches the seat" is a claim about the one list
	// a seat reads, and this test passed for as long as the affordance existed on a surface the
	// seat's completion check could not see.
	got := SittingOf(b, "lens", "red-lens-r1-L1").Open
	if !mentions(got, "c-a08c9764") {
		t.Errorf("the lens is not shown the unverified citation on its work list: %v", hows(got))
	}
}
