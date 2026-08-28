package cli

import (
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// RED'S ARGUMENT FOR A GAP MUST REACH THE SEATS THAT ANSWER AND WEIGH IT.
//
// `mint` took --reason and threw it away whenever --problem was also given, which is every ordinary
// mint. The write still answered "minted R1-4", and --reason's own help promises "the substance the
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
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "argued", "--class", "metric-conflation",
		"--quote", "The access log retains for 45 days.",
		"--problem", "the figure conflicts with the universal",
		"--fix", "reconcile them", "--check", "no two sections disagree", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", why); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-merge-r1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, why) {
		t.Errorf("a gap minted WITH --reason does not carry it to the board.\n\n"+
			"--reason's help promises \"the substance the report renders and the other side answers\"; if the verb "+
			"discards it whenever --problem is present, that promise is false for every ordinary mint and the write "+
			"still reports success.\n\nboard was:\n%s", out)
	}

	// AND IT IS NOT DUPLICATED when the problem arrived THROUGH --reason, which is the documented
	// alternative form (`--problem "..."|--reason`). Storing it twice would make a reader think red
	// argued something beyond the problem statement when it did not.
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "viareason", "--class", "metric-conflation",
		"--quote", "The audit log retains for 30 days.",
		"--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", "the problem arrived through reason"); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "show", "board", "--run", runDir, "--seat-id", "red-merge-r1")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Open []struct {
			Problem    string `json:"problem"`
			MintReason string `json:"mint_reason"`
		} `json:"open"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("board did not parse: %v", err)
	}
	for _, g := range got.Open {
		if g.MintReason != "" && g.MintReason == g.Problem {
			t.Errorf("a gap carries its problem twice — once as problem and once as mint_reason: %q", g.Problem)
		}
	}
}
