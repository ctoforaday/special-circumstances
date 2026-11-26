package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// closingParties is the seats a plan readies, failing the test where one is engaged on a gap: a party
// sitting for its closings is engaged on none.
func closingParties(t *testing.T, plan record.Plan) []string {
	t.Helper()
	var seats []string
	for _, p := range plan.Parties {
		seats = append(seats, p.SeatID)
		if len(p.GapIDs) != 0 {
			t.Errorf("%s is engaged on %v; a party sitting for its closings is engaged on no gap", p.SeatID, p.GapIDs)
		}
	}
	return seats
}

// T-B37: RED GETS ITS CLOSING AT THE FINAL SITTING (amendment 5, L44′), and held-out blue its own
// (amendment 6). A gap docketed at impasse by the FinalSitting plan, and one docketed before it, each
// ready the gap's assignee and blue for their closings before the terminal bench: the plan's parties
// are exactly those, and the plan says the limit is reached. In that sitting each closing lands and
// every other act of a party is refused naming the final sitting; the terminal bench rules the
// docketing. In the epoch that holds blue out, a docketing readies blue for its closing alone; its
// closing lands and its other acts are refused naming the held-out epoch.
func TestRedGetsItsClosingAtTheFinalSitting(t *testing.T) {
	const lens = "red-lens-evidence"
	final := []string{"epoch limit of 5", "this is the final sitting", "Nothing was recorded"}

	// inTheClosingSitting is the parties' closing sitting on boardGap at the final epoch: each sits,
	// each other act is refused naming the final sitting, and each closing lands.
	inTheClosingSitting := func(t *testing.T, runDir string) {
		t.Helper()
		for _, seat := range []string{lens, "blue-respond"} {
			asSeat(t, runDir, seat, "register")
		}
		actRefusedOnRecord(t, runDir, "the lens's regrade", lens, final, "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
		actRefusedOnRecord(t, runDir, "the lens's mint", lens, final, "mint", "--about-kind", "section", "--about", "Findings", "--class", "spec_underspecification",
			"--problem", "p", "--fix", "f", "--check-kind", "document", "--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--reason", "late")
		actRefusedOnRecord(t, runDir, "blue's edit", "blue-respond", final, "edit", "--quote", "# report", "--new", "# the report", "--answers", boardGap, "--reason", "late")
		actRefusedNaming(t, runDir, "blue's proposal", "blue-respond", []string{"epoch limit of 5", "Nothing was recorded"},
			append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "late"}, closeFlags("not_a_defect")...)...)
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "blue's position at the final sitting")
		asSeat(t, runDir, lens, "closing", "--id", boardGap, "--reason", "red's position at the final sitting")
	}
	// terminalBench is the terminal bench's ruling of the docketing on boardGap, for blue.
	terminalBench := func(t *testing.T, runDir string) {
		t.Helper()
		asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
		t.Run("--for-red applies red's text; with red's closing not filed it is refused", func(t *testing.T) {
			t.Skip("S3: needs `motion docket rule --for-red` and red's closing text (the for-red comparison)")
		})
		if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_BLUE, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Fatalf("the terminal bench's ruling for blue: %v", err)
		}
	}

	t.Run("docketed at impasse by the final plan", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 2) // the chair's fifth sitting is the closing register
		writeTerms(t, runDir, 2, 5)
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if d := lastDocketing(t, runDir, boardGap); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
			t.Fatalf("the final plan's docketing = %+v, want impasse", d)
		}
		if seats := closingParties(t, plan); !plan.FinalSitting || !plan.EpochLimitReached || !slices.Equal(seats, []string{lens, "blue-respond"}) {
			t.Fatalf("plan = %+v, want the final sitting at the limit, readying exactly %s's and blue's closings", plan, lens)
		}
		inTheClosingSitting(t, runDir)
		terminalBench(t, runDir)
	})

	t.Run("docketed before the final plan", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 2)
		writeTerms(t, runDir, 2, 6) // the docketing's plan is epoch 5, which holds blue out
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		if d := lastDocketing(t, runDir, boardGap); d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
			t.Fatalf("the docketing = %+v, want impasse", d)
		}
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if seats := closingParties(t, plan); !plan.FinalSitting || !plan.EpochLimitReached || !slices.Equal(seats, []string{lens, "blue-respond"}) {
			t.Fatalf("plan = %+v, want the final sitting at the limit, readying exactly %s's and blue's closings", plan, lens)
		}
		final = []string{"epoch limit of 6", "this is the final sitting", "Nothing was recorded"}
		defer func() { final = []string{"epoch limit of 5", "this is the final sitting", "Nothing was recorded"} }()
		for _, seat := range []string{lens, "blue-respond"} {
			asSeat(t, runDir, seat, "register")
		}
		actRefusedOnRecord(t, runDir, "the lens's regrade", lens, final, "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "blue's position at the final sitting")
		asSeat(t, runDir, lens, "closing", "--id", boardGap, "--reason", "red's position at the final sitting")
		terminalBench(t, runDir)
	})

	t.Run("a proposal standing at the final sitting", func(t *testing.T) {
		runDir := newRun(t)
		stageRetiredLens(t, runDir, true)
		setEpochLimit(t, runDir, 5)
		m := proposeClose(t, runDir, "red-chair", limitGap, "not_a_defect", "not converging, and the bench's to settle")
		asSeat(t, runDir, "red-chair", "register")
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if seats := closingParties(t, plan); !plan.FinalSitting || !slices.Equal(seats, []string{lens, "blue-respond"}) {
			t.Fatalf("plan = %+v, want the final sitting readying exactly the proposal's two closings", plan)
		}
		asSeat(t, runDir, lens, "register")
		actRefusedOnRecord(t, runDir, "the assignee's accept", lens, final, "motion", "docket", "accept", "--id", m, "--reason", "the close is right")
		argue(t, runDir)
		asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
		ruleProposalDocketing(t, runDir, m)
	})

	t.Run("held-out blue", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 2)
		writeTerms(t, runDir, 2, 6) // epoch 5 holds blue out
		plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if plan.EpochLimitReached || plan.FinalSitting {
			t.Fatalf("plan = %+v, want epoch 5 of 6: blue held out, the limit not reached", plan)
		}
		blueOn := -1
		for i, p := range plan.Parties {
			if p.SeatID == "blue-respond" {
				blueOn = i
			}
		}
		if blueOn < 0 || len(plan.Parties[blueOn].GapIDs) != 0 {
			t.Fatalf("plan = %+v, want blue readied for its closing on %s alone, engaged on no gap", plan, boardGap)
		}
		said := false
		for _, w := range plan.Why {
			said = said || (strings.Contains(w, "blue-respond") && strings.Contains(w, "closing"))
		}
		if !said {
			t.Errorf("the plan's why does not name blue's closing: %q", plan.Why)
		}
		held := []string{"blue is held out of it", "Nothing was recorded"}
		asSeat(t, runDir, "blue-respond", "register")
		actRefusedOnRecord(t, runDir, "held-out blue's edit", "blue-respond", held, "edit", "--quote", "# report", "--new", "# the report", "--answers", boardGap, "--reason", "late")
		t.Run("held-out blue's retire and inquiry fate", func(t *testing.T) {
			t.Skip("no citation, computation or inquiry on this board to retire or fate; the act judge's held-out arm is the edit's above")
		})
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "blue's position, held out")
		asSeat(t, runDir, lens, "register")
		asSeat(t, runDir, lens, "closing", "--id", boardGap, "--reason", "red's position")
		// The next chair sitting opens the final epoch; no closing is awaited, so the docketing waits
		// on the bench alone.
		asSeat(t, runDir, "red-chair", "register")
		next := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
		if !next.FinalSitting || slices.ContainsFunc(next.Parties, func(p record.Party) bool { return p.SeatID == "blue-respond" || p.SeatID == lens }) ||
			!slices.Contains(next.Blockers, record.PlanBlocker{Kind: record.BlockerDocketed, Subject: boardGap, Owner: "judge"}) {
			t.Errorf("the plan opening the final epoch = %+v, want no closing party and the docketing the bench's", next)
		}
	})
}

// actRefusedOnRecord is actRefusedNaming for an act that is not a motion: the refusal holds every want,
// and the record's event count moves by no more than the refusal's own log row.
func actRefusedOnRecord(t *testing.T, runDir, what, seat string, want []string, args ...string) {
	t.Helper()
	argv := append(append([]string{}, args...), "--run", runDir, "--seat-id", seat)
	// The delivery the act would stand on is the seat's before it acts, not a row the refusal left.
	deliverBeforeActing(t, argv, seat)
	before := len(events(t, runDir))
	out, err := runAt(t, argv...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, argv...)
	}
	if err == nil {
		t.Errorf("%s was recorded: %s", what, out)
		return
	}
	for _, w := range want {
		if !strings.Contains(err.Error()+out, w) {
			t.Errorf("%s: the refusal does not say %q: %v", what, w, err)
		}
	}
	if after := len(events(t, runDir)); after > before+1 {
		t.Errorf("%s: the refusal left %d event(s) on the record, want at most %d", what, after, before+1)
	}
}
