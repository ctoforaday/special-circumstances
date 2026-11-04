package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// UNIVERSE-M38'S G-29f07d26, REPLAYED (ruled by gblock 2026-11-04: minting with --supersedes closes
// the ancestor in the same act, pointing at the successor). The lens's acceptance check on its gap
// could never pass, so it minted a successor superseding it (ev624). The run then had no close for
// "moved to the successor", and the ancestor, open, refused every verdict until the run ended CEILING.
// Here the successor's mint closes G1: G1 reads closed with the fate "superseded" and names G2; no
// blocker and no verdict refusal names it; and a regrade, a close, a docket and a carry of G1 are
// refused, each naming G2.
func TestM38ASupersededAncestorHoldsNoVerdict(t *testing.T) {
	st := mintedOn(t)
	st.add(evLens, successorMint("G2", "G1")).add(evLens, repairClose("G2"))
	run := st.register("red-chair").seed()

	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	g := fam.Gap("G1")
	if g == nil || g.Open || g.SupersededBy != "G2" || g.Closure != nil || g.ClosureReason() != ClosedSuperseded || g.Successor() != "G2" {
		t.Fatalf("G1 = %+v, want closed by G2's mint, fate %q, no closing body", g, ClosedSuperseded)
	}
	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	onBoard := false
	for _, c := range bj.Closed {
		if c.ID == "G1" {
			onBoard = c.SupersededBy == "G2"
		}
	}
	if !onBoard || bj.Counts.Open != 0 {
		t.Errorf("board: G1 closed and superseded by G2 = %v, open %d, want true and 0; anomalies %v", onBoard, bj.Counts.Open, bj.Anomalies)
	}

	for _, b := range mustBlockers(t, run) {
		if b.Subject == "G1" {
			t.Errorf("blocker %s holds on G1, which G2's mint closed", b.Kind)
		}
	}
	for _, v := range []recordpb.Verdict{recordpb.Verdict_VERDICT_PASS, recordpb.Verdict_VERDICT_FAIL} {
		err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, &recordpb.Gate{Verdict: recordtest.P(v)})
		if err != nil && strings.Contains(err.Error(), "G1") {
			t.Errorf("%s refused naming G1: %v", v, err)
		}
	}

	names := func(what string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "G1, which closed when G2 was minted superseding it") || !strings.Contains(err.Error(), "Act on G2") {
			t.Errorf("%s of the superseded G1 = %v, want it refused naming its successor G2", what, err)
		}
	}
	names("a regrade", validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_REGRADE,
		regradeOf("G1", recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY, recordpb.Grade_GRADE_LOW)))
	names("a close", validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_CLOSE, repairClose("G1")))
	names("a docket", validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_MOTION, docketMotion("M1", "G1")))
	names("a carry", validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{GapId: proto.String("G1"),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), CarriedFrom: proto.String("1"), Prose: proto.String("as before")}))
}

// ONLY A GAP'S MINTER SUPERSEDES IT WHILE IT IS OPEN, AND NOT WHILE THE BENCH HAS IT. A mint
// superseding an open gap closes it, so another lens's is refused naming the minter, and the minter's
// own is refused while a docket motion stands unruled on the gap. Once the gap is closed, a mint
// naming it is lineage and closes nothing, whoever writes it.
func TestASupersessionIsTheMintersAndWaitsOnTheBench(t *testing.T) {
	const logic = "red-lens-logic"
	run := newStage(t).cast(evLens, logic, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).register(logic).mint(evLens, "G1", "high").seed()
	err := validateRun(run, logic, recordpb.EventType_EVENT_TYPE_MINT, successorMint("G2", "G1"))
	if err == nil || !strings.Contains(err.Error(), "--supersedes names G1, which is open and was minted by "+evLens) {
		t.Errorf("another lens superseding the open G1 = %v, want it refused naming its minter", err)
	}
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_MINT, successorMint("G2", "G1")); err != nil {
		t.Errorf("the minter superseding its own open G1 = %v, want it admitted", err)
	}

	docketed := mintedOn(t).add("red-chair", docketMotion("M1", "G1")).register("red-chair").seed()
	err = validateRun(docketed, evLens, recordpb.EventType_EVENT_TYPE_MINT, successorMint("G2", "G1"))
	if err == nil || !strings.Contains(err.Error(), "--supersedes names G1, which is docketed for the bench: motion M1 stands unruled on it") {
		t.Errorf("the minter superseding its docketed G1 = %v, want it refused naming the motion", err)
	}

	closed := mintedOn(t).add(evLens, repairClose("G1")).register("red-chair").seed()
	if err := validateRun(closed, logic, recordpb.EventType_EVENT_TYPE_MINT, successorMint("G2", "G1")); err != nil && strings.Contains(err.Error(), "--supersedes") {
		t.Errorf("another lens naming the closed G1 as an ancestor = %v, want lineage admitted", err)
	}
}
