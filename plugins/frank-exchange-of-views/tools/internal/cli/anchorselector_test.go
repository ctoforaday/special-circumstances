package cli

import (
	"strings"
	"testing"
)

// A READ AT ONE ANCHOR TAKES NO SELECTOR, on either view that reads at one. --match and --quote
// filter a projection's rows; a read at an anchor has none, so the pair accepted would return the
// anchor's read and the seat would take it for the selection it asked for.
func TestAReadAtAnAnchorRefusesASelector(t *testing.T) {
	runDir := acceptRun(t)
	cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
	for _, view := range []string{"report", "evidence"} {
		if out, err := run(t, "show", "--run", runDir, "--seat-id", blueSeat, view, "--anchor", cite); err != nil || !strings.Contains(out, cite) {
			t.Fatalf("show %s --anchor %s = %v, want the read at the anchor:\n%s", view, cite, err, out)
		}
		for _, sel := range [][]string{{"--match", "Costs"}, {"--quote", "Costs"}} {
			args := append([]string{"show", "--run", runDir, "--seat-id", blueSeat, view, "--anchor", cite}, sel...)
			_, err := run(t, args...)
			if err == nil || !strings.Contains(err.Error(), "show "+view+": --anchor") || !strings.Contains(err.Error(), "pass one or the other") {
				t.Errorf("show %s --anchor with %s = %v, want the pair refused", view, sel[0], err)
			}
		}
	}
}
