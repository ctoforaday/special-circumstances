package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS, ACROSS SEATS, THROUGH THE COMMANDS EACH SEAT TYPES (gblock, 2026-10-30): the
// bench regrades a gap it is hearing and the minting lens may not move that axis; the bench rules the
// gap closed; blue's docket motion on it is refused without new grounds, and admitted on a proof it
// records since the ruling, named by its id.
func TestTheBenchSetsAGradeAndAReconsiderationComesOnNewGrounds(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "bench-orders", "spec_underspecification")
	motion := docketFile(t, runDir, "red-chair", g, "contested, and the bench's to settle")

	if _, err := runAt(t, "regrade", "--run", runDir, "--seat-id", "judge", "--id", g, "--severity", "high",
		"--reason", "the defect misleads every reader of the section"); err != nil {
		t.Fatalf("the bench's regrade of a gap it is hearing: %v", err)
	}
	_, err := runAt(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "low",
		"--reason", "the lens disagrees")
	if err == nil || !strings.Contains(err.Error(), "which the bench set at event") {
		t.Fatalf("the minting lens moved the axis the bench set: %v", err)
	}

	args := benchRuleArgs(motion, "defect_accepted", "proportionality")
	args = append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge"}, args[3:]...)
	if _, err := runAt(t, args...); err != nil {
		t.Fatalf("motion docket rule: %v", err)
	}

	_, err = runAt(t, "motion", "docket", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", g,
		"--reason", "the ruling was wrong")
	if err == nil || !strings.Contains(err.Error(), "admitted on NEW GROUNDS only") {
		t.Fatalf("a docket motion on a gap the bench ruled closed, with no grounds: %v", err)
	}

	_, err = runAt(t, "motion", "docket", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", g,
		"--quote", "the parser accepts an empty body in this line.", "--reason", "the report says otherwise")
	if err == nil || !strings.Contains(err.Error(), "is not new grounds on gap "+g) {
		t.Fatalf("a docket motion quoting text no edit answering the gap wrote since the ruling: %v", err)
	}

	s := script(t, runDir, "grounds.js", "console.log('what the ruling did not have');")
	if _, err := runAt(t, "prove", "--run", runDir, "--seat-id", "blue-respond", "--quote", "the parser accepts an empty body in this line.",
		"--script", s, "--answers", g, "--reason", "new grounds"); err != nil {
		t.Fatalf("prove: %v", err)
	}
	proof := lastBody(t, runDir, &recordpb.Proof{}).GetProofId()
	// THE GAP MOVED UNDER BLUE'S SITTING (the bench regraded and ruled it), so the first filing is
	// refused showing the moves and the same filing then lands.
	reconsider := func() error {
		_, err := runAt(t, "motion", "docket", "file", "--run", runDir, "--seat-id", "blue-respond", "--id", g,
			"--grounds", proof, "--reason", "the proof was recorded since the ruling")
		return err
	}
	if err := reconsider(); err != nil && !strings.Contains(err.Error(), "changed after your sitting opened") {
		t.Fatalf("a docket motion on new grounds %s was refused: %v", proof, err)
	} else if err != nil {
		if err := reconsider(); err != nil {
			t.Fatalf("a docket motion on new grounds %s was refused: %v", proof, err)
		}
	}
	if got := lastBody(t, runDir, &recordpb.Motion{}).GetDocket().GetGrounds(); got != proof {
		t.Errorf("the motion records grounds %q, want %q", got, proof)
	}
}
