package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// GapStates is the gap family, whole, asked of the record: the view answers the derived facts
// (openness, board order, the regrade-overlaid grades) so no reader re-decides them, and the
// typed loader hands back the acts a renderer quotes — the mint's prose, the closure pair, the
// regrade history. It fills the same *Gap the fold built, because the renders read that shape;
// what it does NOT build is a Board: no event bag, no counts, no board-wide derived state.
//
// It is the one assembly of the family (plans/board-as-views.md wave 3); the markdown renders
// and their oracle read it in place of BoardState.
func GapStates(run Run) ([]*Gap, error) {
	evs, err := EventsOf(run,
		recordpb.EventType_EVENT_TYPE_MINT,
		recordpb.EventType_EVENT_TYPE_REGRADE,
		recordpb.EventType_EVENT_TYPE_CLOSE,
		recordpb.EventType_EVENT_TYPE_OPINION)
	if err != nil {
		return nil, err
	}
	mints := map[string]*recordpb.Mint{}
	regrades := map[string][]*recordpb.Regrade{}
	for _, e := range evs {
		switch m := mustBody(e).(type) {
		case *recordpb.Mint:
			if _, seen := mints[m.GetGapId()]; !seen {
				mints[m.GetGapId()] = m
			}
		case *recordpb.Regrade:
			regrades[m.GetGapId()] = append(regrades[m.GetGapId()], m)
		}
	}
	closures := closureStatesOf(evs)

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "gap_id", "minted_round", "open",
	    "current_severity", "current_likelihood", "current_impact", "current_complexity_cost"
	  FROM "gap" ORDER BY "minted_event"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its gap states: %w", err)
	}
	defer rows.Close()
	var out []*Gap
	for rows.Next() {
		var id string
		var round int
		var open bool
		var sev, lik, imp, cx sql.NullString
		if err := rows.Scan(&id, &round, &open, &sev, &lik, &imp, &cx); err != nil {
			return nil, err
		}
		g := &Gap{
			ID: id, Round: round, Open: open, Mint: mints[id],
			Regrades:       regrades[id],
			Severity:       gradeOrZero(sev),
			Likelihood:     gradeOrZero(lik),
			Impact:         gradeOrZero(imp),
			ComplexityCost: gradeOrZero(cx),
		}
		if c := closures[id]; c != nil && c.hasClosed {
			g.HasClosed, g.ClosedRound, g.ClosedByBench = true, c.closedRound, c.closedByBench
			g.Closure, g.BenchClosure = c.lastClose, c.lastBenchClosure
		}
		out = append(out, g)
	}
	return out, rows.Err()
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
