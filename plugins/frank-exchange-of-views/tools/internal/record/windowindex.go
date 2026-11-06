package record

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// WindowIndex is where each loaded event sits on the record, as the loader read it: its row id, the
// sitting the write path stored it in, and that sitting's owner and ranks (recordsql.Window). It
// is keyed by the event itself, so a fold that reorders or filters the slice
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
	// reach is, for each retire, the gaps it reaches (reachingRetiresOf), by the retire's events.id.
	// The whole-record loaders fill it from the same snapshot as the events; nil on a narrowed
	// index, and on one a test built from bare windows.
	reach map[int64][]reaching
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
	x := windowIndexOf(evs, ws)
	if x.reach, err = reachingRetiresOf(q); err != nil {
		return nil, WindowIndex{}, err
	}
	return evs, x, nil
}

// reaching is one gap a retire reaches: the gap, the retired anchor it reaches it through, whether
// through a reading (the carrying_reading view) rather than the anchor a gap is about, and the gap
// whose reading or about it is (from): gap itself, or an ancestor it carries the demand from.
type reaching struct {
	gap, anchor, from string
	carried           bool
}

// A REACHING RETIRE IS BLUE'S ACT ON EACH GAP IT REACHES (gblock, 2026-11-03; serving Closures are
// history and L1): a retire whose anchors hold an open gap's about_ref, or the anchor of a reading
// the gap carries, meets that gap's demand whatever gap the retire named, and without counting it
// as an act the lens's turn on the gap never comes. It is read here, off the record's rows, and
// never off a retire's own `answers`: `--answers` is optional, and a claim retire takes the anchor
// too. A gap closed before the retire is not reached — a closure is history.
//
// The gap-to-anchor reach is read through the demand_reach view, its one statement: the anchor a
// gap is about, the anchor of each reading it carries, and each of those of every ancestor it carries
// the demand from, which a retire reaches from the successor's mint. The three reads are joined here
// rather than in one statement: a retire's anchor meets a gap on the gap's side by mint.about_ref and
// verify.anchor, and verify.anchor is not indexed (verify_carried_by indexes the carrying side), so a
// join in SQL would build an automatic index on every load. A record with no retire reads the first,
// empty, and stops.
const (
	retiredAnchorsSQL = `SELECT "event_id", "value" FROM "retire_anchors" ORDER BY "event_id", "ord"`
	reachedGapsSQL    = `SELECT "gap_id", "anchor", "reached", "via" = 'reading', "from_gap" FROM "demand_reach" WHERE "anchor" IS NOT NULL
  ORDER BY "from_gap" <> "gap_id"`
	firstClosureSQL = `SELECT "gap_id", MIN("event_id") FROM "closure_act" GROUP BY "gap_id"`
)

// reachingRetiresOf is every reaching retire on the record, by the retire's events.id, each gap
// once (the first anchor of the retire that reaches it, through the gap's own about or reading before
// an ancestor's), in gap order.
func reachingRetiresOf(q recordsql.Querier) (map[int64][]reaching, error) {
	type retired struct {
		at     int64
		anchor string
	}
	var retires []retired
	if err := scanRows(q, retiredAnchorsSQL, func(rows *sql.Rows) error {
		var r retired
		err := rows.Scan(&r.at, &r.anchor)
		retires = append(retires, r)
		return err
	}); err != nil || len(retires) == 0 {
		return map[int64][]reaching{}, err
	}
	type target struct {
		gap, from string
		since     int64
		carried   bool
	}
	targets := map[string][]target{} // anchor -> the gaps about it, or carrying a reading of it
	if err := scanRows(q, reachedGapsSQL, func(rows *sql.Rows) error {
		var t target
		var a string
		err := rows.Scan(&t.gap, &a, &t.since, &t.carried, &t.from)
		targets[a] = append(targets[a], t)
		return err
	}); err != nil {
		return nil, err
	}
	closed := map[string]int64{} // gap -> its first closing act
	if err := scanRows(q, firstClosureSQL, func(rows *sql.Rows) error {
		var g string
		var at int64
		err := rows.Scan(&g, &at)
		closed[g] = at
		return err
	}); err != nil {
		return nil, err
	}
	out := map[int64][]reaching{}
	for _, r := range retires {
		for _, t := range targets[r.anchor] {
			c, isClosed := closed[t.gap]
			if t.since >= r.at || (isClosed && c < r.at) || slices.ContainsFunc(out[r.at], func(o reaching) bool { return o.gap == t.gap }) {
				continue
			}
			out[r.at] = append(out[r.at], reaching{gap: t.gap, anchor: r.anchor, from: t.from, carried: t.carried})
		}
	}
	for at := range out {
		slices.SortFunc(out[at], func(a, b reaching) int { return strings.Compare(a.gap, b.gap) })
	}
	return out, nil
}

// scanRows runs query and hands each row to scan.
func scanRows(q recordsql.Querier, query string, scan func(*sql.Rows) error) error {
	rows, err := q.Query(query)
	if err != nil {
		return fmt.Errorf("asking the record which gaps each retire reaches: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

// Reaches is the gaps the retire e reaches (reachingRetiresOf), in gap order; nil for any other
// event. A narrowed index panics, as in LatestSittingOf: it was not read beside the record's rows.
func (x WindowIndex) Reaches(e *Event) []reaching {
	if e.GetType() != recordpb.EventType_EVENT_TYPE_RETIRE {
		return nil
	}
	x.mustBeWhole("the gaps a retire reaches")
	return x.reach[x.Of(e).ID]
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
		panic(fmt.Sprintf("no window for a %s event by %s — the slice holds an event the loader did not produce, so it has no stored sitting", e.GetType(), e.GetSeatId()))
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
// brackets, its tool-call limits, the cast (HarnessSeat on the envelope) — nor a stored plan, the
// record's copy of what `dispatch next` printed: a chair that has just sat opens a new epoch on the
// record, and its hook's own bookkeeping lands in it, but until something is done in it the current
// one is still the last with work in it — which is what "stale since the current epoch" has to mean
// for an inquiry pursued in the previous one. Counted, the plan would move the epoch by being
// stored.
//
// A DUTY THE CHAIR CLEARS IN ITS OWN SITTING DOES NOT READ THIS. The chair's first act of a sitting
// moves this epoch, so a duty read off it falls due by that act; such a duty reads what it reads, as
// the stale-area spot-check reads each retired lens's pin.
//
// The last work act of a narrowed slice is not the record's, so a narrowed index panics here as it
// does in LatestSittingOf.
func (x WindowIndex) CurrentEpoch(evs []*Event) int {
	x.mustBeWhole("the current epoch")
	cur := 0
	for _, e := range evs {
		if t := e.GetType(); t == recordpb.EventType_EVENT_TYPE_REGISTER || t == recordpb.EventType_EVENT_TYPE_PLAN || e.GetSeatId() == HarnessSeat {
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
		panic(fmt.Sprintf("no whole-record window index to ask for %s — the index was built by hand or by a narrowed read, so it does not hold the record's newest acts", what))
	}
}
