package fuzz

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
)

// THE SPAN BLUE'S FIX QUOTES IS IN THE SENTENCE THE ANCHOR BACKS. Lane 3's part holds two sentences on
// one line (lanePartText), the second cited and the first not: words reaching back to the line's start
// would put the fix's prefix on the first sentence, leave the cited one as it stood, and the edit
// would rightly not list the citation reopened (fixFoundAgainst's oracle, release fuzz seed 14).
func TestWordsBeforeStaysInTheSentenceItsAnchorBacks(t *testing.T) {
	const cite = "C-f9c45cda"
	line := "Across the sampled corpus the rise holds in every year measured. " +
		strings.TrimSuffix(repeatedSentenceSeeds[1], ".") + anchor.Token(cite) + ".\n"
	cur := "# § fuzz\n\n" + line
	words := wordsBefore(cur, cite)
	if want := strings.TrimSuffix(repeatedSentenceSeeds[1], "."); words != want {
		t.Fatalf("wordsBefore = %q, want %q, the cited sentence's words alone", words, want)
	}
	after := strings.Replace(cur, words, "Read against its source, "+words, 1)
	if got := bluedoc.ReopenedAnchors(cur, after); !slices.Contains(got, cite) {
		t.Errorf("the edit of %q lists %v reopened, not %s", words, got, cite)
	}
	// THE SENTENCE HELD TWICE has no span standing once before its anchor, so it is "".
	twice := cur + "\n" + repeatedSentenceSeeds[1] + "\n"
	if got := wordsBefore(twice, cite); got != "" {
		t.Errorf("wordsBefore over a sentence the report holds twice = %q, want \"\"", got)
	}
}
