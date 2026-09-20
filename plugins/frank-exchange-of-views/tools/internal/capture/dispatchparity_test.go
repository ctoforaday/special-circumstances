package capture

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// The relay is audited: every register between two dispatches is a named party or the chair, and
// every named party registered. A stray seat and an absent party each FAIL, by name.
func TestDispatchedPartiesAndRegistersAgree(t *testing.T) {
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	dispatch := func(seat string, gaps ...string) *record.Event {
		return at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: gaps})
	}
	head := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		reg("frontier"), // the base phase: before the first dispatch, outside the window
		reg("red-chair"), dispatch("red-lens-evidence"), dispatch("red-lens-logic"),
		reg("red-lens-evidence"), reg("red-lens-logic"),
		reg("red-chair"), dispatch("blue-respond", "G1"),
	}
	clean := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-chair"), reg("judge"), reg("judge"))
	dir := t.TempDir()
	recordtest.Seed(t, dir, clean...)
	if a := DispatchParityAudit(runtest.Open(t, dir), nil, false); a.Verdict != "PASS" {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}

	n = 100
	stray := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-lens-voice"), reg("red-chair"))
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, stray...)
	if a := DispatchParityAudit(runtest.Open(t, dir2), nil, false); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice registered after dispatch 2") {
		t.Fatalf("a seat nobody dispatched sat unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	n = 200
	absent := append(append([]*record.Event{}, head...), reg("red-chair")) // blue never sat for dispatch 2
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, absent...)
	if a := DispatchParityAudit(runtest.Open(t, dir3), nil, false); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "blue-respond was named in dispatch 2 and never registered") {
		t.Fatalf("an absent party went unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	if a := DispatchParityAudit(runtest.Open(t, func() string { d := t.TempDir(); recordtest.Seed(t, d, reg("red-chair")); return d }()), nil, false); a.Verdict != "SKIP" {
		t.Errorf("a record with no dispatch must SKIP, not judge: %s", a.Verdict)
	}
}

// THE WARM RUN, as the B6 record holds it: the chair registered ONCE and sat four times, and ran
// `dispatch next` twice in its first sitting (prose, then --json), writing the plan twice with
// nobody sitting between. Every party sat for every dispatch. Grouped by the clock's chair-sitting
// count the whole run was "dispatch 1" and every party of it read as never registered; grouped by
// who sat, it is four dispatches and a faithful relay.
func TestAWarmChairsDispatchesAreGroupedByWhoSat(t *testing.T) {
	n := 300
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	dispatch := func(seat string, gaps ...string) *record.Event {
		return at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: gaps})
	}
	lenses := []string{"red-lens-evidence", "red-lens-logic", "red-lens-voice"}
	warm := func(voiceSits, voiceLate bool) []*record.Event {
		evs := []*record.Event{
			at("harness", &recordpb.Cast{SeatIds: append([]string{"red-chair", "blue-respond", "judge"}, lenses...)}),
			at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
			reg("red-chair"), // the chair's only register, all run
		}
		for i := 0; i < 2; i++ { // the plan asked for twice, in one sitting
			for _, l := range lenses {
				evs = append(evs, dispatch(l))
			}
		}
		evs = append(evs, reg("red-lens-logic"), reg("red-lens-evidence"), reg("red-lens-voice"))
		// Sitting 2, warm: no chair register.
		evs = append(evs, dispatch("red-lens-logic", "G1"), dispatch("blue-respond", "G1", "G3"),
			dispatch("red-lens-evidence", "G3"), dispatch("red-lens-voice", "G5"))
		evs = append(evs, reg("red-lens-evidence"), reg("red-lens-logic"))
		if voiceSits {
			evs = append(evs, reg("red-lens-voice"))
		}
		evs = append(evs, reg("blue-respond"))
		// Sitting 3: the bench.
		evs = append(evs, dispatch("judge", "G2"), reg("judge"))
		if voiceLate {
			evs = append(evs, reg("red-lens-voice")) // it sat, but for nothing it was dispatched to since
		}
		return append(evs, reg("judge"))
	}

	dir := t.TempDir()
	recordtest.Seed(t, dir, warm(true, false)...)
	// THE BENCH IS DISPATCHED IN THE FINAL GROUP HERE, and since it collapsed to one seat its
	// closing sittings register in that same window — so "the bench sat" cannot be established from
	// the record, and the audit says NOT MEASURED rather than PASS. Asserting PASS would be
	// asserting a fact the events do not carry: the bookend register and the ruling register are
	// the same seat, and nothing on the record says which question either answered.
	if a := DispatchParityAudit(runtest.Open(t, dir), nil, false); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "NOT MEASURED") {
		t.Fatalf("a warm run whose bench sitting is unmeasurable = %s: %s — want the NOT MEASURED note", a.Verdict, a.Detail)
	}

	n = 400
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, warm(false, false)...)
	if a := DispatchParityAudit(runtest.Open(t, dir2), nil, false); a.Verdict != "FAIL" || !strings.HasPrefix(a.Detail, "red-lens-voice was named in dispatch 2 and never registered before the next") {
		t.Fatalf("a warm run whose voice lens never sat for dispatch 2 = %s: %s", a.Verdict, a.Detail)
	}

	// Registering AFTER the next dispatch is not sitting for this one: the voice lens's register
	// lands in dispatch 3's window, where it is a stray, and dispatch 2 still went unanswered.
	n = 500
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, warm(false, true)...)
	a := DispatchParityAudit(runtest.Open(t, dir3), nil, false)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice was named in dispatch 2 and never registered before the next") ||
		!strings.Contains(a.Detail, "red-lens-voice registered after dispatch 3 and was not a party to it") {
		t.Fatalf("a party that sat only after the next dispatch = %s: %s", a.Verdict, a.Detail)
	}
}

// THE BENCH'S OWN SITTING IS MEASURABLE NOW, AND THIS IS THE TEST THAT SAYS SO.
//
// The case: the chair dispatches the bench onto a gap, and the bench never rules it — it only sits
// for its two closing sittings, the terminal disposition and the assembly, which the ENGINE convenes
// and which register in the same window. Before the register carried an occasion, those bookends
// satisfied the dispatch: `Sat` is the party's first register after it, the bench is one seat id, and
// nothing separated a docket ruling from a bookend. The audit reported that case NOT MEASURED rather
// than passing it — an honest refusal, and a permanent blind spot on any run whose last dispatching
// chair sitting engaged the bench.
//
// This is the acceptance check for the field. A dispatched bench that never rules must FAIL, and its
// own closing sittings must not rescue it.
func TestABenchThatOnlySatForItsBookendsDidNotAnswerItsDispatch(t *testing.T) {
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	benchReg := func(occ recordpb.Occasion) *record.Event {
		return at("judge", &recordpb.Register{Occasion: occ.Enum()})
	}
	head := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		reg("red-chair"),
		at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"}}),
	}

	// The bench sits ONLY for its bookends. Both are its own, both are in the window, and neither
	// answers the chair's dispatch.
	bookends := append(append([]*record.Event{}, head...),
		benchReg(recordpb.Occasion_OCCASION_TERMINAL), benchReg(recordpb.Occasion_OCCASION_ASSEMBLE))
	dir := t.TempDir()
	recordtest.Seed(t, dir, bookends...)
	a := DispatchParityAudit(runtest.Open(t, dir), nil, false)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "judge was named in dispatch 1 and recorded no docket sitting") {
		t.Fatalf("a bench that never ruled its dispatch was satisfied by its own bookends: %s: %s", a.Verdict, a.Detail)
	}
	if strings.Contains(a.Detail, "NOT MEASURED") {
		t.Errorf("the bench's sitting is still reported unmeasurable on a record that carries occasions: %s", a.Detail)
	}

	// AND THE HONEST CASE PASSES, or the check above would be satisfied by a bench that cannot sit
	// at all. The same dispatch, with a docket sitting that answers it and the bookends after.
	n = 100
	ruled := append(append([]*record.Event{}, head...),
		benchReg(recordpb.Occasion_OCCASION_DOCKET),
		benchReg(recordpb.Occasion_OCCASION_TERMINAL), benchReg(recordpb.Occasion_OCCASION_ASSEMBLE))
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, ruled...)
	if a := DispatchParityAudit(runtest.Open(t, dir2), nil, false); a.Verdict != "PASS" {
		t.Fatalf("a bench that DID rule its dispatch was failed: %s: %s", a.Verdict, a.Detail)
	}

	// A RECORD THAT PREDATES THE FIELD STAYS NOT MEASURED — never a guess in either direction.
	// Same shape as the failing case, with no occasion on any register.
	n = 200
	old := append(append([]*record.Event{}, head...), reg("judge"), reg("judge"))
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, old...)
	if a := DispatchParityAudit(runtest.Open(t, dir3), nil, false); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "NOT MEASURED") {
		t.Fatalf("a record with no occasions must say it could not measure the bench, not answer: %s: %s", a.Verdict, a.Detail)
	}
}
