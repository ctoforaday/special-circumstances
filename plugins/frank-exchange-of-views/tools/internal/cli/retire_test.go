package cli

import (
	"strings"
	"testing"
)

// RETIRING IS HOW A REMOVAL IS EXPLAINED, NOT HOW IT IS PERFORMED.
//
// The verb recorded whatever it was told: nothing confirmed the claim had ever been in the
// report, and nothing confirmed it had left. "Substance leaves only through the retire verb"
// — the rule the verb's own doc calls "strictly stronger than the prose rule it replaces" —
// rested entirely on the seat's word.
func TestRetireRefusesAClaimStillInTheReport(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nFive independent approaches agree.\n")

	_, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Five independent approaches agree.", "--reason", "overclaim")
	if err == nil {
		t.Fatal("a claim still present in the report was retired — the record now says it left while it sits there")
	}
	for _, want := range []string{"still in the report", "blue edit"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must say what to do instead, got: %v", err)
		}
	}
}

// A CLAIM THE RECORD CAN SHOW LEAVING is `verified`: absent now, and present in the old span
// of a recorded edit.
func TestRemovalIsVerifiedWhenAnEditTookItOut(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nFive independent approaches agree.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Five independent approaches agree.", "--new", "Two approaches agree.",
		"--answers", "R1-1", "--reason", "narrow the claim"); err != nil {
		t.Fatalf("edit refused: %v", err)
	}
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "Five independent approaches agree.", "--reason", "the independence was never established"); err != nil {
		t.Fatalf("retiring a claim an edit removed was refused: %v", err)
	}
	ev := lastOfType(t, runDir, "retire")
	if got := ev.Payload.Str("removal_basis"); got != "verified" {
		t.Errorf("removal_basis = %q, want verified — the record can show this claim leaving", got)
	}
}

// A PHANTOM RETIRE IS RECORDED AS `asserted`, NOT AS A REMOVAL.
//
// It is worse than useless when it passes for one: the scorecard computes
// unrecorded_claim_loss as (drop in claim_count) MINUS (retire events), so retiring something
// that was never there subtracts from the accounted side and cancels REAL loss — blinding the
// one detector built to catch silent deletion.
func TestAPhantomRetireIsMarkedAsserted(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSomething entirely else.\n")

	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "a claim that was never in this report", "--reason", "fuzz"); err != nil {
		t.Fatalf("retire refused: %v", err)
	}
	ev := lastOfType(t, runDir, "retire")
	if got := ev.Payload.Str("removal_basis"); got != "asserted" {
		t.Errorf("removal_basis = %q, want asserted — nothing on the record shows this claim was ever present", got)
	}
}
