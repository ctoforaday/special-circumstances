// Command racescope is the race gate for a module whose suite is too slow to race whole.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # The rule
//
// Every concurrency start in the shipped binary's import graph, and every test that makes one,
// runs under the race detector. Not every test in the graph: the detector reports a race
// between two goroutines of the test binary, and a test that starts none has nothing for it to
// report while still paying its 12–26x.
//
// # What it does
//
//		go run ./racescope -module <dir> -deps ./cmd/<binary>
//
//	 1. The scope is the binary's module-local import graph (internal/racescope.Scope).
//	 2. Concurrency starts are found by syntax tree and type information: a `go` statement, or a
//	    call to a row of internal/racescope.Table.
//	 3. The race set is DERIVED: a Test, Fuzz or Example function of a scope package is in it
//	    when a start is reachable from it by direct calls through test files and through the
//	    non-test files of packages outside the scope (derive.go states the edges). A start in
//	    shipped source is owned by the test that carries `//racescope:covers <file>:<function>`.
//	 4. The verdict run races the set, with no coverage.
//	 5. The evidence run executes the set again with no race detector and a coverage profile,
//	    and every start in a non-test file must have been executed. One that was not fails the
//	    gate, named by file and line.
//
// The two runs are separate because measured together they are wrong; run.go says why.
//
// # What it does not see
//
// A start reached only through a function value or an interface, and a library call that is
// not a row of the table. `.github/workflows/race-nightly.yml` races the whole graph for those.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
)

// GOTOOLCHAIN=local matches CI, which pins it deliberately.
func init() { _ = os.Setenv("GOTOOLCHAIN", "local") }

func main() {
	module := flag.String("module", "", "directory of the module whose race gate this is")
	target := flag.String("deps", "", "the binary whose module-local import graph is the scope, e.g. ./cmd/feov-record")
	timeout := flag.String("timeout", "30m", "per-package bound on each run")
	parallel := flag.Int("p", runtime.NumCPU(), "packages run at once")
	deriveOnly := flag.Bool("derive", false, "print the race set and the starts it must execute, and run nothing")
	flag.Parse()
	if *module == "" || *target == "" {
		fmt.Fprintln(os.Stderr, "racescope: -module and -deps are both required")
		os.Exit(2)
	}
	os.Exit(run(os.Stdout, *module, *target, *timeout, *parallel, *deriveOnly))
}

func run(w io.Writer, module, target, timeout string, parallel int, deriveOnly bool) int {
	d, err := derive(module, target)
	if err != nil {
		fmt.Fprintln(w, "racescope:", err)
		return 1
	}
	var pkgs []string
	for pkg := range d.sets {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	total := 0
	for _, pkg := range pkgs {
		total += len(d.sets[pkg])
		fmt.Fprintf(w, "%s — %d selected\n", short(d, pkg), len(d.sets[pkg]))
		for _, s := range d.sets[pkg] {
			fmt.Fprintf(w, "    %-60s %s\n", s.name, s.why)
		}
	}
	fmt.Fprintf(w, "\nrace set: %d test(s) in %d package(s); %d start(s) in non-test files to execute\n", total, len(pkgs), len(d.owed))

	if len(d.problems) > 0 {
		fmt.Fprintln(w, "\n── REFUSED before running ──")
		for _, p := range d.problems {
			fmt.Fprintln(w, "  "+p)
		}
		fmt.Fprintf(w, "\nracescope: FAILED — %d problem(s)\n", len(d.problems))
		return 1
	}
	// A scope with starts and an empty set is caught above (nothing links them). A scope with
	// neither has nothing to race, and saying so is not the same as a run that raced nothing.
	if total == 0 {
		fmt.Fprintln(w, "\nracescope: no concurrency start in scope and none in its tests — nothing to race")
		return 0
	}
	if deriveOnly {
		for _, s := range d.owed {
			fmt.Fprintln(w, "  owed: "+s.String())
		}
		return 0
	}

	var problems []string
	fmt.Fprintln(w, "\n── verdict run: -race, no coverage ──")
	for _, v := range raceRun(module, d, pkgs, timeout, parallel) {
		line := fmt.Sprintf("%s — %d selected: %d pass, %d fail, %d skip", short(d, v.pkg), len(d.sets[v.pkg]), v.pass, len(v.failed), len(v.skipped))
		if len(v.skipped) > 0 {
			// A concurrency test that skipped was not raced. Printed by name: a set that
			// shrinks is a number that moved, and a skip is a line a reviewer reads.
			line += " (skipped, so NOT raced: " + strings.Join(v.skipped, ", ") + ")"
		}
		fmt.Fprintln(w, line)
		if !v.red() {
			continue
		}
		var what []string
		if v.race {
			what = append(what, "WARNING: DATA RACE reported")
		}
		if len(v.failed) > 0 {
			what = append(what, "failed: "+strings.Join(v.failed, ", "))
		}
		if len(v.missing) > 0 {
			what = append(what, "NOT MEASURED, no verdict came back for: "+strings.Join(v.missing, ", ")+
				" (an Example needs an `// Output:` comment to run)")
		}
		if v.runErr != nil {
			what = append(what, "go test: "+v.runErr.Error())
		}
		problems = append(problems, fmt.Sprintf("%s under -race — %s\n%s", short(d, v.pkg), strings.Join(what, "; "), v.output))
	}

	fmt.Fprintln(w, "\n── evidence run: coverage, no -race ──")
	tmp, err := os.MkdirTemp("", "racescope-")
	if err != nil {
		fmt.Fprintln(w, "racescope:", err)
		return 1
	}
	defer os.RemoveAll(tmp)
	profiles, runErrs := evidenceRun(module, d, pkgs, timeout, tmp, parallel)
	blocks, err := readProfiles(profiles)
	if err != nil {
		problems = append(problems, runErrs...)
		problems = append(problems, "the starts are NOT MEASURED: "+err.Error())
	} else {
		proved, unproven := prove(d.owed, blocks)
		for _, s := range proved {
			fmt.Fprintln(w, "  executed: "+s.String())
		}
		problems = append(problems, unproven...)
		// A failing evidence run with every start executed is the verdict run's business,
		// and the verdict run has already said so; reported here only when starts are missing.
		if len(unproven) > 0 {
			problems = append(problems, runErrs...)
		}
	}

	if len(problems) > 0 {
		fmt.Fprintln(w, "\n── FAILURES ──")
		for _, p := range problems {
			fmt.Fprintln(w, p)
		}
		fmt.Fprintf(w, "\nracescope: FAILED — %d problem(s)\n", len(problems))
		return 1
	}
	fmt.Fprintf(w, "\nracescope: OK — %d test(s) raced, %d start(s) executed\n", total, len(d.owed))
	return 0
}

// short prints a package without the module prefix every line would share.
func short(d *derived, pkg string) string {
	if s := strings.TrimPrefix(pkg, d.module+"/"); s != "" {
		return s
	}
	return pkg
}
