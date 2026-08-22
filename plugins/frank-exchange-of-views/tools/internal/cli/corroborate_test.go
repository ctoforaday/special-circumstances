package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// corroborateRun is a run with a registered lens and a seeded report.
func corroborateRun(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	seedBlueReport(t, runDir)
	return runDir
}

const corroborated = "§2 the finding prose lands in a quoted sentence."

// assembled runs the composer and returns the report's TEXT. Assemble writes the file and
// returns its PATH, which reads like markdown to a Contains check and always fails to match.
func assembled(t *testing.T, runDir string) string {
	t.Helper()
	path, err := report.Assemble(runDir)
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(md)
}

// A SUPPORTING CORROBORATION BECOMES AN ORDINARY FOOTNOTE.
//
// A human reader cares that the text has appropriate references, not which team inserted them.
// Before this, red's independent corroboration reached no reader of the document at all:
// `internal/report` has no reader for a Verify event, and citationid.go stated red's exclusion
// as a property ("Red's `lens cite` carries no label and is EXCLUDED").
func TestASupportingCorroborationRendersAsAFootnote(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--url", "https://example.org/red-found-this", "--title", "A source red found",
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the source says exactly this at page 4"); err != nil {
		t.Fatalf("a supporting corroboration was refused: %v", err)
	}

	// The anchor is spliced into the report, invisibly, exactly as blue's cite and red's
	// finding markers already are.
	md, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), "<!--cite:c-") {
		t.Fatalf("no citation anchor was spliced at the corroborated sentence:\n%s", md)
	}

	// And it reaches the READER, which is the whole point: assembly weaves it into the
	// bibliography with no knowledge of which seat wrote it.
	out := assembled(t, runDir)
	if !strings.Contains(out, "## Bibliography") || !strings.Contains(out, "example.org/red-found-this") {
		t.Errorf("red's corroborating source is absent from the assembled bibliography:\n%s", out)
	}
	if !strings.Contains(out, "A source red found") {
		t.Error("the source's title did not travel into the footnote")
	}
}

// A REFUTING CORROBORATION IS NOT A REFERENCE, and must not become one.
//
// A source that CONTRADICTS the sentence, rendered in the bibliography, reads as backing it —
// and the report's own assembly check already treats a live refuted citation as a failure.
func TestARefutingCorroborationIsNotSplicedAsAFootnote(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--url", "https://example.org/contradicts", "--title", "A source that disagrees",
		"--quote", corroborated, "--as", "refutes", "--confidence", "high",
		"--reason", "the source says the opposite at page 9"); err != nil {
		t.Fatalf("a refuting corroboration was refused — it is red's strongest finding on this axis and must still record: %v", err)
	}
	md, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(md), "<!--cite:c-") {
		t.Errorf("a REFUTING source was spliced as a citation — the reader would meet it as a reference backing the sentence it contradicts:\n%s", md)
	}
	if strings.Contains(assembled(t, runDir), "example.org/contradicts") {
		t.Error("a refuting source reached the bibliography")
	}
}

// THE QUOTE MUST BE IN THE LIVE REPORT, so a corroboration of a claim blue has since edited
// away is refused rather than spliced blind — the same rule blue's own cite is held to.
func TestACorroborationOfAnAbsentClaimIsRefused(t *testing.T) {
	runDir := corroborateRun(t)
	_, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--url", "https://example.org/x", "--title", "T",
		"--quote", "a sentence that is nowhere in the report", "--as", "supports",
		"--confidence", "high", "--reason", "r")
	if err == nil {
		t.Fatal("a corroboration of a claim absent from the report was accepted")
	}
	if !strings.Contains(err.Error(), "--quote") {
		t.Errorf("the refusal does not name the flag to fix, so a seat cannot act on it: %v", err)
	}
	// AND NOTHING WAS RECORDED. A rejected splice must leave no event behind, or the record
	// carries a corroboration whose footnote does not exist.
	srcs, err := record.CitedSources(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 0 {
		t.Errorf("the splice was refused and %d source(s) still recorded: %+v", len(srcs), srcs)
	}
}
