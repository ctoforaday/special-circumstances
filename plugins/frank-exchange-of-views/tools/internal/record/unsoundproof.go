package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// A STANDING UNSOUND JUDGEMENT IS A PASS BLOCKER (gblock, 2026-10-24): "The judging lens is readied
// to mint its gap, and blue to fix or retire the proof. It clears when the proof is re-run sound or
// retired."
//
// universe-m32 ended UNVERIFIED with pass_permitted true: the computation lens judged a proof's
// sentence unsound, minted no gap, and retired; nothing readied anybody for it, and the chair would
// not PASS over a sentence the record called unsound while the gate admitted one.
//
// WHAT STANDS IS THE LEDGER'S (anchorchecks.go). A judgement stands while it is its seat's latest
// check of the sentence the report carries: the judging seat's own later `sound` answers it, another
// seat's never does, an edit that changes the sentence answers it by reopening the anchor (which is
// then the reopened blocker, owed a fresh judgement), and a retire takes the anchor out. So one
// blocker stands per proof and judging seat, and it clears exactly when the ruling says.
//
// THE GAP ANSWERS THE JUDGE'S DEBT, NOT THE BLOCKER. The judging lens owes a gap about the proof —
// `about_kind` proof, `about_ref` its anchor, a field the write checks — which carries the defect to
// blue with a grade and a fix. Once it is minted the lens owes nothing more on it, and the blocker
// still holds the PASS until the proof is judged sound again or retired.

// unsoundProof is one proof anchor a lens's standing judgement calls unsound.
type unsoundProof struct {
	anchor, sha string
	// judge is the seat whose standing check is unsound; since is that check's events.id.
	judge string
	since int64
	// answeredBy is the gap the judge minted about this proof from the judgement's sitting on, "" until
	// it has; answeredAt is that mint's events.id.
	answeredBy string
	answeredAt int64
}

// unsoundProofsOf is every standing unsound judgement on the record, in the order the record placed
// the anchors and, within an anchor, the judgements. win supplies each event's events.id and sitting.
func unsoundProofsOf(evs []*Event, win WindowIndex) []unsoundProof {
	type aboutMint struct {
		gap, seat string
		at        int64
	}
	mints := map[string][]aboutMint{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_PROOF {
			mints[m.GetAboutRef()] = append(mints[m.GetAboutRef()], aboutMint{gap: m.GetGapId(), seat: e.GetSeatId(), at: win.Of(e).ID})
		}
	}
	var out []unsoundProof
	for _, s := range anchorChecksOf(evs, win).order {
		if !s.proof {
			continue
		}
		for _, c := range s.standing() {
			if c.soundness != recordpb.Soundness_SOUNDNESS_UNSOUND {
				continue
			}
			u := unsoundProof{anchor: s.anchor, sha: s.sha, judge: c.seat, since: c.id}
			// A GAP MINTED IN THE JUDGEMENT'S OWN SITTING ANSWERS IT, before or after the reproduce: the
			// seat read the same sentence for both. One from an earlier sitting carried an earlier
			// judgement of a sentence an edit has since replaced.
			from := c.sitting
			if from == 0 {
				from = c.id
			}
			for _, m := range mints[s.anchor] {
				if m.seat == c.seat && m.at > from {
					u.answeredBy, u.answeredAt = m.gap, m.at
					break
				}
			}
			out = append(out, u)
		}
	}
	return out
}

// answersOwnUnsound is the proof a mint by seatID about kind/ref answers: seatID's own standing
// unsound judgement of that proof, which no gap answers yet. "" where it answers none.
func answersOwnUnsound(evs []*Event, win WindowIndex, seatID string, kind recordpb.AboutKind, ref string) string {
	if kind != recordpb.AboutKind_ABOUT_KIND_PROOF {
		return ""
	}
	for _, u := range unsoundProofsOf(evs, win) {
		if u.anchor == ref && u.judge == seatID && u.answeredBy == "" {
			return u.anchor
		}
	}
	return ""
}

// ownUnsoundAnswers is the gaps seatID minted that answer a standing unsound judgement of its own:
// outside its mint budget while open (lensMintStanding).
func ownUnsoundAnswers(evs []*Event, win WindowIndex, seatID string) map[string]bool {
	out := map[string]bool{}
	for _, u := range unsoundProofsOf(evs, win) {
		if u.judge == seatID && u.answeredBy != "" {
			out[u.answeredBy] = true
		}
	}
	return out
}
