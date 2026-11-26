package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// THE LANDING COMPARES WHATEVER THE CHECK JUDGED: the read in a landing is its own mode, so a judged
// sentence that is empty is compared like any other, never read as the read before a check.
func TestTheLandingComparesAnEmptyJudgedSentence(t *testing.T) {
	id := "C-0badcafe"
	report := "# Findings\n\nThe grass is green" + anchor.Token(id) + ".\n"
	render := func() (string, error) { return report, nil }
	notRetired := func() (AnchorRemoval, bool, error) { return AnchorRemoval{}, false, nil }
	if _, err := liveSentence(id, render, notRetired, inLanding("")); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Errorf("a landing whose check judged an empty sentence landed on %q: %v", "The grass is green.", err)
	}
	if got, err := liveSentence(id, render, notRetired, inLanding("The grass is green.")); err != nil || got != "The grass is green." {
		t.Errorf("a landing on the sentence judged: %q, %v", got, err)
	}
}

// AN ANCHOR FIRST IN A FENCE OR PROSELESS SENTENCE READS ITS PROSE SENTENCE (#1780): when a token
// first appears inside a fenced block or proseless sentence and later backs prose, SentencesIn
// returns the prose sentence it backs, not the fence or an empty string. When an anchor appears
// ONLY in a fence, it is not live (sentenceAbsent).
func TestAnAnchorInAFenceOrProselessSentenceReadsItsLiveSentenceInProse(t *testing.T) {
	fencedLive, proselessLive, onlyFenced := "C-11223344", "C-55667788", "C-99aabbcc"
	report := "# Findings\n\n```\n" + anchor.Token(fencedLive) + "\n" + anchor.Token(onlyFenced) + "\n```\n\n" +
		"One.\n\n" + anchor.Token(proselessLive) + "\n\n" +
		"The sky is blue" + anchor.Token(fencedLive) + ".\n\n" +
		"The grass is green" + anchor.Token(proselessLive) + ".\n"

	sentences := SentencesIn(report, []string{fencedLive, proselessLive, onlyFenced})
	if got := sentences[fencedLive]; got != "The sky is blue." {
		t.Errorf("%s (in fence then prose) = %q, want %q", fencedLive, got, "The sky is blue.")
	}
	if got := sentences[proselessLive]; got != "The grass is green." {
		t.Errorf("%s (in proseless then prose) = %q, want %q", proselessLive, got, "The grass is green.")
	}
	if got, ok := sentences[onlyFenced]; ok {
		t.Errorf("%s (only in fence) = %q, want absent (not in map)", onlyFenced, got)
	}
}
