package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// result is one gate's outcome. SKIPPED is its own state, never folded into PASS: a run that
// could not execute a gate has not checked it, and reporting that as green is the exact
// failure this tool was built after.
type result struct {
	gate    gate
	status  string // PASS | FAIL | SKIP
	took    time.Duration
	details string
	// notMeasured distinguishes the two kinds of SKIP. A DECLARED skip (gate.skip) is a gate
	// this environment cannot run and CI will; nothing is owed here. A NOT-MEASURED skip is a
	// gate that ran and found nothing to look at — CI runs it for real, where it can fail.
	// Reporting both as "CI still runs them" is true for one and a lie for the other.
	notMeasured bool
}

func main() {
	list := flag.Bool("list", false, "print the gate set and where each one runs, then exit")
	only := flag.String("only", "", "run only gates whose id contains one of these (comma-separated)")
	base := flag.String("base", "origin/main", "ref the diff-based gates compare against")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "check:", err)
		os.Exit(1)
	}
	gs := gateSet()

	if *list {
		fmt.Printf("%-24s %-8s %-14s %s\n", "GATE", "KIND", "CI JOB", "WHY")
		for _, g := range gs {
			note := g.why
			if g.skip != "" {
				note = "SKIPS HERE: " + g.skip
			}
			fmt.Printf("%-24s %-8s %-14s %s\n", g.id, g.kind, g.ciJob, note)
		}
		fmt.Printf("\n%d gate(s). Parity with %s is enforced by parity_test.go.\n", len(gs), workflowPath)
		return
	}

	selected := selectGates(gs, *only)
	if len(selected) == 0 {
		fmt.Fprintf(os.Stderr, "check: -only %q matched no gate. `-list` shows the set.\n", *only)
		os.Exit(2)
	}

	// A diff-based gate against a stale ref passes locally and fails in CI — measured on
	// #400, where origin/main was two commits behind. Fetching is cheap; being wrong is not.
	if anyNeedsBase(selected) {
		if _, err := gitx.Run(root, "fetch", "origin", strings.TrimPrefix(*base, "origin/")); err != nil {
			fmt.Fprintf(os.Stderr, "check: could not fetch %s (%v) — diff-based gates would compare "+
				"against a stale ref, so they are not run.\n", *base, err)
			os.Exit(1)
		}
	}

	var results []result
	for _, g := range selected {
		results = append(results, runGate(root, g, *base))
	}
	os.Exit(report(results))
}

func selectGates(gs []gate, only string) []gate {
	if strings.TrimSpace(only) == "" {
		return gs
	}
	wants := strings.Split(only, ",")
	var kept []gate
	for _, g := range gs {
		for _, w := range wants {
			if w = strings.TrimSpace(w); w != "" && strings.Contains(g.id, w) {
				kept = append(kept, g)
				break
			}
		}
	}
	return kept
}

func anyNeedsBase(gs []gate) bool {
	for _, g := range gs {
		if g.needsBase && g.skip == "" {
			return true
		}
	}
	return false
}

// runGate executes one gate and captures its output. Output is CAPTURED rather than streamed
// so the summary is readable; a failing gate prints its output in full, a passing one prints
// nothing. Twelve gates streaming at once is how a real failure gets scrolled past.
func runGate(root string, g gate, base string) result {
	if g.skip != "" {
		return result{gate: g, status: "SKIP", details: g.skip}
	}
	dir := filepath.Join(root, g.dir)
	start := time.Now()

	var cmd *exec.Cmd
	switch g.kind {
	case kindFmt:
		// gofmt -l names the offenders; empty output is the pass. Expressed here rather than
		// shelled out so it behaves the same on Windows, where CI's `test -z "$(gofmt -l .)"`
		// is a shell idiom this tool cannot rely on.
		//
		// cmd.Dir is set EXPLICITLY. Without it gofmt lints the process's cwd — which is
		// scripts/ when this is invoked the documented way, so every module's fmt gate would
		// have reported on scripts/ and three of the four would have been checking nothing
		// while printing a pass.
		fmtCmd := exec.Command("gofmt", "-l", ".")
		fmtCmd.Dir = dir
		out, err := fmtCmd.Output()
		took := time.Since(start)
		if err != nil {
			return result{gate: g, status: "FAIL", took: took, details: "gofmt: " + err.Error()}
		}
		if s := strings.TrimSpace(string(out)); s != "" {
			return result{gate: g, status: "FAIL", took: took, details: "unformatted:\n" + s}
		}
		return result{gate: g, status: "PASS", took: took}
	case kindNode:
		cmd = exec.Command("node", g.args...)
	case kindTool:
		// BUILT, then run — never `go run`. `go run` collapses ANY non-zero exit to 1, so a
		// tool's exit code cannot reach this function through it: exitNotMeasured (3) arrived
		// here as 1 and was reported as a failure, indistinguishable from a real one. Measured
		// while building the not-measured path (#409). The build is cached, so this costs
		// nothing after the first run, and it makes every tool gate's exit code honest rather
		// than only rulesweep's.
		bin, cleanup, err := buildTool(dir, g)
		if err != nil {
			return result{gate: g, status: "FAIL", took: time.Since(start), details: err.Error()}
		}
		defer cleanup()
		var args []string
		if g.needsBase {
			args = append(args, "-base", base)
		}
		args = append(args, toolFlags(g)...)
		cmd = exec.Command(bin, args...)
	default:
		cmd = exec.Command("go", g.args...)
	}
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	took := time.Since(start)
	if err != nil {
		// EXIT 3 = the gate ran and could NOT MEASURE (see exitNotMeasured). Not a pass: a
		// gate reporting "clean" when it had nothing to look at is the plausible zero this
		// tool exists to stop printing. Not a failure either — the developer has done nothing
		// wrong, the gate simply cannot fire yet. Its own reason travels with it.
		//
		// Measured: rulesweep reads Rule-Class:/Sibling-Sweep: from COMMIT messages, and its
		// `git diff base...HEAD` is committed-only. Run with the work still in the working
		// tree it saw zero changed files, said "no protocol surface touched", and exited 0.
		// check reported 26/26; CI then failed rule-sweep (#409).
		if exitCodeOf(err) == exitNotMeasured {
			return result{gate: g, status: "SKIP", took: took, notMeasured: true, details: strings.TrimSpace(string(out))}
		}
		return result{gate: g, status: "FAIL", took: took, details: strings.TrimSpace(string(out))}
	}
	return result{gate: g, status: "PASS", took: took}
}

// toolFlags returns a tool gate's own flags — everything in args after `run` and the package
// path (e.g. mutate's `-selftest`).
func toolFlags(g gate) []string {
	if len(g.args) > 2 {
		return g.args[2:]
	}
	return nil
}

// buildTool compiles a scripts/ tool to a temp binary so its exit code survives. args[1] is
// the package path (`./rulesweep`).
func buildTool(dir string, g gate) (bin string, cleanup func(), err error) {
	if len(g.args) < 2 {
		return "", func() {}, fmt.Errorf("gate %s: no package path in args %v", g.id, g.args)
	}
	tmp, err := os.MkdirTemp("", "check-"+g.id+"-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup = func() { _ = os.RemoveAll(tmp) }
	bin = filepath.Join(tmp, g.id)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, g.args[1])
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("building %s: %v\n%s", g.args[1], err, out)
	}
	return bin, cleanup, nil
}

// exitNotMeasured is the exit code a gate uses to say it could not measure anything. A CODE
// rather than a phrase in the output: a runner matching on prose breaks the moment someone
// rewords the message, and reports the reword as a pass.
const exitNotMeasured = 3

func exitCodeOf(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// indentLines keeps a multi-line gate reason inside the report's left margin.
func indentLines(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\n", "\n"+strings.Repeat(" ", 27))
}

// GOTOOLCHAIN=local matches CI, which pins it deliberately: under `auto` a runner silently
// DOWNLOADS a newer toolchain to build a module that pins one, and the build goes green while
// testing a toolchain nobody chose. The workflow's own comment records that happening.
func init() { _ = os.Setenv("GOTOOLCHAIN", "local") }

// report prints the summary and returns the exit code.
//
// SKIPPED IS COUNTED AND NAMED. A summary that says "11 passed" when three gates never ran is
// the plausible-zero this repo keeps finding; this one says what it did not check.
func report(rs []result) int {
	var failed, skipped int
	fmt.Println()
	for _, r := range rs {
		mark := "  ok  "
		switch r.status {
		case "FAIL":
			mark, failed = "  FAIL", failed+1
		case "SKIP":
			mark, skipped = "  skip", skipped+1
		}
		took := ""
		if r.took > 0 {
			took = fmt.Sprintf("%6.1fs", r.took.Seconds())
		}
		fmt.Printf("%s %-24s %8s %s\n", mark, r.gate.id, took, shortReason(r))
	}

	fmt.Printf("\n%d gate(s): %d passed, %d failed, %d not run here\n",
		len(rs), len(rs)-failed-skipped, failed, skipped)

	var declared, unmeasured []result
	for _, r := range rs {
		switch {
		case r.status != "SKIP":
		case r.notMeasured:
			unmeasured = append(unmeasured, r)
		default:
			declared = append(declared, r)
		}
	}
	if len(declared) > 0 {
		fmt.Println("\nNOT RUN HERE — these are declared, not forgotten, and CI still runs them:")
		for _, r := range declared {
			fmt.Printf("  %-24s %s\n", r.gate.id, r.details)
		}
	}
	if len(unmeasured) > 0 {
		fmt.Println("\nDID NOT MEASURE — these RAN and found nothing to look at. CI runs them for real,")
		fmt.Println("where they can FAIL. A gate that did not fire is not a gate that passed:")
		for _, r := range unmeasured {
			fmt.Printf("  %-24s %s\n", r.gate.id, indentLines(r.details))
		}
	}
	if failed == 0 {
		if len(unmeasured) > 0 {
			fmt.Printf("\ncheck: %d gate(s) did not measure — resolve that and re-run before pushing\n", len(unmeasured))
			return 0
		}
		fmt.Println("\ncheck: OK")
		return 0
	}
	fmt.Println("\n── FAILURES ──")
	for _, r := range rs {
		if r.status == "FAIL" {
			fmt.Printf("\n── %s (%s, CI job: %s)\n%s\n", r.gate.id, r.gate.why, r.gate.ciJob, r.details)
		}
	}
	fmt.Printf("\ncheck: FAILED — %d gate(s)\n", failed)
	return 1
}

func shortReason(r result) string {
	if r.status == "SKIP" {
		return "(not run here)"
	}
	return ""
}
