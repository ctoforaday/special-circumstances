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
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := spanBoundaryOK(c.s, c.a, c.b); got != c.want {
				t.Errorf("spanBoundaryOK(%q, %d, %d) = %v, want %v — segment %q", c.s, c.a, c.b, got, c.want, c.s[c.a:c.b])
			}
		})
	}
}
