package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// provenProof runs a real script through `prove` so the fixture is a proof the engine actually
// recorded — cache layout, sha and basis included — rather than a hand-built imitation of one.
func provenProof(t *testing.T, runDir, name, body string) *proof.Result {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "blue", "candidates"), 0o755); err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("blue", "candidates", name)
	if err := os.WriteFile(filepath.Join(runDir, rel), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := proof.Run(runDir, rel)
	if err != nil {
		t.Skipf("no interpreter for %s in this environment: %v", name, err)
	}
	return res
}

func seedProofEvents(t *testing.T, runDir string, rs ...*proof.Result) {
	t.Helper()
	var evs []*recordpb.Event
	for i, r := range rs {
		evs = append(evs, recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:proof:#"+itoa(i+1),
			&recordpb.Proof{
				ProofSha:   proto.String(r.SHA),
				ProofBasis: proto.String(r.Basis),
				Script:     proto.String(r.Script),
				Exit:       proto.Int32(int32(r.Exit)),
			}))
	}
	recordtest.Seed(t, runDir, evs...)
}

// THE ACCEPTANCE CASE, seeded: a proof whose recorded output no longer reproduces.
//
// The script reads a file, so its output is a measurement of state rather than a computation.
// Changing that state after the proof is recorded is exactly `pattern_ephemeral_instrument` — the
// pattern staged in the 2026-08-23 run's own gap-pattern memory and applied to none of its three
// instances, because nothing re-ran a proof.
func TestProofRerunCatchesARecordedOutputThatNoLongerReproduces(t *testing.T) {
	run := t.TempDir()
	if err := os.WriteFile(filepath.Join(run, "corpus.txt"), []byte("three\nfiles\nhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := provenProof(t, run, "count.sh", "#!/bin/sh\nwc -l < corpus.txt\n")
	if res.Basis != proof.Reproducible {
		t.Fatalf("fixture must record as reproducible, got %s", res.Basis)
	}
	seedProofEvents(t, run, res)

	// It reproduces while the state it measured is unchanged.
	if got := ProofRerunAudit(runtest.Open(t, run), 3); got.Verdict != "PASS" {
		t.Fatalf("unchanged state: want PASS, got %s — %s", got.Verdict, got.Detail)
	}

	// Move the state the proof measured. The artifact still reads as evidence; it no longer is.
	if err := os.WriteFile(filepath.Join(run, "corpus.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ProofRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an ephemeral instrument must FAIL: got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"NO LONGER REPRODUCES", "first divergence at line 1", "re-ran 1 of 1"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
}

// AN `observed` PROOF DIVERGING IS ITS NATURE, NOT A DEFECT. The basis was recorded by running
// twice and watching the output move; failing it here would punish the one evidence class the
// engine deliberately admits over documentation.
func TestProofRerunDoesNotFailAProofRecordedAsObserved(t *testing.T) {
	run := t.TempDir()
	counter := filepath.Join(run, "n")
	if err := os.WriteFile(counter, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Each execution increments, so `prove`'s two runs disagree and the basis records as observed.
	res := provenProof(t, run, "tick.sh", "#!/bin/sh\nn=$(cat n)\nn=$((n+1))\necho $n > n\necho \"tick $n\"\n")
	if res.Basis != proof.Observed {
		t.Skipf("fixture did not produce an observed basis (got %s)", res.Basis)
	}
	seedProofEvents(t, run, res)

	got := ProofRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict == "FAIL" {
		t.Fatalf("an observed proof must not FAIL for moving: %s", got.Detail)
	}
	if !strings.Contains(got.Detail, "as its recorded `observed` basis says it may") {
		t.Errorf("the report must say WHY it is not a failure; got:\n%s", got.Detail)
	}
}

// A proof whose stored artifact is gone cannot be checked by anything, which is the one state a
// proof exists to prevent. It is a failure, not a skip.
func TestProofRerunFailsAProofThatCannotBeReRunAtAll(t *testing.T) {
	run := t.TempDir()
	res := provenProof(t, run, "yes.sh", "#!/bin/sh\necho ok\n")
	seedProofEvents(t, run, res)
	if err := os.RemoveAll(filepath.Join(run, "proofs", res.SHA)); err != nil {
		t.Fatal(err)
	}
	got := ProofRerunAudit(runtest.Open(t, run), 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("a proof with no artifact must FAIL, got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "could not be re-run") {
		t.Errorf("detail must say it could not be re-run; got:\n%s", got.Detail)
	}
}

// A RUN WITH NO PROOFS AND A RUN NOBODY COULD READ ARE NOT THE SAME ANSWER, and neither is a PASS.
func TestProofRerunSeparatesNoProofsFromNoRecord(t *testing.T) {
	empty := t.TempDir()
	recordtest.Seed(t, empty)
	if got := ProofRerunAudit(runtest.Open(t, empty), 3); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "recorded no proofs") {
		t.Errorf("a run with no proofs: got %s — %s", got.Verdict, got.Detail)
	}
	// A record that cannot be READ is the third state, and it is the one worth being careful
	// about: an absent record directory is a legal empty run (the record layer draws that line
	// deliberately), so the failure has to be a real one — here, a record file that is not a
	// database. It must still say the proofs were not checked rather than implying they were.
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "records", "record.db"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ProofRerunAudit(runtest.Open(t, broken), 3)
	if got.Verdict != "SKIP" {
		t.Fatalf("unreadable record: got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "NOT a run whose proofs reproduce") {
		t.Errorf("the miss must not read as a clean board; got:\n%s", got.Detail)
	}
}

// The sample is bounded and prefers proofs no seat has re-run — where the audit is worth most —
// and the count of what it actually looked at always reaches the report.
func TestProofRerunSamplesTheUnauditedFirstAndSaysHowManyItRan(t *testing.T) {
	run := t.TempDir()
	var rs []*proof.Result
	for _, n := range []string{"a", "b", "c", "d"} {
		rs = append(rs, provenProof(t, run, n+".sh", "#!/bin/sh\necho "+n+"\n"))
	}
	seedProofEvents(t, run, rs...)
	// One of them has been re-run by a seat already; it must be sampled LAST.
	recordtest.Seed(t, run, recordtest.At(t, "red-lens-r1-evidence", 1, "red-lens-r1-evidence:reproduce:#1",
		&recordpb.Reproduce{ProofSha: proto.String(rs[0].SHA), Reproduced: proto.Bool(true)}))

	got := ProofRerunAudit(runtest.Open(t, run), 2)
	if !strings.Contains(got.Detail, "re-ran 2 of 4") {
		t.Errorf("the sample size must reach the report; got:\n%s", got.Detail)
	}
	if !strings.Contains(got.Detail, "3 had never been re-run by any seat") {
		t.Errorf("the unaudited count is the reason this audit exists; got:\n%s", got.Detail)
	}
}
