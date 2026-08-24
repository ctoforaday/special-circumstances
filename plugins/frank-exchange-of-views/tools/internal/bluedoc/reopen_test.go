package bluedoc

import (
	"strings"
	"testing"
)

// A REFERENCE WHOSE REFERENT MOVED IS REOPENED.
//
// The anchor is never lost — that promise is enforced elsewhere and holds. This is the other
// half: an anchor that SURVIVES onto rewritten prose backs a sentence nobody read, and until
// now nothing said so.
func TestReopenedAnchorsCatchesTextMovingUnderAReference(t *testing.T) {
	const tok = "<!--cite:c-abc123-->"
	before := "# H\n\nThe sky is blue and the grass is green" + tok + ".\n\nAnother sentence entirely.\n"

	// The inversion, which is the case worth naming: same anchor, opposite claim.
	after := "# H\n\nThe sky is green and the grass is on fire" + tok + ".\n\nAnother sentence entirely.\n"
	got := ReopenedAnchors(before, after)
	if len(got) != 1 || got[0] != "c-abc123" {
		t.Errorf("ReopenedAnchors = %v, want [c-abc123] — the citation now backs the opposite of what it was placed against", got)
	}

	// AN UNTOUCHED ANCHOR IS NOT REOPENED. Reopening everything on every edit is the same as
	// reopening nothing: a reader learns to skip it.
	elsewhere := "# H\n\nThe sky is blue and the grass is green" + tok + ".\n\nA completely rewritten tail.\n"
	if got := ReopenedAnchors(before, elsewhere); len(got) != 0 {
		t.Errorf("ReopenedAnchors = %v, want none — an edit elsewhere in the document did not move this reference's text", got)
	}

	// A SECOND ANCHOR ARRIVING IN THE SAME SENTENCE IS NOT A CHANGE TO THE FIRST. Anchors are
	// stripped before comparing, or every cite would reopen its own neighbours.
	twin := "# H\n\nThe sky is blue and the grass is green" + tok + "<!--fx:f-L1-F1-->.\n\nAnother sentence entirely.\n"
	if got := ReopenedAnchors(before, twin); len(got) != 0 {
		t.Errorf("ReopenedAnchors = %v, want none — a neighbouring anchor is not a change to this one's referent", got)
	}

	// A DROPPED ANCHOR IS NOT REOPENED — it is dropped, and that is a refusal the caller already
	// makes. Reporting both would file one fault as two.
	dropped := "# H\n\nThe sky is blue and the grass is green.\n\nAnother sentence entirely.\n"
	if got := ReopenedAnchors(before, dropped); len(got) != 0 {
		t.Errorf("ReopenedAnchors = %v, want none — a dropped anchor is a different fault with its own refusal", got)
	}

	// Whitespace-only reflow is not a change of referent.
	reflowed := strings.Replace(before, "blue and the grass", "blue  and   the grass", 1)
	if got := ReopenedAnchors(before, reflowed); len(got) != 0 {
		t.Errorf("ReopenedAnchors = %v, want none — reflowing whitespace did not change what the sentence says", got)
	}
}

// A QUOTE MAY NOT STOP JUST SHORT OF THE ANCHOR ON THE TEXT IT REPLACES.
//
// The markers are the mechanism and they are VISIBLE for this reason: `show report` prints them
// as they are, so a seat rewriting a sentence sees the token in it and copies it into --new the
// way it copies every other character. An edit mimics how anyone edits a document — quote what is
// there, write what should be there.
//
// The tolerance broke that. normalizeQuote skips annotation spans, so a quote omitting the marker
// still matched, and the located span then ENDED BEFORE it: the transit guard never fired on a
// whole-sentence edit, and the marker was left beside prose it was never placed against.
func TestAQuoteMayNotStopShortOfTheAnchorItIsRewriting(t *testing.T) {
	const tok = "<!--cite:c-abc123-->"
	report := "# H\n\nThe sky is blue and the grass is green" + tok + ".\n\nA second sentence with no anchor.\n"

	// Quoting the sentence WITHOUT its anchor is refused, and the refusal names the token to carry.
	_, _, err := LocateUniqueReplacing("blue edit", report, "The sky is blue and the grass is green.")
	if err == nil {
		t.Fatal("a quote that stops just before the anchor on its own sentence was accepted — the marker would be stranded beside rewritten prose")
	}
	if !strings.Contains(err.Error(), tok) {
		t.Errorf("the refusal does not print the token the seat must carry: %v", err)
	}

	// Quoting it WITH the anchor, as `show report` prints it, locates normally.
	if _, _, err := LocateUniqueReplacing("blue edit", report, "The sky is blue and the grass is green"+tok+"."); err != nil {
		t.Errorf("the sentence quoted AS PRINTED was refused: %v", err)
	}

	// A FRAGMENT that does not reach the anchor is still allowed — that is the class of edit the
	// strict rule would otherwise cost, and the anchor keeps its position while `reopened` records
	// that its sentence moved.
	if _, _, err := LocateUniqueReplacing("blue edit", report, "The sky is blue"); err != nil {
		t.Errorf("a fragment edit that does not touch the anchored text was refused: %v", err)
	}

	// And a sentence with no anchor is unaffected.
	if _, _, err := LocateUniqueReplacing("blue edit", report, "A second sentence with no anchor."); err != nil {
		t.Errorf("an unanchored sentence was refused: %v", err)
	}
}
