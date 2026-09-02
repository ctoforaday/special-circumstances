package fuzz

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	// execFailWhy keeps the FIRST refusal per path, verbatim. A path refused on every
	// invocation is a coverage hole the count alone cannot diagnose — "26 refused" sends a
	// reader hunting, "26 refused: --seat-id IS REQUIRED HERE" names the fix. The refusal
	// text is what the tool said; r.exec already attaches the tool's output to the error.
	execFailWhy = map[string]string{}
	// execEdges is the OBSERVED GRAPH: seat id -> command path -> invocations.
	//
	// #535 step 1. execRuns above is keyed on the path alone, so the seat that ran it is thrown
	// away at the moment it is known — and "the surface was driven" is a different claim from
	// "this seat can reach this verb". The second is what says whether a verb nobody has ever run
	// is unreachable in principle or merely unused, which is the question #525's census could not
	// answer and #535 exists to make answerable.
	//
	// Keyed on the seat rather than the role, and the operator's own calls are kept under their
	// seat id too: an edge only the harness walks is still an edge, and hiding it would make the
	// graph flatter than the run.
	execEdges = map[string]map[string]int{}
)

// noteExec records one invocation. Called from BOTH paths that shell the binary — the
// runner's exec and the direct exec.Command oracles — so nothing reaches the tool untracked.
func noteExec(args []string, err error, out []byte) {
	execMu.Lock()
	defer execMu.Unlock()
	k := commandPathOf(args)
	if k == "" {
		return
	}
	execRuns[k]++
	if seat := seatOfArgs(args); seat != "" {
		if execEdges[seat] == nil {
			execEdges[seat] = map[string]int{}
		}
		execEdges[seat][k]++
	}
	if err != nil {
		execFail[k]++
		if _, seen := execFailWhy[k]; !seen {
			execFailWhy[k] = refusalDigest(out, err)
		}
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

// seatOfArgs recovers the acting SEAT ID from an invocation — "blue-respond-r3", not "blue".
//
// roleOfArgs below collapses that to a role, which is what the parity oracles want. The graph
// wants the seat: "blue ran edit" is true of a lane and of a respond seat four rounds apart, and
// a graph that cannot tell them apart cannot say which sitting reached a verb.
func seatOfArgs(args []string) string {
	for i, a := range args {
		switch {
		case a == "--seat-id" && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(a, "--seat-id="):
			return strings.TrimPrefix(a, "--seat-id=")
		}
	}
	return ""
}

// roleOfArgs recovers the acting role from the --seat-id in an invocation, which is the only
// place it appears now.
func roleOfArgs(args []string) string {
	for i, a := range args {
		var id string
		switch {
		case a == "--seat-id" && i+1 < len(args):
			id = args[i+1]
		case strings.HasPrefix(a, "--seat-id="):
			id = strings.TrimPrefix(a, "--seat-id=")
		default:
			continue
		}
		if r := cli.RoleOfSeat(id); r != "" && r != record.OperatorRole {
			return r
		}
	}
	return ""
}

func commandPathOf(args []string) string {
	// THE LONGEST ARGV PREFIX THAT IS A REAL COMMAND PATH, asked of the cobra tree.
	//
	// This used to take at most TWO tokens and require the first to be a role — a two-level tree
	// hard-coded into the tally. The command-groups work makes the tree deeper (`motion grade
	// file` is three), and under the old rule that collapsed to `motion`, matched no real path,
	// and reported all seven motion verbs as NEVER INVOKED while the record showed 35 motion
	// events. A tally that cannot see a path it was given reports a false absence, which is the
	// same plausible-zero shape the gate exists to catch — one level up, in the gate itself.
	//
	// Asking the tree removes the assumption entirely: any depth works, and a path the tree does
	// not have is not a path.
	var toks []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			if len(toks) == 0 {
				continue // `--json merge mint …` is a real mint; skip leading globals
			}
			break
		}
		toks = append(toks, a)
	}
	// THE ROLE IS PUT BACK ON, FROM THE SEAT ID. A seat no longer TYPES its role — the tree is
	// scoped to the injected identity — but cli.CommandPaths still keys a command by (role, verb),
	// because `closing`, `position`, `friction`, `register` and `show` each exist under several
	// roles with different contracts. Comparing role-less argv against role-keyed paths reported
	// every seat verb as never invoked: the plausible-zero shape this tally exists to catch,
	// arriving in the tally.
	known := knownPaths()
	role := roleOfArgs(args)
	for n := len(toks); n > 0; n-- {
		bare := strings.Join(toks[:n], " ")
		if role != "" && known[role+" "+bare] {
			return role + " " + bare
		}
		if known[bare] {
			return bare
		}
	}
	if len(toks) > 0 {
		return toks[0] // an unknown root command still counts as invoked, and as itself
	}
	return ""
}

var (
	knownOnce  sync.Once
	knownCache map[string]bool
)

func knownPaths() map[string]bool {
	knownOnce.Do(func() {
		knownCache = map[string]bool{}
		for _, p := range cli.CommandPaths() {
			knownCache[p] = true
		}
	})
	return knownCache
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
	noteExec(args, err, out)
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

// refusalDigest is what the TOOL said, not what the harness ran.
//
// It takes the combined output rather than the error, because `exec.ExitError` renders as
// "exit status 2" and the refusal — the whole reason the caller is looking — is in the output
// beside it. That distance is the one r.exec's own comment was written about: "six defects this
// session were a discarded refusal, and the cost was always the distance between the refusal and
// the symptom." Capped, because one cobra help page would swamp the report.
func refusalDigest(out []byte, err error) string {
	for _, ln := range strings.Split(string(out), "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			if len(ln) > 160 {
				return ln[:160] + "…"
			}
			return ln
		}
	}
	return err.Error()
}

// refusalExpected names paths the sweep drives ON PURPOSE to be refused, each with its reason.
// Empty today: every entry here is a claim that a verb's success path does not need exercising,
// and that claim should be argued in review rather than inherited.
var refusalExpected = map[string]string{}

// neverSucceeded returns the paths the sweep INVOKED and never once got a zero exit from.
//
// DRIVING A VERB IS NOT EXERCISING IT, and this repository has now learned that three times by
// hand. `merge carry` was "20 of 20 refused, and the coverage line read as a driven verb"; then
// it "fired ONCE across 60 runs — and that one was refused, so the verb's real coverage was zero
// while the tally read 1" (see the drive's own comment). Both were found by a person reading the
// tally, and both were fixed as instances. The count and the failure count sat side by side in
// execReport the whole time, and nothing compared them.
//
// The full-surface gate asks "was this path invoked". That question is satisfied by a path that
// has never done anything but be refused — the plausible zero this sweep exists to detect,
// wearing the sweep's own coverage line.
func neverSucceeded() []string {
	execMu.Lock()
	defer execMu.Unlock()
	var out []string
	for p, n := range execRuns {
		if n == 0 || execFail[p] < n {
			continue
		}
		if why := refusalExpected[p]; why != "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s — %d of %d invocations refused, never once succeeded; first refusal: %s",
			p, execFail[p], n, execFailWhy[p]))
	}
	sort.Strings(out)
	return out
}
