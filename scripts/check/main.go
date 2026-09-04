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
	"github.com/ctoforaday/special-circumstances/scripts/internal/goldenmods"
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
	for _, g := range subsumeGoldenLegs(selected) {
		results = append(results, runGate(root, g, *base))
	}
	os.Exit(report(results))
}

// subsumeGoldenLegs tells the golden gate which of its go legs THIS run already executes.
//
// golden's go leg per module is `go test -count=1 ./...` — byte-for-byte the module's `:test`
// gate — so a check run that includes both runs the same suite twice, sequentially: ~620s of a
// ~1550s local run re-deriving a result the run already has (#626). The subsumption is only
// claimed when the module's test gate is actually in the selected set (an `-only golden` run
// still drives the leg itself), and golden REPORTS the leg as subsumed rather than dropping it —
// a leg that vanishes from the output is indistinguishable from one that passed. The module list
// is internal/goldenmods, the one record both tools read; if the lists lived apart, this would
// be a hope about two files agreeing.
func subsumeGoldenLegs(selected []gate) []gate {
	testGateOf := map[string]string{}
	for _, m := range modules {
		testGateOf[m.dir] = m.ciJob + ":test"
	}
	present := map[string]bool{}
	for _, g := range selected {
		if g.skip == "" {
			present[g.id] = true
		}
	}
	var pairs []string
	for _, mod := range goldenmods.Modules {
		if id := testGateOf[mod]; id != "" && present[id] {
			pairs = append(pairs, mod+"="+id)
		}
	}
	if len(pairs) == 0 {
		return selected
	}
	out := make([]gate, len(selected))
	copy(out, selected)
	for i, g := range out {
		if g.id == "golden" && g.skip == "" {
			args := make([]string, len(g.args), len(g.args)+1)
			copy(args, g.args)
			out[i].args = append(args, "-subsume="+strings.Join(pairs, ","))
		}
	}
	return out
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
// expandScope replaces a deps:<target> scope marker with the module-local packages in the
// target's import graph, resolved in dir at the moment the gate runs — the same derivation
// the workflow's race step performs, so neither carrier holds a copy the other could drift
// from. The module path comes from the module itself (`go list -m`), not a constant here.
//
// AN EMPTY EXPANSION IS AN ERROR, NEVER AN EMPTY GATE. `go test -race` with no packages
// would test the current directory and report a pass that measured almost nothing — the
// plausible zero again. A target that resolves to no module-local packages means the marker
// or the module is broken, and the gate says so instead of going green.
func expandScope(dir string, args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if !strings.HasPrefix(a, depsScopePrefix) {
			out = append(out, a)
			continue
		}
		target := strings.TrimPrefix(a, depsScopePrefix)
		mod := exec.Command("go", "list", "-m")
		mod.Dir = dir
		modOut, err := mod.Output()
		if err != nil {
			return nil, fmt.Errorf("expanding %s: go list -m in %s: %w", a, dir, err)
		}
		modPath := strings.TrimSpace(string(modOut))
		deps := exec.Command("go", "list", "-deps", target)
		deps.Dir = dir
		depsOut, err := deps.Output()
		if err != nil {
			return nil, fmt.Errorf("expanding %s: go list -deps %s in %s: %w", a, target, dir, err)
		}
		var pkgs []string
		for _, p := range strings.Split(strings.TrimSpace(string(depsOut)), "\n") {
			if strings.HasPrefix(p, modPath) {
				pkgs = append(pkgs, p)
			}
		}
		if len(pkgs) == 0 {
			return nil, fmt.Errorf("scope %s expanded to no packages under module %s — not measured, not clean", a, modPath)
		}
		out = append(out, pkgs...)
	}
	return out, nil
}

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
		args, err := expandScope(dir, g.args)
		if err != nil {
			return result{gate: g, status: "FAIL", took: time.Since(start), details: err.Error()}
		}
		cmd = exec.Command("go", args...)
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
