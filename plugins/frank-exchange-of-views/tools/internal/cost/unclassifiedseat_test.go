package cost

import (
	"strings"
	"testing"
)

// AN UNCLASSIFIABLE SEAT MUST NOT LEAVE THE TIER AUDIT SILENT.
//
// TierMismatch is the fable trap — a judgment seat quietly running on a dearer model. It used to
// `continue` past any row whose seat classified as `other`, on a comment calling those "not
// tier-bound". Every seat in SeatClass HAS a tier, so ClassOf returns "" for exactly one input:
// the seat ClassifySeat could not name.
//
// So a prompt-wording drift in debate.js sent a seat to `other` and its spend was never checked,
// while seatclass's own doc promised `other` was "a visible bucket, never folded away, so a
// prompt-wording drift is spottable". The precedent is in that doc: cost-audit once lacked the
// terminal-disposition case and misattributed that seat's spend.
func TestAnUnidentifiableSeatIsReportedRatherThanSkipped(t *testing.T) {
	rows := []Row{{Seat: "other", Round: 0, T: "opus", Turns: 12}}
	findings := TierMismatch(rows, "haiku", "sonnet")
	if len(findings) == 0 {
		t.Fatal("an unidentifiable seat produced NO finding — its spend went unchecked and the run reads as fully audited")
	}
	f := findings[0]
	if f.Verdict != "WARN" {
		t.Errorf("verdict = %q, want WARN", f.Verdict)
	}
	// The message has to name the CAUSE, or the next reader treats it as a seat that is exempt.
	for _, want := range []string{"could not be identified", "NOT checked", "drift"} {
		if !strings.Contains(f.Why, want) {
			t.Errorf("the finding does not say %q — it must name a classifier drift, not imply an exemption:\n%s", want, f.Why)
		}
	}
}

// AND A CLASSIFIED SEAT ON ITS CONFIGURED TIER STILL PASSES SILENTLY, so the new finding cannot
// turn every clean run into a warning.
func TestAMatchingTierStillProducesNoFinding(t *testing.T) {
	rows := []Row{{Seat: "red-lens", Round: 1, T: "haiku", Turns: 5}}
	if got := TierMismatch(rows, "haiku", "sonnet"); len(got) != 0 {
		t.Errorf("a bulk seat on the configured bulk tier produced %d finding(s), want 0: %+v", len(got), got)
	}
}
