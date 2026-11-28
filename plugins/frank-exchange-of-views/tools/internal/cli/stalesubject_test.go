package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN AMEND IS A CHANGE TO ITS GAP, AND A CHANGE TO A GAP IS A CHANGE TO EVERY GAP ABOUT IT (§III.B7; the
// stale guard, gblock ruling B). otherLens raises G2 on the evidence lens's G and its sitting stays
// open; the evidence lens amends G. otherLens's regrade of G2, read off the G its sitting opened on, is
// refused stale naming the amend of G, a Stale event moves its read point, and the same regrade then
// lands. Without the amend, the regrade lands at once.
func TestAnAmendOfAGapIsNewsToAnActOnAGapAboutIt(t *testing.T) {
	stage := func(t *testing.T, amend bool) (string, string) {
		runDir := seatRun(t)
		g := mintGap(t, runDir, "under", "spec_underspecification")
		asSeat(t, runDir, otherLens, "register")
		g2, err := mintOnGap(t, runDir, otherLens, "on-gap", g)
		if err != nil {
			t.Fatal(err)
		}
		if amend {
			if out, err := amendGap(t, runDir, lensSeat, g, "--acceptance-check", "the check names an observable outcome", "--reason", "the check was vague"); err != nil {
				t.Fatalf("the minter's amend of %s: %v %s", g, err, out)
			}
		}
		return runDir, g2
	}
	regrade := func(t *testing.T, runDir, g2 string) (string, error) {
		return runAt(t, "regrade", "--run", runDir, "--seat-id", otherLens, "--id", g2, "--severity", "low", "--reason", "the amend narrowed it")
	}

	runDir, g2 := stage(t, false)
	if out, err := regrade(t, runDir, g2); err != nil {
		t.Fatalf("the regrade of %s with its subject unchanged: %v %s", g2, err, out)
	}

	runDir, g2 = stage(t, true)
	stales := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE)
	out, err := regrade(t, runDir, g2)
	if err == nil || !strings.Contains(err.Error()+out, "amended G") {
		t.Fatalf("the regrade of %s after its subject's amend = %v %s, want refused stale naming the amend", g2, err, out)
	}
	if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_STALE); got != stales+1 {
		t.Errorf("stale events %d, want %d: the refusal moves the read point", got, stales+1)
	}
	if out, err := regrade(t, runDir, g2); err != nil {
		t.Errorf("the same regrade once the amend is read: %v %s", err, out)
	}
}
