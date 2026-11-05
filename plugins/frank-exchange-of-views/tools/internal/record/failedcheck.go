package record

import (
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A FAILED CHECK IS A PASS BLOCKER. Three states of the anchor ledger (anchorchecks.go) are a check
// that was made and failed: a computation whose sentence a lens's standing judgement calls unsound, a
// citation whose every check since its sentence last changed read nothing — its source was
// unreachable, or the copy read was not the work — and a citation a lens's standing reading finds
// against: it read the work and the source says otherwise than the sentence, or not at all.
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
// A FOUND-AGAINST CHECK OWES NOTHING OF ITS READER (gblock, 2026-10-30, F4 and F8). Its gap is
// minted in the reading's own act and named by the reading's `carried_by` (AppendReading), followed
// through its successors; one standing reading is one check, so an anchor read against by two lenses
// holds two. It readies blue alone. It clears by the ledger's rule — its reader's own later reading
// of the sentence as it stands that backs it, blue's edit of the sentence, or blue's retire of the
// citation; another lens's backing reading beside it is a disagreement and clears nothing. The close
// of a gap that carries it is refused while it stands; a closing ruling of the bench closes that gap
// and the check still holds the PASS. Where no gap carries it — a migrated archived reading — nothing
// is owed either.

// CheckState is which failed check a failed_check blocker is: a citation whose source no check read,
// a computation a standing judgement calls unsound, or a citation a standing reading finds against.
// Owed-ness and every printed sentence key on it, never on whether a gap carries the check.
type CheckState string

const (
	CheckUnread       CheckState = "unread"
	CheckUnsound      CheckState = "unsound"
	CheckFoundAgainst CheckState = "found_against"
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
	// edit, the reader of a found-against reading — and since that check's events.id.
	owner string
	since int64
	// outcome is a found-against reading's outcome; UNSPECIFIED for the other states.
	outcome recordpb.SourceOutcome
	// answeredBy is the one gap that carries the check, "" while none does; answeredAt is that gap's
	// mint's events.id. For unread and unsound it is the first gap owner minted about this anchor from
	// the check's sitting on; for found_against, the OPEN gap of answers minted last, else the gap of
	// answers minted last.
	answeredBy string
	answeredAt int64
	// answers is every gap that carries the check: for unread and unsound, every gap owner minted
	// about this anchor from the check's sitting on, in record order; for found_against, the reading's
	// carried_by and every gap on its successor tree (carrierTree).
	answers []string
}

// owed reports whether the check's own seat owes a gap for it: an unread or unsound check no gap
// carries. A found-against check is never owed by its reader, carried or not — its gap is minted at
// the reading, and where none is named (a migrated archived reading) no mint answers it either.
func (f failedCheck) owed() bool { return f.state != CheckFoundAgainst && f.answeredBy == "" }

// failedChecksOf is every standing failed check on the record, in the order the record placed the
// anchors and, within an anchor, its standing failing checks. win supplies each event's events.id and
// sitting; open is the gap view's open set, read in the same snapshot as evs — a closing ruling of
// the bench closes a gap and writes no Close, so open-ness is never read off Close events.
func failedChecksOf(evs []*Event, win WindowIndex, open map[string]bool) []failedCheck {
	type aboutMint struct {
		gap, seat string
		at        int64
	}
	mints := map[string][]aboutMint{}
	mintAt := map[string]int64{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			mintAt[m.GetGapId()] = win.Of(e).ID
			if m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR {
				mints[m.GetAboutRef()] = append(mints[m.GetAboutRef()], aboutMint{gap: m.GetGapId(), seat: e.GetSeatId(), at: win.Of(e).ID})
			}
		}
	}
	var succ map[string][]string
	var out []failedCheck
	addFoundAgainst := func(s *anchorStanding, c anchorCheck) {
		v, _ := recordpb.BodyAs[*recordpb.Verify](c.ev)
		f := failedCheck{anchor: s.anchor, state: CheckFoundAgainst, owner: c.seat, since: c.id, outcome: v.GetOutcome()}
		if g := v.GetCarriedBy(); g != "" {
			if succ == nil {
				succ = successorsOf(evs)
			}
			f.answers, f.answeredBy = carrierTree(g, succ, mintAt, open)
			f.answeredAt = mintAt[f.answeredBy]
		}
		out = append(out, f)
	}
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
		case stateFoundAgainst:
			for _, c := range s.standing() {
				if !c.holds {
					addFoundAgainst(s, c)
				}
			}
		}
	}
	return out
}

// carrierTree is the gaps that carry a reading whose carried_by is g: g and every gap on its
// successor tree — each gap's close's successor and every gap whose mint supersedes it
// (successorsOf), transitively, each visited once — and the one that names the check: the OPEN gap of
// them minted last, else the gap of them minted last. A carrier closed into a successor hands the
// check to it, so no successor closes out from under a reading that stands.
func carrierTree(g string, succ map[string][]string, mintAt map[string]int64, open map[string]bool) (answers []string, by string) {
	seen := map[string]bool{g: true}
	answers = []string{g}
	for i := 0; i < len(answers); i++ {
		for _, s := range succ[answers[i]] {
			if !seen[s] {
				seen[s] = true
				answers = append(answers, s)
			}
		}
	}
	last := func(only bool) string {
		pick := ""
		for _, a := range answers {
			if only && !open[a] {
				continue
			}
			if pick == "" || mintAt[a] > mintAt[pick] {
				pick = a
			}
		}
		return pick
	}
	if by = last(true); by == "" {
		by = last(false)
	}
	return answers, by
}

// openGapIDsAt is the gap view's open set, asked of q.
func openGapIDsAt(q recordsql.Querier) (map[string]bool, error) {
	out := map[string]bool{}
	if noRecord(q) {
		return out, nil
	}
	rows, err := q.Query(`SELECT "gap_id" FROM "gap" WHERE "open"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		out[g] = true
	}
	return out, rows.Err()
}

// failedChecksAt is failedChecksOf asked of q: the stream and the gap view's open set on one snapshot.
// held is the stream q holds where the reader already took it (heldStream), nil where none is held.
func failedChecksAt(q recordsql.Querier, held *heldStream) ([]failedCheck, error) {
	evs, win, err := held.all(q)
	if err != nil {
		return nil, err
	}
	open, err := openGapIDsAt(q)
	if err != nil {
		return nil, err
	}
	return failedChecksOf(evs, win, open), nil
}

// answersOwnFailedCheck is the anchor a mint by seatID about kind/ref answers: seatID's own standing
// unread or unsound check of that anchor, which no gap answers yet. "" where it answers none. A
// found-against check is never answered this way: its gap is the reading's own, outside the budget by its act.
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
		if f.anchor != anchor || f.owner != seatID || f.state == CheckFoundAgainst {
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

// requireNotARepeatReading refuses a reading that repeats one of the seat's own: over a citation, a
// reading that reads nothing meets the unread arm (requireNotARepeatFailedCheck); a reading that
// mints (ReadingGapClasses non-empty) by the same lens, of the same sentence from the same source —
// the anchor for `verify --anchor`, the url and claim for `verify --url` — with the same outcome as
// an earlier such reading of its own, the sentence unchanged, while a gap of that earlier reading's
// carrier tree is open, is refused, naming the earlier reading's event and its gap. A reading with a
// different outcome is not a repeat. Within one sitting requireNotARetriedReading answers a
// found-source retry first.
func requireNotARepeatReading(q recordsql.Querier, seatID string, v *recordpb.Verify) error {
	if v.GetAnchor() != "" && !VerifyReads(v) {
		return requireNotARepeatFailedCheck(q, seatID, v.GetAnchor(), true)
	}
	if len(ReadingGapClasses(v)) == 0 {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err == nil {
		var open map[string]bool
		if open, err = openGapIDsAt(q); err == nil {
			return repeatReadingRefusal(evs, win, open, seatID, v)
		}
	}
	return feov.Wrap(feov.Conflict, err, "the record's readings could not be read, so whether %s already stands on this sentence cannot be told: %v", seatID, err)
}

// repeatReadingRefusal is requireNotARepeatReading's fold, nil where v repeats nothing. Its refusal is
// worded by form: a reading of a citation finds the source against the sentence; a reading of a
// source the lens found may back it, so it says only that the gap carries it.
func repeatReadingRefusal(evs []*Event, win WindowIndex, open map[string]bool, seatID string, v *recordpb.Verify) error {
	refuse := func(what string, at int64, gap string) error {
		return feov.Errorf(feov.Conflict,
			"your reading of %s at event %d finds the source against this sentence (%s), the sentence has not changed, and gap %s carries that reading and is open — nothing is recorded; the reading and its gap stand. "+
				"A reading of yours with another outcome is recorded",
			what, at, recordpb.Word(v.GetOutcome()), gap)
	}
	refuseFound := func(url string, at int64, gap string) error {
		return feov.Errorf(feov.Conflict,
			"your reading of %s for this sentence at event %d (%s) stands, gap %s carries it and is open, and the sentence has not changed — nothing is recorded; the reading and its gap stand. "+
				"A reading of yours with another outcome is recorded",
			url, at, recordpb.Word(v.GetOutcome()), gap)
	}
	if v.GetAnchor() != "" {
		for _, f := range failedChecksOf(evs, win, open) {
			if f.state != CheckFoundAgainst || f.anchor != v.GetAnchor() || f.owner != seatID || f.outcome != v.GetOutcome() {
				continue
			}
			if slices.ContainsFunc(f.answers, func(g string) bool { return open[g] }) {
				return refuse("citation "+f.anchor, f.since, f.answeredBy)
			}
		}
		return nil
	}
	mintAt := map[string]int64{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			mintAt[m.GetGapId()] = win.Of(e).ID
		}
	}
	var succ map[string][]string
	claim := strings.TrimSpace(v.GetClaim())
	for _, e := range evs {
		prev, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || e.GetSeatId() != seatID || prev.GetAnchor() != "" || prev.GetCarriedBy() == "" ||
			strings.TrimSpace(prev.GetClaim()) != claim || prev.GetUrl() != v.GetUrl() ||
			prev.GetOutcome() != v.GetOutcome() || len(ReadingGapClasses(prev)) == 0 {
			continue
		}
		if succ == nil {
			succ = successorsOf(evs)
		}
		answers, by := carrierTree(prev.GetCarriedBy(), succ, mintAt, open)
		if slices.ContainsFunc(answers, func(g string) bool { return open[g] }) {
			return refuseFound(v.GetUrl(), win.Of(e).ID, by)
		}
	}
	return nil
}

// demandAfterClear is what every text that says a failed check clears and the gap that carries it may
// then close says beside it: the clearing admits the close of a `source` or `computation` gap only
// under its live-answer demand, which the close verb asks first (recordpb.DemandClause).
const demandAfterClear = "The check clearing admits the close only under the gap's live-answer demand, which the close asks first: " + recordpb.DemandClause

// requireFailedCheckAnswered refuses the close of a gap that carries a failed check still standing
// (failedChecksOf): the gap is the failure's carrier to blue, and a closure — repaired, accepted, not
// a defect — while the record still holds the anchor unread, unsound or found against would say the
// defect left the report when it did not. b6's G3 closed repaired over two dead anchors the report
// still cited. A found-against check's carriers are every gap of its reading's carrier tree, so no
// branch closes out from under it. A successor minted superseding the gap moves the defect rather
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
			if f.state == CheckFoundAgainst {
				return feov.Errorf(feov.Conflict,
					"gap %s carries citation %s, which %s's reading at event %d finds the source against, and the report still carries that sentence — the defect has not left the report, so this close is refused. Nothing is recorded. "+
						"The check clears once %s's own later reading of the sentence as it stands backs it, blue edits the sentence, or blue retires the citation; another lens's backing reading is a disagreement and clears nothing. %s; a gap that carries the defect on is minted with --supersedes naming this one, which closes it, and while nothing answers this one's `source` or `computation` demand the successor carries its check kind and the demand with it",
					g, f.anchor, ownerOrSeat(f.owner), f.since, ownerOrSeat(f.owner), demandAfterClear)
			}
			if f.computation {
				return feov.Errorf(feov.Conflict,
					"gap %s carries computation %s, which %s's judgement at event %d still calls unsound, and the report still carries it — the defect has not left the report, so this close is refused. Nothing is recorded. "+
						"The check clears once the computation is judged sound again, blue edits its sentence (the computation is then owed a fresh judgement) or blue retires it. %s; a gap that carries the defect on is minted with --supersedes naming this one, which closes it, and while nothing answers this one's `source` or `computation` demand the successor carries its check kind and the demand with it",
					g, f.anchor, f.owner, f.since, demandAfterClear)
			}
			return feov.Errorf(feov.Conflict,
				"gap %s carries citation %s, whose source no check since its sentence last changed has read (%s's check at event %d read nothing), and the report still cites it — the defect has not left the report, so this close is refused. Nothing is recorded. "+
					"The check clears once blue retires the citation (its sentence may stay), blue edits its sentence (the citation is then owed a fresh check) or a lens reads its source. %s; a gap that carries the defect on is minted with --supersedes naming this one, which closes it, and while nothing answers this one's `source` or `computation` demand the successor carries its check kind and the demand with it",
				g, f.anchor, f.owner, f.since, demandAfterClear)
		}
	}
	return nil
}

// ownFailedCheckAnswers is the gaps seatID minted that answer a standing unread or unsound check of
// its own: outside its mint budget while open (lensMintStanding). A found-against check's carrier is
// never among them — a reading's gap is outside the budget by its act (carrier_minted), so a held
// gap another lens minted cannot enter this lens's set, and no mint is exempted twice.
func ownFailedCheckAnswers(fs []failedCheck, seatID string) map[string]bool {
	out := map[string]bool{}
	for _, f := range fs {
		if f.owner == seatID && f.state != CheckFoundAgainst && f.answeredBy != "" {
			out[f.answeredBy] = true
		}
	}
	return out
}

// ownerOrSeat is a failed check's reader, or the fallback where the record names none.
func ownerOrSeat(owner string) string {
	if owner == "" {
		return failedChecker
	}
	return owner
}
