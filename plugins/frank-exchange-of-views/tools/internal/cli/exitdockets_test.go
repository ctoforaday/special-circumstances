package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// proposalActs is each act a proposal takes on gap, as its filing's words.
func proposalActs(gap string) map[string][]string {
	return map[string][]string{
		"regrade": {"motion", "docket", "file", "regrade", "--id", gap, "--severity", "low", "--reason", "the harm needs two failures"},
		"close":   append([]string{"motion", "docket", "file", "close", "--id", gap, "--reason", "not a defect"}, closeFlags("not_a_defect")...),
	}
}

// T-B25: THE EXIT DOCKETS A STANDING PROPOSAL (L29). A proposal standing at the final sitting, and one an
// engine stop ends before any plan follows it, are each docketed by the bench's terminal or assemble
// register, answered for a side of the docketing's pair, and the outcome lands. A proposal filed after
// a final-epoch plan — of each act, by a lens or blue — is refused naming the plan, and records
// nothing (the chair files no proposal: only a gap's author does, T-B39). Fails if a filing the record knows no plan follows lands, or the exit's proposal has
// no answer that lands.
func TestTheExitDocketsAStandingProposal(t *testing.T) {
	// THE LIVE PATH (L29): a lens author files its proposal on a gap about its gap in the last epoch a
	// lens is dispatched, after the gap's raiser sat there; the raiser sits again for its closings only.
	// The final plan dockets nothing while the proposal stands — with no closing filed, and with both
	// parties' closings filed at the final sitting — and the terminal register dockets it.
	for _, closings := range []bool{false, true} {
		t.Run(map[bool]string{false: "standing at the final sitting: the terminal register dockets it", true: "both closings at the final sitting, nothing docketed: the terminal register dockets it"}[closings], func(t *testing.T) {
			runDir := newRun(t)
			lensGapBoard(t, runDir, 4, 5)
			asSeat(t, runDir, "red-chair", "register")
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); plan.Epoch != 4 || plan.FinalSitting {
				t.Fatalf("plan = %+v, want epoch 4 of 5, the last that dispatches a lens", plan)
			}
			asSeat(t, runDir, otherLens, "register")
			asSeat(t, runDir, lensSeat, "register")
			m := proposeClose(t, runDir, lensSeat, onGapGap, "not_a_defect", "the check names its outcome already")
			asSeat(t, runDir, "red-chair", "register")
			if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); !plan.FinalSitting || !plan.EpochLimitReached {
				t.Fatalf("plan = %+v, want the final sitting at the epoch limit", plan)
			}
			if closings {
				asSeat(t, runDir, otherLens, "register")
				asSeat(t, runDir, otherLens, "closing", "--id", onGapGap, "--reason", "the raiser's case", "--no-text")
				asSeat(t, runDir, lensSeat, "register")
				asSeat(t, runDir, lensSeat, "closing", "--id", onGapGap, "--reason", "the author's case: the proposal states it")
			}
			if ds, _ := docketingsOn(t, runDir, onGapGap); len(ds) != 0 {
				t.Fatalf("docketed before the exit: %+v", ds)
			}
			asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
			if d := lastDocketing(t, runDir, onGapGap); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL || d.GetMotionId() != m {
				t.Fatalf("the terminal register wrote %+v, want PROPOSAL holding %s", d, m)
			}
			if err := benchRules(t, runDir, onGapGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
				t.Fatalf("the bench's ruling for the author: %v", err)
			}
			asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
			asSeat(t, runDir, "judge", "outcome", "--as", "CEILING", "--reason", "the run's term ran out")
		})
	}

	t.Run("an engine stop: the assemble register dockets it", func(t *testing.T) {
		runDir := newRun(t)
		stageRetiredLens(t, runDir, true)
		setEpochLimit(t, runDir, 8)
		m := proposeClose(t, runDir, "blue-respond", limitGap, "not_a_defect", "not converging, and the bench's to settle")
		// No plan follows the filing, and no red closing is filed: the engine stopped.
		asSeat(t, runDir, "judge", "register", "--occasion", "assemble")
		d := lastDocketing(t, runDir, limitGap)
		if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL || d.GetMotionId() != m {
			t.Fatalf("the assemble register wrote %+v, want PROPOSAL holding %s", d, m)
		}
		t.Run("red's side, with no red closing", func(t *testing.T) {
			t.Skip("S3: red's side is refused naming the missing red closing text — needs red's closing text and the for-raiser comparison")
		})
		if err := benchRules(t, runDir, limitGap, recordpb.SidedWith_SIDED_WITH_AUTHOR, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Fatalf("the filer's side at the assembly: %v", err)
		}
		asSeat(t, runDir, "judge", "outcome", "--as", "UNVERIFIED", "--reason", "the engine stopped with the bench's ruling on the record")
	})

	// A FILING AFTER A FINAL-EPOCH PLAN. The board: G1 has had its exchange, and the chair's fourth
	// sitting is the run's last under a limit of 4.
	for _, filer := range []string{otherLens, "blue-respond"} {
		for act, args := range proposalActs(boardGap) {
			t.Run("after the final plan: "+act+" by "+filer, func(t *testing.T) {
				runDir := newRun(t)
				impasseBoard(t, runDir, 3, 1)
				writeTerms(t, runDir, 3, 4)
				if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); !plan.FinalSitting || plan.Epoch != 4 {
					t.Fatalf("plan = %+v, want the final sitting at epoch 4", plan)
				}
				if filer != "red-chair" {
					asSeat(t, runDir, filer, "register")
				}
				// The plan's refusal (HeardInTime) or, for a party, the closings-only sitting's (L44′):
				// each names the run's last epoch under its limit.
				actRefusedNaming(t, runDir, act+" filed after the final plan", filer, []string{"epoch limit of 4", "Nothing was recorded"}, args...)
			})
		}
	}
	// THE CONTROL: two epochs earlier — the last that admits blue's filing, the next holding it out —
	// blue's close lands.
	t.Run("before the final plan: blue's close lands", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 3, 1)
		writeTerms(t, runDir, 3, 6)
		if plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")); plan.FinalSitting || plan.Epoch != 4 {
			t.Fatalf("plan = %+v, want epoch 4 of 6, a sitting before the final one", plan)
		}
		proposeClose(t, runDir, blueSeat, boardGap, "not_a_defect", "not a defect")
	})
}
