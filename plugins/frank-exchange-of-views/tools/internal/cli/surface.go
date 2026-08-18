package cli

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
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
	walk(newRoot(), "")
	sort.Strings(out)
	return out
}

// NewRootForTest builds the real command tree for the gates that must walk it exactly as a seat
// meets it — the same tree Execute runs, not a reconstruction.
//
// Exported for the same reason CommandPaths is: a gate that asks "can a seat discover this" has to
// ask the tree, and a second tree built for testing answers about itself.
func NewRootForTest() *cobra.Command { return newRoot() }
