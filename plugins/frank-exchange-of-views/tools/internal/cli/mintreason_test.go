package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// RED'S ARGUMENT FOR A GAP MUST REACH THE SEATS THAT ANSWER AND WEIGH IT.
//
// `mint` took --reason and threw it away whenever --problem was also given, which is every ordinary
// mint. The write still answered "minted G1", and --reason's own help promises "the substance the
// report renders and the other side answers" — a promise this verb did not keep.
//
// MEASURED 2026-08-16 BY ASKING THE BENCH rather than by watching one. Dispatched to a petition
// sitting about what a required_fix may demand, it listed first among the things it could not find:
// "Red's closing or reasoning — Why did red mint this gap? … The gap's problem statement is clear,
// but red's intent for the fix isn't visible to me." It was being asked to adjudicate red's intent
// with red's intent unrecorded.
func TestAGapCarriesRedsArgumentAndNotOnlyItsProblem(t *testing.T) {
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	board := seatprobe.Boards()["audit"]
	exec := func(args ...string) (string, error) { return run(t, args...) }
	if err := seatprobe.Build(runtest.Open(t, runDir), board, exec); err != nil {
		t.Fatal(err)
	}

	const why = "blue may well have done the work; what is missing is the record of it"
	registerLensOnce(t, runDir)
	if _, err := runMint(t, runDir, "--run", runDir, "--seat-id", lensSeat,
		"--key", "argued", "--class", "metric-conflation",
		"--quote", "The access log retains for 45 days.",
		"--problem", "the figure conflicts with the universal",
		"--fix", "reconcile them", "--check", "no two sections disagree", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", why); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, why) {
		t.Errorf("a gap minted WITH --reason does not carry it to the board.\n\n"+
			"--reason's help promises \"the substance the report renders and the other side answers\"; if the verb "+
			"discards it whenever --problem is present, that promise is false for every ordinary mint and the write "+
			"still reports success.\n\nboard was:\n%s", out)
	}

	// --REASON IS NEVER THE PROBLEM. A mint that gives the argument and no --problem is refused, in
	// the problem's own words: the text it would have stored was red's argument, and filing it as
	// the problem — which the report prints — puts red's reasoning in the risk matrix.
	_, err = runMint(t, runDir, "--run", runDir, "--seat-id", lensSeat,
		"--key", "viareason", "--class", "metric-conflation",
		"--quote", "The audit log retains for 30 days.",
		"--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", "the problem arrived through reason")
	if err == nil || !strings.Contains(err.Error(), "--problem") {
		t.Fatalf("a mint with --reason and no --problem = %v, want a refusal naming --problem", err)
	}
	out, err = run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "the problem arrived through reason") {
		t.Errorf("the refused mint reached the board:\n%s", out)
	}
}
