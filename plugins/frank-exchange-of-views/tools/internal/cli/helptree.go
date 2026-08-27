package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// HELP IS A RUN INPUT, NOT A DISCOVERY TASK.
//
// MEASURED across the 2026-08-23 programme (#593): roughly a third of ALL record-tool traffic in
// both runs was --help surface-walking. Run A, 198 help calls against ~410 record-writing
// invocations; run B, 209 help-ish of 622 total, with usage screens printed 394 times. The worst
// seats spent more calls learning the surface than using it — judge-terminal 21 help against 16
// writes, red-merge-r1 25 of 54. By round 4 the same seats' successors needed about three: the
// knowledge exists by then, it just dies with each seat.
//
// The constitutions are right to demand the whole-tree read before deciding — deciding first was
// the measured failure that duty was written for. What was missing is that nothing PERSISTED the
// result, so every seat paid the same tax to learn the same surface.
//
// So setup renders it once, per role, into inputs/. GENERATED, never hand-written: this is the
// [[facts-are-fields]] preference for generating the derived carrier over guarding two copies of
// it. A staged tree cannot drift from the binary because it IS the binary's output, taken at the
// moment the run's other inputs are pinned.

// HelpTreeFor renders one role's ENTIRE command surface as markdown — the same pages `--help`
// prints, in one document, in the order a seat meets them.
//
// It builds the role's real tree via NewRootFor, so a verb that exists is here and a verb that
// does not cannot be. The alternative — a maintained document describing the surface — is the
// shape this repository keeps deleting.
func HelpTreeFor(role string) (string, error) {
	seatID := record.SampleSeatOf(role)
	if seatID == "" {
		return "", fmt.Errorf("helptree: %q is not a role with a seat namespace (roles: %s)",
			role, strings.Join(record.SeatRoles(), ", "))
	}
	root := NewRootFor(seatID)
	if root == nil {
		return "", fmt.Errorf("helptree: no command tree for role %q", role)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# `%s` — the %s seat's whole surface\n\n", InvokedAs(), role)
	b.WriteString("GENERATED AT RUN SETUP from the record binary this run will call, so it cannot " +
		"describe a surface the binary does not have. Read this ONCE, before you decide what to do — " +
		"that duty has not changed. What has changed is that you no longer walk the tree to satisfy " +
		"it: this IS the walk, already done.\n\n")
	b.WriteString("Per-command `--help` is still there and still worth running at USE time, when you " +
		"are about to pass a flag and want its exact wording. It is the detail; this is the map.\n\n")
	// THE MENU FIRST, THEN THE PAGES. A seat deciding what to do needs the discriminator — what
	// each verb is FOR — and that is what Short carries (help_test.go holds Short to exactly that:
	// no flags, no mechanics). The full pages below are for the verb it has already chosen. Put
	// the pages first and the map is 70KB down the file, which is the same problem the tree walk
	// had: the answer exists and is expensive to reach.
	b.WriteString("## The whole surface, one line each\n\n")
	for _, c := range invocable(root) {
		fmt.Fprintf(&b, "- `%s` — %s\n", c.CommandPath(), c.Short)
	}
	b.WriteString("\n---\n\n## Every page in full\n\n")

	var walk func(c *cobra.Command, depth int) error
	walk = func(c *cobra.Command, depth int) error {
		path := c.CommandPath()
		fmt.Fprintf(&b, "%s `%s`\n\n", strings.Repeat("#", min(depth+2, 6)), path)
		var page bytes.Buffer
		c.SetOut(&page)
		c.SetErr(&page)
		if err := c.Help(); err != nil {
			return fmt.Errorf("helptree: %s: %w", path, err)
		}
		b.WriteString("```\n")
		b.WriteString(strings.TrimRight(page.String(), "\n"))
		b.WriteString("\n```\n\n")

		for _, k := range children(c) {
			if err := walk(k, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, 0); err != nil {
		return "", err
	}
	return b.String(), nil
}

// children are c's real subcommands, sorted — cobra's own scaffolding (`help`, `completion`) is
// not surface a seat can run, and staging it would teach two verbs that do not exist in the role.
func children(c *cobra.Command) []*cobra.Command {
	kids := []*cobra.Command{}
	for _, k := range c.Commands() {
		if !k.IsAvailableCommand() || k.Name() == "help" || k.Name() == "completion" {
			continue
		}
		kids = append(kids, k)
	}
	sort.Slice(kids, func(i, j int) bool { return kids[i].Name() < kids[j].Name() })
	return kids
}

// invocable flattens the tree to the commands a seat can actually RUN — leaves, plus a group
// whose BARE form is itself a capability.
//
// RUNNABLE IS THE WRONG TEST, and CommandPaths already records why: the role groups and the
// motion subjects are runnable too, and all they do is REFUSE ("a verb is required, pick one
// below"). The first cut of this menu used `RunE != nil` and duly offered `motion grade` beside
// `motion grade file` — a line whose only capability is teaching the three lines under it. So the
// rule here is the one the tree already carries, the BareIsACapability annotation, and the two
// walks agree because they ask the same question rather than because someone kept them in step.
func invocable(root *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		kids := children(c)
		if c.CommandPath() != root.CommandPath() {
			if len(kids) == 0 || c.Annotations[BareIsACapability] == "yes" {
				out = append(out, c)
			}
		}
		for _, k := range kids {
			walk(k)
		}
	}
	walk(root)
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
