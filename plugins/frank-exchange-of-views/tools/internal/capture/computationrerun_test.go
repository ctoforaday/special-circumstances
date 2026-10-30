package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// provenComputation runs a real program through `compute` so the fixture is a computation the engine actually
// recorded — store layout, sha and basis included — rather than a hand-built imitation of one. The program
// is the one file of its own bundle directory, run by the interpreter its extension names.
func provenComputation(t *testing.T, runDir, name, body string) *compute.Result {
	t.Helper()
	dir := filepath.Join(runDir, "scratch", strings.TrimSuffix(name, filepath.Ext(name)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	interp := map[string]string{".sh": "bash", ".py": "python3", ".js": "node"}[filepath.Ext(name)]
	prog, err := compute.ReadProgram(dir, []string{interp, name}, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	res, err := compute.Run(runDir, prog)
	if err != nil {
		t.Skipf("no interpreter for %s in this environment: %v", name, err)
	}
	if err := compute.WriteBundle(runDir, res.Bundle, res.Argv, prog.Files); err != nil {
		t.Fatal(err)
	}
	return res
}

func seedComputationEvents(t *testing.T, runDir string, rs ...*compute.Result) {
	t.Helper()
	var evs []*recordpb.Event
	for i, r := range rs {
		id := fmt.Sprintf("K-%08x", i+1)
		if err := compute.WriteAct(runDir, r.Bundle, id, []byte(r.Output), r.Out); err != nil {
			t.Fatal(err)
		}
		c := &recordpb.Computation{ComputationId: proto.String(id), Argv: r.Argv, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
			BundleSha: proto.String(r.Bundle), Exit: proto.Int32(int32(r.Exit))}
		if r.Drift != "" {
			c.Basis, c.Drift, c.Invariant = recordpb.Basis_BASIS_VARIANT.Enum(), proto.String(r.Drift), proto.String("the fixture's invariant")
		}
		evs = append(evs, recordtest.At(t, "blue-lane-1", "blue-lane-1:computation:#"+itoa(i+1), c))
	}
	recordtest.Seed(t, runDir, evs...)
}

// THE ACCEPTANCE CASE, seeded: a computation whose recorded output no longer reproduces.
//
// The program reads a file OUTSIDE its bundle, by absolute path, so its output is a measurement of
// state rather than a computation.
// Changing that state after the computation is recorded is the ephemeral instrument — a failure the
// 2026-08-23 run made three times, because nothing re-ran a computation.
func TestComputationRerunCatchesARecordedOutputThatNoLongerReproduces(t *testing.T) {
	run := t.TempDir()
	if err := os.WriteFile(filepath.Join(run, "corpus.txt"), []byte("three\nfiles\nhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := provenComputation(t, run, "count.sh", "#!/bin/sh\nwc -l < '"+filepath.Join(run, "corpus.txt")+"'\n")
	if res.Drift != "" {
		t.Fatalf("fixture must record as reproducible, drifted: %s", res.Drift)
	}
	seedComputationEvents(t, run, res)

	// It reproduces while the state it measured is unchanged.
	if got := ComputationRerunAudit(runtest.Open(t, run), 3); got.Verdict != "PASS" {
		t.Fatalf("unchanged state: want PASS, got %s — %s", got.Verdict, got.Detail)
	}

	// Move the state the computation measured. The artifact still reads as evidence; it no longer is.
	if err := os.WriteFile(filepath.Join(run, "corpus.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ComputationRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an ephemeral instrument must FAIL: got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"NO LONGER REPRODUCES", "first divergence at line 1", "re-ran 1 of 1"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
}

// A `variant` COMPUTATION DIVERGING IS ITS NATURE, NOT A DEFECT. The basis was recorded by running
// twice and watching the output move; failing it here would punish the one evidence class the
// engine deliberately admits over documentation.
func TestComputationRerunDoesNotFailAComputationRecordedAsVariant(t *testing.T) {
	run := t.TempDir()
	counter := filepath.Join(run, "n")
	if err := os.WriteFile(counter, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Each execution increments, so `compute`'s two runs disagree and the basis records as variant.
	// The counter lies outside the bundle: each run starts from a fresh copy, so one inside it never moves.
	res := provenComputation(t, run, "tick.sh", "#!/bin/sh\nn=$(cat '"+counter+"')\nn=$((n+1))\necho $n > '"+counter+"'\necho \"tick $n\"\n")
	if res.Drift == "" {
		t.Skip("fixture's two runs agreed")
	}
	seedComputationEvents(t, run, res)

	got := ComputationRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict == "FAIL" {
		t.Fatalf("a variant computation must not FAIL for moving: %s", got.Detail)
	}
	if !strings.Contains(got.Detail, "as its recorded `variant` basis says it may") {
		t.Errorf("the report must say WHY it is not a failure; got:\n%s", got.Detail)
	}
}

// A computation whose stored artifact is gone cannot be checked by anything, which is the one state a
// computation exists to prevent. It is a failure, not a skip.
func TestComputationRerunFailsAComputationThatCannotBeReRunAtAll(t *testing.T) {
	run := t.TempDir()
	res := provenComputation(t, run, "yes.sh", "#!/bin/sh\necho ok\n")
	seedComputationEvents(t, run, res)
	if err := os.RemoveAll(compute.BundlePath(run, res.Bundle)); err != nil {
		t.Fatal(err)
	}
	got := ComputationRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("a computation with no artifact must FAIL, got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "could not be re-run") {
		t.Errorf("detail must say it could not be re-run; got:\n%s", got.Detail)
	}
}

// A RUN WITH NO COMPUTATIONS AND A RUN NOBODY COULD READ ARE NOT THE SAME ANSWER, and neither is a PASS.
func TestComputationRerunSeparatesNoComputationsFromNoRecord(t *testing.T) {
	empty := t.TempDir()
	recordtest.Seed(t, empty)
	if got := ComputationRerunAudit(runtest.Open(t, empty), 3); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "recorded no computations") {
		t.Errorf("a run with no computations: got %s — %s", got.Verdict, got.Detail)
	}
	// A record that cannot be READ is the third state, and it is the one worth being careful
	// about: an absent record directory is a legal empty run (the record layer draws that line
	// deliberately), so the failure has to be a real one — here, a record file that is not a
	// database. It must still say the computations were not checked rather than implying they were.
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "records", "record.db"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ComputationRerunAudit(runtest.Open(t, broken), 3)
	if got.Verdict != "SKIP" {
		t.Fatalf("unreadable record: got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "NOT a run whose computations reproduce") {
		t.Errorf("the miss must not read as a clean board; got:\n%s", got.Detail)
	}
}

// The sample is bounded and prefers computations no seat has re-run — where the audit is worth most —
// and the count of what it actually looked at always reaches the report.
func TestComputationRerunSamplesTheUnauditedFirstAndSaysHowManyItRan(t *testing.T) {
	run := t.TempDir()
	var rs []*compute.Result
	for _, n := range []string{"a", "b", "c", "d"} {
		rs = append(rs, provenComputation(t, run, n+".sh", "#!/bin/sh\necho "+n+"\n"))
	}
	seedComputationEvents(t, run, rs...)
	// One of them has been re-run by a seat already; it must be sampled LAST.
	recordtest.Seed(t, run, recordtest.At(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1",
		&recordpb.Reproduce{ComputationId: proto.String("K-00000001"), BundleSha: proto.String(rs[0].Bundle), Rerun: recordpb.Rerun_RERUN_REPRODUCED.Enum(), Ran: proto.Bool(true)}))

	got := ComputationRerunAudit(runtest.Open(t, run), 2)
	if !strings.Contains(got.Detail, "re-ran 2 of 4 distinct bundle(s), behind 2 of the 4 recorded computation(s)") {
		t.Errorf("the sample size must reach the report; got:\n%s", got.Detail)
	}
	if !strings.Contains(got.Detail, "3 bundle(s) had never been re-run by any seat") {
		t.Errorf("the unaudited count is the reason this audit exists; got:\n%s", got.Detail)
	}
}

// COMPUTATIONS THAT SHARE A BUNDLE ARE COUNTED AS COMPUTATIONS, and the bundle as one re-run (smoke
// 07d: 9 computations over 6 bundles read "3 of 6 recorded proof(s)"). Three computations over
// two bundles, one of them recorded twice: re-running both covers all three.
func TestComputationRerunCountsTheComputationsBehindEachBundle(t *testing.T) {
	run := t.TempDir()
	a := provenComputation(t, run, "a.sh", "#!/bin/sh\necho a\n")
	b := provenComputation(t, run, "b.sh", "#!/bin/sh\necho b\n")
	seedComputationEvents(t, run, a, b, a)
	// A sample of one takes the bundle whose sha sorts first; the computations behind it are the count.
	behind := 1
	if a.Bundle < b.Bundle {
		behind = 2
	}
	got := ComputationRerunAudit(runtest.Open(t, run), 1)
	if want := fmt.Sprintf("re-ran 1 of 2 distinct bundle(s), behind %d of the 3 recorded computation(s)", behind); !strings.Contains(got.Detail, want) {
		t.Errorf("the count must name the bundles and the computations behind the ones re-run; got:\n%s", got.Detail)
	}
	got = ComputationRerunAudit(runtest.Open(t, run), 3)
	if !strings.Contains(got.Detail, "re-ran 2 of 2 distinct bundle(s), behind 3 of the 3 recorded computation(s)") {
		t.Errorf("got:\n%s", got.Detail)
	}
}
