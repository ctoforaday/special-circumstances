package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE EXIT CODE IS THE ONLY PART OF A CAPTURE ANYTHING AUTOMATED READS, and it had no
// test. Run assembles the audits and then decides, in four lines, whether the capture
// FAILED; invert that decision and a healthy run exits non-zero while a run whose audits
// all failed exits 0 — a green capture over a broken record, which is the shape every
// gate in this repository exists to refuse.
//
// The assertion does not predict any verdict: it recomputes the contract INDEPENDENTLY
// from the audits Run itself returned, so it holds for whatever a fixture produces and
// says the one thing that must always be true — the code Run reports agrees with the
// verdicts Run reports.
func TestTheExitCodeAgreesWithTheAuditsItReports(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(run.Dir(), "trajectories"), 0o755); err != nil {
		t.Fatal(err)
	}
	transcripts := t.TempDir()
	if err := os.WriteFile(filepath.Join(transcripts, "journal.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	audits, report, exitFail, err := Run(run, transcripts, time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(audits) == 0 {
		t.Fatal("Run returned no audits — the gate below would be vacuous, which is the " +
			"failure this test exists to refuse")
	}

	wantFail := false
	for _, a := range audits {
		if a.Verdict == "FAIL" {
			wantFail = true
		}
	}
	if exitFail != wantFail {
		t.Errorf("Run reported exitFail=%v while its own audits say %v", exitFail, wantFail)
		for _, a := range audits {
			t.Logf("  %s: %s", a.Check, a.Verdict)
		}
	}

	// And the mechanics the exit code sits on top of: the journal reached the run, and the
	// audit report was written where a later reader looks for it.
	for _, p := range []string{
		filepath.Join(run.Dir(), "trajectories", "journal.jsonl"),
		filepath.Join(run.Dir(), "run-record-audit.md"),
	} {
		if _, serr := os.Stat(p); serr != nil {
			t.Errorf("capture did not leave %s: %v", p, serr)
		}
	}
	if report == "" {
		t.Error("capture returned an empty report")
	}
}

// A RUN WITH NO WORKFLOW JOURNAL CAN BE CLOSED, AND IS NOT CALLED CLEAN FOR IT.
//
// Capture is the only code that clears a run's live marker, and it used to fail at its first
// line when journal.jsonl was absent. So a run driven outside the Workflow tool — a headless
// harness, or a workflow killed before it wrote its journal — could never be captured, and its
// marker blocked plugin updates indefinitely: three such runs did exactly that on 2026-09-10.
//
// The other half is what the fix must NOT do. Without a journal the envelope side of log-parity
// is empty, nothing is "silent", and LogAudit would PASS — a clean verdict produced by the
// absence of one of its inputs. It must say SKIP, and the report must say why.
func TestARunWithNoJournalIsCapturedAndItsEnvelopeAuditsAreNotMeasured(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(run.Dir(), "trajectories"), 0o755); err != nil {
		t.Fatal(err)
	}
	transcripts := t.TempDir() // deliberately empty: no journal.jsonl
	audits, report, _, err := Run(run, transcripts, time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("capture refused a run with no journal: %v — so its live marker can never be cleared", err)
	}
	var parity *Audit
	for i := range audits {
		if audits[i].Check == "log-parity" {
			parity = &audits[i]
		}
	}
	if parity == nil {
		t.Fatal("no log-parity audit in the result")
	}
	if parity.Verdict != "SKIP" {
		t.Errorf("log-parity = %s with no envelopes to compare — an absent input read as a clean pass: %s", parity.Verdict, parity.Detail)
	}
	if !strings.Contains(report, "journal: none") {
		t.Error("the capture report does not say the journal was absent, so a reader cannot tell a run with no envelopes from one whose envelopes all agreed")
	}
	if _, serr := os.Stat(filepath.Join(run.Dir(), "trajectories", "journal.jsonl")); serr == nil {
		t.Error("capture wrote a journal.jsonl into the run although none existed")
	}
}
