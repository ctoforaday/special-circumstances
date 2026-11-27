package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT THIS BOARD MAKES AVAILABLE, WHICH IS NOT WHAT THIS SEAT OWES.
//
// This file used to open by arguing that the affordances had to be a SECOND list, kept off
// `complete` so the view could not disagree with the gates. The argument was sound and the
// conclusion was wrong, and the difference is worth keeping because it is a shape that recurs.
//
// The rule it was protecting — sitting.go's "nothing here invents an obligation", because a seat
// told it is finished by one surface and refused by another learns to trust neither — is real and
// still holds. But a duty is derived only where omission already carries a mechanical consequence,
// so it can name 4 of blue's verbs, 5 of merge's, 3 of bench's and 2 of lens's. Every act whose
// omission is merely a QUALITY loss — an inquiry never revisited, an archive never sampled, a citation nobody verified, a source blue never cited, a computation
// nobody re-ran, a grade accepted and never moved — got no line. Those are exactly the acts the
// probe boards bait and the measured runs skip.
//
// Protecting the rule by making a SEPARATE list made those acts absent from the surface a seat
// consults to ask whether anything is left. Three seats interviewed about verbs they never touched
// all stopped the same way and said so in the same words — the outstanding array emptied — and one
// put the mechanism exactly: the work list "supplied the goal, and a goal crowds out survey.
// Things off the list weren't declined. They were invisible."
//
// The rule never required a second LIST. It required that `complete` be computed from the
// enforced half only. That is a property of one field, not of two surfaces, so it is now a flag
// on each item (record.Item.Blocks) and there is one list. See sitting.go.
//
// The carriage argument that used to close this comment is also settled and is left here because
// it is the measurement, not the opinion: across 24 probe dispatches the duty-carrying projection
// was read 0.33-2.00 times a sitting while `board` was read 2.7-4.3. The answer to a list that
// does not arrive is not a second copy of it somewhere better-trafficked. It is one list, on the
// one command a seat is told to run.

// availableOf derives the acts this board affords this seat right now.
//
// EVERY ENTRY NAMES A CONDITION READ OFF THE RECORD, never a general encouragement. "You could
// cite something" is not an affordance, it is advice; "claim c-3 is cited and nobody has verified
// it" is a fact about this board with one act attached to it.
//
// NOT DERIVED, AND SAID RATHER THAN LEFT AS A SILENT HOLE: `motion docket file` and `retire` turn on a judgement
// (whether you disagree, whether the claim should go) that no derivation can make for the seat.
// An affordance line for those would be an expectation that cannot be honestly met, which is the
// defect TestEveryExpectationIsReachableOnItsBoard exists to catch one layer up.
//
// blueOut is the run in the epoch that holds blue out (blueHeldOut, under the run's terms): the
// chair's row for a remanded gap says what the dispatch can still do there.
func availableOf(evs []*Event, win WindowIndex, gaps []WorkGapState, role, seatID string, blueOut bool) []Item {
	out := []Item{}
	// Blocks is FALSE on every one of these: an affordance is open work, not owed work,
	// and that distinction is the whole content of the rule this file used to split a list
	// in half to honour.
	add := func(it Item) { out = append(out, it) }
	// A DOCKETED GAP IS OFF A PARTY'S LIST until it is ruled, apart from its closing argument (ruled by
	// gblock 2026-10-31); the chair's arm reads every gap.
	if role == "blue" || role == "lens" {
		gaps = undocketed(gaps)
	}

	switch role {
	case "blue":
		// An inquiry proposed and never revisited records an intention rather than a choice.
		//
		// THE PREDICATE IS StaleInquiries, NOT A SECOND COPY OF IT. A followed inquiry (underway)
		// rests there, and the `How` never tells a seat to abandon or defer it. Blue-respond's list
		// carries each open inquiry as owed instead (blueInquiryItemsOf), so it is not offered here too.
		if seatID != blueRespondSeat {
			for _, a := range StaleInquiriesOf(evs, win) {
				add(Item{Kind: ItemStaleInquiry, Subject: about(SubjectInquiry, a.ID), What: fmt.Sprintf("inquiry %s is at %q and no fate has been recorded on it since epoch %d — an inquiry declared once and never revisited records an intention rather than a choice", a.ID, a.Status, a.Epoch)})
			}
		}
		// A LENS THAT KEPT ITS GRADE SAID WHAT ITS CHECK STILL FINDS FAILING, and that is what blue's
		// answer still lacks — the work, not the argument for the gap, which stays on the board with
		// the mint's reason and every regrade that moved a grade (gblock, 2026-10-18). It stays here
		// until blue answers the gap again.
		out = append(out, keptGradeItems(evs, win, gaps)...)
	case "chair":
		// AN OPEN GAP IS EITHER CLOSED OR DOCKETED, and until the docket was a motion this could
		// not be said here at all — the note above used to record that "`closing` wants the
		// docket, which is not recoverable from board state alone". It is now: a docket motion is
		// an event, so "which gaps are before the bench" is a question the board answers.
		//
		// AN AFFORDANCE, NOT A DUTY. The open-gap row in sitting.go already refuses PASS over a
		// material gap, so a second blocking item would be a duplicate — and worse, a seat that OBEYED it
		// would be further from complete than before, because filing adds an unruled motion. What
		// filing buys is not completeness but the gap moving from this seat's undecided pile to a
		// question the bench owes an answer to.
		//
		// NO FLAGS IN THE TEXT. This file's own contract, and the reason is recorded above: an
		// affordance once handed a seat a flag the verb does not have, so the only instruction the
		// duty ever gave could not run.
		// PENDING, NOT EVER-FILED.
		//
		// A docket motion that has been RULED is answered. If the gap is still open after that
		// answer, the answer was `remanded`: every other disposition closes the gap, so an open gap
		// with a ruled docket motion is EXACTLY a remand, and the ruling states its direction. Keyed
		// on ever-filed, this row went silent permanently at the first filing, and a remanded gap
		// was listed as blocking PASS with nothing said about what happens to it next. Measured:
		// after M2 was ruled `remanded` on R1-2, `show work` offered nothing on the gap, while a
		// never-docketed R1-3 got the affordance.
		//
		// AN UNRULED PROPOSAL OR A STANDING DOCKETING SUPPRESSES IT: a gap holds one proposal at a time,
		// and no proposal is filed on a docketed gap (L38) — the parties' closings are its statements.
		pending := map[string]bool{}
		for _, m := range MotionsOf(evs, win) {
			if m != nil && m.Subject == "docket" && m.GapID != "" && !m.Ruled() {
				pending[m.GapID] = true
			}
		}
		for _, k := range docketingsAt(evs, win.IDs(evs)) {
			if k.Standing() {
				pending[k.GapID] = true
			}
		}
		for _, g := range gaps {
			// ONLY A GAP THAT HOLDS PASS IS OFFERED. A docket motion holds PASS until the bench
			// rules it, so docketing a gap that is not material would create the block the gap
			// itself does not — the one the chair's by-class listing exists to avoid. The verb is
			// not refused (a party may still docket such a gap); the list just does not offer it.
			if !g.Open || pending[g.ID] || !g.Material {
				continue
			}
			// ONE ROW PER GAP, BETTER WORDS WHEN THE BOARD KNOWS MORE (#759).
			//
			// A remanded gap and a gap nobody has ever docketed both reach this line, and until
			// the view could tell them apart they got the same sentence — true of both and
			// actionable on neither. A SECOND row for the remanded case would be the wrong fix:
			// two items naming one gap is a duplicate, and the open-gap row in sitting.go
			// already blocks PASS over a material gap. So the row is REPLACED, not added.
			//
			// The direction is the substance. "The bench remanded this" tells a seat the history;
			// "the bench sent it back for blue to report what the stated direction found" tells it
			// what has to happen, and that sentence is the bench's own words off the record rather
			// than this file's paraphrase of them. What happens next is the dispatch's, and the row
			// says it by the plan's own predicate (routeOf).
			if g.Remanded {
				add(gapItem(ItemKind(BlockerRemandOwed), g.ID, remandedGapForChair(g, blueOut)))
				continue
			}
		}
	case "lens":
		// WHETHER BLUE ANSWERED, PER GAP, AND WHETHER THE RECORD CAN SAY (#1122).
		//
		// A lens dispatched on its own open gap has one question before it re-reads anything: did
		// blue answer this, or am I being asked to re-argue a gap nothing has touched? Nothing said.
		// The lens's other surface — the report's own movement — answers a document-level question
		// ("the report was at N and is at M"), which is not the same question: the report can move a
		// great deal without a single edit answering this gap.
		//
		// ON THE ORIGINATOR'S LIST: only the minting seat re-audits and closes its gap, so this is news for that seat and noise for any other.
		out = append(out, blueAnswers(evs, win, gaps, seatID)...)
	}
	return out
}

// assigneeOf maps each gap to its assignee, the seat that minted it — the fact requireOriginator
// reads off the gap table — read here off the events this derivation already holds (mintersOf).
func assigneeOf(evs []*Event) map[string]string { return mintersOf(evs) }

// computationSeat is the computation lens's seat — the owner of every computation's check
// (anchorOwnerOf).
const computationSeat = "red-lens-computation"

// anchorOwnerOf is the lens that owes the check of an anchor of one kind — computation says which: a
// citation's is the evidence lens, a computation's the computation lens. Every run seats both. It is the
// one answer to "who owes this anchor's check": the owner of an unchecked anchor's blocker, whose work
// list it blocks, whose debt it is (lensOwesOf), and whom dispatch readies (passBlockersOf).
//
// ONE LENS OWNS A COMPUTATION (gblock, 2026-10-17): "doesn't it belong to computation/proof to run, and
// for the rest it's just content to judge?" A computation item on every lens's list put lenses dispatched
// together in a race to re-run and judge one script — universe-m26 re-ran one script 16 times across
// five lenses.
//
// ONE LENS OWNS A CITATION (gblock, 2026-10-25), and for the same reason: every lens's list offered
// every citation nobody had verified, and lenses opened together verified the same ones — on the
// run measured, 12 of the 15 first-epoch verifies were by lenses other than the evidence lens, and
// later runs put 42 of 58 and 35 of 45 there.
//
// Every other lens reads the anchor's standing check and judges its sentence as content.
func anchorOwnerOf(computation bool) string {
	if computation {
		return computationSeat
	}
	return evidenceSeat
}

// remandItemsOf is the item a remand's exchange puts on the list of blue-respond, which the dispatch
// readies for it, and of the gap's assignee, whose turn comes after blue's act — carrying the
// research direction the ruling states, so the bench's words reach both parties to the exchange. blue
// says which of the two seatID is.
//
// BLUE'S ITEM BLOCKS UNTIL BLUE HAS ACTED ON THE GAP (gblock, 2026-10-25), by the per-seat envelope
// gate (EnvelopeRefusal): an act on the gap after the ruling (gapActsOf) clears it, and until then
// blue's sitting cannot end. On universe-m34 the item was offered and not owed, blue worked other
// gaps, and a sitting that never touched the remanded gap was the exchange the remand granted.
//
// THE LENS'S ITEM NEVER BLOCKS (ruled by gblock 2026-10-30): the exchange begins with blue's act, and
// the lens's turn on the gap comes with it (lensTurnOn), so while the exchange is owed blue has not
// acted since the ruling and there is nothing for the lens to audit. A lens that sits for another
// reason reads the direction its next audit of the gap is held to.
//
// A BLUE SITTING FOR A DISPATCH MADE BEFORE THE RULING IS NOT HELD BY IT: the plan that dispatched
// it carried no remand, the next plan readies blue, and that sitting owes it. The dispatch row is the
// reading, not blue's latest sitting, because a sitting is not the seat's until it registers: blue's
// harness bracket names no seat, so before its register the latest sitting of blue-respond is the
// one before. On universe-m35 the item read "the ruling landed after this sitting opened"
// (non-blocking, and false) before register, and blocked after it. The dispatch row is on the record
// before the seat sits, so the item says the same thing before and after register. Once blue has
// acted the item stays on the list, not owed, while the remand's exchange is: the direction is still
// what the exchange is about.
//
// It stands while the plan's own route (routeOf) owes the remand's exchange, so it goes when the gap
// closes, when a docket motion on it puts it back before the bench, and when an exchange in which
// blue acted on the gap has been had.
func remandItemsOf(evs []*Event, win WindowIndex, gaps []WorkGapState, seatID string, blue bool) []Item {
	minted := assigneeOf(evs)
	var acts map[string][]gapAct
	var dispatched int64 // blue's latest dispatch row, read with acts
	var out []Item
	for _, g := range gaps {
		if !g.Open || g.route != routeRemandOwed || (!blue && minted[g.ID] != seatID) {
			continue
		}
		what := remandItem(g, blue)
		if !blue {
			out = append(out, gapItem(ItemKind(BlockerRemandOwed), g.ID, what))
			continue
		}
		if acts == nil {
			acts = gapActsOf(evs, win.IDs(evs), win)
			dispatched = latestDispatchOf(evs, win, seatID)
		}
		switch {
		case seatActedSince(acts[g.ID], seatID, g.remandRuled):
			out = append(out, gapItem(ItemKind(BlockerRemandOwed), g.ID, what+" — you have recorded an act on it since the ruling, so it does not hold your sitting"))
		case dispatched < g.remandRuled:
			out = append(out, gapItem(ItemKind(BlockerRemandOwed), g.ID, what+" — the ruling landed after the dispatch you are sitting for, so your next dispatch owes it"))
		default:
			it := gapItem(ItemKind(BlockerRemandOwed), g.ID, what)
			it.Blocks = true
			out = append(out, it)
		}
	}
	return out
}

// latestDispatchOf is the events.id of the latest dispatch row naming seatID — the chair's, or an
// engage blue wrote for its own next sitting — or 0 when none does.
func latestDispatchOf(evs []*Event, win WindowIndex, seatID string) int64 {
	var at int64
	for _, e := range evs {
		if d, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok && d.GetSeatId() == seatID {
			at = max(at, win.Of(e).ID)
		}
	}
	return at
}

// remandItem is the remand's sentence for blue or for the gap's assignee: the direction, the act
// that discharges it, and where the gap goes next.
func remandItem(g WorkGapState, blue bool) string {
	const next = ". If this exchange leaves the gap at impasse, it goes back to the bench, to be closed"
	if blue {
		return "gap " + g.ID + " was REMANDED by the bench and is back in the debate for ONE more exchange, which begins with your act — what you owe is the research direction the ruling states: " + g.DocketDirection +
			". Your sitting cannot end until you record an act on " + g.ID + ": an edit answering it, a citation or computation placed as its answer, a changed fate on the inquiry it is about or one the section it is about prints, or a motion on it. Its assignee audits your answer after you have acted" + next
	}
	return "gap " + g.ID + ", which you minted, was REMANDED by the bench and is back in the debate for ONE more exchange — it begins when blue records an act on " + g.ID + " against the research direction the ruling states: " + g.DocketDirection +
		". The dispatch readies you on " + g.ID + " once blue has acted or an edit has changed its sentence, and you audit it against that direction then: your audit is what ends this exchange" + next
}

// remandedGapForChair is the chair's row for a remanded gap: what the dispatch does with it next, by
// the plan's own predicate (routeOf). In the epoch that holds blue out (blueOut), an exchange still
// waiting on blue's act cannot be had before the run ends, and the plan lists it as owed work nowhere.
func remandedGapForChair(g WorkGapState, blueOut bool) string {
	const head = "gap "
	switch {
	case g.route == routeRemandOwed && blueOut && g.lensDue:
		return head + g.ID + " is open because the BENCH REMANDED it — blue has acted on it, and the dispatch readies its assignee to audit that answer, which ends the one more exchange; " + blueHeldOutClause + ": " + g.DocketDirection
	case g.route == routeRemandOwed && blueOut:
		return head + g.ID + " is open because the BENCH REMANDED it — its one more exchange begins with blue's act, and " + blueHeldOutClause + ", so that exchange cannot be had before the run ends; the gap stays open and holds the PASS: " + g.DocketDirection
	}
	switch g.route {
	case routeBench:
		return head + g.ID + " is open because the BENCH REMANDED it, and it is docketed again — the dispatch readies its parties for their closings, then the bench"
	case routeProposal:
		return head + g.ID + " is open because the BENCH REMANDED it, and a proposal stands on it — the dispatch readies its assignee for its window, and dockets the proposal for the bench if the assignee does not accept it there"
	case routeRemandOwed:
		return head + g.ID + " is open because the BENCH REMANDED it — the dispatch readies blue for one more exchange on the research direction the ruling states, and its assignee once blue has acted: " + g.DocketDirection
	case routeRedocket:
		return head + g.ID + " is open because the BENCH REMANDED it, and the exchange the remand granted left it at impasse — the dispatch dockets it for the bench again, to be closed"
	case routeDocket:
		return head + g.ID + " is open, the BENCH REMANDED it before it reached impasse, and it is at impasse now — the dispatch dockets it for the bench"
	case routeNobody:
		return head + g.ID + " is open, the BENCH REMANDED it, and it is not material: the exchange its remand granted has been had, and the dispatch readies nobody for it"
	case routeGradeMovement:
		return head + g.ID + " is open because the BENCH REMANDED it, and its grades have moved since, with the open gaps' grade movement the bench has not ruled on past its threshold — the dispatch dockets it for the bench"
	case routeLensFirst:
		return head + g.ID + " is open because the BENCH REMANDED it, and it is at its limits — the dispatch readies its assignee alone on it, and dockets it for the bench if it stands after that audit"
	}
	return head + g.ID + " is open because the BENCH REMANDED it, and it is below its limits — the dispatch readies blue on it, and its assignee once blue has acted on it since the lens last did"
}
