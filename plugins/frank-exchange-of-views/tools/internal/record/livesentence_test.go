package record

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// ONE BARE RULE: an anchor claimcount.BareAnchorIDs lists stands bare for every reader of the live
// sentence, wherever its first occurrence sits. Here the first occurrence is in a code fence, outside
// the claim stream BareAnchorIDs reads, and the one in prose opens its sentence with no prose before
// it: bare, not absent and not live.
// A live anchor and one the report does not hold read as they always have.
func TestAnAnchorBareInProseIsBareWhereverItFirstOccurs(t *testing.T) {
	fenced, live, gone := "C-a1b2c3d4", "C-0badcafe", "C-99887766"
	report := "# Findings\n\n```\n" + anchor.Token(fenced) + "\n```\n\n" +
		"The grass is green" + anchor.Token(live) + ".\n\n" +
		anchor.Token(fenced) + " Next comes plain prose.\n"
	if !slices.Contains(claimcount.BareAnchorIDs(report), fenced) {
		t.Fatalf("the fixture does not stand %s bare by claimcount.BareAnchorIDs:\n%s", fenced, report)
	}
	sentences, bare := SentencesIn(report, []string{fenced, live, gone})
	if _, ok := sentences[fenced]; ok || !bare[fenced] {
		t.Errorf("%s, bare in prose with its first occurrence in a fence, reads live %v / bare %v, want bare", fenced, ok, bare[fenced])
	}
	if got := sentences[live]; got != "The grass is green." || bare[live] {
		t.Errorf("%s reads %q (bare %v), want its live sentence", live, got, bare[live])
	}
	if _, ok := sentences[gone]; ok || bare[gone] {
		t.Errorf("%s, not in the report, reads live %v / bare %v, want neither", gone, ok, bare[gone])
	}
}

// THE LANDING COMPARES WHATEVER THE CHECK JUDGED: the read in a landing is its own mode, so a judged
// sentence that is empty is compared like any other, never read as the read before a check.
func TestTheLandingComparesAnEmptyJudgedSentence(t *testing.T) {
	id := "C-0badcafe"
	report := "# Findings\n\nThe grass is green" + anchor.Token(id) + ".\n"
	render := func() (string, error) { return report, nil }
	notRetired := func() (AnchorRetire, bool, error) { return AnchorRetire{}, false, nil }
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
// ONLY in a fence, it is neither live nor bare (sentenceAbsent).
func TestAnAnchorInAFenceOrProselessSentenceReadsItsLiveSentenceInProse(t *testing.T) {
	fencedLive, proselessLive, onlyFenced := "C-11223344", "C-55667788", "C-99aabbcc"
	report := "# Findings\n\n```\n" + anchor.Token(fencedLive) + "\n" + anchor.Token(onlyFenced) + "\n```\n\n" +
		"One.\n\n" + anchor.Token(proselessLive) + "\n\n" +
		"The sky is blue" + anchor.Token(fencedLive) + ".\n\n" +
		"The grass is green" + anchor.Token(proselessLive) + ".\n"

	sentences, bare := SentencesIn(report, []string{fencedLive, proselessLive, onlyFenced})
	if got := sentences[fencedLive]; got != "The sky is blue." || bare[fencedLive] {
		t.Errorf("%s (in fence then prose) = %q (bare %v), want %q", fencedLive, got, bare[fencedLive], "The sky is blue.")
	}
	if got := sentences[proselessLive]; got != "The grass is green." || bare[proselessLive] {
		t.Errorf("%s (in proseless then prose) = %q (bare %v), want %q", proselessLive, got, bare[proselessLive], "The grass is green.")
	}
	if got, ok := sentences[onlyFenced]; ok || bare[onlyFenced] {
		t.Errorf("%s (only in fence) = %q (bare %v), want absent (not in map)", onlyFenced, got, bare[onlyFenced])
	}
}
