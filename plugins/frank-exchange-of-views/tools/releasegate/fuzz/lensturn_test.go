package fuzz

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE LENS'S TURN ON A GAP COMES WITH BLUE'S ACT (ruled by gblock 2026-10-30).
//
// The engine sits the lenses before blue in an epoch, so a lens engaged on a gap blue has not acted
// on since the lens's own last act there audits an answer that does not exist yet: on universe-m36,
// seven lens sittings each recorded "blue has not edited" before blue sat. tallyLensTurns holds every
// plan the chair relays to the rule from the record's own acts, read here and not through the
// dispatch's fold: a lens engaged on a gap is a failure unless blue recorded an act on it after the
// lens's latest act there — its mint included — and after the bench's latest remand of it; and a gap
// blue is engaged on whose minting lens's turn HAS come must engage that lens too, so the rule never
// holds a lens off blue's answer. A gap a successor names is the exception both ways: its lens owes
// the close the supersession promised, whatever blue did.

// gapActs is, per gap, the stream position of each act on it and the seat that recorded it, by the
// fields the acts name a gap in.
type gapActs map[string][]struct {
	seat string
	at   int
}

// actsOnGaps reads every act on a gap off evs, and each gap's mint position and stranded state.
func actsOnGaps(evs []*record.Event) (acts gapActs, minted map[string]int, superseded map[string]bool) {
	acts, minted, superseded = gapActs{}, map[string]int{}, map[string]bool{}
	proofAnswers := map[string]string{} // proof id → the gap it answers
	anchored := map[string][]string{}   // inquiry id → the gaps minted about it
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
		if b, ok := recordpb.BodyAs[*recordpb.Proof](e); ok {
			proofAnswers[b.GetProofId()] = b.GetAnswers()
			add(seat, i, b.GetAnswers())
		}
		if b, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
			add(seat, i, proofAnswers[b.GetProofId()])
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

// remandRulings is the position of the bench's latest remand of each gap.
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
	remanded := remandRulings(b.Events)
	minter := map[string]string{}
	for _, e := range b.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			if _, seen := minter[m.GetGapId()]; !seen {
				minter[m.GetGapId()] = e.GetSeatId()
			}
		}
	}
	// turnCame is whether blue recorded an act on gap after lens's latest act there, its mint and the
	// gap's latest remand ruling.
	turnCame := func(lens, gap string) bool {
		last := max(minted[gap], remanded[gap])
		for _, a := range acts[gap] {
			if a.seat == lens {
				last = max(last, a.at)
			}
		}
		for _, a := range acts[gap] {
			if a.at > last && seatclass.RoleOfSeat(a.seat) == "blue" {
				return true
			}
		}
		return false
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
	turns, waits := 0, 0
	for gap, seats := range engaged {
		lens := minter[gap]
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
			r.noteActErr(fmt.Sprintf("lens-turn oracle: the plan engages blue on %s and not its minting lens %s, whose turn has come — blue acted on %s since the lens's last act there", gap, lens, gap))
			return
		case came:
			turns++
		default:
			waits++
		}
	}
	r.mu.Lock()
	r.lensTurns += turns
	r.lensWaits += waits
	r.mu.Unlock()
}
