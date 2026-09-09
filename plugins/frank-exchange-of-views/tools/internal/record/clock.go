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
