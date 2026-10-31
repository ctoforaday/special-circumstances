package fuzz

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// liftLens is the one lens of liftRun's cast.
const liftLens = "red-lens-logic"

// liftRun is a record holding one material gap of liftLens's, minted with the scenario given, and a docket motion blue filed on it: the bench's first hearing stands before it. Every
// coin of its runner lands tails, so every 20% arm alone declines.
func liftRun(t *testing.T, ledger *benchOrderLedger, scenario string) (*runner, string) {
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
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
		[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":1,"k":1,"kMax":2,"mintBudget":1,"convergenceFraction":0.25,"maxEpochs":10}`)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: []string{liftLens, "red-chair", "blue-synthesize", "blue-respond", "judge"}}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, &lockedRand{r: rand.New(steadySource(99))})
	r.benchOrders = ledger
	r.register("blue", "blue-synthesize")
	if _, err := r.do("ingest", "blue-synthesize").result(); err != nil {
		t.Fatalf("ingest the report: %v", err)
	}
	r.chairPlans(t)
	r.dispatchSits("lens", liftLens)
	r.mustExec(t, "--json", "--seat-id", liftLens, "show", "report")
	r.mustExec(t, "--json", "--seat-id", liftLens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", scenario, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	r.dispatchReturns(liftLens, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	// THE MINTER, as r.mint records it for every gap the sweep mints.
	if r.minter == nil {
		r.minter = map[string]string{}
	}
	r.minter[open[0]] = liftLens
	r.dispatchSits("blue", "blue-respond")
	r.mustExec(t, "motion", "docket", "file", "--seat-id", "blue-respond", "--id", open[0], "--reason", "fuzz: the bench's to settle")
	r.dispatchReturns("blue-respond", false)
	return r, open[0]
}

// chairPlans is the chair's sitting that asks the dispatch for the plan and returns.
func (r *runner) chairPlans(t *testing.T) {
	t.Helper()
	r.dispatchSits("chair", "red-chair")
	r.mustExec(t, "--json", "dispatch", "next", "--seat-id", "red-chair")
	r.dispatchReturns("red-chair", false)
}

// mustExec is exec failing the test on a refusal.
func (r *runner) mustExec(t *testing.T, args ...string) string {
	t.Helper()
	out, err := r.exec(args...)
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return out
}

// hearing is the bench's docket sitting over gapID as the drive sits it: the bench sets a grade, then
// remands the gap.
func (r *runner) hearing(t *testing.T, gapID string) {
	t.Helper()
	r.chairPlans(t)
	r.dispatchSits("bench", "judge", "docket")
	r.benchSetsGrade("judge", gapID)
	r.benchDisposes("judge", gapID, "remanded", "--direction", "fuzz: one more exchange", "--reason", "fuzz: remanded for "+gapID)
	r.dispatchReturns("judge", false)
	if r.actErr != "" {
		t.Fatalf("the bench's hearing: %s", r.actErr)
	}
}

// THE SWEEP'S LEDGER DRIVES THE LIFT, not the coin: a run whose sweep still owes the lens's move of an
// axis the bench released takes each act of the sequence — the bench locks an axis at the first
// hearing, the lens is refused moving it, blue puts the gap before the bench again, the bench's second
// move releases the axis, and the lens moves it — and the ledger then owes nothing. Every coin lands
// tails, so the 20% arms alone would take none of it, and only the ledger drives this.
func TestTheSweepsLedgerDrivesTheLensMoveOfAReleasedAxis(t *testing.T) {
	ledger := &benchOrderLedger{}
	r, gap := liftRun(t, ledger, dirIgnore)

	r.hearing(t, gap)
	held := r.benchGrades(gap)
	if len(held) != 1 || r.benchRegrades != 1 {
		t.Fatalf("after the first hearing the bench holds %v by %d regrade(s), want one axis locked by one", held, r.benchRegrades)
	}

	r.dispatchSits("lens", liftLens)
	if got := r.lensMovesBenchAxis(liftLens); got != gap || r.lockRefused != 1 {
		t.Fatalf("the lens regraded %q with %d move(s) of the locked axis refused, want %s regraded after one refusal: %s", got, r.lockRefused, gap, r.actErr)
	}
	r.dispatchReturns(liftLens, false)

	r.dispatchSits("blue", "blue-respond")
	r.docketLockedGap("blue-respond")
	r.dispatchReturns("blue-respond", false)
	if len(r.unruledDocketMotions(gap)) != 1 {
		t.Fatalf("blue put no docket motion before the bench on %s, which the bench holds %v of: %s", gap, held, r.actErr)
	}

	r.hearing(t, gap)
	if left := r.benchGrades(gap); len(left) != 0 || len(r.benchLifted[gap]) != 1 {
		t.Fatalf("after the second hearing the bench holds %v and released %v, want the locked axis released", left, r.benchLifted[gap])
	}

	r.dispatchSits("lens", liftLens)
	if got := r.lensMovesReleasedAxis(liftLens, ""); got != gap || r.lockLifted != 1 {
		t.Fatalf("the lens regraded %q with %d move(s) of a released axis admitted, want %s once: %s", got, r.lockLifted, gap, r.actErr)
	}
	r.dispatchReturns(liftLens, false)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if ledger.owesLift() {
		t.Error("the ledger still owes the lift after the lens's move of the released axis was admitted")
	}
}

// A RUNNER WITHOUT THE SWEEP'S LEDGER DRAWS ON THE COIN ALONE: on a tails coin the bench sets no grade
// at its hearing, and blue files no docket motion to convene another.
func TestARunnerWithoutTheLedgerDrawsTheLiftOnTheCoin(t *testing.T) {
	r, gap := liftRun(t, nil, dirIgnore)
	r.hearing(t, gap)
	if r.benchRegrades != 0 || len(r.benchGrades(gap)) != 0 {
		t.Fatalf("the bench regraded %d time(s) and holds %v on a tails coin with no ledger, want neither", r.benchRegrades, r.benchGrades(gap))
	}
	// A lock the drive did not take: blue, owing no lift, leaves it alone.
	r.benchSet, r.benchLifted = map[string]map[string]string{gap: {"severity": "high"}}, map[string]map[string]bool{}
	r.dispatchSits("blue", "blue-respond")
	r.docketLockedGap("blue-respond")
	r.dispatchReturns("blue-respond", false)
	if ids := r.unruledDocketMotions(gap); len(ids) != 0 {
		t.Errorf("blue owing no lift filed docket motion(s) %v on %s", ids, gap)
	}
}

// AN OWED LIFT KEEPS THE LOCK WHERE THE HEARING CLOSES THE GAP: released there, the axis would leave
// the board with it, and no lens could move it.
func TestTheBenchKeepsTheLockAtAHearingThatClosesTheGap(t *testing.T) {
	r, gap := liftRun(t, &benchOrderLedger{}, dirDisputeLost)
	r.benchSet, r.benchLifted = map[string]map[string]string{gap: {"severity": r.currentGrade(gap, "severity")}}, map[string]map[string]bool{}
	r.chairPlans(t)
	r.dispatchSits("bench", "judge", "docket")
	r.benchSetsGrade("judge", gap)
	if r.benchRegrades != 0 || len(r.benchGrades(gap)) != 1 {
		t.Errorf("the bench moved %d axis(es) and holds %v at a hearing its scenario closes, want the lock kept", r.benchRegrades, r.benchGrades(gap))
	}
}

// steadySource draws n on every Intn(100): its 31-bit draw is n, below every rejection bound.
type steadySource int64

func (s steadySource) Int63() int64 { return int64(s) << 32 }
func (steadySource) Seed(int64)     {}
