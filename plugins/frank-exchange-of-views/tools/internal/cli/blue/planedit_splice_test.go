package blue

import (
	"strings"
	"testing"
)

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
