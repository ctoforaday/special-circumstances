package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
	"github.com/ctoforaday/special-circumstances/scripts/internal/testsplit"
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

// workflow is every workflow that holds a gate, read as one text: hooks.yml, and the nightly
// whole-graph race beside it. A sweep that read only the first would hold the nightly's race leg
// to nothing.
func workflow(t *testing.T) string {
	t.Helper()
	return workflowFile(t, workflowPath) + "\n" + workflowFile(t, nightlyRacePath)
}

// workflowFile is ONE workflow, for an assertion about which of them carries a line.
func workflowFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), path))
	if err != nil {
		t.Fatalf("reading %s, which is the AUTHORITY this table mirrors: %v", path, err)
	}
	return string(b)
}

var (
	reGoRunTool = regexp.MustCompile(`go run \./([a-z/]+)(?:\s+(-[a-z]+))?`)
	reNodeSuite = regexp.MustCompile(`(plugins/[^\s"']+\.test\.mjs)`)
)

// toolID is the identity of a tool invocation, and the distinction it draws is the one this
// comparison turns on.
//
// A MODE flag is part of the identity: `protogen -check` and `protogen` are different gates
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

// ciTools is every dev tool the workflow invokes, by its package path: `agentgen` under scripts/,
// `devcmd/enginegen` under a plugin's tools module.
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
// mjsparity and decisions, and the hand-kept list never learned about them.
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
// coverage CI does not have. feov-record's scope is the shipped binary's import graph, and two
// workflows read it: the pull-request gate hands the target to racescope, which races the tests
// in that graph that start goroutines; the nightly races the graph whole. What this test holds
// together is the TARGET each derives from, and which workflow carries which form — the
// whole-graph line appearing in hooks.yml, or the tool line in the nightly, is the two having
// been swapped or copied.
func TestRaceScopesMatchTheWorkflow(t *testing.T) {
	wf := workflowFile(t, workflowPath)
	const target = "./cmd/feov-record"
	if scope := raceScope["plugins/frank-exchange-of-views/tools"]; len(scope) != 1 || scope[0] != depsScopePrefix+target {
		t.Errorf("raceScope for feov = %v, want exactly %s%s — any other claim here "+
			"would report concurrency coverage the workflow does not run", scope, depsScopePrefix, target)
	}
	var local gate
	for _, g := range gateSet() {
		if g.id == "feov-record:race" {
			local = g
		}
	}
	localLine := "go " + strings.Join(local.args, " ")
	if local.kind != kindTool || !strings.Contains(localLine, "go run ./racescope ") || !strings.Contains(localLine, " -deps "+target) {
		t.Errorf("the local feov-record:race gate runs %q, want racescope over %s", localLine, target)
	}
	feov := jobBlock(t, wf, "feov-record")
	var tool string
	for _, line := range strings.Split(feov, "\n") {
		if trimmed := strings.TrimSpace(line); !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "go run ./racescope ") {
			tool = trimmed
		}
	}
	switch {
	case tool == "":
		t.Error("the feov-record job no longer runs racescope; raceScope still claims its race gate is derived")
	case !strings.Contains(tool, " -deps "+target):
		t.Errorf("the feov-record job's racescope line derives from another target than %s: %q", target, tool)
	case !strings.Contains(tool, raceTimeout):
		t.Errorf("the feov-record job's racescope line lacks %s, so CI and the local gate diverge: %q", raceTimeout, tool)
	}
	if strings.Contains(feov, "xargs go test -race") {
		t.Error("the feov-record job races the whole import graph again; that form belongs to " + nightlyRacePath)
	}
	nightly := workflowFile(t, nightlyRacePath)
	if !strings.Contains(nightly, "go list -deps "+target) || !strings.Contains(nightly, "xargs go test -race") {
		t.Errorf("%s no longer races the whole import graph of %s, and it is the only run that does", nightlyRacePath, target)
	}
	if strings.Contains(nightly, "go run ./racescope") {
		t.Errorf("%s runs racescope; the nightly exists for what racescope cannot see", nightlyRacePath)
	}
	// scripts GAINED a -race leg when its concurrency became real: golden/interrupt.go guards
	// signal state with a mutex against its interrupt handler, which puts a file back. The
	// unsynchronised version of that shipped and was found by reading, because this was the one
	// module the detector never ran over.
	//
	// Scoped to the job's own block: `-race ./...` appears in several jobs, so a bare
	// Contains over the whole workflow would pass on somebody else's leg — which is the
	// shape of claim this whole test exists to refuse.
	// The command and its scope, not their adjacency: a flag between them (the race bound) must
	// not read as the leg having gone.
	if !regexp.MustCompile(`go test -race\b[^\n]*\s\./\.\.\.`).MatchString(jobBlock(t, wf, "scripts")) {
		t.Error("the scripts job no longer runs -race over ./..., but raceScope claims it does")
	}
	if scope := raceScope["scripts"]; len(scope) != 1 || scope[0] != "./..." {
		t.Errorf("raceScope for scripts = %v, want exactly ./...", scope)
	}
}

// jobBlock returns one job's YAML, so a workflow assertion cannot be satisfied by a step in
// a different job. Jobs are two-space keys under `jobs:`; the block runs to the next one.
func jobBlock(t *testing.T, wf, job string) string {
	t.Helper()
	start := strings.Index(wf, "\n  "+job+":\n")
	if start < 0 {
		t.Fatalf("job %q is not in the workflow at all", job)
	}
	rest := wf[start+1:]
	next := regexp.MustCompile(`\n  [a-z][a-z0-9-]*:\n`).FindStringIndex(rest[len(job)+2:])
	if next == nil {
		return rest
	}
	return rest[:len(job)+2+next[0]]
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
		// A suite run by scripts/testsplit shows no -count=1 in the gate's arguments: the tool
		// passes it on both of its paths, and its own tests hold it to that
		// (TestBothPathsDefeatTheTestCache). What is held HERE is that such a gate runs that
		// tool and nothing else.
		if g.kind == kindSuite {
			if len(g.args) < 2 || g.args[1] != "./testsplit" {
				t.Errorf("suite gate %q runs %v — a tool that is not testsplit, whose cache defeat nothing here vouches for", g.id, g.args)
			}
			continue
		}
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

// A SPLIT SUITE IS STILL THE WORKFLOW'S SUITE. A module with a row in testsplit's table has its
// suite run by that tool from `check` and from `golden`, both through testsplit.SuiteArgs. The
// workflow runs plain `go test ./...` over the same module instead — its runners have too few
// cores for splitting to pay — and the tool runs the whole of `./...` whatever it splits, so
// the two cover the same packages. What is held here: the job still runs that plain suite, and
// does not ALSO run the tool, which would be the suite twice on the slowest box.
func TestTheSplitSuiteIsTheWorkflowsSuite(t *testing.T) {
	wf := workflow(t)
	var held int
	for _, m := range modules {
		if !testsplit.Splits(m.dir) {
			continue
		}
		held++
		var suite, tool int
		for _, line := range strings.Split(jobBlock(t, wf, m.ciJob), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(trimmed, "go test ") && strings.HasSuffix(trimmed, " ./...") && !strings.Contains(trimmed, "-race") {
				suite++
			}
			if strings.Contains(trimmed, "./testsplit") {
				tool++
			}
		}
		if suite != 1 {
			t.Errorf("job %s runs plain `go test ... ./...` %d time(s), want once — it is the CI half of the local %s:test gate", m.ciJob, suite, m.ciJob)
		}
		if tool != 0 {
			t.Errorf("job %s runs testsplit; on a four-core runner that is slower than the plain suite it also runs", m.ciJob)
		}
	}
	if held == 0 {
		t.Fatal("no module has a row in testsplit's table, and this test measured nothing")
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
	if got, un := selectGates(gs, ""); len(got) != len(gs) || len(un) != 0 {
		t.Errorf("empty -only selected %d (unmatched %v), want all %d", len(got), un, len(gs))
	}
	if got, un := selectGates(gs, "golden"); len(got) != 1 || got[0].id != "golden" || len(un) != 0 {
		t.Errorf("-only golden selected %v, unmatched %v", got, un)
	}
	if got, un := selectGates(gs, "nosuchgate"); len(got) != 0 || len(un) != 1 || un[0] != "nosuchgate" {
		t.Errorf("-only nosuchgate selected %d, unmatched %v — want 0 and the term reported", len(got), un)
	}
}

// #999: a term that matched nothing must be reported even when ANOTHER term matched.
//
// Measured before the fix: `-only versionguard,no-such-gate` printed "1 gate(s): 1 passed, 0
// failed" and exited 0, so a typo'd gate name read exactly like a gate that passed. The real
// case was `-only rule-sweep,...` against a gate named `rulesweep`, on a release PR.
func TestSelectGatesReportsATypoBesideAMatch(t *testing.T) {
	gs := gateSet()
	got, un := selectGates(gs, "golden,rule-sweep")
	if len(un) != 1 || un[0] != "rule-sweep" {
		t.Errorf("unmatched = %v, want [rule-sweep] so main can refuse the whole run", un)
	}
	if len(got) != 1 || got[0].id != "golden" {
		t.Errorf("selected %v, want the one gate that did match", got)
	}
}

// One gate, one entry: two terms matching the same gate must not run it twice.
func TestSelectGatesDoesNotRepeatAGateMatchedTwice(t *testing.T) {
	gs := gateSet()
	got, un := selectGates(gs, "golden,golde")
	if len(un) != 0 {
		t.Fatalf("unmatched %v, want none", un)
	}
	if len(got) != 1 {
		t.Errorf("selected %d gates for two terms naming one gate, want 1", len(got))
	}
}

// A -race leg scoped to ONE TEST is still a gate, and the module-wide raceScope cannot hold it.
//
// #641 added `go test -race -run TestTheRunHandleIsImmutableAfterConstruction ./releasegate/fuzz/`
// to feov-record and declared it nowhere: raceScope is asserted to be EXACTLY ./internal/record/,
// so the new leg was invisible to this file — a CI gate with no local counterpart, which is the
// first thing this command's package comment names.
func TestTheNarrowRaceLegsAreDeclared(t *testing.T) {
	wf := workflow(t)
	for _, g := range narrowRace {
		var name string
		for i, a := range g.args {
			if a == "-run" && i+1 < len(g.args) {
				name = strings.Trim(g.args[i+1], "^$")
			}
		}
		if name == "" {
			t.Fatalf("%s declares no -run name; a narrow race leg that races everything is not narrow", g.id)
		}
		if !strings.Contains(jobBlock(t, wf, g.ciJob), name) {
			t.Errorf("%s claims the %s job runs %q under -race, and the workflow does not mention it",
				g.id, g.ciJob, name)
		}
	}
}

// A deps: scope reaches `go test` through nothing: the gate that carries one is racescope's, or
// is declared and skipped. One that is neither would hand the marker to go test as a package.
func TestADepsScopeIsNeverAGoTestArgument(t *testing.T) {
	var markers int
	for _, g := range gateSet() {
		for _, a := range g.args {
			if !strings.HasPrefix(a, depsScopePrefix) {
				continue
			}
			markers++
			if g.skip == "" {
				t.Errorf("gate %q runs here with the scope marker %s in its arguments %v", g.id, a, g.args)
			}
		}
	}
	if markers == 0 {
		t.Error("no gate carries a deps: scope at all — the whole-graph race is no longer declared")
	}
}

// THE RACE BOUND IS ONE NUMBER IN TWO PLACES, and both are held to it.
//
// Go's default ten-minute test timeout was catching a correct serial suite rather than a hang:
// feov-record's internal/cli took 435s under -race on an IDLE box and 776s under this box's
// usual load, turning the race gate red three times in one week with zero data races reported.
// raceTimeout fixes the local gate; the workflow carries the same bound. If one moves without
// the other, `check` and CI measure different things — the drift this command exists to remove.
func TestEveryModuleRaceLegCarriesTheBound(t *testing.T) {
	var bounded int
	for _, g := range gateSet() {
		// A module's gate, and the declared whole-graph run. The narrow single-test legs run
		// one test and need no suite bound.
		if !strings.HasSuffix(g.id, ":race") && !strings.HasSuffix(g.id, ":race-nightly") {
			continue
		}
		bounded++
		var found bool
		for _, a := range g.args {
			if a == raceTimeout {
				found = true
			}
		}
		if !found {
			t.Errorf("local race gate %q runs %v without %s", g.id, g.args, raceTimeout)
		}
	}
	if bounded != len(raceScope)+1 {
		t.Errorf("held %d race gates to the bound, want one per module in raceScope and the nightly (%d)", bounded, len(raceScope)+1)
	}
	// Every SUITE race leg in either workflow — `./...` or the xargs'd import graph — carries it
	// too. racescope's line is held to the bound by TestRaceScopesMatchTheWorkflow.
	// Whole lines, so a mention in a YAML comment — the workflow explains itself in prose — is
	// told apart from an invocation. The first version matched "`go test -race` with no packages
	// would" out of a comment and failed on it.
	var legs int
	for _, line := range strings.Split(workflow(t), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "go test -race") {
			continue
		}
		if strings.Contains(line, "-run ") {
			continue // a single named test, like the local narrow legs
		}
		legs++
		if !strings.Contains(line, raceTimeout) {
			t.Errorf("a workflow race leg lacks %s, so CI and the local gate diverge: %q", raceTimeout, trimmed)
		}
	}
	// A NO-MATCH MUST NOT READ AS A PASS. If the workflow's shape moves and this finds nothing, it
	// would report every leg bounded while checking none.
	if legs == 0 {
		t.Fatal("found no suite race legs in the workflow at all — the check measured nothing")
	}
}
