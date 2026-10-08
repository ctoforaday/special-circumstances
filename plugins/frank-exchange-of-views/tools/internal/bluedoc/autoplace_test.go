package bluedoc

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// AutoPlace PUTS AN ANCHOR BACK ONLY WHERE ITS SENTENCE SURVIVES WORD FOR WORD, ONCE, and there at
// the place the anchor held in it: mid-sentence, inside a run of anchors in the run's order, across
// a re-wrapped line, never on a heading. Every other replacement comes back as it went in.
func TestAutoPlace(t *testing.T) {
	const a, b = "<!--fx:f-aaaa1111-->", "<!--cite:c-bbbb2222-->"
	for _, c := range []struct{ name, span, new, want string }{
		{"mid-sentence", "The sieve is fast" + a + " and simple", "It is old. The sieve is fast and simple", "It is old. The sieve is fast" + a + " and simple"},
		{"after the anchors that preceded it", "Wet" + a + b, "Dry. Wet" + a, "Dry. Wet" + a + b},
		{"a whole run, in order", "Wet" + a + b, "Dry. Wet", "Dry. Wet" + a + b},
		{"across a re-wrapped line", "Costs\nrose" + a, "Costs rose", "Costs rose" + a},
		{"twice in new", "Costs rose" + a + ".", "Costs rose sharply. Costs rose.", "Costs rose sharply. Costs rose."},
		{"rewritten", "Costs rose" + a, "Costs fell", "Costs fell"},
		{"its terminator changed", "Costs rose" + a + "!", "Costs rose?", "Costs rose?"},
		{"across a blank line in new", "Costs rose" + a, "Costs\n\nrose", "Costs\n\nrose"},
		{"carried already", "Costs rose" + a, "Costs rose" + a + ". More.", "Costs rose" + a + ". More."},
		{"onto a heading", "Costs rose" + a + ".", "## Costs rose.\n\nMore.", "## Costs rose.\n\nMore."},
	} {
		if got := AutoPlace(c.span, c.new); got != c.want {
			t.Errorf("%s: AutoPlace(%q, %q) = %q, want %q", c.name, c.span, c.new, got, c.want)
		}
	}
}

// An anchor dropped with no prose around it is refused as bare, not "on the sentence """.
func TestTheDropRefusalSaysWhenTheAnchorIsBare(t *testing.T) {
	_, err := AnchorsTransitUnchanged("blue edit", "<!--fx:f-aaaa1111-->", "Gone.")
	if err == nil || !strings.Contains(err.Error(), anchor.Label("f-aaaa1111")+" bare of any sentence") {
		t.Errorf("refusal = %v, want it to say the anchor is bare", err)
	}
}

// Of several anchors left out, the refusal names the first in the span, every time.
func TestTheDropRefusalNamesTheFirstAnchorLeftOut(t *testing.T) {
	span := "One<!--fx:f-aaaa0001-->. Two<!--fx:f-aaaa0002-->. Three<!--fx:f-aaaa0003-->. Four<!--fx:f-aaaa0004-->."
	for range 8 {
		if _, err := AnchorsTransitUnchanged("blue edit", span, "Gone."); err == nil || !strings.Contains(err.Error(), anchor.Label("f-aaaa0001")+" on the sentence \"One.\"") {
			t.Fatalf("refusal = %v, want it to name f-aaaa0001 on \"One.\"", err)
		}
	}
}
