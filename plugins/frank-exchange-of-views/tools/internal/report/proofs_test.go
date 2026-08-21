package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// seedProof lays down the on-disk artifact a proof event points at.
func seedProof(t *testing.T, runDir, sha, script, output string) {
	t.Helper()
	dir := filepath.Join(runDir, "proofs", sha)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "script.js"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "output"), []byte(output), 0o644); err != nil {
		t.Fatal(err)
	}
}

// THE DEFECT THIS PASS EXISTS FOR. Before it, a proof ran, cached, anchored and was
// auditable by red — and the assembled report carried the RAW anchor while mentioning the
// computation zero times. The evidence existed everywhere except the document a human reads.
func TestProofsAreWovenIntoTheDeliverableWithSourceAndOutput(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "console.log('divisors of 9:', 3);", "divisors of 9: 3\n")

	md := "Nine is composite by trial division<!--proof:p-deadbeef-->.\n"
	out := weaveProofs(runDir, md, []record.Proof{{
		Label: "p-deadbeef", SHA: sha, Basis: "reproducible",
		Script: "nine.js", Reason: "trial division settles it", Cites: "c-1234",
	}})

	if strings.Contains(out, "<!--proof:") {
		t.Error("a raw proof anchor shipped into the deliverable")
	}
	for _, want := range []string{
		"[^P1]",                    // the visible marker at the sentence
		"## Proofs",                // the section
		"console.log('divisors",    // THE SCRIPT — a computation's source IS its evidence
		"divisors of 9: 3",         // and its output
		"basis**: reproducible",    // how far it can be trusted
		"nine.js",                  // which script
		"applies the method cited", // the cite/compute pair, legible as one argument
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the woven report is missing %q:\n%s", want, out)
		}
	}
}

// An `observed` proof must say so IN THE REPORT, with its drift — a reader must be able to
// tell a measurement of a moving system from a proof.
func TestAnObservedProofIsLabelledInTheReport(t *testing.T) {
	runDir := newRun(t)
	const sha = "feed0000"
	seedProof(t, runDir, sha, "console.log(Math.random());", "0.42\n")

	out := weaveProofs(runDir, "The sampled rate varies<!--proof:p-1-->.\n", []record.Proof{{
		Label: "p-1", SHA: sha, Basis: "observed", Script: "s.js",
		Drift: "output differs from byte 2 between runs", Reason: "live sample",
	}})
	if !strings.Contains(out, "observed") || !strings.Contains(out, "differs from byte 2") {
		t.Errorf("an observed proof was not distinguished from a reproducible one:\n%s", out)
	}
}

// A missing artifact is STATED. A proofs section quietly short of its proof is worse than
// one that admits the artifact is gone.
func TestAMissingArtifactIsStatedNotSkipped(t *testing.T) {
	runDir := newRun(t)
	out := weaveProofs(runDir, "A claim<!--proof:p-9-->.\n", []record.Proof{{
		Label: "p-9", SHA: "notonthisdisk", Basis: "reproducible", Script: "gone.js",
	}})
	if !strings.Contains(out, "missing from this run directory") {
		t.Errorf("a missing artifact was silently omitted, so the report shows a proof it cannot produce:\n%s", out)
	}
}

// An anchor with no event behind it becomes an explicit unresolved line — never a crash and
// never a silent drop. Same defence weaveCitations makes for a dangling cite.
func TestADanglingProofAnchorIsMarkedUnresolved(t *testing.T) {
	// A VALID hex id with no event behind it. (An id-shaped-but-not-hex token like
	// "p-nothing" is not an anchor at all — the regex correctly ignores it, which is a
	// different case and not the one under test.)
	out := weaveProofs(t.TempDir(), "A claim<!--proof:p-abcdef01-->.\n", nil)
	if !strings.Contains(out, "unresolved proof") {
		t.Errorf("a dangling proof anchor vanished silently:\n%s", out)
	}
}

// With no proofs the report is returned unchanged — no empty section.
func TestNoProofsLeavesTheReportAlone(t *testing.T) {
	const md = "A report with no computations in it.\n"
	if got := weaveProofs(t.TempDir(), md, nil); got != md {
		t.Errorf("an empty Proofs section was appended:\n%s", got)
	}
}

// One proof cited twice shares one number, like a citation reused.
func TestOneProofUsedTwiceSharesItsNumber(t *testing.T) {
	runDir := newRun(t)
	seedProof(t, runDir, "s1", "x", "y")
	out := weaveProofs(runDir, "First<!--proof:p-a-->. Second<!--proof:p-a-->.\n",
		[]record.Proof{{Label: "p-a", SHA: "s1", Basis: "reproducible", Script: "a.js"}})
	if strings.Count(out, "[^P1]") != 3 { // two references plus the section heading
		t.Errorf("a proof used twice did not share one number:\n%s", out)
	}
	if strings.Contains(out, "[^P2]") {
		t.Errorf("the same proof was numbered twice:\n%s", out)
	}
}
