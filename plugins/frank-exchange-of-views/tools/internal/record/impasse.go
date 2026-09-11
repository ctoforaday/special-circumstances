package record

import (
	"database/sql"
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
type GapExchanges struct {
	GapID     string
	Exchanges int  // total exchanges on G — monotone, never resets
	Stalled   int  // consecutive exchanges with no movement — resets on movement
	Impasse   bool // Stalled >= K or Exchanges >= KMax
}

// partySitting is one seat's sitting engaged on a gap by a dispatch: when it began (the seat's
// register after the dispatch), when it ended (the seat's next register, or the chair's next
// register — the workflow came back to the chair, so the party had returned), and its side.
type partySitting struct {
	start, end int64
	red        bool
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
	return exchangesOf(evs, ids, p), nil
}

// eventIDs is events."id" per position — the sequence the folds compare against, which the proto
// Event does not carry (see recordsql.Window for why nothing derived is stamped on the row).
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

func exchangesOf(evs []*Event, ids []int64, p Params) map[string]*GapExchanges {
	const chair = "red-chair"
	minted := map[string]string{}    // gap -> the lens that minted it
	grades := map[string][3]string{} // gap -> current severity, likelihood, impact
	movement := map[string][]int64{} // gap -> ids of movement events
	dispatches, registers := dispatchLedger(evs, ids)
	chairRegisters := registers[chair]

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

	// A party's sitting for a dispatch: sittingFor, its first register after the dispatch. It is
	// COMPLETE once the chair has registered again after it began — the workflow awaits the parties
	// before it comes back to the chair — and it ends at the seat's next register or that chair
	// register, whichever is first.
	sittings := map[string][]partySitting{}
	for _, d := range dispatches {
		start, sat := sittingFor(registers[d.seat], d)
		if !sat {
			continue // engaged, never sat: no exchange yet, and rule 1 keeps readying it
		}
		chairNext, back := firstAfter(chairRegisters, start)
		if !back {
			continue // the sitting is still open
		}
		end := chairNext
		if next, ok := firstAfter(registers[d.seat], start); ok && next < end {
			end = next
		}
		for _, g := range d.gaps {
			switch {
			case d.seat == minted[g]:
				sittings[g] = append(sittings[g], partySitting{start: start, end: end, red: true})
			case roleOfSeat(d.seat) == "blue":
				sittings[g] = append(sittings[g], partySitting{start: start, end: end})
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
