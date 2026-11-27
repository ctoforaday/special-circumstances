package consistency

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A STANDING CHECK THE RULE CANNOT HOLD TO A SENTENCE IS COUNTED UNMEASURED, never a violation and never
// agreement: one of an anchor whose kind backs nothing, which has no evidence row to read it by. Beside
// it, the rule's two findings still fire: a check of another wording, and one of an anchor the report
// does not hold.
func TestAStandingCheckWithNoSentenceToHoldItToIsUnmeasured(t *testing.T) {
	const (
		live   = "K-0000000a"
		absent = "K-0000000c"
		gap    = "G-00000001"
	)
	report := "# Findings\n\nAlpha holds<!--computation:" + live + "-->.\n"
	standing := []record.StandingCheck{
		{Anchor: live, Event: 10, Seat: "red-lens-computation", Holds: true, Sentence: "Alpha once held."},
		{Anchor: absent, Event: 12, Seat: "red-lens-computation", Holds: true, Sentence: "Gamma held."},
		{Anchor: gap, Event: 13, Seat: "red-lens-logic", Holds: true, Sentence: "Delta held."},
	}
	var violations []string
	unmeasured := stalenessIn(standing, report, func(v Violation) { violations = append(violations, v.Anchor) })
	if !slices.Equal(violations, []string{live, absent}) {
		t.Errorf("staleness violations on %v, want [%s %s]", violations, live, absent)
	}
	if len(unmeasured) != 1 || !strings.Contains(unmeasured[0], gap) {
		t.Errorf("unmeasured = %q, want the standing check of %s alone", unmeasured, gap)
	}
}
