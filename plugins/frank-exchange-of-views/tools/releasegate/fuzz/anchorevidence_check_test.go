package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE EVIDENCE HELPER IS ITSELF EXERCISED, because a diagnostic that renders nothing is worse
// than none: it appends an empty line to a failure and reads like a run that carried no anchors.
func TestAnchorEvidenceNamesTheMarkersTheReportCarries(t *testing.T) {
	runDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# § fuzz\n\nA sentence.<!--fx:f-00abc123--> Another.<!--cite:c-00def456-->\n"
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := anchorEvidence(runDir)
	for _, want := range []string{"f-00abc123", "c-00def456", "2 protected anchor(s)"} {
		if !strings.Contains(got, want) {
			t.Errorf("anchor evidence does not carry %q — a diagnostic that names nothing cannot diagnose:\n%s", want, got)
		}
	}
}

// AND IT SAYS SO WHEN THERE IS NOTHING TO READ, rather than returning an empty string that
// would append silence to the failure it is meant to explain.
func TestAnchorEvidenceIsLoudWhenTheReportIsUnreadable(t *testing.T) {
	got := anchorEvidence(t.TempDir()) // no blue/report.md at all
	if !strings.Contains(got, "unavailable") {
		t.Errorf("anchor evidence should name its own failure, got %q", got)
	}
}
