package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// TWO THINGS, TWO NAMES (#1002, gblock 2026-09-18). A sitting is every prompt a seat is handed, so
// Clock counts a sitting-record repair as a turn of its own. A repair completes the record ANOTHER
// sitting owes, so ActClock attributes its acts to that sitting. This holds the two apart on one
// record and shows exactly where they differ: at the repair register and after it, on that seat.
func TestTheTwoClocksDifferOnlyOnARepairsSeat(t *testing.T) {
	blue := func(key string, body proto.Message) *Event { return recordtest.At(t, "blue-respond", key, body) }
	evs := []*Event{
		recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		blue("blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		blue("blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		// The re-prompt: a second agent, a second prompt, a register that names the sitting it repairs.
		blue("blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String("blue-respond:register:#1")}),
		blue("blue-respond:manifest_row:m1", &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("recomputed")}),
		// A later seat's act: neither clock moves for it, and blue's number is unchanged after it.
		recordtest.At(t, "red-lens-logic", "red-lens-logic:register:#1", &recordpb.Register{}),
	}

	type window struct{ epoch, sitting int }
	var turns, acts []window
	var tc Clock
	var ac ActClock
	for _, e := range evs {
		wt, wa := tc.Advance(e), ac.Advance(e)
		turns = append(turns, window{wt.Epoch, wt.Sitting})
		acts = append(acts, window{wa.Epoch, wa.Sitting})
		if wt.Epoch != wa.Epoch {
			t.Errorf("the epochs disagree at %s: turns %d, acts %d — only blue may repair, so the chair's count is the same on both", e.GetKey(), wt.Epoch, wa.Epoch)
		}
	}
	wantTurns := []int{1, 1, 1, 1, 0, 2, 2, 1}
	wantActs := []int{1, 1, 1, 1, 0, 1, 1, 1}
	for i := range evs {
		if turns[i].sitting != wantTurns[i] || acts[i].sitting != wantActs[i] {
			t.Errorf("%s: turns #%d acts #%d, want turns #%d acts #%d",
				evs[i].GetKey(), turns[i].sitting, acts[i].sitting, wantTurns[i], wantActs[i])
		}
	}
}

// A REGISTER NOTHING CAN READ OPENS A SITTING. The repair is a claim the body carries, so a register
// whose body did not decode has claimed nothing — counting it as a turn invents no repair, while
// the other reading folds two sittings into one and reads exactly like a record with one fewer.
func TestARegisterWithNoBodyOpensASitting(t *testing.T) {
	bodiless := &recordpb.Event{
		SeatId: proto.String("blue-respond"),
		Ts:     proto.String("2026-09-18T00:00:00Z"),
		Type:   recordpb.EventType_EVENT_TYPE_REGISTER.Enum(),
		Key:    proto.String("blue-respond:register:#1"),
	}
	if _, ok := recordpb.BodyAs[*recordpb.Register](bodiless); ok {
		t.Fatal("the fixture carries a body, so it tests nothing")
	}
	var ac ActClock
	if w := ac.Advance(bodiless); w.Sitting != 1 {
		t.Errorf("a register with no body = sitting %d, want 1", w.Sitting)
	}
}

// A MOTION FILED IN A REPAIR IS THE REPAIRED SITTING'S (#1002). The motion board attributes three
// acts to a sitting — the proposal, the filing and the ruling — and the report prints the filer's as
// `filed by <seat> #N`. A grade motion blue files while finishing sitting 1's record is sitting 1's.
func TestAMotionFiledInARepairCarriesTheSittingItCompletes(t *testing.T) {
	evs := []*Event{
		recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String("blue-respond:register:#1")}),
		recordtest.At(t, "blue-respond", "blue-respond:motion:#1", &recordpb.Motion{MotionId: proto.String("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE), Basis: proto.String("the consequence is bounded"),
			Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
				Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY), Proposed: recordtest.P(recordpb.Grade_GRADE_LOW)}}}),
	}
	ms := MotionsOf(evs)
	if len(ms) != 1 || ms[0].Sitting != 1 || ms[0].Epoch != 1 {
		t.Fatalf("motions = %+v, want M1 filed in blue-respond's sitting 1, epoch 1", ms)
	}
}
