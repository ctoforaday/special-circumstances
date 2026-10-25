package record

import (
	"fmt"
	"strings"

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
// omission is merely a QUALITY loss — an inquiry never revisited, a repair with no manifest
// row, an archive never sampled, a citation nobody verified, a source blue never cited, a proof
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
// NOT DERIVED, AND SAID RATHER THAN LEFT AS A SILENT HOLE: `motion grade file` and `retire` turn on a judgement
// (whether you disagree, whether the claim should go) that no derivation can make for the seat.
// An affordance line for those would be an expectation that cannot be honestly met, which is the
// defect TestEveryExpectationIsReachableOnItsBoard exists to catch one layer up.
func availableOf(evs []*Event, win WindowIndex, gaps []WorkGapState, role, seatID string) []Item {
	out := []Item{}
	// Blocks is FALSE on every one of these: an affordance is open work, not owed work,
	// and that distinction is the whole content of the rule this file used to split a list
	// in half to honour.
	add := func(what string) { out = append(out, Item{What: what}) }

	switch role {
	case "blue":
		// An inquiry proposed and never revisited records an intention rather than a choice.
		//
		// THE PREDICATE IS StaleInquiries, NOT A SECOND COPY OF IT. A followed inquiry (underway)
		// rests there, and the `How` never tells a seat to abandon or defer it.
		for _, a := range StaleInquiriesOf(evs, win) {
			add(fmt.Sprintf("inquiry %s is at %q and no fate has been recorded on it since epoch %d — an inquiry declared once and never revisited records an intention rather than a choice", a.ID, a.Status, a.Epoch))
		}
		// A repair with no receipt is one nobody audited, including its author. THE PREDICATE IS
		// ManifestOwed, the one the report and the scorecard read: a gap blue was dispatched onto,
		// open when it sat, and answered by its edit that sitting.
		//
		// NO NOT-MEASURED LINE HERE. A later register closes every earlier blue sitting, so the only
		// sitting the record can leave unresolved is the latest — and while this list is read, that
		// is the sitting reading it. Its unreceipted edits are the receipts it can still file.
		for _, id := range ManifestUnreceipted(evs, win, WhileRunning).Gaps {
			add("gap " + id + " was answered by an edit and carries no manifest row — the report names a gap YOU repaired that carries no row as a repair nobody audited, including its author")
		}
		// A LENS THAT KEPT ITS GRADE SAID WHAT ITS CHECK STILL FINDS FAILING, and that is what blue's
		// answer still lacks — the work, not the argument for the gap, which stays on the board with
		// the mint's reason and every regrade that moved a grade (gblock, 2026-10-18). It stays here
		// until blue answers the gap again.
		for _, w := range keptGradeItems(evs, win, gaps) {
			add(w)
		}
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
		// An UNRULED motion suppresses it, because filing a second time while the first is pending
		// asks the bench the same question twice.
		pending := map[string]bool{}
		for _, m := range MotionsOf(evs, win) {
			if m != nil && m.Subject == "docket" && m.GapID != "" && !m.Ruled() {
				pending[m.GapID] = true
			}
		}
		for _, g := range gaps {
			// ONLY A GAP THAT HOLDS PASS IS OFFERED. A docket motion holds PASS until the bench
			// rules it, so docketing a gap that is not material would create the block the gap
			// itself does not — the one the chair's by-class listing exists to avoid. The verb is
			// not refused (a party may still docket such a gap); the list just does not offer it.
			// A stranded gap is held as material, so the offer stays on it.
			if !g.Open || pending[g.ID] || !(g.Material || g.Stranded) {
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
				add(remandedGapForChair(g))
				continue
			}
			add("gap " + g.ID + " is open and you have not closed it — `motion docket file` puts it before the bench, which is the channel for a gap you cannot settle yourself")
		}
	case "lens":
		// Accepting a grade motion does not move the grade. Saying so is not doing it.
		//
		// ON THE ORIGINATOR'S LIST, AND ONLY WHERE THE WRITE PATH ADMITS THE REGRADE: requireOriginator
		// refuses it from any other seat, requireOpenGap refuses it on a closed gap, and
		// writeIn refuses a second regrade of one gap in one sitting. regradesAfforded holds
		// all three, so no line here offers a seat an act its own write path refuses.
		minted := mintedBy(evs)
		//
		// IT NAMES THE AXIS AND THE GRADE EACH RULING ACCEPTED, off the same reader the gap's
		// `regrades_owed` field is filled from (WorkOfSeat). Without them the item said a regrade
		// was owed and not which way: universe-wave-b's architecture lens found the motion's
		// dimension and proposed grade on no surface it holds, and regraded another axis upward.
		for _, o := range regradesAfforded(evs, win, gaps, minted, seatID) {
			add("gap " + o.GapID + " had a grade motion ACCEPTED and no regrade followed it: " + o.accepted() + " — accepting a dispute does not move the grade, and a grade that moved with no regrade event reads as though the dispute was answered by silence")
		}
		// A CITATION OR PROOF NOBODY CHECKED IS ON ITS OWNER'S LIST ALONE (anchorOwnerOf). Every other
		// lens reads the anchor's standing check off `show evidence` and audits its sentence as content.
		if seatID == anchorOwnerOf(false) {
			for _, key := range citedClaimsWithoutVerify(evs) {
				add("citation " + key + " is on the record and nobody has verified it against what the source actually says")
			}
		}
		// THE PROOF ITEM CARRIES THE ID THE VERB TAKES: the proof's anchor, which `reproduce --anchor`
		// takes and the seat reads in the report. When the verb took the script's sha256 and the
		// item named the anchor, three lenses on universe-m11 spent 22 proof-file reads and 14
		// filesystem hunts translating one into the other, for 12 re-runs.
		if seatID == anchorOwnerOf(true) {
			for _, o := range proofChecksOwed(evs, win) {
				//
				// AND IT SAYS WHERE THE SCRIPT IS READ, IN THE ORDER THE ACT NEEDS. It said "audited by
				// RE-RUNNING it, not by reading it" and handed over the re-run's id — so the first thing
				// a lens could do with the item was record a verdict on a script it had not read (wave A,
				// 2026-10-09: three placeholder verdicts, each of which stands).
				add(o.item())
			}
		}
		// WHETHER BLUE ANSWERED, PER GAP, AND WHETHER THE RECORD CAN SAY (#1122).
		//
		// A lens dispatched on its own open gap has one question before it re-reads anything: did
		// blue answer this, or am I being asked to re-argue a gap nothing has touched? Nothing said.
		// The lens's other surface — the report's own movement — answers a document-level question
		// ("the report was at N and is at M"), which is not the same question: the report can move a
		// great deal without a single edit answering this gap.
		//
		// ON THE ORIGINATOR'S LIST, by the same rule as the regrade line above: only the minting seat
		// re-audits and closes its gap, so this is news for that seat and noise for any other.
		for _, w := range blueAnswers(evs, win, gaps, seatID) {
			add(w)
		}
		// ANOTHER LENS'S ARGUMENT ABOUT YOUR GAP REACHES YOU, because nobody else can act on it.
		//
		// A gap is its minter's from mint to close, so a lens that thinks another lens's gap is
		// wrong has nothing it may DO about it — and until this item, nothing it filed reached the
		// seat that could. So it filed in the log. Measured on universe-m12: three of ten `defect`
		// entries were lenses arguing about gaps they did not mint ("G2 is incorrectly labeled as
		// unverified", "G3 remains open and unanswered"), and the one interviewed said who it
		// expected to read it — the operator, because the log was the only channel it believed
		// reached anybody. `finding --about-kind gap` already existed for exactly this and was
		// used zero times, because a finding anchored to a gap went to no one's list.
		//
		// It stays until the minter acts on that gap after the finding — a regrade, one that keeps the
		// grade included, or a close — so it asks for an answer, not for the finding to be
		// acknowledged.
		for _, w := range findingsAboutYourGaps(evs, win, gaps, minted, seatID) {
			add(w)
		}
	}
	return out
}

// findingsAboutYourGaps is each finding another seat anchored to an open gap this seat minted, that
// this seat has not acted on since.
func findingsAboutYourGaps(evs []*Event, win WindowIndex, gaps []WorkGapState, minted map[string]string, seatID string) []string {
	open := map[string]bool{}
	for _, g := range gaps {
		if g.Open {
			open[g.ID] = true
		}
	}
	var out []string
	for _, f := range findingsAboutGapsOf(evs, win, open, minted) {
		if f.minter != seatID {
			continue
		}
		what := f.text
		if r := []rune(what); len(r) > 160 {
			what = string(r[:160]) + "…"
		}
		out = append(out, f.filer+" filed a finding about your gap "+f.gap+" — "+what+
			" — only you can act on "+f.gap+", so answer it: regrade it, close it, or keep its grade with a regrade that says why it stands")
	}
	return out
}

// gapFinding is a finding another seat anchored to an open gap, standing until the gap's minter acts
// on the gap after it.
type gapFinding struct {
	id, gap, minter, filer, text string
	at                           int64 // the finding's events.id
}

// findingsAboutGapsOf is every finding a seat anchored to an open gap of open it did not mint, that
// the minter has not acted on since: the one reading of it, which the minter is handed
// (findingsAboutYourGaps) and owes while it has not sat since the finding (lensOwesOf).
func findingsAboutGapsOf(evs []*Event, win WindowIndex, open map[string]bool, minted map[string]string) []gapFinding {
	var out []gapFinding
	for i, e := range evs {
		f := e.GetFinding()
		if f == nil || f.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_GAP {
			continue
		}
		gap := f.GetAboutRef()
		minter := minted[gap]
		if !open[gap] || minter == "" || e.GetSeatId() == minter || actedOnGapAfter(evs[i+1:], minter, gap) {
			continue
		}
		out = append(out, gapFinding{id: f.GetId(), gap: gap, minter: minter, filer: e.GetSeatId(), text: f.GetText(), at: win.Of(e).ID})
	}
	return out
}

// actedOnGapAfter is whether seatID — the gap's minting lens — recorded a regrade or a close on gap
// among evs. A regrade that keeps the grade is one: it answers the finding with why the gap stands.
func actedOnGapAfter(evs []*Event, seatID, gap string) bool {
	for _, e := range evs {
		if e.GetSeatId() != seatID {
			continue
		}
		if r := e.GetRegrade(); r != nil && r.GetGapId() == gap {
			return true
		}
		if c := e.GetClose(); c != nil && c.GetGapId() == gap {
			return true
		}
	}
	return false
}

// mintedBy maps each gap to the seat that minted it — the same fact requireOriginator reads off
// the gap table, read here off the events this derivation already holds.
func mintedBy(evs []*Event) map[string]string {
	out := map[string]string{}
	for _, e := range evs {
		if m := e.GetMint(); m != nil && m.GetGapId() != "" {
			if _, seen := out[m.GetGapId()]; !seen {
				out[m.GetGapId()] = e.GetSeatId()
			}
		}
	}
	return out
}

// OwedRegradeJSON is one accepted grade motion that still waits for its regrade, as the gap's work
// item carries it (WorkGapJSON.RegradesOwed): the motion, the axis it contested and the grade the
// ruling accepted for that axis — the GradeDimension and Grade words, as the filing holds them.
type OwedRegradeJSON struct {
	MotionID  string `json:"motion_id"`
	Dimension string `json:"dimension"`
	Proposed  string `json:"proposed"`
}

// owedRegrade is one gap's owed regrade: every accepted grade motion on it that no regrade has
// followed, in ruling order. One row per gap, because a sitting records one regrade of a gap and
// that one act carries every axis.
type owedRegrade struct {
	GapID   string
	Motions []OwedRegradeJSON
	ruled   int64 // the events.id of the latest of their rulings
}

// accepted states each waiting motion's axis and grade for the work item's sentence. A filing that
// states no axis or no grade says so in its place: an empty word would read as an axis named "".
func (o owedRegrade) accepted() string {
	word := func(w, missing string) string {
		if w == "" {
			return missing
		}
		return w
	}
	parts := make([]string, 0, len(o.Motions))
	for _, m := range o.Motions {
		parts = append(parts, fmt.Sprintf("%s to %s (motion %s)", word(m.Dimension, "an axis the motion does not state"), word(m.Proposed, "a grade the motion does not state"), m.MotionID))
	}
	return strings.Join(parts, ", ")
}

// regradesAfforded is each gap whose accepted grade motion still waits for its regrade, and whose
// regrade record.Append admits from seatID now, with every such motion's axis and accepted grade.
// It reads "which regrade does a ruling still wait for" off regradesAwaitedOf and adds the write
// path's filters: the work item's sentence (availableOf) and the gap's `regrades_owed` field
// (WorkOfSeat) both come from here.
//
// THE DEBT IS THE REGRADE AFTER THE RULING. An accepted grade motion asks the minting lens to move
// the grade, and a regrade recorded before the ruling answers nothing the ruling said. Any regrade
// of the gap after the ruling discharges it, whichever way it moves the grade: the item does not
// block, and a lens that regrades and keeps the grade has answered the ruling on the record, in
// its basis.
//
// IT IS LISTED ONLY WHERE THE WRITE PATH ADMITS IT. Append is validate, then writeIn, and
// each refusal a regrade can meet on that path has its filter here:
//
//   - requireOriginator (the gap is another seat's)    → minted[gap] == seatID
//   - requireGap (no such gap)                         → the gap is a row of gaps
//   - requireOpenGap (the gap is CLOSED)               → that row's Open
//   - writeIn's duplicate key (the seat's              → the seat's current sitting holds no
//     current sitting already holds a regrade of it)     regrade of the gap
//
// The once-per-sitting duty does not reach a regrade (it is no singleton), and the required-field
// walk refuses the seat's own input, never a state of the gap.
// TestARegradeOnTheListIsOneTheWritePathAdmits holds the list to Append itself, row by row.
//
// THE SAME-SITTING HOLD IS EXACT. A regrade's key carries the count of its seat's sittings
// (deriveKey) and its stored sitting is the seat's newest at the write; both read the `sittings`
// view. So an earlier regrade's key is the one a regrade now would derive exactly when its stored
// sitting is still the seat's latest. Nothing stands in for it this sitting: the first regrade
// stands, and the duplicate-key refusal names the next sitting's regrade as the act that answers
// it (supersede.go). The item appears at the seat's next sitting, which is the first that can
// record the regrade.
//
// evs is one read of the whole record in id order. The reader is the work list's whole-record
// snapshot (WorkOfSeat), so win knows every seat's latest sitting.
//
// THE JOIN IS motionsAt, NOT A SECOND COPY OF IT: the gap is on the FILING and the verdict on the
// RULING, and motion.go's projection is the one that pairs them and says WHICH ruling a motion
// carries and WHERE it sits — the first, in the place it was made. regradesAwaitedOf asks it for the
// verdict and the place together, so the regrade is weighed against the ruling whose verdict is read.
func regradesAfforded(evs []*Event, win WindowIndex, gaps []WorkGapState, minted map[string]string, seatID string) []owedRegrade {
	latest := win.LatestSittingOf(seatID)
	heldThisSitting := map[string]bool{} // gap → the seat's current sitting holds a regrade of it
	for _, e := range evs {
		if r := e.GetRegrade(); r != nil && r.GetGapId() != "" && e.GetSeatId() == seatID && win.Of(e).SittingID == latest {
			heldThisSitting[r.GetGapId()] = true
		}
	}
	open := map[string]bool{}
	for _, g := range gaps {
		if g.Open {
			open[g.ID] = true
		}
	}
	var out []owedRegrade
	for _, o := range regradesAwaitedOf(evs, win, open) {
		if minted[o.GapID] == seatID && !heldThisSitting[o.GapID] {
			out = append(out, o)
		}
	}
	return out
}

// regradesAwaitedOf is each open gap of open whose accepted grade motion no regrade has followed,
// with every such motion's axis and accepted grade and the events.id of the latest ruling among
// them. It is the one reading of "which regrade does a ruling still wait for" before the write
// path's filters: the minting lens is offered it (regradesAfforded), and owes it while it has not
// sat since the ruling (lensOwesOf).
func regradesAwaitedOf(evs []*Event, win WindowIndex, open map[string]bool) []owedRegrade {
	ids := win.IDs(evs)
	lastRegrade := map[string]int64{} // gap → events.id of its last regrade, by any seat
	for i, e := range evs {
		if r := e.GetRegrade(); r != nil && r.GetGapId() != "" {
			lastRegrade[r.GetGapId()] = ids[i]
		}
	}
	var out []owedRegrade
	row := map[string]int{} // gap → its row of out
	for _, m := range motionsAt(evs, ids, win) {
		if m.Subject != "grade" || m.Ruling != "accepted" {
			continue
		}
		id := m.GapID
		if id == "" || !open[id] {
			continue
		}
		if at, regraded := lastRegrade[id]; regraded && at > m.ruled {
			continue
		}
		k, listed := row[id]
		if !listed {
			k, row[id] = len(out), len(out)
			out = append(out, owedRegrade{GapID: id})
		}
		out[k].Motions = append(out[k].Motions, OwedRegradeJSON{MotionID: m.ID, Dimension: m.Fields["dimension"], Proposed: m.Fields["proposed"]})
		out[k].ruled = max(out[k].ruled, m.ruled)
	}
	return out
}

// citedClaimsWithoutVerify finds citations on the record that nobody checked.
//
// THE JOIN IS THE CITATION ANCHOR, and it was neither of the keys this used to read.
//
// It looked for `cite_key`, then `key`, on both sides. A `cite` event carries neither — its keys
// are access_date, claim, label, location, sha256, title, url — so the lookup always missed, the
// slice was always empty, and no lens was ever told about an unverified citation. "Nothing is
// outstanding" and "the join key is not on the event" were the same bytes, and the affordance had
// never fired in the tool's lifetime.
//
// It also hid a second defect: the instruction that dead path handed over named `--key`, a flag
// `lens verify` does not have, so the one time it fired it would have failed. Neither was visible
// because the other kept it off the board.
//
// `blue cite` records the anchor id as `label` (c-<hex>, the token it splices into the report);
// `lens verify` records the citation it checked as `anchor`. An INDEPENDENT verify carries no
// anchor and deliberately discharges nothing: it is a check against a source blue never cited.
//
// A citation whose anchor left the report with a retired claim is owed nothing: `verify` refuses
// it, having no sentence to read the source against, so offering it would offer a refused act.
func citedClaimsWithoutVerify(evs []*Event) []string {
	verified := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		if r, isRetire := body.(*recordpb.Retire); isRetire {
			for _, a := range r.GetAnchors() {
				verified[a] = true
			}
		}
		if v, isVerify := body.(*recordpb.Verify); isVerify {
			// A CORROBORATION carries no anchor and discharges nothing: the empty anchor is what
			// makes a verification one.
			if k := v.GetAnchor(); k != "" {
				verified[k] = true
			}
		}
	}
	var out []string
	seen := map[string]bool{}
	for i := range evs {
		body, ok := recordpb.Body(evs[i])
		if !ok {
			continue
		}
		c, isCite := body.(*recordpb.Cite)
		if !isCite {
			continue
		}
		k := c.GetLabel()
		if k == "" || verified[k] || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out
}

// proofChecksOwed finds the proof anchors whose sentence nobody has judged. `print("7 is prime")`
// reproduces perfectly forever, which is why the audit is a re-run AND a reading, never either alone.
//
// AN ITEM IS PER ANCHOR, AND IT SAYS WHAT THE SCRIPT ALREADY HAS. A re-run belongs to the script
// and answers for every proof on it; a soundness verdict belongs to one sentence (proofverdict.go).
// So a second anchor on a script that was already re-run is still owed — its sentence was never
// judged — and the act it is owed re-runs nothing. Keyed by sha, the second anchor's item vanished
// with the first anchor's re-run, and its sentence shipped judged by a verdict on another.
//
// A reopened anchor — judged, then edited — is not an item here: it is the reopened row, owed by
// the same owner and holding a PASS (reopenedEvidence, anchorOwnerOf).
//
// Every proof carries its script's sha: the write refuses one without (validate), because no act
// could discharge the item it would put here.
func proofChecksOwed(evs []*Event, win WindowIndex) []proofOwed {
	ledger := anchorChecksOf(evs, win)
	reruns := scriptRerunsOf(evs)
	var out []proofOwed
	seen := map[string]bool{}
	for _, e := range evs {
		p, isProof := recordpb.BodyAs[*recordpb.Proof](e)
		if !isProof || p.GetProofId() == "" || seen[p.GetProofId()] {
			continue
		}
		seen[p.GetProofId()] = true
		s := ledger.at(p.GetProofId())
		v := proofVerdictOf(s, reruns)
		if s.retired || v.soundness != SoundnessNotJudged {
			continue
		}
		out = append(out, proofOwed{ProofRef: ProofRef{Anchor: s.anchor, Sha: s.sha}, rerun: v.rerunWord()})
	}
	return out
}

// computationSeat is the computation lens's seat — the owner of every proof's check
// (anchorOwnerOf).
const computationSeat = "red-lens-computation"

// anchorOwnerOf is the lens that owes the check of an anchor of one kind — proof says which: a
// citation's is the evidence lens, a proof's the computation lens. Every run seats both. It is the
// one answer to "who owes this anchor's check": the work list that offers a citation or proof nobody
// checked, the debt that keeps a lens from retiring (lensOwesOf), and the owner of a reopened
// anchor's blocker, whom dispatch readies (passBlockersOf).
//
// ONE LENS OWNS A PROOF (gblock, 2026-10-17): "doesn't it belong to computation/proof to run, and
// for the rest it's just content to judge?" A proof item on every lens's list put lenses dispatched
// together in a race to re-run and judge one script — universe-m26 re-ran one script 16 times across
// five lenses.
//
// ONE LENS OWNS A CITATION (gblock, 2026-10-25), and for the same reason: every lens's list offered
// every citation nobody had verified, and lenses opened together verified the same ones — on the
// run measured, 12 of the 15 first-epoch verifies were by lenses other than the evidence lens, and
// later runs put 42 of 58 and 35 of 45 there.
//
// Every other lens reads the anchor's standing check and judges its sentence as content.
func anchorOwnerOf(proof bool) string {
	if proof {
		return computationSeat
	}
	return evidenceSeat
}

// ProofRef is a recorded proof named BOTH ways: the anchor the report carries and the verb takes,
// and the sha256 of the script every proof on it shares.
type ProofRef struct {
	Anchor string
	Sha    string
}

// proofOwed is one proof anchor whose sentence nobody judged, with its script's standing re-run.
type proofOwed struct {
	ProofRef
	// rerun is the script state, one of reproduced | differs | calls_the_tool | not_rerun.
	rerun string
}

// item is the work-list line, in the order the act needs: read the script, then judge. A script
// with a re-run is never re-run again, whatever that re-run showed, so the act an item names turns
// on one thing: whether the script has a re-run.
func (o proofOwed) item() string {
	read := "`show evidence --anchor " + o.Anchor + "` prints the script and its recorded output"
	if o.rerun == RerunNone {
		return "proof " + o.Anchor + " is recorded and nobody has re-run it — a proof is audited by READING its script and then RE-RUNNING it, never by reading alone: " +
			read + ", and `reproduce --anchor " + o.Anchor + "` then re-runs it and records your verdict on its sentence in one act"
	}
	showed := "did not reproduce (" + o.rerun + ")"
	if o.rerun == recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED) {
		showed = "REPRODUCED"
	}
	return "proof " + o.Anchor + "'s script was re-run and " + showed + ", and nobody has judged whether it establishes THIS sentence — one script can back several, and a verdict on another is not a verdict on this one: " +
		read + ", and `reproduce --anchor " + o.Anchor + "` records your verdict on its sentence without re-running it"
}

// remandItemsOf is the item a remand's exchange puts on the list of each seat the dispatch readies
// for it — blue-respond, and the gap's minting lens — carrying the research direction the ruling
// states, so the bench's words reach the two seats that owe them. blue says which of the two seatID
// is.
//
// IT BLOCKS UNTIL THE SEAT HAS ACTED ON THE GAP (gblock, 2026-10-25), by the per-seat envelope gate
// (EnvelopeRefusal): an act on the gap after the ruling (gapActsOf) clears it, and until then the
// seat's sitting cannot end. On universe-m34 the item was offered and not owed, blue worked other
// gaps, and a sitting that never touched the remanded gap was the exchange the remand granted.
//
// A SITTING OPENED BEFORE THE RULING IS NOT HELD BY IT: the remand landed while it sat, the next plan
// readies the seat, and that sitting owes it. Once the seat has acted the item stays on the list,
// not owed, while the remand's exchange is: the direction is still what the exchange is about.
//
// It stands while the plan's own route (routeOf) owes the remand's exchange, so it goes when the gap
// closes, when a docket motion on it puts it back before the bench, and when the exchange has been
// had.
func remandItemsOf(evs []*Event, win WindowIndex, gaps []WorkGapState, seatID string, blue bool) []Item {
	minted := mintedBy(evs)
	var acts map[string][]gapAct
	var out []Item
	for _, g := range gaps {
		if !g.Open || g.route != routeRemandOwed || (!blue && minted[g.ID] != seatID) {
			continue
		}
		if acts == nil {
			acts = gapActsOf(evs, win.IDs(evs))
		}
		what := remandItem(g, blue)
		switch {
		case seatActedSince(acts[g.ID], seatID, g.remandRuled):
			out = append(out, Item{What: what + " — you have recorded an act on it since the ruling, so it does not hold your sitting"})
		case win.LatestSittingOf(seatID) < g.remandRuled:
			out = append(out, Item{What: what + " — the ruling landed after this sitting opened, so your next sitting owes it"})
		default:
			out = append(out, Item{What: what, Blocks: true})
		}
	}
	return out
}

// remandItem is the remand's sentence for blue or for the gap's minting lens: the direction, the act
// that discharges it, and where the gap goes next.
func remandItem(g WorkGapState, blue bool) string {
	const next = ". If this exchange leaves the gap at impasse, it goes back to the bench, to be closed"
	if blue {
		return "gap " + g.ID + " was REMANDED by the bench and is back in the debate for ONE more exchange — you and its minting lens are dispatched on it, and what you owe is the research direction the ruling states: " + g.DocketDirection +
			". Your sitting cannot end until you record an act on " + g.ID + ": an edit or a proof that answers it, a fate on the inquiry it is anchored on, or a motion on it" + next
	}
	return "gap " + g.ID + ", which you minted, was REMANDED by the bench and is back in the debate for ONE more exchange — audit blue's answer against the research direction the ruling states: " + g.DocketDirection +
		". Your sitting cannot end until you record an act on " + g.ID + ": a regrade (one that keeps the grade says what your check still finds failing), a close, or a finding about it" + next
}

// remandedGapForChair is the chair's row for a remanded gap: what the dispatch does with it next, by
// the plan's own predicate (routeOf).
func remandedGapForChair(g WorkGapState) string {
	const head = "gap "
	switch g.route {
	case routeRemandOwed:
		return head + g.ID + " is open because the BENCH REMANDED it — the dispatch readies its minting lens and blue for one more exchange on the research direction the ruling states: " + g.DocketDirection
	case routeRedocket:
		return head + g.ID + " is open because the BENCH REMANDED it, and the exchange the remand granted left it at impasse — the dispatch dockets it for the bench again, to be closed"
	case routeDocket:
		return head + g.ID + " is open, the BENCH REMANDED it before it reached impasse, and it is at impasse now — the dispatch dockets it for the bench"
	case routeNobody:
		return head + g.ID + " is open, the BENCH REMANDED it, and it is not material: the exchange its remand granted has been had, and the dispatch readies nobody for it"
	case routeGradeMovement:
		return head + g.ID + " is open because the BENCH REMANDED it, and its grades have moved since, with the open gaps' grade movement the bench has not ruled on past its threshold — the dispatch dockets it for the bench"
	}
	return head + g.ID + " is open because the BENCH REMANDED it, and it is below its limits — the dispatch readies its minting lens and blue on it"
}
