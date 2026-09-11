package record

import (
	"database/sql"
	"fmt"
	"strings"

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
		// REGISTER IS NOT READ FOR ITS BODY. The fold's Clock counts chair registers to place each
		// closure in its epoch, and a slice filtered to the body types would leave every closure
		// in epoch 0 — silently, which is how the parity test found it.
		recordpb.EventType_EVENT_TYPE_REGISTER,
		recordpb.EventType_EVENT_TYPE_MINT,
		recordpb.EventType_EVENT_TYPE_REGRADE,
		recordpb.EventType_EVENT_TYPE_CLOSE,
		recordpb.EventType_EVENT_TYPE_MOTION,
		recordpb.EventType_EVENT_TYPE_MOTION_RULE)
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
	closures, unpairedDocket := closureStatesOf(evs)
	if len(unpairedDocket) > 0 {
		return nil, fmt.Errorf("record: docket ruling(s) on motion(s) %s have no filing on this record — the gap each settles rides its FILING, so an unpaired ruling would leave a disposed gap reading as open", strings.Join(unpairedDocket, ", "))
	}

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "gap_id", "minted_epoch", "open",
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
			ID: id, Epoch: round, Open: open, Mint: mints[id],
			Regrades:       regrades[id],
			Severity:       gradeOrZero(sev),
			Likelihood:     gradeOrZero(lik),
			Impact:         gradeOrZero(imp),
			ComplexityCost: gradeOrZero(cx),
		}
		if c := closures[id]; c != nil && c.hasClosed {
			g.HasClosed, g.ClosedEpoch, g.ClosedByBench = true, c.closedEpoch, c.closedByBench
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

// Family is the record handed to a whole-record presenter — the gap family in board order and
// the typed event stream, nothing derived beyond the family itself. It is what the report, the
// verifiers and the operators' consumers take instead of a Board (plans/board-as-views.md waves
// 4-6): no fold builds it, and no board-wide state rides along to drift.
type Family struct {
	Gaps []*Gap
	// Events is the WHOLE stream, struck acts included: a listing shows a corrected act struck,
	// never hides it. A reader picking a winner or discharging a duty reads Live() instead.
	Events []*Event
	// Struck is every same-sitting correction on the stream, indexed once.
	Struck StruckIndex
	byID   map[string]*Gap
}

// Live is the stream a fold reads: each struck act replaced in place by the act that stands.
func (f Family) Live() []*Event { return Live(f.Events) }

// Listing is the stream a listing renders: the acts that stand, each preceded by what it struck.
func (f Family) Listing() []Listed { return Listing(f.Events) }

// FamilyOf assembles the family from the record: the gap view + typed loader (GapStates) and
// the stream.
func FamilyOf(run Run) (Family, error) {
	gaps, err := GapStates(run)
	if err != nil {
		return Family{}, err
	}
	m, err := MergedEvents(run)
	if err != nil {
		return Family{}, err
	}
	return NewFamily(gaps, m.Events), nil
}

// NewFamily indexes the family once, so a lineage lookup is a map hit for every consumer.
func NewFamily(gaps []*Gap, evs []*Event) Family {
	return Family{Gaps: gaps, Events: evs, Struck: StruckIndexOf(evs), byID: GapsByID(gaps)}
}

// Gap is the family's index: the gap by id, or nil — the same answer b.Gaps[id] gave.
func (f Family) Gap(id string) *Gap { return f.byID[id] }
