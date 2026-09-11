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
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge", "judge-terminal", "assemble"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		reg("frontier"), // the base phase: before the first dispatch, outside the window
		reg("red-chair"), dispatch("red-lens-evidence"), dispatch("red-lens-logic"),
		reg("red-lens-evidence"), reg("red-lens-logic"),
		reg("red-chair"), dispatch("blue-respond", "G1"),
	}
	clean := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-chair"), reg("judge-terminal"), reg("assemble"))
	dir := t.TempDir()
	recordtest.Seed(t, dir, clean...)
	if a := DispatchParityAudit(runtest.Open(t, dir)); a.Verdict != "PASS" {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}

	n = 100
	stray := append(append([]*record.Event{}, head...), reg("blue-respond"), reg("red-lens-voice"), reg("red-chair"))
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, stray...)
	if a := DispatchParityAudit(runtest.Open(t, dir2)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice registered after dispatch 2") {
		t.Fatalf("a seat nobody dispatched sat unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	n = 200
	absent := append(append([]*record.Event{}, head...), reg("red-chair")) // blue never sat for dispatch 2
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, absent...)
	if a := DispatchParityAudit(runtest.Open(t, dir3)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "blue-respond was named in dispatch 2 and never registered") {
		t.Fatalf("an absent party went unnoticed: %s: %s", a.Verdict, a.Detail)
	}

	if a := DispatchParityAudit(runtest.Open(t, func() string { d := t.TempDir(); recordtest.Seed(t, d, reg("red-chair")); return d }())); a.Verdict != "SKIP" {
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
			at("harness", &recordpb.Cast{SeatIds: append([]string{"red-chair", "blue-respond", "judge", "assemble"}, lenses...)}),
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
		return append(evs, reg("assemble"))
	}

	dir := t.TempDir()
	recordtest.Seed(t, dir, warm(true, false)...)
	if a := DispatchParityAudit(runtest.Open(t, dir)); a.Verdict != "PASS" || !strings.HasPrefix(a.Detail, "3 dispatch(es)") {
		t.Fatalf("a warm run whose parties all sat = %s: %s — want PASS over 3 dispatches", a.Verdict, a.Detail)
	}

	n = 400
	dir2 := t.TempDir()
	recordtest.Seed(t, dir2, warm(false, false)...)
	if a := DispatchParityAudit(runtest.Open(t, dir2)); a.Verdict != "FAIL" || a.Detail != "red-lens-voice was named in dispatch 2 and never registered before the next" {
		t.Fatalf("a warm run whose voice lens never sat for dispatch 2 = %s: %s", a.Verdict, a.Detail)
	}

	// Registering AFTER the next dispatch is not sitting for this one: the voice lens's register
	// lands in dispatch 3's window, where it is a stray, and dispatch 2 still went unanswered.
	n = 500
	dir3 := t.TempDir()
	recordtest.Seed(t, dir3, warm(false, true)...)
	a := DispatchParityAudit(runtest.Open(t, dir3))
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice was named in dispatch 2 and never registered before the next") ||
		!strings.Contains(a.Detail, "red-lens-voice registered after dispatch 3 and was not a party to it") {
		t.Fatalf("a party that sat only after the next dispatch = %s: %s", a.Verdict, a.Detail)
	}
}
