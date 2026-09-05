package surface

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// hookBinaries are the binaries the plugin manifest registers as hooks. EVERY ONE of them is
// held to the same graph, and that uniformity is the point: the rule is PER-INVOCATION COST, and
// all three fire often enough to pay it.
//
// feov-pretooluse fires once per TOOL CALL — ~2,457 times a run. The sitting hooks fire on every
// SubagentStart and SubagentStop, which sounds rarer and is not: SubagentStop also fires at the
// MAIN AGENT'S TURN END, 50 times against 19 seats in one measured session (§7a), and it fires in
// EVERY session including every session with no run at all. A globally installed plugin paying a
// SQLite driver's init() on every turn end of every user's session is the same defect as #684 F2,
// on a different event.
//
// Measured while this was being decided: feov-subagentstart was 13.06 MB and 3.555 ms with the
// record linked, and 3.16 MB and 1.210 ms without — level with feov-pretooluse. The write moved to
// feov-sitting-write, which IS heavy and is NOT here, because the hooks spawn it only once they
// know there is a live run and a real seat: ~38 times in a run, never in an ordinary session.
//
// Nothing here applies to feov-record, which is a long-lived-enough invocation that its
// 398-package graph is not a cost worth guarding, or to feov-sitting-write for the same reason.
var hookBinaries = []string{
	"./cmd/feov-pretooluse",
	"./cmd/feov-subagentstart",
	"./cmd/feov-subagentstop",
}

// hookLocalPackages is the module-local import graph each registered hook binary is ALLOWED to
// have, PER BINARY rather than shared.
//
// A single shared list cannot state this: internal/sittinghook is legitimate for the sitting
// hooks and would be a dead entry for feov-pretooluse, and the staleness arm below — which is
// what stops the allowlist rotting into decoration — cannot tell a dead entry from a deliberate
// one. So each binary's graph is its own claim, and the repetition is the claim being made three
// times rather than one claim being made loosely.
//
// A WHITELIST, AND DELIBERATELY. facts-are-fields warns that a guard whose own allowlist is
// hand-kept has reproduced the defect one level up — that warning is about a guard whose miss is
// SILENT, and this one cannot miss: any package added to a hook's graph fails this test by name,
// whether or not anyone thought of it in advance. A blacklist of known-heavy packages is the
// shape that would go quietly stale, because the next expensive dependency is by definition the
// one nobody listed.
//
// Adding a line here is a real decision, not a formality: it means that hook now pays that
// package's init() on every firing of its event.
var hookLocalPackages = map[string][]string{
	"./cmd/feov-pretooluse": {
		"internal/feov", "internal/seatenv", "internal/hookgate", "internal/runlive", "internal/hookcmd",
	},
	"./cmd/feov-subagentstart": {
		"internal/feov", "internal/seatenv", "internal/hookgate", "internal/runlive", "internal/hookcmd",
		"internal/sittinghook",
	},
	"./cmd/feov-subagentstop": {
		"internal/feov", "internal/seatenv", "internal/hookgate", "internal/runlive", "internal/hookcmd",
		"internal/sittinghook",
	},
}

// hookForbidden are packages whose init() registers global state before main runs — a SQL driver
// registering itself, protobuf registering every message descriptor in the linked schema, cobra
// building a command tree. They are named individually because the whitelist above cannot see
// them: they arrive transitively through a module-local package that is already allowed, or
// directly, and neither shows up as a new module-local name.
//
// This half IS a hand-kept list, and it says why it could not be generated: "has an expensive
// init()" is not a property `go list` reports, and there is no derivation from the import graph
// that separates a package which registers a driver from one that declares a constant. The
// whitelist is the part that cannot go stale; this is the part that names what we already know
// hurt, so the failure message can say WHICH thing came back.
var hookForbidden = []string{
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record",
	"modernc.org/sqlite",
	"google.golang.org/protobuf/proto",
	"github.com/spf13/cobra",
}

// THE HOOK FIRES ONCE PER BASH CALL, AND ITS COST IS init(), NOT ITS WORK.
//
// Measured on an idle 4-core box, interleaved A/B so a co-tenant's load hits both arms: the hook
// went 4.361 ms -> 1.359 ms and 13.3 MB -> 3.08 MB when `InferRunDir` moved out of
// internal/cli/seat, against a 1.15 ms empty-binary floor. It was importing one function and
// inheriting cobra, protobuf and a SQLite driver behind it (#684 F2). None of that work ever ran
// — the driver registers itself before main and the hook never opens a database — which is
// exactly why nothing caught it: the hook was correct, tested, and three times slower than it
// needed to be, on ~2,457 invocations a run.
//
// THE GRAPH IS THE ASSERTION, NOT THE CLOCK. A timing test would be the honest thing to check
// and the wrong thing to gate: the first sequential measurement of this very change read 7.8 ms
// because another worktree's test binary was at 98% CPU. The import graph is the stable proxy
// for the cost and it is what a future change actually gets wrong.
func TestHookBinaryLinksNothingExpensive(t *testing.T) {
	for _, b := range hookBinaries {
		t.Run(path.Base(b), func(t *testing.T) { assertGraph(t, b) })
	}
}

func assertGraph(t *testing.T, hookBinary string) {
	if _, ok := hookLocalPackages[hookBinary]; !ok {
		t.Fatalf("%s is gated but has no allowlist, so every package it links would read as an "+
			"addition; state its graph in hookLocalPackages", hookBinary)
	}
	deps := hookDeps(t, hookBinary)

	local := map[string]bool{}
	for _, d := range deps {
		if rel, ok := strings.CutPrefix(d, modulePath+"/"); ok {
			local[rel] = true
		}
	}
	// Each binary's OWN cmd package is derived rather than listed: three entries in a shared
	// allowlist would each be stale for two of the three binaries, which is the same
	// "reads as coverage, checks nothing" failure the staleness arm below exists to catch.
	// The binary's OWN cmd package is derived rather than listed — it is true by construction.
	allowed := map[string]bool{strings.TrimPrefix(hookBinary, "./"): true}
	for _, p := range hookLocalPackages[hookBinary] {
		allowed[p] = true
		if !local[p] {
			t.Errorf("%s no longer imports %s. If that is deliberate, drop it from hookLocalPackages — "+
				"a stale allowlist entry checks nothing while reading as coverage", hookBinary, p)
		}
	}
	var added []string
	for p := range local {
		if !allowed[p] {
			added = append(added, p)
		}
	}
	sort.Strings(added)
	if len(added) > 0 {
		t.Errorf("%s gained module-local package(s) %s.\n\n"+
			"This binary runs once per Bash call in every session, and it pays every linked package's "+
			"init() before main. If the new dependency is genuinely needed, add it to hookLocalPackages "+
			"WITH the reason; if it is one function from a large package, move that function to a leaf "+
			"instead — which is what #684 F2 did with InferRunDir.", hookBinary, strings.Join(added, ", "))
	}

	for _, bad := range hookForbidden {
		for _, d := range deps {
			if d == bad {
				t.Errorf("%s links %s, whose init() runs before main on every invocation. "+
					"The hook resolves a run directory and rewrites a command string; it opens no record "+
					"and parses no protobuf. See #684 F2 — this cost 3 ms and 10 MB the last time.",
					hookBinary, bad)
			}
		}
	}
}

const modulePath = "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools"

// hookDeps asks the toolchain for the binary's transitive import graph.
//
// `go list` READS the module graph; it does not build or drive anything, which is why this gate
// belongs beside the other static agreements here rather than in integration/fuzz. It costs
// about as much as one compile of a small package, and it is the only way to ask this question:
// the answer lives in the build graph, not in any file a text sweep could read.
func hookDeps(t *testing.T, hookBinary string) []string {
	t.Helper()
	dir, err := repotree.Plugin("tools")
	if err != nil {
		t.Fatalf("locating the tools module: %v", err)
	}
	cmd := exec.Command("go", "list", "-deps", hookBinary)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// A gate that cannot ask its question must SAY so. Returning an empty graph here would
		// pass every assertion below and read exactly like a clean binary.
		t.Fatalf("go list -deps %s in %s: %v", hookBinary, dir, err)
	}
	deps := strings.Fields(string(out))
	if len(deps) == 0 {
		t.Fatalf("go list -deps %s returned nothing — the graph was not measured", hookBinary)
	}
	return deps
}

// EVERY HOOK BINARY IS EITHER GATED OR NAMED AS DELIBERATELY UNGATED.
//
// The gate above guards ONE binary by name. A fourth hook added later — firing per tool call, and
// linking whatever was convenient — would not fail it, because the gate never asks what the set
// of hook binaries IS. That is the plausible zero one level up: the guard passes because it is
// looking at a list that stopped being the whole list.
//
// So the set is derived from the plugin manifest, which is the thing that actually decides what
// runs, and every member must be accounted for in one of the two categories above.
func TestEveryRegisteredHookBinaryIsGatedOrNamedUngated(t *testing.T) {
	root, err := repotree.Root()
	if err != nil {
		t.Fatalf("cannot locate the repository root: %v", err)
	}
	manifest := filepath.Join(root, "plugins", "frank-exchange-of-views", "hooks", "hooks.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("cannot read the hook manifest, which is what decides which binaries run: %v", err)
	}

	// The command is a shell line naming ${CLAUDE_PLUGIN_ROOT}/bin/<binary>; the binary name is
	// what this needs and the rest is the bootstrap guard.
	found := map[string]bool{}
	for _, m := range regexp.MustCompile(`/bin/(feov-[a-z]+)`).FindAllStringSubmatch(string(raw), -1) {
		found[m[1]] = true
	}
	if len(found) == 0 {
		t.Fatal("no hook binaries found in the manifest — this test is reading the wrong thing and would pass forever")
	}

	accounted := map[string]bool{}
	for _, b := range hookBinaries {
		accounted[path.Base(b)] = true
	}
	for b := range found {
		if !accounted[b] {
			t.Errorf("hooks.json registers %q and hookBinaries does not list it, so its import graph is "+
				"unguarded. Every registered hook pays its init() on every firing of its event; add it "+
				"here, or move its expensive work into a binary the hook spawns only when it has "+
				"something to do (see feov-sitting-write).", b)
		}
	}
	for b := range accounted {
		if !found[b] {
			t.Errorf("hookBinaries lists %q and hooks.json registers no such binary — the gate is "+
				"guarding something that does not run", b)
		}
	}
}
