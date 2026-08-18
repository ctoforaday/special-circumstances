package blue

import (
	"strings"
	"testing"
)

// The exact splice damage measured on the 2026-08-04 smoke, where blue spent 6 of 17 round-2
// edits repairing punctuation it had created itself. Each case is a seam an edit manufactures.
func TestTidySeamRemovesSpliceArtifacts(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"double period", "It is prime.. The next claim.", "It is prime. The next claim."},
		{"double colon", "Sources:: Cuemath", "Sources: Cuemath"},
		{"double semicolon", "one;; two", "one; two"},
		{"double comma", "a,, b", "a, b"},
		{"space before period", "It is prime . The next", "It is prime. The next"},
	} {
		t.Run(c.name, func(t *testing.T) {
			// the seam sits at the artifact
			at := 0
			for i := 1; i < len(c.in); i++ {
				if c.in[i] == c.in[i-1] && (c.in[i] == '.' || c.in[i] == ':' || c.in[i] == ';' || c.in[i] == ',') {
					at = i
					break
				}
				if c.in[i-1] == ' ' && (c.in[i] == '.' || c.in[i] == ',') {
					at = i
					break
				}
			}
			got, changed := tidySeam(c.in, at)
			if got != c.want {
				t.Errorf("tidySeam = %q, want %q", got, c.want)
			}
			if !changed {
				t.Error("changed = false, want true")
			}
		})
	}
}

// CONTENT IS NEVER TOUCHED. An ellipsis and deliberate emphasis are things a human meant to
// write; a prose normalizer that "fixes" them is worse than the artifacts it removes.
func TestTidySeamLeavesContentAlone(t *testing.T) {
	for _, s := range []string{
		"an ellipsis... is content",
		"emphasis!! stays",
		"a question?? stays",
		"an anchor<!--fx:f-abc--> is untouched",
		"no artifact here at all",
	} {
		for at := 1; at < len(s); at++ {
			if got, changed := tidySeam(s, at); changed || got != s {
				t.Fatalf("tidySeam(%q, %d) altered content -> %q", s, at, got)
			}
		}
	}
}

// AMBIGUITY MUST BE REFUSED, NOT GUESSED. LocateSpan takes the FIRST match and says nothing, so a
// quote appearing twice silently edits whichever came first. blue is explicitly instructed to
// propagate a correction to every site stating a claim, so repeated text is the expected shape of
// a real report — this is the case where guessing is worst.
func TestPlanEditRefusesAnAmbiguousSpan(t *testing.T) {
	const rep = "# H\n\nThe value is stable.\n\nElsewhere, again: The value is stable.\n"
	if _, err := planEdit(rep, "The value is stable.", "The value is steady."); err == nil {
		t.Fatal("an --quote span matching twice was accepted — it silently edits the first site")
	} else if !strings.Contains(err.Error(), "MORE THAN ONCE") {
		t.Errorf("error = %v, want it to name the ambiguity", err)
	}
	// A quote carrying enough context to be unique still applies.
	out, err := planEdit(rep, "Elsewhere, again: The value is stable.", "Elsewhere, again: The value is steady.")
	if err != nil {
		t.Fatalf("a uniquely-quoted span was refused: %v", err)
	}
	if !strings.Contains(out, "Elsewhere, again: The value is steady.") || !strings.Contains(out, "\n\nThe value is stable.") {
		t.Errorf("the disambiguated edit hit the wrong site:\n%s", out)
	}
}

// ANCHORS TRANSIT; THEY ARE NEVER CREATED OR DESTROYED. The old rule ("reject any span containing
// an anchor") deadlocked against the uniqueness guard: when a word appears twice and the only
// disambiguating context carries red's anchor, the minimal quote is ambiguous and the contextual
// quote is anchor-spanning, so the site RED FLAGGED becomes the one site blue cannot edit.
func TestAnchorsMayTransitButNeverChange(t *testing.T) {
	const rep = "# H\n\nThe methods agree independently<!--fx:f-abc--> here.\n\nOthers agree independently there.\n"

	// The deadlock case: disambiguating context carries the anchor, carried across verbatim.
	out, err := planEdit(rep,
		"The methods agree independently<!--fx:f-abc--> here.",
		"The methods agree independent<!--fx:f-abc--> here.")
	if err != nil {
		t.Fatalf("an edit carrying its anchor across was refused: %v", err)
	}
	if strings.Count(out, "<!--fx:f-abc-->") != 1 {
		t.Errorf("anchor count changed: %q", out)
	}
	if !strings.Contains(out, "agree independent<!--fx:f-abc--> here") {
		t.Errorf("the anchored occurrence was not edited:\n%s", out)
	}
	if !strings.Contains(out, "Others agree independently there.") {
		t.Error("the OTHER occurrence must be untouched — that was the silent-mis-target risk")
	}

	// Dropping it is still refused.
	if _, err := planEdit(rep,
		"The methods agree independently<!--fx:f-abc--> here.",
		"The methods agree independent here."); err == nil {
		t.Error("an edit that DROPPED its anchor was accepted")
	}
	// Duplicating it is still refused.
	if _, err := planEdit(rep,
		"The methods agree independently<!--fx:f-abc--> here.",
		"The methods agree independent<!--fx:f-abc--> and<!--fx:f-abc--> here."); err == nil {
		t.Error("an edit that DUPLICATED its anchor was accepted")
	}
	// Introducing one that was not in the span is still refused. (The id must be valid hex to BE
	// an anchor — "f-zzz" is inert text the regex correctly ignores, which is why this uses f-dead.)
	if _, err := planEdit(rep,
		"Others agree independently there.",
		"Others agree independent<!--fx:f-dead--> there."); err == nil {
		t.Error("an edit that INVENTED an anchor was accepted")
	}
}
