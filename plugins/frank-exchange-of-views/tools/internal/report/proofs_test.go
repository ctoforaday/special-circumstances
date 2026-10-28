package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
		Label: "P-deadbeef", SHA: sha, Basis: "reproducible",
		Script: "nine.js", Reason: "trial division settles it", Cites: "C-00001234",
	}}
	md := "Nine is composite by trial division<!--proof:P-deadbeef-->.\n"
	out, used := weaveProofRefs(md, proofs)

	if strings.Contains(out, "<!--proof:") {
		t.Error("a raw proof anchor shipped into the deliverable")
	}
	if len(used) != 1 || used[0] != "P-deadbeef" {
		t.Errorf("the weave did not report which proofs the document anchors: %v", used)
	}
	for _, want := range []string{
		"[^P1]",                  // the visible marker at the sentence
		"[^P1]: trial division",  // AND ITS DEFINITION — the half #590 shipped without
		"evidence.md#P-deadbeef", // pointing at the computation in full
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the woven document is missing %q:\n%s", want, out)
		}
	}

	ev := evidenceDoc(runtest.Open(t, runDir), proofs, map[string]bool{"P-deadbeef": true})
	for _, want := range []string{
		"## Proofs",                      // the section
		proofAnchorMark("P-deadbeef"),    // where the anchor the footnote links to ships
		"### P1 — trial division",        // a HEADING, not a second dangling reference (#590)
		"console.log('divisors",          // THE SCRIPT — a computation's source IS its evidence
		"divisors of 9: 3",               // and its output
		"basis**: reproducible",          // how far it can be trusted
		"`proofs/" + sha + "/script.js`", // which script: the run's own copy, never the path it ran from
		"applies the method cited",       // the cite/compute pair, legible as one argument
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

// A PROOF WITH NO REASON IS NAMED BY THE RUN'S COPY OF ITS SCRIPT, in the report's note and in the
// evidence heading alike: `prove --reason` is optional, and the path the seat ran is a working file
// in the operator's home directory on a real run.
func TestAProofWithNoReasonIsNamedByItsStoredScript(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "print(91 % 7)", "0\n")
	proofs := []record.Proof{{Label: "P-deadbeef", SHA: sha, Basis: "reproducible", Script: "/home/someone/scratch/nine.py"}}
	out, _ := weaveProofRefs("Seven divides 91<!--proof:P-deadbeef-->.\n", proofs)
	ev := evidenceDoc(runtest.Open(t, runDir), proofs, map[string]bool{"P-deadbeef": true})
	const stored = "proofs/" + sha + "/script.py"
	if !strings.Contains(out, "[^P1]: "+stored) {
		t.Errorf("the report's note does not name the stored script:\n%s", out)
	}
	if !strings.Contains(ev, "### P1 — "+stored) {
		t.Errorf("the evidence heading does not name the stored script:\n%s", ev)
	}
	if strings.Contains(out+ev, "/home/") {
		t.Errorf("a home-directory path shipped:\n%s\n%s", out, ev)
	}
}

// EVERY RE-RUN AND EVERY JUDGEMENT NAMES RED, never the lens's seat and sitting: on each cell of the
// two axes — the script's re-run and this sentence's verdict.
func TestEveryReRunOfAProofIsAttributedToRed(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "print(91 % 7)", "0\n")
	for _, rerun := range []recordpb.Rerun{recordpb.Rerun_RERUN_REPRODUCED, recordpb.Rerun_RERUN_DIFFERS, recordpb.Rerun_RERUN_CALLS_THE_TOOL} {
		for _, sound := range []bool{true, false} {
			soundness := record.SoundnessUnsound
			if sound {
				soundness = record.SoundnessSound
			}
			proofs := []record.Proof{{Label: "P-deadbeef", SHA: sha, Basis: "reproducible", Script: "nine.py", Reason: "trial division",
				Rerun:      &record.ProofRerun{SeatID: "red-lens-evidence", Sitting: 5, Rerun: rerun},
				Soundness:  soundness,
				Judgements: []record.ProofJudgement{{SeatID: "red-lens-evidence", Sound: sound, Note: "read it"}}}}
			ev := evidenceDoc(runtest.Open(t, runDir), proofs, map[string]bool{"P-deadbeef": true})
			if strings.Contains(ev, "red-lens-evidence") || strings.Contains(ev, "#5") {
				t.Errorf("%s, sound=%v: a re-run is attributed to a seat:\n%s", rerun, sound, ev)
			}
			if !strings.Contains(ev, "**re-run by red") || !strings.Contains(ev, "**judged by red") {
				t.Errorf("%s, sound=%v: the re-run and the judgement do not name red:\n%s", rerun, sound, ev)
			}
		}
	}
}

// THE TWO HALVES ARE TWO LINES. A proof whose script reproduced and whose sentence nobody judged
// says both, and one that nobody re-ran says that — the absence of an audit is information.
func TestAProofSaysWhatOfItWasCheckedAndWhatWasNot(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "print(91 % 7)", "0\n")
	for _, c := range []struct {
		name string
		p    record.Proof
		want []string
	}{
		{"nothing checked", record.Proof{Soundness: record.SoundnessNotJudged}, []string{"**not re-run**", "**not judged**"}},
		{"reproduced, sentence not judged", record.Proof{Rerun: &record.ProofRerun{Rerun: recordpb.Rerun_RERUN_REPRODUCED}, Soundness: record.SoundnessNotJudged},
			[]string{"it REPRODUCES", "**not judged**"}},
		{"reproduced, sentence edited after its judgement", record.Proof{Rerun: &record.ProofRerun{Rerun: recordpb.Rerun_RERUN_REPRODUCED}, Soundness: record.SoundnessReopened},
			[]string{"it REPRODUCES", "**not judged since the sentence changed**"}},
		{"the script calls the tool", record.Proof{Rerun: &record.ProofRerun{Rerun: recordpb.Rerun_RERUN_CALLS_THE_TOOL}, Soundness: record.SoundnessUnsound},
			[]string{"IT CANNOT BE CHECKED", "IT DOES NOT PROVE THIS SENTENCE"}},
	} {
		c.p.Label, c.p.SHA, c.p.Basis, c.p.Script, c.p.Reason = "P-deadbeef", sha, "reproducible", "nine.py", "trial division"
		ev := evidenceDoc(runtest.Open(t, runDir), []record.Proof{c.p}, map[string]bool{"P-deadbeef": true})
		for _, w := range c.want {
			if !strings.Contains(ev, w) {
				t.Errorf("%s: the evidence document does not say %q:\n%s", c.name, w, ev)
			}
		}
	}
}

// EVERY REFERENCE THIS WEAVE EMITS HAS A DEFINITION IN THE SAME DOCUMENT.
//
// #590 in one assertion. A footnote definition cannot cross a file boundary, so a set of seven
// documents fails this the moment the layer is resolved globally and split afterwards.
func TestEveryProofReferenceHasADefinitionInTheSameDocument(t *testing.T) {
	proofs := []record.Proof{
		{Label: "P-0000000a", SHA: "s1", Basis: "reproducible", Script: "a.js", Reason: "first"},
		{Label: "P-0000000b", SHA: "s2", Basis: "observed", Script: "b.js", Reason: "second"},
	}
	out, _ := weaveProofRefs("One<!--proof:P-0000000b-->. Two<!--proof:P-0000000a-->.\n", proofs)
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
		{Label: "P-0000000a", SHA: "s1", Basis: "reproducible", Script: "a.js", Reason: "first"},
		{Label: "P-0000000b", SHA: "s2", Basis: "reproducible", Script: "b.js", Reason: "second"},
	}
	// A document that anchors ONLY the second proof still calls it P2.
	out, _ := weaveProofRefs("Only the second<!--proof:P-0000000b-->.\n", proofs)
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

	out := evidenceDoc(runtest.Open(t, runDir), []record.Proof{{
		Label: "P-00000001", SHA: sha, Basis: "observed", Script: "s.js",
		Drift: "output differs from byte 2 between runs", Reason: "live sample",
	}}, map[string]bool{"P-00000001": true})
	if !strings.Contains(out, "observed") || !strings.Contains(out, "differs from byte 2") {
		t.Errorf("an observed proof was not distinguished from a reproducible one:\n%s", out)
	}
}

// A missing artifact is STATED. A proofs section quietly short of its proof is worse than
// one that admits the artifact is gone.
func TestAMissingArtifactIsStatedNotSkipped(t *testing.T) {
	runDir := newRun(t)
	out := evidenceDoc(runtest.Open(t, runDir), []record.Proof{{
		Label: "P-00000009", SHA: "notonthisdisk", Basis: "reproducible", Script: "gone.js",
	}}, map[string]bool{"P-00000009": true})
	if !strings.Contains(out, "missing from proofs/notonthisdisk/") {
		t.Errorf("a missing artifact was silently omitted, so the report shows a proof it cannot produce:\n%s", out)
	}
}

// A computation on the record that NO document references is still shown, and said to be
// unreferenced. Dropping it would hide a proof the run paid to run.
func TestAnUnanchoredProofIsShownAndLabelled(t *testing.T) {
	runDir := newRun(t)
	seedProof(t, runDir, "s5", "x", "y")
	out := evidenceDoc(runtest.Open(t, runDir), []record.Proof{{
		Label: "P-00000005", SHA: "s5", Basis: "reproducible", Script: "e.js", Reason: "nobody cited it",
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
	out, _ := weaveProofRefs("A claim<!--proof:P-abcdef01-->.\n", nil)
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
	if ev := evidenceDoc(runtest.Open(t, recordtest.TmpRun(t)), nil, nil); ev != "" {
		t.Errorf("evidence.md was composed for a run with no proofs:\n%s", ev)
	}
}

// One proof cited twice shares one number, like a citation reused.
func TestOneProofUsedTwiceSharesItsNumber(t *testing.T) {
	out, _ := weaveProofRefs("First<!--proof:P-0000000a-->. Second<!--proof:P-0000000a-->.\n",
		[]record.Proof{{Label: "P-0000000a", SHA: "s1", Basis: "reproducible", Script: "a.js"}})
	if strings.Count(out, "[^P1]") != 3 { // two references plus the definition
		t.Errorf("a proof used twice did not share one number:\n%s", out)
	}
	if strings.Contains(out, "[^P2]") {
		t.Errorf("the same proof was numbered twice:\n%s", out)
	}
}

// AN EMPTY OUTPUT IS SAID TO BE EMPTY. A re-run that printed nothing was shown as two backticks with
// nothing between them, which are no code span: a reader saw two stray characters where the record
// holds no output at all.
func TestAnEmptyReRunOutputSaysSo(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	seedProof(t, runDir, sha, "print(1)", "1\n")
	proofs := []record.Proof{{Label: "P-deadbeef", SHA: sha, Basis: "reproducible", Script: "s.py",
		Rerun: &record.ProofRerun{Rerun: recordpb.Rerun_RERUN_DIFFERS, Recorded: "1", Observed: " \n"}}}
	ev := evidenceDoc(runtest.Open(t, runDir), proofs, map[string]bool{"P-deadbeef": true})
	if !strings.Contains(ev, "  - on re-run: _(empty)_\n") || strings.Contains(ev, "on re-run: ``") {
		t.Errorf("the empty re-run output is not said to be empty:\n%s", ev)
	}
}

// A NOTE IS READ WHERE IT STANDS: the first, and every note of one line, after the judgement on its
// item's line, so its first line is prose behind the tool's text, and a "```" opening it opens no
// fence whose info string would keep a token. A token in a code span it holds stays, and a note that
// is a paragraph of its own keeps one in its fence's info string.
func TestAJudgementNoteIsStrippedWhereItShips(t *testing.T) {
	for _, c := range []struct {
		notes []string
		keep  int // tokens that are code where they ship
	}{
		{[]string{"```x <!--gap:G-00000001-->"}, 0},
		{[]string{"```x <!--gap:G-00000001-->\nmore"}, 0},
		{[]string{"```x <!--gap:G-00000001-->", "second"}, 0},
		{[]string{"first", "```x <!--gap:G-00000001-->"}, 0},
		{[]string{"a `<!--gap:G-00000001-->` b\nmore"}, 1},
		{[]string{"first\nline", "```x <!--gap:G-00000001-->\ny\n```"}, 1},
	} {
		got := judgementNotes(append([]string(nil), c.notes...))
		if n := strings.Count(got, "<!--gap:"); n != c.keep {
			t.Errorf("judgementNotes(%q) keeps %d tokens, want %d:\n%s", c.notes, n, c.keep, got)
		}
	}
}
