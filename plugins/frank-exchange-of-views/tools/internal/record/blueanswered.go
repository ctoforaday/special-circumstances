package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// DID BLUE ANSWER THIS GAP, AND CAN THE RECORD SAY? (#1122)
//
// THE ANSWERING PREDICATE IS THE ONE EVERY OTHER READER USES. `BlueEdit.answers` is blue's CLAIM to
// have answered, validated at the write (requireGap refuses an unknown gap), and with `old != new` it
// is what `exchangesOf` counts as movement and what ManifestOwed counts as answered. An edit that
// changed a gap's sentence without naming the gap (WorkGapState.Edits) is not an answer: reading it
// would invent a third definition of "answered", disagreeing with the impasse fold and the manifest
// audit, and would accuse blue on any gap anchored by --about, which has no sentence to change. It
// gives the lens its turn all the same (lensTurnActsOf), and the item says which of the two it is.
//
// A GAP ANCHORED ON AN INQUIRY IS ANSWERED BY THAT INQUIRY'S FATE, NOT BY AN EDIT. A mint made
// with `--about-kind inquiry` names no report text, so no edit can be its answer: blue answers it by
// recording a fate on the inquiry it names. Read by edits alone, such a gap read "NO edit answers
// it" after blue had recorded the fate, and the lens had to read the inquiries to learn otherwise.
//
// A GAP WHOSE ANCHOR LEFT THE REPORT IS NOT UNANSWERED. Blue cutting the sentence is an act on it
// whether or not the edit named the gap, so a `gone` gap blue was engaged on is put to its lens to
// judge, never reported as silence.
//
// THE NOT-MEASURED ARM IS THE POINT OF THE WHOLE FUNCTION. "Blue has not answered" and "blue is
// still sitting and may yet answer" are the same silence on the record, and a lens handed the first
// where the second is true files a closing argument about a dispute that is still in progress.
// BlueSitting.Unresolved is that fact, and it is read here rather than re-derived.
//
// A REMAND IS ANSWERED ONLY AFTER IT (ruled by gblock 2026-10-27). The bench remanded the gap
// having heard every answer blue gave before the ruling, so none of them answers the remand: on
// universe-m35 the voice lens's list said "ANSWERED by an edit" beside the remand item while every
// answering edit predated the ruling. blueAnswersOf reads each gap's answers after its latest remand
// ruling, so this list, the lens's debt (lensOwesOf) and blue's kept-grade items agree.
//
// IT IS ITEMS ON THE LENS'S TURN (ruled by gblock 2026-10-31). A gap whose turn has not come and that
// the route does not engage its lens on (partiesOnRoute) has nothing here, and a docketed gap is off
// the list altogether (sittingOf): universe-m37's lenses closed docketed gaps from this list and kept
// grades on gaps whose turn had not come, because it read "open, and I minted it" while the plan read
// the route. And an answer the lens has read is no news: only what arrived after lensRead is named, so
// ANSWERED clears once the lens has audited that answer. A gap that is not material readies nobody,
// and its lens hears of blue's answer here all the same when it sits for anything else.
func blueAnswers(evs []*Event, win WindowIndex, gaps []WorkGapState, seatID string) []string {
	minted := mintedBy(evs)
	var mine []string
	open, gone, read, turn := map[string]bool{}, map[string]bool{}, map[string]int64{}, map[string]bool{}
	for _, g := range gaps {
		if g.Open && (g.lensDue || g.lensTurn || g.LocationState == LocationGone) && minted[g.ID] == seatID {
			mine = append(mine, g.ID)
			open[g.ID] = true
			gone[g.ID] = g.LocationState == LocationGone
			read[g.ID], turn[g.ID] = g.lensRead, g.lensTurn
		}
	}
	if len(mine) == 0 {
		return nil
	}
	answers := blueAnswersOf(evs, win, open, motionsAt(evs, win.IDs(evs), win))
	var out []string
	for _, g := range mine {
		a := answers[g]
		switch {
		case a.editedAt > read[g]:
			out = append(out, fmt.Sprintf("gap %s has been ANSWERED by a change to the report (an edit or a citation retire answering it) — re-audit it against the acceptance check you set at mint, and either close it or keep its grade with a regrade that states what your own check still fails — blue reads that beside the gap, and the bench in the transcript", g))
		case a.computedAt > read[g]:
			out = append(out, fmt.Sprintf("gap %s has been ANSWERED by a computation — blue recorded %s answering it; re-audit it against the acceptance check you set at mint, and either close it or keep its grade with a regrade that states what your own check still fails — blue reads that beside the gap, and the bench in the transcript", g, a.computation))
		case a.fateAt > read[g]:
			out = append(out, fmt.Sprintf("gap %s has been ANSWERED by an inquiry record — blue recorded the fate `%s` on %s, the inquiry the gap is anchored on; re-audit it against the acceptance check you set at mint, and either close it or keep its grade with a regrade that states what your own check still fails — blue reads that beside the gap, and the bench in the transcript", g, a.fate, a.inquiry))
		case a.pending:
			out = append(out, fmt.Sprintf("gap %s: whether blue has answered it is NOT MEASURED — a blue sitting engaged on it is open on the record, so nothing here is a report that no answer came", g))
		case a.engaged && gone[g] && a.at > read[g]:
			out = append(out, fmt.Sprintf("gap %s is open and its anchor is no longer in the report — judge whether the report as it stands meets your acceptance check, then close it, or keep its grade with a regrade that says what still fails", g))
		case turn[g]:
			// THE LENS'S TURN, NOT BLUE'S ANSWER: an edit answering another gap changed this gap's
			// sentence, and the row's `edited_since` lists it.
			out = append(out, fmt.Sprintf("gap %s: an edit changed its sentence since you last acted on it, and no act of blue's answers it — `edited_since` lists the edit; re-audit the sentence as it stands against the acceptance check you set at mint, and either close it or keep its grade with a regrade that states what your own check still fails", g))
		case a.engaged && a.inquiry != "":
			out = append(out, fmt.Sprintf("gap %s is open and NO inquiry record answers it since you last acted on it — it is anchored on %s, which an edit cannot answer, and blue has been engaged on it and its sitting closed without recording a fate on %s", g, a.inquiry, a.inquiry))
		case a.engaged:
			out = append(out, fmt.Sprintf("gap %s is open and NO edit or computation answers it since you last acted on it — blue has been engaged on it and its sitting closed without one", g))
		}
		// A GAP BLUE HAS NEVER BEEN ENGAGED ON GETS NOTHING, and the silence is correct: it is
		// waiting for a dispatch, which is the chair's business and not news about blue.
	}
	return out
}

// blueAnswer is what the record says of blue's answer to one gap: what blueAnswers words for the
// gap's lens, and when it last arrived, which its debt reads (lensOwesOf).
type blueAnswer struct {
	editedAt    int64  // the latest change to the report answering it: an edit, old != new, or a citation retire
	computedAt  int64  // the latest computation answering it
	computation string // that computation's id
	fateAt      int64  // the latest fate blue recorded on the inquiry it is anchored on, while engaged on it
	fate        string // that fate
	inquiry     string // the inquiry it is anchored on; "" for a gap anchored on report text
	pending     bool   // a blue sitting engaged on it is open on the record
	engaged     bool   // a blue sitting was engaged on it
	// at is the events.id of blue's latest answer to it: an edit answering it, a computation, a fate on
	// its inquiry, or the latest act of a closed sitting engaged on it — the rebuttal, or the silence
	// its lens states. 0 when blue has given none.
	at int64
}

// answered is whether blue has given an answer the record can state — an edit, a fate, or a closed
// sitting engaged on it. A sitting still open is no answer yet: whether it answers is not measured.
func (a blueAnswer) answered() bool { return a.at != 0 }

// blueAnswersOf is blue's answer to each gap of open, by gap: the one reading of it, which the gap's
// lens is handed (blueAnswers) and owes while it has not sat since (lensOwesOf). The inquiry a gap is
// anchored on is its mint's, as the gap view's about columns are. A remanded gap's answer is read
// after its latest remand ruling only (remandFloorsOf, off motions — motionsAt's fold of evs, which
// a caller that already holds it passes rather than folding twice): a sitting opened before it, and
// an act recorded before it, answers nothing the remand asks.
func blueAnswersOf(evs []*Event, win WindowIndex, open map[string]bool, motions []*Motion) map[string]blueAnswer {
	out := map[string]blueAnswer{}
	inquiryOf := map[string]string{} // gap -> the inquiry it is anchored on
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && open[m.GetGapId()] && m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
			inquiryOf[m.GetGapId()] = m.GetAboutRef()
		}
	}
	floor := remandFloorsOf(motions, open)
	after := func(g string, e *Event) bool { return win.Of(e).ID > floor[g] }
	at := func(g string, e *Event) {
		a := out[g]
		if id := win.Of(e).ID; id > a.at {
			a.at = id
		}
		out[g] = a
	}
	// WhileRunning, because this list is read by a seat that is sitting: a later reading is not the
	// question, and the latest blue sitting is exactly the one that may still be answering.
	for _, s := range BlueSittings(evs, win, WhileRunning) {
		opening := evs[s.opened]
		for _, g := range s.Engaged {
			if !open[g] || !after(g, opening) {
				continue
			}
			a := out[g]
			a.engaged, a.inquiry = true, inquiryOf[g]
			a.pending = a.pending || s.Unresolved
			out[g] = a
			if !s.Unresolved && len(s.Acts) > 0 {
				at(g, s.Acts[len(s.Acts)-1])
			}
		}
		for _, e := range s.Acts {
			if in, ok := recordpb.BodyAs[*recordpb.Inquiry](e); ok {
				for _, g := range s.Engaged {
					if q, anchored := inquiryOf[g]; anchored && open[g] && q == in.GetInquiryId() && after(g, e) {
						a := out[g]
						a.fate, a.fateAt = recordpb.Word(in.GetStatus()), win.Of(e).ID
						out[g] = a
						at(g, e)
					}
				}
				continue
			}
			if g := retireAnswers(e); g != "" && open[g] && after(g, e) {
				a := out[g]
				a.editedAt, a.inquiry = win.Of(e).ID, inquiryOf[g]
				out[g] = a
				at(g, e)
				continue
			}
			if c, ok := recordpb.BodyAs[*recordpb.Computation](e); ok {
				if g := c.GetAnswers(); g != "" && open[g] && after(g, e) {
					a := out[g]
					a.computedAt, a.computation, a.inquiry = win.Of(e).ID, c.GetComputationId(), inquiryOf[g]
					out[g] = a
					at(g, e)
				}
				continue
			}
			be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
			if !ok {
				continue
			}
			// old != new is the same substantive-movement test exchangesOf applies: an edit that
			// replaced a span with itself is an act and not an answer.
			if g := be.GetAnswers(); g != "" && open[g] && be.GetOld() != be.GetNew() && after(g, e) {
				a := out[g]
				a.editedAt, a.inquiry = win.Of(e).ID, inquiryOf[g]
				out[g] = a
				at(g, e)
			}
		}
	}
	return out
}

// retireAnswers is the gap a citation retire answers — the key BlueEdit.answers is on the other act
// that changes the report's text, and every reader of blue's answers reads both. "" for any other
// event, and for a retire answering no gap; a claim retire carries none (validate refuses one).
func retireAnswers(e *Event) string {
	if r, ok := recordpb.BodyAs[*recordpb.Retire](e); ok && r.GetClaimKept() {
		return r.GetAnswers()
	}
	return ""
}

// remandFloorsOf is the events.id of the latest remand ruling on each gap of open — the place after
// which blue's answer to it is read (blueAnswersOf) — off the motion fold the bench's other readers
// take (motionsAt, read at events.id), so a remand is the ruling the docket's disposition says it
// is. A gap the bench never remanded has none, and every answer counts.
func remandFloorsOf(motions []*Motion, open map[string]bool) map[string]int64 {
	out := map[string]int64{}
	remanded := recordpb.Word(recordpb.Disposition_DISPOSITION_REMANDED)
	for _, m := range motions {
		if m.Subject == "docket" && m.Ruling == remanded && open[m.GapID] {
			out[m.GapID] = max(out[m.GapID], m.ruled)
		}
		// AN UPHELD POSITION NAMING GAPS REMANDS EACH, as the docket's disposition does.
		for _, g := range upheldPositionGaps(m) {
			if open[g] {
				out[g] = max(out[g], m.ruled)
			}
		}
	}
	return out
}
