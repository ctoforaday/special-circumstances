package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"strings"
	"testing"
)

// A REPLY MUST NAME WHAT IT ANSWERS — and after #344 it names an ID, not a pair.
//
// THE HISTORY MATTERS BECAUSE IT IS THE ARGUMENT FOR THE ID. A dispute was filed against ONE
// grade, and blue could contest more than one on the same gap. The answer recorded only the gap,
// so two disputes and one "rejected" left a record nobody could resolve. The fix at the time was
// to make `--dimension` required on the answer, so the join became (gap_id, dimension) — the
// orchestrator had always matched on that pair off the ENVELOPE, which evaporates with the
// session, while the durable record kept the lossier half.
//
// That pair was still a composed key, and #312 is where it ran out: two petitions from one seat
// in one round could not be told apart at all, because their pair had no third component to
// reach for. A motion has an id. `motion grade rule --id M1` answers M1 and nothing else, so
// both failures below stop being checks against a validator and become unrepresentable.
//
// These tests survive the collapse asserting the SAME property against the new mechanism: an
// answer joins to exactly one ask, and an answer to an ask nobody filed is refused.

// TWO MOTIONS ON THE SAME GAP AND THE SAME DIMENSION ARE STILL DISTINCT.
//
// This is the case the old pair key could not represent at all, and it is why the id exists. Blue
// contests R1-1's severity, is refused, and contests it again on new grounds; the two filings are
// separate motions with separate ids, and a ruling on one leaves the other open.
func TestTwoMotionsOnOneGradeAreTellableApart(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nA claim.\n")
	mintGap(t, runDir, "G1", "overclaim")

	for _, why := range []string{"the harm needs two failures", "and the second failure is gated upstream"} {
		if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", why); err != nil {
			t.Fatalf("file refused: %v", err)
		}
	}

	// Rule only M1. Under the old pair key this ruling would have matched BOTH filings — same
	// gap, same dimension — and the record would have shown one answer to two questions.
	if _, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "M1", "--as", "rejected", "--reason", "the grade stands"); err != nil {
		t.Fatalf("a properly addressed ruling was refused: %v", err)
	}

	ev := lastBody(t, runDir, &recordpb.MotionRule{})
	if got := ev.GetMotionId(); got != "M1" {
		t.Errorf("motion_id = %q, want M1 — the record must carry the join, not leave it to be inferred", got)
	}
	if got := ev.GetGrade(); got != recordpb.GradeRuling_GRADE_RULING_REJECTED {
		t.Errorf("ruling = %q, want rejected", recordpb.Word(got))
	}
}

// AN ANSWER TO A QUESTION NOBODY ASKED IS REFUSED — the dangling-reference rule this codebase
// applies to gap ids, applied to the other side of the conversation.
func TestARulingOnAnUnfiledMotionIsRefused(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nA claim.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "s"); err != nil {
		t.Fatalf("file refused: %v", err)
	}

	_, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r2",
		"--id", "M7", "--as", "accepted", "--reason", "sure")
	if err == nil {
		t.Fatal("red ruled a motion blue never filed — a reply with no question, joining to nothing")
	}
	if !strings.Contains(err.Error(), "which no filing created") {
		t.Errorf("the refusal must say the reference resolves to nothing: %v", err)
	}
}

// AN APPEAL AGAINST NO RULING IS REFUSED.
//
// The other half of the same discipline, and it did not exist before the collapse: `contests_ruling`
// was a field set as a side effect of a status move, so there was nothing to refuse. An appeal is
// pressing on AFTER an answer; against no answer the event would replay as a motion both unruled
// and appealed, which the report has no honest sentence for.
func TestAnAppealAgainstNoRulingIsRefused(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nA claim.\n")
	mintGap(t, runDir, "G1", "overclaim")
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "s"); err != nil {
		t.Fatalf("file refused: %v", err)
	}

	_, err := run(t, "motion", "grade", "appeal", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "M1", "--reason", "pressing it to the bench")
	if err == nil {
		t.Fatal("an appeal was recorded against a motion nobody had ruled — there is nothing to press against")
	}
	if !strings.Contains(err.Error(), "no ruling to appeal") {
		t.Errorf("the refusal must say WHY there is nothing to appeal: %v", err)
	}
}
