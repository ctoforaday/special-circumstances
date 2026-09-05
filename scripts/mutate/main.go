// Command mutate answers the question coverage cannot: does the suite actually TEST the
// code, or merely execute it?
//
// Dev tooling for this repository only (CLAUDE.md: dev concerns live at the root,
// consumer behaviour lives in plugins/). Nothing here ships to an installing project.
//
// WHY THIS EXISTS. internal/secrets reported 100.0% of statements while two of its eight
// secret patterns could be DELETED OUTRIGHT with the suite green — on the one gate whose
// contract is agent-guardrails: never send secrets outward. Statement coverage could not
// see it, and could never see it: Scan() loops over the pattern table, so every pattern is
// "executed" whether or not anything asks it to match. Coverage measures REACH. This
// measures whether the tests would NOTICE.
//
// The method: flip one operator, run the narrowest suite that covers it, restore. A mutant
// that still passes is a branch no test pins. Measured 2026-07-30 over
// prosthetic-conscience/tools: 53 of 277 behavioural mutants survived at 77–100% statement
// coverage per package; after the pass that closed them, 19 of 280.
//
// # Why this is its own module, and not under a plugin's tools/cmd
//
// sc-doctor discovers hook binaries with ReadDir(<plugin>/tools/cmd) and treats every
// directory there as a binary the operator is expected to have built. A dev tool parked in
// that tree would show as "✗ not built" in the doctor table and `--fix` would try to fetch
// a release asset for it. So this lives at the repo root with its own go.mod, where the
// doctor cannot mistake it for something a session needs.
//
// # An audit day to day, a gate at the release boundary
//
//   - Minutes, not seconds: one suite run per mutant, serially, because the mutant lives in
//     the file on disk. Parallelism would need a checkout per worker.
//   - The residue is JUDGEMENT. An equivalent mutant cannot be killed by any test
//     (a separator token can never equal a git verb, so the comparison decides nothing);
//     a platform-conditional one cannot be killed on one OS. Driving the number to zero
//     buys contorted tests. There is deliberately no threshold here: survivors are a list
//     to EXPLAIN, and the ones nobody can explain are the findings.
//
// Ruled 2026-09-03: what a RELEASE must prove is that the explaining HAPPENED. `-gate`
// (see gate.go) reconciles the sweep against the module's mutation-survivors.json — every
// survivor judged in writing, no judgement outliving its survivor — and the release job
// runs it on the tagged plugin's module. Still no threshold, still no per-PR leg: pull
// requests pay nothing, and the day-to-day audit is unchanged.
//
// Also in CI is -selftest. See selftest() for the failure mode that justifies it.
//
// Usage (from this directory, since it is its own module — the sweep target is resolved
// from the git root, so the working directory does not change what is swept):
//
//	go run . -selftest                             what CI runs on every pull request
//	go run .                                       sweep prosthetic-conscience/tools
//	go run . -module plugins/gray-area/tools
//	go run . -filter sc-doctor                     only files matching a substring
//	go run . -gate -module plugins/gray-area/tools what the release job runs on a tag
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
)

// mutation is one operator flip.
type mutation struct{ from, to string }

// mutations is kept STABLE so numbers stay comparable run to run — widening the set
// changes the denominator, which makes "survivors went up" unreadable.
var mutations = []mutation{
	{"&&", "||"},
	{"||", "&&"},
	{"==", "!="},
	{"!=", "=="},
	{">=", ">"},
	{"<=", "<"},
}

// candidate is one operator occurrence in real code.
type candidate struct {
	line   int // 0-based
	column int
	mutation
}

// survivor is a mutant the suite did not notice.
type survivor struct {
	rel  string
	line int // 1-based, for citation
	mutation
	text string
}

// codeMask returns line with string literals, rune literals and comments blanked, so an
// operator found by index in the mask is REAL CODE.
//
// Without this, a `!=` inside a message string gets mutated, the test fails on the changed
// text, and the mutant is scored "killed" — inflating the kill rate with mutants that never
// tested a branch at all. The earlier throwaway version of this tool had exactly that bug.
// inBlock carries /* */ state across lines.
func codeMask(line string, inBlock bool) (string, bool) {
	var out strings.Builder
	block := inBlock
	for i := 0; i < len(line); {
		if block {
			if strings.HasPrefix(line[i:], "*/") {
				block = false
				out.WriteString("  ")
				i += 2
				continue
			}
			out.WriteByte(' ')
			i++
			continue
		}
		if strings.HasPrefix(line[i:], "//") {
			out.WriteString(strings.Repeat(" ", len(line)-i))
			break
		}
		if strings.HasPrefix(line[i:], "/*") {
			block = true
			out.WriteString("  ")
			i += 2
			continue
		}
		if q := line[i]; q == '"' || q == '`' || q == '\'' {
			out.WriteByte(' ')
			i++
			for i < len(line) {
				if q != '`' && line[i] == '\\' {
					out.WriteString("  ")
					i += 2
					continue
				}
				if line[i] == q {
					break
				}
				out.WriteByte(' ')
				i++
			}
			if i < len(line) {
				out.WriteByte(' ')
				i++
			}
			continue
		}
		out.WriteByte(line[i])
		i++
	}
	return out.String(), block
}

// candidates lists every real-code operator occurrence in a source file.
func candidates(source string) []candidate {
	var out []candidate
	inBlock := false
	for n, line := range strings.Split(source, "\n") {
		mask, next := codeMask(line, inBlock)
		inBlock = next
		for _, m := range mutations {
			// Overlapping scan, offset-based. Written this way deliberately: the
			// obvious `at = strings.Index(mask[at+1:], from) + at + 1` form leaves `at`
			// UNCHANGED when Index returns -1, so the loop never terminates and appends
			// until the process is OOM-killed. That was the first version of this
			// function, and it is why -selftest exists to be run before trusting output.
			for off := 0; off < len(mask); {
				i := strings.Index(mask[off:], m.from)
				if i < 0 {
					break
				}
				at := off + i
				out = append(out, candidate{line: n, column: at, mutation: m})
				off = at + 1
			}
		}
	}
	return out
}

// goFiles lists a module's non-test .go files, deterministically.
func goFiles(moduleDir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(moduleDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == "testdata" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

// suiteFor is the package OWNING a file — the first suite a mutant is tried against.
func suiteFor(rel string) string {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		return "./"
	}
	return "./" + dir + "/"
}

// TWO STAGES, AND THE FIRST ONE IS WHY THIS TOOL CAN BE RUN AT ALL.
//
// Every `internal/` mutant used to run `go test ./...` — the whole module — on the argument
// that "an internal/ mutant can affect anything". The argument is TRUE and the cost made the
// tool unusable for exactly the packages worth auditing: the module suite is ~15 minutes here,
// most of it an integration sweep running 60 debate simulations, so a single flipped operator
// cost a quarter of an hour and a 200-site file would take fifty. Measured: two hours of
// sweeping `internal/record` completed about eight mutants and produced no output at all.
//
// A mutant is tried against its OWN package, where nearly all die in milliseconds.
//
// THE WIDE STAGE IS OPT-IN (`-confirm`), AND MAKING IT SO IS THE SECOND HALF OF THE FIX. Trying
// every survivor against the rest of the module is correct and costs ~8 minutes EACH: measured,
// a `citationid.go` sweep ran an hour because its survivors each paid for a full-module run, and
// the operator waiting on it could not be told how many were left. Two stages made the common
// case free and left the reported case ruinous.
//
// So a survivor is reported as what it IS — "survived its own package" — and the expensive
// confirmation is asked for rather than assumed. That matches what this tool already is: an
// on-demand audit whose output is a list to EXPLAIN, not a gate. An operator who wants to know
// whether another package would have caught it can ask; one who is auditing the package in front
// of them should not pay for an answer they did not want.
//
// The correctness argument for the wide stage is unchanged and still true — an `internal/`
// mutant CAN be killed by another package's tests — which is why `-confirm` exists rather than
// the stage being deleted. What changed is who decides to pay.
//
// With `-confirm` the wide run uses `-short` and skips ./releasegate/... : that suite shrinks
// under -short by design and exists to exercise whole debates, not to notice a one-operator flip.
func runSuite(moduleDir, pkg string, confirm bool) (passed, broke bool) {
	passed, broke = runOne(moduleDir, "go", "test", pkg)
	if broke || !passed || !confirm {
		return passed, broke
	}
	pkgs, err := widePackages(moduleDir)
	if err != nil || len(pkgs) == 0 {
		return true, false
	}
	return runOne(moduleDir, append([]string{"go", "test", "-short"}, pkgs...)...)
}

func runOne(moduleDir string, argv ...string) (passed, broke bool) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = moduleDir
	out, err := cmd.CombinedOutput()
	return err == nil, strings.Contains(string(out), "build failed")
}

// widePackages lists the module's packages EXCEPT the integration suite, which is a whole-system
// exercise rather than a unit oracle and dominates the runtime.
func widePackages(moduleDir string) ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = moduleDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var keep []string
	for _, p := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if p == "" || strings.Contains(p, "/releasegate/") {
			continue
		}
		keep = append(keep, p)
	}
	return keep, nil
}

type result struct {
	survived  []survivor
	killed    int
	nocompile int
}

// sweep mutates one occurrence at a time, ALWAYS restoring the file — including on SIGINT,
// because a mutant left in the working tree is worse than no measurement.
// progress is where the per-mutant heartbeat goes. It is SEPARATE FROM out because the two have
// different readers: out carries the report (survivors, the summary), progress carries "still
// alive, here is where I am". Sending both to stdout would put a moving cursor through a document
// somebody pipes to a file.
// sweep runs single-threaded, which is what the selftest and any caller wanting the old shape get.
func sweep(moduleDir, filter string, confirm bool, out *os.File, progress io.Writer) (*result, error) {
	return sweepN(moduleDir, filter, confirm, out, progress, 1, nil)
}

// job is one mutant: which file, which flip. The path is RELATIVE so a worker resolves it inside
// its OWN sandbox — no two workers share a tree, which is the whole reason the pool is safe.
type job struct {
	rel string
	pkg string
	c   candidate
}

type jobResult struct {
	job
	verdict string // "killed" | "SURVIVED" | "no-compile"
	text    string
	seconds float64
	err     error
}

// sweepN runs the sweep across `workers` private trees.
//
// # The premise that ruled this out is gone
//
// The header used to say: "one suite run per mutant, serially, because the mutant lives in the file
// on disk. Parallelism would need a checkout per worker." The first clause stopped being true when
// the sweep moved into a sandbox, and a checkout per worker is exactly what sandbox() already
// makes. The conclusion outlived its premise.
//
// # Where the time goes, measured
//
// internal/record: ~4.0s per mutant — ~1.5s compile and link, ~2.7s the package's own tests
// executing. Both halves are CPU, so a pool moves the whole figure rather than a slice of it.
// Nothing else on offer does: -vet=off touches part of the 1.5s, and the 2.7s IS the suite the
// mutant exists to be judged by.
//
// # Determinism
//
// Jobs are dispatched in order and results reassembled in order, so the report never depends on
// which worker finished first. Only the heartbeat interleaves, and it goes to stderr precisely so
// the ordered report on stdout stays a document.
func sweepN(moduleDir, filter string, confirm bool, out *os.File, progress io.Writer, workers int, mkSandbox func(string) (string, func(), error)) (*result, error) {
	files, err := goFiles(moduleDir)
	if err != nil {
		return nil, err
	}
	if filter != "" {
		var kept []string
		for _, f := range files {
			if strings.Contains(f, filter) {
				kept = append(kept, f)
			}
		}
		files = kept
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no non-test .go files under %s (filter %q)", moduleDir, filter)
	}

	// THE DENOMINATOR IS COMPUTED FIRST, and it is cheap: candidates() only parses, so counting
	// every mutant costs one pass over the source and no test runs at all. This is what the header
	// meant by "the operator waiting on it could not be told how many were left" — the number was
	// always available and simply never asked for.
	var jobs []job
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(moduleDir, path)
		pkg := suiteFor(rel)
		for _, c := range candidates(string(src)) {
			jobs = append(jobs, job{rel: filepath.ToSlash(rel), pkg: pkg, c: c})
		}
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(jobs) {
		workers = len(jobs)
	}
	fmt.Fprintf(progress, "mutate: %d mutant(s) across %d file(s), %d worker(s)\n", len(jobs), len(files), workers)

	// Worker 0 uses the tree the caller already made; the rest get their own. A worker that cannot
	// get one is a REFUSAL rather than a quietly smaller pool — a sweep that halved itself in
	// silence would report a mutant count it never ran.
	trees := make([]string, workers)
	trees[0] = moduleDir
	for i := 1; i < workers; i++ {
		if mkSandbox == nil {
			return nil, fmt.Errorf("mutate: %d workers requested but no tree factory was supplied", workers)
		}
		dir, cleanup, err := mkSandbox(moduleDir)
		if err != nil {
			return nil, fmt.Errorf("mutate: making a tree for worker %d: %w", i, err)
		}
		defer cleanup()
		trees[i] = dir
	}

	in := make(chan int)
	results := make([]jobResult, len(jobs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(tree string) {
			defer wg.Done()
			for idx := range in {
				j := jobs[idx]
				path := filepath.Join(tree, filepath.FromSlash(j.rel))
				source, err := os.ReadFile(path)
				if err != nil {
					results[idx] = jobResult{job: j, err: err}
					continue
				}
				lines := strings.Split(string(source), "\n")
				original := lines[j.c.line]
				lines[j.c.line] = original[:j.c.column] + j.c.to + original[j.c.column+len(j.c.from):]
				if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
					results[idx] = jobResult{job: j, err: err}
					continue
				}

				started := time.Now()
				passed, broke := runSuite(tree, j.pkg, confirm)
				elapsed := time.Since(started).Seconds()

				// RESTORED BEFORE THE NEXT MUTANT, not merely before the next file: a worker takes
				// its next job from anywhere in the queue, so a file left mutated would be measured
				// against by whichever mutant lands on it next.
				if err := restoreFile(path, source); err != nil {
					results[idx] = jobResult{job: j, err: err}
					continue
				}

				r := jobResult{job: j, verdict: "killed", seconds: elapsed, text: strings.TrimSpace(original)}
				switch {
				case broke:
					r.verdict = "no-compile"
				case passed:
					r.verdict = "SURVIVED"
				}
				results[idx] = r

				// ONE LINE PER MUTANT, WHATEVER IT DID. A killed mutant used to print nothing, so a
				// sweep that was working, one that was wedged and one that was dead all produced the
				// same bytes: none. Three runs were misread that way — twice as progress that was
				// not happening, once as a cost that was not real — before anybody suspected the
				// output rather than the tool.
				mu.Lock()
				done++
				fmt.Fprintf(progress, "[%d/%d] %s:%d %s->%s %s (%.1fs)\n",
					done, len(jobs), j.rel, j.c.line+1, j.c.from, j.c.to, r.verdict, elapsed)
				mu.Unlock()
			}
		}(trees[w])
	}
	for i := range jobs {
		in <- i
	}
	close(in)
	wg.Wait()

	var res result
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		switch r.verdict {
		case "no-compile":
			res.nocompile++
		case "SURVIVED":
			s := survivor{rel: r.rel, line: r.c.line + 1, mutation: r.c.mutation, text: r.text}
			res.survived = append(res.survived, s)
			text := s.text
			if len(text) > 100 {
				text = text[:100]
			}
			fmt.Fprintf(out, "SURVIVED %s:%d  %s->%s  |  %s\n", s.rel, s.line, s.from, s.to, text)
		default:
			res.killed++
		}
	}
	return &res, nil
}

func selftest(goVersion string) bool {
	dir, err := os.MkdirTemp("", "sc-mutate-selftest-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "selftest FAILED:", err)
		return false
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// pinned() has its branch asserted; loose() does not. A working sweep reports exactly
	// one survivor, in loose(). The `!=` inside loose()'s string must NOT be mutated — if
	// the mask regressed that becomes a second survivor, which the count below catches.
	write := func(name, body string) bool {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "selftest FAILED:", err)
			return false
		}
		return true
	}
	ok := write("go.mod", "module selftest\n\ngo "+goVersion+"\n") &&
		write("x.go", `package selftest

func pinned(a, b int) bool {
	if a == b {
		return true
	}
	return false
}

func loose(a, b int) string {
	if a == b {
		return "a != b is false"
	}
	return "differs"
}
`) &&
		write("x_test.go", `package selftest

import "testing"

func TestPinned(t *testing.T) {
	if !pinned(1, 1) {
		t.Fatal("equal must be true")
	}
	if pinned(1, 2) {
		t.Fatal("unequal must be false")
	}
}

func TestLooseIsCalledButNotAsserted(t *testing.T) {
	_ = loose(1, 1)
	_ = loose(1, 2)
}
`)
	if !ok {
		return false
	}

	res, err := sweep(dir, "", false, os.Stdout, io.Discard)
	if err != nil {
		fmt.Fprintln(os.Stderr, "selftest FAILED:", err)
		return false
	}

	pass := true
	fail := func(format string, a ...any) {
		fmt.Fprintf(os.Stderr, "selftest FAILED: "+format+"\n", a...)
		pass = false
	}
	if res.killed < 1 {
		fail("no mutant was killed (%d) — the sweep is not running the suite, and a sweep that generates nothing reports a perfect score", res.killed)
	}
	switch {
	case len(res.survived) != 1:
		fail("expected exactly 1 survivor (the unasserted branch), got %d: %+v", len(res.survived), res.survived)
	case !strings.Contains(res.survived[0].text, "if a == b"):
		fail("the survivor is not the unasserted comparison: %q", res.survived[0].text)
	}
	for _, s := range res.survived {
		if strings.Contains(s.text, "a != b is false") {
			fail("an operator inside a string literal was mutated — codeMask has regressed")
		}
	}
	if pass {
		fmt.Printf("selftest ok: 1 survivor in the unasserted branch, %d killed, string literals untouched\n", res.killed)
	}
	return pass
}

// gitRoot resolves the repository root, so the sweep target does not depend on the working
// directory this module happens to be run from.
func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	module := flag.String("module", "plugins/prosthetic-conscience/tools", "module to sweep, relative to the repo root")
	filter := flag.String("filter", "", "only sweep files whose path contains this substring")
	doSelftest := flag.Bool("selftest", false, "prove the tool can still mutate and observe, then exit")
	goVersion := flag.String("go-version", "1.25", "go directive for the -selftest fixture module")
	jobs := flag.Int("jobs", runtime.NumCPU(), "how many mutants to try at once; each worker gets its own copy of the module")
	confirm := flag.Bool("confirm", false, "re-test each SURVIVOR against the rest of the module (correct, and ~8 minutes per survivor)")
	gateMode := flag.Bool("gate", false, "release gate: fail unless every survivor is explained in "+RecordName+" and no entry is stale (see gate.go)")
	flag.Parse()

	// A verdict over a PARTIAL sweep would claim the whole module: the filtered-out files'
	// survivors would read as absent, and their record entries as stale. The audit composes
	// with -filter; the gate does not.
	if *gateMode && *filter != "" {
		fmt.Fprintln(os.Stderr, "mutate: -gate refuses -filter — a gated verdict must be over the whole module")
		os.Exit(1)
	}

	if *doSelftest {
		if !selftest(*goVersion) {
			os.Exit(1)
		}
		return
	}

	root, err := gitRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	moduleDir := filepath.Join(root, *module)

	// NO DIRTY-TREE REFUSAL. It existed because the mutant lived in YOUR tree, where a
	// crashed run could not be told apart from your own edit — "commit or stash first" was
	// the price. The sweep now runs in a copy, so a dirty tree is not a hazard, and copying
	// it is the point: mutation testing is most useful on the tests you just wrote, which
	// are exactly the ones the refusal forbade measuring.

	// THE SWEEP RUNS IN A COPY. Everything this tool has ever got wrong came from mutating
	// the tree other people are using — a failed undo, a signal landing mid-record, and the
	// one no amount of care in the writer can answer: a second agent or a human editing a
	// file while the sweep holds a stale copy of it. 0.06s for the largest module here.
	work, cleanup, err := sandbox(moduleDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mutate:", err)
		os.Exit(1)
	}
	defer cleanup()

	res, err := sweepN(work, *filter, *confirm, os.Stdout, os.Stderr, *jobs, sandbox)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	behavioural := len(res.survived) + res.killed
	rate := 0
	if behavioural > 0 {
		rate = res.killed * 100 / behavioural
	}
	fmt.Printf("\n%d survived / %d behavioural (%d%% killed); %d did not compile and were discarded.\n",
		len(res.survived), behavioural, rate, res.nocompile)
	fmt.Println("Survivors are a list to EXPLAIN, not a number to drive to zero: equivalent mutants and")
	fmt.Println("platform-conditional branches cannot be killed by any test. The ones nobody can explain")
	fmt.Println("are the findings.")

	// The record is read from the REAL module, not the sandbox copy — same bytes today, but
	// the original is the record a release actually ships. cleanup() runs by hand because
	// os.Exit does not run defers, and a leaked sandbox per gated release adds up.
	if *gateMode {
		code := gateVerdict(moduleDir, res, os.Stdout)
		cleanup()
		os.Exit(code)
	}

	// NO "HEAD MOVED" WARNING EITHER. It caught a real defect — a readState guard committed
	// INVERTED by a `git add -A` that ran while a mutant was on disk, under a commit subject
	// about an unrelated test file. That capture is now impossible: the mutant never exists
	// anywhere git is looking. The sweep measures a SNAPSHOT taken at its start, which is
	// also a cleaner thing to report than a measurement of a tree that moved underneath it.
}
