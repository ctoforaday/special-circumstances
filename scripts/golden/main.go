// Package main implements golden: ONE command for every golden in the repo, both
// languages.
//
// Dev tooling for this repository only (CLAUDE.md: dev concerns live at the root,
// consumer behaviour lives in plugins/). Nothing here ships to an installing project.
//
// THE PROBLEM THIS SOLVES is not diffing — go-cmp already does that — it is REVIEW
// EROSION. Golden suites rot in a predictable way: goldens spread across suites and
// languages, a wave regenerates some of them by hand, the diff is hundreds of lines, it
// gets rubber-stamped, and a real behavioural drift ships inside the noise. Prompt goldens
// per seat class make that pressure worse, because seat prompts are long and every wave
// edits them.
//
// Three countermeasures, all here:
//  1. ONE command updates everything, so nobody hand-picks which goldens to refresh and
//     no suite silently falls behind.
//  2. An update run prints a CHANGE REPORT — what moved, and by how much — so the author
//     sees the shape of the change before committing it.
//  3. A plain run fails when goldens are stale, so an unrecorded behaviour change cannot
//     merge quietly. That is what CI runs.
//
// # Two claims the .mjs header made that this one does not
//
// It advertised "four countermeasures, all here". Only three were here.
//
//  1. ORPHAN DETECTION — "a golden no test claims is a failure" — is real, but it does not
//     live in the orchestrator and never did. It is implemented in the SUITE helper,
//     tests/simulator/golden.mjs, as orphanGoldens() over a `claimed` set, and
//     prompts.test.mjs invokes it from an after() hook. Two different files named
//     golden.mjs, one claiming the other's feature. The check is genuinely running; the
//     orchestrator simply is not the thing running it, so this header does not say it is.
//     (The distinction is not pedantic: anyone porting or deleting the orchestrator on the
//     strength of that sentence would have been looking for code that was never there.)
//
//  2. `--check`, documented as "compare, exit 1 on any staleness", was NEVER PARSED. Its
//     effect was right by accident — a plain run already exits non-zero on staleness — so
//     the flag was decorative. It is dropped rather than carried over as a no-op alias.
//
// Orphan detection stays where it works: in the suite that knows which goldens it claimed.
// An orchestrator cannot compute that from outside, because Go golden tests name their
// files dynamically, so "claimed by a test" is not statically decidable in general.
//
// Usage:
//
//	go run ./golden             compare (what CI runs)
//	go run ./golden -review     verify, show each proposed golden diff, accept nothing
//	go run ./golden -review -accept a.golden,b.golden    keep those, revert the rest
//	go run ./golden -review -accept-all -reason "..."    keep all, with the reason stated
//	go run ./golden -update     re-record everything, then print the change report
//
// PREFER -review. `-update` rewrites every golden without ever showing you a failing test,
// which is how a semantics regression rides in inside a format change — see review.go for
// the measured case where the regenerate command had been written down as the check itself.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
	"github.com/ctoforaday/special-circumstances/scripts/internal/goldenmods"
)

// mjsSuites are the .mjs suites that carry goldens. Enumerated, never globbed: `node
// --test` over a directory fails on Windows (a documented recurrence), and an explicit
// list is also the thing CI copies.
var mjsSuites = []string{
	"plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs",
}

// goModules are the Go modules whose suites carry goldens. The list itself lives in
// internal/goldenmods so `check` reads the same record when deciding what its own test
// gates subsume (#626); this var stays swappable for the tests that fake a module.
var goModules = goldenmods.Modules

// subsumed maps a module dir to the check gate that already ran its exact go leg in the
// same invocation. Set only by -subsume on a plain verify run; a subsumed leg is REPORTED,
// never silently absent — a leg that vanishes from the output is indistinguishable from
// one that passed.
var subsumed = map[string]string{}

// runLegs drives every golden-carrying suite in both languages and reports whether any
// failed. Extracted so -review can run the SAME legs twice — once to verify, once to
// re-verify after accepting — instead of a second code path that could drift from the one
// CI runs.
func runLegs(root string, update bool) bool {
	suffix := ""
	if update {
		suffix = " (recording)"
	}
	failed := false

	// ---- mjs ----
	var present []string
	for _, s := range mjsSuites {
		if _, err := os.Stat(filepath.Join(root, s)); err == nil {
			present = append(present, s)
		}
	}
	if len(present) > 0 {
		node, err := exec.LookPath("node")
		if err != nil {
			// Absent node is machine state, not a stale golden. Saying which is which
			// keeps a missing toolchain from reading as a behavioural change.
			fmt.Fprintf(os.Stderr, "\n── mjs goldens: node not on PATH — cannot verify %d suite(s)\n", len(present))
			failed = true
		} else if run("mjs goldens"+suffix, root, node, append([]string{"--test"}, present...), update) {
			failed = true
		}
	} else {
		fmt.Print("\n── mjs goldens: no suites yet (skipped)\n")
	}

	// ---- go ----
	for _, mod := range goModules {
		// A SUBSUMED LEG IS STATED, NOT SKIPPED SILENTLY. check passes -subsume only for a
		// module whose `:test` gate runs this exact command in the same invocation (#626) —
		// and never on an update run, where the gate does not record.
		if by, ok := subsumed[mod]; ok && !update {
			fmt.Printf("\n── go goldens in %s: not re-run — %s runs `go test -count=1 ./...` here in this same check\n", mod, by)
			continue
		}
		// -count=1 defeats Go's TEST CACHE. Without it an update run can be satisfied
		// from cache — the leg reports ok, writes nothing, and the goldens on disk stay
		// stale while the command says "recorded". That happened: CI, which has no cache,
		// caught goldens this runner had just claimed to update. A check satisfiable
		// without doing the work is not a check.
		if run(fmt.Sprintf("go goldens in %s%s", mod, suffix), filepath.Join(root, mod),
			"go", []string{"test", "-count=1", "./..."}, update) {
			failed = true
		}
	}
	return failed
}

// run executes one leg, streaming its output, and reports whether it failed.
func run(label, dir, name string, args []string, update bool) bool {
	fmt.Printf("\n── %s\n", label)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "UPDATE_GOLDENS="+map[bool]string{true: "1", false: ""}[update])
	// Start/Wait rather than Run so the leg is TRACKED: a signal arriving mid-review has to be
	// able to end the child, or the main flow sits waiting on a process writing into a pipe
	// nobody is reading. See interrupt.go.
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "golden: could not start %s: %v\n", name, err)
		return true
	}
	trackLeg(cmd)
	err := cmd.Wait()
	clearLeg()
	return err != nil
}

func main() {
	update := flag.Bool("update", false, "re-record goldens, then print the change report")
	review := flag.Bool("review", false, "verify first, then show each proposed golden diff; accepts nothing unless told to")
	accept := flag.String("accept", "", "with -review: comma-separated goldens to keep (bare filename is enough); everything else is reverted")
	acceptAll := flag.Bool("accept-all", false, "with -review: keep every proposal; requires -reason")
	reason := flag.String("reason", "", "with -accept-all: why this whole batch is a deliberate contract change")
	subsume := flag.String("subsume", "", "comma-separated module=gate pairs whose go leg the named gate already ran in this same check invocation (passed by scripts/check, #626)")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "golden:", err)
		os.Exit(1)
	}
	watchSignals()

	// Verify runs only. An update or review run must drive every leg itself — the gate that
	// "already ran this" ran it without UPDATE_GOLDENS, so it recorded nothing.
	if *subsume != "" && (*update || *review) {
		fmt.Fprintln(os.Stderr, "golden: -subsume is a verify-run optimization; -update and -review must drive every leg themselves.")
		os.Exit(2)
	}

	if *review {
		if *update {
			fmt.Fprintln(os.Stderr, "golden: -review and -update are different postures — -review verifies before it rewrites, -update does not. Pick one.")
			os.Exit(2)
		}
		os.Exit(reviewMode(root, *accept, *acceptAll, *reason))
	}
	if *accept != "" || *acceptAll || *reason != "" {
		fmt.Fprintln(os.Stderr, "golden: -accept/-accept-all/-reason only mean anything with -review; without it nothing would be reviewed before being accepted.")
		os.Exit(2)
	}
	if *subsume != "" {
		for _, pair := range strings.Split(*subsume, ",") {
			mod, by, ok := strings.Cut(strings.TrimSpace(pair), "=")
			if !ok || mod == "" || by == "" {
				fmt.Fprintf(os.Stderr, "golden: -subsume %q — want module=gate pairs\n", pair)
				os.Exit(2)
			}
			subsumed[mod] = by
		}
	}

	failed := runLegs(root, *update)

	// ---- change report: git is the review surface, so ask git what moved ----
	if *update {
		fmt.Print("\n══ CHANGE REPORT ══\n")
		fmt.Print(changeReport(root))
	}

	if failed {
		fmt.Fprint(os.Stderr, "\ngolden: FAILED — goldens are stale or a test errored.\n"+
			"If the change is intentional: go run ./golden -update\n")
		os.Exit(1)
	}
	fmt.Printf("\ngolden: OK%s\n", map[bool]string{true: " (recorded)", false: ""}[*update])
}

// changeReport describes what moved, for a human about to commit it.
//
// `git diff` cannot see a golden that did not exist before, and a NEW golden is exactly
// what a new seat class or scenario produces — so modified and untracked are asked for
// separately, or the report under-reports on the runs that matter most.
func changeReport(root string) string {
	diff, diffErr := gitx.Run(root, "diff", "--stat", "--", "*.golden")
	untracked, listErr := gitx.Run(root, "ls-files", "--others", "--exclude-standard", "--", "*.golden")
	if diffErr != nil || listErr != nil {
		// Never fail the run over the report: the goldens are already recorded by here,
		// and losing the summary must not look like losing the recording.
		return fmt.Sprintf("could not read the golden diff (%v; %v) — the goldens ARE recorded; run `git status` to review\n", diffErr, listErr)
	}

	added := gitx.Lines(untracked)
	var parts []string
	if diff != "" {
		parts = append(parts, diff)
	}
	if len(added) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d new golden(s):\n", len(added))
		for _, f := range added {
			fmt.Fprintf(&b, "  + %s\n", f)
		}
		parts = append(parts, strings.TrimRight(b.String(), "\n"))
	}
	if len(parts) == 0 {
		return "no golden changed — behaviour is unmoved\n"
	}
	return strings.Join(parts, "\n") + "\n" +
		"\nCommit the testdata change ON ITS OWN, separately from the code that caused it.\n" +
		"A golden diff mixed into a feature commit is a diff nobody reads.\n"
}
