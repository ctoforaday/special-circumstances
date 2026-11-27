package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ANCHOR A QUOTE CARRIES IS INSIDE THE SPAN, AT EITHER END. Smoke m26: blue quoted a span opening
// with finding markers, as `show report` printed it, and copied them into the replacement. The span was
// located from the first character after the anchor layer, so the opening markers stood outside it and
// the edit was refused for introducing them (ev220, ev221); a quote of one marker and its full stop
// alone was refused as not found (m26 L169), though the edit help says a punctuation repair quotes the
// anchors as the report prints them.

// markedPair writes two finding markers on two sentences a blank line apart, each on its words, the
// second sentence before a heading, and returns the run and the two markers' tokens.
func markedPair(t *testing.T) (runDir, t1, t2 string) {
	t.Helper()
	runDir = newRun(t)
	writeReport(t, runDir, "# Findings\n\nLead stands. Alpha holds for every input.\n\nBeta holds too.\n\n## Open questions\n\nNone.\n")
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)
	find := func(key, quote string) string {
		t.Helper()
		if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
			"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatalf("finding %s: %v", key, err)
		}
		return anchor.Token(lastBody(t, runDir, &recordpb.Finding{}).GetId())
	}
	t1, t2 = find("F1", "Alpha holds for every input."), find("F2", "Beta holds too.")
	if rep := readReport(t, runDir); !strings.Contains(rep, "Alpha holds for every input"+t1+".\n\nBeta holds too"+t2+".\n\n## Open questions") {
		t.Fatalf("the setup did not place the markers:\n%s", rep)
	}
	return runDir, t1, t2
}

func TestAQuoteOpeningWithAnAnchorCarriesIt(t *testing.T) {
	cases := []struct{ name, quote, new string }{
		{"one marker", "%2.\n\n## Open questions", "%2\n\n## Open questions"},
		{"markers across a blank line", "%1.\n\nBeta holds too%2.\n\n## Open questions", "%1.\n\nBeta holds too%2\n\n## Open questions"},
		{"a marker and its full stop alone", "%2.", "%2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runDir, t1, t2 := markedPair(t)
			fill := strings.NewReplacer("%1", t1, "%2", t2).Replace
			if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
				"--quote", fill(c.quote), "--new", fill(c.new), "--reason", "drop the full stop"); err != nil {
				t.Fatalf("the repair was refused: %v", err)
			}
			// readReport IS the replay, so this is also what every later render reproduces.
			if rep := readReport(t, runDir); !strings.Contains(rep, "Lead stands. Alpha holds for every input"+t1+".\n\nBeta holds too"+t2+"\n\n## Open questions") {
				t.Errorf("the full stop is still there, or more moved than it:\n%s", rep)
			}
		})
	}
}

// The same rule at the end: an anchor the quote carries past a blank line is inside the span.
func TestAQuoteClosingOnAnAnchorPastABlankLineCarriesIt(t *testing.T) {
	runDir, t1, t2 := markedPair(t)
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Lead stands. Alpha holds for every input"+t1+".\n\nBeta holds too"+t2,
		"--new", "Lead stands firm. Alpha holds for every input"+t1+".\n\nBeta holds too"+t2, "--reason", "reword"); err != nil {
		t.Fatalf("the edit was refused: %v", err)
	}
	if rep := readReport(t, runDir); !strings.Contains(rep, "Lead stands firm. Alpha holds for every input"+t1+".\n\nBeta holds too"+t2+".\n\n## Open questions") {
		t.Errorf("the markers moved or doubled:\n%s", rep)
	}
}
