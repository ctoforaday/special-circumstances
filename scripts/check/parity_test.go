package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// THIS IS THE HALF THAT MATTERS.
//
// gates.go is a second declaration of CI's gate set, and a second declaration is worth
// nothing on its own — it is the prose validation loop rewritten in Go, and it drifts the
// same way and in the same direction, toward the shorter list. What stops that is a test
// that reads the workflow and fails when the two disagree.
//
// The repo answers drift this way twice already (pluginparity, mjsparity), both after the
// same class recurred. This is the third, aimed at the tooling itself.
//
// The comparison is deliberately COARSE — tool names and per-module step kinds, not exact
// argument strings. A test that pinned every flag would fail on formatting and get relaxed
// until it meant nothing. What must not drift is WHICH GATES EXIST.

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := gitx.Root()
	if err != nil {
		t.Skipf("not in a git repo: %v", err)
	}
	return root
}

func workflow(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), workflowPath))
	if err != nil {
		t.Fatalf("reading the workflow, which is the AUTHORITY this table mirrors: %v", err)
	}
	return string(b)
}

var (
	reGoRunTool = regexp.MustCompile(`go run \./([a-z]+)(?:\s+(-[a-z]+))?`)
	reNodeSuite = regexp.MustCompile(`(plugins/[^\s"']+\.test\.mjs)`)
)

// toolID is the identity of a tool invocation, and the distinction it draws is the one this
// comparison turns on.
//
// A MODE flag is part of the identity: `mutate -selftest` and `mutate` are different gates
// asking different questions, and losing one while keeping the other is exactly the drift
// being guarded. A PARAMETER is not: `-base origin/main` names the ref to compare against,
// its value differs per invocation, and the table already carries that fact in the
// `needsBase` FIELD rather than in a string. Folding a parameter into the identity would make
// the two sides disagree forever over a difference that is not one — which is what the first
// run of this test did, and the reason it is written down here.
func toolID(name, flag string) string {
	if flag == "" || flag == "-base" {
		return name
	}
	return name + " " + flag
}

// ciTools is every scripts/ tool the workflow invokes.
func ciTools(wf string) map[string]bool {
	got := map[string]bool{}
	for _, m := range reGoRunTool.FindAllStringSubmatch(wf, -1) {
		got[toolID(m[1], m[2])] = true
	}
	return got
}

// tableTools is the same set, derived from the gate table.
func tableTools() map[string]bool {
	got := map[string]bool{}
	for _, g := range gateSet() {
		if g.kind != kindTool {
			continue
		}
		// args are {"run", "./tool"} or {"run", "./tool", "-flag"}
		flag := ""
		if len(g.args) > 2 {
			flag = g.args[2]
		}
		got[toolID(strings.TrimPrefix(g.args[1], "./"), flag)] = true
	}
	return got
}

// EVERY tool CI runs must be declared here. This is the direction that actually bit: CI grew
// mjsparity, mutate -selftest and decisions, and the hand-kept list never learned about them.
func TestEveryToolCIRunsIsDeclared(t *testing.T) {
	ci, table := ciTools(workflow(t)), tableTools()
	// The release job shells `gh`, not `go run ./x`, so nothing from it lands in ci.
	var missing []string
	for name := range ci {
		if !table[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("CI runs %d tool(s) this table does not declare: %v\n"+
			"That is the drift this test exists for — add them to gates.go rather than "+
			"deleting this assertion.", len(missing), missing)
	}
}

// And the reverse: a gate declared here but no longer in CI is a gate that looks covered and
// is not. It fails in the quieter direction, which is why it gets its own test.
func TestEveryDeclaredToolIsStillRunByCI(t *testing.T) {
	ci, table := ciTools(workflow(t)), tableTools()
	var stale []string
	for name := range table {
		if !ci[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("this table declares %d tool(s) CI no longer runs: %v\n"+
			"Either CI lost a gate, or the table is describing coverage that does not exist.", len(stale), stale)
	}
}

// The four Go modules and their per-module steps. A module CI builds that this table omits
// means vet/fmt/test/race never run locally for it.
func TestEveryModuleCIBuildsIsDeclared(t *testing.T) {
	wf := workflow(t)
	declared := map[string]bool{}
	for _, m := range modules {
		declared[m.dir] = true
	}
	// working-directory lines naming a module path.
	re := regexp.MustCompile(`working-directory:\s+(plugins/[a-z-]+/tools|scripts)`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(wf, -1) {
		seen[m[1]] = true
	}
	if len(seen) == 0 {
		t.Fatal("no module working-directory found in the workflow — this test would pass " +
			"vacuously, which is the failure it is meant to catch elsewhere")
	}
	for dir := range seen {
		if !declared[dir] {
			t.Errorf("CI builds %s and the gate table does not declare it", dir)
		}
	}
}

// The -race scopes are NOT uniform across modules, and encoding them wrongly would claim
// coverage CI does not have. feov-record races only internal/record.
func TestRaceScopesMatchTheWorkflow(t *testing.T) {
	wf := workflow(t)
	if !strings.Contains(wf, "go test -race ./internal/record/") {
		t.Error("the workflow no longer races ./internal/record/ in feov-record; raceScope still claims it")
	}
	if scope := raceScope["plugins/frank-exchange-of-views/tools"]; len(scope) != 1 || scope[0] != "./internal/record/" {
		t.Errorf("raceScope for feov = %v, want exactly ./internal/record/ — a broader claim here "+
			"would report concurrency coverage the workflow does not run", scope)
	}
	if _, ok := raceScope["scripts"]; ok {
		t.Error("scripts has no -race leg in CI; declaring one here claims coverage that does not exist")
	}
}

// Node suites are enumerated in both places for the same Windows reason; mjsparity guards the
// workflow against the tree, and this guards the table against the workflow.
func TestNodeSuitesMatchTheWorkflow(t *testing.T) {
	wf := workflow(t)
	inCI := map[string]bool{}
	for _, m := range reNodeSuite.FindAllStringSubmatch(wf, -1) {
		inCI[m[1]] = true
	}
	if len(inCI) == 0 {
		t.Fatal("no node --test suite found in the workflow; if Node is gone this table must " +
			"be emptied deliberately rather than left describing suites nobody runs")
	}
	declared := map[string]bool{}
	for _, s := range nodeSuites {
		declared[s] = true
	}
	for s := range inCI {
		if !declared[s] {
			t.Errorf("CI runs %s and the gate table does not declare it", s)
		}
	}
	for s := range declared {
		if !inCI[s] {
			t.Errorf("the gate table declares %s and CI does not run it", s)
		}
	}
}

// A skipped gate must carry a REASON. "Not run here" with no explanation is how a gate stops
// being run and nobody notices it was ever meant to be.
func TestEverySkippedGateSaysWhy(t *testing.T) {
	for _, g := range gateSet() {
		if g.skip != "" && len(strings.TrimSpace(g.skip)) < 20 {
			t.Errorf("gate %q skips with a reason too thin to act on: %q", g.id, g.skip)
		}
		if g.why == "" {
			t.Errorf("gate %q has no `why`; a gate nobody can justify is one nobody will keep", g.id)
		}
	}
}

// Ids must be unique — -only matches on them, and two gates sharing an id means selecting one
// silently runs the other.
// EVERY SUITE GATE DEFEATS THE TEST CACHE, because the local run is the only one that has one.
//
// CI runs on a fresh runner with `cache: false`, so a stale PASS cannot happen there. Locally it
// can, and it did: `go test ./...` in the feov module reported the whole module green while the
// same package under `-count=1` failed on a help golden a commit three back had invalidated. The
// contract tests shell out to a built binary, so the inputs that changed are inputs the cache
// cannot see.
//
// Without this, `go run ./check` can pass where CI fails — which is the exact drift the command
// exists to remove, reproduced inside it.
func TestEverySuiteGateDefeatsTheTestCache(t *testing.T) {
	for _, g := range gateSet() {
		if g.kind != kindTest && g.kind != kindRace {
			continue
		}
		var found bool
		for _, a := range g.args {
			if a == noCache {
				found = true
			}
		}
		if !found {
			t.Errorf("gate %q runs %v without %s — a cached PASS is byte-identical to a real one, "+
				"and only the local run has a cache to be fooled by", g.id, g.args, noCache)
		}
	}
}

func TestGateIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range gateSet() {
		if seen[g.id] {
			t.Errorf("duplicate gate id %q", g.id)
		}
		seen[g.id] = true
	}
}

// -only must never silently match nothing; main exits 2 on that, and this pins the selector.
func TestSelectGatesMatchesAndReportsEmpty(t *testing.T) {
	gs := gateSet()
	if got := selectGates(gs, ""); len(got) != len(gs) {
		t.Errorf("empty -only selected %d, want all %d", len(got), len(gs))
	}
	if got := selectGates(gs, "golden"); len(got) != 1 || got[0].id != "golden" {
		t.Errorf("-only golden selected %v", got)
	}
	if got := selectGates(gs, "nosuchgate"); len(got) != 0 {
		t.Errorf("-only nosuchgate selected %d, want 0 so main can refuse", len(got))
	}
}
