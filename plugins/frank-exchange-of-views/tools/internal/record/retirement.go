package record

import (
	"fmt"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// LENS READINESS AND RETIREMENT (plans/feov-lens-bar.md III.5).
//
// A LENS WITH NO GAP ENGAGED IS READY WHEN THE REPORT HEAD HAS MOVED SINCE ITS LAST SITTING, AND NOT
// OTHERWISE. A second look at an unchanged report finds what the first did: across the archived
// runs, the lenses readied at the head they had last sat at minted nothing and checked nothing,
// and four in five recorded no act at all. A lens that has never sat is ready for its first look.
// The gaps a lens is engaged on, a contradiction it read and a reopened citation or proof ready it
// on their own grounds (foldPlan), whatever the head.
//
// A lens that keeps finding nothing stops being asked: b5's chair readied the voice lens alone
// eleven times as the head moved. So a lens RETIRES at its third sitting in a row that mints no
// fresh material, and a head move does not ready a retired lens.
//
//   - A SITTING is the first opening after a dispatch naming the lens (sittingFor) — the hook's
//     bracket, or a register no bracket preceded — including the sittings it is engaged for on its
//     own gaps. Its PIN is the head that dispatch recorded.
//   - A mint belongs to the sitting whose opening most recently precedes it.
//   - A sitting is PRODUCTIVE when it minted a fresh gap (superseding nothing) that is material NOW.
//     A productive sitting returns the count to zero, a retired lens's included.
//
// Nothing is ready before a report is ingested. A retired lens still answers its own open gaps.

// retiresAfter is how many sittings in a row with no fresh material mint retire a lens.
const retiresAfter = 3

// lensFold is one cast lens, folded.
type lensFold struct {
	seat     string
	barren   int   // trailing sittings with no fresh material mint
	sittings int   // sittings on the record
	pin      int64 // pin of the last dispatch it sat for; 0 when it never sat
	ready    bool
	why      string
}

// retired is a lens whose last retiresAfter sittings minted no fresh material.
func (f lensFold) retired() bool { return f.barren >= retiresAfter }

// StaleArea is a retired lens whose pin the report head has moved past: text it will not audit
// again changed after it last sat. The chair reads those changes against the area's duties and
// names the area in its spot-check before a PASS.
type StaleArea struct {
	SeatID string `json:"seat_id"`
	Pin    int64  `json:"pin" min:"0"`
}

// lensStates folds every cast lens, in cast order. ids are the events' row ids, aligned with evs;
// head is the report head (0: nothing ingested); fresh is the set of gap ids minted fresh and
// material now.
func lensStates(evs []*Event, ids []int64, win WindowIndex, head int64, fresh map[string]bool) []lensFold {
	cast := castOfEvents(evs)
	ds, openings := dispatchLedger(evs, ids, win)
	var out []lensFold
	for _, seat := range cast {
		if seatclass.RoleOfSeat(seat) != "lens" {
			continue
		}
		f := lensFold{seat: seat}
		sat := map[int64]bool{}
		var sittings []int64
		for _, d := range ds {
			if d.seat != seat {
				continue
			}
			r, ok := sittingFor(openings[seat], d)
			if !ok {
				continue
			}
			f.pin = d.pin
			if !sat[r] {
				sat[r] = true
				sittings = append(sittings, r)
			}
		}
		sort.Slice(sittings, func(a, b int) bool { return sittings[a] < sittings[b] })
		productive := map[int64]bool{}
		for i, e := range evs {
			if e.GetSeatId() != seat {
				continue
			}
			m, ok := recordpb.BodyAs[*recordpb.Mint](e)
			if !ok || !fresh[m.GetGapId()] {
				continue
			}
			// the sitting whose opening most recently precedes the mint
			k := sort.Search(len(sittings), func(j int) bool { return sittings[j] >= ids[i] }) - 1
			if k >= 0 {
				productive[sittings[k]] = true
			}
		}
		for _, s := range sittings {
			f.sittings++
			if productive[s] {
				f.barren = 0
			} else {
				f.barren++
			}
		}
		switch {
		case head == 0:
			f.why = fmt.Sprintf("%s: not ready, no report has been ingested", seat)
		case f.retired():
			f.why = fmt.Sprintf("%s: retired (pin %d) — its last %d sittings minted no fresh material; not ready", seat, f.pin, f.barren)
		case f.sittings == 0:
			f.ready, f.why = true, fmt.Sprintf("%s: never sat — audits the report at head %d", seat, head)
		case head > f.pin:
			f.ready, f.why = true, fmt.Sprintf("%s: head %d moved past its last sitting's pin %d — audits the report (%d sitting(s) in a row with no fresh material mint; it retires at %d)", seat, head, f.pin, f.barren, retiresAfter)
		default:
			f.why = fmt.Sprintf("%s: sat at the head (pin %d) — ready when the head moves", seat, f.pin)
		}
		out = append(out, f)
	}
	return out
}

// staleAreasOf is every retired lens whose pin is behind the head, in cast order.
func staleAreasOf(folds []lensFold, head int64) []StaleArea {
	out := []StaleArea{}
	for _, f := range folds {
		if f.retired() && f.pin < head {
			out = append(out, StaleArea{SeatID: f.seat, Pin: f.pin})
		}
	}
	return out
}

// freshMaterialOf is the gap view's answer to "which gaps were minted fresh and are material now".
func freshMaterialOf(db recordsql.Querier) (map[string]bool, error) {
	rows, err := db.Query(`SELECT "gap_id" FROM "gap" WHERE "material" AND "supersedes_count" = 0`)
	if err != nil {
		return nil, fmt.Errorf("asking the record for its fresh material gaps: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// freshMaterialOfStates is the same answer off the work path's gap states, which carry the view's
// columns.
func freshMaterialOfStates(gaps []WorkGapState) map[string]bool {
	out := map[string]bool{}
	for _, g := range gaps {
		if g.Material && len(g.Supersedes) == 0 {
			out[g.ID] = true
		}
	}
	return out
}

// reportHeadOf is the report head off the stream: the last blue_edit or base_ingest, by row id.
func reportHeadOf(evs []*Event, ids []int64) int64 {
	var head int64
	for i, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_BLUE_EDIT, recordpb.EventType_EVENT_TYPE_BASE_INGEST:
			if ids[i] > head {
				head = ids[i]
			}
		}
	}
	return head
}

// passLensGate is the PASS gate's lens conditions as ONE fold: passBlockersOf reads its ready
// lenses and uncovered stale areas, so the gate, the chair's work list and the plan state them
// from the same answer.
// The dispatch readies from the same lensStates fold.
type passLensGate struct {
	cast    bool
	head    int64
	lenses  []lensFold
	stale   []StaleArea
	covered map[string]bool // areas a spot-check in the chair's current sitting named
}

func passLensGateOf(evs []*Event, ids []int64, win WindowIndex, fresh map[string]bool) passLensGate {
	g := passLensGate{cast: castOfEvents(evs) != nil, head: reportHeadOf(evs, ids), covered: map[string]bool{}}
	if !g.cast {
		return g
	}
	g.lenses = lensStates(evs, ids, win, g.head, fresh)
	g.stale = staleAreasOf(g.lenses, g.head)
	// THIS SITTING is the chair's latest sitting as the write path stored it: its spot-checks in that
	// sitting — the same window the work list reads (thisSitting). A spot-check filed between the
	// chair's bracket and its register is in it, which a scan back to the chair's last register would
	// cut out.
	latest := win.LatestSittingOf(chairSeat)
	for _, e := range evs {
		if e.GetSeatId() != chairSeat || latest == 0 || win.Of(e).SittingID != latest {
			continue
		}
		if sc, ok := recordpb.BodyAs[*recordpb.SpotCheck](e); ok {
			for _, a := range sc.GetAreas() {
				g.covered[a] = true
			}
		}
	}
	return g
}

// ready is each cast lens the fold leaves ready.
func (g passLensGate) ready() []lensFold {
	var out []lensFold
	for _, f := range g.lenses {
		if f.ready {
			out = append(out, f)
		}
	}
	return out
}

// uncovered is each stale area no spot-check in the chair's current sitting names.
func (g passLensGate) uncovered() []StaleArea {
	var out []StaleArea
	for _, a := range g.stale {
		if !g.covered[a.SeatID] {
			out = append(out, a)
		}
	}
	return out
}

// passLensGateOfRun reads the gate's inputs off the run, on one snapshot (readSnapshot): a gap
// minted between the stream read and the fresh-material read would be fresh material no lens fold
// holds. The fold runs after the transaction closes.
func passLensGateOfRun(run Run) (passLensGate, bool, error) {
	var evs []*Event
	var win WindowIndex
	var fresh map[string]bool
	absent := false
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		if absent = noRecord(q); absent {
			return nil
		}
		var err error
		if evs, win, err = eventsAt(q); err != nil {
			return err
		}
		fresh, err = freshMaterialOf(q)
		return err
	}); err != nil || absent {
		return passLensGate{}, false, err
	}
	return passLensGateOf(evs, win.IDs(evs), win, fresh), true, nil
}

// LastSittingJSON is a lens's last sitting, as its work view carries it: the latest sitting STRICTLY
// BEFORE the dispatch it is now sitting for. Kind is first | behind | unchanged | undispatched.
type LastSittingJSON struct {
	Kind string `json:"kind"`
	Pin  int64  `json:"pin"`
	Head int64  `json:"head"`
}

// lastSittingBefore is NOT the latest dispatch's own sitting. The lens reads its work view once its
// sitting has opened — at the hook's bracket or at its register — and sittingFor counts that
// opening as having sat, so the sitting for the latest dispatch is the current one, and every lens
// would read `unchanged`. D is the latest dispatch naming the seat (the one it sits for, or owes);
// a prior sitting is a dispatch d before D whose sitting opened before D, so an unsat earlier
// dispatch cannot borrow the current opening.
func lastSittingBefore(evs []*Event, ids []int64, win WindowIndex, seatID string) LastSittingJSON {
	ds, openings := dispatchLedger(evs, ids, win)
	var cur *dispatchRow
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].seat == seatID {
			cur = &ds[i]
			break
		}
	}
	if cur == nil {
		return LastSittingJSON{Kind: "undispatched"}
	}
	var prior *dispatchRow
	for i := range ds {
		d := ds[i]
		if d.seat != seatID || d.at >= cur.at {
			continue
		}
		if r, ok := sittingFor(openings[seatID], d); ok && r < cur.at {
			prior = &ds[i]
		}
	}
	ls := LastSittingJSON{Head: cur.pin}
	switch {
	case prior == nil:
		ls.Kind = "first"
	case prior.pin < cur.pin:
		ls.Kind, ls.Pin = "behind", prior.pin
	default:
		ls.Kind, ls.Pin = "unchanged", prior.pin
	}
	return ls
}
