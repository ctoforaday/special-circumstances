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
// register after the dispatch), the window its acts fall in, whether the record can say it ENDED,
// and its side.
//
// ONLY FACTS ABOUT THIS SEAT CLOSE IT, AND THE EARLIER OF TWO DOES. A sitting ends at whichever
// comes first after it began:
//
//   - the seat's next register. A register is a seat's first act of a sitting, so the next one is
//     proof the sitting before it ended — written by that seat.
//   - the stop of the agent that sat it: a sitting_close whose agent_id is the one on THIS
//     sitting's register. The SubagentStop hook writes it when that agent returns, so it is the
//     harness's observation of this seat, not an act of another seat.
//
// The register alone closes a sitting only when the seat sits AGAIN, which put every exchange —
// and so every impasse — one epoch behind the sittings that made it; under an epoch limit landing
// on that epoch, the run ended at its ceiling with a material gap the bench was never asked to
// rule. The stop closes it when it ends.
//
// NOTHING ANOTHER SEAT WRITES BOUNDS IT. A party's sitting was once closed at the chair's next
// register, on the reasoning that the workflow comes back to the chair only after the parties sit.
// That made a READ depend on a rule enforced at ANOTHER seat's write path, and a warm chair — one
// whose later sittings resume the same session — registers once per run: on five archived runs
// every party sitting read as still open and no gap could reach the bench by impasse (#1002).
//
// A STOP THAT JOINS NO REGISTER CLOSES NOTHING. The hook records every typed subagent in a project
// whose run marker is live, a developer's own subagents included, and a register with no agent_id
// (a run the PreToolUse hook never reached) has nothing a stop can join. A headless seat run as a
// `claude -p` main session fires no SubagentStop at all, so its sittings close by register alone.
//
// A sitting with neither past it is UNRESOLVED, not complete and not absent. It may be in flight,
// or it may have ended with nobody recording anything after it, and the record holds no fact that
// tells the two apart. Counting it would let an in-flight sitting be scored as a null turn and
// stall a gap that is still being answered; reporting it as zero would make the miss read exactly
// like a gap nobody has disputed. So it is neither: it is its own answer.
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
	return exchangesOf(evs, ids, p), nil
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

func exchangesOf(evs []*Event, ids []int64, p Params) map[string]*GapExchanges {
	minted := map[string]string{}    // gap -> the lens that minted it
	grades := map[string][3]string{} // gap -> current severity, likelihood, impact
	movement := map[string][]int64{} // gap -> ids of movement events
	dispatches, registers := dispatchLedger(evs, ids)
	sittingAgent := map[int64]string{} // register id -> the agent_id it carries
	stops := map[string][]int64{}      // agent_id -> ids of its sitting_close events, ascending

	for i, e := range evs {
		id := ids[i]
		switch b := mustBody(e).(type) {
		case *recordpb.Register:
			if a := b.GetAgentId(); a != "" {
				sittingAgent[id] = a
			}
		case *recordpb.SittingClose:
			if a := b.GetAgentId(); a != "" {
				stops[a] = append(stops[a], id)
			}
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

	// A party's sitting for a dispatch: sittingFor, its first register after the dispatch, ended by
	// the earlier of THAT SEAT'S next register and the stop of the agent on THAT register
	// (partySitting). Only registers[d.seat] and that agent's stops are read here — no other seat's
	// acts bound this seat's sitting, so no rule enforced at another seat's write path can decide
	// what this read reports.
	sittings := map[string][]partySitting{}
	for _, d := range dispatches {
		start, sat := sittingFor(registers[d.seat], d)
		if !sat {
			continue // engaged, never sat: no exchange yet, and rule 1 keeps readying it
		}
		end, closed := firstAfter(registers[d.seat], start)
		if agent := sittingAgent[start]; agent != "" {
			if stop, stopped := firstAfter(stops[agent], start); stopped && (!closed || stop < end) {
				end, closed = stop, true
			}
		}
		if !closed {
			end = math.MaxInt64 // nothing about this seat closes it: the window runs to the end of the record
		}
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
