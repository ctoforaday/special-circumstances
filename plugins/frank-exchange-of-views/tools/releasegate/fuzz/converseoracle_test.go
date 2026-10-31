package fuzz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// converseLens is the one lens of converseRun's cast.
const converseLens = "red-lens-logic"

// conversePath is how converseRun's one gap ends and what stands at the last chair sitting.
type conversePath int

const (
	// redCloses: the lens closes its gap, and nothing else is filed.
	redCloses conversePath = iota
	// redClosesAndPetitions: the lens closes its gap and files a petition.
	redClosesAndPetitions
	// benchCloses: blue dockets the gap, and the bench closes it at a docket sitting of epoch 2.
	benchCloses
	// redClosesAndWithholds: the lens closes its gap, and the chair's last sitting withholds the PASS
	// its plan permits, files its position, and asks for the plan again.
	redClosesAndWithholds
	// redClosesAndWithholdsEarly is redClosesAndWithholds one epoch before the limit.
	redClosesAndWithholdsEarly
)

// converseRun is a record at its epoch limit, over one gap that ends as path says: at epoch 1 the
// lens reads the report and mints the gap; the chair then sits for the last epoch and stores its
// plan. It returns the runner and the plan that sitting printed.
func converseRun(t *testing.T, path conversePath) (*runner, record.Plan) {
	t.Helper()
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	const sentence = "A § fuzz sentence to anchor findings."
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte("# § fuzz\n\n"+sentence+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	maxEpochs := 2
	if path == benchCloses || path == redClosesAndWithholdsEarly {
		maxEpochs = 3
	}
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
		fmt.Appendf(nil, `{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":1,"k":1,"kMax":2,"mintBudget":1,"convergenceFraction":0.25,"maxEpochs":%d}`, maxEpochs)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: []string{converseLens, "red-chair", "blue-synthesize", "blue-respond", "judge"}}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(7))
	must := func(args ...string) string {
		t.Helper()
		out, err := r.exec(args...)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	r.register("blue", "blue-synthesize")
	if _, err := r.do("ingest", "blue-synthesize").result(); err != nil {
		t.Fatalf("ingest the report: %v", err)
	}

	// EACH SITTING ENDS AT ITS RETURN, as a dispatched seat's does (dispatchReturns). These seats put
	// no envelope to the record, so none is admitted.
	r.dispatchSits("chair", "red-chair")
	must("--json", "dispatch", "next", "--seat-id", "red-chair")
	r.dispatchReturns("red-chair", false)
	r.dispatchSits("lens", converseLens)
	must("--json", "--seat-id", converseLens, "show", "report")
	must("--json", "--seat-id", converseLens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", "fuzz fix", "--severity", "medium", "--impact", "medium", "--likelihood", "medium", "--reason", "fuzz reason")
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	switch path {
	case redCloses, redClosesAndPetitions, redClosesAndWithholds, redClosesAndWithholdsEarly:
		must("--seat-id", converseLens, "close", "--id", open[0], "--as", "not_a_defect", "--reason", "fuzz: the argument held",
			"--verified-by", converseLens, "--verified-with", "fuzz", "--verified-against", "the report")
		if path == redClosesAndPetitions {
			must("--json", "motion", "petition", "file", "--seat-id", converseLens, "--class", "safety",
				"--relief", "fuzz relief", "--reason", "fuzz petition basis")
		}
		r.dispatchReturns(converseLens, false)
	case benchCloses:
		r.dispatchReturns(converseLens, false)
		r.register("blue", "blue-respond")
		must("motion", "docket", "file", "--seat-id", "blue-respond", "--id", open[0], "--reason", "fuzz: the bench's to settle")
		r.dispatchSits("chair", "red-chair")
		must("--json", "dispatch", "next", "--seat-id", "red-chair")
		r.dispatchReturns("red-chair", false)
		r.dispatchSits("bench", "judge", "docket")
		r.benchDisposes("judge", open[0], "not_a_defect", "--reason", "fuzz: the argument held", "--final")
		if r.actErr != "" {
			t.Fatalf("the bench's ruling: %s", r.actErr)
		}
		r.dispatchReturns("judge", false)
	}

	r.dispatchSits("chair", "red-chair")
	out := must("--json", "dispatch", "next", "--seat-id", "red-chair")
	if path == redClosesAndWithholds || path == redClosesAndWithholdsEarly {
		must("--json", "position", "--seat-id", "red-chair", "--reason", "fuzz: the PASS is withheld, the core claim rests on one source")
		out = must("--json", "dispatch", "next", "--seat-id", "red-chair")
	}
	r.dispatchReturns("red-chair", false)
	requireEveryDispatchReturned(t, r)
	var plan record.Plan
	if err := json.Unmarshal([]byte(out), &plan); err != nil {
		t.Fatalf("the plan of the last chair sitting does not read: %v\n%s", err, out)
	}
	return r, plan
}

// THE CONVERSE VERDICT ORACLE JUDGES THE PLAN THE DEBATE ENDED ON, not the board after the exit —
// pinned deterministically, because the sweep reaches it on about one run in forty (seed 3,
// 2026-10-28).
//
// A petition filed in the epoch before the limit readies the bench at the final chair sitting, the
// limit leaves it undispatched, and the run ends CEILING (record's epochLimitHeldTheBench). The bench
// rules the petition at the exit, so the plan recomputed after it names no blocker: read there, the
// oracle reported a CEILING nothing explains over a board red emptied.
func TestTheConverseVerdictOracleJudgesThePlanTheDebateEndedOn(t *testing.T) {
	t.Run("a petition the bench rules at the exit holds the run at its ceiling", func(t *testing.T) {
		r, last := converseRun(t, redClosesAndPetitions)
		var petitionID string
		for _, b := range last.Blockers {
			if b.Kind == record.BlockerUnruledMotion && b.Owner == "judge" {
				petitionID = b.Subject
			}
		}
		if !last.EpochLimitReached || !last.Ceiling || petitionID == "" {
			t.Fatalf("the last chair sitting's plan = %+v, want the epoch limit reached over the bench readied for the petition", last)
		}
		r.dispatchSits("bench", "judge", "terminal")
		if _, err := r.exec("motion", "petition", "rule", "--seat-id", "judge", "--id", petitionID,
			"--as", "denied", "--reason", "fuzz: denied on its merits"); err != nil {
			t.Fatal(err)
		}
		r.dispatchReturns("judge", false)
		requireEveryDispatchReturned(t, r)
		if got, why, ok, err := record.DeriveVerdict(r.run()); err != nil || !ok || got != "CEILING" {
			t.Fatalf("derived (%q, %q, %v, %v), want CEILING: the limit held the bench", got, why, ok, err)
		}
		board, err := record.FamilyOf(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if m := converseVerdictOracle("CEILING", board); m != "" {
			t.Errorf("the oracle fired over a CEILING the plan it ended on explains: %s", m)
		}
		// THE READ IT REPLACES FIRES HERE: the plan recomputed after the exit has lost the petition.
		after, err := record.PlanDispatch(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if !unfinishedOverRedsWork("CEILING", len(board.Gaps), 0, 0, after.Blockers) {
			t.Errorf("the plan recomputed after the exit = %+v; this test no longer reproduces the board the sweep reached", after)
		}
	})
	// A PASS THE CHAIR WITHHELD AT ITS FINAL SITTING IS HEARD BY THE TERMINAL BENCH, and the run ends at
	// its ceiling however it is ruled (gblock, 2026-10-30). Before the bench heard it, nobody was
	// readied and the run ended UNVERIFIED over a board red emptied, with nothing on the plan to say
	// why; now the plan it ended on holds the position as the bench's motion.
	for _, ruling := range [][]string{{"--as", "upheld", "--lens", converseLens, "--direction", "fuzz: audit the claim the chair would not pass"}, {"--as", "overruled"}} {
		t.Run("a PASS withheld at the final sitting, ruled "+ruling[1]+" at the exit, holds the run at its ceiling", func(t *testing.T) {
			r, last := converseRun(t, redClosesAndWithholds)
			var positionID string
			for _, b := range last.Blockers {
				if b.Kind == record.BlockerUnruledMotion && b.Owner == "judge" {
					positionID = b.Subject
				}
			}
			if !last.EpochLimitReached || !last.Ceiling || positionID == "" {
				t.Fatalf("the plan after the position = %+v, want the epoch limit reached over the bench readied for the position", last)
			}
			r.dispatchSits("bench", "judge", "terminal")
			if _, err := r.exec(append([]string{"motion", "position", "rule", "--seat-id", "judge", "--id", positionID, "--reason", "fuzz: ruled on its merits"}, ruling...)...); err != nil {
				t.Fatal(err)
			}
			r.dispatchReturns("judge", false)
			requireEveryDispatchReturned(t, r)
			if got, why, ok, err := record.DeriveVerdict(r.run()); err != nil || !ok || got != "CEILING" {
				t.Fatalf("derived (%q, %q, %v, %v), want CEILING: the limit held the bench on the position", got, why, ok, err)
			}
			board, err := record.FamilyOf(r.run())
			if err != nil {
				t.Fatal(err)
			}
			if m := converseVerdictOracle("CEILING", board); m != "" {
				t.Errorf("the oracle fired over a CEILING the position the plan ended on explains: %s", m)
			}
		})
	}
	t.Run("a CEILING the plan it ended on does not explain still fires", func(t *testing.T) {
		r, last := converseRun(t, redCloses)
		if last.Ceiling || !last.PassPermitted {
			t.Fatalf("the last chair sitting's plan = %+v, want PASS permitted and no ceiling", last)
		}
		if got, _, ok, err := record.DeriveVerdict(r.run()); err != nil || ok {
			t.Fatalf("derived (%q, %v, %v), want no verdict: no PASS, no halt, no ceiling", got, ok, err)
		}
		board, err := record.FamilyOf(r.run())
		if err != nil {
			t.Fatal(err)
		}
		for _, recorded := range []string{"CEILING", "UNVERIFIED"} {
			if m := converseVerdictOracle(recorded, board); m == "" {
				t.Errorf("the oracle held over %s on a board red emptied, ended on a plan whose blockers are the chair's own", recorded)
			}
		}
	})
	t.Run("gaps the bench closed explain a run that did not pass", func(t *testing.T) {
		r, _ := converseRun(t, benchCloses)
		board, err := record.FamilyOf(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if len(board.Gaps) != 1 || board.Gap(board.Gaps[0].ID) == nil || !board.Gap(board.Gaps[0].ID).ClosedByBench {
			t.Fatalf("the board = %+v, want its one gap closed by the bench", board.Gaps)
		}
		if m := converseVerdictOracle("UNVERIFIED", board); m != "" {
			t.Errorf("the oracle fired over a board the bench closed: %s", m)
		}
	})
}
