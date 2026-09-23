package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

const voiceLens = "red-lens-voice"

// logEntry files a real entry. The type that asserted a clean sitting is retired — clean is
// derived from a bracketed sitting that filed nothing — so a fixture that wants an ENTRY on the
// record files one that asserts something.
func (b *stage) logEntry(seat string) *stage {
	return b.add(seat, &recordpb.Log{Text: proto.String("no verb renders a gap's lineage in one read"),
		Type: recordpb.LogType_LOG_TYPE_REQUEST.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()})
}

// owedItems is the blocking items that name the owed sitting.
func owedItems(s SittingJSON) []Item {
	var out []Item
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, "register for this sitting") {
			out = append(out, it)
		}
	}
	return out
}

func planNames(p Plan, seat string) bool {
	for _, x := range p.Parties {
		if x.SeatID == seat {
			return true
		}
	}
	return false
}

// b5Shape is the 2026-09-11 B5 record in miniature: the voice lens registered once, the report
// head moved, and every later dispatch at the new head found it logging "nothing to do" and
// leaving without registering — so it never sat, its state never moved, and dispatch readied it again.
func b5Shape(t *testing.T) *stage {
	return newStage(t).cast(voiceLens, "red-chair", "blue-respond", "judge").ingest(). // head 2
												register("red-chair").dispatch(2, voiceLens).register(voiceLens).logEntry(voiceLens). // it sat once
												edit("", "a", "b").                                                                   // head 7
												register("red-chair").dispatch(7, voiceLens).logEntry(voiceLens).
												register("red-chair").dispatch(7, voiceLens).logEntry(voiceLens) // and again, unregistered
}

// THE WORK LIST TELLS THE TRUTH THE DISPATCH ACTS ON. A lens dispatched and not registered since
// has not sat: dispatch readies it again, and its own work list says the sitting is owed and that
// the sitting may not close — where it used to say `complete: true` to a seat dispatch was about to
// re-ready, which is the loop B5 ran eleven times at one head.
func TestADispatchedLensThatHasNotRegisteredOwesItsSitting(t *testing.T) {
	run := b5Shape(t).seed()
	s := sittingOfRunT(t, run, "lens", voiceLens)
	if s.Complete {
		t.Fatalf("complete = true for a lens dispatched at head 7 that never registered; open = %+v", s.Open)
	}
	owed := owedItems(s)
	if len(owed) != 1 || !strings.Contains(owed[0].What, "head 7") {
		t.Fatalf("owed items = %+v, want one blocking item naming head 7 and the register", owed)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !planNames(plan, voiceLens) {
		t.Fatalf("the work list says the sitting is owed but dispatch does not ready the lens: %+v", plan)
	}

	// It registers and does nothing else: the sitting is on the record, the owed-sitting item goes,
	// and NOTHING is owed. A sitting with no acts does not owe a log — the emptiness is the clean
	// case, derived rather than asserted (#1089), and this is the state a woken lens with no work
	// should be able to end its turn in.
	registered := b5Shape(t).register(voiceLens).seed()
	s = sittingOfRunT(t, registered, "lens", voiceLens)
	if len(owedItems(s)) != 0 || !s.Complete {
		t.Fatalf("after its register an empty sitting still owes: complete=%v open=%+v", s.Complete, s.Open)
	}
	// But a sitting that ACTED owes its log, and its first sitting's entry does not discharge it.
	acted := b5Shape(t).register(voiceLens).
		add(voiceLens, &recordpb.Finding{Label: proto.String("voice-F9"), Text: proto.String("an act this sitting made")}).seed()
	s = sittingOfRunT(t, acted, "lens", voiceLens)
	if !hasItem(s, "the log is open") {
		t.Fatalf("a sitting that acted and filed no log is not told: complete=%v open=%+v", s.Complete, s.Open)
	}
	// It logs: the work list is complete, and dispatch agrees.
	sat := b5Shape(t).register(voiceLens).logEntry(voiceLens).seed()
	s = sittingOfRunT(t, sat, "lens", voiceLens)
	if !s.Complete || len(owedItems(s)) != 0 {
		t.Fatalf("after its register and log the lens still owes: complete=%v open=%+v", s.Complete, s.Open)
	}
	plan, err = PlanDispatch(sat)
	if err != nil {
		t.Fatal(err)
	}
	if planNames(plan, voiceLens) {
		t.Fatalf("the lens registered for head 7 and dispatch still readies it: %+v", plan)
	}
}

// Every seat a dispatch names is readied again by the same predicate — blue and the minting lens
// get no exchange counted, the bench has not sat for its docketing — so each owes the sitting on
// its own list, and a seat no dispatch names owes none.
func TestEveryDispatchedSeatOwesTheSittingItWasDispatchedFor(t *testing.T) {
	base := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high").
			register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").dispatch(2, "judge", "G1")
	}
	for _, c := range []struct{ role, seat string }{{"lens", evLens}, {"blue", "blue-respond"}, {"bench", "judge"}} {
		if n := len(owedItems(sittingOfRunT(t, base().seed(), c.role, c.seat))); n != 1 {
			t.Errorf("%s dispatched and unregistered: %d owed item(s), want 1", c.seat, n)
		}
		if n := len(owedItems(sittingOfRunT(t, base().register(c.seat).seed(), c.role, c.seat))); n != 0 {
			t.Errorf("%s registered after its dispatch: %d owed item(s), want 0", c.seat, n)
		}
	}
	if n := len(owedItems(sittingOfRunT(t, base().seed(), "chair", "red-chair"))); n != 0 {
		t.Errorf("the chair is never dispatched and owes no dispatched sitting; got %d", n)
	}
}
