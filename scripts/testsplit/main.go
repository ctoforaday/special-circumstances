// Command testsplit runs a module's whole suite, with the packages named in
// internal/testsplit's table run as several processes each.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # Why
//
// `go test` gives a package one process, and a package whose tests cannot be parallel —
// t.Setenv, a package-level swap — runs them one at a time in it. Such a package uses one core
// however many the box has, and the gate waits on it. Process isolation gives each share of
// the tests its own environment and package state, which is the thing that kept them serial.
//
// # What it does
//
//	go run ./testsplit -module <dir> [-n N] [-timeout D]
//
// It runs the WHOLE of `./...`, so no package can fall between two invocations. A package
// without a row runs under plain `go test`; a package with one is compiled once, its tests
// listed, and the names dealt round-robin to N processes of that binary, each started in the
// package's directory with the caller's environment. Every listed name must come back with a
// verdict, read from the JSON event stream; a name that does not is NOT MEASURED and fails the
// run. `-n 1` splits nothing: every package runs under plain `go test`, and that is the
// default on a box of four cores or fewer (defaultProcesses says why).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ctoforaday/special-circumstances/scripts/internal/testsplit"
)

// GOTOOLCHAIN=local matches CI, which pins it deliberately.
func init() { _ = os.Setenv("GOTOOLCHAIN", "local") }

func main() {
	module := flag.String("module", "", "directory of the module whose suite this runs")
	n := flag.Int("n", defaultProcesses(runtime.NumCPU()), "processes per split package")
	timeout := flag.String("timeout", "10m", "bound on each package, and on each process of a split one")
	flag.Parse()
	if *module == "" {
		fmt.Fprintln(os.Stderr, "testsplit: -module is required")
		os.Exit(2)
	}
	path, err := goOutput(*module, "list", "-m")
	if err != nil {
		fmt.Fprintln(os.Stderr, "testsplit:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Stdout, *module, testsplit.For(strings.TrimSpace(path)), *n, *timeout))
}

// defaultProcesses is half the cores, at most eight — and ONE on a box of four cores or fewer,
// which splits nothing.
//
// Splitting pays only where cores sit idle. Plain `go test` already runs as many packages at
// once as there are cores, so on a small box every core is busy and more processes are only
// contention. MEASURED on a four-core CI runner: the suite step took 111s and 134s through
// two processes per split package, against 73-104s for plain `go test`; internal/cli alone
// went from 57s to 100s. Sixteen cores: 94s down to 59-71s. Past eight processes the shares
// are short enough that each process's own start-up dominates.
func defaultProcesses(cpus int) int {
	if cpus <= 4 {
		return 1
	}
	n := cpus / 2
	if n > 8 {
		n = 8
	}
	return n
}

// outcome is one package's result, whichever path ran it.
type outcome struct {
	pkg     string
	split   int // processes it ran as; 0 for the plain path
	took    time.Duration
	tally   tally
	problem string // non-empty is a failure; carries what to print
}

func run(w io.Writer, moduleDir string, rows []testsplit.Row, n int, timeout string) int {
	if n < 1 {
		n = 1
	}
	listed, err := goOutput(moduleDir, "list", "./...")
	if err != nil {
		fmt.Fprintln(w, "testsplit:", err)
		return 1
	}
	pkgs := strings.Fields(listed)
	// AN EMPTY SCOPE IS AN ERROR. `go test` over no packages tests the current directory.
	if len(pkgs) == 0 {
		fmt.Fprintf(w, "testsplit: %s lists no packages — not measured, not clean\n", moduleDir)
		return 1
	}
	inScope := map[string]bool{}
	for _, p := range pkgs {
		inScope[p] = true
	}
	split := map[string]bool{}
	for _, r := range rows {
		if !inScope[r.Package] {
			fmt.Fprintf(w, "testsplit: the table names %s, which is not a package of this module — the row splits nothing\n", r.Package)
			return 1
		}
		split[r.Package] = true
	}
	// ONE PROCESS IS PLAIN `go test`. With nothing to deal to, a split package would be compiled
	// on its own and run beside the rest for no gain, so the rows are checked above and then
	// set aside: every package takes the plain path.
	if n == 1 {
		rows, split = nil, map[string]bool{}
	}
	var plain []string
	for _, p := range pkgs {
		if !split[p] {
			plain = append(plain, p)
		}
	}

	tmp, err := os.MkdirTemp("", "testsplit-")
	if err != nil {
		fmt.Fprintln(w, "testsplit:", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	var (
		mu       sync.Mutex
		outcomes []outcome
		wg       sync.WaitGroup
	)
	add := func(os ...outcome) {
		mu.Lock()
		defer mu.Unlock()
		outcomes = append(outcomes, os...)
	}
	if len(plain) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// The plain path keeps every core. `-p` bounds the BUILD as well as the packages
			// run at once, and holding cores back for the split processes was measured on a
			// four-core runner: -p 2 took 134s where plain `go test` took 73-104s. The split
			// processes overlap it only while their packages run.
			parallel := runtime.NumCPU()
			add(runPlain(moduleDir, plain, parallel, timeout)...)
		}()
	}
	for i, r := range rows {
		wg.Add(1)
		go func() {
			defer wg.Done()
			add(runSplit(moduleDir, r.Package, filepath.Join(tmp, fmt.Sprint(i)), n, timeout))
		}()
	}
	wg.Wait()

	sort.Slice(outcomes, func(i, j int) bool { return outcomes[i].pkg < outcomes[j].pkg })
	failed := 0
	for _, o := range outcomes {
		mark, how := "ok  ", ""
		if o.problem != "" {
			mark, failed = "FAIL", failed+1
		}
		if o.split > 0 {
			how = fmt.Sprintf("  [%d processes]", o.split)
		}
		fmt.Fprintf(w, "%s  %-90s %7.1fs  %s%s\n", mark, o.pkg, o.took.Seconds(), o.tally, how)
	}
	if failed > 0 {
		fmt.Fprintln(w, "\n── FAILURES ──")
		for _, o := range outcomes {
			if o.problem != "" {
				fmt.Fprintf(w, "\n── %s\n%s\n", o.pkg, o.problem)
			}
		}
		fmt.Fprintf(w, "\ntestsplit: FAILED — %d package(s)\n", failed)
		return 1
	}
	fmt.Fprintf(w, "\ntestsplit: OK — %d package(s)\n", len(outcomes))
	return 0
}

// runPlain is the path for every package without a row: one `go test`, as before.
func runPlain(moduleDir string, pkgs []string, parallel int, timeout string) []outcome {
	start := time.Now()
	cmd := exec.Command("go", testsplit.UnsplitArgs(timeout, parallel, pkgs)...)
	cmd.Dir = moduleDir
	stream, runErr := cmd.CombinedOutput()
	byPkg, stray := fold(stream)
	var out []outcome
	var silent []string
	for _, pkg := range pkgs {
		p := byPkg[pkg]
		if p == nil || p.verdict == "" {
			silent = append(silent, pkg)
			continue
		}
		o := outcome{pkg: pkg, took: p.elapsed, tally: p.tally()}
		if p.verdict == "fail" {
			o.problem = p.output
			if strings.Contains(p.output, "[build failed]") {
				o.problem += stray
			}
		}
		out = append(out, o)
	}
	// A package that sent no verdict did not run: a build failure upstream of it, or go test
	// itself failing. One failure carries them, with everything that was not an event.
	if len(silent) > 0 || (runErr != nil && !anyFailed(out)) {
		out = append(out, outcome{pkg: "(go test)", took: time.Since(start),
			problem: fmt.Sprintf("go test: %v\nNOT MEASURED, no verdict for: %s\n%s", runErr, strings.Join(silent, " "), stray)})
	}
	return out
}

func anyFailed(os []outcome) bool {
	for _, o := range os {
		if o.problem != "" {
			return true
		}
	}
	return false
}

// runSplit compiles one package's test binary, deals its tests to n processes and requires a
// verdict for every name.
func runSplit(moduleDir, pkg, work string, n int, timeout string) outcome {
	start := time.Now()
	o := outcome{pkg: pkg}
	fail := func(format string, a ...any) outcome {
		o.took, o.problem = time.Since(start), fmt.Sprintf(format, a...)
		return o
	}
	if err := os.MkdirAll(work, 0o755); err != nil {
		return fail("%v", err)
	}
	bin := filepath.Join(work, "pkg.test")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", testsplit.CompileArgs(bin, pkg)...)
	build.Dir = moduleDir
	if out, err := build.CombinedOutput(); err != nil {
		return fail("building the test binary: %v\n%s", err, out)
	}
	if _, err := os.Stat(bin); err != nil {
		return fail("%s has no tests to split; its row splits nothing", pkg)
	}
	// A test binary runs in its package's directory: that is where its testdata is.
	dir, err := goOutput(moduleDir, "list", "-f", "{{.Dir}}", pkg)
	if err != nil {
		return fail("%v", err)
	}
	dir = strings.TrimSpace(dir)
	list := exec.Command(bin, "-test.list", ".*")
	list.Dir = dir
	listed, err := list.Output()
	if err != nil {
		return fail("listing tests: %v\n%s", err, listed)
	}
	names := testNames(string(listed))
	if len(names) == 0 {
		return fail("%s lists no tests — not measured, not clean", pkg)
	}
	shares := deal(names, n)
	o.split = len(shares)

	type part struct {
		stream []byte
		err    error
		raw    []byte
	}
	parts := make([]part, len(shares))
	var wg sync.WaitGroup
	for i, share := range shares {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(bin, testsplit.SplitArgs(timeout, share)...)
			cmd.Dir = dir
			raw, runErr := cmd.CombinedOutput()
			conv := exec.Command("go", "tool", "test2json", "-p", pkg)
			conv.Stdin = strings.NewReader(string(raw))
			stream, convErr := conv.Output()
			if convErr != nil {
				runErr = fmt.Errorf("test2json: %w (the process itself: %v)", convErr, runErr)
			}
			parts[i] = part{stream: stream, err: runErr, raw: raw}
		}()
	}
	wg.Wait()

	var problems []string
	for i, p := range parts {
		byPkg, _ := fold(p.stream)
		got := byPkg[pkg]
		if got == nil {
			got = &pkgEvents{tests: map[string]string{}}
		}
		var missing, failedNames []string
		for _, name := range shares[i] {
			switch got.tests[name] {
			case "pass":
				o.tally.pass++
			case "skip":
				o.tally.skip++
			case "fail":
				o.tally.fail++
				failedNames = append(failedNames, name)
			default:
				missing = append(missing, name)
			}
		}
		if p.err == nil && len(missing) == 0 && len(failedNames) == 0 {
			continue
		}
		var what []string
		if len(failedNames) > 0 {
			what = append(what, "failed: "+strings.Join(failedNames, ", "))
		}
		if len(missing) > 0 {
			what = append(what, "NOT MEASURED, no verdict came back for: "+strings.Join(missing, ", "))
		}
		if p.err != nil {
			what = append(what, "the process: "+p.err.Error())
		}
		problems = append(problems, fmt.Sprintf("process %d of %d — %s\n%s", i+1, len(shares), strings.Join(what, "; "), got.outputOr(p.raw)))
	}
	o.took = time.Since(start)
	// EVERY LISTED NAME, ONCE. The shares are dealt from the list, so the tally is the list
	// unless a name went unanswered — which is a problem above, never a smaller total.
	if got := o.tally.pass + o.tally.skip + o.tally.fail; len(problems) == 0 && got != len(names) {
		problems = append(problems, fmt.Sprintf("%d names were listed and %d answered", len(names), got))
	}
	o.problem = strings.Join(problems, "\n\n")
	return o
}

// testNames reads `-test.list` output: one name per line, Test, Fuzz and Example alike. A
// line that is not one of those is something the binary printed, and is not a test.
func testNames(listed string) []string {
	var names []string
	for _, line := range strings.Split(listed, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Fuzz") || strings.HasPrefix(line, "Example") {
			names = append(names, line)
		}
	}
	return names
}

// deal hands the names round-robin to at most n shares, in listed order, dropping none.
func deal(names []string, n int) [][]string {
	if n > len(names) {
		n = len(names)
	}
	shares := make([][]string, n)
	for i, name := range names {
		shares[i%n] = append(shares[i%n], name)
	}
	return shares
}

func goOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return string(out), nil
}
