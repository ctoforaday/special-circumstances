package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

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
// and its prescription's pair as the mint recorded it.
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
// from the run's RNG.
func (r *runner) rewordAppliedText(seatID string) {
	if !r.forceEstoppel || seatID != "blue-respond" {
		return
	}
	switch r.estoppelRewordStage() {
	case rewordIdle:
		gap, _ := r.aVerbatimGap()
		if gap == "" {
			return
		}
		cur, err := record.RenderRun(r.run())
		if err != nil {
			return
		}
		gid, old, new_ := r.proposalFor(gap, cur)
		if gid == "" || strings.Count(cur, rewordTail) != 1 || !strings.Contains(anchortext.Visible(cur), anchortext.Visible(new_)) {
			return
		}
		if _, ok := r.do("edit", seatID).set("--quote", rewordTail).set("--new", rewordedTail).
			set("--reason", "fuzz: the sentence red's text was applied to, reworded in blue's own words").admit(); !ok {
			return
		}
		if !r.lastEditReopened(seatID, gap) {
			r.noteActErr("blue's reword of the sentence red's fix on " + gap + " was applied to does not list the gap reopened")
			return
		}
		r.mu.Lock()
		r.reword = estoppelReword{stage: rewordReworded, gap: gap, old: old, new_: new_}
		r.mu.Unlock()
	case rewordLanded:
		r.mu.Lock()
		gap, old, new_ := r.reword.gap, r.reword.old, r.reword.new_
		r.mu.Unlock()
		cur, err := record.RenderRun(r.run())
		if err != nil {
			return
		}
		// RED'S WORDS ARE PUT BACK where the applied text still stands, so the accept applies again.
		if vis := anchortext.Visible(cur); !strings.Contains(vis, anchortext.Visible(old)) && strings.Contains(vis, anchortext.Visible(new_)) {
			if _, ok := r.do("edit", seatID).set("--quote", new_).set("--new", old).
				set("--reason", "fuzz: blue's words again, before red's are re-applied").admit(); !ok {
				return
			}
		}
		if _, err := r.do("edit", seatID).set("--answers", gap).bare("--accept").
			set("--reason", "fuzz: accepting red's prescribed fix on "+gap+" again").result(); err != nil {
			if _, ok := r.do("edit", seatID).set("--answers", gap).set("--quote", old).set("--new", new_).
				set("--reason", "fuzz: applying red's proposed text verbatim again").admit(); !ok {
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
func (r *runner) mintRewordedText(seatID, gapID, sentence string) bool {
	switch r.estoppelRewordStage() {
	case rewordReworded, rewordLanded:
		args := []string{"--json", "mint", "--seat-id", seatID, "--class", "spec_underspecification",
			"--problem", "fuzz: raising a defect against text blue reworded", "--check-kind", "document",
			"--check", "acc", "--fix", dirCounter, "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
			"--quote", sentence, "--reason", "fuzz: the reworded text is blue's, not red's prescription on " + gapID}
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
		case strings.Contains(out, estoppelWords):
			r.noteActErr("a mint quoting applied text blue has since reworded was refused by estoppel; a reword ends it (F12): " + firstLine(out))
		default:
			noteEither(args[:2], "a mint of reworded text, refused by another rule", false)
		}
		return true
	case rewordRestored:
		r.mu.Lock()
		r.reword.stage = rewordDone
		r.mu.Unlock()
		out, err := r.exec("--json", "mint", "--seat-id", seatID, "--class", "spec_underspecification",
			"--problem", "fuzz: raising a defect against text I prescribed", "--check-kind", "document",
			"--check", "acc", "--fix", "reword it", "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
			"--quote", sentence, "--reason", "fuzz: estoppel probe — this quotes red's re-applied prescription on "+gapID)
		switch {
		case err == nil:
			r.noteActErr("a mint quoting red's text blue re-applied verbatim was admitted; the re-application estops again (F12): " + firstLine(out))
		case !strings.Contains(out, estoppelWords):
			r.noteActErr("a mint quoting red's re-applied text was refused, and not by estoppel: " + firstLine(out))
		default:
			r.countReading(func(c *readingCounts) { c.estoppelRestored++ })
		}
		return true
	}
	return false
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
	if g, _ := r.aVerbatimGap(); g != gap {
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
