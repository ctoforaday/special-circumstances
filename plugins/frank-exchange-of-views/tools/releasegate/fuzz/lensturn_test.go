package fuzz

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE LENS'S TURN ON A GAP COMES WITH BLUE'S ACT (ruled by gblock 2026-10-30), OR WITH AN EDIT OF ITS
// SENTENCE UNDER ANOTHER GAP'S NAME, AND THE LENS GOES BEFORE THE DOCKET (ruled by gblock 2026-10-31).
//
// The engine sits the lenses before blue in an epoch, so a lens engaged on a gap blue has not acted
// on since the lens's own last act there audits an answer that does not exist yet: on universe-m36,
// seven lens sittings each recorded "blue has not edited" before blue sat. tallyLensTurns holds every
// plan the chair relays to the rule from the record's own acts, read here and not through the
// dispatch's fold: a lens engaged on a gap is a failure unless blue moved it after the lens's latest
// turn there — its mint, an act, or the opening of a closed sitting it was engaged on, silent or not —
// and after the bench's latest remand of it; and a gap blue is engaged on whose minting lens's turn HAS
// come must engage that lens too, so the rule never holds a lens off blue's answer. A gap a successor
// names is the exception both ways: its lens owes the close the supersession promised, whatever blue
// did.
//
// A plan's docket comes after the lens: on universe-m37 the dispatch docketed six gaps with the lens's
// audit of blue's answer still to come, and the lens closed five of them minutes after. A gap the plan
// dockets whose lens's turn stands is a failure unless the lens has sat on it since its floor, which is
// the audit the docket waited for.

// gapActs is, per gap, the stream position of each act on it and the seat that recorded it, by the
// fields the acts name a gap in.
type gapActs map[string][]struct {
	seat string
	at   int
}

// actsOnGaps reads every act on a gap off evs, and each gap's mint position and stranded state.
func actsOnGaps(evs []*record.Event) (acts gapActs, minted map[string]int, superseded map[string]bool) {
	acts, minted, superseded = gapActs{}, map[string]int{}, map[string]bool{}
	computationGap := map[string]string{} // computation id → the gap it answers
	anchored := map[string][]string{}     // inquiry id → the gaps minted about it
	add := func(seat string, at int, gaps ...string) {
		for _, g := range gaps {
			if g != "" {
				acts[g] = append(acts[g], struct {
					seat string
					at   int
				}{seat, at})
			}
		}
	}
	for i, e := range evs {
		seat := e.GetSeatId()
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			if _, seen := minted[m.GetGapId()]; !seen {
				minted[m.GetGapId()] = i
			}
			for _, s := range m.GetSupersedes() {
				superseded[s] = true
			}
			if m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
				anchored[m.GetAboutRef()] = append(anchored[m.GetAboutRef()], m.GetGapId())
			}
			add(seat, i, m.GetSupersedes()...)
		}
		if b, ok := recordpb.BodyAs[*recordpb.Regrade](e); ok {
			add(seat, i, b.GetGapId())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
			add(seat, i, b.GetGapId())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Closing](e); ok {
			add(seat, i, b.GetGapId())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_GAP {
			add(seat, i, b.GetAboutRef())
		}
		if b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
			add(seat, i, b.GetAnswers())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Retire](e); ok && b.GetClaimKept() {
			add(seat, i, b.GetAnswers())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Computation](e); ok {
			computationGap[b.GetComputationId()] = b.GetAnswers()
			add(seat, i, b.GetAnswers())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
			add(seat, i, computationGap[b.GetComputationId()])
		}
		if b, ok := recordpb.BodyAs[*recordpb.Inquiry](e); ok {
			add(seat, i, anchored[b.GetInquiryId()]...)
		}
		if b, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			add(seat, i, b.GetDocket().GetGapId(), b.GetGrade().GetGapId())
		}
	}
	return acts, minted, superseded
}

// turnsOnGaps is acts with what else ends or begins a lens's turn: blue's edit of a sentence a gap is
// anchored on, under whatever name it answers (the record's `reopened`), and the opening of each closed
// sitting a seat was dispatched onto the gap for. A sitting nothing has closed yet is no turn taken.
func turnsOnGaps(evs []*record.Event, at record.WindowIndex, acts gapActs) (turns, sittings gapActs) {
	turns, sittings = gapActs{}, gapActs{}
	for g, as := range acts {
		turns[g] = append(turns[g], as...)
	}
	add := func(to gapActs, seat string, i int, gaps ...string) {
		for _, g := range gaps {
			to[g] = append(to[g], struct {
				seat string
				at   int
			}{seat, i})
		}
	}
	openings := map[string][]int{}
	agentOf := map[int]string{}
	stops := map[string][]int{}
	for i, e := range evs {
		if seat, opens := at.Opens(e); opens {
			openings[seat] = append(openings[seat], i)
		}
		if a := recordpb.AgentOpening(e); a != "" {
			agentOf[i] = a
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() != "" {
			stops[b.GetAgentId()] = append(stops[b.GetAgentId()], i)
		}
		if b, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok && b.GetOld() != b.GetNew() {
			add(turns, e.GetSeatId(), i, b.GetReopened()...)
		}
	}
	firstAfter := func(xs []int, i int) (int, bool) {
		for _, x := range xs {
			if x > i {
				return x, true
			}
		}
		return 0, false
	}
	for i, e := range evs {
		d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
		if !ok {
			continue
		}
		start, sat := firstAfter(openings[d.GetSeatId()], i)
		if !sat {
			continue
		}
		_, closed := firstAfter(openings[d.GetSeatId()], start)
		if !closed && agentOf[start] != "" {
			_, closed = firstAfter(stops[agentOf[start]], start)
		}
		if closed {
			add(sittings, d.GetSeatId(), start, d.GetGapIds()...)
			add(turns, d.GetSeatId(), start, d.GetGapIds()...)
		}
	}
	return turns, sittings
}

// remandRulings is the position of the bench's latest remand of each gap: a docket ruled remanded,
// and an upheld position naming the gap, which the engine reads as the same remand (blueanswered.go).
func remandRulings(evs []*record.Event) map[string]int {
	docketGap := map[string]string{} // docket motion id → its gap
	out := map[string]int{}
	for i, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && m.GetDocket().GetGapId() != "" {
			docketGap[m.GetMotionId()] = m.GetDocket().GetGapId()
		}
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetDocket().GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
			if g := docketGap[r.GetMotionId()]; g != "" {
				out[g] = i
			}
		}
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetPosition().GetRuling() == recordpb.PositionRuling_POSITION_RULING_UPHELD {
			for _, g := range r.GetPosition().GetGapIds() {
				out[g] = i
			}
		}
	}
	return out
}

// tallyLensTurns holds plan, as relayed, to the lens-turn rule on the record as the plan read it.
func (r *runner) tallyLensTurns(plan map[string]any) {
	if plan["epoch_limit_reached"] == true {
		return // the plan at the epoch limit dispatches nobody
	}
	b, ok := r.board()
	if !ok {
		r.noteActErr("lens-turn oracle: the record could not be read after the chair's dispatch")
		return
	}
	acts, minted, superseded := actsOnGaps(b.Events)
	turnActs, sittings := turnsOnGaps(b.Events, b.At, acts)
	remanded := remandRulings(b.Events)
	minter := map[string]string{}
	for _, e := range b.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			if _, seen := minter[m.GetGapId()]; !seen {
				minter[m.GetGapId()] = e.GetSeatId()
			}
		}
	}
	// turnCame is whether blue moved gap after lens's latest turn there, its mint and the gap's latest
	// remand ruling.
	turnCame := func(lens, gap string) bool {
		last := max(minted[gap], remanded[gap])
		for _, a := range turnActs[gap] {
			if a.seat == lens {
				last = max(last, a.at)
			}
		}
		for _, a := range turnActs[gap] {
			if a.at > last && seatclass.RoleOfSeat(a.seat) == "blue" {
				return true
			}
		}
		return false
	}
	// witness is what the oracle read on gap, and the plan's and the fold's readings beside it: a failure
	// here is one reading of the record against another, and the sweep's log is all a red leaves behind.
	witness := func(gap string) string {
		fold := "unread"
		if p, err := record.RunParams(r.run()); err == nil {
			if xs, err := record.Exchanges(r.run(), p); err == nil && xs[gap] != nil {
				x := xs[gap]
				fold = fmt.Sprintf("%d exchange(s), %d stalled, %d unresolved, impasse %v, remand %v", x.Exchanges, x.Stalled, x.Unresolved, x.Impasse, x.Remand)
			}
		}
		var turns, why []string
		for _, a := range turnActs[gap] {
			turns = append(turns, fmt.Sprintf("%s@%d:%s", a.seat, a.at, b.Events[a.at].GetType()))
		}
		whys, _ := plan["why"].([]any)
		for _, w := range whys {
			if text := fmt.Sprint(w); strings.Contains(text, gap) {
				why = append(why, text)
			}
		}
		return fmt.Sprintf("minted %d, remanded %d, of %d events; turns %v; sittings %v; the fold: %s; the plan: %v", minted[gap], remanded[gap], len(b.Events), turns, sittings[gap], fold, why)
	}
	engaged := map[string]map[string]bool{} // gap → seats the plan engages on it
	parties, _ := plan["parties"].([]any)
	for _, p := range parties {
		m, _ := p.(map[string]any)
		gs, _ := m["gap_ids"].([]any)
		for _, g := range gs {
			id := fmt.Sprint(g)
			if engaged[id] == nil {
				engaged[id] = map[string]bool{}
			}
			engaged[id][fmt.Sprint(m["seat_id"])] = true
		}
	}
	turns, waits, alone := 0, 0, 0
	for gap, seats := range engaged {
		lens := minter[gap]
		if lens != "" && seats[lens] && !seats["blue-respond"] {
			alone++
		}
		for seat := range seats {
			if seatclass.RoleOfSeat(seat) == "lens" && !superseded[gap] && !turnCame(seat, gap) {
				r.noteActErr(fmt.Sprintf("lens-turn oracle: the plan engages %s on %s, and blue has recorded no act on %s since %s's last act there — the lens would audit an answer that does not exist yet", seat, gap, gap, seat))
				return
			}
		}
		if !seats["blue-respond"] || lens == "" || superseded[gap] {
			continue
		}
		switch came := turnCame(lens, gap); {
		case came && !seats[lens]:
			r.noteActErr(fmt.Sprintf("lens-turn oracle: the plan engages blue on %s and not its minting lens %s, whose turn has come — blue acted on %s since the lens's last act there: %s", gap, lens, gap, witness(gap)))
			return
		case came:
			turns++
		default:
			waits++
		}
	}
	// A DOCKET COMES AFTER THE LENS'S SITTING, read off the record's sittings and not through the fold
	// that holds it back. Only what this dispatch filed: `docket` also names every motion already standing,
	// and a party's filing is not the dispatch's order. The dispatch's motions are the tail of the record,
	// written since the chair last opened a sitting.
	after := 0
	for i := len(b.Events) - 1; i >= 0; i-- {
		e := b.Events[i]
		if _, opens := b.At.Opens(e); opens || e.GetSeatId() != "red-chair" {
			break
		}
		m, ok := recordpb.BodyAs[*recordpb.Motion](e)
		if !ok || m.GetDocket().GetGapId() == "" {
			continue
		}
		gap := m.GetDocket().GetGapId()
		lens := minter[gap]
		if seatclass.RoleOfSeat(lens) != "lens" || superseded[gap] {
			continue
		}
		floor, sat := max(minted[gap], remanded[gap]), false
		for _, s := range sittings[gap] {
			sat = sat || (s.seat == lens && s.at > floor)
		}
		switch {
		case sat:
			after++
		case turnCame(lens, gap):
			r.noteActErr(fmt.Sprintf("docket-order oracle: the dispatch filed %s on %s, blue has moved it since %s's last turn there, and the lens has not sat on it since its floor — the bench is asked before the lens has read the answer: %s", m.GetMotionId(), gap, lens, witness(gap)))
			return
		}
	}
	r.mu.Lock()
	r.lensTurns += turns
	r.lensWaits += waits
	r.lensAlone += alone
	r.docketsAfterAudit += after
	r.mu.Unlock()
}
