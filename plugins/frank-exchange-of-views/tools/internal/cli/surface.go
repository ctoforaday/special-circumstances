package cli

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
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
				continue
			}
			out = append(out, path)
		}
	}
	walk(newRoot(), "")
	sort.Strings(out)
	return out
}
