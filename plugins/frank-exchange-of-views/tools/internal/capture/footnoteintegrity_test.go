package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeReport(t *testing.T, body string) string {
	t.Helper()
	run := t.TempDir()
	if err := os.WriteFile(filepath.Join(run, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return run
}

// THE SHAPE THAT SHIPPED TWICE. Assembly wrote the proof appendix headings as `### [^P1] …`,
// which markdown reads as a second REFERENCE, and defined neither — so `[^P1]` rendered as
// literal text at every site in both 2026-08-23 reports.
func TestFootnoteIntegrityCatchesTheShippedDanglingProofFootnotes(t *testing.T) {
	run := writeReport(t, `# A report

None of the plan's artifacts exist[^P1]: `+"`skills/x`"+`, and the wait halves[^P4].

[^1]: Settles, Active Learning Literature Survey (2009). http://example.invalid (accessed 2026-08-23)

## Proofs

### [^P1] the build-state script

- **basis**: reproducible
`)
	a := FootnoteIntegrity(runtest.Open(t, run))
	if a.Verdict != "FAIL" {
		t.Fatalf("verdict = %s, want FAIL — a reader sees literal [^P1] here\n%s", a.Verdict, a.Detail)
	}
	for _, want := range []string{"[^P1]", "[^P4]"} {
		if !strings.Contains(a.Detail, want) {
			t.Errorf("detail must name %s; got:\n%s", want, a.Detail)
		}
	}
	if strings.Contains(a.Detail, "[^1]") {
		t.Errorf("[^1] IS defined and must not be reported: %s", a.Detail)
	}
}

// A REFERENCE MAY BE FOLLOWED BY A COLON IN PROSE, and that is not a definition. This is the
// case that makes a naive `\[\^X\]:` check wrong in the permissive direction: it would read
// "exist[^P1]:" as defining P1 and pass a report that is broken.
func TestFootnoteIntegrityDoesNotMistakeAMidLineColonForADefinition(t *testing.T) {
	run := writeReport(t, "Artifacts exist[^P1]: none of them.\n")
	if a := FootnoteIntegrity(runtest.Open(t, run)); a.Verdict != "FAIL" {
		t.Fatalf("a mid-line `[^P1]:` is a reference, not a definition; verdict = %s (%s)", a.Verdict, a.Detail)
	}
}

func TestFootnoteIntegrityPassesAWellFormedReport(t *testing.T) {
	run := writeReport(t, `Backed by a source[^1] and a computation[^P1].

[^1]: Someone, A Paper (2024). http://example.invalid (accessed 2026-08-23)
[^P1]: Proof `+"`p-abd56845`"+` — `+"`blue/candidates/lane3_buildstate.sh`"+`, exit 0.
`)
	a := FootnoteIntegrity(runtest.Open(t, run))
	if a.Verdict != "PASS" {
		t.Fatalf("verdict = %s, want PASS: %s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "2 footnote(s) referenced") {
		t.Errorf("detail should count what it checked; got %s", a.Detail)
	}
}

// CODE IS NOT PROSE. A report that quotes a proof's output in a fenced block, or writes the
// footnote surface inline while describing it, references nothing — and the report that first
// carried this audit's own repair note does exactly that, so scanning code would fail the very
// document that fixed the defect.
func TestFootnoteIntegrityIgnoresFootnotesInsideCode(t *testing.T) {
	run := writeReport(t, "This report referenced its proof footnotes (`[^P…]`) without defining them.\n\n"+
		"```\n[^P9]: quoted output that is not this document's footnote\n```\n\n"+
		"A real one[^1].\n\n[^1]: Someone, A Paper (2024). http://example.invalid\n")
	a := FootnoteIntegrity(runtest.Open(t, run))
	if a.Verdict != "PASS" {
		t.Fatalf("footnote-shaped text inside code is not a reference; verdict = %s (%s)", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "1 footnote(s) referenced") {
		t.Errorf("only the prose reference counts; got %s", a.Detail)
	}
}

// Before assembly there is nothing to judge, and saying so beats a PASS that means "the file I
// was looking for was not there" — the distinction AssemblyScreen already draws.
func TestFootnoteIntegritySkipsBeforeAssembly(t *testing.T) {
	if a := FootnoteIntegrity(runtest.Open(t, t.TempDir())); a.Verdict != "SKIP" {
		t.Fatalf("verdict = %s, want SKIP (%s)", a.Verdict, a.Detail)
	}
}
