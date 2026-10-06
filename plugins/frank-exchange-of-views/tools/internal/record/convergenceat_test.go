package record

import (
	"strconv"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A past verdict is judged on the grades AS THEY STOOD AT IT: every gap open at the verdict, each
// graded by its latest regrade at or before it (else its mint), against the run's own fraction. The
// refusal at the write path asks the same question at the record's newest event, so the scorecard
// and the gate give every FAIL the same answer.

func gradedMint(gap string, sev, lik, imp recordpb.Grade) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(sev), Likelihood: recordtest.P(lik), Impact: recordtest.P(imp)}
}

func gradedRegrade(gap string, sev, lik, imp recordpb.Grade) *recordpb.Regrade {
	return &recordpb.Regrade{GapId: proto.String(gap), Severity: recordtest.P(sev), Likelihood: recordtest.P(lik),
		Impact: recordtest.P(imp), Basis: proto.String("moved")}
}

func repairClose(gap string) *recordpb.Close {
	return &recordpb.Close{GapId: proto.String(gap), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}
}

var failVerdict = func() *recordpb.Gate { return &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)} }

const (
	gLow  = recordpb.Grade_GRADE_LOW
	gMed  = recordpb.Grade_GRADE_MEDIUM
	gHigh = recordpb.Grade_GRADE_HIGH
)

// setFraction writes the run's terms with the given convergence fraction, as setup does.
func setFraction(t *testing.T, run Run, f float64) {
	t.Helper()
	writeRunConfig(t, run.Dir(), `{"convergenceFraction":`+strconv.FormatFloat(f, 'g', -1, 64)+`}`)
}

// B4: G1 (high/high) minted, FAILed over and closed; G2 minted low/low and regraded to high/high
// before the second FAIL. The open mass at FAIL #2 is 9 against a peak of 9 — not converged — so
// the write path admits it, and the scorecard must not call it divergent.
func TestAFailTheGateAdmitsIsNotDivergentAfterTheFact(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).
		add(evLens, repairClose("G1")).
		add(evLens, gradedMint("G2", gLow, gLow, gLow)).
		add(evLens, &recordpb.Regrade{GapId: proto.String("G2"), Likelihood: recordtest.P(gHigh),
			Impact: recordtest.P(gHigh), Basis: proto.String("worse than it looked")}).
		register("red-chair")
	run := b.seed()
	gate, err := convergenceOf(run)
	if err != nil {
		t.Fatal(err)
	}
	if gate.Holds || gate.Mass != 9 || gate.Peak != 9 {
		t.Fatalf("the gate's convergence before FAIL #2 = %+v, want mass 9 on the regraded G2, peak 9, not holding", gate)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, failVerdict()); err != nil {
		t.Fatalf("FAIL #2 over a board of mass 9 against a peak of 9 was refused: %v", err)
	}
	eps, err := ConvergenceVsVerdict(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("verdict rows = %+v, want 2", eps)
	}
	if got := eps[1]; got.Divergent || got.Mass != 9 || got.Peak != 9 {
		t.Errorf("FAIL #2 after the fact = %+v, want mass 9 (G2 as regraded before it), peak 9, not divergent — the gate admitted it", got)
	}
}

// A regrade AFTER a verdict does not move that verdict: the scorecard judges a past act on the facts
// it was issued over, not on later ones.
func TestARegradeAfterAVerdictDoesNotMoveIt(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).
		add(evLens, repairClose("G1")).
		add(evLens, gradedMint("G2", gLow, gLow, gLow)).
		add(evLens, gradedRegrade("G2", gLow, gHigh, gHigh)).
		register("red-chair").
		add("red-chair", failVerdict()).
		// After FAIL #2: G2 falls back to low/low. FAIL #2 stood over mass 9 — neither its mint
		// grades (1) nor its current ones (1) — and stays there.
		add(evLens, gradedRegrade("G2", gLow, gLow, gLow))
	eps, err := ConvergenceVsVerdict(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("verdict rows = %+v, want 2", eps)
	}
	if got := eps[1]; got.Mass != 9 || got.Divergent {
		t.Errorf("FAIL #2 = %+v, want mass 9 — a regrade written after it does not reach back", got)
	}
}

// A verdict's PEAK is the run's peak as it stood at that verdict, not one a later board set. FAIL #1
// stands over a trifle of mass 1 — its own peak, nothing converged — and a mass-10 board at FAIL #2
// must not reach back and make FAIL #1 a FAIL over a tenth of the peak.
func TestAVerdictsPeakIsThePeakAsItStoodThen(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gLow, gLow, gLow)).
		add("red-chair", failVerdict()).
		register("red-chair").
		add(evLens, gradedMint("G2", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict())
	eps, err := ConvergenceVsVerdict(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("verdict rows = %+v, want 2", eps)
	}
	if got := eps[0]; got.Peak != 1 || got.Divergent {
		t.Errorf("FAIL #1 = %+v, want peak 1 — the larger board came after it — and not divergent", got)
	}
	if got := eps[1]; got.Peak != 10 || got.Mass != 10 {
		t.Errorf("FAIL #2 = %+v, want mass and peak 10", got)
	}
}

// The fraction is the run's own: a run set up at 0.05 does not hold a mass-1 board against a peak of
// 9 converged (1 > 0.45), and neither the gate nor the scorecard may judge it at setup's default.
func TestTheConvergenceFractionIsTheRuns(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).
		register("red-chair").register(evLens).
		add(evLens, repairClose("G1")).
		add(evLens, gradedMint("G2", gLow, gLow, gLow)).
		register("red-chair").sit(2, evLens).sit(2, evLens).register("red-chair")
	run := b.seed()
	setFraction(t, run, 0.05)
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, failVerdict()); err != nil {
		t.Fatalf("FAIL over mass 1 against peak 9 at fraction 0.05 was refused: %v", err)
	}
	eps, err := ConvergenceVsVerdict(run)
	if err != nil {
		t.Fatal(err)
	}
	last := eps[len(eps)-1]
	if last.Fraction != 0.05 || last.Divergent || last.Mass != 1 || last.Peak != 9 {
		t.Errorf("the admitted FAIL = %+v, want fraction 0.05, mass 1, peak 9, not divergent", last)
	}
}

// THE GATE AND THE SCORECARD COMPUTE ONE QUANTITY. Every edge a gap can stand in relative to a
// verdict — never regraded, regraded before it, regraded after it, closed before it, closed after it,
// minted after it in its own epoch, and an act the verdict read corrected after it or before it — is
// seeded, and at each FAIL the write path's convergence over the record as it stood just before that
// FAIL is compared field by field with the scorecard's row for that FAIL, read off the whole record
// afterwards. The run's fraction is 0.5, which is what makes FAIL #4 converged (mass 21 against a
// peak of 46); at setup's default it would not be.
//
// A CORRECTION WRITTEN AFTER A VERDICT DOES NOT CHANGE THAT VERDICT'S BOARD. A corrected act keeps
// its place in every ordering ("pos"), but whether it STOOD at the verdict is a question of when it
// and its correction were written: R's regrade and S's close are corrected after FAIL #2, which read
// the originals, and K's regrade is corrected before it, so FAIL #2 read the replacement.
func TestTheGateAndTheScorecardJudgeEveryVerdictAlike(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		// Epoch 1: eight gaps, then FAIL #1 over all of them.
		register("red-chair").register(evLens).
		add(evLens, gradedMint("A", gHigh, gHigh, gHigh)). // regraded only after FAIL #3
		add(evLens, gradedMint("B", gLow, gLow, gLow)).    // regraded up before FAIL #2
		add(evLens, gradedMint("C", gMed, gMed, gMed)).    // regraded down after FAIL #2, up after FAIL #4
		add(evLens, gradedMint("D", gHigh, gHigh, gHigh)). // closed before FAIL #2
		add(evLens, gradedMint("E", gHigh, gHigh, gHigh)). // closed after FAIL #2
		add(evLens, gradedMint("R", gLow, gLow, gLow)).    // regraded before FAIL #2, the regrade corrected after it
		add(evLens, gradedMint("S", gLow, gLow, gLow)).    // closed before FAIL #2, the close corrected after it
		add(evLens, gradedMint("K", gLow, gLow, gLow)).    // regraded and the regrade corrected, both before FAIL #2
		add("red-chair", failVerdict()).
		// Epoch 2.
		register("red-chair").
		add(evLens, gradedRegrade("B", gHigh, gHigh, gHigh)).
		add(evLens, repairClose("D")).
		add(evLens, gradedRegrade("R", gLow, gHigh, gHigh))
	regradeR := b.lastKey()
	b.add(evLens, repairClose("S"))
	closeS := b.lastKey()
	b.add(evLens, gradedRegrade("K", gLow, gMed, gMed))
	regradeK := b.lastKey()
	b.add(evLens, gradedRegrade("K", gLow, gHigh, gMed))
	b.add(evLens, &recordpb.Correction{Corrects: proto.String(regradeK), Replacement: proto.String(b.lastKey()), Why: proto.String("impact misread")}).
		add("red-chair", failVerdict()).
		add(evLens, gradedRegrade("C", gLow, gLow, gLow)).
		add(evLens, repairClose("E")).
		add(evLens, gradedRegrade("R", gLow, gMed, gMed))
	b.add(evLens, &recordpb.Correction{Corrects: proto.String(regradeR), Replacement: proto.String(b.lastKey()), Why: proto.String("overstated")})
	reworded := repairClose("S")
	reworded.Prose = proto.String("fixed, as the source now reads")
	b.add(evLens, reworded)
	b.add(evLens, &recordpb.Correction{Corrects: proto.String(closeS), Replacement: proto.String(b.lastKey()), Why: proto.String("reworded")}).
		// Epoch 3: F is fresh and material before FAIL #3; G is minted after it in the same epoch.
		register("red-chair").
		add(evLens, gradedMint("F", gMed, gMed, gMed)).
		add("red-chair", failVerdict()).
		add(evLens, gradedMint("G", gHigh, gHigh, gHigh)).
		add(evLens, gradedRegrade("A", gLow, gLow, gLow)).
		// Epoch 4: nothing material stands open; G keeps its mass at a low severity.
		register("red-chair").
		add(evLens, repairClose("B")).
		add(evLens, repairClose("F")).
		add(evLens, gradedRegrade("G", gLow, gHigh, gHigh)).
		add("red-chair", failVerdict()).
		add(evLens, gradedRegrade("C", gHigh, gHigh, gHigh)).
		add(evLens, repairClose("C")).
		add(evLens, gradedMint("H", gMed, gLow, gLow)) // after FAIL #4, in its epoch: material, and not fresh to it

	// What each FAIL stood over, by hand. Masses: low 1, medium 2, high 3. R, S and K are low in
	// severity throughout, so they move the mass and never the material count.
	want := []Convergence{
		{Mass: 35, Peak: 35, MaxSeverityMass: 3, MaterialOpen: 4, FreshMaterialMints: 4, Fraction: 0.5},              // A9 B1 C4 D9 E9 R1 S1 K1
		{Mass: 46, Peak: 46, MaxSeverityMass: 3, MaterialOpen: 4, FreshMaterialMints: 0, Fraction: 0.5},              // A9 B9 C4 E9 R9 K6
		{Mass: 33, Peak: 46, MaxSeverityMass: 3, MaterialOpen: 3, FreshMaterialMints: 1, Fraction: 0.5},              // A9 B9 C1 F4 R4 K6
		{Mass: 21, Peak: 46, MaxSeverityMass: 1, MaterialOpen: 0, FreshMaterialMints: 0, Fraction: 0.5, Holds: true}, // A1 C1 G9 R4 K6
	}

	// The gate: the record as it stood just before each FAIL was written.
	var gates []Convergence
	for i, e := range b.evs {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_VERDICT {
			continue
		}
		dir := newRun(t)
		recordtest.Seed(t, dir, b.evs[:i]...)
		run := mustRun(t, dir)
		setFraction(t, run, 0.5)
		c, err := convergenceOf(run)
		if err != nil {
			t.Fatal(err)
		}
		gates = append(gates, c)
	}

	run := b.seed()
	setFraction(t, run, 0.5)
	eps, err := ConvergenceVsVerdict(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != len(want) || len(gates) != len(want) {
		t.Fatalf("verdict rows %d, gate readings %d, want %d", len(eps), len(gates), len(want))
	}
	for i := range want {
		if gates[i] != want[i] {
			t.Errorf("FAIL #%d at the gate = %+v, want %+v", i+1, gates[i], want[i])
		}
		if eps[i].Convergence != gates[i] {
			t.Errorf("FAIL #%d: the scorecard read %+v, the gate %+v", i+1, eps[i].Convergence, gates[i])
		}
		if eps[i].Divergent != gates[i].Holds {
			t.Errorf("FAIL #%d: divergent %v but the gate's convergence held = %v", i+1, eps[i].Divergent, gates[i].Holds)
		}
	}

	// At the newest event "as it stood" is "now": the board at that point is the gap view's open
	// board on its CURRENT grades and its own "material" column — the rule convergence_at restates
	// at every other point.
	now, err := convergenceOf(run)
	if err != nil {
		t.Fatal(err)
	}
	var cur Convergence
	if _, err := queryRow(run, []any{&cur.Mass, &cur.MaxSeverityMass, &cur.MaterialOpen}, `
	  SELECT COALESCE(SUM(COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0)), 0.0),
	         COALESCE(MAX(COALESCE(gs."mass", 0.0)), 0.0),
	         COALESCE(SUM(g."material"), 0)
	  FROM "gap" g
	  LEFT JOIN "enum_grade" gl ON gl."value" = g."current_likelihood"
	  LEFT JOIN "enum_grade" gi ON gi."value" = g."current_impact"
	  LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
	  WHERE g."open"`); err != nil {
		t.Fatal(err)
	}
	if now.Mass != cur.Mass || now.MaxSeverityMass != cur.MaxSeverityMass || now.MaterialOpen != cur.MaterialOpen || now.Mass != 21 || now.MaterialOpen != 1 {
		t.Errorf("at the newest event the board = %+v, the gap view's current board = %+v (want mass 21 — A1 G9 H1 R4 K6 — and H material)", now, cur)
	}
}

// The rule reads ONE term of the run, and only that term can refuse it. An epoch limit of 0 is a bad
// term for the dispatch, and says nothing about the fraction; a run with no record has nothing to
// judge, whatever its config holds. A fraction outside (0, 1] is the one config error the detector
// must report.
func TestTheDetectorIsRefusedOnlyByABadFraction(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).seed()

	writeRunConfig(t, run.Dir(), `{"maxEpochs":0,"convergenceFraction":0.5}`)
	if _, err := RunParams(run); err == nil {
		t.Fatal("RunParams accepted maxEpochs 0 — the fixture no longer holds a term the detector must ignore")
	}
	eps, err := ConvergenceVsVerdict(run)
	if err != nil || len(eps) != 1 || eps[0].Fraction != 0.5 {
		t.Errorf("ConvergenceVsVerdict beside maxEpochs 0 = (%+v, %v), want one row at fraction 0.5", eps, err)
	}
	if c, err := convergenceOf(run); err != nil || c.Fraction != 0.5 {
		t.Errorf("the gate beside maxEpochs 0 = (%+v, %v), want fraction 0.5", c, err)
	}

	writeRunConfig(t, run.Dir(), `{"convergenceFraction":0}`)
	if eps, err := ConvergenceVsVerdict(run); err == nil {
		t.Errorf("ConvergenceVsVerdict at fraction 0 = %+v, want the error", eps)
	}

	empty := mustRun(t, newRun(t))
	writeRunConfig(t, empty.Dir(), `{`)
	if eps, err := ConvergenceVsVerdict(empty); err != nil || eps != nil {
		t.Errorf("ConvergenceVsVerdict over no record with a config that does not parse = (%+v, %v), want (nil, nil)", eps, err)
	}
}

// A RECORD THAT CANNOT ANSWER IS NOT A BOARD THAT HAS NOT CONVERGED. With events on the record and no
// convergence_at row at the newest one, a zero Convergence reads exactly like a board the rule does
// not hold on, and the refusal would admit the FAIL it exists to stop.
func TestAMissingConvergenceRowIsAnError(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gLow, gLow, gLow)).seed()
	if _, err := convergenceOf(run); err != nil {
		t.Fatalf("the healthy read failed: %v", err)
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open the run's handle: (%v, %v)", db, err)
	}
	if _, err := db.Exec(`DROP VIEW "convergence_at"`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE VIEW "convergence_at" AS SELECT 0 AS "seq", 0.0 AS "mass", 0.0 AS "peak",
	  0.0 AS "max_severity_mass", 0 AS "material_open", 0 AS "fresh_material_mints" WHERE 0`); err != nil {
		t.Fatal(err)
	}
	if c, err := convergenceOf(run); err == nil {
		t.Errorf("convergenceOf with no row at the newest event = %+v, nil — want the error", c)
	}
}
