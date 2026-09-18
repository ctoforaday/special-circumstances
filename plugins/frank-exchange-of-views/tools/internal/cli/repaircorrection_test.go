package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// blueDispatchedOnG1 puts the chair's dispatch row on the record so blue's next register opens a
// sitting for it — which is what makes that sitting repairable.
func blueDispatchedOnG1(t *testing.T, runDir string) {
	t.Helper()
	recordtest.Seed(t, runDir, recordtest.Event(t, "red-chair",
		&recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}))
}

// A CORRECTION IN A REPAIR IS STILL THE SAME SITTING (#1002, gblock 2026-09-18). "While it is still
// yours alone" is a question about the sitting an act BELONGS TO, and a sitting-record repair's acts
// are the repaired sitting's — so the seat re-prompted to finish sitting 1's record may correct what
// sitting 1 wrote. Counting the repair as a turn refuses it with "that was an earlier sitting" at
// the one moment the seat is back on the record precisely to put that sitting right.
func TestACorrectionInARepairIsTheSameSitting(t *testing.T) {
	runDir := corrFixture(t)
	blueDispatchedOnG1(t, runDir)
	must(t, runDir, "register", "--seat-id", "blue-respond")
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"manifest-row", "--id", "G1", "--reason", "recomputed the  figure"})

	must(t, runDir, "register", "--seat-id", "blue-respond", "--repair-sitting")
	if _, err := run(t, "manifest-row", "--run", runDir, "--seat-id", "blue-respond", "--id", "G1",
		"--reason", "recomputed the figure", "--corrects", k, "--correction-why", "the shell deleted a word"); err != nil {
		t.Fatalf("a correction filed in the repair of the sitting that wrote the act was refused: %v", err)
	}
	if c := lastBody(t, runDir, &recordpb.Correction{}); c.GetCorrects() != k {
		t.Errorf("the correction on the record = %v, want it striking %s", c, k)
	}
}

// THE CONTROL, and it is the rule this one bends around: a seat DISPATCHED AGAIN is in a new
// sitting, and the act it wrote in the last one is answered rather than corrected.
func TestACorrectionAcrossARealSecondSittingIsRefused(t *testing.T) {
	runDir := corrFixture(t)
	blueDispatchedOnG1(t, runDir)
	must(t, runDir, "register", "--seat-id", "blue-respond")
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"manifest-row", "--id", "G1", "--reason", "recomputed the  figure"})

	must(t, runDir, "register", "--seat-id", "blue-respond")
	_, err := run(t, "manifest-row", "--run", runDir, "--seat-id", "blue-respond", "--id", "G1",
		"--reason", "recomputed the figure", "--corrects", k, "--correction-why", "the shell deleted a word")
	if err == nil || !strings.Contains(err.Error(), "was written in an earlier sitting of blue-respond") {
		t.Fatalf("a correction across a real second sitting = %v, want the earlier-sitting refusal", err)
	}
}

// THE OFFER READS THE SAME SITTING THE SAME WAY. A first-wins refusal offers the correction only
// when the act it found is the asker's own AND from the sitting it is in now — the second reader of
// "same sitting", and it follows the repair for the same reason the gate does.
func TestTheFirstWinsOfferSurvivesARepair(t *testing.T) {
	runDir := corrFixture(t)
	m := fileGradeMotion(t, runDir)
	must(t, runDir, "motion", "grade", "rule", "--seat-id", "red-chair", "--id", m, "--as", "rejected", "--reason", "the evidence does not reach it")
	blueDispatchedOnG1(t, runDir)
	must(t, runDir, "register", "--seat-id", "blue-respond")
	must(t, runDir, "motion", "grade", "appeal", "--seat-id", "blue-respond", "--id", m, "--reason", "the consequence is not bounded")

	must(t, runDir, "register", "--seat-id", "blue-respond", "--repair-sitting")
	_, err := run(t, "motion", "grade", "appeal", "--run", runDir, "--seat-id", "blue-respond", "--id", m, "--reason", "a second try")
	if err == nil || !strings.Contains(err.Error(), "It is your own appeal from this sitting") {
		t.Fatalf("the second appeal's refusal = %v, want the offer to correct the appeal this sitting filed", err)
	}
}

// AN ACT FILED INSIDE THE REPAIR IS CORRECTED INSIDE IT TOO (#1026). Both arms of the shared SQL
// constant have to skip the repair register, and only the SECOND was held: every test corrected an
// act written BEFORE the repair, so the `before` subquery's filter could be dropped outright with
// the suite green. Here the act is written inside the repair — `before` counts the registers ahead
// of it, the repair's included — and an unfiltered count makes it sitting 2 against a current
// sitting 1, refusing the correction of an act the seat wrote moments earlier.
func TestAnActFiledInsideTheRepairIsCorrectedInsideIt(t *testing.T) {
	runDir := corrFixture(t)
	blueDispatchedOnG1(t, runDir)
	must(t, runDir, "register", "--seat-id", "blue-respond")
	must(t, runDir, "register", "--seat-id", "blue-respond", "--repair-sitting")

	k := correctionKeyOf(t, runDir, "blue-respond", []string{"manifest-row", "--id", "G1", "--reason", "recomputed the  figure"})
	if _, err := run(t, "manifest-row", "--run", runDir, "--seat-id", "blue-respond", "--id", "G1",
		"--reason", "recomputed the figure", "--corrects", k, "--correction-why", "the shell deleted a word"); err != nil {
		t.Fatalf("an act filed inside the repair was refused its own correction: %v", err)
	}
	if c := lastBody(t, runDir, &recordpb.Correction{}); c.GetCorrects() != k {
		t.Errorf("the correction on the record = %v, want it striking %s", c, k)
	}
}
