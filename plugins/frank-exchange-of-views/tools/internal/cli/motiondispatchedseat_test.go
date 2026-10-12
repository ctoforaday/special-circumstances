package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// A MOTION IS FILED BY A SEAT THE ENGINE CREATED.
//
// A motion's command parent is its SUBJECT ("grade"), which names no role, so there is no role to
// hold the filer to — and the filer is still held to being a seat: seat.Begin refuses an id the
// roster gives no role (record.RequireDispatchedSeat), and
// TestEveryWritingVerbRefusesWhatBeginRefuses holds every writing verb to Begin. A motion under an
// invented identity would be rendered in the report, ruled on, and joined to a gap.
//
// Every seat may FILE any motion; that asymmetry is the mechanism and this does not touch it. Being
// a seat at all is not the optional half.
func TestAMotionRequiresASeatTheEngineCreated(t *testing.T) {
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	board := seatprobe.Boards()["docket"]
	exec := func(args ...string) (string, error) { return run(t, args...) }
	if err := seatprobe.Build(runtest.Open(t, runDir), board, exec); err != nil {
		t.Fatal(err)
	}

	_, err := runAt(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "totally-invented",
		"--id", "G1", "--dimension", "severity", "--proposed", "low", "--reason", "an invented identity")
	if err == nil {
		t.Fatal("a motion filed under an invented seat id was ACCEPTED — it is rendered in the report, ruled on, and joined to a gap, under an identity no dispatch created")
	}
	if !strings.Contains(err.Error(), "role namespace") {
		t.Errorf("refused for the wrong reason: %v", err)
	}

	// AND THE LEGITIMATE PATH IS UNTOUCHED. A guard that also refused real filings would be
	// caught by the suite, but the point of the asymmetry is that ANY seat may file — so blue
	// filing a grade motion must still work, and that is the half worth asserting here.
	if _, err := runAt(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond",
		"--id", "G1", "--dimension", "severity", "--proposed", "low", "--reason", "a real filing"); err != nil {
		t.Fatalf("blue could not file a grade motion: %v", err)
	}
}
