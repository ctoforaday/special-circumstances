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
	const tok = "<!--cite:C-00abc123-->"
	before := "# H\n\nThe sky is blue and the grass is green" + tok + ".\n\nAnother sentence entirely.\n"

	// The inversion, which is the case worth naming: same anchor, opposite claim.
	after := "# H\n\nThe sky is green and the grass is on fire" + tok + ".\n\nAnother sentence entirely.\n"
	got := ReopenedAnchors(before, after)
	if len(got) != 1 || got[0] != "C-00abc123" {
		t.Errorf("ReopenedAnchors = %v, want [C-00abc123] — the citation now backs the opposite of what it was placed against", got)
	}

	// AN UNTOUCHED ANCHOR IS NOT REOPENED. Reopening everything on every edit is the same as
	// reopening nothing: a reader learns to skip it.
	elsewhere := "# H\n\nThe sky is blue and the grass is green" + tok + ".\n\nA completely rewritten tail.\n"
	if got := ReopenedAnchors(before, elsewhere); len(got) != 0 {
		t.Errorf("ReopenedAnchors = %v, want none — an edit elsewhere in the document did not move this reference's text", got)
	}

	// A SECOND ANCHOR ARRIVING IN THE SAME SENTENCE IS NOT A CHANGE TO THE FIRST. Anchors are
	// stripped before comparing, or every cite would reopen its own neighbours.
	twin := "# H\n\nThe sky is blue and the grass is green" + tok + "<!--fx:F-009f2a1c-->.\n\nAnother sentence entirely.\n"
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

// THE QUOTE DECIDES WHICH ANCHORS THE SPAN HOLDS, anchor by anchor (settleAbuttingAnchor).
//
// Attach places an anchor after a quote's last content character, so the anchors on a sentence stand
// before its terminator — where a whole-sentence quote, its trailing punctuation trimmed, stops. An
// anchor the quote does not carry is outside the span and the splice never touches it; one the quote
// carries is inside, with every anchor of the run before it. Nothing here refuses: replay runs the
// same locate on every recorded edit.
func TestTheQuoteDecidesWhichAnchorsTheSpanHolds(t *testing.T) {
	const a, b, g = "<!--fx:F-000000a1-->", "<!--cite:C-000000b2-->", "<!--gap:G-000000c3-->"
	const body = "# H\n\nThe sky is blue"
	for _, c := range []struct{ name, run, quote, held string }{
		{"one anchor, the sentence quoted without it", a, "The sky is blue.", ""},
		{"one anchor, the sentence quoted without it or its terminator", a, "The sky is blue", ""},
		{"a run holding a gap anchor, quoted without it", a + g + b, "The sky is blue.", ""},
		{"one anchor, quoted as printed", a, "The sky is blue" + a + ".", a},
		{"a run, quoted whole", a + g + b, "The sky is blue" + a + g + b + ".", a + g + b},
		{"a quote that stops between two anchors takes the first", a + b, "The sky is blue" + a, a},
		{"a quote that carries only the second takes both: a span is one stretch of text", a + b, "The sky is blue" + b, a + b},
	} {
		t.Run(c.name, func(t *testing.T) {
			report := body + c.run + ".\n\nA second sentence with no anchor.\n"
			start, end, err := LocateUniqueReplacing("blue edit", report, c.quote)
			if err != nil {
				t.Fatalf("LocateUniqueReplacing = %v, want the span located", err)
			}
			if got, want := report[start:end], "The sky is blue"+c.held; got != want {
				t.Errorf("the span is %q, want %q", got, want)
			}
		})
	}

	// A sentence with no anchor, and a fragment that stops short of one, locate as they are quoted.
	report := body + a + ".\n\nA second sentence with no anchor.\n"
	for _, q := range []string{"The sky", "A second sentence with no anchor."} {
		start, end, err := LocateUniqueReplacing("blue edit", report, q)
		if err != nil || strings.Contains(report[start:end], "<!--") {
			t.Errorf("%q located %q, %v", q, report[start:end], err)
		}
	}
}
