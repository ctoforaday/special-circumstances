package cli

import (
	"strings"
	"testing"
)

// AN ID OFF THE ROSTER SELECTS NO SURFACE, however much of a seat's spelling it shares.
//
// `judge-x` is not the bench and `red-lens-banana` is not a lens: a surface handed to either is the
// whole verb set of a role, `halt` included, under an identity no dispatch creates. The root built
// for such an id carries no verb, and a command named under it is refused with the reason.
func TestAnIdOffTheRosterSelectsNoSurface(t *testing.T) {
	for id, verb := range map[string]string{
		"judge-x": "halt", "judgement": "halt", "red-lens-banana": "mint", "red-chair-x": "verdict",
		"blue-x": "edit", "frontierx": "edit", "blue-lane-0": "edit", "blue-lane-01": "edit",
		"operatorX": "setup", "harness": "setup",
	} {
		root := NewRootFor(id)
		for _, c := range root.Commands() {
			t.Errorf("--seat-id %s builds a surface carrying %q — an id off the roster selects none", id, c.Name())
		}
		if strings.Contains(root.Long, "the bench —") || strings.Contains(root.Long, "red lens seats") || strings.Contains(root.Long, "the red chair —") || strings.Contains(root.Long, "blue seats —") {
			t.Errorf("--seat-id %s is described as a seat's surface:\n%s", id, root.Long)
		}
		_, err := run(t, verb, "--seat-id", id)
		if err == nil {
			t.Errorf("`%s --seat-id %s` ran — an id off the roster names nobody", verb, id)
			continue
		}
		if want := id + " does not belong to any role namespace, so it selects no surface"; !strings.Contains(err.Error(), want) {
			t.Errorf("`%s --seat-id %s` is refused without the reason (%q):\n%v", verb, id, want, err)
		}
	}
	// A seat on the roster, and a lane, keep their surface.
	for id, verb := range map[string]string{"judge": "halt", "red-lens-evidence": "mint", "blue-lane-2": "edit", "operator": "setup"} {
		if cmdAt(NewRootFor(id), []string{verb}) == nil {
			t.Errorf("--seat-id %s builds a surface without %q", id, verb)
		}
	}
}
