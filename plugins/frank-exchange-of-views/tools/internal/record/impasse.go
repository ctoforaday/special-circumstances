package record

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// GapExchanges is what the record says about one gap's dispute (plans/roundless.md §III.B.2).
//
// The exchange unit is defined over SITTINGS, not acts: one exchange on G is a red-party sitting
// engaged on G followed by a blue-party sitting engaged on G, both complete. A sitting that closes
// having recorded nothing on G is a NULL TURN — silence is a turn taken — and the exchange still
// counts. Movement is a STATE change: the gap's grade changed, its anchored text changed (a blue
// edit answering G with old != new), its lineage moved (a mint superseding G), or it closed. A
// closing argument, a motion, a spot-check, a regrade to the same grade or an edit that changed
// nothing are acts but not movement.
//
// THE COUNTS ARE A FLOOR AND UNRESOLVED IS THE REST OF THE ANSWER. Exchanges and Stalled state
// what the record can establish; Unresolved states what it cannot, so a reader is never handed a
// zero that might mean either.
type GapExchanges struct {
	GapID     string
	Exchanges int  // total exchanges on G — monotone, never resets
	Stalled   int  // consecutive exchanges with no movement — resets on movement
	Impasse   bool // Stalled >= K or Exchanges >= KMax
	// Unresolved is the sittings engaged on G that the record cannot close (see partySitting):
	// the seat sat, and neither its next register nor its agent's stop is on the record since.
	// They are NOT counted as exchanges and NOT reported as zero — the miss and the honest zero are
	// different answers, and Counted words them apart.
	Unresolved int
}

// Counted is the fold's numbers in the words a seat reads, and it is THE ONE PLACE the not-measured
// case is worded — a consumer that formats the fields itself can fold the miss back into a zero,
// which is the whole defect (#1002), so every consumer renders through here.
func (x *GapExchanges) Counted() string {
	switch {
	case x.Unresolved == 0:
		return fmt.Sprintf("%d exchange(s) (%d stalled)", x.Exchanges, x.Stalled)
	case x.Exchanges == 0:
		return fmt.Sprintf("exchanges NOT MEASURED — %d sitting(s) the record cannot close (the seat has not registered since and its agent's stop is not on the record), so this is not a count of zero", x.Unresolved)
	default:
		return fmt.Sprintf("%d exchange(s) (%d stalled), %d sitting(s) NOT MEASURED — the record cannot close them", x.Exchanges, x.Stalled, x.Unresolved)
	}
}

// partySitting is one seat's sitting engaged on a gap by a dispatch: when it began (the seat's
// register after the dispatch, sittingFor), the window its acts fall in, whether the record can
// say it ENDED (sittingCloser, the one rule every sitting-bounding read shares), and its side.
//
// The register alone closed a sitting only when the seat sat AGAIN, which put every exchange —
// and so every impasse — one epoch behind the sittings that made it; under an epoch limit landing
// on that epoch, the run ended at its ceiling with a material gap the bench was never asked to
// rule. The agent's stop closes it when it ends.
//
// AN UNRESOLVED SITTING IS NOT COUNTED AND NOT ZERO. Counting it would let an in-flight sitting be
// scored as a null turn and stall a gap that is still being answered; reporting it as zero would
// make the miss read exactly like a gap nobody has disputed. So it is neither: it is its own answer.
type partySitting struct {
	start, end int64
	red        bool
	// open is the sitting the record cannot close: end is the end of the record rather than an
	// act, so the acts in the window are everything written since — a floor, not a sitting.
	open bool
}

// Exchanges folds the record into per-gap exchange counts under the run's terms. It reads the
// dispatch events for who was engaged on what, the registers for when each party actually sat,
// and the acts for movement — nothing is asserted by a seat; the counts are the record's.
func Exchanges(run Run, p Params) (map[string]*GapExchanges, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return map[string]*GapExchanges{}, err
	}
	evs, _, err := recordsql.EventsW(db)
	if err != nil {
		return nil, err
	}
	ids, err := eventIDs(db)
	if err != nil {
		return nil, err
	}
	return exchangesOf(evs, ids, p, WhileRunning), nil
}

// eventIDs is events."id" per position — the sequence the folds compare against, which the proto
// Event does not carry (see recordsql.Window for why nothing derived is stamped on the row).
// eventIDsOfRun is eventIDs for a caller holding the run: the events' row ids in stream order,
// aligned with MergedEvents, or nil for a run with no record yet.
func eventIDsOfRun(run Run) ([]int64, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	return eventIDs(db)
}

func eventIDs(db *sql.DB) ([]int64, error) {
	rows, err := db.Query(`SELECT "id" FROM "events" ORDER BY "id"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// exchangesOf is the fold. It takes when the record is read from its caller: the chair's dispatch
// plan is its reader, and a plan is computed while the run is running.
func exchangesOf(evs []*Event, ids []int64, p Params, when ReadWhen) map[string]*GapExchanges {
	minted := map[string]string{}    // gap -> the lens that minted it
	grades := map[string][3]string{} // gap -> current severity, likelihood, impact
	movement := map[string][]int64{} // gap -> ids of movement events
	dispatches, registers := dispatchLedger(evs, ids)
	closer := sittingCloserOf(evs, ids, registers, when)

	for i, e := range evs {
		id := ids[i]
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			g := b.GetGapId()
			if _, seen := minted[g]; !seen {
				minted[g] = e.GetSeatId()
				grades[g] = [3]string{recordpb.Word(b.GetSeverity()), recordpb.Word(b.GetLikelihood()), recordpb.Word(b.GetImpact())}
			}
			for _, s := range b.GetSupersedes() {
				movement[s] = append(movement[s], id)
			}
		case *recordpb.Regrade:
			g := b.GetGapId()
			cur := grades[g]
			moved := false
			for k, w := range []string{recordpb.Word(b.GetSeverity()), recordpb.Word(b.GetLikelihood()), recordpb.Word(b.GetImpact())} {
				if w != "" && w != cur[k] {
					cur[k], moved = w, true
				}
			}
			grades[g] = cur
			if moved {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.BlueEdit:
			if g := b.GetAnswers(); g != "" && b.GetOld() != b.GetNew() {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.Close:
			movement[b.GetGapId()] = append(movement[b.GetGapId()], id)
		}
	}

	// A party's sitting for a dispatch: sittingFor, its first register after the dispatch, ended
	// where sittingCloser says — by facts about THAT seat alone, so no rule enforced at another
	// seat's write path can decide what this read reports.
	sittings := map[string][]partySitting{}
	for _, d := range dispatches {
		start, sat := sittingFor(registers[d.seat], d)
		if !sat {
			continue // engaged, never sat: no exchange yet, and rule 1 keeps readying it
		}
		_, end, closed := closer.bounds(d.seat, start) // unclosed: the window runs to the end of the record
		for _, g := range d.gaps {
			switch {
			case d.seat == minted[g]:
				sittings[g] = append(sittings[g], partySitting{start: start, end: end, red: true, open: !closed})
			case roleOfSeat(d.seat) == "blue":
				sittings[g] = append(sittings[g], partySitting{start: start, end: end, open: !closed})
			}
		}
	}

	out := map[string]*GapExchanges{}
	for g := range minted {
		x := &GapExchanges{GapID: g}
		out[g] = x
		ss := sittings[g]
		sort.Slice(ss, func(a, b int) bool { return ss[a].start < ss[b].start })
		moves := movement[g]
		pendingRed := int64(math.MinInt64)
		for _, s := range ss {
			if s.open {
				// NOT COUNTED AND NOT ZERO. A sitting the record cannot close is reported as
				// itself; folding it either way would state something the record does not hold.
				x.Unresolved++
				continue
			}
			if s.red {
				pendingRed = s.start
				continue
			}
			if pendingRed == math.MinInt64 {
				continue // blue answering nobody is not an exchange
			}
			x.Exchanges++
			moved := false
			for _, m := range moves {
				if m > pendingRed && m <= s.end {
					moved = true
					break
				}
			}
			if moved {
				x.Stalled = 0
			} else {
				x.Stalled++
			}
			pendingRed = math.MinInt64
		}
		x.Impasse = x.Stalled >= p.K || x.Exchanges >= p.KMax
	}
	return out
}
