package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A FAILED CHECK IS A PASS BLOCKER. Two states of the anchor ledger (anchorchecks.go) are a check
// that was made and failed: a proof whose sentence a lens's standing judgement calls unsound, and a
// citation whose every check since its sentence last changed read nothing — its source was
// unreachable, or the copy read was not the work.
//
// THE PROOF RULING (gblock, 2026-10-24): "The judging lens is readied to mint its gap, and blue to
// fix or retire the proof. It clears when the proof is re-run sound or retired." universe-m32 ended
// UNVERIFIED with pass_permitted true: the computation lens judged a proof's sentence unsound, minted
// no gap, and retired; nothing readied anybody for it.
//
// THE CITATION RULING (gblock, 2026-10-25): "A citation that isn't checked is a lie in waiting - the
// doc is flawed, and red should demand the citation withdrawn or text changed or *something. Blocks
// lens and pass. Citation checking isn't optional. Ever." is-91-prime-b6 PASSed over two citations
// whose only check was `unreachable`: the evidence lens's gap was closed repaired once blue cited a
// mirror, and the two dead anchors stayed in the report.
//
// ONE SHAPE FOR BOTH. The seat whose check failed — the judge, or the first reader since the edit
// that found the source unreadable — owes a gap about the anchor (`about_kind` anchor, `about_ref`
// it), outside its mint budget, which carries the defect to blue with a grade and a fix; dispatch
// readies it until the gap exists, then blue. The gap answers the seat's debt, NOT the blocker: the
// blocker holds the PASS until the anchor is checked and holds, is changed by an edit (which leaves
// it unchecked, owed a fresh check), or is retired — and the gap cannot close while it stands
// (requireFailedCheckAnswered), so a gap closed "repaired" cannot stand over the anchor it carries.
// No bench ruling clears it either: the record's state of the anchor is not a dispute.

// failedCheck is one standing failed check of one anchor.
type failedCheck struct {
	anchor string
	// proof is whether the anchor is a proof's; sha is then its proof_sha, "" for a citation.
	proof bool
	sha   string
	// owner is the seat whose check failed — a proof's judge, a citation's first reader since the
	// edit — and since that check's events.id.
	owner string
	since int64
	// answeredBy is the gap owner minted about this anchor from the check's sitting on, "" until it
	// has; answeredAt is that mint's events.id.
	answeredBy string
	answeredAt int64
	// answers is every gap owner minted about this anchor from the check's sitting on, in record
	// order; answeredBy is the first.
	answers []string
}

// failedChecksOf is every standing failed check on the record, in the order the record placed the
// anchors and, within a proof, its standing unsound judgements. win supplies each event's events.id
// and sitting.
func failedChecksOf(evs []*Event, win WindowIndex) []failedCheck {
	type aboutMint struct {
		gap, seat string
		at        int64
	}
	mints := map[string][]aboutMint{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR {
			mints[m.GetAboutRef()] = append(mints[m.GetAboutRef()], aboutMint{gap: m.GetGapId(), seat: e.GetSeatId(), at: win.Of(e).ID})
		}
	}
	var out []failedCheck
	add := func(s *anchorStanding, c anchorCheck) {
		f := failedCheck{anchor: s.anchor, proof: s.proof, sha: s.sha, owner: c.seat, since: c.id}
		// A GAP MINTED IN THE CHECK'S OWN SITTING ANSWERS IT, before or after the check: the seat read
		// the same sentence for both. One from an earlier sitting carried an earlier check of a
		// sentence an edit has since replaced.
		from := c.sitting
		if from == 0 {
			from = c.id
		}
		for _, m := range mints[s.anchor] {
			if m.seat == c.seat && m.at > from {
				if f.answeredBy == "" {
					f.answeredBy, f.answeredAt = m.gap, m.at
				}
				f.answers = append(f.answers, m.gap)
			}
		}
		out = append(out, f)
	}
	for _, s := range anchorChecksOf(evs, win).order {
		switch s.state() {
		case stateUnsound:
			for _, c := range s.standing() {
				if c.soundness == recordpb.Soundness_SOUNDNESS_UNSOUND {
					add(s, c)
				}
			}
		case stateUnread:
			for _, c := range s.checks {
				if s.sinceEdit(c) {
					add(s, c)
					break
				}
			}
		}
	}
	return out
}

// answersOwnFailedCheck is the anchor a mint by seatID about kind/ref answers: seatID's own standing
// failed check of that anchor, which no gap answers yet. "" where it answers none.
func answersOwnFailedCheck(evs []*Event, win WindowIndex, seatID string, kind recordpb.AboutKind, ref string) string {
	if kind != recordpb.AboutKind_ABOUT_KIND_ANCHOR {
		return ""
	}
	for _, f := range failedChecksOf(evs, win) {
		if f.anchor == ref && f.owner == seatID && f.answeredBy == "" {
			return f.anchor
		}
	}
	return ""
}

// A REPEAT FAILED CHECK IS REFUSED (gblock, 2026-10-24, for the proof): a lens whose failed check of
// an anchor stands — its `unsound` judgement, or its reading that read nothing — and whose gap about
// the anchor, minted from that check's sitting on, is open, is refused the same failed check again,
// and the refusal names the gap.
//
// UNCHANGED IS THE LEDGER'S STANDING: the check stands exactly while the report carries the sentence
// it checked. An edit naming the anchor leaves it unchecked, and a check after that is a fresh one,
// admitted; a retire takes the anchor out. The open gap already carries the failure to blue; a second
// one of the same sentence would owe a second gap for the defect the open one holds — the release-gate
// fuzz minted ~490 such gaps per 40 runs. A check that holds is never refused here: it clears the
// blocker. Another lens's check is its own.
//
// Replay exempts it: an archived check is what a seat DID, recorded before this refusal existed.
func requireNotARepeatFailedCheck(q recordsql.Querier, seatID, anchor string, fails bool) error {
	if Migrating || !fails {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's checks could not be read, so whether %s already stands as a failed check by %s cannot be told: %v", anchor, seatID, err)
	}
	for _, f := range failedChecksOf(evs, win) {
		if f.anchor != anchor || f.owner != seatID {
			continue
		}
		for _, g := range f.answers {
			closed, err := gapState(q, g)
			if err != nil {
				return err
			}
			if closed {
				continue
			}
			if f.proof {
				return feov.Errorf(feov.Conflict,
					"you judged the sentence proof %s backs unsound at event %d, the sentence has not changed since that judgement, and gap %s, which you minted about the proof, is still open — "+
						"that gap stands for the judgement, so `--as unsound` is not recorded again and no second gap is owed. Nothing is recorded. "+
						"The judgement and its gap stand until blue edits the sentence, which leaves the proof unchecked for a fresh judgement, or retires the proof; "+
						"`--as sound`, where your reading of the sentence has changed, is recorded and clears the judgement",
					f.anchor, f.since, g)
			}
			return feov.Errorf(feov.Conflict,
				"your check of citation %s at event %d read nothing of its source, the sentence has not changed since, and gap %s, which you minted about the citation, is still open — "+
					"that gap stands for the unread citation, so a reading that reads nothing is not recorded again and no second gap is owed. Nothing is recorded. "+
					"The citation and its gap stand until blue retires the citation or edits its sentence; a reading of the source — any outcome but `unreachable`, of a copy that is the work — is recorded and clears it",
				f.anchor, f.since, g)
		}
	}
	return nil
}

// requireFailedCheckAnswered refuses the close of a gap that carries a failed check still standing
// (failedChecksOf): the gap is the failure's carrier to blue, and a closure — repaired, accepted, not
// a defect — while the record still holds the anchor unread or unsound would say the defect left the
// report when it did not. b6's G3 closed repaired over two dead anchors the report still cited. A
// close that hands the defect to a live successor (`--superseded-by`) moves it rather than ending it,
// and is admitted. Replay exempts it: an archived close is what a seat DID.
func requireFailedCheckAnswered(q recordsql.Querier, b *recordpb.Close) error {
	if Migrating || b.Successor != nil || b.CarriedFrom != nil {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's checks could not be read, so whether gap %s carries a failed check that still stands cannot be told: %v", b.GetGapId(), err)
	}
	for _, f := range failedChecksOf(evs, win) {
		for _, g := range f.answers {
			if g != b.GetGapId() {
				continue
			}
			if f.proof {
				return feov.Errorf(feov.Conflict,
					"gap %s carries proof %s, which %s's judgement at event %d still calls unsound, and the report still carries it — the defect has not left the report, so the gap does not close. Nothing is recorded. "+
						"It closes once the proof is judged sound again, blue edits its sentence (the proof is then owed a fresh judgement) or blue retires it; a close that hands the defect on names its successor with `--superseded-by`",
					g, f.anchor, f.owner, f.since)
			}
			return feov.Errorf(feov.Conflict,
				"gap %s carries citation %s, whose source no check since its sentence last changed has read (%s's check at event %d read nothing), and the report still cites it — the defect has not left the report, so the gap does not close. Nothing is recorded. "+
					"It closes once blue retires the citation (its sentence may stay), blue edits its sentence (the citation is then owed a fresh check) or a lens reads its source; a close that hands the defect on names its successor with `--superseded-by`",
				g, f.anchor, f.owner, f.since)
		}
	}
	return nil
}

// ownFailedCheckAnswers is the gaps seatID minted that answer a standing failed check of its own:
// outside its mint budget while open (lensMintStanding).
func ownFailedCheckAnswers(evs []*Event, win WindowIndex, seatID string) map[string]bool {
	out := map[string]bool{}
	for _, f := range failedChecksOf(evs, win) {
		if f.owner == seatID && f.answeredBy != "" {
			out[f.answeredBy] = true
		}
	}
	return out
}
