package cli

import (
	"io"
	"runtime"
	"testing"
	"weak"

	"github.com/spf13/cobra"
)

// A BUILT TREE MUST BE COLLECTABLE once nothing in the caller holds it. Every invocation builds a
// fresh tree, so a package-level table keyed by *cobra.Command (or anything reaching one) keeps
// every tree a process ever built. Two such tables did, and this package's test binary held 3.4 GB
// of dead trees at exit and peaked at 5 GB resident — three concurrent suites were enough for the
// kernel to kill a session. Nothing else notices: the leak passes every assertion and only costs
// memory.
//
// Each root also renders its help, because a table filled on the READ path leaks the same way.
func TestABuiltTreeIsCollected(t *testing.T) {
	roles := []string{}
	for role := range AllRoots() {
		roles = append(roles, role)
	}
	for _, role := range roles {
		// Built inside the call, so the weak pointer is the only reference that outlives it.
		w := retainedAfterUse(t, func() *cobra.Command { return AllRoots()[role] })
		for range 3 {
			runtime.GC()
		}
		if w.Value() != nil {
			t.Errorf("the %s tree is still reachable after its last use: something package-level holds a *cobra.Command", role)
		}
	}
}

// retainedAfterUse renders the root's help and hands back a weak pointer to its deepest command,
// so a table keyed by a leaf verb is caught as well as one keyed by the root.
func retainedAfterUse(t *testing.T, build func() *cobra.Command) weak.Pointer[cobra.Command] {
	t.Helper()
	root := build()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("--help: %v", err)
	}
	leaf := root
	for len(leaf.Commands()) > 0 {
		leaf = leaf.Commands()[len(leaf.Commands())-1]
	}
	return weak.Make(leaf)
}
