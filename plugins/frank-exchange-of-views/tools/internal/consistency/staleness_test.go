package consistency

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A STANDING CHECK THE RULE CANNOT HOLD TO A SENTENCE IS COUNTED UNMEASURED, never a violation and never
// agreement: one of an anchor standing bare (the sentence it backed was cut, and no edit names a bare
// anchor reopened), and one of an anchor whose kind backs nothing, which has no evidence row to read
// it by. Beside them, the rule's two findings still fire: a check of another wording, and one of an
// anchor the report does not hold.
func TestAStandingCheckWithNoSentenceToHoldItToIsUnmeasured(t *testing.T) {
	const (
		live   = "K-0000000a"
		bare   = "K-0000000b"
		absent = "K-0000000c"
		gap    = "G-00000001"
	)
	report := "# Findings\n\nAlpha holds<!--computation:" + live + "-->. <!--computation:" + bare + "-->\n"
	if !slices.Contains(claimcount.BareAnchorIDs(report), bare) {
		t.Fatalf("the fixture's %s does not stand bare in %q", bare, report)
	}
	standing := []record.StandingCheck{
		{Anchor: live, Event: 10, Seat: "red-lens-computation", Holds: true, Sentence: "Alpha once held."},
		{Anchor: bare, Event: 11, Seat: "red-lens-computation", Holds: true, Sentence: "Beta held."},
		{Anchor: absent, Event: 12, Seat: "red-lens-computation", Holds: true, Sentence: "Gamma held."},
		{Anchor: gap, Event: 13, Seat: "red-lens-logic", Holds: true, Sentence: "Delta held."},
	}
	var violations []string
	unmeasured := stalenessIn(standing, report, func(v Violation) { violations = append(violations, v.Anchor) })
	if !slices.Equal(violations, []string{live, absent}) {
		t.Errorf("staleness violations on %v, want [%s %s]", violations, live, absent)
	}
	for _, id := range []string{bare, gap} {
		if !slices.ContainsFunc(unmeasured, func(u string) bool { return strings.Contains(u, id) }) {
			t.Errorf("the standing check of %s is not counted unmeasured: %q", id, unmeasured)
		}
	}
	if len(unmeasured) != 2 {
		t.Errorf("%d unmeasured, want 2: %q", len(unmeasured), unmeasured)
	}
}
