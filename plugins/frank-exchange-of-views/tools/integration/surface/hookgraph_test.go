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

// hookBinary is the ONE binary in this module that runs once per TOOL CALL. Everything below is
// about keeping its process image small; nothing here applies to feov-record, which is a
// long-lived-enough invocation that its 398-package graph is not a cost worth guarding.
//
// THERE ARE NOW THREE HOOK BINARIES AND ONLY THIS ONE IS GATED, which is a decision rather than
// an oversight, and the arithmetic is the whole of it. feov-subagentstart and feov-subagentstop
// fire ONCE PER SEAT — thirteen to twenty times in a run — against this one's ~2,457 invocations.
// At the 4.361 ms image this gate exists to prevent, that is ~86 ms across a run whose seats span
// ~36 minutes each. They deliberately link internal/record, because writing an event to the
// record is the entire job (#265): a sitting hook that could not open a record would have to put
// its fact in a file beside the record instead, which is the shape facts-are-fields refuses.
//
// The rule is therefore PER-INVOCATION COST, not "hooks are light". If a future hook fires per
// tool call, it belongs in this gate; if it fires per seat, the graph is not what to guard.
const hookBinary = "./cmd/feov-pretooluse"

// perSeatHookBinaries fire once per seat and are NOT graph-gated, listed so "why is only one of
// three here" is answered where the question is asked.
var perSeatHookBinaries = []string{"./cmd/feov-subagentstart", "./cmd/feov-subagentstop"}

// hookLocalPackages is the module-local import graph feov-pretooluse is ALLOWED to have.
//
// A WHITELIST, AND DELIBERATELY. facts-are-fields warns that a guard whose own allowlist is
// hand-kept has reproduced the defect one level up — that warning is about a guard whose miss is
// SILENT, and this one cannot miss: any package added to the hook's graph fails this test by
// name, whether or not anyone thought of it in advance. A blacklist of known-heavy packages is
// the shape that would go quietly stale, because the next expensive dependency is by definition
// the one nobody listed.
//
// Adding a line here is a real decision, not a formality: it means the hook now pays that
// package's init() on every Bash call in every session.
var hookLocalPackages = []string{
	"cmd/feov-pretooluse",
	"internal/feov",
	"internal/hookcmd",
	"internal/hookgate",
	"internal/runlive",
	"internal/seatenv",
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
	deps := hookDeps(t)

	local := map[string]bool{}
	for _, d := range deps {
		if rel, ok := strings.CutPrefix(d, modulePath+"/"); ok {
			local[rel] = true
		}
	}
	allowed := map[string]bool{}
	for _, p := range hookLocalPackages {
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
func hookDeps(t *testing.T) []string {
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

	accounted := map[string]bool{path.Base(hookBinary): true}
	for _, b := range perSeatHookBinaries {
		accounted[path.Base(b)] = true
	}
	for b := range found {
		if !accounted[b] {
			t.Errorf("hooks.json registers %q and this file neither gates it nor names it as deliberately "+
				"ungated. Decide which: a hook firing per TOOL CALL belongs in hookBinary's gate; one firing "+
				"per seat belongs in perSeatHookBinaries with the arithmetic that says why.", b)
		}
	}
	for b := range accounted {
		if !found[b] {
			t.Errorf("this file accounts for %q and hooks.json registers no such binary — the gate is "+
				"guarding something that does not run", b)
		}
	}
}
