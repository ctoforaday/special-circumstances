package seatprobe

import (
	"fmt"
	"os"
	"path/filepath"
)

// BUILDING A BOARD, THROUGH WHATEVER RUNS THE TOOL.
//
// The caller supplies Exec — the test drives the in-process cobra root, the harness shells out to
// the real binary — and everything else is shared. That split is the point: a board built one way
// for the test and another way for the probe is two fixtures with one name, and the probe's is the
// one that matters.
//
// EVERY STEP GOES THROUGH THE REAL VERBS. Writing events directly would let this record a state no
// seat could ever have reached — a gap with no acceptance check, a closure with no anchor, a
// ruling on a motion nobody filed — and a probe run against an impossible board teaches nothing
// about the possible one.
//
// EVERY FAILURE IS FATAL, and that is not fastidiousness. The first draft logged a failed `cite`
// and carried on, and two boards built with ZERO cited claims — which is precisely the state their
// expectations are about. `lens verify` had nothing to verify and `blue claim-index` had nothing
// to index, so both would have reported UNMET against a seat with no way to meet them, and the
// report would have read as a finding about the SEAT rather than about the fixture.

// Exec runs one tool invocation and returns its stdout.
type Exec func(args ...string) (string, error)

// Seats are registered on every board: a motion names its filer, a ruling names its ruler, and an
// unregistered seat is refused before any board state exists.
var Seats = []struct{ Role, ID string }{
	{"lens", "red-lens-r1-L1"},
	{"merge", "red-merge-r1"},
	{"blue", "blue-respond-r1"},
	{"bench", "judge-r1"},
}

// Build materialises a board into runDir.
func Build(runDir string, b Board, exec Exec) error {
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(b.Report), 0o644); err != nil {
		return err
	}
	for _, s := range Seats {
		if _, err := exec(s.Role, "register", "--run", runDir, "--seat-id", s.ID); err != nil {
			return fmt.Errorf("register %s: %w", s.ID, err)
		}
	}

	for i, g := range b.Gaps {
		if _, err := exec("merge", "mint", "--run", runDir, "--seat-id", "red-merge-r1",
			"--key", g.Key, "--class", g.Class,
			"--location", g.Location, "--problem", g.Problem, "--fix", g.Fix,
			"--check", g.Check, "--check-kind", g.CheckKind,
			"--severity", g.Severity, "--likelihood", g.Likelihood,
			"--impact", g.Impact, "--cx", g.Complexity,
			"--reason", g.Problem+" (baits "+g.Baits+": "+g.Why+")"); err != nil {
			return fmt.Errorf("mint %s: %w", g.Key, err)
		}
		if !g.Closed {
			continue
		}
		// A CLOSED gap so the archive is not empty: `spot-check` against an empty one has nothing
		// to sample, so a board that wants the duty exercised has to give it something.
		id := fmt.Sprintf("R1-%d", i+1)
		if _, err := exec("merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", id, "--as", "closed", "--anchor-seat", "L1", "--anchor-tool", "git show",
			"--anchor-target", "HEAD:config", "--reason", "verified at the leaf against the pinned config"); err != nil {
			return fmt.Errorf("close %s: %w", id, err)
		}
	}

	for i, a := range b.Avenues {
		if _, err := exec("blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--line", a.Line, "--hypothesis", a.Hypothesis); err != nil {
			return fmt.Errorf("avenue %d: %w", i+1, err)
		}
		if a.Ruled == "" {
			continue
		}
		id := fmt.Sprintf("A%d", i+1)
		if _, err := exec("motion", "direction", "rule", "--run", runDir, "--seat-id", "red-merge-r1",
			"--id", id, "--as", a.Ruled,
			"--reason", "ruled "+a.Ruled+" on the line as it was proposed"); err != nil {
			return fmt.Errorf("rule %s: %w", id, err)
		}
	}

	for i, m := range b.Motions {
		args := []string{"motion", m.Subject, "file", "--run", runDir, "--seat-id", m.Filer,
			"--reason", m.Basis}
		switch m.Subject {
		case "grade":
			args = append(args, "--id", m.GapID, "--dimension", m.Dimension, "--proposed", m.Proposed)
		case "petition":
			args = append(args, "--petition-class", m.Class, "--relief", m.Relief)
		}
		if _, err := exec(args...); err != nil {
			return fmt.Errorf("motion %d (%s): %w", i+1, m.Subject, err)
		}
		if m.Ruled == "" {
			continue
		}
		ruler := map[string]string{"grade": "red-merge-r1", "petition": "judge-r1"}[m.Subject]
		if _, err := exec("motion", m.Subject, "rule", "--run", runDir, "--seat-id", ruler,
			"--id", fmt.Sprintf("M%d", i+1), "--as", m.Ruled,
			"--reason", "ruled "+m.Ruled+" on the filing as it stands"); err != nil {
			return fmt.Errorf("rule motion %d: %w", i+1, err)
		}
	}

	for i, pr := range b.Proofs {
		script := filepath.Join(runDir, fmt.Sprintf("probe-proof-%d.py", i+1))
		if err := os.WriteFile(script, []byte(pr.Script+"\n"), 0o644); err != nil {
			return err
		}
		args := []string{"blue", "prove", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--location", pr.Location, "--script", script,
			"--reason", "the computation behind this sentence"}
		if pr.Answers != "" {
			args = append(args, "--answers", pr.Answers)
		}
		if _, err := exec(args...); err != nil {
			return fmt.Errorf("prove %d: %w", i+1, err)
		}
	}

	for i, claim := range b.Claims {
		// A REACHABLE url, because `cite` FETCHES and caches. An unreachable one is refused and
		// logged as friction, which is correct behaviour and useless here.
		if _, err := exec("blue", "cite", "--run", runDir, "--seat-id", "blue-respond-r1",
			"--key", fmt.Sprintf("C%d", i+1), "--location", claim,
			"--title", "the pinned source", "--url", "https://example.com/",
			"--reason", "the source this claim rests on"); err != nil {
			return fmt.Errorf("cite %d (%q): %w — the board declares this claim and its expectations are about acting on it; building without it would produce a board whose demands cannot be met", i+1, claim, err)
		}
	}
	return nil
}
