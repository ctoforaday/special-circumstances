package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Clock derives the record's two windows for a fold that already holds the events in id order:
// the EPOCH (how many times the chair has registered, this row included) and the SITTING (how
// many times this row's own seat has). It is the same definition events_w computes in SQL
// (plans/roundless.md §III.A.0), for the readers that have the slice and not the connection —
// and it is the ONLY other copy: TestClockAgreesWithEventsW pins the two together over a record
// with several seats, a re-sat chair and work between the registers.
//
// Advance once per event, in order. A fold that skips an event before advancing desyncs the
// clock, so the call is the first statement of the loop body, ahead of any continue. And the
// slice MUST contain the register events: a fold over events filtered to body types has no
// chair registers to count and reads every row as epoch 0 — the same bytes as a run in which no
// chair ever sat. A loader that narrows by type includes EVENT_TYPE_REGISTER for this reason.
//
// NOTHING IS STAMPED. The envelope carried this number as `epoch` and the record paid for it at
// every write; the number is a count over rows already written, so it is counted when wanted.
type Clock struct {
	epoch    int
	sittings map[string]int
}

// Advance folds one event in and returns the window it sits in.
func (c *Clock) Advance(e *Event) recordsql.Window {
	if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
		if c.sittings == nil {
			c.sittings = map[string]int{}
		}
		c.sittings[e.GetSeatId()]++
		if e.GetSeatId() == "red-chair" {
			c.epoch++
		}
	}
	return recordsql.Window{Epoch: c.epoch, Sitting: c.sittings[e.GetSeatId()]}
}

// ActClock is the OTHER half of the ruling, and the two halves are a pair: Clock COUNTS TURNS, and
// ActClock says WHICH SITTING AN ACT BELONGS TO. A sitting-record repair is a sitting — the seat was
// handed a prompt — so Clock counts it, and its tool-call cap, its dispatch number and the events_w
// ordinal are its own. But a repair exists to complete the record ANOTHER sitting owes, so its acts
// are that sitting's, and every reader that attributes an act to a sitting reads them here.
//
// A REPAIR REGISTER OPENS NONE, which is the whole difference: this counts only the registers that
// open a sitting (repairs_sitting absent), so a repair and everything after it carries the number of
// the sitting it repairs. That is exact rather than approximate — a repair may name only its seat's
// LATEST sitting (checkRepair), so the sitting it completes is always the one this counter is on.
// It is the same set of registers dispatchLedger opens a sitting from, so the number a reader
// renders and the span sittingCloser bounds cannot disagree about which sittings exist.
//
// THE EPOCH IS THE SAME NUMBER, and by construction, not by luck: only a blue seat may repair
// (checkRepair refuses every other role), and the epoch counts red-chair's registers. So a caller
// that wants the epoch may read it off either clock, and TestTheTwoClocksDifferOnlyOnARepairsSeat
// holds that.
//
// Advance once per event, in order, exactly as Clock's contract says. A REGISTER WHOSE BODY DID NOT
// DECODE OPENS A SITTING: the repair is a claim the body carries, so a register that cannot be read
// has not claimed one, and counting it as a turn is the degradation that invents nothing — the
// alternative folds two sittings into one and reads exactly like a record with one fewer.
type ActClock struct {
	epoch    int
	sittings map[string]int
}

// Advance folds one event in and returns the window its acts belong to.
func (c *ActClock) Advance(e *Event) recordsql.Window {
	if b, ok := recordpb.BodyAs[*recordpb.Register](e); e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && (!ok || b.RepairsSitting == nil) {
		if c.sittings == nil {
			c.sittings = map[string]int{}
		}
		c.sittings[e.GetSeatId()]++
		if e.GetSeatId() == "red-chair" {
			c.epoch++
		}
	}
	return recordsql.Window{Epoch: c.epoch, Sitting: c.sittings[e.GetSeatId()]}
}

// CurrentEpochOf is the epoch the debate's work has reached: the epoch of the last event that
// is not a register. A chair that has just sat opens a new epoch on the record, but until
// something is done in it the current one is still the last with work in it — which is what
// "stale since the current epoch" has to mean for a line of inquiry pursued in the previous one.
func CurrentEpochOf(evs []*Event) int {
	var clk Clock
	cur := 0
	for _, e := range evs {
		w := clk.Advance(e)
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			cur = w.Epoch
		}
	}
	return cur
}
