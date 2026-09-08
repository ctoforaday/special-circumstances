package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE ROUND A WRITE STAMPS IS THE EPOCH — the count of red-chair registers on the record at the
// moment of the write — and a seat's NAME has nothing to do with it (plans/roundless.md §III.A.0).
//
// These three tests replace the ones that asked the round-inference what a name meant: a terminal seat used to
// be special-cased to "the last round on the record", synthesis seats to 0, and an empty run to
// -1 "unknown". Under the epoch none of those is a rule — they are all the same count, and the
// count is always defined. An empty record has no chair register, so the epoch is 0, and 0 is the
// honest answer: no dispatch cycle has opened. There is no -1 any more because there is nothing
// unknown; the record is the source.
func TestATerminalSeatIsStampedWithTheEpochItActsIn(t *testing.T) {
	runDir := newRun(t)
	for range 3 {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, ""); err != nil {
		t.Fatal(err)
	}
	ev, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, &recordpb.Observe{Text: proto.String("closing")})
	if err != nil {
		t.Fatal(err)
	}
	if got := int(ev.GetRound()); got != 3 {
		t.Errorf("terminal act stamped round %d, want 3 — the chair has registered three times, so this is epoch 3", got)
	}
}

func TestSynthesisSeatsAreEpochZeroBecauseNoChairHasSat(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"frontier", "blue-synthesize", "blue-lane-2"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: s}, ""); err != nil {
			t.Fatal(err)
		}
		ev, err := Append(Identity{Run: mustRun(t, runDir), SeatID: s}, &recordpb.Observe{Text: proto.String("x")})
		if err != nil {
			t.Fatal(err)
		}
		if got := int(ev.GetRound()); got != 0 {
			t.Errorf("%s stamped round %d before any chair register; want 0 — the base phase is epoch 0", s, got)
		}
	}
	// And the chair's own first register opens epoch 1 — the register row is the first of its epoch.
	_, ev, err := registerReturningEvent(t, runDir, "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	if got := int(ev.GetRound()); got != 1 {
		t.Errorf("the chair's first register stamped round %d, want 1 — it opens its own epoch", got)
	}
}

func TestAnEmptyRunIsEpochZeroNotUnknown(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, ""); err != nil {
		t.Fatal(err)
	}
	ev, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, &recordpb.Observe{Text: proto.String("x")})
	if err != nil {
		t.Fatal(err)
	}
	if got := int(ev.GetRound()); got != 0 {
		t.Errorf("round on a run with no chair = %d, want 0: there is no unknown, the count is simply zero", got)
	}
}

// registerReturningEvent registers and hands back the register event itself, which RegisterSeat's
// signature does not — it returns the seq and nonce — so the test can read what the register row
// was stamped with.
func registerReturningEvent(t *testing.T, runDir, seatID string) (int, *recordpb.Event, error) {
	t.Helper()
	run := mustRun(t, runDir)
	seq, _, err := RegisterSeat(Identity{Run: run, SeatID: seatID}, "")
	if err != nil {
		return 0, nil, err
	}
	m, err := MergedEvents(run)
	if err != nil {
		return 0, nil, err
	}
	for i := len(m.Events) - 1; i >= 0; i-- {
		e := m.Events[i]
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && e.GetSeatId() == seatID {
			return seq, e, nil
		}
	}
	t.Fatalf("no register event for %s after RegisterSeat", seatID)
	return 0, nil, nil
}
