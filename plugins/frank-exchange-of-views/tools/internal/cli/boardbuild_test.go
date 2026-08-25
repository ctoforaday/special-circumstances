package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// EVERY PROBE BOARD MUST STILL BUILD THROUGH THE REAL WRITE PATHS.
//
// MEASURED 2026-08-15, and it is why this exists: EIGHT OF NINE boards were unbuildable on a green
// HEAD. `merge mint --location` was tightened (#359) to refuse a span that is not literally in
// report.md — "a location a reader cannot find is not a location" — and the boards went on writing
// the old prose form, `## Findings — "the sentence"`. Every dispatch died in Build before a seat
// was ever handed the board.
//
// Nothing reported it for one reason: the only thing that built a board was
// TestWriteSeatProbeFixture, which is env-guarded (FEOV_SEAT_PROBE_DIR) and skips by default, and
// `cmd/seatprobe` itself, which is run by hand. So the fixtures rotted behind a full green sweep —
// the write path moved, the carrier kept speaking the old model, and the half-state read as done.
//
// This is that gate, and it is deliberately NOT env-guarded. It costs one temp directory per board
// and it fails on the day a write path moves rather than on the day someone next runs the probe by
// hand and reads the error as their own mistake.
//
// It builds and stops. Dispatching a seat needs a model and a network, which is the operator's
// business (#363: this package reports, it does not gate); whether the STATE those expectations are
// written against can still be reached is a question about this repository's own code, and that
// belongs here.
func TestEveryProbeBoardStillBuilds(t *testing.T) {
	boards := seatprobe.Boards()
	if len(boards) == 0 {
		t.Fatal("no probe boards at all — an empty set builds cleanly forever and reports nothing")
	}
	for name, b := range boards {
		b := b
		t.Run(name, func(t *testing.T) {
			// A BARE DIRECTORY, ON PURPOSE — not newRun(t). This gate exists to prove the probe's
			// boards still build, so it must give Build exactly what the probe gives it: nothing.
			// Staging a class registry here made the gate green on the day every real dispatch
			// died on its first mint, because the fixture was configured a way the probe never is.
			runDir := recordtest.TmpRun(t)
			t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
			// seatprobe.Build, not a second copy of it. The harness dispatches through that
			// function; a gate that rebuilt the steps here would be asserting its own reading of
			// the build rather than the build the probe actually runs.
			exec := func(args ...string) (string, error) { return run(t, args...) }
			if err := seatprobe.Build(runDir, b, exec); err != nil {
				t.Fatalf("board %q no longer builds: %v\n\n"+
					"The board is a fixture for a write path that has since moved. Until it builds, every\n"+
					"dispatch against it dies before a seat sees it — and the probe reports a FAILED board,\n"+
					"which reads as a harness fault rather than as this fixture speaking a retired model.",
					name, err)
			}
		})
	}
}
