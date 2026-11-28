package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// actOnce is runAt with one retry where the gap moved after the seat's sitting opened: the refusal
// moves the read point past what it shows, and the same act lands when made again.
func actOnce(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, err := runAt(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, args...)
	}
	return out, err
}

// T-B50 (the concession rows, whose mechanism R2 holds): PER CAUSE, EACH PARTY'S CONCESSION ENDS THE
// DOCKETING AND THE PROPOSAL IT HOLDS WITH NO RULING — the raiser's close of G and its `--accept M`, the
// author's `--accept G` (an edit taking the raiser's stated pair) at impasse and under both proposals —
// in both twins (the Go fold and gap_state), after which the bench's ruling is refused, no closing is
// owed on the docketing, and the chair's next plan dockets the proposal again nowhere. Under grade
// movement the author's accept is refused (TestTheAuthorAcceptsTheRaisersStatement) and the raiser's
// close concedes. The rows whose mechanism is R3's (lensSitsAfter, partiesSitAfter, the freeze and the
// docket order) and R4's (the amend, a gap subject's statement) are carried there.
func TestEveryConcessionEndsItsDocketingUnruled(t *testing.T) {
	raiserCloses := func(runDir, g, _ string) []string {
		return []string{"close", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--as", "not_a_defect", "--verified-by", lensSeat,
			"--verified-with", "the report read", "--verified-against", "the report's sentence", "--reason", "it is not a defect"}
	}
	raiserAccepts := func(verb string) func(runDir, g, m string) []string {
		return func(runDir, _, m string) []string { return acceptArgs(runDir, verb, m) }
	}
	authorAccepts := func(runDir, g, _ string) []string {
		return []string{"edit", "--run", runDir, "--seat-id", blueSeat, "--accept", g, "--reason", "red is right", "--json"}
	}
	// proposed files the author's proposal of act on g after one exchange, and returns it.
	proposed := func(t *testing.T, runDir, g, act string) string {
		exchangeOn(t, runDir, g)
		if act == "close" {
			return proposeClose(t, runDir, blueSeat, g, "not_a_defect", "the sentence says what the gap asks")
		}
		out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
			out, err = propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		}
		if err != nil {
			t.Fatalf("the author's regrade proposal on %s: %v %s", g, err, out)
		}
		return filedMotionID(t, out)
	}
	for _, c := range []struct {
		name     string
		cause    recordpb.DocketingCause
		proposal string // the act the docketed proposal proposes; "" for none
		concede  func(runDir, g, m string) []string
		party    string
	}{
		{"impasse: the raiser's close", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "", raiserCloses, lensSeat},
		{"impasse: the author's --accept G", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "", authorAccepts, blueSeat},
		{"a close proposal: the raiser's close", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "close", raiserCloses, lensSeat},
		{"a close proposal: the raiser's close --accept M", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "close", raiserAccepts("close"), lensSeat},
		{"a close proposal: the author's --accept G", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "close", authorAccepts, blueSeat},
		{"a regrade proposal: the raiser's close", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "regrade", raiserCloses, lensSeat},
		{"a regrade proposal: the raiser's regrade --accept M", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "regrade", raiserAccepts("regrade"), lensSeat},
		{"a regrade proposal: the author's edit --accept G", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "regrade", authorAccepts, blueSeat},
		{"grade movement: the raiser's close", recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, "", raiserCloses, lensSeat},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir, g := prescribedRun(t)
			m := ""
			if c.proposal != "" {
				m = proposed(t, runDir, g, c.proposal)
			} else if c.cause == recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
				exchangeOn(t, runDir, g)
			}
			d := docketNow(t, runDir, g, c.cause, m)
			asSeat(t, runDir, c.party, "register")
			if out, err := actOnce(t, c.concede(runDir, g, m)...); err != nil {
				t.Fatalf("the concession under %s: %v %s", d, err, out)
			}
			ms, ds, _ := partyRecord(t, runDir)
			if k := ds[d]; k == nil || k.Standing() || k.Ruled() {
				t.Errorf("docketing %s reads %+v in the Go fold after the concession; want ended, unruled", d, k)
			}
			if p, sd := standingCols(t, runDir, g); p != "" || sd != "" {
				t.Errorf("gap_state reads proposal %q and docketing %q standing after the concession; want none", p, sd)
			}
			if m != "" {
				if got := ms[m]; got == nil || got.Standing() {
					t.Errorf("proposal %s reads %+v after the concession; want ended", m, got)
				}
			}
			all, err := record.MergedEvents(runtest.Open(t, runDir))
			if err != nil {
				t.Fatal(err)
			}
			for _, duty := range record.ClosingDuties(all.Events, all.At) {
				if duty.Docketing == d && (duty.State == record.ClosingAwaited || duty.State == record.ClosingMissing) {
					t.Errorf("a closing is owed on %s after its concession: %+v", d, duty)
				}
			}
			conveneBench(t, runDir, g)
			if _, err := answerDocketing(t, runDir, d, "for-author"); err == nil {
				t.Errorf("the bench's ruling of %s after its concession was admitted; want refused", d)
			}
			if m == "" {
				return
			}
			asSeat(t, runDir, "red-chair", "register")
			asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
			_, ds, _ = partyRecord(t, runDir)
			for id, k := range ds {
				if id != d && k.MotionID == m {
					t.Errorf("the chair's next plan dockets conceded proposal %s again as %s", m, id)
				}
			}
		})
	}
}
