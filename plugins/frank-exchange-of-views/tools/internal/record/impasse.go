package record

import (
	"fmt"
	"math"
	"slices"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
	Exchanges int // total exchanges on G — monotone, never resets
	Stalled   int // consecutive exchanges with no movement — resets on movement
	// Impasse is the run's terms applied to the exchanges (impasseOf): Stalled >= K or Exchanges
	// >= KMax, and after the bench remands it at impasse, the terms counted from that ruling (impasseOf).
	Impasse bool
	// Unresolved is the sittings engaged on G that the record cannot close (see partySitting):
	// the seat sat, and neither its next sitting nor its agent's stop is on the record since.
	// They are NOT counted as exchanges and NOT reported as zero — the miss and the honest zero are
	// different answers, and Counted words them apart.
	Unresolved int
	// Remand is where the gap stands against the bench's remands (remandStage): the dispatch plan,
	// the work lists and the chair's row read it from here, so no reader re-derives it.
	Remand remandStage
	// rounds is each counted exchange in order — where it began (its minting lens's sitting), where
	// it ended (blue's), whether it moved the gap and whether blue acted on it — so "was the gap at
	// impasse when the bench ruled" and "has the remand's exchange been had" are asked of the same
	// fold that counts.
	rounds []exchangeRound
	// remandSittings is the place of each bench sitting the fold COUNTED as a remand: one that remanded
	// the gap while it was at impasse, keyed on the latest remand ruling in that sitting.
	remandSittings []int64
	// latestRemandOwed is the latest remand the bench ruled on the gap, counted or not, with no
	// exchange since its ruling in which blue acted on the gap (latestRemandOwed): what routes a
	// remand of a gap that is not material (routeOf).
	latestRemandOwed bool
	// latestRemand is the place of that ruling (the latest of remandSittingsOf), 0 where the bench
	// never remanded the gap: each party's act on the gap after it is what clears the remand's item
	// on that party's work list (remandItemsOf).
	latestRemand int64
}

// exchangeRound is one counted exchange: the places its red and blue sittings began and ended,
// whether anything moved the gap between them, and whether blue recorded an act on the gap there
// (gapActsOf) — the exchange a remand grants begins only in a round in which blue acted.
type exchangeRound struct {
	start, end int64
	moved      bool
	acted      bool
}

// Counted is the fold's numbers in the words a seat reads, and it is THE ONE PLACE the not-measured
// case is worded — a consumer that formats the fields itself can fold the miss back into a zero,
// which is the whole defect (#1002), so every consumer renders through here.
func (x *GapExchanges) Counted() string {
	switch {
	case x.Unresolved == 0:
		return fmt.Sprintf("%d exchange(s) (%d stalled)", x.Exchanges, x.Stalled)
	case x.Exchanges == 0:
		return fmt.Sprintf("exchanges NOT MEASURED — %d sitting(s) the record cannot close (the seat has not sat again since and its agent's stop is not on the record), so this is not a count of zero", x.Unresolved)
	default:
		return fmt.Sprintf("%d exchange(s) (%d stalled), %d sitting(s) NOT MEASURED — the record cannot close them", x.Exchanges, x.Stalled, x.Unresolved)
	}
}

// partySitting is one seat's sitting engaged on a gap by a dispatch: when it began (the first
// sitting the seat opened after the dispatch, by a hook's bracket or a register — sittingFor), the
// window its acts fall in, whether the record can say it ENDED (sittingCloser, the one rule every
// sitting-bounding read shares), and its side.
//
// A sitting closed by the seat's next opening alone would close only when the seat sat AGAIN,
// putting every exchange — and so every impasse — one epoch behind the sittings that made it; under
// an epoch limit landing on that epoch, the run would end at its ceiling with a material gap the
// bench was never asked to rule. The agent's stop closes it when it ends.
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
// dispatch events for who was engaged on what, the sittings each party opened (a hook's bracket or
// a register) for when it actually sat, and the acts for movement — nothing is asserted by a seat;
// the counts are the record's.
//
// The stream and the remands are read on one snapshot (readSnapshot) and folded after it closes: a
// remand ruled between two reads would count against a stream that does not hold its ruling.
func Exchanges(run Run, p Params) (map[string]*GapExchanges, error) {
	var evs []*Event
	var win WindowIndex
	var remands map[string][]remandRow
	absent := false
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		if absent = noRecord(q); absent {
			return nil
		}
		var err error
		if evs, win, err = eventsAt(q); err != nil {
			return err
		}
		remands, err = remandRulingsOf(q)
		return err
	}); err != nil {
		return nil, err
	}
	if absent {
		return map[string]*GapExchanges{}, nil
	}
	return exchangesOf(evs, win.IDs(evs), win, p, WhileRunning, remands), nil
}

// exchangesOf is the fold under the run's terms and the bench's remands. It takes when the record
// is read from its caller: the chair's dispatch plan is its reader, and a plan is computed while the
// run is running. ids is evs's places; remands is the "remand" view's rows by gap (remandRulingsOf).
func exchangesOf(evs []*Event, ids []int64, win WindowIndex, p Params, when ReadWhen, remands map[string][]remandRow) map[string]*GapExchanges {
	out := exchangeFold(evs, ids, win, when, remands)
	for g, x := range out {
		ruled := remandSittingsOf(remands[g])
		x.remandSittings = countedRemands(x.rounds, ruled, p)
		x.latestRemandOwed = latestRemandOwed(x.rounds, ruled)
		if len(ruled) > 0 {
			x.latestRemand = ruled[len(ruled)-1]
		}
		x.Impasse = impasseOf(x.rounds, x.remandSittings, p, math.MaxInt64)
		x.Remand = remandStageOf(x)
	}
	return out
}

// exchangeFold counts each gap's exchanges and stalls, and records each counted exchange as a round;
// exchangesOf applies the run's terms and the remands to them. remands is the "remand" view's rows
// by gap (remandRulingsOf), read here for where each ruling hands the turn to blue.
//
// A ROUND OPENS ON THE LENS'S TURN, and the lens's turn is not only a sitting it was engaged in
// (lensTurnOn, ruled by gblock 2026-10-30):
//
//   - the MINT is the minting lens's opening turn, so blue's first sitting on a fresh gap completes
//     the gap's first exchange — the plan readies blue alone on it, and the lens sits after blue;
//   - a REMAND RULING hands the turn to blue: the bench heard both parties, so blue's next sitting
//     completes the exchange the remand grants, beginning at the ruling (impasseOf);
//   - a lens's turn that has not come is PASSED: a counted round whose end leaves blue with no act
//     on the gap since the lens's last act there readies the lens on nothing new, so the plan does
//     not engage it, and the round blue's next sitting completes opens where this one ended. A gap
//     blue never answers stalls exchange by exchange and reaches impasse as it did when the lens sat
//     a null turn on it each epoch.
func exchangeFold(evs []*Event, ids []int64, win WindowIndex, when ReadWhen, remands map[string][]remandRow) map[string]*GapExchanges {
	minted := map[string]string{}    // gap -> the lens that minted it
	mintedAt := map[string]int64{}   // gap -> the place of its mint
	grades := map[string][3]string{} // gap -> current severity, likelihood, impact
	movement := map[string][]int64{} // gap -> ids of movement events
	dispatches, openings := dispatchLedger(evs, ids, win)
	closer := sittingCloserOf(evs, ids, win, openings, when)
	acts := gapActsOf(evs, ids)

	for i, e := range evs {
		id := ids[i]
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			g := b.GetGapId()
			if _, seen := minted[g]; !seen {
				minted[g] = e.GetSeatId()
				mintedAt[g] = id
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
		case *recordpb.Retire:
			// A citation retire moves the report as an edit does: the citation leaves its sentence.
			if g := retireAnswers(e); g != "" {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.Close:
			movement[b.GetGapId()] = append(movement[b.GetGapId()], id)
		}
	}

	// A party's sitting for a dispatch: sittingFor, the first it opened after the dispatch, ended
	// where sittingCloser says — by facts about THAT seat alone, so no rule enforced at another
	// seat's write path can decide what this read reports.
	sittings := map[string][]partySitting{}
	for _, d := range dispatches {
		start, sat := sittingFor(openings[d.seat], d)
		if !sat {
			continue // engaged, never sat: no exchange yet, and rule 1 keeps readying it
		}
		_, end, closed := closer.bounds(d.seat, start) // unclosed: the window runs to the end of the record
		for _, g := range d.gaps {
			switch {
			case d.seat == minted[g]:
				sittings[g] = append(sittings[g], partySitting{start: start, end: end, red: true, open: !closed})
			case seatclass.RoleOfSeat(d.seat) == "blue":
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
		// handed is each place that hands the turn to blue outside a sitting: the mint, then each
		// remand ruling, in place order; next is the first not yet passed.
		rulings := remandPlaces(remands[g])
		handed, next := append([]int64{mintedAt[g]}, rulings...), 0
		for _, s := range ss {
			for ; next < len(handed) && handed[next] < s.start; next++ {
				pendingRed = handed[next]
			}
			if s.open {
				// NOT COUNTED AND NOT ZERO. A sitting the record cannot close is reported as
				// itself; folding it either way would state something the record does not hold. The
				// lens's turn it is taking is unmeasured, so no turn handed before it opens a round
				// the next blue sitting could complete.
				x.Unresolved++
				if s.red {
					pendingRed = math.MinInt64
				}
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
			x.rounds = append(x.rounds, exchangeRound{start: pendingRed, end: s.end, moved: moved, acted: blueActed(acts[g], pendingRed, s.end)})
			pendingRed = math.MinInt64
			if !lensTurnOn(acts[g], minted[g], mintedAt[g], latestBefore(rulings, s.end), s.end) {
				pendingRed = s.end // the lens's turn has not come, so it passes: blue's next sitting completes a round
			}
		}
	}
	return out
}

// remandPlaces is the place of each of a gap's remand rulings, in record order.
func remandPlaces(rows []remandRow) []int64 {
	out := make([]int64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.eventID)
	}
	slices.Sort(out)
	return out
}

// latestBefore is the latest of places (sorted) at or before upto, 0 when none is.
func latestBefore(places []int64, upto int64) int64 {
	var at int64
	for _, p := range places {
		if p <= upto {
			at = p
		}
	}
	return at
}
