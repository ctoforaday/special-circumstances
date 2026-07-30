// Command mjs-suite-parity checks that CI's enumerated .mjs suite list matches the tree.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// The .mjs suites are ENUMERATED in .github/workflows/hooks.yml rather than globbed, because
// `node --test <dir>` does not work on Windows. Nothing checked that the enumeration matched
// the tree, and `node --test` EXITS 0 on a path that does not exist — it does not warn, and
// it does not count the file it could not find. So the list could name a deleted suite (CI
// stays green, testing nothing) or omit a live one (the suite never runs once), and both read
// exactly like a healthy build.
//
// Measured on 2026-07-30, before this guard existed: of four enumerated suites, TWO did not
// exist (read-batching.test.mjs, run-dashboard.test.mjs). That was the third instance of the
// class — record.test.mjs went unlisted for its whole life and never ran in CI once — and the
// previous two were fixed as instances, which is why it recurred.
//
// This guard outlives Node in the repo: it stays useful until the last .mjs suite is gone, and
// when that happens it must fail LOUDLY (no `node --test` step found) rather than pass
// silently, so retiring it is a decision somebody makes rather than a thing that drifts.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const workflowPath = ".github/workflows/hooks.yml"

// listedSuite matches a suite path on its own line in the workflow's `node --test` block.
var listedSuite = regexp.MustCompile(`(?m)^\s*(plugins/[^\s]+\.test\.mjs)\s*$`)

// onDiskSuites lists every tracked .mjs suite. The working corpus (ideas/, research/,
// projects/) is not swept: it is seeded at runtime and is not CI's business.
func onDiskSuites(root string) ([]string, error) {
	out, err := gitx.Run(root, "ls-files", "*.test.mjs")
	if err != nil {
		return nil, err
	}
	var kept []string
	for _, f := range gitx.Lines(out) {
		slash := filepath.ToSlash(f)
		if strings.Contains(slash, "node_modules") {
			continue
		}
		if strings.HasPrefix(slash, "plugins/") || strings.HasPrefix(slash, "scripts/") {
			kept = append(kept, slash)
		}
	}
	sort.Strings(kept)
	return kept, nil
}

// compare is the pure core: what CI lists versus what exists.
func compare(workflow string, want, have []string) []string {
	var problems []string
	fail := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }

	switch {
	case !strings.Contains(workflow, "node --test"):
		fail("%s: no `node --test` step found at all — this guard greps for one, so it has been passing without reading anything. If the suites moved to another runner, move this check with them.", workflowPath)
	case len(have) == 0:
		fail("%s: found `node --test` but could not extract a single suite path — the pattern this guard greps for has moved, so it has been passing without reading anything.", workflowPath)
	}

	for _, p := range have {
		if !contains(want, p) {
			fail("%s runs %s, which does not exist. `node --test` exits 0 on a missing path, so this has been counting as a pass while testing nothing.", workflowPath, p)
		}
	}
	for _, p := range want {
		if !contains(have, p) {
			fail("%s exists and is not in %s — it has never run in CI. Add it to the `node --test` list (enumerated, not globbed: `node --test <dir>` fails on Windows).", p, workflowPath)
		}
	}
	for i, p := range have {
		if i > 0 && have[i-1] == p {
			fail("%s lists %s more than once.", workflowPath, p)
		}
	}
	return problems
}

func listed(workflow string) []string {
	var have []string
	for _, m := range listedSuite.FindAllStringSubmatch(workflow, -1) {
		have = append(have, m[1])
	}
	sort.Strings(have)
	return have
}

func main() {
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "mjs-suite-parity:", err)
		os.Exit(1)
	}
	wf, err := os.ReadFile(filepath.Join(root, workflowPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "mjs-suite-parity: cannot read %s: %v\n", workflowPath, err)
		os.Exit(1)
	}
	want, err := onDiskSuites(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mjs-suite-parity:", err)
		os.Exit(1)
	}

	problems := compare(string(wf), want, listed(string(wf)))
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
	fmt.Printf("mjs suite list matches the tree: %d suite(s) — %s\n", len(want), strings.Join(want, ", "))
}

func contains(h []string, n string) bool {
	for _, x := range h {
		if x == n {
			return true
		}
	}
	return false
}
