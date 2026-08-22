package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeScript lays a proof script under the run directory, where `blue prove` resolves it.
func writeScript(t *testing.T, runDir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(runDir, name), []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// THE DEMAND RED COULD NOT MAKE (#277).
//
// Across six recorded runs no seat ever wrote a program to settle a question, and the
// 2026-08-05 smoke shows why: it was never the invitation that was missing, it was the ASK.
// All TEN of red's acceptance checks were document probes — R1-1 was literally "execute the
// assembly step" — so red could only ever ask whether the report SAYS something. Meanwhile
// blue/frontier.md stated the divisor enumeration in executable terms and blue ran it zero
// times.
//
// `--check-kind computation` is a demand prose cannot answer, and these tests are what make
// it a demand rather than a label.

func mintComputation(t *testing.T, runDir, key string) {
	t.Helper()
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", key, "--class", "unverified-arithmetic",
		"--problem", "the primality claim is asserted, not computed",
		"--check-kind", "computation", "--check", "trial division over 2..6 returns no divisor",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("minting a computation check was refused: %v", err)
	}
}

// A computation check cannot be closed on prose. This is the round the smoke actually spent:
// R1-2 asked blue to test a false claim, blue ASSERTED the test had happened, and red's R2-2
// correctly refused it — a full round for something three lines settle.
func TestAComputationCheckCannotBeClosedWithoutAProof(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintComputation(t, runDir, "G1")

	_, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "blue/report.md", "--reason", "blue says it checked; looks right to me")
	if err == nil {
		t.Fatal("a computation gap closed with no proof — red accepted the one kind of evidence it declared insufficient")
	}
	for _, want := range []string{"check-kind computation", "blue prove"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not mention %q, so nobody learns the way out:\n%v", want, err)
		}
	}
}

// A proof that NAMES the gap closes it. The join is a field, not a gap id mentioned in prose
// — the convention blue_edit's --answers replaced after measuring 73% reliable.
func TestAProofAnsweringTheGapClosesIt(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintComputation(t, runDir, "G1")
	writeScript(t, runDir, "seven.js", "console.log([2,3,4,5,6].filter(n=>7%n===0).length===0);")

	if _, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Seven is prime.", "--script", "seven.js", "--answers", "R1-1",
		"--reason", "trial division settles it"); err != nil {
		t.Fatalf("prove refused: %v", err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "lens reproduce",
		"--verified-against", "proofs/", "--reason", "re-ran the proof; same bytes"); err != nil {
		t.Fatalf("a proved computation gap would not close: %v", err)
	}
}

// A proof that answers a DIFFERENT gap does not close this one. Otherwise any computation
// anywhere in the run would satisfy every computation check, which is the join being
// decorative.
func TestAProofForAnotherGapDoesNotClose(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\nNine is composite.\n")
	mintComputation(t, runDir, "G1")
	mintComputation(t, runDir, "G2")
	writeScript(t, runDir, "nine.js", "console.log(9%3===0);")

	if _, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Nine is composite.", "--script", "nine.js", "--answers", "R1-2",
		"--reason", "settles the second claim"); err != nil {
		t.Fatalf("prove refused: %v", err)
	}
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "x", "--reason", "there is a proof in this run"); err == nil {
		t.Error("a proof answering R1-2 closed R1-1 — the --answers join is not being read")
	}
}

// A DOCUMENT check is unaffected. The guard is scoped to the kind that declared prose
// insufficient; making every closure need a proof would be a different (and wrong) rule.
func TestADocumentCheckStillClosesOnProse(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintGap(t, runDir, "G1", "overclaim")

	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--as", "repaired", "--verified-by", "L1", "--verified-with", "read",
		"--verified-against", "blue/report.md", "--reason", "the section no longer claims it"); err != nil {
		t.Fatalf("a document check was blocked by the computation guard: %v", err)
	}
}

// --answers is checked against the board like every other reference. A dangling one is
// refused at the write, not accepted and dropped at replay.
func TestProveRefusesAnUnknownGap(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	mintComputation(t, runDir, "G1")
	writeScript(t, runDir, "seven.js", "console.log(true);")

	if _, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Seven is prime.", "--script", "seven.js", "--answers", "R9-9",
		"--reason", "x"); err == nil {
		t.Error("prove --answers accepted a gap no mint created")
	}
}

// The kind is a CLOSED set: a near-miss must fail rather than record a word every downstream
// comparison will silently miss.
func TestCheckKindIsEnforcedAsAnEnum(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "x",
		"--problem", "p",
		"--check-kind", "compute", "--check", "c",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	if err == nil {
		t.Fatal("--check-kind compute was recorded; a near-miss reads as neither kind downstream")
	}
	if !strings.Contains(err.Error(), "computation") {
		t.Errorf("the refusal does not list the legal set:\n%v", err)
	}
}
