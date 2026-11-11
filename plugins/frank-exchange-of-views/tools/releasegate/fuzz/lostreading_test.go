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

// THE ONE-GAP REFUSAL OVER A GAP THAT CLOSED UNDER THE DRIVE: the refusal names a sibling's open gap
// of the class on the sentence, and the board and the debt are read after it; a close landing between
// — the minter's own, in its parallel sitting, or the bench's — leaves the debt standing by the
// record's rule (a gap minted before the reading's sitting discharges only while open), which the
// drive notes rather than fails. The gap still open and the debt standing is the failure, naming
// where the record places the gap.
func TestAOneGapRefusalOverAGapClosedUnderTheDriveIsNoted(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const a, b = "red-lens-architecture", "red-lens-logic"
	s := foundSourceSentences[0]
	r.sit("lens", b)
	if r.readFound(b, foundSourceURL(b), s, "refutes") == 0 {
		t.Fatalf("b's reading: %s", r.actErr)
	}
	g := r.mintOnReading(b, classContradicts, quoteArgs(s), 0, "quote").gap
	if g == "" || r.actErr != "" {
		t.Fatalf("b's mint: %q (%s)", g, r.actErr)
	}
	r.sit("lens", a)
	if r.readFound(a, foundSourceURL(a), s, "refutes") == 0 {
		t.Fatalf("a's reading: %s", r.actErr)
	}
	if _, stands := r.debtOf(a, s); stands {
		t.Fatal("a's debt stands over b's open gap on the sentence")
	}
	args := append(readingMintArgs(a, classContradicts), quoteArgs(s)...)
	out, err := r.exec(args...)
	if err == nil || !strings.Contains(out, oneGapWords) || !strings.Contains(out, g) {
		t.Fatalf("a's mint over b's gap: %v %s", err, firstLine(out))
	}
	// THE OPEN GAP: the debt is discharged, and the refusal is read as the other seat's gap.
	r.oneGapOnReading(a, classContradicts, out, args, "quote")
	if r.actErr != "" || r.reading.debtsDischarged["other_seat_gap"] != 1 {
		t.Fatalf("over the open gap: discharged %d (%s)", r.reading.debtsDischarged["other_seat_gap"], r.actErr)
	}
	// THE GAP CLOSED BETWEEN THE REFUSAL AND THE READ: a's debt stands, and the drive notes it.
	r.closeGap(b, g, false)
	if r.actErr != "" {
		t.Fatalf("b's close of %s: %s", g, r.actErr)
	}
	if _, stands := r.debtOf(a, s); !stands {
		t.Fatal("a's debt does not stand once b's gap, minted before a's sitting, is closed")
	}
	r.oneGapOnReading(a, classContradicts, out, args, "quote")
	if r.actErr != "" || r.reading.debtsDischarged["other_seat_gap"] != 1 {
		t.Fatalf("over the closed gap: discharged %d (%s)", r.reading.debtsDischarged["other_seat_gap"], r.actErr)
	}
}
