package record

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// storedPlan is the plan event a chair's `dispatch next` stores, permitting a PASS or not.
func storedPlan(t *testing.T, permitted bool) *Event {
	p := emptyPlan()
	p.PassPermitted, p.Head = permitted, 1
	if !permitted {
		p.Blockers = []PlanBlocker{{Kind: BlockerMaterialGap, Subject: "G1", Owner: evLens}}
	}
	return recordtest.Event(t, chairSeat, &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())})
}

func gate(t *testing.T, v recordpb.Verdict) *Event {
	return recordtest.Event(t, chairSeat, &recordpb.Gate{Verdict: v.Enum()})
}

// THE CHAIR'S POSITION IS OWED WHERE IT WITHHELD A PASS ITS PLAN PERMITTED, AND NOWHERE ELSE — each
// state from the record shape that defines it, at both readings. A sitting nothing about the chair
// has closed is unresolved while the run runs and missing after it.
func TestTheChairOwesAPositionWhereItWithheldAPermittedPass(t *testing.T) {
	position := func(t *testing.T) *Event {
		return recordtest.Event(t, chairSeat, &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("why the PASS is withheld")})
	}
	cases := map[string]struct {
		evs            func(t *testing.T) []*Event
		running, after []PositionState
	}{
		"permitted, no PASS, none filed": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionMissing}, after: []PositionState{PositionMissing},
		},
		"permitted, a FAIL recorded, none filed": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), gate(t, recordpb.Verdict_VERDICT_FAIL), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionMissing}, after: []PositionState{PositionMissing},
		},
		"permitted, no PASS, filed": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), position(t), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionFiled}, after: []PositionState{PositionFiled},
		},
		"permitted, no PASS, in flight": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true)}
			},
			running: []PositionState{PositionUnresolved}, after: []PositionState{PositionMissing},
		},
		"permitted, the PASS recorded": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), gate(t, recordpb.Verdict_VERDICT_PASS), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionNotOwed}, after: []PositionState{PositionNotOwed},
		},
		"not permitted": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, false), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionNotOwed}, after: []PositionState{PositionNotOwed},
		},
		"no plan stored": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionNotOwed}, after: []PositionState{PositionNotOwed},
		},
		// THE SITTING'S LAST PLAN DECIDES: asked again after a motion, the plan the chair relays is
		// the last one, and a permitted plan stored in an earlier sitting answers nothing here.
		"the last plan of the sitting decides": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), storedPlan(t, false), agentStops(t, "chair-a")}
			},
			running: []PositionState{PositionNotOwed}, after: []PositionState{PositionNotOwed},
		},
		"a permitted plan in the sitting before": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), gate(t, recordpb.Verdict_VERDICT_PASS), agentStops(t, "chair-a"),
					registersAs(t, chairSeat, "chair-b"), agentStops(t, "chair-b")}
			},
			running: []PositionState{PositionNotOwed, PositionNotOwed}, after: []PositionState{PositionNotOwed, PositionNotOwed},
		},
		"the chair's next opening closes a sitting with none": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, chairSeat, "chair-a"), storedPlan(t, true), registersAs(t, chairSeat, "chair-b"), position(t)}
			},
			running: []PositionState{PositionMissing, PositionNotOwed}, after: []PositionState{PositionMissing, PositionNotOwed},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := loadedT(t, append(gapsExist(t, "G1"), c.evs(t)...)...)
			for _, read := range []struct {
				when ReadWhen
				want []PositionState
			}{{WhileRunning, c.running}, {AfterTheRun, c.after}} {
				var got []PositionState
				for _, p := range PositionSittings(m.Events, m.At, read.when) {
					got = append(got, p.State)
				}
				if !reflect.DeepEqual(got, read.want) {
					t.Errorf("read %v: states = %v, want %v", read.when, got, read.want)
				}
			}
		})
	}
}

// THE CHAIR'S WORK LIST DOES NOT MOVE WITH ITS POSITION. Its blocking items are the PASS gate's
// blockers; two chair sittings that differ only in holding a position read the same list.
func TestTheChairsWorkListIsTheSameWithAndWithoutAPosition(t *testing.T) {
	sat := func(t *testing.T) *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).mint(evLens, "G1", "medium").
			register("red-chair")
	}
	without := sittingOfRunT(t, sat(t).seed(), "chair", "red-chair")
	with := sittingOfRunT(t, sat(t).position("red-chair").seed(), "chair", "red-chair")
	if hasItem(without, "position") || hasItem(with, "position") {
		t.Fatalf("the chair's list names a position:\nwithout: %+v\nwith: %+v", without.Open, with.Open)
	}
	blocking := func(s SittingJSON) []string {
		var out []string
		for _, it := range s.Open {
			if it.Blocks {
				out = append(out, it.What)
			}
		}
		return out
	}
	if len(blocking(without)) == 0 {
		t.Fatal("the fixture's chair has nothing blocking: the comparison below would hold of two empty lists")
	}
	if !reflect.DeepEqual(blocking(without), blocking(with)) || without.Complete != with.Complete {
		t.Errorf("the chair's blocking items moved with its position:\nwithout: %+v\nwith: %+v", blocking(without), blocking(with))
	}
}

// BLUE-RESPOND OWES NO POSITION: its work list holds no position item at any sitting, open gaps
// and all. Its argument on a gap is its closing.
func TestBlueRespondsWorkListAsksForNoPosition(t *testing.T) {
	s := sittingOfRunT(t, newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "medium").
		register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-a").stop("blue-a").seed(), "blue", "blue-respond")
	if listsItem(s, "position") {
		t.Fatalf("blue-respond's list names a position: %+v", s.Open)
	}
}
