package cli

import (
	"strings"
	"testing"
)

// A REPLY MUST NAME WHAT IT ANSWERS.
//
// A dispute is filed against ONE grade, and blue may contest more than one on the same gap.
// The answer recorded only the gap, so two disputes and one "rejected" left a record nobody
// could resolve. The ORCHESTRATOR always had the dimension — debate.js matches on (gap_id,
// dimension) — but it came off the envelope, which evaporates with the session, while the
// durable record kept the lossier half.
func TestDisputeReplyMustNameItsDimension(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nA claim.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "the harm needs two failures"); err != nil {
		t.Fatalf("dispute refused: %v", err)
	}

	_, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--as", "rejected", "--reason", "no")
	if err == nil {
		t.Fatal("an answer naming only the gap was recorded — with two dimensions contested it would resolve to neither")
	}
	if !strings.Contains(err.Error(), "--dimension is required") {
		t.Errorf("the refusal must name the missing field: %v", err)
	}

	if _, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--dimension", "severity", "--as", "rejected", "--reason", "the grade stands"); err != nil {
		t.Fatalf("a properly addressed answer was refused: %v", err)
	}
	ev := lastOfType(t, runDir, "dispute-respond")
	if got := ev.Payload.Str("dimension"); got != "severity" {
		t.Errorf("dimension = %q, want severity — the record must carry what the envelope always did", got)
	}
}

// AN ANSWER TO A QUESTION NOBODY ASKED IS REFUSED — the dangling-reference rule this codebase
// applies to gap ids, applied to the other side of the conversation.
func TestAnswerToAnUnfiledDisputeIsRefused(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nA claim.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "s"); err != nil {
		t.Fatalf("dispute refused: %v", err)
	}

	_, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "R1-1", "--dimension", "impact", "--as", "accepted", "--reason", "sure")
	if err == nil {
		t.Fatal("red answered a grade blue never contested — a reply with no question, joining to nothing")
	}
	if !strings.Contains(err.Error(), "R1-1.impact, on which no dispute was filed") {
		t.Errorf("the refusal must name the pair it could not find: %v", err)
	}
}
