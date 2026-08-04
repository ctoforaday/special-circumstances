package blue

import "testing"

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
