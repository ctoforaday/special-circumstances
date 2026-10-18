package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

	_, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "Five independent approaches agree.", "--reason", "overclaim")
	if err == nil {
		t.Fatal("a claim still present in the report was retired — the record now says it left while it sits there")
	}
	for _, want := range []string{"still in the report", "`edit`"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must say what to do instead, got: %v", err)
		}
	}
}

// A RETIRE THAT NAMES NO CLAIM IS REFUSED IN THE CLAIM'S OWN WORDS. On universe-m29 `retire --quote ""`
// was told `"" is still in the report`, since every report contains the empty string, and the seat
// was sent to edit out a claim it had not named.
func TestRetireRefusesABlankQuoteAsMissing(t *testing.T) {
	for _, quote := range [][]string{{"--quote", ""}, {"--quote", "  "}, nil} {
		runDir := newRun(t)
		writeReport(t, runDir, "# H\n\nFive independent approaches agree.\n")
		args := append([]string{"retire", "--run", runDir, "--seat-id", "blue-respond", "--reason", "overclaim"}, quote...)
		_, err := run(t, args...)
		if err == nil || !strings.Contains(err.Error(), "requires --quote") || strings.Contains(err.Error(), "still in the report") {
			t.Errorf("retire %q: want the missing-quote refusal, got %v", quote, err)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
			t.Errorf("retire %q recorded %d retire(s)", quote, n)
		}
	}
}

// A CLAIM THE RECORD CAN SHOW LEAVING is `verified`: absent now, and present in the old span
// of a recorded edit.
func TestRemovalIsVerifiedWhenAnEditTookItOut(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nFive independent approaches agree.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := runAt(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "Five independent approaches agree.", "--new", "Two approaches agree.",
		"--answers", "G1", "--reason", "narrow the claim"); err != nil {
		t.Fatalf("edit refused: %v", err)
	}
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "Five independent approaches agree.", "--reason", "the independence was never established"); err != nil {
		t.Fatalf("retiring a claim an edit removed was refused: %v", err)
	}
	ev := lastBody(t, runDir, &recordpb.Retire{})
	if got := ev.GetRemovalBasis(); got != "verified" {
		t.Errorf("removal_basis = %q, want verified — the record can show this claim leaving", got)
	}
}

// A PHANTOM RETIRE IS RECORDED AS `asserted`, NOT AS A REMOVAL.
//
// It is worse than useless when it passes for one: the changelog lists every retirement as a
// claim that left the report, so retiring something that was never there tells the reader of a
// removal the report never had.
func TestAPhantomRetireIsMarkedAsserted(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSomething entirely else.\n")

	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "a claim that was never in this report", "--reason", "fuzz"); err != nil {
		t.Fatalf("retire refused: %v", err)
	}
	ev := lastBody(t, runDir, &recordpb.Retire{})
	if got := ev.GetRemovalBasis(); got != "asserted" {
		t.Errorf("removal_basis = %q, want asserted — nothing on the record shows this claim was ever present", got)
	}
}
