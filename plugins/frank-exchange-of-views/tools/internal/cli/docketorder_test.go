package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE DOCKET ORDER (#1695 §III.B4; gblock 2026-11-26, "One run-wide queue"): the bench rules the run's
// docketings one at a time, in the order written. Two docketings, G's then H's, written at once: in a
// docket sitting convened on both, in the terminal sitting and in the assembly, H's ruling is refused
// naming G's docketing, G's lands, and then H's lands in the same sitting.
func TestTheBenchRulesOnlyTheDocketsFirst(t *testing.T) {
	for _, occasion := range []string{"docket", "terminal", "assemble"} {
		t.Run(occasion, func(t *testing.T) {
			runDir := acceptRun(t)
			g, err := mintQuote(t, runDir, "G", acceptS+".")
			if err != nil {
				t.Fatal(err)
			}
			h, err := mintQuote(t, runDir, "H", acceptS+".", "--class", otherClass)
			if err != nil {
				t.Fatal(err)
			}
			dg := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
			dh := docketNow(t, runDir, h, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
			if occasion == "docket" {
				conveneBench(t, runDir, g, h)
			} else {
				asSeat(t, runDir, "judge", "register", "--occasion", occasion)
			}
			out, err := answerDocketing(t, runDir, dh, "for-author", "--final")
			heldRefusal(t, "H's ruling before G's", out, err, dg, "not the docket's first", "Nothing was recorded")
			if out, err := answerDocketing(t, runDir, dg, "for-author", "--final"); err != nil {
				t.Fatalf("G's ruling, the docket's first: %v\n%s", err, out)
			}
			if out, err := answerDocketing(t, runDir, dh, "for-author", "--final"); err != nil {
				t.Fatalf("H's ruling, first once G's is ruled: %v\n%s", err, out)
			}
		})
	}
}

// THE DISPATCH CONVENES THE BENCH ON THE DOCKET'S READY RUN: its first, and each next whose closings
// are in, up to the first that is not. G's docketing is first and awaits its closings; H's has both. The
// chair's plan readies no bench, and says H waits behind G. Once G's closings are in, the plan convenes
// the bench on G and H, in that order. A bench that sits for the run and rules nothing is not readied
// again for G, and H, behind it, waits too.
func TestTheDispatchConvenesTheBenchOnTheReadyRun(t *testing.T) {
	runDir := acceptRun(t)
	g, err := mintQuote(t, runDir, "G", acceptS+".")
	if err != nil {
		t.Fatal(err)
	}
	h, err := mintQuote(t, runDir, "H", freezeLater, "--class", otherClass)
	if err != nil {
		t.Fatal(err)
	}
	g, h = handleText(t, runDir, g), handleText(t, runDir, h)
	docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	docketNow(t, runDir, h, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	states := func(gap string, quote, with string) {
		t.Helper()
		if out, err := redStates(t, runDir, gap, "--quote", quote, "--new", with); err != nil {
			t.Fatalf("red's closing on %s: %v\n%s", gap, err, out)
		}
		if out, err := blueStates(t, runDir, gap, ""); err != nil {
			t.Fatalf("blue's closing on %s: %v\n%s", gap, err, out)
		}
	}
	benchGaps := func() ([]string, []string) {
		t.Helper()
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		for _, p := range plan.Parties {
			if p.SeatID == "judge" {
				return p.GapIDs, plan.Why
			}
		}
		return nil, plan.Why
	}
	states(h, freezeLater, "Volume recovered by June.")
	gaps, why := benchGaps()
	if gaps != nil {
		t.Errorf("the bench is convened on %v while G, the docket's first, awaits its closings", gaps)
	}
	if !slices.ContainsFunc(why, func(w string) bool {
		return strings.HasPrefix(w, h+":") && strings.Contains(w, g+" is ahead of it and not ready")
	}) {
		t.Errorf("the plan does not say H waits behind G:\n%s", strings.Join(why, "\n"))
	}
	states(g, acceptS+".", acceptN+".")
	if gaps, why := benchGaps(); !slices.Equal(gaps, []string{g, h}) {
		t.Fatalf("the bench is convened on %v, want the ready run [%s %s]:\n%s", gaps, g, h, strings.Join(why, "\n"))
	}
	asSeat(t, runDir, "judge", "register", "--occasion", "docket")
	if gaps, why := benchGaps(); gaps != nil {
		t.Errorf("the bench is readied again on %v after a sitting that ruled nothing:\n%s", gaps, strings.Join(why, "\n"))
	}
}
