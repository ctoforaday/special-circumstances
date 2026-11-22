package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// actRefusedNaming runs seat's act, expecting a refusal whose text holds every want, and that the act
// recorded nothing. A gap that moved under the seat's sitting refuses the first call with the move;
// that refusal is not the one under test, so the act is asked once more.
func actRefusedNaming(t *testing.T, runDir, what, seat string, want []string, args ...string) {
	t.Helper()
	before := motionsOnRecord(t, runDir)
	argv := append(append([]string{}, args...), "--run", runDir, "--seat-id", seat)
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
	if after := motionsOnRecord(t, runDir); after != before {
		t.Errorf("%s: the refusal left %d motion(s) on the record, want %d", what, after, before)
	}
}

// T-B22: A DOCKETED GAP TAKES NO PROPOSAL AT EVERY K-MAX (L38). G1 reaches impasse; `dispatch next`
// writes Docketing{IMPASSE} and no motion; in the closing sitting it readies, blue's proposal and a
// non-assignee lens's are refused naming the docketing; the parties' closings land; the bench's docket
// sitting holds the docketing and rules it, naming it and no motion. At k-max 1 no proposal lands
// between the mint and the ruling; at k-max 2 a proposal filed after one exchange defers the impasse
// docketing (L28), its window ends in Docketing{PROPOSAL, M}, and the bench's ruling names M and D.
// Fails if a proposal lands on a docketed gap, the bench cannot rule D, or the dispatch writes a motion.
func TestADocketedGapTakesNoProposalAtEveryKMax(t *testing.T) {
	for _, kMax := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("k-max %d", kMax), func(t *testing.T) {
			runDir := newRun(t)
			// At k-max 1 the one exchange reaches impasse; above it, two stalled exchanges do.
			impasseBoard(t, runDir, kMax, min(kMax, 2))
			if kMax == 1 {
				// BEFORE THE EXCHANGE that reaches impasse, a proposal has had none and is refused.
				early := newRun(t)
				impasseBoard(t, early, 1, 0)
				actRefusedNaming(t, early, "blue's proposal before the gap's one exchange", "blue-respond", []string{"exchange"},
					append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "not a defect"}, closeFlags("not_a_defect")...)...)
				t.Log("what L38 still costs (§III.B13): at k-max 1 the ≥1-exchange rule is first met by the exchange that reaches impasse, and the docketing then freezes the gap — no proposal is reachable on it, and blue's position is its closing")
			}
			plan := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			if len(plan.Docket) != 1 || plan.Docket[0] != boardGap {
				t.Fatalf("plan = %+v, want %s docketed at impasse", plan, boardGap)
			}
			if n := motionsOnRecord(t, runDir); n != 0 {
				t.Fatalf("the dispatch left %d motion(s) on the record, want none", n)
			}
			d := lastDocketing(t, runDir, boardGap)
			if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE || d.MotionId != nil {
				t.Fatalf("the docketing = %+v, want impasse holding no proposal", d)
			}
			id := d.GetDocketingId()

			// THE CLOSING SITTING: blue and a lens that is not the gap's assignee sit, and neither may
			// propose on the docketed gap.
			for _, seat := range []string{"blue-respond", otherLens, "red-lens-evidence"} {
				asSeat(t, runDir, seat, "register")
			}
			actRefusedNaming(t, runDir, "blue's proposal of a close on the docketed gap", "blue-respond", []string{"docketing " + id},
				append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "not a defect"}, closeFlags("not_a_defect")...)...)
			actRefusedNaming(t, runDir, "a non-assignee lens's proposal of a regrade on the docketed gap", otherLens, []string{"docketing " + id},
				"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
			asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--as", "defect_owed_elsewhere", "--reason", "the report already owes this elsewhere")
			asSeat(t, runDir, "red-lens-evidence", "closing", "--id", boardGap, "--reason", "the section still misleads")

			// THE BENCH'S DOCKET SITTING holds D and rules it: no motion id, docketing id D.
			asSeat(t, runDir, "red-chair", "register")
			bench := planOf(t, asSeat(t, runDir, "red-chair", "dispatch", "next", "--json"))
			benchOn := false
			for _, p := range bench.Parties {
				benchOn = benchOn || (p.SeatID == "judge" && len(p.GapIDs) == 1 && p.GapIDs[0] == boardGap)
			}
			if !benchOn {
				t.Fatalf("plan after both closings = %+v, want the bench readied on %s", bench, boardGap)
			}
			asSeat(t, runDir, "judge", "register", "--occasion", "docket")
			if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_BLUE, recordpb.Disposition_DISPOSITION_DEFECT_OWED_ELSEWHERE); err != nil {
				t.Fatalf("the bench's ruling of %s, for blue: %v", id, err)
			}
			rule := lastDocketRuling(t, runDir)
			if rule.MotionId != nil || rule.GetDocket().GetDocketingId() != id {
				t.Errorf("the bench's ruling = %+v, want docketing %s and no motion", rule, id)
			}
			t.Run("the ruling tombstone names the docketing", func(t *testing.T) {
				requireRulingTombstone(t, runDir, boardGap, id)
			})
			if n := motionsOnRecord(t, runDir); n != 0 {
				t.Errorf("%d motion(s) landed between the mint and the ruling, want none", n)
			}
		})
	}

	// AT K-MAX 2 A PROPOSAL FILED AFTER ONE EXCHANGE outranks the impasse its second exchange reaches
	// (L28): the window's end writes Docketing{PROPOSAL, M}, and the bench's ruling names M and D.
	t.Run("k-max 2, a proposal after one exchange", func(t *testing.T) {
		runDir := newRun(t)
		impasseBoard(t, runDir, 2, 1)
		m := proposeClose(t, runDir, "blue-respond", boardGap, "not_a_defect", "not a defect, and the lens will not close it")
		asSeat(t, runDir, "red-chair", "register")
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		if ds, _ := docketingsOn(t, runDir, boardGap); len(ds) != 0 {
			t.Fatalf("the plan docketed %s with proposal %s in its window: %+v", boardGap, m, ds)
		}
		asSeat(t, runDir, "blue-respond", "register")
		asSeat(t, runDir, "blue-respond", "closing", "--id", boardGap, "--reason", "not a defect")
		asSeat(t, runDir, "red-lens-evidence", "register")
		asSeat(t, runDir, "red-lens-evidence", "closing", "--id", boardGap, "--reason", "the section still misleads")
		d := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap)
		if d.GetMotionId() != m {
			t.Fatalf("the window's end wrote %+v, want PROPOSAL holding %s", d, m)
		}
		if ds, _ := docketingsOn(t, runDir, boardGap); len(ds) != 1 {
			t.Fatalf("docketings = %+v, want the proposal's one", ds)
		}
		actRefusedNaming(t, runDir, "the chair's proposal on the docketed gap", "red-chair", []string{"docketing " + d.GetDocketingId()},
			append([]string{"motion", "docket", "file", "close", "--id", boardGap, "--reason", "a second ask"}, closeFlags("not_a_defect")...)...)
		asSeat(t, runDir, "red-chair", "register")
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		asSeat(t, runDir, "judge", "register", "--occasion", "docket")
		if err := benchRules(t, runDir, boardGap, recordpb.SidedWith_SIDED_WITH_BLUE, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT); err != nil {
			t.Fatalf("the bench's ruling of %s: %v", d.GetDocketingId(), err)
		}
		if rule := lastDocketRuling(t, runDir); rule.GetMotionId() != m || rule.GetDocket().GetDocketingId() != d.GetDocketingId() {
			t.Errorf("the bench's ruling = %+v, want motion %s and docketing %s", rule, m, d.GetDocketingId())
		}
	})
}

// lastDocketRuling is the newest docket ruling on the record.
func lastDocketRuling(t *testing.T, runDir string) *recordpb.MotionRule {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_MOTION_RULE)
	if err != nil {
		t.Fatal(err)
	}
	for i := len(evs) - 1; i >= 0; i-- {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](evs[i]); ok && r.GetDocket() != nil {
			return r
		}
	}
	t.Fatal("no docket ruling on the record")
	return nil
}

// requireRulingTombstone fails unless a ruling tombstone on gap names docketing. It skips on a record
// holding no tombstone row at all: the write site, writeTombstones in writeIn, is phase B3's.
func requireRulingTombstone(t *testing.T, runDir, gap, docketing string) {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), recordpb.EventType_EVENT_TYPE_TOMBSTONE)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) == 0 {
		t.Skip("B3: no tombstone row on the record — writeTombstones in writeIn is not built")
	}
	for _, e := range evs {
		if s, ok := recordpb.BodyAs[*recordpb.Tombstone](e); ok && s.GetKind() == recordpb.TombstoneKind_TOMBSTONE_KIND_RULING && s.GetGapId() == gap {
			if s.GetDocketingId() != docketing {
				t.Errorf("the ruling tombstone on %s names %q, want docketing %s", gap, s.GetDocketingId(), docketing)
			}
			return
		}
	}
	t.Errorf("no ruling tombstone on %s after the bench ruled docketing %s (%d tombstone row(s) on the record)", gap, docketing, len(evs))
}
