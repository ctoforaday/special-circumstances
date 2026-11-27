package record

import (
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A CLOSING IS A PARTY'S ARGUMENT TO THE BENCH ON ONE DOCKETED GAP, AND THIS FILE IS WHERE WHO OWES
// ONE IS SAID (gblock, 2026-10-20).
//
// A gap's dispute has two parties: its assignee, which holds it from mint to close, and
// blue's responding seat, which answers it. Once the gap is docketed — the dispatch writes a docketing
// at impasse, for its grade movement, or for a proposal its assignee did not accept — the bench rules it
// on the two parties' closings, the transcript and the artifacts as they stand, so each party owes one
// closing for each docketing. The chair is party to no dispute and files none. A lens's regrade that keeps
// its grade is its move while the dispute runs; the closing is its brief once the gap is before the
// bench.
//
// ONE READING, AND EVERY SURFACE THAT ENFORCES, LISTS OR AUDITS THE DUTY CALLS IT: the PASS gate's
// list (closing_owed, passBlockersOf) — through which the plan readies the party that owes and
// convenes the bench only once no party is awaited, the work list blocks the party, and the lens
// stays unretired — and capture's record-parity audit.
//
// A CLOSING ANSWERS ONE DOCKETING: a closing of the party's on the gap filed after the docketing and
// before its ruling. A PROPOSAL docketing's span opens at the proposal's filing (L30), so the closing
// its assignee files in its window sitting answers it. The write path refuses a closing on a gap no
// docketing and no proposal stands on, so one filed for an earlier docketing, before a remand, answers
// that docketing and no later one.
//
// OWED BY EACH PARTY, WITH TWO EXCEPTIONS, each a state of its own (ClosingNotOwed):
//   - the gap closed after the docketing: its raiser's concession, which ends the docketing with no
//     ruling (requireRaiserActOnGap), so no closing is owed on it;
//   - the gap's minter is no lens the cast seats — an archived run's merge mints translate to the
//     chair — so there is no minting party to dispatch, and blue's closing is the only one owed.
//
// AWAITED AND MISSING ARE APART BY WHETHER THE PARTY HAS SAT SINCE THE DOCKETING. A party that has
// sat through no sitting opened after the docketing's span opened is awaited: the plan readies it, and the
// bench waits. One that sat through such a sitting and filed none is missing: the plan readies it no
// more for this docketing (once per cause, as for every blocker another seat owes), the bench sits
// on the record as it stands, the item stays on the party's work list until the bench rules, and
// capture fails it.

// ClosingState is what the record says of one party's closing on one docketing.
type ClosingState string

const (
	// ClosingNotOwed: the gap closed after the docketing, or the party is a minter the cast seats as
	// no lens. NotOwedWhy says which.
	ClosingNotOwed ClosingState = "not_owed"
	// ClosingFiled: the party filed a closing on the gap after the docketing and before its ruling.
	ClosingFiled ClosingState = "filed"
	// ClosingAwaited: none filed, and the party has sat through no sitting opened since the docketing
	// (before the ruling, where there is one). Read after the run it never sat after the docketing:
	// nothing it could have filed was refused it, and nothing is short.
	ClosingAwaited ClosingState = "awaited"
	// ClosingMissing: none filed, and the party sat through a sitting opened after the docketing —
	// before the ruling, where there is one — and filed none.
	ClosingMissing ClosingState = "missing"
)

// ClosingDuty is one party's closing on one docketing.
type ClosingDuty struct {
	GapID string
	// Docketing is the docketing's id; Ruled whether the bench has ruled it. Proposal is the proposal
	// in its assignee's window a row of no docketing yet is owed for ("" on a docketing's row).
	Docketing string
	Proposal  string
	Ruled     bool
	Party     string
	State     ClosingState
	// NotOwedWhy is the exception that holds, for ClosingNotOwed; "" otherwise.
	NotOwedWhy string
	// filed is where the docketing's span opens — its row, or for a PROPOSAL docketing the proposal's
	// filing — and ruled its ruling's place, or the place of the Close that ended it unruled (0 while
	// it stands).
	filed, ruled int64
}

// Why a closing is not owed.
const (
	closingGapClosed    = "the gap closed after it was docketed, so the dispute it would argue is over"
	closingMinterNoLens = "the gap's minter is no lens the cast seats, so there is no minting party"
)

// closingPartiesOf is the raiser a gap's docketing owes a closing from: its minter where the cast seats
// it as a lens. ok is false for a minter that is not one. The other party is the subject's author
// (authorOf).
func closingPartiesOf(cast []string, assignee string) (lens string, ok bool) {
	if seatclass.RoleOfSeat(assignee) == "lens" && slices.Contains(cast, assignee) {
		return assignee, true
	}
	return "", false
}

// IsClosingParty reports whether seatID is a party to the dispute on a gap its minter assignee
// minted, about a subject author wrote (authorOf): the minter itself, or the author — the seats the
// write path admits a closing from.
func IsClosingParty(seatID, assignee, author string) bool {
	return seatID == author || (seatID == assignee && seatclass.RoleOfSeat(seatID) == "lens")
}

// ClosingDuties is one row per party per docketing on the record, in the docketings' order, read off
// the whole stream: the reading capture audits after the run.
func ClosingDuties(evs []*Event, win WindowIndex) []ClosingDuty {
	return closingDutiesOf(evs, win.IDs(evs), win)
}

// closingReads is what every closing duty is read against: the cast, each gap's assignee and its
// subject's author, where each gap first closed, the closings filed, and the sittings each seat sat
// through.
type closingReads struct {
	cast      []string
	assignees map[string]string
	author    func(gap string) string
	closedAt  map[string]int64
	closings  []closingFiled
	through   map[string][]int64
}

type closingFiled struct {
	seat, gap string
	at        int64
}

// closingReadsOf reads closingReads off evs and its places ids.
func closingReadsOf(evs []*Event, ids []int64, win WindowIndex) closingReads {
	r := closingReads{cast: castOfEvents(evs), assignees: assigneeOf(evs), author: authorsOn(evs), closedAt: map[string]int64{}}
	for i, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Close:
			if _, seen := r.closedAt[b.GetGapId()]; !seen {
				r.closedAt[b.GetGapId()] = ids[i]
			}
		case *recordpb.Closing:
			r.closings = append(r.closings, closingFiled{seat: e.GetSeatId(), gap: b.GetGapId(), at: ids[i]})
		}
	}
	_, openings := dispatchLedger(evs, ids, win)
	r.through = satThroughOf(win, openings).through
	return r
}

// dutiesOn is the closing rows of one span on gapID — from its opening place to end (0 while it
// stands) — for the docketing id (the gap's assignee's row where the cast seats it, then its subject's
// author's).
func (r closingReads) dutiesOn(gapID, docketing string, from, end int64, ruled bool) []ClosingDuty {
	within := func(at int64) bool { return at > from && (end == 0 || at < end) }
	row := func(party string) ClosingDuty {
		d := ClosingDuty{GapID: gapID, Docketing: docketing, Ruled: ruled, Party: party, filed: from, ruled: end}
		for _, c := range r.closings {
			if c.seat == party && c.gap == gapID && within(c.at) {
				d.State = ClosingFiled
				return d
			}
		}
		// The Close that ended the docketing unruled is the span's end, and owes nothing after it.
		if at, closed := r.closedAt[gapID]; closed && at > from && (end == 0 || at <= end) {
			d.State, d.NotOwedWhy = ClosingNotOwed, closingGapClosed
			return d
		}
		d.State = ClosingAwaited
		if o, sat := firstAfter(r.through[party], from); sat && (end == 0 || o < end) {
			d.State = ClosingMissing
		}
		return d
	}
	var out []ClosingDuty
	if lens, ok := closingPartiesOf(r.cast, r.assignees[gapID]); ok {
		out = append(out, row(lens))
	} else {
		out = append(out, ClosingDuty{GapID: gapID, Docketing: docketing, Ruled: ruled, Party: r.assignees[gapID],
			State: ClosingNotOwed, NotOwedWhy: closingMinterNoLens, filed: from, ruled: end})
	}
	return append(out, row(r.author(gapID)))
}

// closingDutiesOf is ClosingDuties over evs and its places ids: per docketing, its span from its row —
// for a PROPOSAL docketing from the proposal's filing (L30) — to its ruling, or to the Close that ended
// it unruled.
func closingDutiesOf(evs []*Event, ids []int64, win WindowIndex) []ClosingDuty {
	r := closingReadsOf(evs, ids, win)
	filed := map[string]int64{}
	for _, m := range motionsAt(evs, ids, win) {
		if m != nil && m.Subject == "docket" {
			filed[m.ID] = m.filed
		}
	}
	var out []ClosingDuty
	docketed := map[string]bool{} // the proposals a docketing holds
	for _, d := range docketingsAt(evs, ids) {
		from := d.at
		if f, ok := filed[d.MotionID]; ok && d.MotionID != "" {
			from = f
			docketed[d.MotionID] = true
		}
		end := d.ruled
		if end == 0 {
			end = d.ended
		}
		out = append(out, r.dutiesOn(d.GapID, d.ID, from, end, d.Ruled())...)
	}
	// A PROPOSAL IN ITS ASSIGNEE'S WINDOW OWES THE CLOSINGS ITS DOCKETING WILL: that docketing's span
	// opens at the filing (L30), so the window sitting is the closing sitting. Rows of no docketing yet,
	// for the standing proposal alone; one its assignee accepted was answered with no docketing, and owes
	// none.
	_, ended := proposalStates(evs, ids, win)
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != "docket" || m.Ruled() || docketed[m.ID] || ended[m.ID] {
			continue
		}
		for _, d := range r.dutiesOn(m.GapID, "", m.filed, 0, false) {
			d.Proposal = m.ID
			out = append(out, d)
		}
	}
	return out
}

// standingClosingDuties is the duties of each gap's standing docketing: one docketing stands per gap,
// and one ended by a Close owes nothing more.
func standingClosingDuties(duties []ClosingDuty) []ClosingDuty {
	var out []ClosingDuty
	for _, d := range duties {
		if !d.Ruled && d.ruled == 0 {
			out = append(out, d)
		}
	}
	return out
}

// closingBlockersOf is the gate list's closing arm: each party's closing an open gap's standing
// docketing still lacks — awaited or missing — owned by the party.
func closingBlockersOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap) []Blocker {
	open := openSet(gaps)
	var out []Blocker
	for _, d := range standingClosingDuties(closingDutiesOf(evs, ids, win)) {
		if !open[d.GapID] || (d.State != ClosingAwaited && d.State != ClosingMissing) {
			continue
		}
		before := "docketing " + d.Docketing
		if d.Docketing == "" {
			before = "proposal " + d.Proposal + " in its assignee's window"
		}
		out = append(out, Blocker{Kind: BlockerClosingOwed, Subject: d.GapID, Owner: d.Party, Detail: before, About: d.GapID, since: d.filed,
			awaited: d.State == ClosingAwaited})
	}
	return out
}

// owedBlockersOf is the blockers a seat owes by an act of its own beyond the gaps' own: the stream's
// arms (eventBlockersOf, the reading debts among them, read off found), the closings owed
// (closingBlockersOf), the inquiries' (inquiryBlockersOf), which read the open gaps for the wanting
// ones, and the question's (partBlockersOf), off parts — the question_part_state view's rows on the
// same snapshot. A lens owes each one it owns (lensOwesOf).
func owedBlockersOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap, found []locatedReading, parts []QuestionPartRow) []Blocker {
	out := append(eventBlockersOf(evs, ids, win, openSet(gaps), found), closingBlockersOf(evs, ids, win, gaps)...)
	out = append(out, inquiryBlockersOf(evs, ids, win, gaps)...)
	return append(out, partBlockersOf(castOfEvents(evs) != nil, parts)...)
}
