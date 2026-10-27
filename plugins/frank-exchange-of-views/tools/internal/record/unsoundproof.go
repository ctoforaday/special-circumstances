package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

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
	// answers is every gap the judge minted about this proof from the judgement's sitting on, in
	// record order; answeredBy is the first.
	answers []string
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
					if u.answeredBy == "" {
						u.answeredBy, u.answeredAt = m.gap, m.at
					}
					u.answers = append(u.answers, m.gap)
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

// A REPEAT UNSOUND JUDGEMENT IS REFUSED (gblock, 2026-10-24): a lens that judges a proof's sentence
// unsound again, while its earlier `unsound` of that sentence stands and a gap it minted about the
// proof from that judgement's sitting on is open, is refused, and the refusal names the gap.
//
// UNCHANGED IS THE LEDGER'S STANDING (anchorStanding.stands): the judgement stands exactly while the
// report carries the sentence it judged. An edit naming the anchor reopens it, and a judgement after
// that is a fresh one, admitted; a retire takes the anchor out. The open gap already carries the
// judgement to blue. A second judgement of the same sentence would owe a second gap for the defect
// the open one holds (answeredBy reads from the judgement's own sitting on) — the release-gate fuzz
// minted ~490 such gaps per 40 runs. A `sound`
// re-judgement is never refused here: it clears the blocker. Another lens's judgement is its own.
//
// Replay exempts it: an archived judgement is what a seat DID, recorded before this refusal existed.
func requireNotARepeatUnsound(q recordsql.Querier, seatID string, b *recordpb.Reproduce) error {
	if Migrating || b.GetSoundness() != recordpb.Soundness_SOUNDNESS_UNSOUND {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's proof judgements could not be read, so whether %s already stands judged unsound by %s cannot be told: %v", b.GetProofId(), seatID, err)
	}
	for _, u := range unsoundProofsOf(evs, win) {
		if u.anchor != b.GetProofId() || u.judge != seatID {
			continue
		}
		for _, g := range u.answers {
			closed, err := gapState(q, g)
			if err != nil {
				return err
			}
			if closed {
				continue
			}
			return feov.Errorf(feov.Conflict,
				"you judged the sentence proof %s backs unsound at event %d, the sentence has not changed since that judgement, and gap %s, which you minted about the proof, is still open — "+
					"that gap stands for the judgement, so `--as unsound` is not recorded again and no second gap is owed. Nothing is recorded. "+
					"The judgement and its gap stand until blue edits the sentence, which reopens the proof for a fresh judgement, or retires the proof; "+
					"`--as sound`, where your reading of the sentence has changed, is recorded and clears the judgement",
				u.anchor, u.since, g)
		}
	}
	return nil
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
