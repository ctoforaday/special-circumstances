package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ONE MECHANISM, ONE ID — the thing three verb pairs could not do.
//
// #312 is the shape this closes: `petition-rule` joined on (petitioner, class) with no id, so two
// filings by one seat in one round were indistinguishable, and #320 had to render filings and
// rulings SIDE BY SIDE rather than joined because pairing them would have been a guess.
func TestAMotionJoinsItsAskToItsAnswerOnAnID(t *testing.T) {
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "motion-join", "motion-test")

	out, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", gap, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation")
	if err != nil {
		t.Fatalf("file: %v", err)
	}
	if !strings.Contains(out, "M1") {
		t.Fatalf("the tool must assign the id, got %q", out)
	}
	if _, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "M1", "--as", "rejected", "--reason", "the bound does not hold across the retry path"); err != nil {
		t.Fatalf("rule: %v", err)
	}
	if _, err := run(t, "motion", "grade", "appeal", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "M1", "--reason", "the retry path is guarded upstream; taking it to the bench"); err != nil {
		t.Fatalf("appeal: %v", err)
	}

	b, err := record.BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	ms := record.AllMotions(b)
	if len(ms) != 1 {
		t.Fatalf("one exchange, one motion; got %d", len(ms))
	}
	m := ms[0]
	if !m.Ruled() || m.Ruling != "rejected" {
		t.Errorf("the ask lost its answer: %+v", m)
	}
	if !m.Appealed {
		t.Error("the appeal was lost — a ruling is an argument, not a command, and pressing on is the substance")
	}
	if m.Fields["dimension"] != "severity" || m.Fields["gap_id"] != gap {
		t.Errorf("subject payload lost: %+v", m.Fields)
	}
	// The motion's OWN id and the gap it references must not collide in one payload.
	if m.ID == m.Fields["gap_id"] {
		t.Error("the motion id and the gap id are the same value — two identities in one key")
	}
}

// THE RULER IS PART OF THE MECHANISM, and the refusal NAMES who holds the gavel.
//
// Under a scoped surface "not yours" would otherwise read as "does not exist" — the measured
// failure where a seat handed an unavailable verb logs friction and works around it, losing the
// capability for the run.
func TestOnlyTheRulingSeatMayRule(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-merge-r1",
		"--petition-class", "integrity", "--relief", "strike the demand",
		"--reason", "the instruction would require asserting what I believe false"); err != nil {
		t.Fatal(err)
	}
	// A petition is the BENCH's to rule. The merge filing it may not also decide it.
	_, err := run(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "M1", "--as", "granted", "--reason", "granting my own petition")
	if err == nil {
		t.Fatal("the merge seat ruled a petition; a motion is filed by any seat and ruled by ONE, and that asymmetry is the mechanism")
	}
	for _, want := range []string{"bench", "red-merge-r1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the seat that holds the gavel AND the one that asked; missing %q in %v", want, err)
		}
	}
	if _, err := run(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "judge-r1",
		"--id", "M1", "--as", "granted", "--reason", "the objection is sound"); err != nil {
		t.Fatalf("the bench must be able to rule a petition: %v", err)
	}
}

// A PETITION HAS NO APPEAL, expressed by ABSENCE rather than a runtime refusal: it is heard
// before the debate continues, so there is nothing to escalate to.
func TestAPetitionHasNoAppealVerb(t *testing.T) {
	if h := help(t, "motion", "petition", "--help"); strings.Contains(h, "appeal") {
		t.Error("a petition grew an appeal verb; it is heard BEFORE the debate continues, so there is nothing to appeal to")
	}
	if h := help(t, "motion", "grade", "--help"); !strings.Contains(h, "appeal") {
		t.Error("a grade motion must be appealable — a rejected dispute goes to the bench")
	}
}

// A RULING NAMING NO FILING IS REFUSED — the dangling reference every other cross-reference gets
// checked for, because it is accepted at write time and dropped at replay where nobody sees it.
func TestARulingMustNameAMotionThatExists(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "motion", "grade", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", "M99", "--as", "accepted", "--reason", "ruling on nothing"); err == nil {
		t.Fatal("a ruling naming a motion no filing created was accepted; it would be dropped at replay in silence")
	}
}
