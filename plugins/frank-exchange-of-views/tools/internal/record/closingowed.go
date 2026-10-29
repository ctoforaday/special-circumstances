package record

import (
	"fmt"
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A CLOSING IS A PARTY'S ARGUMENT TO THE BENCH ON ONE DOCKETED GAP, AND THIS FILE IS WHERE WHO OWES
// ONE IS SAID (gblock, 2026-10-20).
//
// A gap's dispute has two parties: the lens that minted it, which holds it from mint to close, and
// blue's responding seat, which answers it. Once the gap is docketed — the dispatch files the docket
// motion at impasse or for its grade movement, or a party files one — the bench rules it on the two
// parties' closings, the transcript and the artifacts as they stand, so each party owes one closing
// for each docketing. The chair is party to no dispute and files none. A lens's regrade that keeps
// its grade is its move while the dispute runs; the closing is its brief once the gap is before the
// bench.
//
// ONE READING, AND EVERY SURFACE THAT ENFORCES, LISTS OR AUDITS THE DUTY CALLS IT: the PASS gate's
// list (closing_owed, passBlockersOf) — through which the plan readies the party that owes and
// convenes the bench only once no party is awaited, the work list blocks the party, and the lens
// stays unretired — and capture's record-parity audit.
//
// A CLOSING ANSWERS ONE DOCKETING: a closing of the party's on the gap filed after the docket motion
// and before its ruling. The write path refuses one on a gap no unruled docket motion stands on, so
// one filed for an earlier docketing, before a remand, answers that docketing and no later one.
//
// OWED BY EACH PARTY, WITH TWO EXCEPTIONS, each a state of its own (ClosingNotOwed):
//   - the gap closed after the docketing: its lens closed it, so the dispute is over, and the bench
//     rules the motion on the record as it stands;
//   - the gap's minter is no lens the cast seats — an archived run's merge mints translate to the
//     chair — so there is no minting party to dispatch, and blue's closing is the only one owed.
//
// AWAITED AND MISSING ARE APART BY WHETHER THE PARTY HAS SAT SINCE THE DOCKETING. A party that has
// sat through no sitting opened after the motion was filed is awaited: the plan readies it, and the
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
	// Motion is the docket motion's id; Ruled whether the bench has ruled it.
	Motion string
	Ruled  bool
	Party  string
	State  ClosingState
	// NotOwedWhy is the exception that holds, for ClosingNotOwed; "" otherwise.
	NotOwedWhy string
	// filed and ruled are the motion's places (ruled 0 while unruled).
	filed, ruled int64
	// closedBy is the seat whose close ended the duty, for ClosingNotOwed on a closed gap.
	closedBy string
}

// Why a closing is not owed.
const (
	closingGapClosed    = "the gap closed after it was docketed, so the dispute it would argue is over"
	closingMinterNoLens = "the gap's minter is no lens the cast seats, so there is no minting party"
)

// closingPartiesOf is the parties a gap's docketing owes closings from: its minter where the cast
// seats it as a lens, and blue's responding seat. ok is false for a minter that is not one.
func closingPartiesOf(cast []string, mintedBy string) (lens string, ok bool) {
	if seatclass.RoleOfSeat(mintedBy) == "lens" && slices.Contains(cast, mintedBy) {
		return mintedBy, true
	}
	return "", false
}

// IsClosingParty reports whether seatID is a party to the dispute on a gap its minter mintedBy
// minted: the minter itself, or blue's responding seat — the seats the write path admits a closing
// from.
func IsClosingParty(seatID, mintedBy string) bool {
	return seatID == blueRespondSeat || (seatID == mintedBy && seatclass.RoleOfSeat(seatID) == "lens")
}

// ClosingDuties is one row per party per docket motion on the record, in the motions' order, read
// off the whole stream: the reading capture audits after the run.
func ClosingDuties(evs []*Event, win WindowIndex) []ClosingDuty {
	return closingDutiesOf(evs, win.IDs(evs), win)
}

// closingDutiesOf is ClosingDuties over evs and its places ids.
func closingDutiesOf(evs []*Event, ids []int64, win WindowIndex) []ClosingDuty {
	cast := castOfEvents(evs)
	minter := map[string]string{}
	closedAt := map[string]int64{}
	closer := map[string]string{}
	type filing struct {
		seat, gap string
		at        int64
	}
	var closings []filing
	for i, e := range evs {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.Mint:
			minter[b.GetGapId()] = e.GetSeatId()
		case *recordpb.Close:
			if _, seen := closedAt[b.GetGapId()]; !seen {
				closedAt[b.GetGapId()], closer[b.GetGapId()] = ids[i], e.GetSeatId()
			}
		case *recordpb.Closing:
			closings = append(closings, filing{seat: e.GetSeatId(), gap: b.GetGapId(), at: ids[i]})
		}
	}
	_, openings := dispatchLedger(evs, ids, win)
	ends := satThroughOf(evs, ids, win, openings)
	var out []ClosingDuty
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != "docket" || m.GapID == "" {
			continue
		}
		// before is the place the docketing ends at: its ruling, or nowhere while it stands.
		before := func(at int64) bool { return at > m.filed && (m.ruled == 0 || at < m.ruled) }
		row := func(party string) ClosingDuty {
			d := ClosingDuty{GapID: m.GapID, Motion: m.ID, Ruled: m.Ruled(), Party: party, filed: m.filed, ruled: m.ruled}
			for _, c := range closings {
				if c.seat == party && c.gap == m.GapID && before(c.at) {
					d.State = ClosingFiled
					return d
				}
			}
			if at, closed := closedAt[m.GapID]; closed && before(at) {
				d.State, d.NotOwedWhy, d.closedBy = ClosingNotOwed, closingGapClosed, closer[m.GapID]
				return d
			}
			d.State = ClosingAwaited
			if o, sat := firstAfter(ends.through[party], m.filed); sat && (m.ruled == 0 || o < m.ruled) {
				d.State = ClosingMissing
			}
			return d
		}
		if lens, ok := closingPartiesOf(cast, minter[m.GapID]); ok {
			out = append(out, row(lens))
		} else {
			out = append(out, ClosingDuty{GapID: m.GapID, Motion: m.ID, Ruled: m.Ruled(), Party: minter[m.GapID],
				State: ClosingNotOwed, NotOwedWhy: closingMinterNoLens, filed: m.filed, ruled: m.ruled})
		}
		out = append(out, row(blueRespondSeat))
	}
	return out
}

// standingClosingDuties is the duties of each gap's standing docketing: its newest unruled docket
// motion — the filing the bench's sitting for the gap is keyed on (the gap view's
// `unruled_docket_filed`, benchSatFor). A closing filed after it answers every older one as well.
func standingClosingDuties(duties []ClosingDuty) []ClosingDuty {
	newest := map[string]int64{}
	for _, d := range duties {
		if !d.Ruled && d.filed > newest[d.GapID] {
			newest[d.GapID] = d.filed
		}
	}
	var out []ClosingDuty
	for _, d := range duties {
		if !d.Ruled && d.filed == newest[d.GapID] {
			out = append(out, d)
		}
	}
	return out
}

// closingsNotOwedItemsOf is what a party's work list states of each standing docketing it is a party
// to and owes no closing on because another seat closed the gap (ClosingNotOwed): the gap, the
// motion, the exception that holds and whose close it was. The party that closed the gap itself is
// told nothing: the close is its own act.
//
// THE DISPATCH NAMED THE CLOSING, AND THE LIST HAS TO SAY WHY IT IS GONE. The plan is read at the
// epoch's start; a lens that closes the gap before blue sits ends the duty, and a list that only
// drops the item leaves blue holding a dispatch that names a duty and a list that names nothing —
// the absent case and the discharged case the same bytes (#1601 smoke, blue ev722).
func closingsNotOwedItemsOf(evs []*Event, ids []int64, win WindowIndex, seatID string) []string {
	var out []string
	for _, d := range standingClosingDuties(closingDutiesOf(evs, ids, win)) {
		if d.Party == seatID && d.State == ClosingNotOwed && d.closedBy != "" && d.closedBy != seatID {
			out = append(out, fmt.Sprintf("gap %s is docketed for the bench (motion %s) and no closing argument is owed from you on it: %s (%s closed it); the bench rules the motion on the record as it stands", d.GapID, d.Motion, d.NotOwedWhy, d.closedBy))
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
		out = append(out, Blocker{Kind: BlockerClosingOwed, Subject: d.GapID, Owner: d.Party, Detail: d.Motion, About: d.GapID, since: d.filed,
			awaited: d.State == ClosingAwaited})
	}
	return out
}

// owedBlockersOf is the blockers a seat owes by an act of its own beyond the gaps' own: the stream's
// arms (eventBlockersOf) and the closings owed (closingBlockersOf). A lens owes each one it owns
// (lensOwesOf).
func owedBlockersOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap) []Blocker {
	return append(eventBlockersOf(evs, ids, win), closingBlockersOf(evs, ids, win, gaps)...)
}
