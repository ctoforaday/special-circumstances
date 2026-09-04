// Command check runs THE GATE SET — every check CI runs — from one place, and holds the
// gate set as DATA rather than as prose somebody retypes.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # Why this exists, measured
//
// The gate set lived in two places: `.github/workflows/hooks.yml`, which is authoritative,
// and a hand-written validation loop in a session checkpoint, which is what actually got run
// during development. Nothing reconciled them, and the copy drifted in the direction that
// always wins — toward the shorter list:
//
//	mjsparity          in CI, absent from the loop
//	mutate -selftest   in CI, absent from the loop
//	decisions          in CI, absent from the loop
//	node --test suites in CI, absent from the loop
//	go vet   × 4       in CI, absent from the loop
//	gofmt    × 4       in CI, absent from the loop
//	-race    × 3       in CI, one module in the loop
//	qlty               in CI, absent from the loop
//	versionguard       in CI, absent until a PR went red on it (2026-08-15)
//
// Roughly half the gates. Two failures came out of that gap in one week: a pull request that
// went red on a gate with no local counterpart, and a goldens entry that recorded the
// REGENERATE command as the check — a check that rewrites its own expectations and therefore
// could never fail.
//
// This is `facts-are-fields` pointed at our own toolchain. The gate set is a record with
// fields; recovering it from prose is a hope about a document nobody validates.
//
// # The half that is load-bearing
//
// A second declaration is worth nothing on its own — it is the same prose duplicate rewritten
// in Go, and it drifts the same way. What makes this hold is parity_test.go: it parses the
// workflow and FAILS when CI names a gate this table does not, or the table names one CI does
// not run. The repo already answers drift that way twice (pluginparity, mjsparity), both after
// the same class recurred.
//
// # What "declared" means, and why a skip is not an absence
//
// Some gates cannot run on a developer machine — `qlty` needs a binary that is not a
// dependency here, and `release` only fires on a tag. Those are DECLARED with a stated reason
// and skipped. Declared-and-skipped is honest and visible; absent is the bug. A gate that
// cannot run must say so in the report rather than quietly not appear in it.
//
// Usage:
//
//	go run ./check              run every gate that can run here
//	go run ./check -list        print the gate set and where each one runs
//	go run ./check -only a,b    run just these (substring match on the gate id)
//	go run ./check -base main   the ref diff-based gates compare against (default origin/main)
package main

// workflowPath is the authority FOR THE GATE SET, and the scoping matters: it is not the only
// workflow in the repository. `.github/workflows/labels.yml` runs on `issues` events, marks an
// issue carrying no `area:`/`priority:` label, and is deliberately absent from this table — it
// gates nothing, cannot run locally, and has no merge to block (docs/issue-labels.md says why).
// A gate that cannot run here is DECLARED and skipped; a workflow that is not a gate is neither.
//
// This table is the mirror of hooks.yml, and parity_test.go is what keeps the mirror honest —
// see the package comment.
const workflowPath = ".github/workflows/hooks.yml"

// kind groups a gate by what it is, so the report can be read at a glance and so parity can
// compare like with like.
type kind string

const (
	kindVet     kind = "vet"     // go vet over a module
	kindFmt     kind = "fmt"     // gofmt -l must be empty over a module
	kindTest    kind = "test"    // go test over a module
	kindRace    kind = "race"    // go test -race over some subset of a module
	kindTool    kind = "tool"    // a scripts/ tool, run via go run
	kindNode    kind = "node"    // a node --test suite
	kindExtern  kind = "extern"  // needs a binary this repo does not vendor
	kindRelease kind = "release" // only meaningful on a tag
)

// gate is one check. The fields are the record; nothing here is recovered from a string.
//
// `id` is stable and is what -only matches and what the report prints. `dir` is relative to
// the repository root. `needsBase` marks the gates that compare against a ref, because those
// are the ones a stale origin/main makes pass locally and fail in CI — measured, twice.
type gate struct {
	id        string
	kind      kind
	dir       string
	args      []string
	needsBase bool
	// skip, when non-empty, is why this gate cannot run here. It is REPORTED, never hidden:
	// a gate that silently vanishes from the run is indistinguishable from one that passed.
	skip string
	// ciJob is the workflow job that runs it. Parity is checked per job, so a gate moving
	// between jobs is a change somebody makes deliberately rather than one that slips.
	ciJob string
	why   string
}

// noCache is the difference between running the suite and reading a memory of it, and it is
// the one flag here that CI does NOT pass — deliberately, because CI cannot need it.
//
// A GREEN CACHE AND A GREEN SUITE ARE THE SAME BYTES. Go caches a test package's PASS keyed on
// the inputs it can see, and a test that shells out to a built binary — which is what the help
// and golden contracts do — has inputs the cache cannot see. Change the help text in
// `internal/cli`, and `go test ./internal/difftest/` still answers `ok (cached)`.
//
// Measured 2026-08-16, and it is why this flag exists rather than being an abundance of caution:
// `go test ./...` in `plugins/frank-exchange-of-views/tools` reported the whole module green,
// exit 0. The same package under `-count=1` FAILED on a stale help golden that a commit three
// back had invalidated. The green was reported to a human as verification.
//
// CI never sees this because every job runs on a fresh runner with an empty cache and
// `cache: false` on setup-go — so a stale PASS is structurally impossible there and structurally
// available here. That asymmetry is precisely the drift this whole command was built to remove;
// leaving it in place would mean the local gate set can pass where CI fails, which is the
// original defect wearing this tool's name. parity_test.go compares gate IDENTITY (which module,
// which race scope, which tool), not argv, so declaring it here claims no coverage CI lacks.
const noCache = "-count=1"

// modules are the four Go modules. Enumerated rather than globbed for the same reason CI
// enumerates them: a glob that matches nothing is a silent pass.
var modules = []struct{ dir, ciJob string }{
	{"plugins/prosthetic-conscience/tools", "run"},
	{"plugins/frank-exchange-of-views/tools", "feov-record"},
	{"plugins/gray-area/tools", "gray-area"},
	{"scripts", "scripts"},
}

// raceScope is what each job passes to `go test -race`. NOT uniform, and the difference is
// deliberate. Encoding the real scopes means this table cannot claim coverage the workflow
// does not have.
//
// feov-record races the SHIPPED BINARY'S IMPORT GRAPH and nothing else (ruled 2026-09-03: no
// race-instrumenting the test harnesses). The scope is a deps: marker expanded at run time —
// by expandScope here and by `go list -deps | grep` in the workflow — because the graph is
// the record and a hand-copied package list beside it would be free to drift. What the
// module's harness packages (releasegate/fuzz, testbuild, difftest) do with their
// concurrency is SPAWN the real binary, which no race detector observes from the parent.
// This is also the module's slowest gate now: the graph includes fetchcache and cli, whose
// suites are minutes, not seconds.
var raceScope = map[string][]string{
	"plugins/prosthetic-conscience/tools":   {"./..."},
	"plugins/frank-exchange-of-views/tools": {depsScopePrefix + "./cmd/feov-record"},
	"plugins/gray-area/tools":               {"./..."},
	// scripts WAS the one module with no -race leg, absent on purpose while it was
	// straight-line tooling. It stopped being that: golden/interrupt.go guards signal state
	// with a mutex and mutate arms a file for an interrupt handler to put back — and the
	// second of those shipped an unsynchronised read of the path/bytes pair, so an interrupt
	// could pair one file's path with another file's contents and write it over real source.
	// It was found by reading, not by CI, because CI could not look. `./...` rather than a
	// package list: a hand-kept list of "the concurrent ones" is the next thing to go stale,
	// and this module's suite is seconds.
	"scripts": {"./..."},
}

// depsScopePrefix marks a race scope that is a binary's module-local import graph, resolved
// in the gate's own directory when the gate runs. See expandScope.
const depsScopePrefix = "deps:"

// narrowRace are -race legs scoped to ONE NAMED TEST rather than to a package.
//
// Deliberately not entries in raceScope: that map is module -> scope and the parity test holds
// it to exactly what CI races, so putting "./releasegate/fuzz/" there would claim the whole
// package is raced when one test is. Under-claiming is the other half of the same defect,
// which is why these are declared here instead of left out — a CI gate with no local
// counterpart is the drift this command's package comment opens with.
//
// The run-handle guard is the HARNESS CARVE-OUT from the scope rule above, kept by ruling
// ("keep fuzz on the hook", #686). A harness race cannot ship, but it makes fuzz verdicts
// flaky, and a flaky verdict is a measurement defect; one dedicated ~1s test keeps the
// harness accountable without paying the race tax on the whole package.
var narrowRace = []gate{
	{id: "feov-record:race-runhandle", kind: kindRace, dir: "plugins/frank-exchange-of-views/tools",
		args:  []string{"test", "-race", noCache, "-run", "^TestTheRunHandleIsImmutableAfterConstruction$", "./releasegate/fuzz/"},
		ciJob: "feov-record",
		why:   "the fuzz runner's seats are concurrent; its run handle was briefly resolved lazily, and two full debate runs under -race did not report it"},
}

// tools are the scripts/ commands CI runs.
var tools = []gate{
	{id: "rulesweep", kind: kindTool, dir: "scripts", args: []string{"run", "./rulesweep"}, needsBase: true, ciJob: "rule-sweep",
		why: "a protocol-surface change must name a registry class"},
	{id: "versionguard", kind: kindTool, dir: "scripts", args: []string{"run", "./versionguard"}, needsBase: true, ciJob: "rule-sweep",
		why: "a version that goes BACKWARDS ships content under a number a consumer already has"},
	{id: "archaeology", kind: kindTool, dir: "scripts", args: []string{"run", "./archaeology"}, needsBase: true, ciJob: "rule-sweep",
		why: "prose about a thing that no longer exists costs every seat context it cannot act on"},
	{id: "archaeology-selftest", kind: kindTool, dir: "scripts", args: []string{"run", "./archaeology", "-selftest"}, ciJob: "rule-sweep",
		why: "a matcher whose patterns went stale reports zero findings, which is a clean diff's bytes"},
	{id: "stackguard", kind: kindTool, dir: "scripts", args: []string{"run", "./stackguard"}, needsBase: true, ciJob: "stackguard",
		why: "a stacked pull request merged into an already-merged base lands in a branch nobody reads (#633)"},
	// The sweep is a SECOND declaration rather than a flag on the first, because it is a
	// different question asked at a different moment: -base checks this pull request at open
	// time, -sweep checks every open pull request each time main moves. Folding them would let
	// the table claim the push-side coverage while CI ran only the pull-request leg.
	{id: "stackguard-sweep", kind: kindTool, dir: "scripts", args: []string{"run", "./stackguard", "-sweep"}, ciJob: "stackguard",
		skip: "reads OPEN pull requests over the API; it is a push-to-main gate, and there is no local push to main",
		why:  "the base of an open stack can be merged AFTER that stack's checks went green — #633's actual timeline"},
	{id: "golden", kind: kindTool, dir: "scripts", args: []string{"run", "./golden"}, ciJob: "goldens",
		why: "a stale golden is an unrecorded behaviour change"},
	{id: "mjsparity", kind: kindTool, dir: "scripts", args: []string{"run", "./mjsparity"}, ciJob: "debate-sim",
		why: "node --test exits 0 on a path that does not exist"},
	{id: "mutate-selftest", kind: kindTool, dir: "scripts", args: []string{"run", "./mutate", "-selftest"}, ciJob: "debate-sim",
		why: "a mutation sweep that cannot mutate flatters the suite"},
	// Declared and SKIPPED, like the release gates it runs beside: the sweep takes minutes
	// to hours and only a tag decides which module it judges. Running it here would sweep
	// the default module on every check invocation for a verdict nothing local consumes.
	{id: "mutate-gate", kind: kindTool, dir: "scripts", args: []string{"run", "./mutate", "-gate"}, ciJob: "release",
		skip: "only meaningful on a tag; the release job sweeps the tagged plugin's module",
		why:  "a release must prove every mutation survivor was JUDGED, not merely counted"},
	{id: "validatejson", kind: kindTool, dir: "scripts", args: []string{"run", "./validatejson"}, ciJob: "debate-sim",
		why: "manifests break silently"},
	{id: "frontmatter", kind: kindTool, dir: "scripts", args: []string{"run", "./frontmatter"}, ciJob: "debate-sim",
		why: "a broken block drops ALL its fields, silently"},
	{id: "decisions", kind: kindTool, dir: "scripts", args: []string{"run", "./decisions"}, ciJob: "debate-sim",
		why: "every recorded decision names the issue that tracks it"},
	{id: "pluginparity", kind: kindTool, dir: "scripts", args: []string{"run", "./pluginparity"}, ciJob: "debate-sim",
		why: "every marketplace plugin is installed by both bootstrap paths"},
	{id: "fixtureparity", kind: kindTool, dir: "scripts", args: []string{"run", "./fixtureparity"}, ciJob: "debate-sim",
		why: "a file shape two modules must agree on drifted once and nothing failed"},
	{id: "classgen", kind: kindTool, dir: "scripts", args: []string{"run", "./classgen", "-check"}, ciJob: "debate-sim",
		why: "fixtures must stage the vocabulary the plugin actually ships"},
	{id: "buildidgen", kind: kindTool, dir: "scripts", args: []string{"run", "./buildidgen", "-check"}, ciJob: "debate-sim",
		why: "one authored buildid, copied into modules that cannot import it"},
	{id: "schemagen", kind: kindTool, dir: "scripts", args: []string{"run", "./schemagen", "-check"}, ciJob: "debate-sim",
		why: "the binary and the plugin must read the same event-schema epoch"},
}

// nodeSuites are the .mjs suites CI drives directly. Enumerated in the workflow because
// `node --test <dir>` fails on Windows; mjsparity is the guard that this list matches the tree.
var nodeSuites = []string{
	"plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs",
	"plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs",
}

// gateSet builds the full set. One function so there is exactly one answer to "what are the
// gates", and both the runner and the parity test read it.
func gateSet() []gate {
	var gs []gate
	for _, m := range modules {
		gs = append(gs,
			gate{id: m.ciJob + ":vet", kind: kindVet, dir: m.dir, args: []string{"vet", "./..."}, ciJob: m.ciJob,
				why: "vet catches what compiles and cannot be right"},
			gate{id: m.ciJob + ":fmt", kind: kindFmt, dir: m.dir, ciJob: m.ciJob,
				why: "unformatted Go keeps reaching CI (#352)"},
			gate{id: m.ciJob + ":test", kind: kindTest, dir: m.dir, args: []string{"test", noCache, "./..."}, ciJob: m.ciJob,
				why: "the suite"},
		)
		if scope, ok := raceScope[m.dir]; ok {
			gs = append(gs, gate{id: m.ciJob + ":race", kind: kindRace, dir: m.dir,
				args: append([]string{"test", "-race", noCache}, scope...), ciJob: m.ciJob,
				why: "the concurrency guards are the ones that fail in production, not in a suite"})
		}
	}
	gs = append(gs, narrowRace...)
	gs = append(gs, tools...)
	for _, s := range nodeSuites {
		gs = append(gs, gate{id: "node:" + baseName(s), kind: kindNode, dir: ".", args: []string{"--test", s}, ciJob: "debate-sim",
			why: "the zero-token regression suite for the debate engine"})
	}
	gs = append(gs,
		gate{id: "qlty", kind: kindExtern, dir: ".", ciJob: "qlty", needsBase: true,
			skip: "qlty is not a dependency of this repo; CI installs it. Run it there, or install qlty locally.",
			why:  "lint and static analysis against the merge base"},
		gate{id: "bootstrap-guard", kind: kindExtern, dir: ".", ciJob: "debate-sim",
			skip: "an inline shell block in the workflow, not a tool this can invoke",
			why:  "hook bootstrap guards must degrade gracefully in the empty-bin window"},
		gate{id: "release", kind: kindRelease, dir: ".", ciJob: "release",
			skip: "only meaningful on a tag",
			why:  "cross-compile and publish"},
		// The feov sweeps' release leg. Declared for the reason every release gate is: a
		// gate that cannot run here must still appear in the report. Locally the same
		// sweeps run with FEOV_RELEASE_GATE=1 against releasegate/fuzz.
		gate{id: "release-fuzz-sweep", kind: kindRelease, dir: "plugins/frank-exchange-of-views/tools", ciJob: "release",
			skip: "only meaningful on a feov tag; the release job runs releasegate/fuzz with FEOV_RELEASE_GATE=1",
			why:  "the debate sweeps hold at the boundary where the binary ships"},
		// The RELEASE BOUNDARY invariant (#405). Declared and skipped for the same reason as
		// `release` itself — it runs in the release job, on a tag, before anything is published
		// — but declared, because a gate that cannot run here must still appear in the report.
		// The parity test caught its absence the moment CI grew it, which is what that test is
		// for.
		gate{id: "release-tag-matches-manifest", kind: kindTool, dir: "scripts",
			args: []string{"run", "./versionguard", "-tag"}, ciJob: "release",
			skip: "needs the tag it validates; the release job runs it on a tag before publishing",
			why:  "a tag and the manifest it publishes cannot disagree — one release act writes both"},
	)
	return gs
}

func baseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}
