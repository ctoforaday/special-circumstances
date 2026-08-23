package bluedoc

import "testing"

// The boundary rule: legal spans sit at whitespace, the string edge, or an ANNOTATION edge.
// The anchor cases matter most — red anchors exactly the sentences blue is asked to repair, and
// anchors splice flush against the last content char, so a naive whitespace-only rule would
// reject the common case.
func TestSpanBoundaryOK(t *testing.T) {
	const anchored = "The sieve is fast<!--fx:f-abc--> and simple."
	for _, c := range []struct {
		name string
		s    string
		a, b int
		want bool
	}{
		{"whole string", "abc", 0, 3, true},
		{"word at spaces", "one two three", 4, 7, true},
		{"splits a word at the start", "independently", 3, 13, false},
		{"splits a word at the end", "independently", 0, 11, false},
		// The matcher trims trailing punctuation, so a matched span ALWAYS stops before the
		// period. That must stay legal or every sentence-final edit is refused.
		{"stops before a period", "It is prime. Next.", 0, 11, true},
		{"span ending at an anchor", anchored, 0, 17, true},

		// THE CHARACTER CLASS ITSELF, at the edges nothing else reaches.
		//
		// Every case above uses mid-range lowercase letters, and a mutation sweep showed what that
		// costs: `||`→`&&` in word() survived, because Go binds `&&` tighter, leaving the letter
		// clauses intact while digits and `_` stopped counting as word characters. Seven mutants
		// of one line survived for the same reason — the endpoints `0 9 a z A Z` and `_` were
		// never exercised, so the class could be wrong at its edges and every test still passed.
		//
		// It is not a hypothetical shape here: gap ids (`R1-1`), years and schema field names
		// (`closure_class`) are exactly the tokens a span lands in the middle of.
		{"splits a number", "value 2026 here", 6, 9, false},
		{"splits at the digit 0", "value 20 here", 6, 7, false},
		{"splits at the digit 9", "value 99 here", 6, 7, false},
		{"splits an underscored identifier", "the closure_class field", 4, 11, false},
		{"splits at an underscore", "a_b", 0, 1, false},
		{"splits at lowercase z", "zz here", 0, 1, false},
		{"splits at lowercase a", "aa here", 0, 1, false},
		{"splits at uppercase A", "AA here", 0, 1, false},
		{"splits at uppercase Z", "ZZ here", 0, 1, false},
		// And a digit-to-space boundary is still LEGAL — the rule rejects splitting a word, not
		// touching one.
		{"whole number at spaces", "value 2026 here", 6, 10, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := spanBoundaryOK(c.s, c.a, c.b); got != c.want {
				t.Errorf("spanBoundaryOK(%q, %d, %d) = %v, want %v — segment %q", c.s, c.a, c.b, got, c.want, c.s[c.a:c.b])
			}
		})
	}
}
