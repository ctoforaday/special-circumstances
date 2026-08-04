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

// AMBIGUITY MUST BE REFUSED, NOT GUESSED. LocateSpan takes the FIRST match and says nothing, so a
// quote appearing twice silently edits whichever came first. blue is explicitly instructed to
// propagate a correction to every site stating a claim, so repeated text is the expected shape of
// a real report — this is the case where guessing is worst.
func TestPlanEditRefusesAnAmbiguousSpan(t *testing.T) {
	const rep = "# H\n\nThe value is stable.\n\nElsewhere, again: The value is stable.\n"
	if _, err := planEdit(rep, "The value is stable.", "The value is steady."); err == nil {
		t.Fatal("an --old span matching twice was accepted — it silently edits the first site")
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
