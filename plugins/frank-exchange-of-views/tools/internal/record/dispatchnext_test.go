package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// atImpasse is a run whose chair has just registered with G1 at impasse: `dispatch next` dockets G1
// and convenes the bench on it.
func atImpasse(t *testing.T) Run {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high")
	for i := 0; i < 3; i++ {
		b.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register(evLens).register("blue-respond")
	}
	return b.register("red-chair").seed()
}

// askedElsewhere is what another asking of `dispatch next` on run writes for plan: the docket motion
// for each gap it files at impasse, and its rows.
func askedElsewhere(t *testing.T, plan Plan) []*Event {
	var acts []*Event
	for _, g := range plan.ToFile {
		acts = append(acts, recordtest.At(t, "red-chair", "race:motion:"+g, docketMotion("M-race-"+g, g)))
	}
	for _, p := range plan.Parties {
		var occ []recordpb.Occasion
		for _, w := range p.Occasions {
			o, _ := OccasionOf(w)
			occ = append(occ, o)
		}
		acts = append(acts, recordtest.At(t, "red-chair", "race:dispatch:"+p.SeatID, &recordpb.Dispatch{
			Pin: proto.Int64(plan.Head), SeatId: proto.String(p.SeatID), GapIds: p.GapIDs, Occasions: occ}))
	}
	return acts
}

// recorded counts the docket motions on gap and the dispatch rows on run's record.
func recorded(t *testing.T, run Run, gap string) (motions, rows int) {
	t.Helper()
	evs, _, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if m := e.GetMotion(); m != nil && m.GetDocket().GetGapId() == gap {
			motions++
		}
		if e.GetDispatch() != nil {
			rows++
		}
	}
	return motions, rows
}

// TWO ASKINGS OF `dispatch next` AT ONCE RECORD ONE DECISION. The chair asks for the prose and the
// --json in one turn; each asking that finds no standing plan writes the rows, and at impasse files
// the docket motion. Here the other asking's motion and rows commit, on their own connection, as this
// one opens its transaction: it must find them standing and write nothing, as it does when they are
// already on the record — one motion on G1, one set of rows.
func TestTwoAskingsOfDispatchNextRecordOneDecision(t *testing.T) {
	chair := func(run Run) Identity { return Identity{Run: run, SeatID: "red-chair"} }
	before := atImpasse(t)
	_, rowsBefore := recorded(t, before, "G1")
	plan, err := PlanDispatch(before)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.ToFile) != 1 || len(plan.Parties) == 0 {
		t.Fatalf("fixture: the plan files %v and convenes %d parties, want G1 filed and the bench convened", plan.ToFile, len(plan.Parties))
	}
	if _, standing, err := DispatchNext(chair(before)); err != nil || standing {
		t.Fatalf("fixture: asked once = standing %v, %v", standing, err)
	}
	if m, r := recorded(t, before, "G1"); m != 1 || r != rowsBefore+len(plan.Parties) {
		t.Fatalf("fixture: asked once, %d docket motions on G1 and %d new rows, want 1 and %d", m, r-rowsBefore, len(plan.Parties))
	}

	after := atImpasse(t)
	recordtest.Seed(t, after.Dir(), askedElsewhere(t, plan)...)
	if _, standing, err := DispatchNext(chair(after)); err != nil || !standing {
		t.Fatalf("fixture: asked after the other asking = standing %v, %v — want the plan standing", standing, err)
	}

	raced := atImpasse(t)
	var standing bool
	if err := overtaken(t, raced, askedElsewhere(t, plan), func() error {
		var err error
		_, standing, err = DispatchNext(chair(raced))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if m, r := recorded(t, raced, "G1"); m != 1 || r != rowsBefore+len(plan.Parties) || !standing {
		t.Fatalf("asked while the other asking committed: %d docket motions on G1 and %d new rows (standing %v), want 1 and %d — the plan was folded on a record its rows did not land on",
			m, r-rowsBefore, standing, len(plan.Parties))
	}
}

// THE ROWS `dispatch next` WRITES ARE THE PLAN OF THE RECORD THEY LAND ON. The lens closes G1 as the
// chair asks: the rows must be the plan the record gives with the close on it, so the plan asked of
// the record afterwards stands as the recorded dispatch. Folded before the close and written after
// it, the rows engage a party on a gap that is closed — a decision no state of the record gives. Asked
// of the record afterwards (no further act between), the plan names exactly the rows this asking
// wrote.
func TestDispatchNextRecordsThePlanOfTheRecordItLandsOn(t *testing.T) {
	stage := func() Run {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high").
			register("red-chair").seed()
	}
	closes := func() []*Event { return []*Event{recordtest.At(t, evLens, "race:close:G1", repairClose("G1"))} }
	engaged := func(p Plan) bool {
		for _, party := range p.Parties {
			for _, g := range party.GapIDs {
				if g == "G1" {
					return true
				}
			}
		}
		return false
	}
	before := stage()
	if plan, _, err := DispatchNext(Identity{Run: before, SeatID: "red-chair"}); err != nil || !engaged(plan) {
		t.Fatalf("fixture: with G1 open the plan engages nobody on it (%v)", err)
	}
	after := stage()
	recordtest.Seed(t, after.Dir(), closes()...)
	if plan, _, err := DispatchNext(Identity{Run: after, SeatID: "red-chair"}); err != nil || engaged(plan) {
		t.Fatalf("fixture: with G1 closed the plan still engages a party on it (%v)", err)
	}

	raced := stage()
	_, rowsBefore := recorded(t, raced, "G1")
	if err := overtaken(t, raced, closes(), func() error {
		_, _, err := DispatchNext(Identity{Run: raced, SeatID: "red-chair"})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	evs, _, err := eventsAt(recordOf(t, raced))
	if err != nil {
		t.Fatal(err)
	}
	var written []string
	for _, e := range evs {
		if d := e.GetDispatch(); d != nil {
			written = append(written, d.GetSeatId()+" on "+strings.Join(d.GetGapIds(), ","))
		}
	}
	written = written[rowsBefore:]
	now, err := PlanDispatch(raced)
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, p := range now.Parties {
		want = append(want, p.SeatID+" on "+strings.Join(p.GapIDs, ","))
	}
	if !slices.Equal(written, want) {
		t.Fatalf("the rows recorded while the lens closed G1 are %q, and the plan the record gives is %q — the plan was folded on a record its rows did not land on", written, want)
	}
}

// THE IMPASSE DOCKET IS FILED AT THE FINAL SITTING TOO. A seat's motion filed there is refused
// (HeardInTime), and the docket `dispatch next` files for a gap at impasse is not a seat's motion: it
// is recorded, nobody is dispatched, and the plan lists it as the bench's to rule, which is what
// convenes the terminal bench at the exit.
func TestTheFinalSittingStillDocketsAGapAtImpasse(t *testing.T) {
	run := atImpasse(t)
	writeRunConfig(t, run.dir, `{"maxEpochs":5}`)
	plan, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil {
		t.Fatalf("dispatch next at the final sitting with G1 at impasse: %v", err)
	}
	if !plan.FinalSitting || !plan.EpochLimitReached || len(plan.Parties) != 0 || !slices.Equal(plan.Docket, []string{"G1"}) {
		t.Fatalf("plan = %+v, want the final sitting, nobody dispatched, and G1 docketed", plan)
	}
	if motions, _ := recorded(t, run, "G1"); motions != 1 {
		t.Fatalf("the final sitting recorded %d docket motion(s) on G1, want 1", motions)
	}
	benchOwned := false
	for _, b := range plan.Blockers {
		benchOwned = benchOwned || (b.Kind == BlockerUnruledMotion && b.Owner == "judge")
	}
	if !benchOwned {
		t.Errorf("the plan lists no motion for the bench to rule, so no terminal sitting convenes: %+v", plan.Blockers)
	}
}
