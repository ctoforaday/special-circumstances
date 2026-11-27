package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RETIRE THAT NAMES NO ANCHOR IS REFUSED AS MISSING ONE. On universe-m29 a retire naming nothing
// (`--quote ""`) was answered as if it named something; a retire takes one anchor, so --anchor blank or
// absent is the missing-flag refusal, and nothing is recorded.
func TestRetireRefusesABlankAnchorAsMissing(t *testing.T) {
	for _, named := range [][]string{{"--anchor", ""}, {"--anchor", "  "}, nil} {
		runDir := newRun(t)
		writeReport(t, runDir, "# H\n\nFive independent approaches agree.\n")
		args := append([]string{"retire", "--run", runDir, "--seat-id", "blue-respond", "--reason", "overclaim"}, named...)
		_, err := run(t, args...)
		if err == nil || !strings.Contains(err.Error(), "requires --anchor") {
			t.Errorf("retire %q: want the missing-anchor refusal, got %v", named, err)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
			t.Errorf("retire %q recorded %d retire(s)", named, n)
		}
	}
}
