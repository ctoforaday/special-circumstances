package difftest

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/refusalcheck"
)

// EVERY REFUSAL THE BINARY PRINTS IN THIS PACKAGE NAMES ITS COMMAND EXACTLY ONCE.
//
// `capture` is the one place every invocation of the built binary passes through — the scenarios,
// the error catalogue, the help contracts, the determinism fuzz — so the reading is taken there
// and not per test: a scenario added tomorrow is walked without being listed anywhere. Each
// refusal is resolved to its command from its own argv, on the tree of the seat that ran it, and
// its line is counted by internal/refusalcheck: the renderer's naming, plus any the diagnosis
// adds. A count of 0 or of 2 and more fails the package when the suite ends.
var refusalsWalked struct {
	sync.Mutex
	human, json int
	wrong       []string
}

// noteRefusal counts one invocation's refusal, in whichever form it was written.
func noteRefusal(argv []string, inv invocation) {
	if inv.code != 2 {
		return
	}
	seatID := ""
	for i, a := range argv {
		switch {
		case a == "--seat-id" && i+1 < len(argv):
			seatID = argv[i+1]
		case strings.HasPrefix(a, "--seat-id="):
			seatID = strings.TrimPrefix(a, "--seat-id=")
		}
	}
	roles := make([]string, 0, 5)
	for role := range cli.AllRoots() {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	root := cli.NewRootFor(seatID)

	refusalsWalked.Lock()
	defer refusalsWalked.Unlock()
	if env, ok := feov.DecodeRefusal(strings.TrimSpace(inv.stdout)); ok {
		refusalsWalked.json++
		path, _, _ := refusalcheck.Rendered(root, roles, argv, "")
		if again := refusalcheck.Again(path, roles, env.Error); len(again) > 0 {
			refusalsWalked.wrong = append(refusalsWalked.wrong, fmt.Sprintf("--json, %d times (%s): %s\n      %.300s",
				1+len(again), strings.Join(again, ", "), strings.Join(argv, " "), env.Error))
		}
		return
	}
	if !strings.HasPrefix(inv.stderr, refusalcheck.Tool+": ") {
		return
	}
	refusalsWalked.human++
	if _, n, forms := refusalcheck.Rendered(root, roles, argv, inv.stderr); n != 1 {
		refusalsWalked.wrong = append(refusalsWalked.wrong, fmt.Sprintf("%d times (%s): %s\n      %.300s",
			n, strings.Join(forms, ", "), strings.Join(argv, " "), inv.stderr))
	}
}

// refusalsNameTheirCommandOnce is the package's verdict on every refusal `capture` saw.
func refusalsNameTheirCommandOnce() error {
	refusalsWalked.Lock()
	defer refusalsWalked.Unlock()
	fmt.Fprintf(os.Stderr, "difftest: %d refusals walked (%d human, %d --json), %d naming their command other than once\n",
		refusalsWalked.human+refusalsWalked.json, refusalsWalked.human, refusalsWalked.json, len(refusalsWalked.wrong))
	if len(refusalsWalked.wrong) == 0 {
		return nil
	}
	return fmt.Errorf("FAIL: %d refusal(s) the binary printed name their command other than exactly once — "+
		"the renderer prints the path from the command that resolved, and the diagnosis starts at what is wrong:\n  - %s",
		len(refusalsWalked.wrong), strings.Join(refusalsWalked.wrong, "\n  - "))
}

// walkedSoFar is how many refusals `capture` has read, for a test that asserts it was walked.
func walkedSoFar() int {
	refusalsWalked.Lock()
	defer refusalsWalked.Unlock()
	return refusalsWalked.human + refusalsWalked.json
}
