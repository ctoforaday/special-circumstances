package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// lastBlueEdit is the blue edit b.Events[i] is, where it is one.
func lastBlueEdit(b record.Family, i int) (*recordpb.BlueEdit, bool) {
	return recordpb.BodyAs[*recordpb.BlueEdit](b.Events[i])
}

// gapIDOf is the gap id a mint's --json result names, or "".
func gapIDOf(out string) string {
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil {
		return ""
	}
	return env.Result.GapID
}

// A REWORD ENDS ESTOPPEL (plans/gap-is-the-one-holder.md F12, ruled 2026-11-10): an applied-verbatim
// span whose sentence a later edit lists in `reopened` is blue's, and a red mint quoting it is not
// estopped; a later edit re-applying red's text estops again. estoppelSeed's run drives the pair
// after its verbatim application (rewordAppliedText, blue; mintEstopped, the lenses), through the
// stages below. (A found-source reading debt discharged by a verbatim application — red wrote the
// new sentence — is readingRemovalSeed's arm: the drawn prescription here swaps a phrase shorter than
// the lost-quote lookup's bound, F10, so it names no edit.)

// Stages of the reword pair (estoppelRewordStage).
const (
	rewordIdle     = iota
	rewordReworded // blue reworded the applied sentence; red's mint of it lands
	rewordLanded   // red's mint landed; blue re-applies its text
	rewordRestored // blue re-applied red's text; red's mint of it is refused again
	rewordDone
)

// estoppelReword is the pair's state: the gap whose prescription was applied, reworded and re-applied,
// and its prescription's pair as the reword found it — read for its visible phrases alone, since the
// anchor run it carries is stale once another mint lands on the sentence (rewordAppliedText).
type estoppelReword struct {
	stage          int
	gap, old, new_ string
}

// estoppelRewordStage is the pair's stage.
func (r *runner) estoppelRewordStage() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reword.stage
}

// rewordTail is the phrase blue's reword replaces, in lane 2's sentence where every prescription
// lands (mintWith's proposal swaps its phrasing), and rewordedTail what it writes: the sentence keeps
// red's applied text and loses its exact words, so the old substring test would still estop.
const (
	rewordTail   = "across the whole sampled corpus"
	rewordedTail = "across the whole corpus as sampled"
)

// rewordAppliedText is blue's half of the pair, at its sittings in estoppelSeed's run: once a
// prescription has been applied verbatim, the reword of its sentence (rewordReworded); once red's
// mint of the reworded text has landed, the restore of red's words and the re-application of its
// text by --accept, or typed where the accept no longer applies (rewordRestored). It draws nothing
// from the run's RNG, and every sitting it declines says why (rewordPairDeclined): the pair's two
// counts are gated at zero across the sweep, and a zero with no cause is the shape this drive exists
// to refuse.
//
// THE PAIR IS READ OFF THE RECORD AT EACH SITTING, never off the pair's own memory of it. The
// proposal's pair carries the anchor run after the phrase, and every mint on the sentence since
// prepends a token to that run, so the pair the reword recorded is stale by blue's next sitting;
// proposalFor is the record's live pair, located by the gap's token. The restore of blue's words
// swaps the bare visible phrases and leaves the run standing, as edgeRun's other arm does.
func (r *runner) rewordAppliedText(seatID string) {
	if !r.forceEstoppel || seatID != "blue-respond" {
		return
	}
	declined := func(why string) { r.countReading(func(c *readingCounts) { c.rewordPairDeclined[why]++ }) }
	switch r.estoppelRewordStage() {
	case rewordIdle:
		gap, _, miss := r.aVerbatimGap()
		if gap == "" {
			declined(miss)
			return
		}
		cur, err := record.RenderRun(r.run())
		if err != nil {
			declined("could not render the report")
			return
		}
		if !strings.Contains(cur, anchor.Token(gap)) {
			declined("the applied gap's token has left the page, so no edit of its sentence names it reopened")
			return
		}
		gid, old, new_ := r.proposalFor(gap, cur)
		switch {
		case gid == "":
			declined("the applied gap carries no proposal")
			return
		case strings.Count(cur, rewordTail) != 1:
			declined("the report does not hold the reword's phrase exactly once")
			return
		case !strings.Contains(anchortext.Visible(cur), anchortext.Visible(new_)):
			declined("the applied text is not on the page")
			return
		}
		if _, ok := r.do("edit", seatID).set("--quote", rewordTail).set("--new", rewordedTail).
			set("--reason", "fuzz: the sentence red's text was applied to, reworded in blue's own words").admit(); !ok {
			declined("the reword edit was refused")
			return
		}
		if !r.lastEditReopened(seatID, gap) {
			r.noteActErr("blue's reword of the sentence red's fix on " + gap + " was applied to does not list the gap reopened")
			return
		}
		r.mu.Lock()
		r.reword = estoppelReword{stage: rewordReworded, gap: gap, old: old, new_: new_}
		r.mu.Unlock()
	case rewordLanded, rewordRestored:
		r.mu.Lock()
		gap, old, new_, stage := r.reword.gap, r.reword.old, r.reword.new_, r.reword.stage
		r.mu.Unlock()
		// RESTORED, AND STILL ESTOPPING: nothing to do. Restored and ended — this sitting's response,
		// or another blue sitting's, edited the sentence since — blue applies red's text again, so
		// every sitting of blue's ends with the pair's estoppel standing for the lenses' mint.
		if stage == rewordRestored && r.stillEstops(gap) {
			return
		}
		cur, err := record.RenderRun(r.run())
		if err != nil {
			declined("could not render the report")
			return
		}
		// RED'S WORDS ARE PUT BACK where the applied text still stands, so the accept applies again:
		// the bare phrases swap, and the anchor run after them stays where it is.
		oldVis, newVis := anchortext.Visible(old), anchortext.Visible(new_)
		if vis := anchortext.Visible(cur); !strings.Contains(vis, oldVis) && strings.Contains(vis, newVis) {
			if _, ok := r.do("edit", seatID).set("--quote", newVis).set("--new", oldVis).
				set("--reason", "fuzz: blue's words again, before red's are re-applied").admit(); !ok {
				declined("the restoring edit was refused")
				return
			}
		}
		if _, err := r.do("edit", seatID).set("--answers", gap).bare("--accept").
			set("--reason", "fuzz: accepting red's prescribed fix on "+gap+" again").result(); err != nil {
			cur, rerr := record.RenderRun(r.run())
			if rerr != nil {
				declined("could not render the report")
				return
			}
			gid, liveOld, liveNew := r.proposalFor(gap, cur)
			if gid == "" {
				declined("the pair's gap carries no proposal to type")
				return
			}
			if _, ok := r.do("edit", seatID).set("--answers", gap).set("--quote", liveOld).set("--new", liveNew).
				set("--reason", "fuzz: applying red's proposed text verbatim again").admit(); !ok {
				declined("the re-application was refused by both arms")
				return
			}
		}
		if !r.lastEditVerbatim(seatID) {
			r.noteActErr("blue's re-application of red's text on " + gap + " is not recorded applied_verbatim")
			return
		}
		r.mu.Lock()
		r.reword.stage = rewordRestored
		r.mu.Unlock()
	}
}

// lastEditReopened reports whether seatID's latest edit on the record lists anchorID reopened.
func (r *runner) lastEditReopened(seatID, anchorID string) bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		if e, isE := lastBlueEdit(b, i); isE && b.Events[i].GetSeatId() == seatID {
			for _, id := range e.GetReopened() {
				if id == anchorID {
					return true
				}
			}
			return false
		}
	}
	return false
}

// lastEditVerbatim reports whether seatID's latest edit on the record applied red's text verbatim.
func (r *runner) lastEditVerbatim(seatID string) bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		if e, isE := lastBlueEdit(b, i); isE && b.Events[i].GetSeatId() == seatID {
			return e.GetAppliedVerbatim()
		}
	}
	return false
}

// mintRewordedText is the lenses' half of the pair, inside mintEstopped, for a mint of sentence —
// the applied text as it now stands: while the reword stands, the mint LANDS (estoppel_ended_by_reword)
// or is refused by the one-gap rule over a sibling's landed one, never by estoppel; once blue
// re-applied red's text, it is refused by estoppel again (estoppel_restored). It reports whether the
// pair took the mint; false leaves it to mintEstopped's ordinary refusal.
//
// ONE MINT OF THE REWORDED TEXT. Once one has landed, the lenses wait for blue's re-application and
// mint nothing more on the sentence: each such mint is a gap blue counter-edits, and a counter-edit of
// the sentence lists the pair's gap reopened, ending the re-application's estoppel (F12) before a lens
// mints over it.
//
// EVERY ESTOPPEL EXPECTATION IS ASKED OF THE RECORD AT THE MINT (stillEstops), never of the pair's
// stage: blue sits while the lenses do, an edit of the sentence that is not a verbatim one ends the
// pair's estoppel, and blue's ordinary response can apply the pair's prescription, or a sibling's
// prescribing the same words, verbatim again. The estoppel guard is asked in the write before the act
// judgements (record.appendAct), so while the pair's gap estops, no refusal but estoppel's can answer
// the restored mint; while it does not, the mint is not the pair's to count. An estoppel refusal counts
// for the pair only when it names the pair's gap: the guard answers with whichever applied prescription
// the quote holds. CI run 42943412612 (seed 2) minted over a re-application a later edit had already
// ended, expected estoppel, and met the one-gap rule.
func (r *runner) mintRewordedText(seatID, gapID, sentence string) bool {
	switch r.estoppelRewordStage() {
	case rewordLanded:
		r.noteEstoppelMiss("the reword pair's mint of the reworded text has landed; the lenses wait for blue's re-application")
		return true
	case rewordReworded:
		if r.stillEstops(gapID) {
			// BLUE'S ORDINARY RESPONSE APPLIED THE PAIR'S PRESCRIPTION AGAIN: the text is red's once
			// more, and this sitting's mint is mintEstopped's.
			r.noteEstoppelMiss("blue applied the reword pair's prescription verbatim again before a lens minted over the reworded text")
			return false
		}
		args := []string{"--json", "mint", "--seat-id", seatID, "--class", "spec_underspecification",
			"--problem", "fuzz: raising a defect against text blue reworded", "--check-kind", "document",
			"--check", "acc", "--fix", dirCounter, "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
			"--quote", sentence, "--reason", "fuzz: the reworded text is blue's, not red's prescription on " + gapID}
		// A POOL CLASS NO OPEN GAP HOLDS ON THE SENTENCE, so the one-gap rule is not what answers
		// the mint: the estoppel seed's lenses mint their own gaps on this sentence under free classes.
		r.freeClass(args)
		out, err := r.exec(args...)
		switch {
		case err == nil:
			r.mu.Lock()
			first := r.reword.stage == rewordReworded
			r.reword.stage = rewordLanded
			r.mu.Unlock()
			if first {
				r.countReading(func(c *readingCounts) { c.estoppelEndedByReword++ })
			}
			if g := r.noteReadingMintLandedPlain(seatID, out); g != "" {
				r.checkOneGap(g, args)
			}
		case strings.Contains(out, oneGapWords):
			noteEither(args[:2], "a mint of reworded text over a sibling's landed one, refused by the one-gap rule", false)
		case strings.Contains(out, estoppelWords) && strings.Contains(out, gapID) && !r.stillEstops(gapID):
			r.noteActErr("a mint quoting applied text blue has since reworded was refused by estoppel; a reword ends it (F12): " + firstLine(out))
		case strings.Contains(out, estoppelWords):
			// ANOTHER APPLIED PRESCRIPTION THE SENTENCE HOLDS, or the pair's own applied again between
			// the check and the write: the guard's answer on the record the mint landed on.
			noteEither(args[:2], "a mint of reworded text, refused by estoppel over a prescription applied since", false)
		default:
			noteEither(args[:2], "a mint of reworded text, refused by another rule", false)
		}
		return true
	case rewordRestored:
		if !r.stillEstops(gapID) {
			r.rewordBackToReapply("blue edited the re-applied sentence again before a lens minted over it, which ends the estoppel (F12)")
			return true
		}
		args := []string{"--json", "mint", "--seat-id", seatID, "--class", "spec_underspecification",
			"--problem", "fuzz: raising a defect against text I prescribed", "--check-kind", "document",
			"--check", "acc", "--fix", "reword it", "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
			"--quote", sentence, "--reason", "fuzz: estoppel probe — this quotes red's re-applied prescription on " + gapID}
		// A CLASS NO OPEN GAP HOLDS ON THE SENTENCE, so a mint that blue's edit between the check
		// above and the write leaves unestopped lands rather than meeting the one-gap rule.
		r.freeClass(args)
		out, err := r.exec(args...)
		switch {
		case err != nil && strings.Contains(out, estoppelWords) && strings.Contains(out, gapID):
			r.mu.Lock()
			r.reword.stage = rewordDone
			r.mu.Unlock()
			r.countReading(func(c *readingCounts) { c.estoppelRestored++ })
		case err != nil && strings.Contains(out, estoppelWords):
			// THE GUARD NAMED ANOTHER APPLIED PRESCRIPTION THE SENTENCE HOLDS: a refusal, and not
			// proof of the pair's. The pair stays restored for the next lens's mint.
			noteEither(args[:2], "the reword pair's restored mint, refused by estoppel over another applied prescription", false)
		case !r.stillEstops(gapID):
			// BLUE EDITED THE SENTENCE BETWEEN THE CHECK AND THE WRITE: the record's call on the
			// report the mint landed on, and the pair re-applies.
			noteEither(args[:2], "the reword pair's restored mint, blue having edited the sentence under it (F12)", err == nil)
			if err == nil {
				if g := r.noteReadingMintLandedPlain(seatID, out); g != "" {
					r.checkOneGap(g, args)
				}
			}
			r.rewordBackToReapply("blue edited the re-applied sentence while a lens minted over it, which ends the estoppel (F12)")
		case err == nil:
			r.noteActErr("a mint quoting red's text blue re-applied verbatim was admitted; the re-application estops again (F12): " + firstLine(out))
		default:
			r.noteActErr("a mint quoting red's re-applied text, its estoppel standing, was refused, and not by estoppel: " + firstLine(out))
		}
		return true
	}
	return false
}

// rewordBackToReapply returns the pair to blue's re-application, a later edit having ended its
// estoppel before a lens's mint over it, and counts why.
func (r *runner) rewordBackToReapply(why string) {
	r.mu.Lock()
	if r.reword.stage == rewordRestored {
		r.reword.stage = rewordLanded
	}
	r.mu.Unlock()
	r.noteEstoppelMiss(why)
}

// noteReadingMintLandedPlain notes an admitted plain mint's gap as seatID's (a find, spending its
// stock), and returns the gap id.
func (r *runner) noteReadingMintLandedPlain(seatID, out string) string {
	id := gapIDOf(out)
	if id != "" {
		r.noteMint(seatID, id)
	}
	return id
}

// THE REWORD PAIR, run outside the release gate on the drive's own sittings: red prescribes a swap of
// lane 2's phrasing, blue accepts it verbatim, blue rewords the sentence, red's mint of the reworded
// text lands, blue puts its words back and accepts red's text again, and red's mint is refused again.
func TestARewordEndsEstoppelAndAReapplyRestoresIt(t *testing.T) {
	r := stagedRunner(t)
	r.forceEstoppel = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, lens, other = "blue-respond", "red-lens-logic", "red-lens-adversary"
	r.sit("lens", lens)
	gap := r.mint(lens)
	if gap == "" || r.actErr != "" {
		t.Fatalf("the lens's prescribing mint: %q (%s)", gap, r.actErr)
	}
	r.sit("blue", blue)
	r.blueRespondTo(blue, []string{gap})
	if r.actErr != "" {
		t.Fatalf("blue's verbatim apply: %s", r.actErr)
	}
	if g, _, _ := r.aVerbatimGap(); g != gap {
		t.Fatalf("blue applied %q verbatim, want %s", g, gap)
	}
	r.sit("lens", other)
	r.mintEstopped(other)
	// TWO ENTRIES: the probe's, and the found-source mint's over the same text (readOverPrescription).
	if r.actErr != "" || r.estoppelEntries(other) != 2 {
		t.Fatalf("the ordinary estopped mint and the reading over it: %d entries, want 2 (%s)", r.estoppelEntries(other), r.actErr)
	}
	r.sit("blue", blue)
	r.rewordAppliedText(blue)
	if r.actErr != "" || r.estoppelRewordStage() != rewordReworded {
		t.Fatalf("blue's reword reached stage %d (%s)", r.estoppelRewordStage(), r.actErr)
	}
	r.sit("lens", other)
	r.mintEstopped(other)
	if r.actErr != "" || r.reading.estoppelEndedByReword != 1 || r.estoppelRewordStage() != rewordLanded {
		t.Fatalf("red's mint of the reworded text: ended %d, stage %d (%s)", r.reading.estoppelEndedByReword, r.estoppelRewordStage(), r.actErr)
	}
	r.sit("blue", blue)
	r.rewordAppliedText(blue)
	if r.actErr != "" || r.estoppelRewordStage() != rewordRestored {
		t.Fatalf("blue's re-application reached stage %d (%s)", r.estoppelRewordStage(), r.actErr)
	}
	r.sit("lens", other)
	r.mintEstopped(other)
	if r.actErr != "" || r.reading.estoppelRestored != 1 || r.estoppelRewordStage() != rewordDone {
		t.Fatalf("red's mint of the re-applied text: restored %d, stage %d (%s)", r.reading.estoppelRestored, r.estoppelRewordStage(), r.actErr)
	}
	// THE READING OVER THE PRESCRIPTION: its mint estopped, its debt cleared by the lens's own reading
	// of another source, since the prescribing gap is not its own.
	if r.reading.foundSourceMintsEstopped == 0 || r.reading.debtsDischarged["reread"] == 0 {
		t.Errorf("readings over the prescription: %d mints estopped, %d debts cleared by re-read", r.reading.foundSourceMintsEstopped, r.reading.debtsDischarged["reread"])
	}
}
