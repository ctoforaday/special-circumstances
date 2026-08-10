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
	board := ArithmeticBoard()
	out, err := Report(runDir, []string{"blue-respond-r1", "red-merge-r1", "red-lens-r1-L1", "judge-r1"}, board.Expect)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}
