package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// 35h. A verified fix of a sentence many citations back keeps their tokens, as the accept requires,
// and the tokens are not words it adds: the authoring bound measures the words.
func TestAVerifiedFixOfAHeavilyBackedSentenceIsMeasuredByItsWords(t *testing.T) {
	const sentence = "Costs rose sharply in the first quarter."
	b := readingRunOf(t, rdReport, sentence, sentence, sentence)
	rep := readReport(t, b.runDir)
	start := strings.Index(rep, "Costs rose")
	end := strings.Index(rep[start:], ".") + start + 1
	marked := rep[start:end]
	if n := strings.Count(marked, "<!--cite:"); n != 3 {
		t.Fatalf("the fixture's sentence carries %d citations, want 3: %q", n, marked)
	}
	// 70 characters of words: inside the bound, and past it only with the tokens counted.
	newText := "Read again, with the measured figures for every month of the quarter, " + marked
	if _, err := b.verifyJSON(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source", "--new", newText)...); err != nil {
		t.Fatalf("a %d-character rewording of a sentence whose tokens it keeps was refused: %v",
			len([]rune(claimcount.StripAnchors(newText)))-len([]rune(sentence)), err)
	}
}
