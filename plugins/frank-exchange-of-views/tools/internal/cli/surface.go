package cli

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// CommandPaths returns every invocable command path in the real tree — "verify",
// "merge mint", "blue claim-index" — sorted, with cobra's own scaffolding removed.
//
// WHY IT IS EXPORTED. The fuzz harness asserts that it drives the whole surface, and the
// only honest source for "the whole surface" is the tree itself. A hand-maintained list of
// what exists is how the gap this exists to close was created: a verb ships, nobody adds it
// to the list, and the sweep reports full coverage of a surface it never saw. Measured
// 2026-08-04 — 18 of 44 seat verbs and 7 of 9 root commands were undriven, and every
// undriven seat verb recorded nothing, so no event-based gate could have noticed.
//
// The vocabulary gate walks this same tree for the same reason: a single source of truth
// that nothing reads is a comment.
// ViewNames returns the real projection vocabulary — the names `show --view` accepts.
//
// THE READ SURFACE IS A SURFACE TOO. CommandPaths covers what a seat can RUN; a projection
// name is what a seat can READ, and it drifts the same way: the prompt-verb gate would not
// have caught `--view telemetry` in a constitution, because that is a flag VALUE and not a
// verb. A named view that does not exist fails the way every unmediated fact in this engine
// fails — the seat is told to read something, the tool refuses, and the seat logs friction
// and works around it, so the capability is simply absent for the run.
func ViewNames() []string { return seat.ViewNames() }

// CommandFlags returns every invocable command path with the flag names it declares.
//
// The companion to CommandPaths, and for the same reason: a hand-kept list of what a verb
// accepts is how a flag comes to be never exercised behind a green sweep. CommandPaths asks
// "did the verb run"; this asks "was its surface reached". A verb driven on every sweep can
// have half its flags never passed, and an unpassed flag is code no run has executed.
func CommandFlags() map[string][]string {
	out := map[string][]string{}
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			name := sub.Name()
			if name == "help" || name == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + name)
			if sub.HasSubCommands() {
				walk(sub, path)
				continue
			}
			var fs []string
			sub.Flags().VisitAll(func(f *pflag.Flag) { fs = append(fs, f.Name) })
			out[path] = fs
		}
	}
	walk(newRoot(), "")
	return out
}

// BareIsACapability marks a GROUP whose bare invocation does something, as opposed to one whose
// bare invocation only teaches. `show` returns the seat's pending work; `blue` and `motion grade`
// answer "a verb is required". Only the first is a command path.
const BareIsACapability = "bare-is-a-capability"

func CommandPaths() []string {
	var out []string
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			name := sub.Name()
			// cobra's own, not ours — they are not part of the contract a seat learns.
			if name == "help" || name == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + name)
			if sub.HasSubCommands() {
				walk(sub, path)
				// A GROUP WHOSE BARE FORM IS A CAPABILITY IS ALSO A PATH. `show` became a
				// group answering with the seat's pending work, and this walk collected only
				// leaves — so the prompt gate reported `merge show`, which every constitution
				// names and every seat runs, as a verb that "does not exist in the command tree".
				//
				// RUNNABLE ALONE IS THE WRONG TEST, and the coverage gate said so within a
				// minute: the role groups and the motion subjects are runnable too, and all they
				// do is REFUSE — "a verb is required, pick one below". Counting those as paths
				// demands the sweep drive a teaching message as though it were a capability.
				if sub.Runnable() && sub.Annotations[BareIsACapability] == "yes" {
					out = append(out, path)
				}
				continue
			}
			out = append(out, path)
		}
	}
	// EVERY SEAT'S TREE, PLUS THE OPERATOR'S. No single root holds the whole surface any more:
	// the tree is scoped to the dispatched role, so `newRoot()` with no seat is the operator's.
	//
	// The role is re-composed into the key HERE and nowhere a seat can see it. `blue edit` is not
	// something anyone types — a blue seat types `edit` — but the identity of a command IS
	// (role, verb), because `closing`, `position`, `friction`, `register` and `show` each exist
	// under several roles with different contracts. This function's output is a JOIN KEY for the
	// trigger map and the gates, not an invocation. The motion subtree keeps its bare path: one
	// motion is one object however many seats' trees it appears in.
	seen := map[string]bool{}
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		var roleOut []string
		out = nil
		walk(NewRootFor(dispatchedSeatFor(role)), "")
		roleOut, out = out, nil
		for _, p := range roleOut {
			key := role + " " + p
			// ONE COMMAND MOUNTED IN SEVERAL TREES KEEPS ONE KEY. `motion` is one object however
			// many seats can reach it, and `fetch`/`count-claims` are the operator commands seats
			// genuinely run — a lens reads blue's cached bytes, blue's claim_count is defined as
			// what count-claims prints. Keying them per role would turn one contract into four
			// rows nobody can keep in step.
			if strings.HasPrefix(p, "motion") || p == "fetch" || p == "count-claims" {
				key = p
			}
			if !seen[key] {
				seen[key] = true
			}
		}
	}
	walk(NewRootFor(""), "")
	for _, p := range out {
		seen[p] = true
	}
	out = out[:0]
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// AllRoots is every tree the tool can build: one per seat role, plus the operator's.
//
// NO SINGLE ROOT HOLDS THE WHOLE SURFACE any more — the tree is scoped to the dispatched identity
// — so anything asking a question ABOUT THE SURFACE has to ask it of all of them. A gate that kept
// walking one root would go on passing while it saw a quarter of the tool, which is the shape this
// repository keeps paying for.
func AllRoots() map[string]*cobra.Command {
	out := map[string]*cobra.Command{}
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		out[role] = NewRootFor(dispatchedSeatFor(role))
	}
	out[record.OperatorRole] = NewRootFor(record.OperatorRole)
	return out
}

// dispatchedSeatFor is a seat id of the given role, for building that role's tree in the gates and
// the surface walk. It is derived from the role tables rather than written down here.
func dispatchedSeatFor(role string) string { return record.SampleSeatOf(role) }

// NewRootForTest builds the real command tree for the gates that must walk it exactly as a seat
// meets it — the same tree Execute runs, not a reconstruction.
//
// Exported for the same reason CommandPaths is: a gate that asks "can a seat discover this" has to
// ask the tree, and a second tree built for testing answers about itself.
func NewRootForTest() *cobra.Command { return newRoot() }
