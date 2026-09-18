package record

import (
	"reflect"
	"strings"
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
		"the sitting ends at blue's next register: an edit then answers the new engagement only": {
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
			if got := ManifestOwed(c.evs, WhileRunning).Gaps; !reflect.DeepEqual(got, c.want) {
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
	if got := ManifestUnreceipted(evs, WhileRunning).Gaps; !reflect.DeepEqual(got, []string{"G2"}) {
		t.Errorf("unreceipted = %v, want [G2]", got)
	}
}

func registersAs(t *testing.T, seat, agent string) *Event {
	return recordtest.Event(t, seat, &recordpb.Register{AgentId: proto.String(agent)})
}

// agentStops is the sitting_close the SubagentStop hook writes when that agent returns.
func agentStops(t *testing.T, agent string) *Event {
	return recordtest.Event(t, HarnessSeat, &recordpb.SittingClose{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:blue-researcher")})
}

func chairDispatches(t *testing.T, seat string, gaps ...string) *Event {
	return recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat), GapIds: gaps})
}

// actTypes is a sitting's acts as their event types, for a readable comparison.
func actTypes(s BlueSitting) []string {
	var out []string
	for _, e := range s.Acts {
		out = append(out, e.GetType().String())
	}
	return out
}

// BLUE'S SITTING ENDS ON BLUE'S OWN ACTS (#1002). The chair writing a dispatch row that names blue is
// the chair's act, and a sitting bounded by it depends on a rule enforced at the chair's write path.
// The sitting ends at blue's next register or its agent's stop — sittingCloser, the rule the exchange
// fold reads — and with neither on the record it is Unresolved, its acts read to the end of the record.
func TestBlueSittingEndsOnBluesOwnActsNotTheChairsDispatchRow(t *testing.T) {
	t.Run("a later dispatch row naming blue does not close the sitting", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			chairDispatches(t, "blue-respond", "G2"),
			editAnswers(t, "blue-respond", "G1"),
		}
		ss := BlueSittings(evs, WhileRunning)
		if len(ss) != 1 || !ss[0].Unresolved {
			t.Fatalf("blue has neither registered again nor returned: want one unresolved sitting, got %+v", ss)
		}
		if got := actTypes(ss[0]); !reflect.DeepEqual(got, []string{"EVENT_TYPE_REGISTER", "EVENT_TYPE_BLUE_EDIT"}) {
			t.Errorf("the chair's row is not blue's act and ends nothing: acts = %v, want the register and the edit", got)
		}
		if got := ManifestOwed(evs, WhileRunning); !reflect.DeepEqual(got.Gaps, []string{"G1"}) || got.Unresolved != 1 {
			t.Errorf("the edit is in the sitting and the sitting is not closed: owed = %+v, want G1 with 1 unresolved", got)
		}
	})
	t.Run("the agent's stop closes it, and an act after the stop is outside it", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			agentStops(t, "blue-a"),
			editAnswers(t, "blue-respond", "G1"),
		}
		ss := BlueSittings(evs, WhileRunning)
		if len(ss) != 1 || ss[0].Unresolved {
			t.Fatalf("blue's agent returned: want one closed sitting, got %+v", ss)
		}
		if got := actTypes(ss[0]); !reflect.DeepEqual(got, []string{"EVENT_TYPE_REGISTER"}) {
			t.Errorf("the edit follows the stop: acts = %v, want the register alone", got)
		}
		if got := ManifestOwed(evs, WhileRunning); got.Gaps != nil || got.Unresolved != 0 {
			t.Errorf("owed = %+v, want nothing and nothing unresolved", got)
		}
	})
	t.Run("a stop by another agent closes nothing", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			agentStops(t, "lens-agent"), editAnswers(t, "blue-respond", "G1"),
		}
		if ss := BlueSittings(evs, WhileRunning); len(ss) != 1 || !ss[0].Unresolved || len(ss[0].Acts) != 2 {
			t.Errorf("another agent's return is not blue's: want one unresolved sitting holding the edit, got %+v", ss)
		}
	})
	t.Run("blue's next register closes it, and the next sitting takes the acts after it", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			chairDispatches(t, "blue-respond", "G2"), registersAs(t, "blue-respond", "blue-b"),
			editAnswers(t, "blue-respond", "G1"), agentStops(t, "blue-b"),
		}
		ss := BlueSittings(evs, WhileRunning)
		if len(ss) != 2 || ss[0].Unresolved || ss[1].Unresolved {
			t.Fatalf("want two closed sittings, got %+v", ss)
		}
		if got := actTypes(ss[0]); !reflect.DeepEqual(got, []string{"EVENT_TYPE_REGISTER"}) {
			t.Errorf("first sitting acts = %v, want the register alone", got)
		}
		if !reflect.DeepEqual(ss[1].Engaged, []string{"G2"}) {
			t.Errorf("second sitting engaged = %v, want [G2]", ss[1].Engaged)
		}
	})
}

// AN UNRESOLVED SITTING IS WORDED ONCE. The owed set is a floor while a sitting cannot be closed, and
// NotMeasured is the one place that says so; a record that closes every sitting says nothing.
func TestManifestOwingWordsTheUnresolvedSittings(t *testing.T) {
	open := ManifestOwed([]*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a")}, WhileRunning)
	if open.Unresolved != 1 || !strings.Contains(open.NotMeasured(), "1 blue sitting(s) NOT MEASURED") {
		t.Errorf("an unclosed sitting must be worded as not measured: %+v %q", open, open.NotMeasured())
	}
	closed := ManifestOwed([]*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), agentStops(t, "blue-a")}, WhileRunning)
	if closed.Unresolved != 0 || closed.NotMeasured() != "" {
		t.Errorf("a closed sitting says nothing: %+v %q", closed, closed.NotMeasured())
	}
	if u := ManifestUnreceipted([]*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), editAnswers(t, "blue-respond", "G1")}, WhileRunning); u.Unresolved != 1 {
		t.Errorf("the unreceipted set carries the unresolved sitting: %+v", u)
	}
}
