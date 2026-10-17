package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// A MOTION SUBJECT ON A SURFACE OFFERS AN ACT, for every role's tree rather than for the subjects
// that exist today.

// roleCommands visits every command of every role's tree, groups included.
func roleCommands(visit func(role string, path []string, c *cobra.Command)) {
	roots := AllRoots()
	roles := make([]string, 0, len(roots))
	for r := range roots {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	var walk func(role string, path []string, c *cobra.Command)
	walk = func(role string, path []string, c *cobra.Command) {
		visit(role, path, c)
		for _, sub := range c.Commands() {
			walk(role, append(append([]string{}, path...), sub.Name()), sub)
		}
	}
	for _, role := range roles {
		walk(role, nil, roots[role])
	}
}

// No motion subject on a role's tree is a group with nothing under it.
func TestEveryMotionSubjectOnASurfaceHasAVerb(t *testing.T) {
	var seen, motionGroups int
	roleCommands(func(role string, path []string, c *cobra.Command) {
		seen++
		at := role + " " + strings.Join(path, " ")
		// A SUBJECT WITH NO VERB FOR THIS ROLE IS NOT ON ITS SURFACE. An inquiry motion has a ruling
		// and nothing else, so only its ruler's tree holds the group; mounted bare on the others it
		// is a page that offers no act.
		if len(path) == 2 && path[0] == "motion" {
			motionGroups++
			if !c.HasSubCommands() {
				t.Errorf("%s is a motion subject with no verb on this surface", at)
			}
		}
	})
	if seen < 100 || motionGroups < 12 {
		t.Fatalf("the walk visited %d commands and %d motion subjects; it has stopped seeing the surface", seen, motionGroups)
	}
}
