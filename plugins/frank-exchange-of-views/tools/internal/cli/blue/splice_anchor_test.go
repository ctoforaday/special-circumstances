package blue

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// A SENTENCE WHOSE ANCHOR PRECEDES ITS PERIOD SURVIVES AN EDIT INTACT (#525).
//
// `normalizeQuote` trims a quote's trailing punctuation, so the located span stops before the
// sentence's period and that period is never replaced. In 40 of the 43 anchors across the four
// archived runs the anchor sits immediately after the last WORD — before that period — so a
// replacement ending in "." lands against "<!--fx:…-->." and the two marks are never adjacent.
// Measured through the real command before the fix:
//
//	"The cost is falling.<!--fx:f-abc123-->. Volume grows steadily."
//
// a doubled period AND red's marker displaced out of the sentence it annotates. The controls
// below carry the fixture with no anchor, which is what separates the two defects: the doubled
// mark needs one, the deletion residue does not.
func TestSpliceAroundAnAnchoredSentenceEnd(t *testing.T) {
	const anchored = "Intro.\n\nThe cost is rising over time<!--fx:f-abc123-->. Volume grows steadily.\n"
	const bare = "Intro.\n\nThe cost is rising over time. Volume grows steadily.\n"
	for _, tc := range []struct{ name, report, old, new, want string }{
		{
			"a reword keeps the anchor inside its sentence and does not double the period",
			anchored, "The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling<!--fx:f-abc123-->. Volume grows steadily.\n",
		}, {
			"control: the same reword with no anchor was already clean",
			bare, "The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling. Volume grows steadily.\n",
		}, {
			// The anchor STAYS: it is immortal, and removing it is not blue's call — see
			// AnchorsTransitUnchanged. What goes is the period the deleted sentence owned.
			"a deletion leaves the anchor but not the orphaned period",
			anchored, "The cost is rising over time.", "",
			"Intro.\n\n<!--fx:f-abc123--> Volume grows steadily.\n",
		}, {
			// Fires with NO anchor, which is what makes it a separate defect rather than a
			// consequence of the one above.
			"control: a deletion with no anchor leaves no orphaned period either",
			bare, "The cost is rising over time.", "",
			"Intro.\n\nVolume grows steadily.\n",
		}, {
			// Two lenses anchoring one sentence is a real corpus shape:
			// "verification<!--fx:f-e4bc25ec--><!--fx:f-73a56bd3-->".
			"a run of abutting anchors is stepped over whole, not one token deep",
			"Intro.\n\nThe cost is rising over time<!--fx:f-abc123--><!--cite:c-d4d4d4-->. Volume grows steadily.\n",
			"The cost is rising over time.", "The cost is falling.",
			"Intro.\n\nThe cost is falling<!--fx:f-abc123--><!--cite:c-d4d4d4-->. Volume grows steadily.\n",
		}, {
			// The anchor is INSIDE the span here, so it must be reproduced and no seam rule
			// applies. Included so a change that over-fires shows up as this case moving.
			"a mid-sentence anchor is untouched",
			"Intro.\n\nThe cost<!--fx:f-abc123--> is rising over time. Volume grows steadily.\n",
			"The cost is rising over time.", "The cost<!--fx:f-abc123--> is falling.",
			"Intro.\n\nThe cost<!--fx:f-abc123--> is falling. Volume grows steadily.\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := planEdit(tc.report, tc.old, tc.new)
			if err != nil {
				t.Fatalf("planEdit = %v, want the edit to apply", err)
			}
			if got != tc.want {
				t.Errorf("planEdit =\n  %q\nwant\n  %q", got, tc.want)
			}
		})
	}
}

// SkipRun MUST SEE EXACTLY THE TOKENS THE PROTECTION SWEEP SEES.
//
// tidySeam steps over the anchor layer to find the seam. claimcount's three regexes decide
// which tokens are PROTECTED. A token one reader steps over and the other refuses to see is a
// span where punctuation gets tidied around something that is not an anchor — two readers of
// one vocabulary, drifting, which is the defect this repository keeps re-finding. The two
// patterns are hand-written in different packages (anchor is a stdlib-only leaf), so their
// agreement is pinned here rather than by construction.
func TestSkipRunAgreesWithTheProtectionSweep(t *testing.T) {
	for _, tok := range []string{
		"<!--fx:f-abc123-->", "<!--cite:c-d4d4d4-->", "<!--proof:p-0f0f0f-->",
		"<!--fx:f-seed01-->",   // "s" is not hex — the malformed id that made the fuzzer vacuous
		"<!--fx:f-ABC123-->",   // uppercase is not [0-9a-f]
		"<!--fx:abc123-->",     // no class prefix on the id
		"<!--note:f-abc123-->", // not an anchor class
		"<!--fx:f--->",         // empty id
		"<!-- fx:f-abc123 -->", // spaced: not the minted spelling
	} {
		t.Run(tok, func(t *testing.T) {
			protected := len(claimcount.ProtectedAnchorIDs(tok)) > 0
			skipped := anchor.SkipRun(tok, 0) == len(tok)
			if protected != skipped {
				t.Errorf("claimcount protects=%v but anchor.SkipRun consumes=%v — the two readers disagree about %q", protected, skipped, tok)
			}
		})
	}
	// And a run of two is consumed whole, not one token deep.
	pair := "<!--fx:f-abc123--><!--cite:c-d4d4d4-->"
	if got := anchor.SkipRun(pair+"tail", 0); got != len(pair) {
		t.Errorf("SkipRun over an abutting pair = %d, want %d (%q)", got, len(pair), pair)
	}
	// A token that merely STARTS like one leaves the index untouched rather than half-consuming.
	if got := anchor.SkipRun("<!--fx:f-zz-->x", 0); got != 0 {
		t.Errorf("SkipRun over a malformed token = %d, want 0", got)
	}
}

// THE REFUSAL THAT SENT A SEAT IN A CIRCLE.
//
// Reproducing an anchor is MANDATORY when it sits inside the replaced span and REFUSED when it
// sits just outside, and the seat cannot see which — the trimmed trailing punctuation moves the
// boundary under a quote that appeared to contain the anchor. Each refusal instructed the seat
// to do what the other forbids. blue-respond-r2 in 2026-08-22_record-store-authority met this
// and concluded the removal was impossible, arguing risk-acceptance on R1-1 instead (#525).
func TestTheAdjacentAnchorRefusalSaysWhichSideItIsOn(t *testing.T) {
	const rep = "Intro.\n\nThe cost is rising over time<!--fx:f-abc123-->. Volume grows steadily.\n"

	// OUTSIDE the span: reproducing it must be refused, and the refusal must say it is outside.
	_, err := planEdit(rep, "The cost is rising over time.", "The cost is falling<!--fx:f-abc123-->.")
	if err == nil {
		t.Fatal("reproducing an anchor that sits outside the span was accepted — it would duplicate one")
	}
	for _, want := range []string{"just OUTSIDE the span", "preserved with no action", "f-abc123"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q:\n%v", want, err)
		}
	}
	// And it must NOT repeat the generic line, which is the one that sends the seat back round.
	if strings.Contains(err.Error(), "never typed into a replacement") {
		t.Errorf("the adjacent case still gets the generic invention message:\n%v", err)
	}

	// INSIDE the span: the generic message is correct and must survive, naming the token to copy.
	const mid = "Intro.\n\nThe cost<!--fx:f-abc123--> is rising over time. Volume grows steadily.\n"
	_, err = planEdit(mid, "The cost is rising over time.", "The cost is falling.")
	if err == nil {
		t.Fatal("dropping an anchor from inside the span was accepted")
	}
	for _, want := range []string{"Reproduce it EXACTLY", "<!--fx:f-abc123-->"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the drop refusal does not carry %q:\n%v", want, err)
		}
	}

	// A GENUINE INVENTION — an anchor nowhere near the span — still gets the flat prohibition.
	_, err = planEdit(rep, "Volume grows steadily.", "Volume falls<!--cite:c-d4d4d4-->.")
	if err == nil {
		t.Fatal("inventing an anchor was accepted")
	}
	if !strings.Contains(err.Error(), "never typed into a replacement") {
		t.Errorf("a real invention lost the prohibition message:\n%v", err)
	}
}
