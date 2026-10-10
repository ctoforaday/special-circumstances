package record

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

const positionItem = "this sitting's position is missing"

// THE PREDICATE NAMES TWO SEATS, AND THE ROSTER IS WHAT IT IS HELD AGAINST: every seat the tool
// knows is asked, so a seat added to the roster is on one side of the duty by a decision and not
// by an omission from a hand list.
func TestSeatOwesPositionNamesTheChairAndBlueRespondAndNoOtherSeat(t *testing.T) {
	want := map[string]bool{"red-chair": true, "blue-respond": true}
	asked := 0
	for id := range seatclass.Seats {
		asked++
		if got := SeatOwesPosition(id); got != want[id] {
			t.Errorf("SeatOwesPosition(%q) = %v, want %v", id, got, want[id])
		}
	}
	for id := range want {
		if _, on := seatclass.Seats[id]; !on {
			t.Errorf("%s owes a position and is not on the roster", id)
		}
	}
	if asked < 10 {
		t.Fatalf("the roster holds %d seats — the predicate was asked about too few to mean anything", asked)
	}
}

// BLUE-RESPOND'S WORK LIST BLOCKS ON AN OWED POSITION, and on nothing else of the kind: a sitting
// that filed one, a sitting that found every engaged gap closed, and every blue seat that owes none
// hold no such item.
func TestTheWorkListBlocksBlueRespondOnAnOwedPosition(t *testing.T) {
	engaged := func(t *testing.T) *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "blue-synthesize", "blue-lane-1", "frontier", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
			register("red-chair").dispatch(2, "blue-respond", "G1")
	}
	t.Run("an open engaged gap and no position blocks", func(t *testing.T) {
		s := sittingOfRunT(t, engaged(t).registerAs("blue-respond", "blue-a").seed(), "blue", "blue-respond")
		if !hasItem(s, positionItem) || s.Complete {
			t.Fatalf("blue-respond sits on G1, open, with no position: want a blocking item and complete=false, got complete=%v open=%+v", s.Complete, s.Open)
		}
		for _, it := range s.Open {
			if hasItem(SittingJSON{Open: []Item{it}}, positionItem) && !it.Blocks {
				t.Errorf("the position item does not block: %+v", it)
			}
		}
	})
	t.Run("a sitting its agent's stop closed still lists it", func(t *testing.T) {
		s := sittingOfRunT(t, engaged(t).registerAs("blue-respond", "blue-a").stop("blue-a").seed(), "blue", "blue-respond")
		if !hasItem(s, positionItem) {
			t.Fatalf("a closed sitting short of its position is not listed: %+v", s.Open)
		}
	})
	t.Run("a filed position clears it", func(t *testing.T) {
		s := sittingOfRunT(t, engaged(t).registerAs("blue-respond", "blue-a").position("blue-respond").seed(), "blue", "blue-respond")
		if hasItem(s, positionItem) {
			t.Fatalf("the position is on the record and the list still asks for it: %+v", s.Open)
		}
	})
	t.Run("every engaged gap found closed owes none", func(t *testing.T) {
		s := sittingOfRunT(t, engaged(t).closeGap(evLens, "G1").registerAs("blue-respond", "blue-a").seed(), "blue", "blue-respond")
		if hasItem(s, positionItem) {
			t.Fatalf("G1 was closed before blue sat and the list asks for a position: %+v", s.Open)
		}
	})
	for _, seat := range []string{"blue-lane-1", "frontier", "blue-synthesize"} {
		t.Run(seat+" has no such item", func(t *testing.T) {
			s := sittingOfRunT(t, engaged(t).register(seat).seed(), "blue", seat)
			if hasItem(s, positionItem) {
				t.Fatalf("%s owes no position and its list asks for one: %+v", seat, s.Open)
			}
		})
	}
}

// THE CHAIR'S WORK LIST DOES NOT MOVE WITH ITS POSITION. Its blocking items are the PASS gate's
// blockers; two chair sittings that differ only in holding a position read the same list.
func TestTheChairsWorkListIsTheSameWithAndWithoutAPosition(t *testing.T) {
	sat := func(t *testing.T) *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
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

// THE FOUR STATES, EACH FROM THE RECORD SHAPE THAT DEFINES IT, at both readings. A sitting nothing
// about its seat has closed is unresolved while the run runs and missing after it; every other
// shape reads the same either way. A lane, the frontier and the synthesizer have no row.
func TestPositionSittingsReadsTheFourStates(t *testing.T) {
	position := func(t *testing.T, seat string) *Event {
		return recordtest.Event(t, seat, &recordpb.Position{Text: proto.String("what the bench is asked to weigh")})
	}
	type row struct {
		seat  string
		state PositionState
	}
	cases := map[string]struct {
		evs            func(t *testing.T) []*Event
		running, after []row
	}{
		"blue-respond filed": {
			evs: func(t *testing.T) []*Event {
				return []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), position(t, "blue-respond"), agentStops(t, "blue-a")}
			},
			running: []row{{"blue-respond", PositionFiled}}, after: []row{{"blue-respond", PositionFiled}},
		},
		"blue-respond closed with none": {
			evs: func(t *testing.T) []*Event {
				return []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), agentStops(t, "blue-a")}
			},
			running: []row{{"blue-respond", PositionMissing}}, after: []row{{"blue-respond", PositionMissing}},
		},
		"blue-respond in flight with none": {
			evs: func(t *testing.T) []*Event {
				return []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a")}
			},
			running: []row{{"blue-respond", PositionUnresolved}}, after: []row{{"blue-respond", PositionMissing}},
		},
		"blue-respond found its gap closed": {
			evs: func(t *testing.T) []*Event {
				return []*Event{dispatchBlue(t, "G1"), closeGap(t, "red-lens-logic", "G1"), registersAs(t, "blue-respond", "blue-a"), agentStops(t, "blue-a")}
			},
			running: []row{{"blue-respond", PositionNotOwed}}, after: []row{{"blue-respond", PositionNotOwed}},
		},
		"the chair filed, then sat again with none": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, "red-chair", "chair-a"), position(t, "red-chair"), agentStops(t, "chair-a"), registersAs(t, "red-chair", "chair-b")}
			},
			running: []row{{"red-chair", PositionFiled}, {"red-chair", PositionUnresolved}},
			after:   []row{{"red-chair", PositionFiled}, {"red-chair", PositionMissing}},
		},
		"the chair's next opening closes a sitting with none": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, "red-chair", "chair-a"), registersAs(t, "red-chair", "chair-b"), position(t, "red-chair")}
			},
			running: []row{{"red-chair", PositionMissing}, {"red-chair", PositionFiled}},
			after:   []row{{"red-chair", PositionMissing}, {"red-chair", PositionFiled}},
		},
		"a lane, the frontier and the synthesizer have no row": {
			evs: func(t *testing.T) []*Event {
				return []*Event{registersAs(t, "blue-lane-1", "lane-a"), agentStops(t, "lane-a"),
					registersAs(t, "frontier", "frontier-a"), agentStops(t, "frontier-a"),
					registersAs(t, "blue-synthesize", "synth-a"), position(t, "blue-synthesize"), agentStops(t, "synth-a")}
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := loadedT(t, append(gapsExist(t, "G1"), c.evs(t)...)...)
			for _, read := range []struct {
				when ReadWhen
				want []row
			}{{WhileRunning, c.running}, {AfterTheRun, c.after}} {
				var got []row
				for _, p := range PositionSittings(m.Events, m.At, read.when) {
					got = append(got, row{p.Seat, p.State})
					if (p.Absence() == "") != (p.State == PositionFiled) {
						t.Errorf("%s sitting %d is %s and its absence reads %q: only a filed position has none", p.Seat, p.Ordinal, p.State, p.Absence())
					}
				}
				if !reflect.DeepEqual(got, read.want) {
					t.Errorf("read %v: rows = %+v, want %+v", read.when, got, read.want)
				}
			}
		})
	}
}
