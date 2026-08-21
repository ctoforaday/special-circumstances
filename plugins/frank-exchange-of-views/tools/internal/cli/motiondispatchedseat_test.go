package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// A MOTION IS FILED BY A SEAT THE ENGINE CREATED, AND NOTHING CHECKED THAT.
//
// The role guard lives in seat.Begin. The motion verbs read their context with seat.Of, which only
// parses flags — Begin is not on that path at all — and record.CheckSeatRole took the command's
// parent as the role, which for `motion grade file` is the SUBJECT "grade". No role table has that
// key, so the guard returned nil and the filing landed.
//
// Measured on a real board before the fix:
//
//	blue position      --seat-id totally-invented   REFUSED
//	motion grade file  --seat-id totally-invented   "motion M2 filed (grade)"
//
// Every seat may FILE any motion; that asymmetry is the mechanism and this does not touch it. Being
// a seat at all is not the optional half.
func TestAMotionRequiresASeatTheEngineCreated(t *testing.T) {
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	board := seatprobe.Boards()["docket"]
	exec := func(args ...string) (string, error) { return run(t, args...) }
	if err := seatprobe.Build(runDir, board, exec); err != nil {
		t.Fatal(err)
	}

	_, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "totally-invented",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "an invented identity")
	if err == nil {
		t.Fatal("a motion filed under an invented seat id was ACCEPTED — it is rendered in the report, ruled on, and joined to a gap, under an identity no dispatch created")
	}
	if !strings.Contains(err.Error(), "role namespace") {
		t.Errorf("refused for the wrong reason: %v", err)
	}

	// AND THE LEGITIMATE PATH IS UNTOUCHED. A guard that also refused real filings would be
	// caught by the suite, but the point of the asymmetry is that ANY seat may file — so blue
	// filing a grade motion must still work, and that is the half worth asserting here.
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", "R1-1", "--dimension", "severity", "--proposed", "low", "--reason", "a real filing"); err != nil {
		t.Fatalf("blue could not file a grade motion: %v", err)
	}
}
