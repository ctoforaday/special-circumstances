package capture

import (
	"fmt"
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

// THE SHAPE THAT SHIPPED TWICE. Assembly wrote the computation appendix headings as `### [^K1] …`,
// which markdown reads as a second REFERENCE, and defined neither — so `[^K1]` rendered as
// literal text at every site in both 2026-08-23 reports.
func TestFootnoteIntegrityCatchesTheShippedDanglingComputationFootnotes(t *testing.T) {
	run := writeReport(t, `# A report

None of the plan's artifacts exist[^K1]: `+"`skills/x`"+`, and the wait halves[^K4].

[^1]: Settles, Active Learning Literature Survey (2009). http://example.invalid (accessed 2026-08-23)

## Computations

### [^K1] the build-state script

- **basis**: reproducible
`)
	a := FootnoteIntegrity(runtest.Open(t, run))
	if a.Verdict != "FAIL" {
		t.Fatalf("verdict = %s, want FAIL — a reader sees literal [^K1] here\n%s", a.Verdict, a.Detail)
	}
	for _, want := range []string{"[^K1]", "[^K4]"} {
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
// "exist[^K1]:" as defining P1 and pass a report that is broken.
func TestFootnoteIntegrityDoesNotMistakeAMidLineColonForADefinition(t *testing.T) {
	run := writeReport(t, "Artifacts exist[^K1]: none of them.\n")
	if a := FootnoteIntegrity(runtest.Open(t, run)); a.Verdict != "FAIL" {
		t.Fatalf("a mid-line `[^K1]:` is a reference, not a definition; verdict = %s (%s)", a.Verdict, a.Detail)
	}
}

func TestFootnoteIntegrityPassesAWellFormedReport(t *testing.T) {
	run := writeReport(t, `Backed by a source[^1] and a computation[^K1].

[^1]: Someone, A Paper (2024). http://example.invalid (accessed 2026-08-23)
[^K1]: Computation `+"`K-abd56845`"+` — `+"`blue/candidates/lane3_buildstate.sh`"+`, exit 0.
`)
	a := FootnoteIntegrity(runtest.Open(t, run))
	if a.Verdict != "PASS" {
		t.Fatalf("verdict = %s, want PASS: %s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "2 footnote(s) referenced") {
		t.Errorf("detail should count what it checked; got %s", a.Detail)
	}
}

// CODE IS NOT PROSE. A report that quotes a computation's output in a fenced block, or writes the
// footnote surface inline while describing it, references nothing — and the report that first
// carried this audit's own repair note does exactly that, so scanning code would fail the very
// document that fixed the defect.
func TestFootnoteIntegrityIgnoresFootnotesInsideCode(t *testing.T) {
	run := writeReport(t, "This report referenced its computation footnotes (`[^P…]`) without defining them.\n\n"+
		"```\n[^K9]: quoted output that is not this document's footnote\n```\n\n"+
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

// A SEAT'S WORDS ARE NOT A REFERENCE ONCE ASSEMBLY HAS ESCAPED THEM, and they are one where it has
// not. The two lines are the wave B smoke's: a finding in the board document naming the report's
// footnote, which report.md defines and docket.md does not, and a log entry in the run document
// quoting a regex. Printed raw, each is a reference with no definition in its document — the board
// document's became a citation link to nowhere in report.html — and the audit says so.
func TestFootnoteIntegrityReadsAnEscapedNoteAsText(t *testing.T) {
	const (
		finding = "K-c53c6b2c, the other anchor on the same sentence and the one footnote %s cites, still types the constants."
		logged  = "- **red-lens-evidence** (friction): the form grep -o '%s{0,120}term%s{0,120}', fails with a complexity-limit error."
	)
	set := func(open string) string {
		run := writeReport(t, "# A report\n\nThe norms multiply to 8281[^K5].\n\n[^K5]: The norms of 7, 3 + 2i and 3 − 2i multiply to 8281.\n")
		for name, body := range map[string]string{
			"docket.md": "# the board\n\n### F-76352127 (dark-side)\n" + fmt.Sprintf(finding, open+"K5]") + "\n",
			"run.md":    "# the run\n\n" + fmt.Sprintf(logged, open+"<>]", open+"<>]") + "\n",
		} {
			if err := os.WriteFile(filepath.Join(run, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return run
	}
	escaped := FootnoteIntegrity(runtest.Open(t, set(`\[^`)))
	if escaped.Verdict != "PASS" || !strings.Contains(escaped.Detail, "1 footnote(s) referenced, 1 defined") {
		t.Errorf("escaped seat text is counted as a reference: %s — %s", escaped.Verdict, escaped.Detail)
	}
	raw := FootnoteIntegrity(runtest.Open(t, set("[^")))
	if raw.Verdict != "FAIL" {
		t.Fatalf("seat text printed raw opens a footnote in each document and the audit passed it: %s", raw.Detail)
	}
	// EACH LABEL IS REPORTED AS CODE: the detail is printed in the run document, where a bare
	// label is the next audit's dangling reference.
	for _, want := range []string{"`docket.md:[^K5]`", "`run.md:[^<>]`"} {
		if !strings.Contains(raw.Detail, want) {
			t.Errorf("the detail does not name %s:\n%s", want, raw.Detail)
		}
	}
}
