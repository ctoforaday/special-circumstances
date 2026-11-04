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
// THE ROUND ENDS ON THE LENS'S RE-AUDIT (ruled by gblock 2026-10-31). One exchange on G is blue's
// answer followed by its minting lens's audit of that answer: the turn is blue's from the mint, from a
// remand ruling and from the end of the last exchange, and the exchange ends when a sitting of the
// lens engaged on G, opened with its turn come (lensTurnOn), ends. A sitting that closes having
// recorded nothing on G is a NULL TURN: silence is a turn taken, and the exchange still counts. Where
// blue's sitting leaves nothing to audit the lens's turn never comes, so the exchange ends with blue's
// sitting, and a gap blue never acts on reaches impasse with no lens readied on it.
//
// Movement is a STATE change: the gap's grade changed, its anchored text changed (a blue edit
// answering G with old != new, or a citation retire), blue answered it with a computation or a change
// of fate on the inquiry it is anchored on (ruled by gblock 2026-10-31), its lineage moved (a mint
// superseding G), or it closed. A closing argument, a motion, a spot-check, a regrade to the same
// grade or an edit that changed nothing are acts but not movement.
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
	// rounds is each counted exchange in order — where it began, where it ended, whether it moved the
	// gap and whether blue acted on it — so "was the gap at impasse when the bench ruled" and "has the
	// remand's exchange been had" are asked of the same fold that counts.
	rounds []exchangeRound
	// pending is the exchange blue's answer has opened and no lens sitting has ended: what its round
	// would be if the lens's next sitting ended it now.
	pending exchangeRound
	// lensTurn is whether the minting lens's turn on the gap has come (lensTurnOn), lensRead the place
	// that turn is read from: its last act or sitting on the gap, the mint, the latest remand ruling.
	lensTurn bool
	lensRead int64
	// strandedDue: a successor names the gap and no sitting of its minting lens engaged on it has
	// opened since. What it owes there is the close the supersession promised, whatever blue did.
	strandedDue bool
	// answeredRemand: blue has acted on the gap since the latest remand ruling on it (gapActsOf).
	answeredRemand bool
	// lensFirst: the gap is at its limits, or the audit the lens now owes would bring it there, and the
	// lens has not had that audit: the plan readies the lens alone, and any docket waits on its sitting.
	lensFirst bool
	// lensAudit: the lens's turn stands and its last re-audit was not already at impasse (auditSpent): a
	// docket for the open gaps' grade movement waits on it too.
	lensAudit bool
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

// exchangeRound is one counted exchange: the places it opened and ended, whether anything moved the
// gap between them, whether blue recorded an act on the gap there (gapActsOf) — the exchange a remand
// grants begins only in a round in which blue acted — and whether a lens's re-audit ended it.
type exchangeRound struct {
	start, end int64
	moved      bool
	acted      bool
	audited    bool
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
		x.lensFirst = lensGoesFirst(x, p)
		x.lensAudit = x.lensTurn && !auditSpent(x, p)
	}
	return out
}

// lensGoesFirst is THE LENS'S PREFERENCE IN THE ORDER (ruled by gblock 2026-10-31): a gap at its
// limits, or whose owed audit would bring it there, readies its minting lens alone, and the docket
// comes only after that lens's sitting. Blue is held off the gap for that epoch, because blue sits
// after the lens and its answer would be one the lens has not read when the bench is asked.
//
// ONE AUDIT PAST THE LIMIT, then the docket. Blue can still move the sentence under another gap's
// name, which gives the lens its turn every epoch (lensTurnOn), so a deferral resting on the turn
// alone would never file: the round its audit ended is what spends the preference (auditSpent). A
// stranded gap at its limits has the same one sitting, for its close.
func lensGoesFirst(x *GapExchanges, p Params) bool {
	switch {
	case auditSpent(x, p):
		return false
	case x.Impasse:
		return x.lensTurn || x.strandedDue
	case !x.lensTurn:
		return false
	}
	return impasseOf(append(slices.Clone(x.rounds), x.pending), x.remandSittings, p, math.MaxInt64)
}

// auditSpent reports whether the lens's last re-audit was already at impasse: every round before it
// put the gap there. A remand's own exchange is exempt, since a remand holds the gap at impasse until
// that exchange has been had, and that audit is the one the ruling orders.
func auditSpent(x *GapExchanges, p Params) bool {
	n := len(x.rounds)
	if n == 0 || !x.rounds[n-1].audited {
		return false
	}
	last := x.rounds[n-1]
	if len(x.remandSittings) > 0 {
		ruled := x.remandSittings[len(x.remandSittings)-1]
		had := false
		for _, r := range x.rounds[:n-1] {
			had = had || (r.start >= ruled && r.acted)
		}
		if ruled > last.start || !had {
			return false
		}
	}
	return impasseOf(x.rounds[:n-1], x.remandSittings, p, last.end)
}

// gapSittingsOf is each gap's party sittings for the dispatches engaging them, in place order: a
// party's sitting for a dispatch is sittingFor, the first it opened after the dispatch, ended where
// sittingCloser says — by facts about THAT seat alone, so no rule enforced at another seat's write
// path can decide what this read reports. Two rows answered by one sitting are that one sitting.
func gapSittingsOf(evs []*Event, ids []int64, win WindowIndex, when ReadWhen, minted map[string]string) map[string][]partySitting {
	dispatches, openings := dispatchLedger(evs, ids, win)
	closer := sittingCloserOf(evs, ids, win, openings, when)
	out := map[string][]partySitting{}
	for _, d := range dispatches {
		start, sat := sittingFor(openings[d.seat], d)
		if !sat {
			continue // engaged, never sat: no exchange yet, and rule 1 keeps readying it
		}
		_, end, closed := closer.bounds(d.seat, start) // unclosed: the window runs to the end of the record
		for _, g := range d.gaps {
			switch {
			case d.seat == minted[g]:
				out[g] = append(out[g], partySitting{start: start, end: end, red: true, open: !closed})
			case seatclass.RoleOfSeat(d.seat) == "blue":
				out[g] = append(out[g], partySitting{start: start, end: end, open: !closed})
			}
		}
	}
	for g, ss := range out {
		sort.Slice(ss, func(a, b int) bool { return ss[a].start < ss[b].start })
		out[g] = slices.Compact(ss)
	}
	return out
}

// exchangeFold counts each gap's exchanges and stalls, and records each counted exchange as a round;
// exchangesOf applies the run's terms and the remands to them. remands is the "remand" view's rows
// by gap (remandRulingsOf), read here for where each ruling hands the turn to blue.
//
// THE TURN IS BLUE'S WHEN A ROUND OPENS, and a round opens at the mint, at a remand ruling and at the
// end of the last round. Blue's sitting then either leaves the lens something to audit (lensTurnOn),
// and the round stays open until the lens's sitting ends it, or leaves nothing: the lens's turn does
// not come, and the round ends with blue's sitting, so a gap blue never answers stalls exchange by
// exchange. A minter that is no lens has no re-audit to wait on, and blue's sitting ends its round. A
// lens sitting with no turn come, which only a stranded gap's lens or an archived dispatch has, ends
// nothing.
func exchangeFold(evs []*Event, ids []int64, win WindowIndex, when ReadWhen, remands map[string][]remandRow) map[string]*GapExchanges {
	minted := map[string]string{}     // gap -> the lens that minted it
	mintedAt := map[string]int64{}    // gap -> the place of its mint
	namedAt := map[string]int64{}     // gap -> the place of the first mint superseding it
	grades := map[string][3]string{}  // gap -> current severity, likelihood, impact
	movement := map[string][]int64{}  // gap -> ids of movement events
	anchored := map[string][]string{} // inquiry -> the gaps minted about it
	fates := map[string]recordpb.InquiryStatus{}
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
			if b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
				anchored[b.GetAboutRef()] = append(anchored[b.GetAboutRef()], g)
			}
			for _, s := range b.GetSupersedes() {
				movement[s] = append(movement[s], id)
				if _, seen := namedAt[s]; !seen {
					namedAt[s] = id
				}
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
			// MOVEMENT IS A PARTY'S ACT. The bench's regrade sets the grade its ruling rules on
			// (benchgrade.go); counted here, a bench regrade before a remand would read as the
			// remand's exchange moving the gap and undo the one-remand limit.
			if moved && seatclass.RoleOfSeat(e.GetSeatId()) != "bench" {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.BlueEdit:
			if g := b.GetAnswers(); g != "" && b.GetOld() != b.GetNew() {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.Computation:
			if g := b.GetAnswers(); g != "" && seatclass.RoleOfSeat(e.GetSeatId()) == "blue" {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.Inquiry:
			// A FATE IS THE ANSWER TO A GAP ANCHORED ON THE INQUIRY, as an edit is to one anchored on
			// text: no edit can answer it, and a fate restated moves nothing.
			if seatclass.RoleOfSeat(e.GetSeatId()) == "blue" && b.GetStatus() != fates[b.GetInquiryId()] {
				for _, g := range anchored[b.GetInquiryId()] {
					movement[g] = append(movement[g], id)
				}
			}
			fates[b.GetInquiryId()] = b.GetStatus()
		case *recordpb.Retire:
			// A citation retire moves the report as an edit does: the citation leaves its sentence.
			if g := retireAnswers(e); g != "" {
				movement[g] = append(movement[g], id)
			}
		case *recordpb.Close:
			movement[b.GetGapId()] = append(movement[b.GetGapId()], id)
		}
	}

	sittings := gapSittingsOf(evs, ids, win, when, minted)
	turns := lensTurnActsOf(evs, ids, acts, sittings, minted)

	out := map[string]*GapExchanges{}
	for g := range minted {
		x := &GapExchanges{GapID: g}
		out[g] = x
		lens, moves := minted[g], movement[g]
		audits := seatclass.RoleOfSeat(lens) == "lens"
		movedIn := func(from, to int64) bool {
			return slices.ContainsFunc(moves, func(m int64) bool { return m > from && m <= to })
		}
		rulings := remandPlaces(remands[g])
		start, next := mintedAt[g], 0
		end := func(at int64, audited bool) {
			x.Exchanges++
			moved := movedIn(start, at)
			if moved {
				x.Stalled = 0
			} else {
				x.Stalled++
			}
			x.rounds = append(x.rounds, exchangeRound{start: start, end: at, moved: moved, acted: blueActed(acts[g], start, at), audited: audited})
			start = at
		}
		// A LENS SITTING THE RECORD CANNOT CLOSE IS UNMEASURED, so no blue sitting after it ends a
		// round: whether the lens is auditing the answer is what the record cannot yet say.
		unmeasured := false
		for _, s := range sittings[g] {
			for ; next < len(rulings) && rulings[next] < s.start; next++ {
				start = rulings[next] // a ruling hands the turn to blue: what it heard opens nothing
			}
			switch {
			case s.open:
				// NOT COUNTED AND NOT ZERO. A sitting the record cannot close is reported as itself;
				// folding it either way would state something the record does not hold.
				x.Unresolved++
				unmeasured = unmeasured || s.red
			case s.red:
				unmeasured = false
				if audits && lensTurnOn(turns[g], lens, mintedAt[g], latestBefore(rulings, s.start), s.start-1) {
					end(s.end, true)
				}
			case unmeasured:
			default:
				// A SITTING NO STOP CLOSED ENDS AT ITS SEAT'S NEXT OPENING, after the lens's next sitting
				// opened: the round holds blue's acts from where blue's sitting began, and what blue left
				// unread is read off where the lens stood when blue sat.
				start = min(start, s.start)
				floor := latestBefore(rulings, s.end)
				read := lensReadAt(turns[g], lens, mintedAt[g], floor, s.start)
				if !audits || !blueActed(turns[g], read, s.end) {
					end(s.end, false)
				}
			}
		}
		for ; next < len(rulings); next++ {
			start = rulings[next]
		}
		floor := latestBefore(rulings, math.MaxInt64)
		x.lensRead = lensReadAt(turns[g], lens, mintedAt[g], floor, math.MaxInt64)
		x.lensTurn = audits && lensTurnOn(turns[g], lens, mintedAt[g], floor, math.MaxInt64)
		x.answeredRemand = blueActed(acts[g], floor, math.MaxInt64)
		x.pending = exchangeRound{start: start, end: math.MaxInt64 - 1, moved: movedIn(start, math.MaxInt64), acted: blueActed(acts[g], start, math.MaxInt64), audited: true}
		if at, named := namedAt[g]; named && audits {
			x.strandedDue = !slices.ContainsFunc(sittings[g], func(s partySitting) bool { return s.red && !s.open && s.start > at })
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
