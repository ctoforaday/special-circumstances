package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// familyReads is every answer the gap family takes off the record, read on one snapshot
// (familyReadsAt) and folded after it closes (gapStatesOf): the whole stream with its window index,
// and the gap view's rows.
type familyReads struct {
	merged Merged
	gaps   []familyGapRow
}

// familyGapRow is one row of the `gap` view as the family scans it.
type familyGapRow struct {
	id                     string
	epoch                  int
	open, material         bool
	sev, lik, imp, cx      sql.NullString
	closeEvent, benchEvent sql.NullInt64
	closedByBench          bool
}

// familyReadsAt reads the family's answers off q, the read transaction both come off: the whole
// record once (eventsAt, the one windowed ranking a family read pays) and the `gap` view once.
//
// ONE SNAPSHOT, BECAUSE THE VIEW NAMES AN ACT AND THE STREAM HOLDS ITS BODY. The gap view says which
// act closed a gap by its events.id, and the body that id names is looked up in the stream. Read at
// two moments, a close landing between them names an act the stream read did not load.
func familyReadsAt(q recordsql.Querier) (familyReads, error) {
	var r familyReads
	evs, win, err := eventsAt(q)
	if err != nil {
		return r, err
	}
	r.merged = Merged{Events: evs, At: win}
	if noRecord(q) {
		return r, nil
	}
	rows, err := q.Query(`SELECT "gap_id", "minted_epoch", "open",
	    "current_severity", "current_likelihood", "current_impact", "current_complexity_cost", "material",
	    "close_event", "bench_close_event", "closed_by_bench"
	  FROM "gap" ORDER BY "minted_event"`)
	if err != nil {
		return r, fmt.Errorf("record: asking the record for its gap states: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var g familyGapRow
		if err := rows.Scan(&g.id, &g.epoch, &g.open, &g.sev, &g.lik, &g.imp, &g.cx, &g.material,
			&g.closeEvent, &g.benchEvent, &g.closedByBench); err != nil {
			return r, err
		}
		r.gaps = append(r.gaps, g)
	}
	return r, rows.Err()
}

// gapStatesOf is the gap family, whole, from what familyReadsAt read: the view answers the derived
// facts (openness, board order, the regrade-overlaid grades, materiality, and which act closed the
// gap) so no reader re-decides them, and the stream hands back the acts a renderer quotes — the
// mint's prose, the closing acts' bodies, the regrade history. It asks the record nothing.
//
// It is the one assembly of the family (plans/board-as-views.md wave 3); the markdown renders
// and their oracle read it.
func gapStatesOf(r familyReads) ([]*Gap, error) {
	evs, win := gapFamilyActs(r.merged.Events), r.merged.At
	byID := eventsByID(evs, win)
	// THE HISTORY IS A LISTING: every regrade, a struck one marked, before the fold below narrows
	// the stream to the acts that stand.
	history := map[string][]RegradeEntry{}
	for _, l := range Listing(evs) {
		if r, ok := recordpb.BodyAs[*recordpb.Regrade](l.Event); ok {
			history[r.GetGapId()] = append(history[r.GetGapId()], RegradeEntry{Regrade: r, Struck: l.Struck})
		}
	}
	// The acts that stand: a corrected mint or regrade is folded as its replacement.
	mints := map[string]*recordpb.Mint{}
	regrades := map[string][]*recordpb.Regrade{}
	for _, e := range Live(evs) {
		switch m := mustBody(e).(type) {
		case *recordpb.Mint:
			if _, seen := mints[m.GetGapId()]; !seen {
				mints[m.GetGapId()] = m
			}
		case *recordpb.Regrade:
			regrades[m.GetGapId()] = append(regrades[m.GetGapId()], m)
		}
	}
	var out []*Gap
	for _, row := range r.gaps {
		g := &Gap{
			ID: row.id, Epoch: row.epoch, Open: row.open, Mint: mints[row.id],
			Regrades:       regrades[row.id],
			RegradeHistory: history[row.id],
			Severity:       gradeOrZero(row.sev),
			Likelihood:     gradeOrZero(row.lik),
			Impact:         gradeOrZero(row.imp),
			ComplexityCost: gradeOrZero(row.cx),
			Material:       row.material,
		}
		c, err := closeStateOf(row.id, row.closeEvent, row.benchEvent, row.closedByBench, byID, win)
		if err != nil {
			return nil, err
		}
		if c != nil {
			g.HasClosed, g.ClosedEpoch, g.ClosedByBench = true, c.closedEpoch, c.closedByBench
			g.Closure, g.BenchClosure = c.lastClose, c.lastBenchClosure
		}
		out = append(out, g)
	}
	return out, nil
}

// GapsByID indexes the family for the readers that follow lineage (an ancestor lookup is a map
// hit, never a second query).
func GapsByID(gaps []*Gap) map[string]*Gap {
	out := make(map[string]*Gap, len(gaps))
	for _, g := range gaps {
		out[g.ID] = g
	}
	return out
}

// gradeOrZero maps a view grade word back to the schema's value; NULL is the ungraded zero,
// exactly the fold's answer for an axis nothing set.
func gradeOrZero(v sql.NullString) recordpb.Grade {
	if !v.Valid {
		return recordpb.Grade_GRADE_UNSPECIFIED
	}
	g, _ := GradeOf(v.String)
	return g
}

// ObservationsOf builds the lens observations from the finding events, exactly as the fold
// appended them: seat, key, the type's word, the finding body.
func ObservationsOf(evs []*Event) []*Observation {
	var out []*Observation
	for _, e := range evs {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok {
			out = append(out, &Observation{
				SeatID: e.GetSeatId(), Key: e.GetKey(), Kind: recordpb.Word(e.GetType()), Finding: f,
			})
		}
	}
	return out
}

// Family is the record handed to a whole-record presenter — the gap family in board order and
// the typed event stream, nothing derived beyond the family itself. It is what the report, the
// verifiers and the operators' consumers take instead of a Board (plans/board-as-views.md waves
// 4-6): no fold builds it, and no board-wide state rides along to drift.
type Family struct {
	Gaps []*Gap
	// Events is the WHOLE stream, struck acts included: a listing shows a corrected act struck,
	// never hides it. A reader picking a winner or discharging a duty reads Live() instead.
	Events []*Event
	// At is where each of Events sits on the record, as the loader read it (WindowIndex). A family
	// built by hand has none, and a fold that asks it panics rather than read a zero.
	At WindowIndex
	// Struck is every same-sitting correction on the stream, indexed once.
	Struck StruckIndex
	byID   map[string]*Gap
}

// Live is the stream a fold reads: each struck act replaced in place by the act that stands.
func (f Family) Live() []*Event { return Live(f.Events) }

// Listing is the stream a listing renders: the acts that stand, each preceded by what it struck.
func (f Family) Listing() []Listed { return Listing(f.Events) }

// FamilyOf assembles the family from the record in ONE read transaction: in it, the whole record
// once and the `gap` view once (familyReadsAt). The closure is the view's, and the bodies it names
// are looked up by events.id in that same read (gapStatesOf).
func FamilyOf(run Run) (Family, error) {
	var r familyReads
	err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = familyReadsAt(q)
		return err
	})
	if err != nil {
		return Family{}, err
	}
	gaps, err := gapStatesOf(r)
	if err != nil {
		return Family{}, err
	}
	return NewFamily(gaps, r.merged), nil
}

// NewFamily indexes the family once, so a lineage lookup is a map hit for every consumer. The stream
// and its window index come from the one read that loaded them (m); a Merged built by hand has no
// index, and a fold that asks it panics rather than read a zero.
func NewFamily(gaps []*Gap, m Merged) Family {
	return Family{Gaps: gaps, Events: m.Events, At: m.At, Struck: StruckIndexOf(m.Events), byID: GapsByID(gaps)}
}

// gapFamilyActs narrows the whole stream to the acts the gap family reads — the mints and regrades
// it folds, the closes and docket rulings the gap view names and it looks up, and the corrections
// that strike a mint or a regrade — in record order.
func gapFamilyActs(evs []*Event) []*Event {
	var out []*Event
	for _, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_MINT, recordpb.EventType_EVENT_TYPE_REGRADE,
			recordpb.EventType_EVENT_TYPE_CLOSE, recordpb.EventType_EVENT_TYPE_MOTION_RULE,
			recordpb.EventType_EVENT_TYPE_CORRECTION:
			out = append(out, e)
		}
	}
	return out
}

// Gap is the family's index: the gap by id, or nil — the same answer b.Gaps[id] gave.
func (f Family) Gap(id string) *Gap { return f.byID[id] }
