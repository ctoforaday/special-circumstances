package record

import (
	"fmt"
	"math"
	"sort"
	"strings"

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
// The gaps a lens is engaged on and a contradiction it read ready it on their own grounds
// (foldPlan), whatever the head.
//
// A lens that keeps finding nothing stops being asked: b5's chair readied the voice lens alone
// eleven times as the head moved. So a lens RETIRES at its third sitting in a row that mints no
// fresh material, and a head move does not ready a retired lens.
//
// A LENS THAT OWES IS NOT RETIRED. Judging a proof and closing a gap mint nothing, so a lens could
// retire holding work no other seat may do: universe-m23's voice lens retired with its gap open and
// answered by blue, and nobody was readied to close it; a computation lens that only judges proofs
// retired, and every proof recorded after shipped "not judged". What a lens owes is each item its
// work list names as its own to clear that is still undone (lensOwesOf) — on a gap it minted, what is
// new since it last sat; a blocker whose owner it is; a proof it owns that nobody has judged — and
// while it owes any of it, the head readies it as it readies an unretired lens. That is the bound: it
// sits at most once per head move, and the head moves only by blue's edits, which the gaps' limits and
// the run's epoch limit bound. What is new on a gap is owed only until its next sitting, so past its
// limit it costs at most one sitting per head move that finds it unsat-on; an item it never
// discharges keeps it ready only when the report has changed.
//
//   - A SITTING is the first opening after a dispatch naming the lens (sittingFor) — the hook's
//     bracket, or a register no bracket preceded — including the sittings it is engaged for on its
//     own gaps.
//   - ITS PIN IS THE HEAD THE WHOLE REPORT WAS LAST SERVED TO IT AT (reportread.go): the latest
//     report read in a sitting it sat through. Opening a sitting is not auditing the report, and the
//     dispatch's own pin is the head the chair asked it to audit, not one it was given: a lens whose
//     sitting at the head ended without the whole report served to it is still behind the head, and
//     ready at it, until it is served it or retires.
//   - IT COUNTS WHEN THE LENS SAT THROUGH IT (satThrough). A sitting stopped at the per-sitting call
//     limit, or one that never closed, is neither sat nor barren: a read in it moves no pin, and it
//     adds nothing to the count.
//   - A mint belongs to the sitting whose opening most recently precedes it.
//   - A sitting is PRODUCTIVE when it minted a fresh gap (superseding nothing) that is material NOW.
//     A productive sitting returns the count to zero, a retired lens's included.
//
// Nothing is ready before a report is ingested.

// retiresAfter is how many sittings in a row with no fresh material mint retire a lens.
const retiresAfter = 3

// lensFold is one cast lens, folded.
type lensFold struct {
	seat     string
	barren   int      // trailing sittings with no fresh material mint
	sittings int      // sittings it sat through
	pin      int64    // the head of its latest report read in a sitting it sat through (pinOf); 0 when it has none
	owes     []string // what it owes (lensOwesOf), each named
	ready    bool
	why      string
}

// retired is a lens whose last retiresAfter sittings minted no fresh material and that owes nothing.
func (f lensFold) retired() bool { return f.barren >= retiresAfter && len(f.owes) == 0 }

// retiresWhy is the fold's clause on an unretired lens's retirement, in the terms that hold it: a
// lens past its barren limit is unretired only by what it owes, and the clause names each item; one
// below it retires at the limit, provided it then owes nothing.
func (f lensFold) retiresWhy() string {
	switch {
	case f.barren >= retiresAfter:
		return "it does not retire while it owes " + strings.Join(f.owes, ", ")
	case len(f.owes) > 0:
		return fmt.Sprintf("it retires at %d owing nothing, and it owes %s", retiresAfter, strings.Join(f.owes, ", "))
	}
	return fmt.Sprintf("it retires at %d", retiresAfter)
}

// readiness is whether the folded lens is ready at the report head, and its reason. short is
// unsatSitting's clause on the sitting of its latest dispatch, or "" when it sat through that one.
func (f lensFold) readiness(head int64, short string) (bool, string) {
	switch {
	case head == 0:
		return false, fmt.Sprintf("%s: not ready, no report has been ingested", f.seat)
	case f.retired():
		return false, fmt.Sprintf("%s: retired (pin %d) — its last %d sittings minted no fresh material; not ready", f.seat, f.pin, f.barren)
	case f.sittings == 0 && short != "":
		return true, fmt.Sprintf("%s: has sat through no sitting — %s — audits the report at head %d", f.seat, short, head)
	case f.sittings == 0:
		return true, fmt.Sprintf("%s: never sat — audits the report at head %d", f.seat, head)
	case head > f.pin:
		why := fmt.Sprintf("%s: %s, behind head %d — audits the report (%d sitting(s) in a row with no fresh material mint; %s)", f.seat, servedClause(f.pin), head, f.barren, f.retiresWhy())
		if short != "" {
			why += "; " + short
		}
		return true, why
	case f.barren >= retiresAfter:
		return false, fmt.Sprintf("%s: served the report at the head (pin %d) — ready when the head moves; %s", f.seat, f.pin, f.retiresWhy())
	}
	return false, fmt.Sprintf("%s: served the report at the head (pin %d) — ready when the head moves", f.seat, f.pin)
}

// servedClause is the fold's account of a lens's pin: the head the whole report was last served to it
// at, or that no sitting it sat through was served it.
func servedClause(pin int64) string {
	if pin == 0 {
		return "no sitting it sat through was served the whole report"
	}
	return fmt.Sprintf("the whole report was last served to it at its pin %d", pin)
}

// unsatSitting is the clause on a sitting the lens did not sit through: the place it opened at and
// satThrough's reason.
func unsatSitting(opened int64, reason string) string {
	return fmt.Sprintf("its sitting opened at %d %s, so it counts neither as sat nor toward retirement", opened, reason)
}

// owedBlocker is a blocker the lens owns, as lensOwesOf names it among the lens's debts.
func owedBlocker(b Blocker) string { return string(b.Kind) + " " + b.Subject }

// lensOwesOf is what each lens owes, by seat, each named: an item its work list names as its own to
// clear that is still undone. Each name is that item's alone — a finding and a remand by their own
// ids, since one gap may hold several — so a clause listing two items never reads as one listed twice.
//
//   - ON A GAP IT MINTED, WHAT IS NEW SINCE IT LAST SAT (gblock, 2026-10-18): blue's answer to the gap
//     (blueAnswersOf), an accepted grade motion awaiting its regrade (regradesAwaitedOf), the bench's
//     remand of it, another seat's finding about it (findingsAboutGapsOf) — each owed until the lens
//     sits through a sitting opened after it arrived (satThrough), whether or not that sitting acts
//     on it. A sitting stopped at the call limit, or one that never closed, does not end the debt. A gap it holds with
//     nothing new on it owes nothing: every act a lens has on its own gap waits on another seat's act,
//     or is a judgement it could make at mint, so a later sitting with nothing new gives it nothing to
//     act on. universe-m28's architecture and voice lenses held their own gaps that are not material,
//     which ready nobody and which blue never answers, and sat after every edit to the epoch limit.
//   - A blocker on the stream whose owner it is (events is eventBlockersOf): a contradiction it read,
//     an anchor reopened since its check, a proof it judged unsound that no gap of its carries yet.
//   - For the lens that owns proofs (proofOwnerOf), each proof nobody has judged.
//
// The work list keeps offering an item after the lens has sat on it; only the debt ends there.
func lensOwesOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap, events []Blocker) map[string][]string {
	out := map[string][]string{}
	_, openings := dispatchLedger(evs, ids, win)
	ends := satThroughOf(evs, ids, win, openings)
	minted := map[string]string{}
	for _, g := range gaps {
		minted[g.id] = g.mintedBy
	}
	// owe names item for lens while it has sat through no sitting that opened since at.
	owe := func(lens string, at int64, item string) {
		if _, sat := firstAfter(ends.through[lens], at); lens != "" && !sat {
			out[lens] = append(out[lens], item)
		}
	}
	for _, n := range gapNewsOf(evs, ids, win, gaps) {
		owe(minted[n.gap], n.at, n.item)
	}
	for _, b := range events {
		// A PROOF IT JUDGED UNSOUND IS OWED UNTIL ITS GAP IS MINTED: from then the gap carries it to
		// blue, and what clears the blocker is blue's fix or retire, which the lens cannot do.
		if b.Kind == BlockerUnsoundProof && b.answeredBy != "" {
			continue
		}
		if b.Owner != "" {
			out[b.Owner] = append(out[b.Owner], owedBlocker(b))
		}
	}
	if owner := proofOwnerOf(castOfEvents(evs)); owner != "" {
		for _, o := range proofChecksOwed(evs, win) {
			out[owner] = append(out[owner], "unjudged proof "+o.Anchor)
		}
	}
	return out
}

// gapNews is one act on a gap that its minting lens has not had: what it is, as a work list names
// it, and the events.id it arrived at.
type gapNews struct {
	gap, item string
	at        int64
}

// gapNewsOf IS "WHAT IS NEW ON A GAP" (gblock, 2026-10-18), for each gap of gaps: blue's answer to it
// (blueAnswersOf), an accepted grade motion awaiting its regrade (regradesAwaitedOf), the bench's
// remand of it, another seat's finding about it (findingsAboutGapsOf). It is one reading with two
// readers: the minting lens owes each item until it sits through a sitting opened after it
// (lensOwesOf), and a seat's act on the gap is refused while an item arrived after its sitting opened
// (staleOf).
func gapNewsOf(evs []*Event, ids []int64, win WindowIndex, gaps []blockerGap) []gapNews {
	var out []gapNews
	open, minted := map[string]bool{}, map[string]string{}
	for _, g := range gaps {
		open[g.id], minted[g.id] = true, g.mintedBy
	}
	answers := blueAnswersOf(evs, win, open)
	for _, g := range gaps {
		if a := answers[g.id]; a.answered() {
			out = append(out, gapNews{gap: g.id, at: a.at, item: "blue's answer to " + g.id})
		}
	}
	for _, o := range regradesAwaitedOf(evs, win, open) {
		out = append(out, gapNews{gap: o.GapID, at: o.ruled, item: "the accepted grade motion on " + o.GapID})
	}
	// A RULED DOCKET MOTION ON AN OPEN GAP IS A REMAND: every other disposition closes the gap. Only a
	// gap that is material or stranded carries one back into the debate (routeOf); the bench's ruling
	// on any other leaves it where it was, readying nobody.
	held := map[string]bool{}
	for _, g := range gaps {
		held[g.id] = g.material || g.supersededBy != ""
	}
	for _, m := range motionsAt(evs, ids, win) {
		if m.Subject == "docket" && m.Ruled() && held[m.GapID] {
			out = append(out, gapNews{gap: m.GapID, at: m.ruled, item: "the bench's remand of " + m.GapID + " in " + m.ID})
		}
	}
	for _, f := range findingsAboutGapsOf(evs, win, open, minted) {
		out = append(out, gapNews{gap: f.gap, at: f.at, item: f.filer + "'s finding " + f.id + " about " + f.gap})
	}
	return out
}

// StaleArea is a retired lens whose pin the report head has moved past: text it will not audit
// again changed after the whole report was last served to it. The chair reads those changes against the area's duties and
// names the area in its spot-check before a PASS.
type StaleArea struct {
	SeatID string `json:"seat_id"`
	Pin    int64  `json:"pin" min:"0"`
}

// lensStates folds every cast lens, in cast order. ids are the events' row ids, aligned with evs;
// head is the report head (0: nothing ingested); fresh is the set of gap ids minted fresh and
// material now; owes is lensOwesOf.
func lensStates(evs []*Event, ids []int64, win WindowIndex, head int64, fresh map[string]bool, owes map[string][]string) []lensFold {
	cast := castOfEvents(evs)
	ds, openings := dispatchLedger(evs, ids, win)
	ends := satThroughOf(evs, ids, win, openings)
	reads := servedOf(evs, ids, win)
	var out []lensFold
	for _, seat := range cast {
		if seatclass.RoleOfSeat(seat) != "lens" {
			continue
		}
		f := lensFold{seat: seat, owes: owes[seat], pin: pinOf(reads[seat], ends, seat, math.MaxInt64)}
		sat := map[int64]bool{}
		var sittings []int64
		var short string // why the sitting of its latest dispatch is not one it sat through; "" when it is
		for _, d := range ds {
			if d.seat != seat {
				continue
			}
			r, ok := sittingFor(openings[seat], d)
			if !ok {
				continue
			}
			short = ""
			if !ends.sat(seat, r) {
				short = unsatSitting(r, ends.short[r])
			}
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
		// A SITTING IT DID NOT SIT THROUGH IS NEITHER SAT NOR BARREN (satThrough): it is skipped, so
		// the count runs on as if it had not been dispatched. A fresh material mint it made is on the
		// record all the same, and resets the count.
		for _, s := range sittings {
			through := ends.sat(seat, s)
			if through {
				f.sittings++
			}
			switch {
			case productive[s]:
				f.barren = 0
			case through:
				f.barren++
			}
		}
		f.ready, f.why = f.readiness(head, short)
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

func passLensGateOf(evs []*Event, ids []int64, win WindowIndex, fresh map[string]bool, owes map[string][]string) passLensGate {
	g := passLensGate{cast: castOfEvents(evs) != nil, head: reportHeadOf(evs, ids), covered: map[string]bool{}}
	if !g.cast {
		return g
	}
	g.lenses = lensStates(evs, ids, win, g.head, fresh, owes)
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
// holds, and one closed between it and the open-gap read a debt no fold holds. The fold runs after
// the transaction closes.
func passLensGateOfRun(run Run) (passLensGate, bool, error) {
	var r planReads
	absent := false
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		if absent = noRecord(q); absent {
			return nil
		}
		var err error
		r, err = passBlockerReadsAt(q)
		return err
	}); err != nil || absent {
		return passLensGate{}, false, err
	}
	ids := r.win.IDs(r.evs)
	owes := lensOwesOf(r.evs, ids, r.win, blockerGapsOfOpen(r.gaps, nil), owedBlockersOf(r.evs, ids, r.win, blockerGapsOfOpen(r.gaps, nil)))
	return passLensGateOf(r.evs, ids, r.win, r.fresh, owes), true, nil
}

// LastReadJSON is a lens's pin as its work view carries it: the head the whole report was last served
// to it at, in a sitting it SAT THROUGH that opened before the dispatch it is now sitting for, beside
// that dispatch's head. Kind is first | behind | unchanged | undispatched; first is a lens no such
// sitting was served the report in.
type LastReadJSON struct {
	Kind string `json:"kind"`
	Pin  int64  `json:"pin"`
	Head int64  `json:"head"`
}

// lastReadBefore is NOT the fold's pin. The lens reads its work view in its current sitting, before
// and after it is served the report, and a read in that sitting would turn `behind` into `unchanged`
// halfway through the audit the work view asks for. D is the latest dispatch naming the seat (the one
// it sits for, or owes); the pin is its reads in sittings that opened before D and that it sat through
// (satThrough), so a sitting stopped at the call limit or never closed is not a read of the report it
// was served.
func lastReadBefore(evs []*Event, ids []int64, win WindowIndex, seatID string) LastReadJSON {
	ds, openings := dispatchLedger(evs, ids, win)
	var cur *dispatchRow
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].seat == seatID {
			cur = &ds[i]
			break
		}
	}
	if cur == nil {
		return LastReadJSON{Kind: "undispatched"}
	}
	pin := pinOf(servedOf(evs, ids, win)[seatID], satThroughOf(evs, ids, win, openings), seatID, cur.at)
	ls := LastReadJSON{Pin: pin, Head: cur.pin}
	switch {
	case pin == 0:
		ls.Kind = "first"
	case pin < cur.pin:
		ls.Kind = "behind"
	default:
		ls.Kind = "unchanged"
	}
	return ls
}
