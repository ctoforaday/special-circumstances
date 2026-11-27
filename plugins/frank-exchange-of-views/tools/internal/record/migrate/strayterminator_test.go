package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// TestAStrayTerminatorReadsTheAnchorRunItStandsAfter is m26's edit 222 in small: a quote whose
// leading full stop locates on the sentence's own stop, after the anchor run the locate skips. The
// edit is settled because the span, widened leftward over that run, holds an anchor the edit does
// not name; read without the widening it holds none, and replaying the edit deletes the stop.
func TestAStrayTerminatorReadsTheAnchorRunItStandsAfter(t *testing.T) {
	run := anchor.Token("G-59e61f8c") + anchor.Token("C-8a9958e3")
	text := "Some claim stands here" + run + ".\n\n## Open questions\n"
	for _, c := range []struct {
		old, replacement string
		stray, located   bool
	}{
		{".\n\n## Open questions", "\n\n## Open questions", true, true},
		{"## Open questions", "## Questions", false, true},
		{". Nowhere in the text", "Nowhere in the text", false, false},
	} {
		stray, located := strayAt(text, c.old, c.replacement)
		if stray != c.stray || located != c.located {
			t.Errorf("strayAt(%q -> %q) = %v, %v; want %v, %v", c.old, c.replacement, stray, located, c.stray, c.located)
		}
	}
	if got, want := tokenRunBefore(text, len("Some claim stands here")+len(run)), len("Some claim stands here"); got != want {
		t.Errorf("tokenRunBefore = %d, want %d: the run of two abutting tokens", got, want)
	}
}
