package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// A BOARD FOR PROBING WHETHER A SEAT USES THE TOOL AT ALL.
//
// # The question this exists to ask
//
// Every other harness here asks whether the tool ACCEPTS what a seat sends. This one is built for
// the opposite question, and it is the more expensive failure: **does a seat reach for the verb,
// or does it route around into prose and a hand-written file?**
//
// Nothing in the suite can answer that. The fuzz IS a driver — it calls the verb by construction,
// so it measures the tool and says nothing about whether a real seat would find it. The measured
// shape is in the friction logs of real runs: a seat that cannot find the verb it wants "logs
// friction and works around it, losing the capability for the run", and a seat that finds prose
// easier than a flag does the same thing without even noticing.
//
// So: build a board a seat would actually MEET, hand it to a weak model under its real
// constitution, and read back which verbs it used and which it talked its way around. The output
// is not pass/fail — it is a friction corpus and a list of verbs nobody chose. That belongs to
// the operator, not to CI (#363).
//
// # Why a weak model
//
// A strong seat compensates for a bad help string by inferring what was meant. That hides exactly
// the defect being hunted: the constitution and the `--help` text are the only things standing
// between a seat and a hand-written artifact, and their quality is only visible when the reader
// is not clever enough to paper over it. Haiku is the instrument, not the subject.
//
// # Why it is built through the CLI rather than the record API
//
// The fixture must be a board the real write paths would produce. Writing events directly would
// let this file record a state no seat could ever have reached — a gap with no acceptance check,
// a closure with no anchor — and a probe run against an impossible board teaches nothing about
// the possible one.
//
// # Running it
//
//	FEOV_SEAT_PROBE_DIR=<path> go test ./internal/cli -run TestWriteSeatProbeFixture -count=1
//
// Env-guarded because it writes outside the test's temp directory, which is the one thing a test
// must not do by default.
func TestWriteSeatProbeFixture(t *testing.T) {
	dest := os.Getenv("FEOV_SEAT_PROBE_DIR")
	if dest == "" {
		t.Skip("set FEOV_SEAT_PROBE_DIR to write a seat-probe board")
	}
	name := os.Getenv("FEOV_SEAT_PROBE_BOARD")
	if name == "" {
		name = "arithmetic"
	}
	board, ok := seatprobe.Boards()[name]
	if !ok {
		var have []string
		for n := range seatprobe.Boards() {
			have = append(have, n)
		}
		sort.Strings(have)
		t.Fatalf("no board %q — one of %s", name, strings.Join(have, ", "))
	}
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	buildBoard(t, dest, board)
	t.Logf("board %q written to %s for seat %s: %d gap(s), %d avenue(s), %d expectation(s)",
		board.Name, dest, board.Seat, len(board.Gaps), len(board.Avenues), len(board.Expect))
}

// buildBoard materialises a board THROUGH THE REAL WRITE PATHS.
//
// Writing the events directly would let this record a state no seat could ever have reached — a
// gap with no acceptance check, a closure with no anchor — and a probe run against an impossible
// board teaches nothing about the possible one.
func buildBoard(t *testing.T, runDir string, b seatprobe.Board) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(b.Report), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, s := range []struct{ role, id string }{
		{"lens", "red-lens-r1-L1"},
		{"merge", "red-merge-r1"},
		{"blue", "blue-respond-r1"},
		{"bench", "judge-r1"},
	} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatalf("register %s: %v", s.id, err)
		}
	}

	for i, g := range b.Gaps {
		existence := g.Existence
		if existence == "" {
			existence = "verified"
		}
		args := []string{"merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", g.Key, "--class", g.Class,
			"--location", g.Location, "--problem", g.Problem, "--fix", g.Fix,
			"--check", g.Check, "--check-kind", g.CheckKind,
			"--severity", g.Severity, "--likelihood", g.Likelihood,
			"--impact", g.Impact, "--cx", g.Complexity,
			"--existence", existence,
			"--reason", g.Problem + " (baits " + g.Baits + ": " + g.Why + ")"}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("mint %s: %v", g.Key, err)
		}
		if !g.Closed {
			continue
		}
		// A CLOSED gap so the archive is not empty. spot-check against an empty archive has
		// nothing to sample, so a board that wants the duty exercised has to give it something.
		id := fmt.Sprintf("R1-%d", i+1)
		if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", id, "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "git show",
			"--anchor-target", "HEAD:config", "--reason", "verified at the leaf against the pinned config"); err != nil {
			t.Fatalf("close %s: %v", id, err)
		}
	}

	for i, a := range b.Avenues {
		if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--line", a.Line, "--hypothesis", a.Hypothesis); err != nil {
			t.Fatalf("avenue %d: %v", i, err)
		}
		if a.Ruled == "" {
			continue
		}
		id := fmt.Sprintf("A%d", i+1)
		if _, err := run(t, "motion", "direction", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", id, "--as", a.Ruled,
			"--reason", "ruled "+a.Ruled+" on the line as it was proposed"); err != nil {
			t.Fatalf("rule %s: %v", id, err)
		}
	}

	for i, claim := range b.Claims {
		if _, err := run(t, "blue", "cite", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--key", fmt.Sprintf("C%d", i+1), "--location", claim,
			"--title", "the pinned source", "--url", "https://example.invalid/pinned",
			"--reason", "the source this claim rests on"); err != nil {
			// A cite that cannot reach its url is logged as friction by design; the board still
			// stands without the anchor, so this is reported rather than fatal.
			t.Logf("cite %d not recorded (%v) — the board is usable without it", i+1, err)
		}
	}
}
