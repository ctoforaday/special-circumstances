package fuzz

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE DEBT WHOSE SENTENCE A SUB-BOUND EDIT TOOK OUT: the drawn edit swaps a 16-character phrasing
// of lane 2's sentence, under the lost-quote lookup's floor (F10), so a lens's found reading of that
// sentence is lost with no edit named — the item keeps the located wording, and the plan's residue
// clears it by the lens's reading of the SAME url against a sentence the report holds. The drive
// reads the same url rather than quoting words the report no longer holds, in the sitting after the
// edit and in the sitting the reading was made in (the retry key admits another sentence).
func TestASubBoundEditLeavesTheDebtTheSameURLClears(t *testing.T) {
	const blue, lens = "blue-respond", "red-lens-voice"
	line := "The cost is rising over time across the whole sampled corpus."
	stage := func(t *testing.T) (*runner, string) {
		r := stagedRunner(t)
		if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
			t.Fatalf("record the lanes' parts: %v", err)
		}
		r.sit("lens", lens)
		url := foundSourceURL(lens)
		if r.readFound(lens, url, line, "refutes") == 0 {
			t.Fatalf("the reading: %s", r.actErr)
		}
		r.sit("blue", blue)
		if _, ok := r.do("edit", blue).set("--quote", "rising over time").set("--new", "climbing sharply").
			set("--reason", "fuzz edit: sharper phrasing, answering no gap").admit(); !ok {
			t.Fatalf("blue's edit: %s", r.actErr)
		}
		b, stands := r.debtOf(lens, line)
		if !stands || b.Lost() != 0 || b.Reworded() != "" || !strings.Contains(b.WorkItem(), "finds the source against its sentence") {
			t.Fatalf("after the sub-bound edit the debt stands %v, lost %d, reworded %q: want standing, no edit named, the located wording", stands, b.Lost(), b.Reworded())
		}
		if r.heldOnce(line) {
			t.Fatal("the report still holds the reading's sentence once")
		}
		return r, url
	}
	for _, same := range []bool{false, true} {
		name := "the next sitting"
		if same {
			name = "the same sitting"
		}
		t.Run(name, func(t *testing.T) {
			r, url := stage(t)
			if !same {
				r.sit("lens", lens)
			}
			r.answerReadingDebt(lens)
			if r.actErr != "" || len(r.readingDebts()) != 0 || r.reading.debtsDischarged["reread"] != 1 {
				t.Fatalf("after the lens's answer: %d debts, reread %d (%s)", len(r.readingDebts()), r.reading.debtsDischarged["reread"], r.actErr)
			}
			if ref, err := record.EnvelopeRefusal(r.run(), lens); err != nil || ref != "" {
				t.Errorf("the envelope after the answer: %q (%v)", ref, err)
			}
			// THE SAME URL, another sentence: the one identity the residue leaves the reading.
			b, ok := r.board()
			if !ok {
				t.Fatal("the board did not read")
			}
			reads := 0
			for _, e := range b.Events {
				if v, isV := recordpb.BodyAs[*recordpb.Verify](e); isV && v.GetUrl() == url && e.GetSeatId() == lens {
					reads++
				}
			}
			if reads != 2 {
				t.Errorf("readings of %s by the lens: %d, want the lost one and its re-read", url, reads)
			}
		})
	}
}
