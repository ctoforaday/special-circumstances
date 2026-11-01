package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// seedComputation lays down the store a computation event points at: its one-file bundle, run as
// `python3 <file>`, and the act's output. It returns the computation, its bundle and argv set.
func seedComputation(t *testing.T, runDir, label, file, body, output string) record.Computation {
	t.Helper()
	argv := []string{"python3", file}
	files := []compute.File{{Path: file, Body: []byte(body)}}
	bundle := compute.BundleSha(argv, compute.Sums(files))
	if err := compute.WriteBundle(runDir, bundle, argv, files); err != nil {
		t.Fatal(err)
	}
	if err := compute.WriteAct(runDir, bundle, label, []byte(output), nil); err != nil {
		t.Fatal(err)
	}
	return record.Computation{Label: label, SHA: bundle, Argv: argv, Basis: "reproducible"}
}

// seedStore lays down the store under a bundle name an event already holds, for a test whose
// subject is the document rather than the identity: the bundle's one file is `script.py`, as the
// events' argv names it, and label's act holds output.
func seedStore(t *testing.T, runDir, bundle, label, body, output string) {
	t.Helper()
	if err := compute.WriteBundle(runDir, bundle, []string{"python3", "script.py"}, []compute.File{{Path: "script.py", Body: []byte(body)}}); err != nil {
		t.Fatal(err)
	}
	if err := compute.WriteAct(runDir, bundle, label, []byte(output), nil); err != nil {
		t.Fatal(err)
	}
}

// THE DEFECT THIS PASS EXISTS FOR. Before it, a computation ran, cached, anchored and was
// auditable by red — and the assembled report carried the RAW anchor while mentioning the
// computation zero times. The evidence existed everywhere except the document a human reads.
//
// SPLIT IN TWO SINCE, and the split is the #590 repair: the REFERENCE and its DEFINITION are
// woven into each document that anchors a computation, and the computation itself — command,
// bundle, output, sha256 — lands once, in evidence.md.
func TestComputationsAreWovenIntoTheDeliverableWithSourceAndOutput(t *testing.T) {
	runDir := newRun(t)
	c := seedComputation(t, runDir, "K-deadbeef", "nine.py", "print('divisors of 9:', 3)", "divisors of 9: 3\n")
	c.Reason, c.Cites = "trial division settles it", "C-00001234"
	computations := []record.Computation{c}
	md := "Nine is composite by trial division<!--computation:K-deadbeef-->.\n"
	out, used := weaveComputationRefs(md, computations)

	if strings.Contains(out, "<!--computation:") {
		t.Error("a raw computation anchor shipped into the deliverable")
	}
	if len(used) != 1 || used[0] != "K-deadbeef" {
		t.Errorf("the weave did not report which computations the document anchors: %v", used)
	}
	for _, want := range []string{
		"[^K1]",                  // the visible marker at the sentence
		"[^K1]: trial division",  // AND ITS DEFINITION — the half #590 shipped without
		"evidence.md#K-deadbeef", // pointing at the computation in full
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the woven document is missing %q:\n%s", want, out)
		}
	}

	ev := evidenceDoc(runtest.Open(t, runDir), computations)
	for _, want := range []string{
		"## Computations",                   // the section
		computationAnchorMark("K-deadbeef"), // where the anchor the footnote links to ships
		"### K1 — trial division",           // a HEADING, not a second dangling reference (#590)
		"print('divisors",                   // THE PROGRAM — a computation's source IS its evidence
		"divisors of 9: 3",                  // and its output
		"basis**: reproducible",             // how far it can be trusted
		`**command**: python3 nine\.py`,     // what ran, as the bundle names its file, through Literal
		"applies the method cited",          // the cite/compute pair, legible as one argument
	} {
		if !strings.Contains(ev, want) {
			t.Errorf("evidence.md is missing %q:\n%s", want, ev)
		}
	}
	// The heading must not itself be a footnote reference — the exact shape of #590.
	if strings.Contains(ev, "### [^K") {
		t.Errorf("a computation heading was written as a footnote reference (#590):\n%s", ev)
	}
}

// A COMPUTATION WITH NO REASON IS NAMED BY ITS COMMAND, in the report's note and in the evidence
// heading alike: `compute --reason` is optional, and the command names the bundle's own file, never
// the working path the seat ran it from.
func TestAComputationWithNoReasonIsNamedByItsCommand(t *testing.T) {
	runDir := newRun(t)
	c := seedComputation(t, runDir, "K-deadbeef", "script.py", "print(91 % 7)", "0\n")
	computations := []record.Computation{c}
	out, _ := weaveComputationRefs("Seven divides 91<!--computation:K-deadbeef-->.\n", computations)
	ev := evidenceDoc(runtest.Open(t, runDir), computations)
	if !strings.Contains(out, `[^K1]: python3 script\.py`) {
		t.Errorf("the report's note does not name the command:\n%s", out)
	}
	if !strings.Contains(ev, `### K1 — python3 script\.py`) {
		t.Errorf("the evidence heading does not name the command:\n%s", ev)
	}
}

// EVERY RE-RUN AND EVERY JUDGEMENT NAMES RED, never the lens's seat and sitting: on each cell of the
// two axes — the bundle's re-run and this sentence's verdict.
func TestEveryReRunOfAComputationIsAttributedToRed(t *testing.T) {
	runDir := newRun(t)
	base := seedComputation(t, runDir, "K-deadbeef", "nine.py", "print(91 % 7)", "0\n")
	for _, rerun := range []recordpb.Rerun{recordpb.Rerun_RERUN_REPRODUCED, recordpb.Rerun_RERUN_DIFFERS, recordpb.Rerun_RERUN_CALLS_THE_TOOL} {
		for _, sound := range []bool{true, false} {
			soundness := record.SoundnessUnsound
			if sound {
				soundness = record.SoundnessSound
			}
			c := base
			c.Reason = "trial division"
			c.Rerun = &record.ComputationRerun{SeatID: "red-lens-evidence", Sitting: 5, Rerun: rerun}
			c.Soundness = soundness
			c.Judgements = []record.ComputationJudgement{{SeatID: "red-lens-evidence", Sound: sound, Note: "read it"}}
			ev := evidenceDoc(runtest.Open(t, runDir), []record.Computation{c})
			if strings.Contains(ev, "red-lens-evidence") || strings.Contains(ev, "#5") {
				t.Errorf("%s, sound=%v: a re-run is attributed to a seat:\n%s", rerun, sound, ev)
			}
			if !strings.Contains(ev, "**re-run by red") || !strings.Contains(ev, "**judged by red") {
				t.Errorf("%s, sound=%v: the re-run and the judgement do not name red:\n%s", rerun, sound, ev)
			}
		}
	}
}

// THE TWO HALVES ARE TWO LINES. A computation whose bundle reproduced and whose sentence nobody
// judged says both, and one that nobody re-ran says that — the absence of an audit is information.
func TestAComputationSaysWhatOfItWasCheckedAndWhatWasNot(t *testing.T) {
	runDir := newRun(t)
	base := seedComputation(t, runDir, "K-deadbeef", "nine.py", "print(91 % 7)", "0\n")
	base.Reason = "trial division"
	for _, c := range []struct {
		name string
		p    record.Computation
		want []string
	}{
		{"nothing checked", record.Computation{Soundness: record.SoundnessNotJudged}, []string{"**not re-run**", "**not judged**"}},
		{"reproduced, sentence not judged", record.Computation{Rerun: &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_REPRODUCED}, Soundness: record.SoundnessNotJudged},
			[]string{"it REPRODUCES", "**not judged**"}},
		{"reproduced, sentence edited after its judgement", record.Computation{Rerun: &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_REPRODUCED}, Soundness: record.SoundnessReopened},
			[]string{"it REPRODUCES", "**not judged since the sentence changed**"}},
		{"the program calls the tool", record.Computation{Rerun: &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_CALLS_THE_TOOL}, Soundness: record.SoundnessUnsound},
			[]string{"IT CANNOT BE CHECKED", "IT DOES NOT PROVE THIS SENTENCE"}},
	} {
		p := base
		p.Rerun, p.Soundness = c.p.Rerun, c.p.Soundness
		ev := evidenceDoc(runtest.Open(t, runDir), []record.Computation{p})
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
func TestEveryComputationReferenceHasADefinitionInTheSameDocument(t *testing.T) {
	computations := []record.Computation{
		{Label: "K-0000000a", SHA: "s1", Basis: "reproducible", Argv: []string{"node", "a.js"}, Reason: "first"},
		{Label: "K-0000000b", SHA: "s2", Basis: "variant", Argv: []string{"node", "b.js"}, Reason: "second"},
	}
	out, _ := weaveComputationRefs("One<!--computation:K-0000000b-->. Two<!--computation:K-0000000a-->.\n", computations)
	for _, ref := range []string{"[^K1]", "[^K2]"} {
		if !strings.Contains(out, ref+": ") {
			t.Errorf("reference %s has no definition in the document that carries it:\n%s", ref, out)
		}
	}
}

// The number is the RUN's, not the document's: K2 in the debate and K2 in the report are the
// same computation, because a per-document first-appearance number makes them different ones.
func TestComputationNumbersAreRunWideNotPerDocument(t *testing.T) {
	computations := []record.Computation{
		{Label: "K-0000000a", SHA: "s1", Basis: "reproducible", Argv: []string{"node", "a.js"}, Reason: "first"},
		{Label: "K-0000000b", SHA: "s2", Basis: "reproducible", Argv: []string{"node", "b.js"}, Reason: "second"},
	}
	// A document that anchors ONLY the second computation still calls it K2.
	out, _ := weaveComputationRefs("Only the second<!--computation:K-0000000b-->.\n", computations)
	if !strings.Contains(out, "[^K2]") || strings.Contains(out, "[^K1]") {
		t.Errorf("computation numbering is per-document, so the same computation has two names:\n%s", out)
	}
}

// A `variant` computation must say so IN THE REPORT, with what holds across its runs and what moved
// — a reader must be able to tell a measurement of a moving system from a computation.
func TestAVariantComputationIsLabelledInTheReport(t *testing.T) {
	runDir := newRun(t)
	c := seedComputation(t, runDir, "K-00000001", "s.py", "import random; print(random.random())", "0.42\n")
	c.Basis, c.Reason = "variant", "live sample"
	c.Drift, c.Invariant = "output differs from byte 2 between runs", "a number in [0, 1)"
	out := evidenceDoc(runtest.Open(t, runDir), []record.Computation{c})
	for _, want := range []string{"basis**: variant", "differs from byte 2", "**holds across both runs**: a number in [0, 1)"} {
		if !strings.Contains(out, want) {
			t.Errorf("a variant computation was not distinguished from a reproducible one — missing %q:\n%s", want, out)
		}
	}
}

// A missing artifact is STATED. A computations section quietly short of its computation is worse than
// one that admits the artifact is gone.
func TestAMissingArtifactIsStatedNotSkipped(t *testing.T) {
	runDir := newRun(t)
	out := evidenceDoc(runtest.Open(t, runDir), []record.Computation{{
		Label: "K-00000009", SHA: "notonthisdisk", Basis: "reproducible", Argv: []string{"node", "gone.js"},
	}})
	for _, want := range []string{"the bundle is missing from computations/notonthisdisk/dir/", "the output is missing from computations/notonthisdisk/acts/K-00000009/"} {
		if !strings.Contains(out, want) {
			t.Errorf("a missing artifact was silently omitted, so the report shows a computation it cannot produce — missing %q:\n%s", want, out)
		}
	}
}

// AN ACT SHOWS ITS OWN OUTPUT. Two computations on one bundle each recorded their own run, and the
// evidence document shows each act's output under its own heading, never the bundle's first.
func TestEachActShowsItsOwnOutput(t *testing.T) {
	runDir := newRun(t)
	first := seedComputation(t, runDir, "K-00000001", "s.py", "print(1)", "first act\n")
	second := first
	second.Label = "K-00000002"
	if err := compute.WriteAct(runDir, first.SHA, second.Label, []byte("second act\n"), nil); err != nil {
		t.Fatal(err)
	}
	out := evidenceDoc(runtest.Open(t, runDir), []record.Computation{first, second})
	at := strings.Index(out, "### K2")
	if at < 0 || !strings.Contains(out[:at], "first act") || strings.Contains(out[:at], "second act") || !strings.Contains(out[at:], "second act") {
		t.Errorf("an act's output is not its own:\n%s", out)
	}
}

// A FILE THAT IS NOT TEXT IS NAMED, NOT PRINTED: its path, size and sha256 on one line, never bytes
// a fence would mangle.
func TestABinaryBundleFileIsNamedNotPrinted(t *testing.T) {
	runDir := newRun(t)
	argv := []string{"python3", "s.py"}
	files := []compute.File{{Path: "s.py", Body: []byte("print(1)")}, {Path: "data.bin", Body: []byte{0, 1, 2}}}
	bundle := compute.BundleSha(argv, compute.Sums(files))
	if err := compute.WriteBundle(runDir, bundle, argv, files); err != nil {
		t.Fatal(err)
	}
	if err := compute.WriteAct(runDir, bundle, "K-00000001", []byte("1\n"), nil); err != nil {
		t.Fatal(err)
	}
	out := evidenceDoc(runtest.Open(t, runDir), []record.Computation{{Label: "K-00000001", SHA: bundle, Argv: argv, Basis: "reproducible"}})
	if !strings.Contains(out, `- data\.bin: 3 bytes, sha256 `) || !strings.Contains(out, "print(1)") {
		t.Errorf("a binary file was printed, or the text file was not:\n%s", out)
	}
}

// An anchor with no event behind it becomes an explicit unresolved line — never a crash and
// never a silent drop. Same defence weaveCitations makes for a dangling cite.
func TestADanglingComputationAnchorIsMarkedUnresolved(t *testing.T) {
	// A VALID hex id with no event behind it. (An id-shaped-but-not-hex token is not an anchor at
	// all — the reader ignores it, which is a different case and not the one under test.)
	out, _ := weaveComputationRefs("A claim<!--computation:K-abcdef01-->.\n", nil)
	if !strings.Contains(out, "unresolved computation") {
		t.Errorf("a dangling computation anchor vanished silently:\n%s", out)
	}
}

// With no computations the document is returned unchanged — no empty section, no stray definitions.
func TestNoComputationsLeavesTheReportAlone(t *testing.T) {
	const md = "A report with no computations in it.\n"
	got, used := weaveComputationRefs(md, nil)
	if got != md || used != nil {
		t.Errorf("an empty computation layer was appended:\n%s", got)
	}
	if ev := evidenceDoc(runtest.Open(t, recordtest.TmpRun(t)), nil); ev != "" {
		t.Errorf("evidence.md was composed for a run with no computations:\n%s", ev)
	}
}

// One computation cited twice shares one number, like a citation reused.
func TestOneComputationUsedTwiceSharesItsNumber(t *testing.T) {
	out, _ := weaveComputationRefs("First<!--computation:K-0000000a-->. Second<!--computation:K-0000000a-->.\n",
		[]record.Computation{{Label: "K-0000000a", SHA: "s1", Basis: "reproducible", Argv: []string{"node", "a.js"}}})
	if strings.Count(out, "[^K1]") != 3 { // two references plus the definition
		t.Errorf("a computation used twice did not share one number:\n%s", out)
	}
	if strings.Contains(out, "[^K2]") {
		t.Errorf("the same computation was numbered twice:\n%s", out)
	}
}

// A RE-RUN THAT DIFFERED SHOWS WHAT IT WROTE, fenced under its line: the drift is the program's
// bytes, and a fence one longer than any backtick run in them keeps every line of it code.
func TestADifferingReRunShowsItsDriftFenced(t *testing.T) {
	runDir := newRun(t)
	c := seedComputation(t, runDir, "K-deadbeef", "s.py", "print(1)", "1\n")
	c.Rerun = &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_DIFFERS, Drift: "on re-run the program wrote: ```\n# 2"}
	ev := evidenceDoc(runtest.Open(t, runDir), []record.Computation{c})
	if !strings.Contains(ev, "it DID NOT REPRODUCE") || !strings.Contains(ev, "  ````\n  on re-run the program wrote: ```\n  # 2\n  ````") {
		t.Errorf("the re-run's drift is not shown fenced under its line:\n%s", ev)
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
