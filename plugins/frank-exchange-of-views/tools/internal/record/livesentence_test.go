package record

import (
	"slices"
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
