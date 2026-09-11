package cli

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

var cardHeading = regexp.MustCompile(`(?m)^# (\S+) scorecard$`)

// THE OPERATOR'S SCORECARD PRINTS EVERY CARD UNLESS TOLD OTHERWISE. An operator is not a party and
// has no card, so the question it asks is how the whole run is going; making it name one card
// before it could see any was a selector nothing could default.
func TestScorecardPrintsEveryCardInOrder(t *testing.T) {
	runDir := newRun(t)
	out, err := run(t, "scorecard", "--run", runDir, "--seat-id", "operator")
	if err != nil {
		t.Fatalf("scorecard with no --card refused: %v", err)
	}
	var got []string
	for _, m := range cardHeading.FindAllStringSubmatch(out, -1) {
		got = append(got, m[1])
	}
	if strings.Join(got, ",") != strings.Join(scorecard.Cards, ",") {
		t.Errorf("card headings = %v, want %v in that order:\n%s", got, scorecard.Cards, out)
	}
}

func TestScorecardCardNarrowsToOne(t *testing.T) {
	runDir := newRun(t)
	out, err := run(t, "scorecard", "--run", runDir, "--seat-id", "operator", "--card", "bench")
	if err != nil {
		t.Fatalf("scorecard --card bench refused: %v", err)
	}
	if m := cardHeading.FindAllStringSubmatch(out, -1); len(m) != 1 || m[0][1] != "bench" {
		t.Errorf("--card bench printed %v, want only the bench card:\n%s", m, out)
	}
}

func TestScorecardRefusesAnUnknownCardByName(t *testing.T) {
	runDir := newRun(t)
	_, err := run(t, "scorecard", "--run", runDir, "--seat-id", "operator", "--card", "chair")
	if err == nil {
		t.Fatal("--card chair was accepted; the cards are red, blue and bench")
	}
	for _, want := range []string{`"chair" is not a scorecard`, "red, blue and bench", "omit --card"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
}
