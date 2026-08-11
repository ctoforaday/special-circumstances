package seatprobe

import (
	"fmt"
	"os"
	"testing"
)

// THE OPERATOR REPORT. Not a gate — agent behaviour is not deterministic, and a flaky gate is one
// the next person turns off. Run it against a probe run a real seat was dispatched into:
//
//	FEOV_PROBE_RUN=<runDir> go test ./internal/seatprobe -run TestSeatChoiceReport -v
func TestSeatChoiceReport(t *testing.T) {
	runDir := os.Getenv("FEOV_PROBE_RUN")
	if runDir == "" {
		t.Skip("set FEOV_PROBE_RUN to a probe run directory")
	}
	name := os.Getenv("FEOV_PROBE_BOARD")
	if name == "" {
		name = "arithmetic"
	}
	board, ok := Boards()[name]
	if !ok {
		t.Fatalf("no board %q — set FEOV_PROBE_BOARD to one of the declared boards", name)
	}
	out, err := Report(surface(), runDir, []string{board.Seat}, board.Expect, nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}
