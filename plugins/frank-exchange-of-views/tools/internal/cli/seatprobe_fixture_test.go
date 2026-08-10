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

	// MOTIONS THE BOARD ALREADY CARRIES: an ask a seat must answer, or an answered one it may
	// press. A board that expects a ruling and files no motion made the expectation unreachable,
	// and the probe reported it as a miss BY THE SEAT — four times, before the reachability gate
	// existed.
	for i, m := range b.Motions {
		args := []string{"motion", m.Subject, "file", "--run", runDir, "--seat-id", m.Filer,
			"--reason", m.Basis}
		switch m.Subject {
		case "grade":
			args = append(args, "--id", m.GapID, "--dimension", m.Dimension, "--proposed", m.Proposed)
		case "petition":
			args = append(args, "--petition-class", m.Class, "--relief", m.Relief)
		}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("motion %d (%s): %v", i+1, m.Subject, err)
		}
		if m.Ruled == "" {
			continue
		}
		ruler := map[string]string{"grade": "red-merge-r1", "petition": "judge-r1"}[m.Subject]
		if _, err := run(t, "motion", m.Subject, "rule", "--run", runDir, "--seat-id", ruler,
			"--id", fmt.Sprintf("M%d", i+1), "--as", m.Ruled,
			"--reason", "ruled "+m.Ruled+" on the filing as it stands"); err != nil {
			t.Fatalf("rule motion %d: %v", i+1, err)
		}
	}

	// PROOFS, so a lens has something to RE-RUN rather than only something to read. The tool
	// executes the script twice and records whether it reproduced, which is the whole point of
	// the verb the board is about to demand.
	for i, pr := range b.Proofs {
		script := filepath.Join(runDir, fmt.Sprintf("probe-proof-%d.py", i+1))
		if err := os.WriteFile(script, []byte(pr.Script+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		args := []string{"blue", "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--location", pr.Location, "--script", script,
			"--reason", "the computation behind this sentence"}
		if pr.Answers != "" {
			args = append(args, "--answers", pr.Answers)
		}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("prove %d: %v", i+1, err)
		}
	}

	for i, claim := range b.Claims {
		// A CITE THAT DOES NOT LAND MAKES A HOLLOW BOARD, so this is fatal.
		//
		// The first draft logged it and carried on, and the `sources` and `lens-audit` boards
		// built with ZERO cited claims — which is precisely the state their expectations are
		// about. `lens verify` has nothing to verify and `blue claim-index` has nothing to
		// index, so both would have reported UNMET against a seat that had no way to meet them,
		// and the report would have read as a finding about the seat rather than about the
		// fixture. A builder that degrades quietly produces exactly the plausible zero the rest
		// of this suite exists to remove.
		//
		// The url is a real, reachable one because `cite` FETCHES and caches: an unreachable
		// source is refused and logged as friction, which is correct behaviour and useless here.
		if _, err := run(t, "blue", "cite", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--key", fmt.Sprintf("C%d", i+1), "--location", claim,
			"--title", "the pinned source", "--url", "https://example.com/",
			"--reason", "the source this claim rests on"); err != nil {
			t.Fatalf("cite %d (%q) did not land: %v\n\nThe board declares this claim and its expectations are about acting on it. Building without it would produce a board whose demands cannot be met, and a report that blames the seat for the fixture.", i+1, claim, err)
		}
	}
}
