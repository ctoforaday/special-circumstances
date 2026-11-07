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
	return recordtest.Event(t, lens, &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")})
}

// gapsExist mints each gap a fixture names, ahead of it, so the record can hold the closes that
// name them. The mints are a lens's acts before any sitting of blue's, and no blue
// sitting reads them.
func gapsExist(t *testing.T, gaps ...string) []*Event {
	var out []*Event
	for _, g := range gaps {
		out = append(out, mintsGap(t, "red-lens-logic", g))
	}
	return out
}

func registers(t *testing.T, seat string) *Event {
	return recordtest.Event(t, seat, &recordpb.Register{})
}

func editAnswers(t *testing.T, seat, gap string) *Event {
	return recordtest.Event(t, seat, &recordpb.BlueEdit{Answers: proto.String(gap)})
}

// A SITTING OWES AN ANSWER ON THE GAPS ITS DISPATCH NAMED THAT WERE STILL OPEN WHEN BLUE SAT (#868).
// The lenses sit before blue, so a lens can close a gap between the dispatch that engaged blue on it
// and blue's register; that gap is found closed. A close after the register is a closure of blue's
// answer, and the gap was open when blue sat. A dispatch blue never sat has no sitting.
func TestABlueSittingSplitsItsEngagedGapsAtItsRegister(t *testing.T) {
	cases := map[string]struct {
		evs               []*Event
		sittings          int
		open, foundClosed []string
	}{
		"both closed before blue registered": {
			evs: []*Event{
				registers(t, "red-chair"), dispatchBlue(t, "G1", "G2"),
				registers(t, "red-lens-logic"), closeGap(t, "red-lens-logic", "G1"),
				registers(t, "red-lens-voice"), closeGap(t, "red-lens-voice", "G2"),
				registers(t, "blue-respond"),
			},
			sittings: 1, foundClosed: []string{"G1", "G2"},
		},
		"one closed first, one open": {
			evs: []*Event{
				dispatchBlue(t, "G1", "G2"), closeGap(t, "red-lens-logic", "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G2"),
			},
			sittings: 1, open: []string{"G2"}, foundClosed: []string{"G1"},
		},
		"a close AFTER blue sat leaves the gap open in the sitting": {
			evs: []*Event{
				dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"), closeGap(t, "red-lens-evidence", "G1"),
			},
			sittings: 1, open: []string{"G1"},
		},
		"a dispatch blue never sat has no sitting": {
			evs:      []*Event{dispatchBlue(t, "G1"), editAnswers(t, "blue-respond", "G1")},
			sittings: 0,
		},
		"the lens's own dispatch engages nobody on blue's behalf": {
			evs: []*Event{
				recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("red-lens-logic"), GapIds: []string{"G1"}}),
				registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			sittings: 0,
		},
		"a close from an earlier epoch does not reach into the next window": {
			evs: []*Event{
				closeGap(t, "red-lens-logic", "G1"), dispatchBlue(t, "G1"), registers(t, "blue-respond"), editAnswers(t, "blue-respond", "G1"),
			},
			sittings: 1, open: []string{"G1"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ss := blueSittingsT(t, append(gapsExist(t, "G1", "G2"), c.evs...), WhileRunning)
			if len(ss) != c.sittings {
				t.Fatalf("sittings = %d, want %d: %+v", len(ss), c.sittings, ss)
			}
			if c.sittings == 0 {
				return
			}
			var closed []string
			for _, f := range ss[0].FoundClosed {
				closed = append(closed, f.GapID)
			}
			if !reflect.DeepEqual(ss[0].Open, c.open) || !reflect.DeepEqual(closed, c.foundClosed) {
				t.Errorf("open = %v, found closed = %v; want %v and %v", ss[0].Open, closed, c.open, c.foundClosed)
			}
		})
	}
}

func registersAs(t *testing.T, seat, agent string) *Event {
	return recordtest.Event(t, seat, &recordpb.Register{AgentId: proto.String(agent)})
}

// agentStops is the sitting_close the SubagentStop hook writes when that agent returns.
func agentStops(t *testing.T, agent string) *Event {
	return recordtest.Event(t, HarnessSeat, &recordpb.SittingClose{AgentId: proto.String(agent)})
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
		ss := blueSittingsT(t, evs, WhileRunning)
		if len(ss) != 1 || !ss[0].Unresolved {
			t.Fatalf("blue has neither registered again nor returned: want one unresolved sitting, got %+v", ss)
		}
		if got := actTypes(ss[0]); !reflect.DeepEqual(got, []string{"EVENT_TYPE_REGISTER", "EVENT_TYPE_BLUE_EDIT"}) {
			t.Errorf("the chair's row is not blue's act and ends nothing: acts = %v, want the register and the edit", got)
		}
	})
	t.Run("the agent's stop closes it, and an act after the stop is outside it", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			agentStops(t, "blue-a"),
			editAnswers(t, "blue-respond", "G1"),
		}
		ss := blueSittingsT(t, evs, WhileRunning)
		if len(ss) != 1 || ss[0].Unresolved {
			t.Fatalf("blue's agent returned: want one closed sitting, got %+v", ss)
		}
		if got := actTypes(ss[0]); !reflect.DeepEqual(got, []string{"EVENT_TYPE_REGISTER"}) {
			t.Errorf("the edit follows the stop: acts = %v, want the register alone", got)
		}
	})
	t.Run("a stop by another agent closes nothing", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			agentStops(t, "lens-agent"), editAnswers(t, "blue-respond", "G1"),
		}
		if ss := blueSittingsT(t, evs, WhileRunning); len(ss) != 1 || !ss[0].Unresolved || len(ss[0].Acts) != 2 {
			t.Errorf("another agent's return is not blue's: want one unresolved sitting holding the edit, got %+v", ss)
		}
	})
	t.Run("blue's next register closes it, and the next sitting takes the acts after it", func(t *testing.T) {
		evs := []*Event{
			dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"),
			chairDispatches(t, "blue-respond", "G2"), registersAs(t, "blue-respond", "blue-b"),
			editAnswers(t, "blue-respond", "G1"), agentStops(t, "blue-b"),
		}
		ss := blueSittingsT(t, evs, WhileRunning)
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
