package record

import (
	"database/sql"
	"fmt"

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
	// latest is each seat's newest sitting, by the id of the event that opened it.
	latest map[string]int64
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

// eventsAt reads the whole record with its index — the read every fold over the run's stream takes.
func eventsAt(db *sql.DB) ([]*Event, WindowIndex, error) {
	evs, ws, err := recordsql.Events(db)
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
// 0 is "never sat", so an index the loader did not build panics here as Of does: answering 0 from
// no index would read every seat as one that has not sat.
func (x WindowIndex) LatestSittingOf(seat string) int64 {
	if x.of == nil {
		panic(fmt.Sprintf("record: no window index to ask for %s's latest sitting — the index was not built by the loader, so it holds no stored sitting", seat))
	}
	return x.latest[seat]
}
