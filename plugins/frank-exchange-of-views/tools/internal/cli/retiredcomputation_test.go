package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A RETIRED COMPUTATION IS EVIDENCE FOR NOTHING, AND THE SET SHOWS IT NOWHERE BUT ITS RETIRE. The
// #1663 smoke shipped a computation whose claim blue had retired under its stale reason ("gives
// eight") in the Computations section, beside a report that says seven (G-7dcf6270). The standing
// computation is the one shown, and it is K1: the numbering is over the computations the set anchors.
func TestARetiredComputationIsNotShownAsEvidence(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Plain prose here.\n")
	registerBlue(t, runDir)
	retired := computeSentence(t, runDir, "Alpha holds.")
	standing := computeSentence(t, runDir, "Beta holds.")
	gutToAnchor(t, runDir, "Alpha holds.", anchor.Token(retired))
	retireClaim(t, runDir, "Alpha holds.")

	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(a.Report)
	evidence, err := os.ReadFile(filepath.Join(dir, report.FileEvidence))
	if err != nil {
		t.Fatalf("evidence.md: %v", err)
	}
	if strings.Contains(string(evidence), retired) || strings.Contains(string(evidence), "anchored to nothing") {
		t.Errorf("the retired computation %s is shown as evidence:\n%s", retired, evidence)
	}
	if !strings.Contains(string(evidence), standing) || !strings.Contains(string(evidence), "### K1 —") || strings.Contains(string(evidence), "### K2 —") {
		t.Errorf("the standing computation %s is not the set's one computation, K1:\n%s", standing, evidence)
	}
	changelog, err := os.ReadFile(filepath.Join(dir, report.FileChangelog))
	if err != nil || !strings.Contains(string(changelog), "Alpha holds") {
		t.Errorf("the retire of the computed claim is not in the changelog (%v):\n%s", err, changelog)
	}
}
