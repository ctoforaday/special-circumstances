package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A FAILED CHECK IS A PASS BLOCKER. Two states of the anchor ledger (anchorchecks.go) are a check
// that was made and failed: a computation whose sentence a lens's standing judgement calls unsound, and
// a citation whose every check since its sentence last changed read nothing — its source was
// unreachable, or the copy read was not the work. A citation a lens's reading finds against is no
// state of the ledger: the reading is its lens's debt (readingDebtsOf) until the gap it mints on it
// stands, and the gap's materiality then holds the PASS (gblock, 2026-11-06, #1727).
//
// THE COMPUTATION RULING (gblock, 2026-10-24): "The judging lens is readied to mint its gap, and blue to
// fix or retire the proof. It clears when the proof is re-run sound or retired." universe-m32 ended
// UNVERIFIED with pass_permitted true: the computation lens judged a computation's sentence unsound, minted
// no gap, and retired; nothing readied anybody for it.
//
// THE CITATION RULING (gblock, 2026-10-25): "A citation that isn't checked is a lie in waiting - the
// doc is flawed, and red should demand the citation withdrawn or text changed or *something. Blocks
// lens and pass. Citation checking isn't optional. Ever." is-91-prime-b6 PASSed over two citations
// whose only check was `unreachable`: the evidence lens's gap was closed repaired once blue cited a
// mirror, and the two dead anchors stayed in the report.
//
// ONE SHAPE FOR UNREAD AND UNSOUND. The seat whose check failed — the judge, or the first reader since
// the edit that found the source unreadable — owes a gap about the anchor (`about_kind` anchor,
// `about_ref` it), outside its mint budget, which carries the defect to blue with a grade and a fix;
// dispatch readies it until the gap exists, then blue. The gap answers the seat's debt, NOT the
// blocker: the blocker holds the PASS until the anchor is checked and holds, is changed by an edit
// (which leaves it unchecked, owed a fresh check), or is retired — and the gap's close is refused
// while it stands (requireFailedCheckAnswered), so a gap closed "repaired" cannot stand over the
// anchor it carries. No bench ruling clears it either: the record's state of the anchor is not a
// dispute.
//
// CheckState is which failed check a failed_check blocker is: a citation whose source no check read,
// or a computation a standing judgement calls unsound. Every printed sentence keys on it.
type CheckState string

const (
	CheckUnread  CheckState = "unread"
	CheckUnsound CheckState = "unsound"
)

// failedCheck is one standing failed check of one anchor.
type failedCheck struct {
	anchor string
	// computation is whether the anchor is a computation's; sha is then its bundle_sha, "" for a citation.
	computation bool
	sha         string
	// state is which failed check this is.
	state CheckState
	// owner is the seat whose check failed — a computation's judge, a citation's first reader since the
	// edit — and since that check's events.id.
	owner string
	since int64
	// answeredBy is the one gap that carries the check, "" while none does — the first gap owner
	// minted about this anchor from the check's sitting on; answeredAt is that gap's mint's events.id.
	answeredBy string
	answeredAt int64
	// answers is every gap that carries the check: every gap owner minted about this anchor from the
	// check's sitting on, in record order.
	answers []string
}

// owed reports whether the check's own seat owes a gap for it: a check no gap carries.
func (f failedCheck) owed() bool { return f.answeredBy == "" }

// failedChecksOf is every standing failed check on the record, in the order the record placed the
// anchors and, within an anchor, its standing failing checks. win supplies each event's events.id and
// sitting.
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
		f := failedCheck{anchor: s.anchor, computation: s.computation, sha: s.sha, state: CheckUnread, owner: c.seat, since: c.id}
		if s.computation {
			f.state = CheckUnsound
		}
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

// failedChecksAt is failedChecksOf asked of q. held is the stream q holds where the reader already
// took it (heldStream), nil where none is held.
func failedChecksAt(q recordsql.Querier, held *heldStream) ([]failedCheck, error) {
	evs, win, err := held.all(q)
	if err != nil {
		return nil, err
	}
	return failedChecksOf(evs, win), nil
}

// answersOwnFailedCheck is the anchor a mint by seatID about kind/ref answers: seatID's own standing
// unread or unsound check of that anchor, which no gap answers yet. "" where it answers none.
func answersOwnFailedCheck(fs []failedCheck, seatID string, kind recordpb.AboutKind, ref string) string {
	if kind != recordpb.AboutKind_ABOUT_KIND_ANCHOR {
		return ""
	}
	for _, f := range fs {
		if f.anchor == ref && f.owner == seatID && f.owed() {
			return f.anchor
		}
	}
	return ""
}

// A REPEAT FAILED CHECK IS REFUSED (gblock, 2026-10-24, for the computation): a lens whose failed check of
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
func requireNotARepeatFailedCheck(q recordsql.Querier, seatID, anchor string, fails bool) error {
	if !fails {
		return nil
	}
	fs, err := failedChecksAt(q, nil)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's checks could not be read, so whether %s already stands as a failed check by %s cannot be told: %v", anchor, seatID, err)
	}
	for _, f := range fs {
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
			if f.computation {
				return feov.Errorf(feov.Conflict,
					"you judged the sentence computation %s backs unsound at event %d, the sentence has not changed since that judgement, and gap %s, which you minted about the computation, is still open — "+
						"that gap stands for the judgement, so `--as unsound` is not recorded again and no second gap is owed. Nothing is recorded. "+
						"The judgement and its gap stand until blue edits the sentence, which leaves the computation unchecked for a fresh judgement, or retires the computation; "+
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

// requireNotARepeatReading refuses a reading of a citation that reads nothing (an unread check) where
// the seat's own such check stands with its gap open: the unread arm (requireNotARepeatFailedCheck).
// A reading that reads the work is recorded: a repeated finds-against reading replaces the seat's
// standing check, and the same debt stands (readingDebtsOf, one per standing reading per seat).
// Within one sitting requireNotARetriedReading answers a found-source retry first.
func requireNotARepeatReading(q recordsql.Querier, seatID string, v *recordpb.Verify) error {
	if v.GetAnchor() != "" && !VerifyReads(v) {
		return requireNotARepeatFailedCheck(q, seatID, v.GetAnchor(), true)
	}
	return nil
}

// demandAfterClear is what every text that says a failed check clears and the gap that carries it may
// then close says beside it: the clearing admits the close of a `source` or `computation` gap only
// under its live-answer demand, which the close verb asks first (recordpb.DemandClause).
const demandAfterClear = "The check clearing admits the close only under the gap's live-answer demand, which the close asks first: " + recordpb.DemandClause

// requireFailedCheckAnswered refuses the close of a gap that carries a failed check still standing
// (failedChecksOf): the gap is the failure's carrier to blue, and a closure — repaired, accepted, not
// a defect — while the record still holds the anchor unread or unsound would say the defect left the
// report when it did not. b6's G3 closed repaired over two dead anchors the report still cited. A
// successor minted superseding the gap moves the defect rather
// than ending it, and closes the gap in that act (supersession.go).
func requireFailedCheckAnswered(q recordsql.Querier, b *recordpb.Close) error {
	if b.CarriedFrom != nil {
		return nil
	}
	fs, err := failedChecksAt(q, nil)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the record's checks could not be read, so whether gap %s carries a failed check that still stands cannot be told: %v", b.GetGapId(), err)
	}
	for _, f := range fs {
		for _, g := range f.answers {
			if g != b.GetGapId() {
				continue
			}
			if f.computation {
				return feov.Errorf(feov.Conflict,
					"gap %s carries computation %s, which %s's judgement at event %d still calls unsound, and the report still carries it — the defect has not left the report, so this close is refused. Nothing is recorded. "+
						"The check clears once the computation is judged sound again, blue edits its sentence (the computation is then owed a fresh judgement) or blue retires it. %s; a gap that carries the defect on is minted with --supersedes naming this one, which closes it, and while nothing answers this one's `source` or `computation` demand the successor carries its check kind and the demand with it, and a retire of the anchor, or its own seat's check of it again that holds, made after the successor's mint, meets the successor's demand",
					g, f.anchor, f.owner, f.since, demandAfterClear)
			}
			return feov.Errorf(feov.Conflict,
				"gap %s carries citation %s, whose source no check since its sentence last changed has read (%s's check at event %d read nothing), and the report still cites it — the defect has not left the report, so this close is refused. Nothing is recorded. "+
					"The check clears once blue retires the citation (its sentence may stay), blue edits its sentence (the citation is then owed a fresh check) or a lens reads its source. %s; a gap that carries the defect on is minted with --supersedes naming this one, which closes it, and while nothing answers this one's `source` or `computation` demand the successor carries its check kind and the demand with it, and a retire of the anchor, or its own seat's check of it again that holds, made after the successor's mint, meets the successor's demand",
				g, f.anchor, f.owner, f.since, demandAfterClear)
		}
	}
	return nil
}

// ownFailedCheckAnswers is the gaps seatID minted that answer a standing unread or unsound check of
// its own: outside its mint budget while open (lensMintStanding).
func ownFailedCheckAnswers(fs []failedCheck, seatID string) map[string]bool {
	out := map[string]bool{}
	for _, f := range fs {
		if f.owner == seatID && f.answeredBy != "" {
			out[f.answeredBy] = true
		}
	}
	return out
}
