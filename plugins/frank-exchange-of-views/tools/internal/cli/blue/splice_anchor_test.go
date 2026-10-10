package blue

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
)

// A SENTENCE WHOSE ANCHOR PRECEDES ITS PERIOD SURVIVES AN EDIT INTACT (#525).
//
// `normalizeQuote` trims a quote's trailing punctuation, so the located span stops before the
// sentence's period and that period is never replaced. In 40 of the 43 anchors across the four
// archived runs the anchor sits immediately after the last WORD — before that period — so a naive
// replacement ending in "." lands against "<!--fx:…-->." and the two marks are never adjacent.
// Measured through the real command before the fix:
//
//	"The cost is falling.<!--fx:F-00abc123-->. Volume grows steadily."
//
// a doubled period AND red's marker displaced out of the sentence it annotates.
//
// # THE QUOTE DECIDES (bluedoc.settleAbuttingAnchor)
//
// An anchor run at the end of the quote that the quote does not carry is OUTSIDE the span: the edit
// replaces the quoted text and the run stays where it stands, after the replacement and before the
// terminator, byte for byte — tidySeam takes the terminator the replacement brings against it. A
// quote that carries the run, as `show report` prints it, has it INSIDE the span: the transit check
// then requires each anchor in --new, where the seat writes it.
//
// Smoke m18 measured the refusal this replaces: 37 of the run's 56 refusals were a quote that
// stopped before the run on its sentence, and every first attempt among them quoted and rewrote
// prose only.
func TestSpliceAroundAnAnchoredSentenceEnd(t *testing.T) {
	const anchored = "Intro.\n\nThe cost is rising over time<!--fx:F-00abc123-->. Volume grows steadily.\n"
	const bare = "Intro.\n\nThe cost is rising over time. Volume grows steadily.\n"
	for _, tc := range []struct{ name, report, old, new, want string }{
		{
			"a reword keeps the anchor inside its sentence and does not double the period",
			anchored, "The cost is rising over time<!--fx:F-00abc123-->.", "The cost is falling<!--fx:F-00abc123-->.",
			"Intro.\n\nThe cost is falling<!--fx:F-00abc123-->. Volume grows steadily.\n",
		}, {
			"control: the same reword with no anchor was already clean",
			bare, "The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling. Volume grows steadily.\n",
		}, {
			// The anchor STAYS: it is immortal, and removing it is not blue's call — see
			// AnchorsTransitUnchanged. What goes is the period the deleted sentence owned.
			"a deletion leaves the anchor but not the orphaned period",
			anchored, "The cost is rising over time<!--fx:F-00abc123-->.", "<!--fx:F-00abc123-->",
			"Intro.\n\n<!--fx:F-00abc123--> Volume grows steadily.\n",
		}, {
			// Fires with NO anchor, which is what makes it a separate defect rather than a
			// consequence of the one above.
			"control: a deletion with no anchor leaves no orphaned period either",
			bare, "The cost is rising over time.", "",
			"Intro.\n\nVolume grows steadily.\n",
		}, {
			// Two lenses anchoring one sentence is a real corpus shape:
			// "verification<!--fx:F-e4bc25ec--><!--fx:F-73a56bd3-->".
			"a run of abutting anchors is carried whole, not one token deep",
			"Intro.\n\nThe cost is rising over time<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
			"The cost is rising over time<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->.",
			"The cost is falling<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->.",
			"Intro.\n\nThe cost is falling<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
		}, {
			"the sentence quoted without its anchor: the anchor stays before the terminator",
			anchored, "The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling<!--fx:F-00abc123-->. Volume grows steadily.\n",
		}, {
			"the same with no terminator on either side",
			anchored, "The cost is rising over time", "The cost is falling",
			"Intro.\n\nThe cost is falling<!--fx:F-00abc123-->. Volume grows steadily.\n",
		}, {
			"a run of several holding a gap anchor stays whole and in order",
			"Intro.\n\nThe cost is rising over time<!--fx:F-00abc123--><!--gap:G-00e5e5e5--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
			"The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling<!--fx:F-00abc123--><!--gap:G-00e5e5e5--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
		}, {
			"a replacement of two sentences leaves the run at its end",
			anchored, "The cost is rising over time.", "The cost is falling. Analysts disagree about why.",
			"Intro.\n\nThe cost is falling. Analysts disagree about why<!--fx:F-00abc123-->. Volume grows steadily.\n",
		}, {
			"a quote that stops between two anchors carries the first and leaves the second",
			"Intro.\n\nThe cost is rising over time<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
			"The cost is rising over time<!--fx:F-00abc123-->", "The cost<!--fx:F-00abc123--> is falling",
			"Intro.\n\nThe cost<!--fx:F-00abc123--> is falling<!--cite:C-00d4d4d4-->. Volume grows steadily.\n",
		}, {
			// The anchor is INSIDE the span here, so it must be reproduced and no seam rule
			// applies. Included so a change that over-fires shows up as this case moving.
			"a mid-sentence anchor is untouched",
			"Intro.\n\nThe cost<!--fx:F-00abc123--> is rising over time. Volume grows steadily.\n",
			"The cost is rising over time.", "The cost<!--fx:F-00abc123--> is falling.",
			"Intro.\n\nThe cost<!--fx:F-00abc123--> is falling. Volume grows steadily.\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, _, err := planEdit(tc.report, tc.old, tc.new)
			if err != nil {
				t.Fatalf("planEdit = %v, want the edit to apply", err)
			}
			if got != tc.want {
				t.Errorf("planEdit =\n  %q\nwant\n  %q", got, tc.want)
			}
		})
	}
}

// ONE RULE AT EVERY ANCHOR POSITION (#525).
//
// Reproducing an anchor is required when it sits inside the replaced span and refused when it sits
// outside, and a quote's trimmed trailing punctuation puts the anchor at a sentence's end outside a
// quote that appears to reach it. The seat need not work out which: the quote it wrote decides, and
// each refusal that remains names the anchor and what to do with it.
func TestOneInstructionHoldsAtEveryAnchorPosition(t *testing.T) {
	const rep = "Intro.\n\nThe cost is rising over time<!--fx:F-00abc123-->. Volume grows steadily.\n"

	// AT THE EDGE, QUOTED: the span follows the quote and the anchor goes where --new writes it.
	got, _, _, err := planEdit(rep, "The cost is rising over time<!--fx:F-00abc123-->.", "The cost<!--fx:F-00abc123--> is falling.")
	if err != nil {
		t.Fatalf("quoting the sentence as printed was refused: %v", err)
	}
	if want := "Intro.\n\nThe cost<!--fx:F-00abc123--> is falling. Volume grows steadily.\n"; got != want {
		t.Errorf("planEdit =\n  %q\nwant\n  %q", got, want)
	}

	// AT THE EDGE, NOT QUOTED: it applies, and the anchor stands where it stood.
	got, applied, _, err := planEdit(rep, "The cost is rising over time.", "The cost is falling.")
	if err != nil {
		t.Fatalf("an edit whose quote stops before the anchor at its sentence's end was refused: %v", err)
	}
	if want := "Intro.\n\nThe cost is falling<!--fx:F-00abc123-->. Volume grows steadily.\n"; got != want || applied != "The cost is falling." {
		t.Errorf("planEdit =\n  %q (applied %q)\nwant\n  %q, the replacement recorded as written", got, applied, want)
	}
	// The record still says the claim under it changed: reopened reads sentences, not spans.
	if re := bluedoc.ReopenedAnchors(rep, got); len(re) != 1 || re[0] != "F-00abc123" {
		t.Errorf("reopened = %v, want the anchor whose sentence the edit rewrote", re)
	}

	// AT THE EDGE, TYPED INTO --new WITHOUT BEING QUOTED: refused, saying it stays and how to move it.
	_, _, _, err = planEdit(rep, "The cost is rising over time.", "The cost is falling<!--fx:F-00abc123-->.")
	if err == nil {
		t.Fatal("an anchor outside the span, typed into the replacement, was accepted — the report would hold it twice")
	}
	for _, want := range []string{"<!--fx:F-00abc123-->", "outside it", "Leave it out of the replacement", "quote it as `show report` prints it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q:\n%v", want, err)
		}
	}

	// A QUOTE THAT SKIPS THE FIRST ANCHOR OF A RUN AND CARRIES THE SECOND holds both, so dropping the
	// first is refused by name.
	const pair = "Intro.\n\nThe cost is rising over time<!--fx:F-00abc123--><!--cite:C-00d4d4d4-->. Volume grows steadily.\n"
	_, _, _, err = planEdit(pair, "The cost is rising over time<!--cite:C-00d4d4d4-->", "The cost is falling<!--cite:C-00d4d4d4-->")
	if err == nil || !strings.Contains(err.Error(), "<!--fx:F-00abc123-->") {
		t.Errorf("a quote carrying only the second anchor of a run = %v, want a refusal naming the first", err)
	}

	// INSIDE the span: the generic message is correct and must survive, naming the token to copy.
	const mid = "Intro.\n\nThe cost<!--fx:F-00abc123--> is rising over time. Volume grows steadily.\n"
	_, _, _, err = planEdit(mid, "The cost is rising over time.", "The cost is falling.")
	if err == nil {
		t.Fatal("dropping an anchor from inside the span was accepted")
	}
	if !strings.Contains(err.Error(), "<!--fx:F-00abc123-->") {
		t.Errorf("the drop refusal does not name the token to copy:\n%v", err)
	}

	// A GENUINE INVENTION — an anchor nowhere near the span — still gets the flat prohibition.
	_, _, _, err = planEdit(rep, "Volume grows steadily.", "Volume falls<!--cite:C-00d4d4d4-->.")
	if err == nil {
		t.Fatal("inventing an anchor was accepted")
	}
	if !strings.Contains(err.Error(), "never typed into a replacement") {
		t.Errorf("a genuine invention lost the flat prohibition:\n%v", err)
	}
}
