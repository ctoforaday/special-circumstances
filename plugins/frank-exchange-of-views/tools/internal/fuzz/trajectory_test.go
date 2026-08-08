package fuzz

import (
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
)

// EVERY EXECUTION THROUGH THE HARNESS IS TRACKED, THE WAY A TRAJECTORY IS.
//
// The first version of this coverage work kept a hand-written list of surfaces to check.
// That is the same shape as the defect it was fixing: a list nothing derives from the code
// goes stale silently, and the sweep then reports coverage of a surface it never saw.
//
// So the harness tallies EVERY invocation it makes — argv[0..1] resolved to a command path
// — and the completeness question is answered by diffing that observed set against
// cli.CommandPaths(), which walks the REAL cobra tree. Nothing hand-maintained states what
// exists; the only hand-written list left is the EXEMPTIONS, and each carries its reason.
//
// The tally is also the diagnostic the fuzz never had: which surfaces run how often, across
// arbitrary run shapes, is the same question a live run's trajectory answers about seats.

var (
	execMu   sync.Mutex
	execRuns = map[string]int{} // command path -> invocations
	execFail = map[string]int{} // command path -> non-zero exits
)

// noteExec records one invocation. Called from BOTH paths that shell the binary — the
// runner's exec and the direct exec.Command oracles — so nothing reaches the tool untracked.
func noteExec(args []string, err error) {
	execMu.Lock()
	defer execMu.Unlock()
	k := commandPathOf(args)
	if k == "" {
		return
	}
	execRuns[k]++
	if err != nil {
		execFail[k]++
	}
	// FLAGS AND THEIR VALUES, per command path. The tally saw argv all along and threw the
	// flags away, so "the surface is driven" meant the VERB ran — a verb can be exercised on
	// every sweep with half its flags never passed and every enum value but one unreached.
	for i, a := range args {
		if !strings.HasPrefix(a, "--") {
			continue
		}
		name := strings.TrimPrefix(strings.SplitN(a, "=", 2)[0], "--")
		if execFlags[k] == nil {
			execFlags[k] = map[string]bool{}
		}
		execFlags[k][name] = true
		// The VALUE, when the next token is one rather than another flag. Enum coverage is
		// what this buys: a closed set whose values are never all driven is a set whose
		// unreached members nothing has ever run.
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			if execValues[name] == nil {
				execValues[name] = map[string]bool{}
			}
			execValues[name][args[i+1]] = true
		}
	}
}

// commandPathOf resolves an argv to the command path it invoked: "merge mint", "verify".
// Flags end the path, so `graph --format dot` is "graph" and `blue show --view x` is
// "blue show".
// execFlags: command path -> the flag names ever passed to it.
// execValues: flag name -> the values ever passed for it.
var (
	execFlags  = map[string]map[string]bool{}
	execValues = map[string]map[string]bool{}
)

func commandPathOf(args []string) string {
	var parts []string
	for _, a := range args {
		// SKIP leading flags rather than stopping at them: `--json merge mint …` is a real
		// mint, and stopping made it invisible to the tally — which then reported the most
		// heavily driven verb in the suite as never invoked.
		if strings.HasPrefix(a, "-") {
			if len(parts) == 0 {
				continue
			}
			break
		}
		parts = append(parts, a)
		if len(parts) == 2 {
			break
		}
	}
	if len(parts) == 0 {
		return ""
	}
	// A one-word argv is a root command; two words are only a path if the first is a role.
	if len(parts) == 2 && !isRole(parts[0]) {
		parts = parts[:1]
	}
	return strings.Join(parts, " ")
}

func isRole(s string) bool {
	switch s {
	case "lens", "merge", "blue", "bench":
		return true
	}
	return false
}

// tracked wraps a direct binary call so the oracles' invocations are tallied too.
func tracked(bin string, args ...string) ([]byte, error) {
	out, err := exec.Command(bin, args...).CombinedOutput()
	noteExec(args, err)
	return out, err
}

// exemptSurfaces are the command paths the fuzz deliberately does NOT drive, each with the
// reason. This is the ONLY hand-written list left, and it says why rather than merely
// omitting — an absence with no stated reason is indistinguishable from an oversight, which
// is exactly how the 18 undriven verbs survived.
var exemptSurfaces = map[string]string{
	"setup":   "CREATES a run; this harness builds its run dir directly, so driving setup would fuzz a different thing (internal/setup has its own tests)",
	"capture": "needs a workflow transcript directory the harness has no analogue for (internal/capture has its own tests)",
	"hook":    "reads a JSON hook payload on stdin rather than argv (internal/cli hook tests cover it)",
	"bench halt": "TERMINAL — a halt ends the run and reshapes every downstream oracle, so it is " +
		"deliberately kept out of the random sweep and driven by TestFuzzHaltPath instead",
}

// unreachedSurfaces returns the command paths the tree exposes that no run ever invoked.
func unreachedSurfaces() []string {
	execMu.Lock()
	defer execMu.Unlock()
	var out []string
	for _, path := range cli.CommandPaths() {
		if execRuns[path] > 0 {
			continue
		}
		// `hook` and friends have subcommands, so the tree yields "hook pretooluse";
		// exempt by first word too.
		head := strings.SplitN(path, " ", 2)[0]
		if exemptSurfaces[path] != "" || exemptSurfaces[head] != "" {
			continue
		}
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

// execReport renders the tally for the sweep's log — the fuzz's own trajectory.
func execReport() string {
	execMu.Lock()
	defer execMu.Unlock()
	type row struct {
		path string
		n, f int
	}
	rows := make([]row, 0, len(execRuns))
	for p, n := range execRuns {
		rows = append(rows, row{p, n, execFail[p]})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].n > rows[j].n })
	var b strings.Builder
	total := 0
	for _, r := range rows {
		total += r.n
	}
	b.WriteString("command surface exercised: ")
	b.WriteString(strings.TrimSuffix(strings.Join([]string{}, ""), ""))
	for i, r := range rows {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(r.path)
		b.WriteString("=")
		b.WriteString(itoa(r.n))
		if r.f > 0 {
			b.WriteString("(" + itoa(r.f) + " refused)")
		}
	}
	return b.String() + "  [" + itoa(len(rows)) + " paths, " + itoa(total) + " invocations]"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}
