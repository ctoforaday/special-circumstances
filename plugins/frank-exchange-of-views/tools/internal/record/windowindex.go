package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// WindowIndex is where each loaded event sits on the record, as the loader read it: its row id, the
// sitting the write path stored it in, and that sitting's owner and ranks (recordsql.Window). It
// is keyed by the event itself, so a fold that reorders or filters the slice — Live, Listing —
// still finds each event's window.
//
// THE STORED SITTING IS THE ONE DEFINITION. Which events open a sitting, and which sitting an act
// belongs to, is decided once at the write (recordsql's sittingOf): a hook bracket opens one, a
// register under the agent it bracketed joins it, a repair joins the seat's latest. A reader that
// asks "did this seat sit, and since when" asks this index, never the event types — counting
// registers over the slice is a second definition, and it missed every hook-opened sitting and read
// a bracket and its paired register as two.
//
// AN EVENT THE LOADER DID NOT PRODUCE IS A PANIC, not a zero window. A hand-built slice has no stored
// sitting, and reading it as sitting 0 / epoch 0 is the same bytes as an act before anything sat.
// Tests seed through the write path and load back.
type WindowIndex struct {
	of map[*Event]recordsql.Window
	// latest is each seat's newest sitting, by the id of the event that opened it. nil on an index a
	// narrowed read built: its slice need not hold any act of a seat's newest sitting.
	latest map[string]int64
}

// narrowedIndexOf indexes a narrowed read's windows, aligned by position with evs. Each window is
// the stored one; what the index cannot answer is which sitting is a seat's newest.
func narrowedIndexOf(evs []*Event, ws []recordsql.Window) WindowIndex {
	x := WindowIndex{of: make(map[*Event]recordsql.Window, len(evs))}
	for i, e := range evs {
		x.of[e] = ws[i]
	}
	return x
}

// windowIndexOf indexes the loader's windows, aligned by position with evs.
func windowIndexOf(evs []*Event, ws []recordsql.Window) WindowIndex {
	x := WindowIndex{of: make(map[*Event]recordsql.Window, len(evs)), latest: map[string]int64{}}
	for i, e := range evs {
		w := ws[i]
		x.of[e] = w
		if w.Owner != "" && w.SittingID > x.latest[w.Owner] {
			x.latest[w.Owner] = w.SittingID
		}
	}
	return x
}

// eventsAt reads the whole record with its index — the read every fold over the run's stream takes —
// asked of q: the run's handle, or a read transaction a reader asks every question of. A nil q is a
// run with no record yet, which holds no events.
func eventsAt(q recordsql.Querier) ([]*Event, WindowIndex, error) {
	if noRecord(q) {
		return nil, windowIndexOf(nil, nil), nil
	}
	evs, ws, err := recordsql.Events(q)
	if err != nil {
		return nil, WindowIndex{}, err
	}
	return evs, windowIndexOf(evs, ws), nil
}

// IDs is each event's row id, aligned with evs: the place sequence a fold compares, read off the
// same load as the events rather than a second query that a concurrent write could misalign.
func (x WindowIndex) IDs(evs []*Event) []int64 {
	out := make([]int64, len(evs))
	for i, e := range evs {
		out[i] = x.Of(e).ID
	}
	return out
}

// Of is e's window.
func (x WindowIndex) Of(e *Event) recordsql.Window {
	w, ok := x.of[e]
	if !ok {
		panic(fmt.Sprintf("record: no window for a %s event by %s — the slice holds an event the loader did not produce, so it has no stored sitting", e.GetType(), e.GetSeatId()))
	}
	return w
}

// Opens is the seat whose sitting e opened, and whether it opened one: a hook bracket opens its
// seat's sitting, a register opens one unless it joined its own bracket's or repairs one.
func (x WindowIndex) Opens(e *Event) (string, bool) {
	return x.Of(e).Opens()
}

// LatestSittingOf is the id of the event that opened the seat's newest sitting, or 0 if it has none.
//
// 0 is "never sat", so an index the whole-record loader did not build panics here as Of does:
// answering 0 from no index, or from a narrowed read's, would read a seat as one that has not sat or
// name a sitting before its newest.
func (x WindowIndex) LatestSittingOf(seat string) int64 {
	x.mustBeWhole(seat + "'s latest sitting")
	return x.latest[seat]
}

// CurrentEpoch is the epoch the debate's work has reached: the stored epoch of the last event in
// evs that is work. A register is not, and neither is anything the harness wrote — its sitting
// brackets, its tool-call limits, the cast (HarnessSeat on the envelope): a chair that has just sat
// opens a new epoch on the record, and its hook's own bookkeeping lands in it, but until something
// is done in it the current one is still the last with work in it — which is what "stale since the
// current epoch" has to mean for an avenue pursued in the previous one.
//
// The last work act of a narrowed slice is not the record's, so a narrowed index panics here as it
// does in LatestSittingOf.
func (x WindowIndex) CurrentEpoch(evs []*Event) int {
	x.mustBeWhole("the current epoch")
	cur := 0
	for _, e := range evs {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER || e.GetSeatId() == HarnessSeat {
			continue
		}
		cur = x.Of(e).Epoch
	}
	return cur
}

// LastEpoch is the stored epoch of the last event in evs, bookkeeping included: the epoch the record
// has reached, which counts the chair's sittings opened at or before it. A narrowed index panics, as
// in CurrentEpoch.
func (x WindowIndex) LastEpoch(evs []*Event) int {
	x.mustBeWhole("the epoch the record has reached")
	if len(evs) == 0 {
		return 0
	}
	return x.Of(evs[len(evs)-1]).Epoch
}

// mustBeWhole panics on an index the whole-record loader did not build: what asks it is a question
// about the record's newest acts, which a narrowed slice need not hold.
func (x WindowIndex) mustBeWhole(what string) {
	if x.latest == nil {
		panic(fmt.Sprintf("record: no whole-record window index to ask for %s — the index was built by hand or by a narrowed read, so it does not hold the record's newest acts", what))
	}
}
