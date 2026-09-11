package record

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func dispatchBlue(t *testing.T, gaps ...string) *Event {
	return recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: gaps})
}

func closeGap(t *testing.T, lens, gap string) *Event {
	return recordtest.Event(t, lens, &recordpb.Close{GapId: proto.String(gap)})
}

func registers(t *testing.T, seat string) *Event {
	return recordtest.Event(t, seat, &recordpb.Register{})
}

func editAnswers(t *testing.T, seat, gap string) *Event {
	return recordtest.Event(t, seat, &recordpb.BlueEdit{Answers: proto.String(gap)})
}

// OWED = DISPATCHED ONTO, OPEN WHEN BLUE SAT, AND ANSWERED BY BLUE'S EDIT IN THAT SITTING.
//
// The B4 ordering (#868) is the first case: the lenses sat first and closed both gaps, blue found
// nothing to repair and filed no row, and the engine's abort on that empty manifest killed a
// $16.91 run. The rebuttal case is the second half of the same rule: the constitutions say one row
// per REPAIRED gap, and a gap blue argued against without editing was not repaired.
func TestManifestOwedIsARepairOfAGapStillOpenWhenBlueSat(t *testing.T) {
	cases := map[string]struct {
		evs  []*Event
		want []string
	}{
		"B4: both closed before blue registered": {
			evs: []*Event{
				registers(t, "red-chair"), dispatchBlue(t, "G1", "G2"),
				registers(t, "red-lens-logic"), closeGap(t, "red-lens-logic", "G1"),
				registers(t, "red-lens-voice"), closeGap(t, "red-lens-voice", "G2"),
				registers(t, "blue-respond"),
			},
			want: nil,
		},
		"closed first, then edited anyway: still not owed": {
			evs: []*Event{
				dispatchBlue(t, "G1"), closeGap(t, "red-lens-logic", "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			want: nil,
		},
		"one closed first, one open and repaired: only the repair is owed": {
			evs: []*Event{
				dispatchBlue(t, "G1", "G2"), closeGap(t, "red-lens-logic", "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G2"),
			},
			want: []string{"G2"},
		},
		"rebutted, no edit: not owed": {
			evs: []*Event{
				dispatchBlue(t, "G1", "G2"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G2"),
			},
			want: []string{"G2"},
		},
		"a close AFTER blue sat is a closure of blue's repair, and the row was owed": {
			evs: []*Event{
				dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"), closeGap(t, "red-lens-evidence", "G1"),
			},
			want: []string{"G1"},
		},
		"a dispatch blue never sat owes nothing, whatever was edited": {
			evs:  []*Event{dispatchBlue(t, "G1"), editAnswers(t, "blue-respond", "G1")},
			want: nil,
		},
		"an edit answering a gap blue was not dispatched onto is not owed": {
			evs: []*Event{
				dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G7"),
			},
			want: nil,
		},
		"another seat's edit repairs nothing on blue-respond's account": {
			evs: []*Event{
				dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-synthesize", "G1"),
			},
			want: nil,
		},
		"the lens's own dispatch engages nobody on blue's behalf": {
			evs: []*Event{
				recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("red-lens-logic"), GapIds: []string{"G1"}}),
				registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			want: nil,
		},
		"the sitting ends at the next dispatch: an edit then answers the new engagement only": {
			evs: []*Event{
				dispatchBlue(t, "G1"), registers(t, "blue-respond"),
				dispatchBlue(t, "G2"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			want: nil,
		},
		"owed once per gap across sittings, in first-owed order; a re-prompt register opens nothing": {
			evs: []*Event{
				dispatchBlue(t, "G2"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G2"), registers(t, "blue-respond"),
				dispatchBlue(t, "G1", "G2"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"), editAnswers(t, "blue-respond", "G2"),
			},
			want: []string{"G2", "G1"},
		},
		"a close from an earlier epoch does not reach into the next window": {
			evs: []*Event{
				closeGap(t, "red-lens-logic", "G1"), dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			want: []string{"G1"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ManifestOwed(c.evs); !reflect.DeepEqual(got, c.want) {
				t.Errorf("owed = %v, want %v", got, c.want)
			}
		})
	}
}

func TestManifestUnreceiptedIsOwedLessTheRows(t *testing.T) {
	evs := []*Event{
		dispatchBlue(t, "G1", "G2"), registers(t, "blue-respond"),
		editAnswers(t, "blue-respond", "G1"), editAnswers(t, "blue-respond", "G2"),
		recordtest.Event(t, "blue-respond", &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("recomputed")}),
	}
	if got := ManifestUnreceipted(evs); !reflect.DeepEqual(got, []string{"G2"}) {
		t.Errorf("unreceipted = %v, want [G2]", got)
	}
}
