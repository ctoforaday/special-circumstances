package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// lastWindow reads the LAST event's epoch and sitting back off the record. Nothing stamps them
// (plans/roundless.md §III.A.2): they are the events_w view's answer, which is what every reader
// gets, so this is the assertion a test about "which epoch did this act land in" has to make.
func lastWindow(t *testing.T, runDir string) recordsql.Window {
	t.Helper()
	db, err := openRunForRead(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	evs, ws, err := recordsql.EventsW(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) == 0 {
		t.Fatal("no events on the record")
	}
	return ws[len(ws)-1]
}

func TestATerminalSeatActsInTheEpochTheChairHasReached(t *testing.T) {
	runDir := newRun(t)
	for range 3 {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, &recordpb.Observe{Text: proto.String("closing")}); err != nil {
		t.Fatal(err)
	}
	if got := lastWindow(t, runDir).Epoch; got != 3 {
		t.Errorf("terminal act is in epoch %d, want 3 — the chair has registered three times", got)
	}
}

func TestSynthesisSeatsAreEpochZeroBecauseNoChairHasSat(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"frontier", "blue-synthesize", "blue-lane-2"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: s}, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: s}, &recordpb.Observe{Text: proto.String("x")}); err != nil {
			t.Fatal(err)
		}
		if got := lastWindow(t, runDir).Epoch; got != 0 {
			t.Errorf("%s acts in epoch %d before any chair register; want 0 — the base phase is epoch 0", s, got)
		}
	}
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, ""); err != nil {
		t.Fatal(err)
	}
	if w := lastWindow(t, runDir); w.Epoch != 1 || w.Sitting != 1 {
		t.Errorf("the chair's first register = %+v, want epoch 1 sitting 1 — it opens its own epoch", w)
	}
}

func TestAnEmptyRunIsEpochZeroNotUnknown(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal"}, &recordpb.Observe{Text: proto.String("x")}); err != nil {
		t.Fatal(err)
	}
	if got := lastWindow(t, runDir).Epoch; got != 0 {
		t.Errorf("epoch on a run with no chair = %d, want 0: there is no unknown, the count is simply zero", got)
	}
}
