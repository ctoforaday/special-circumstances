package cli

import (
	"strings"
	"testing"
)

// A READ AT ONE ANCHOR TAKES NO SELECTOR, on either view that reads at one. --quote
// filters a projection's rows; a read at an anchor has none, so the pair accepted would return the
// anchor's read and the seat would take it for the selection it asked for.
func TestAReadAtAnAnchorRefusesASelector(t *testing.T) {
	runDir := acceptRun(t)
	cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
	for _, view := range []string{"report", "evidence"} {
		if out, err := run(t, "show", "--run", runDir, "--seat-id", blueSeat, view, "--anchor", cite); err != nil || !strings.Contains(out, cite) {
			t.Fatalf("show %s --anchor %s = %v, want the read at the anchor:\n%s", view, cite, err, out)
		}
		args := []string{"show", "--run", runDir, "--seat-id", blueSeat, view, "--anchor", cite, "--quote", "Costs"}
		_, err := run(t, args...)
		if err == nil || !strings.Contains(err.Error(), "--anchor reads") || !strings.Contains(err.Error(), "pass one or the other") {
			t.Errorf("show %s --anchor with --quote = %v, want the pair refused", view, err)
		}
	}
	// Report also takes --search, which is similarly refused beside --anchor.
	args := []string{"show", "--run", runDir, "--seat-id", blueSeat, "report", "--anchor", cite, "--search", "Costs"}
	_, err := run(t, args...)
	if err == nil || !strings.Contains(err.Error(), "--anchor reads") || !strings.Contains(err.Error(), "pass one or the other") {
		t.Errorf("show report --anchor with --search = %v, want the pair refused", err)
	}
}
