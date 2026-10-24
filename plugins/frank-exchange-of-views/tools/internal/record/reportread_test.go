package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AT THE FINAL HEAD, A LENS THAT WAS NEVER SERVED THE REPORT HOLDS THE PASS (#1458). It sat at the
// head — its sitting opened and closed — and recorded nothing: that is the same record whether it
// read the report and found nothing or never looked. The dispatch's pin moved; its own did not, so it
// is ready at the head and the gate refuses the PASS naming it. Served the report there, it is not.
func TestALensNeverServedTheReportAtTheFinalHeadHoldsThePass(t *testing.T) {
	final := func() *stage { b := retBoard(t).sit(2, evLens).ingest(); return b }
	unserved := final()
	unserved.sitUnserved(unserved.head(), evLens).register("red-chair")
	run := unserved.seed()
	if plan, err := PlanDispatch(run); err != nil || plan.PassPermitted || !lensReady(plan) {
		t.Fatalf("plan over a lens never served the final head = %+v; want it ready and no PASS", plan)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err == nil ||
		!strings.Contains(err.Error(), evLens+": the whole report was last served to it at its pin 2, behind head ") {
		t.Fatalf("a PASS over a lens never served the final head: %v; want it refused naming the lens and its pin", err)
	}

	served := final()
	served.sit(served.head(), evLens).register("red-chair")
	if _, err := Append(Identity{Run: served.seed(), SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS once the lens was served the final head was refused: %v", err)
	}
}

// THE FALLBACK IS THE CHAIR'S SPOT-CHECK, AS IT WAS. A lens that keeps sitting without being served
// the report counts those sittings toward retirement like any barren sitting, so the readying is
// bounded: retired, its pin is the head it was LAST SERVED, not the dispatch's, so its area is stale —
// the chair reads the changes since that pin and names the area in its spot-check, and the PASS lands.
func TestALensRetiredWithoutBeingServedTheHeadLeavesItsAreaToTheSpotCheck(t *testing.T) {
	board := func() (*stage, int64) {
		b := retBoard(t).sit(2, evLens).ingest()
		h := b.head()
		b.sitUnserved(h, evLens).sitUnserved(h, evLens)
		return b, h
	}
	b, h := board()
	f := lensStateOf(t, b.seed())
	if !f.retired() || f.pin != 2 || f.ready {
		t.Fatalf("three barren sittings, the last two never served head %d: %+v — want retired at pin 2", h, f)
	}
	b, _ = board()
	run := b.register("red-chair").seed()
	if plan, err := PlanDispatch(run); err != nil || len(plan.StaleAreas) != 1 || plan.StaleAreas[0] != (StaleArea{SeatID: evLens, Pin: 2}) {
		t.Fatalf("stale areas = %+v (%v), want %s at pin 2", plan.StaleAreas, err, evLens)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err == nil || !strings.Contains(err.Error(), evLens+" (pin 2)") {
		t.Fatalf("a PASS over the area of a lens retired behind the head without its read: %v; want the stale-area refusal", err)
	}
	b, _ = board()
	b.register("red-chair").add("red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the changes since pin 2")})
	if _, err := Append(Identity{Run: b.seed(), SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS after the spot-check named the area was refused: %v", err)
	}
}

// A READ IN A SITTING THE LENS DID NOT SIT THROUGH IS NO PIN: stopped at the call limit, the sitting
// is not one it sat (satThrough), and nothing it was served there is its audit.
func TestAReadInASittingStoppedAtTheLimitMovesNoPin(t *testing.T) {
	b := retBoard(t).sit(2, evLens).ingest()
	h := b.head()
	b.sitHooked(h, evLens, "a1", false)
	b.limit(evLens, 2).stop("a1").register("red-chair")
	if f := lensStateOf(t, b.seed()); f.pin != 2 || !f.ready {
		t.Fatalf("a read in a sitting stopped at the limit moved the pin: %+v", f)
	}
}

// THE REPORT READ IS NOT AN ACT. It is what the tool served, as a stale refusal is what it showed:
// a sitting holding nothing else recorded nothing, so the log is not asked of it; it is no movement
// on a gap, so an exchange it falls in is a null turn as before; and it opens no sitting.
func TestAReportReadIsNotAnAct(t *testing.T) {
	quiet := retBoard(t).sitUnserved(2, evLens)
	read := retBoard(t).sit(2, evLens)
	for name, b := range map[string]*stage{"registered": quiet, "registered and served": read} {
		run := b.seed()
		m, err := MergedEvents(run)
		if err != nil {
			t.Fatal(err)
		}
		if !sittingRecordedNothing(m.Events, m.At, evLens) {
			t.Errorf("%s: a sitting holding only a report read reads as having acted", name)
		}
		if s := sittingOfRunT(t, run, "lens", evLens); listsItem(s, "the log is open") {
			t.Errorf("%s: the log is asked of a sitting that recorded no act: %+v", name, s.Open)
		}
	}

	exchange := func(served bool) *GapExchanges {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(evLens).mint(evLens, "G1", "high")
		for range 3 { // each epoch: both parties engaged, the lens served the report (or not), blue silent
			b.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register(evLens)
			if served {
				b.served(evLens, 2)
			}
			b.register("blue-respond")
		}
		x, err := Exchanges(b.seed(), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		return x["G1"]
	}
	if without, with := exchange(false), exchange(true); without == nil || with == nil || with.Exchanges == 0 ||
		without.Exchanges != with.Exchanges || without.Stalled != with.Stalled || without.Unresolved != with.Unresolved || with.Stalled != with.Exchanges {
		t.Errorf("exchanges on G1 without a read %+v, with one %+v — want the same null turn", without, with)
	}

	m, err := MergedEvents(retBoard(t).sitUnserved(2, evLens).served(evLens, 2).served(evLens, 2).seed())
	if err != nil {
		t.Fatal(err)
	}
	if _, openings := dispatchLedger(m.Events, m.At.IDs(m.Events), m.At); len(openings[evLens]) != 1 {
		t.Errorf("report reads opened sittings: %v", openings[evLens])
	}
}

// A REPORT READ IS A LENS'S, AT THE HEAD THE TEXT WAS SERVED AT. The write refuses one by any other
// seat — no other seat's audit is pinned — and one naming a head the report is not at.
func TestAReportReadIsALensesAtTheReportsHead(t *testing.T) {
	run := retBoard(t).seed()
	if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_REPORT_READ, &recordpb.ReportRead{Head: proto.Int64(2)}); err == nil || !strings.Contains(err.Error(), "is not a lens") {
		t.Errorf("a chair's report read: %v; want refused", err)
	}
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_REPORT_READ, &recordpb.ReportRead{Head: proto.Int64(1)}); err == nil || !strings.Contains(err.Error(), "the report's head is 2") {
		t.Errorf("a read naming head 1 over head 2: %v; want refused", err)
	}
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_REPORT_READ, &recordpb.ReportRead{Head: proto.Int64(2)}); err != nil {
		t.Errorf("a lens's read at the head: %v", err)
	}
}
