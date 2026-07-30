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
// # An audit, not a gate — deliberately not in CI
//
//   - Minutes, not seconds: one suite run per mutant, serially, because the mutant lives in
//     the file on disk. Parallelism would need a checkout per worker.
//   - The residue is JUDGEMENT. An equivalent mutant cannot be killed by any test
//     (a separator token can never equal a git verb, so the comparison decides nothing);
//     a platform-conditional one cannot be killed on one OS. Driving the number to zero
//     buys contorted tests. There is deliberately no threshold here: survivors are a list
//     to EXPLAIN, and the ones nobody can explain are the findings.
//
// What IS in CI is -selftest. See selftest() for the failure mode that justifies it.
//
// Usage (from this directory, since it is its own module — the sweep target is resolved
// from the git root, so the working directory does not change what is swept):
//
//	go run . -selftest                             what CI runs
//	go run .                                       sweep prosthetic-conscience/tools
//	go run . -module plugins/gray-area/tools
//	go run . -filter sc-doctor                     only files matching a substring
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
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

// suiteFor is the narrowest package covering a file. A cmd/ mutant cannot affect another
// cmd/, and running one package instead of the module is most of the speed here; an
// internal/ mutant can affect anything, so it gets ./...
func suiteFor(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if parts[0] == "cmd" && len(parts) > 1 {
		return "./cmd/" + parts[1] + "/"
	}
	return "./..."
}

// runSuite reports whether the mutant passed, and whether it failed to build (a
// non-compiling mutant is not behavioural and is discarded rather than counted).
func runSuite(moduleDir, pkg string) (passed, broke bool) {
	cmd := exec.Command("go", "test", pkg)
	cmd.Dir = moduleDir
	out, err := cmd.CombinedOutput()
	return err == nil, strings.Contains(string(out), "build failed")
}

type result struct {
	survived  []survivor
	killed    int
	nocompile int
}

// sweep mutates one occurrence at a time, ALWAYS restoring the file — including on SIGINT,
// because a mutant left in the working tree is worse than no measurement.
func sweep(moduleDir, filter string, out *os.File) (*result, error) {
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

	var (
		res     result
		current string
		backup  []byte
	)
	restore := func() {
		if current != "" {
			_ = os.WriteFile(current, backup, 0o644)
			current = ""
		}
	}
	defer restore()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	go func() {
		<-sig
		restore()
		fmt.Fprintln(os.Stderr, "\ninterrupted — file restored")
		os.Exit(130)
	}()

	for _, path := range files {
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(moduleDir, path)
		pkg := suiteFor(rel)
		current, backup = path, source
		lines := strings.Split(string(source), "\n")

		for _, c := range candidates(string(source)) {
			original := lines[c.line]
			lines[c.line] = original[:c.column] + c.to + original[c.column+len(c.from):]
			if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				return nil, err
			}
			lines[c.line] = original

			passed, broke := runSuite(moduleDir, pkg)
			switch {
			case broke:
				res.nocompile++
			case passed:
				s := survivor{rel: filepath.ToSlash(rel), line: c.line + 1, mutation: c.mutation, text: strings.TrimSpace(original)}
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
		restore()
	}
	return &res, nil
}

// selftest proves the tool can still both MUTATE and OBSERVE. Without it this joins the
// class of guards that silently stop guarding.
//
// The failure mode to fear is the FLATTERING one: a sweep that generates no mutants at all
// — a masking bug, an operator table that stopped matching, a file walk that found nothing
// — reports "0 survived / 0 behavioural" and reads as a clean bill of health. `killed < 1`
// is the assertion that catches that, which is why it is checked SEPARATELY from the
// survivor count rather than folded into it.
//
// (The opposite break is loud: if the mutation is never written to disk, every mutant
// passes and survivors INFLATE. Both directions were verified by breaking them on purpose.)
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

	res, err := sweep(dir, "", os.Stdout)
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
	flag.Parse()

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

	// The mutant lives in the working tree. A dirty target means a crash cannot be told
	// apart from your own edit, so refuse rather than risk it. This is not hypothetical:
	// during the pass that closed these survivors, a sweep was running while files were
	// staged, and the tree had to be checked mutant-by-mutant before committing was safe.
	dirty, err := exec.Command("git", "-C", root, "status", "--porcelain", "--", moduleDir).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot read git status:", err)
		os.Exit(1)
	}
	if len(strings.TrimSpace(string(dirty))) > 0 {
		fmt.Fprintf(os.Stderr, "refusing to sweep: %s has uncommitted changes.\n"+
			"This tool writes mutants into the working tree and restores them; an interrupted run\n"+
			"against a dirty tree cannot be told apart from your own edits. Commit or stash first.\n\n%s",
			*module, dirty)
		os.Exit(1)
	}

	res, err := sweep(moduleDir, *filter, os.Stdout)
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
}
