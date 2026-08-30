package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
//
// SPLIT IN TWO SINCE, and the split is the #590 repair: the REFERENCE and its DEFINITION are
// woven into each document that anchors a proof, and the computation itself — script, output,
// sha256 — lands once, in evidence.md.
func TestProofsAreWovenIntoTheDeliverableWithSourceAndOutput(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "console.log('divisors of 9:', 3);", "divisors of 9: 3\n")

	proofs := []record.Proof{{
		Label: "p-deadbeef", SHA: sha, Basis: "reproducible",
		Script: "nine.js", Reason: "trial division settles it", Cites: "c-1234",
	}}
	md := "Nine is composite by trial division<!--proof:p-deadbeef-->.\n"
	out, used := weaveProofRefs(md, proofs)

	if strings.Contains(out, "<!--proof:") {
		t.Error("a raw proof anchor shipped into the deliverable")
	}
	if len(used) != 1 || used[0] != "p-deadbeef" {
		t.Errorf("the weave did not report which proofs the document anchors: %v", used)
	}
	for _, want := range []string{
		"[^P1]",                  // the visible marker at the sentence
		"[^P1]: trial division",  // AND ITS DEFINITION — the half #590 shipped without
		"evidence.md#p-deadbeef", // pointing at the computation in full
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the woven document is missing %q:\n%s", want, out)
		}
	}

	ev := evidenceDoc(runDir, proofs, map[string]bool{"p-deadbeef": true})
	for _, want := range []string{
		"## Proofs",                // the section
		"<a id=\"p-deadbeef\">",    // the anchor the footnote links to
		"### P1 — trial division",  // a HEADING, not a second dangling reference (#590)
		"console.log('divisors",    // THE SCRIPT — a computation's source IS its evidence
		"divisors of 9: 3",         // and its output
		"basis**: reproducible",    // how far it can be trusted
		"nine.js",                  // which script
		"applies the method cited", // the cite/compute pair, legible as one argument
	} {
		if !strings.Contains(ev, want) {
			t.Errorf("evidence.md is missing %q:\n%s", want, ev)
		}
	}
	// The heading must not itself be a footnote reference — the exact shape of #590.
	if strings.Contains(ev, "### [^P") {
		t.Errorf("a proof heading was written as a footnote reference (#590):\n%s", ev)
	}
}

// EVERY REFERENCE THIS WEAVE EMITS HAS A DEFINITION IN THE SAME DOCUMENT.
//
// #590 in one assertion. A footnote definition cannot cross a file boundary, so a set of seven
// documents fails this the moment the layer is resolved globally and split afterwards.
func TestEveryProofReferenceHasADefinitionInTheSameDocument(t *testing.T) {
	proofs := []record.Proof{
		{Label: "p-a", SHA: "s1", Basis: "reproducible", Script: "a.js", Reason: "first"},
		{Label: "p-b", SHA: "s2", Basis: "observed", Script: "b.js", Reason: "second"},
	}
	out, _ := weaveProofRefs("One<!--proof:p-b-->. Two<!--proof:p-a-->.\n", proofs)
	for _, ref := range []string{"[^P1]", "[^P2]"} {
		if !strings.Contains(out, ref+": ") {
			t.Errorf("reference %s has no definition in the document that carries it:\n%s", ref, out)
		}
	}
}

// The number is the RUN's, not the document's: P2 in the debate and P2 in the report are the
// same computation, because a per-document first-appearance number makes them different ones.
func TestProofNumbersAreRunWideNotPerDocument(t *testing.T) {
	proofs := []record.Proof{
		{Label: "p-a", SHA: "s1", Basis: "reproducible", Script: "a.js", Reason: "first"},
		{Label: "p-b", SHA: "s2", Basis: "reproducible", Script: "b.js", Reason: "second"},
	}
	// A document that anchors ONLY the second proof still calls it P2.
	out, _ := weaveProofRefs("Only the second<!--proof:p-b-->.\n", proofs)
	if !strings.Contains(out, "[^P2]") || strings.Contains(out, "[^P1]") {
		t.Errorf("proof numbering is per-document, so the same computation has two names:\n%s", out)
	}
}

// An `observed` proof must say so IN THE REPORT, with its drift — a reader must be able to
// tell a measurement of a moving system from a proof.
func TestAnObservedProofIsLabelledInTheReport(t *testing.T) {
	runDir := newRun(t)
	const sha = "feed0000"
	seedProof(t, runDir, sha, "console.log(Math.random());", "0.42\n")

	out := evidenceDoc(runDir, []record.Proof{{
		Label: "p-1", SHA: sha, Basis: "observed", Script: "s.js",
		Drift: "output differs from byte 2 between runs", Reason: "live sample",
	}}, map[string]bool{"p-1": true})
	if !strings.Contains(out, "observed") || !strings.Contains(out, "differs from byte 2") {
		t.Errorf("an observed proof was not distinguished from a reproducible one:\n%s", out)
	}
}

// A missing artifact is STATED. A proofs section quietly short of its proof is worse than
// one that admits the artifact is gone.
func TestAMissingArtifactIsStatedNotSkipped(t *testing.T) {
	runDir := newRun(t)
	out := evidenceDoc(runDir, []record.Proof{{
		Label: "p-9", SHA: "notonthisdisk", Basis: "reproducible", Script: "gone.js",
	}}, map[string]bool{"p-9": true})
	if !strings.Contains(out, "missing from this run directory") {
		t.Errorf("a missing artifact was silently omitted, so the report shows a proof it cannot produce:\n%s", out)
	}
}

// A computation on the record that NO document references is still shown, and said to be
// unreferenced. Dropping it would hide a proof the run paid to run.
func TestAnUnanchoredProofIsShownAndLabelled(t *testing.T) {
	runDir := newRun(t)
	seedProof(t, runDir, "s5", "x", "y")
	out := evidenceDoc(runDir, []record.Proof{{
		Label: "p-5", SHA: "s5", Basis: "reproducible", Script: "e.js", Reason: "nobody cited it",
	}}, map[string]bool{})
	if !strings.Contains(out, "anchored to nothing") {
		t.Errorf("a proof no document references was rendered as if it were cited:\n%s", out)
	}
}

// An anchor with no event behind it becomes an explicit unresolved line — never a crash and
// never a silent drop. Same defence weaveCitations makes for a dangling cite.
func TestADanglingProofAnchorIsMarkedUnresolved(t *testing.T) {
	// A VALID hex id with no event behind it. (An id-shaped-but-not-hex token like
	// "p-nothing" is not an anchor at all — the regex correctly ignores it, which is a
	// different case and not the one under test.)
	out, _ := weaveProofRefs("A claim<!--proof:p-abcdef01-->.\n", nil)
	if !strings.Contains(out, "unresolved proof") {
		t.Errorf("a dangling proof anchor vanished silently:\n%s", out)
	}
}

// With no proofs the document is returned unchanged — no empty section, no stray definitions.
func TestNoProofsLeavesTheReportAlone(t *testing.T) {
	const md = "A report with no computations in it.\n"
	got, used := weaveProofRefs(md, nil)
	if got != md || used != nil {
		t.Errorf("an empty proof layer was appended:\n%s", got)
	}
	if ev := evidenceDoc(recordtest.TmpRun(t), nil, nil); ev != "" {
		t.Errorf("evidence.md was composed for a run with no proofs:\n%s", ev)
	}
}

// One proof cited twice shares one number, like a citation reused.
func TestOneProofUsedTwiceSharesItsNumber(t *testing.T) {
	out, _ := weaveProofRefs("First<!--proof:p-a-->. Second<!--proof:p-a-->.\n",
		[]record.Proof{{Label: "p-a", SHA: "s1", Basis: "reproducible", Script: "a.js"}})
	if strings.Count(out, "[^P1]") != 3 { // two references plus the definition
		t.Errorf("a proof used twice did not share one number:\n%s", out)
	}
	if strings.Contains(out, "[^P2]") {
		t.Errorf("the same proof was numbered twice:\n%s", out)
	}
}
