package bluedoc

import "testing"

// AutoPlace PUTS AN ANCHOR BACK ONLY WHERE ITS SENTENCE SURVIVES WORD FOR WORD, ONCE, and there at
// the place the anchor held in it: mid-sentence, inside a run of anchors in the run's order, across
// a re-wrapped line. Every other replacement comes back as it went in.
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
	} {
		if got := AutoPlace(c.span, c.new); got != c.want {
			t.Errorf("%s: AutoPlace(%q, %q) = %q, want %q", c.name, c.span, c.new, got, c.want)
		}
	}
}
