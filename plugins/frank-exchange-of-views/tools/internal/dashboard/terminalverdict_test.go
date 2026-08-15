package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE VERDICT COMES OFF THE RECORD, NOT OUT OF THE PROSE IT WAS RENDERED INTO.
//
// The regex was the only reader, and it is coupled to a sentence assemble.go owns: basisNote
// appends " — **derived from the record**…" directly after the verdict word, so the pattern has to
// stop at an em-dash. A rewording there would blank the dashboard, and a blank dashboard is what an
// un-assembled run looks like too.
//
// The disagreement case is the one worth pinning: when the record says one thing and the rendered
// prose says another, the record wins. Anything else makes the dashboard a reader of a reader.
func TestTerminalVerdictPrefersTheRecordOverTheRenderedProse(t *testing.T) {
	runDir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if _, _, err := record.RegisterSeat(runDir, "judge-terminal"); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(runDir, "judge-terminal", "outcome",
		record.NewPayload().Set("verdict", "HALTED").Set("prose", "ended on safety grounds")); err != nil {
		t.Fatal(err)
	}
	// The rendered artifact says something else. It is the derived carrier; the event is the fact.
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Verdict:** VERIFIED — **derived from the record**, not claimed.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readTerminalVerdict(runDir); got != "HALTED" {
		t.Errorf("readTerminalVerdict = %q, want \"HALTED\" — the record holds the verdict as a field and the report is a rendering of it", got)
	}
}

// AND THE FALLBACK STILL WORKS, so a run assembled before the terminal act was an event is not
// silently blanked by the fix.
func TestTerminalVerdictFallsBackToTheReportWhenTheRecordCannotSay(t *testing.T) {
	runDir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Verdict:** CEILING-TERMINATED — the run hit its round ceiling.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readTerminalVerdict(runDir); got != "CEILING-TERMINATED" {
		t.Errorf("readTerminalVerdict = %q, want \"CEILING-TERMINATED\" from the report fallback", got)
	}
}
