package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// THE DUTY NAMES A THING THE HELP MUST REACH.
//
// A duty says WHAT is owed and nothing about how — the help page is the only page that instructs.
// That holds only if every verb a duty points at is findable from the seat's own help. Motions
// were the exception: the group sits at the ROOT (any seat files, one role rules), so
// `merge --help` listed no motion verb and no pointer to one, while the worklist says "motion M1
// was filed and never ruled — PASS is refused while it stands".

func TestEveryRoleHelpPointsAtTheMotionGroup(t *testing.T) {
	root := newRoot()
	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		c, _, err := root.Find([]string{role})
		if err != nil {
			t.Fatalf("no %s group: %v", role, err)
		}
		if !strings.Contains(c.Long, "motion --help") {
			t.Errorf("%s --help never names the motion group, and no motion verb is mounted under it — "+
				"a seat told a motion is unruled has nowhere to go", role)
		}
	}
}

// THE POINTER MUST POINT AT SOMETHING. A footer naming a group that has been renamed or removed
// is worse than no footer: it sends a seat somewhere and the trip fails.
func TestTheMotionGroupTheFooterNamesExists(t *testing.T) {
	root := newRoot()
	c, _, err := root.Find([]string{"motion"})
	if err != nil || c == nil || c.Name() != "motion" {
		t.Fatalf("every role's help sends a seat to `motion --help`, and the root has no motion group: %v", err)
	}
	if !c.HasSubCommands() {
		t.Error("the motion group has no subjects, so `motion --help` teaches a seat nothing")
	}
	// And each subject must itself list its verbs, or the second hop dead-ends too.
	for _, sub := range c.Commands() {
		if sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		if !sub.HasSubCommands() {
			t.Errorf("motion %s offers no verbs — the pointer resolves to a page with nothing on it", sub.Name())
		}
	}
}

// The role groups are built by one constructor, so this catches a role added later that misses
// the footer — the drift a hand-applied footer invites.
func TestNoRoleGroupIsMissingAFooter(t *testing.T) {
	root := newRoot()
	for _, c := range root.Commands() {
		if !isRoleGroup(c) {
			continue
		}
		for _, want := range []string{"motion --help", "friction"} {
			if !strings.Contains(c.Long, want) {
				t.Errorf("role group %q is missing %q from its help", c.Name(), want)
			}
		}
	}
}

// isRoleGroup reports whether a top-level command is one of the four seat roles, keyed on the
// verb set rather than a name list so a new role is covered without editing this test.
func isRoleGroup(c *cobra.Command) bool {
	if !c.HasSubCommands() {
		return false
	}
	for _, sub := range c.Commands() {
		if sub.Name() == "register" {
			return true
		}
	}
	return false
}
