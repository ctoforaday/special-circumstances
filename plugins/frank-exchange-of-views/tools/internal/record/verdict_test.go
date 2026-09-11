package record

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

func runWith(t *testing.T, maxRounds string, evs []*Event) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	if maxRounds != "" {
		if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"),
			[]byte(`{"maxRounds":"`+maxRounds+`"}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, evs...)
	return dir
}

// vev builds a fixture act. The key is composed from what the event actually SAYS — the type is
// derived from the body — rather than from a word passed alongside it that could disagree.
func vev(t *testing.T, seat string, round int, body proto.Message) *Event {
	t.Helper()
	ev := recordtest.Event(t, seat, body)
	// One chair, several epochs: the key carries the round so two verdicts by red-chair are two rows.
	ev.Key = proto.String(seat + ":" + recordpb.Word(ev.GetType()) + ":" + strconv.Itoa(round))
	return ev
}

// A PASS on the record is VERIFIED, without anyone saying so.
func TestVerifiedIsDerivedFromThePassEvent(t *testing.T) {
	dir := runWith(t, "3", []*Event{vev(t, "red-chair", 1, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})})
	got, why, ok := DeriveVerdict(mustRun(t, dir))
	if !ok || got != "VERIFIED" {
		t.Errorf("got %q (ok=%v) — want VERIFIED: %s", got, ok, why)
	}
}

// Reaching the ceiling with no pass is CEILING — computed from the rounds on the record
// against the bound setup wrote, so nobody has to be told.
// CEILING is derived from the board: every open material gap at impasse, docketed and ruled
// carried, with nobody ready — no clock is consulted.
func TestCeilingIsDerivedFromEveryMaterialGapAtItsLimit(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high")
	for i := 0; i < 2; i++ {
		st.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register(evLens).register("blue-respond")
	}
	// At impasse: the chair's verb dockets G1 and the bench sits and rules it carried.
	st.register("red-chair").
		add("red-chair", &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
			Basis: proto.String("at impasse"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}).
		dispatch(2, "judge", "G1").register("judge").
		add("judge", &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
			Opinion: proto.String("the parties have said what they can"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED), Principle: proto.String("p"), Tension: proto.String("t"),
				ReviewFlag: proto.String("none"), Settled: proto.String("nothing"), Final: proto.Bool(true)}}}).
		register("red-chair")
	run := st.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parties) != 0 || !plan.Ceiling || plan.PassPermitted {
		t.Fatalf("plan = %+v, want nobody ready and ceiling", plan)
	}
	got, why, ok := DeriveVerdict(run)
	if !ok || got != "CEILING" {
		t.Errorf("got %q (ok=%v) — want CEILING: %s", got, ok, why)
	}
}

// A HALT OUTRANKS A PASS. A run stopped on safety or integrity grounds did not end by
// passing, however clean the board looked when it stopped.
func TestHaltOutranksAPass(t *testing.T) {
	dir := runWith(t, "3", []*Event{
		vev(t, "red-chair", 1, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}),
		vev(t, "judge", 1, &recordpb.Halt{Opinion: proto.String("consent gate")}),
	})
	got, _, ok := DeriveVerdict(mustRun(t, dir))
	if !ok || got != "HALTED" {
		t.Errorf("got %q (ok=%v) — a halt must outrank a pass", got, ok)
	}
}

// THE ONE CASE THE RECORD CANNOT DECIDE, and it must say so rather than guess. A run that
// ends early with no pass, no halt and no board at its ceiling ended UNVERIFIED — and the
// record cannot tell that from a run still in flight, so it refuses to derive rather than guess.
func TestARunThatEndedEarlyIsNotDerivable(t *testing.T) {
	dir := runWith(t, "5", []*Event{vev(t, "red-chair", 1, &recordpb.Position{Text: proto.String("x")})})
	got, why, ok := DeriveVerdict(mustRun(t, dir))
	if ok {
		t.Errorf("derived %q from a record that cannot decide — the ended-early case must stay honest", got)
	}
	if why == "" {
		t.Error("the refusal must explain WHY it cannot decide, or the gap is invisible")
	}
}

// An absent or unparseable ceiling degrades CEILING to underivable rather than inventing a
// bound — the same posture as InferRunDir's "say nothing rather than guess".
