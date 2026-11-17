package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
	gap := mintGap(t, runDir, "motion-join", "spec_underspecification")

	exchangeOn(t, runDir, gap)
	out, err := run(t, closeProposalArgs(runDir, "blue-respond", gap, "not_a_defect",
		"the consequence is bounded by the caller's own validation; the bench should decide the gap")...)
	if err != nil {
		t.Fatalf("file: %v", err)
	}
	if !strings.Contains(out, handleText(t, runDir, "M1")) {
		t.Fatalf("the tool must assign the id, got %q", out)
	}
	conveneBench(t, runDir, gap)
	rule := benchRuleArgs("M1", "remanded", "correctness")
	if _, err := runAt(t, append([]string{rule[0], rule[1], rule[2], "--run", runDir, "--seat-id", "judge"}, rule[3:]...)...); err != nil {
		t.Fatalf("rule: %v", err)
	}

	b, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	ms := record.MotionsOf(b.Events, b.At)
	if len(ms) != 1 {
		t.Fatalf("one exchange, one motion; got %d", len(ms))
	}
	m := ms[0]
	if !m.Ruled() || m.Ruling != "remanded" {
		t.Errorf("the ask lost its answer: %+v", m)
	}
	if m.Fields["gap_id"] != gap {
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
// failure where a seat handed an unavailable verb logs a request and works around it, losing the
// capability for the run.
func TestOnlyTheRulingSeatMayRule(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-chair",
		"--class", "integrity", "--relief", "strike the demand",
		"--reason", "the instruction would require asserting what I believe false"); err != nil {
		t.Fatal(err)
	}
	// A petition is the BENCH's to rule. The chair filing it may not also decide it.
	_, err := runAt(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "red-chair",
		"--id", "M1", "--as", "granted", "--relief", "Restate the section as the bench orders.", "--binds", "blue", "--reason", "granting my own petition")
	if err == nil {
		t.Fatal("the chair ruled a petition; a motion is ruled by ONE seat, and that asymmetry is the mechanism")
	}
	for _, want := range []string{"bench", "red-chair"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the seat that holds the gavel AND the one that asked; missing %q in %v", want, err)
		}
	}
	if _, err := runAt(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "judge",
		"--id", "M1", "--as", "granted", "--relief", "Restate the section as the bench orders.", "--binds", "blue", "--reason", "the objection is sound"); err != nil {
		t.Fatalf("the bench must be able to rule a petition: %v", err)
	}
}

// A RULING NAMING NO FILING IS REFUSED — the dangling reference every other cross-reference gets
// checked for, because it is accepted at write time and dropped at replay where nobody sees it.
func TestARulingMustNameAMotionThatExists(t *testing.T) {
	runDir := seatRun(t)
	if _, err := runAt(t, "motion", "petition", "rule", "--run", runDir, "--seat-id", "judge",
		"--id", "M99", "--as", "denied", "--reason", "ruling on nothing"); err == nil {
		t.Fatal("a ruling naming a motion no filing created was accepted; it would be dropped at replay in silence")
	}
}
