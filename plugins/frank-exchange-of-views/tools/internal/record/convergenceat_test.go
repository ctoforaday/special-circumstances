package record

import (
	"strconv"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A FAIL is judged on the grades AS THEY STAND WHEN IT IS WRITTEN: every gap open at the record's
// newest event, each graded by its latest standing regrade (else its mint), against the run's own
// fraction.

func gradedMint(gap string, sev, lik, imp recordpb.Grade) *recordpb.Mint {
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
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
// the write path admits it.
func TestARegradeBeforeAFailIsTheGradeTheGateReads(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).
		add(evLens, repairClose("G1")).
		add(evLens, gradedMint("G2", gLow, gLow, gLow)).
		add(evLens, &recordpb.Regrade{GapId: proto.String("G2"), Likelihood: recordtest.P(gHigh),
			Impact: recordtest.P(gHigh), Basis: proto.String("worse than it looked")}).
		register("red-chair")
	run := b.seed()
	gate, err := convergenceOf(viewOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	if gate.Holds || gate.Mass != 9 || gate.Peak != 9 {
		t.Fatalf("the gate's convergence before FAIL #2 = %+v, want mass 9 on the regraded G2, peak 9, not holding", gate)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, failVerdict()); err != nil {
		t.Fatalf("FAIL #2 over a board of mass 9 against a peak of 9 was refused: %v", err)
	}
}

// The fraction is the run's own: a run set up at 0.05 does not hold a mass-1 board against a peak of
// 9 converged (1 > 0.45), and the gate may not judge it at setup's default.
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
	c, err := convergenceOf(viewOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	if c.Fraction != 0.05 || c.Holds || c.Mass != 1 || c.Peak != 9 {
		t.Errorf("the gate's convergence = %+v, want fraction 0.05, mass 1, peak 9, not holding", c)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, failVerdict()); err != nil {
		t.Fatalf("FAIL over mass 1 against peak 9 at fraction 0.05 was refused: %v", err)
	}
}

// THE GATE READS THE BOARD AS IT STANDS. A gap is seeded in every state it can hold when a FAIL is
// written — never regraded, regraded, closed, minted fresh in the FAIL's epoch, and regraded twice —
// and at each FAIL the write path's convergence over the record just before it
// is compared field by field with the board worked out by hand. The run's fraction is 0.5, which is
// what makes FAIL #4 converged (mass 21 against a peak of 46); at setup's default it would not be.
//
// THE LATER ACT IS READ: K is regraded twice before FAIL #2, so FAIL #2 reads the second regrade.
func TestTheGateJudgesTheBoardAsItStandsAtEachFail(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		// Epoch 1: eight gaps, then FAIL #1 over all of them.
		register("red-chair").register(evLens).
		add(evLens, gradedMint("A", gHigh, gHigh, gHigh)). // regraded only after FAIL #3
		add(evLens, gradedMint("B", gLow, gLow, gLow)).    // regraded up before FAIL #2
		add(evLens, gradedMint("C", gMed, gMed, gMed)).    // regraded down after FAIL #2, up after FAIL #4
		add(evLens, gradedMint("D", gHigh, gHigh, gHigh)). // closed before FAIL #2
		add(evLens, gradedMint("E", gHigh, gHigh, gHigh)). // closed after FAIL #2
		add(evLens, gradedMint("R", gLow, gLow, gLow)).    // regraded before FAIL #2, regraded again after it
		add(evLens, gradedMint("S", gLow, gLow, gLow)).    // closed before FAIL #2, closed again after it
		add(evLens, gradedMint("K", gLow, gLow, gLow)).    // regraded twice, both before FAIL #2
		add("red-chair", failVerdict()).
		// Epoch 2.
		register("red-chair").
		add(evLens, gradedRegrade("B", gHigh, gHigh, gHigh)).
		add(evLens, repairClose("D")).
		add(evLens, gradedRegrade("R", gLow, gHigh, gHigh))
	b.add(evLens, repairClose("S"))
	b.add(evLens, gradedRegrade("K", gLow, gMed, gMed))
	b.add(evLens, gradedRegrade("K", gLow, gHigh, gMed)).
		add("red-chair", failVerdict()).
		add(evLens, gradedRegrade("C", gLow, gLow, gLow)).
		add(evLens, repairClose("E")).
		add(evLens, gradedRegrade("R", gLow, gMed, gMed))
	reworded := repairClose("S")
	reworded.Prose = proto.String("fixed, as the source now reads")
	b.add(evLens, reworded).
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
		c, err := convergenceOf(viewOf(t, run))
		if err != nil {
			t.Fatal(err)
		}
		gates = append(gates, c)
	}

	if len(gates) != len(want) {
		t.Fatalf("gate readings %d, want %d", len(gates), len(want))
	}
	for i := range want {
		if gates[i] != want[i] {
			t.Errorf("FAIL #%d at the gate = %+v, want %+v", i+1, gates[i], want[i])
		}
	}

	// At the newest event "as it stood" is "now": the board at that point is the gap view's open
	// board on its CURRENT grades and its own "material" column — the rule convergence_at restates.
	run := b.seed()
	setFraction(t, run, 0.5)
	now, err := convergenceOf(viewOf(t, run))
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
// judge, whatever its config holds. A fraction outside (0, 1] is the one config error the gate must
// report.
func TestTheGateIsRefusedOnlyByABadFraction(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gHigh, gHigh, gHigh)).
		add("red-chair", failVerdict()).seed()

	writeRunConfig(t, run.Dir(), `{"maxEpochs":0,"convergenceFraction":0.5}`)
	if _, err := RunParams(run); err == nil {
		t.Fatal("RunParams accepted maxEpochs 0 — the fixture no longer holds a term the gate must ignore")
	}
	if c, err := convergenceOf(viewOf(t, run)); err != nil || c.Fraction != 0.5 {
		t.Errorf("the gate beside maxEpochs 0 = (%+v, %v), want fraction 0.5", c, err)
	}

	writeRunConfig(t, run.Dir(), `{"convergenceFraction":0}`)
	if c, err := convergenceOf(viewOf(t, run)); err == nil {
		t.Errorf("the gate at fraction 0 = %+v, want the error", c)
	}

	empty := mustRun(t, newRun(t))
	writeRunConfig(t, empty.Dir(), `{`)
	if c, err := convergenceOf(viewOf(t, empty)); err != nil || c.Holds {
		t.Errorf("the gate over no record with a config that does not parse = (%+v, %v), want no convergence and no error", c, err)
	}
}

// A RECORD THAT CANNOT ANSWER IS NOT A BOARD THAT HAS NOT CONVERGED. With events on the record and no
// convergence_at row at the newest one, a zero Convergence reads exactly like a board the rule does
// not hold on, and the refusal would admit the FAIL it exists to stop.
func TestAMissingConvergenceRowIsAnError(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		add(evLens, gradedMint("G1", gLow, gLow, gLow)).seed()
	if _, err := convergenceOf(viewOf(t, run)); err != nil {
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
	if c, err := convergenceOf(viewOf(t, run)); err == nil {
		t.Errorf("convergenceOf with no row at the newest event = %+v, nil — want the error", c)
	}
}
