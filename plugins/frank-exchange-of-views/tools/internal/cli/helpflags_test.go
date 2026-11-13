package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// EVERY FLAG A HELP PAGE NAMES MUST BE A FLAG SOMETHING REGISTERS.
//
// The help page is the seat's whole teacher, and a page naming a flag no command takes sends the
// seat that obeys it into cobra's unknown-flag refusal. `show report`'s lens page named
// `--assembled` after the assembled set became the operator's own verb, and a smoke seat typed it.
//
// The census is the union of every tree's registered flags, the same cheap half
// TestNoRefusalNamesAFlagThatDoesNotExist takes, and for the same reason: a page legitimately
// names other verbs' flags, across roles too — the lens's `show source` page says which sentence
// blue's `cite --source-span` takes, the operator's `ops log` names the seats' `log --type`. A name
// NO tree registers needs no judgement about which sentence it sits in.
func TestNoHelpPageNamesAFlagThatDoesNotExist(t *testing.T) {
	roots := AllRoots()
	known := map[string]bool{"help": true}
	register := func(c *cobra.Command) {
		for _, fs := range []*pflag.FlagSet{c.Flags(), c.PersistentFlags(), c.InheritedFlags()} {
			fs.VisitAll(func(f *pflag.Flag) { known[f.Name] = true })
		}
	}
	for _, root := range roots {
		register(root)
		walk(root, func(c *cobra.Command, _ []string) { register(c) })
	}
	if len(known) < 20 {
		t.Fatalf("only %d flags registered across every tree — the walk is not reaching them, and an empty census refuses every page", len(known))
	}

	var pages, named int
	for _, role := range SurfaceRoles() {
		walk(roots[role], func(c *cobra.Command, path []string) {
			// THE PAGE AS THE SEAT READS IT: the role-scoped long text, the generated document and
			// cobra's flag block, rendered by the same help function `--help` runs.
			var page bytes.Buffer
			c.SetOut(&page)
			c.SetErr(&page)
			if err := c.Help(); err != nil {
				t.Fatalf("%s %s: rendering --help: %v", role, strings.Join(path, " "), err)
			}
			pages++
			for _, m := range flagToken.FindAllStringSubmatch(page.String(), -1) {
				named++
				if !known[m[1]] {
					t.Errorf("%s's `%s --help` names `--%s`, which NO command in any tree registers. "+
						"A seat that obeys it is refused an unknown flag. The page names only what exists: "+
						"drop the name, or name the verb that now carries the concept.",
						role, strings.Join(path, " "), m[1])
				}
			}
		})
	}
	// A FLOOR ON WHAT WAS READ: a render that comes back empty finds no names and passes forever.
	if pages < 50 || named < 100 {
		t.Fatalf("read %d pages naming %d flags — the render is not reaching the help text", pages, named)
	}
}
