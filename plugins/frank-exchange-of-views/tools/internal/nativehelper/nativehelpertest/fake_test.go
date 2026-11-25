package nativehelpertest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(Run(m)) }

// WHAT BUILD MAKES IS GONE AFTER THE REMOVAL TestMain RUNS. Before it, the build directory was
// never removed: one nativehelpertest-* per spec per test binary, 3.5 GB of them in /tmp.
func TestTheRemovalTakesEveryBuiltFake(t *testing.T) {
	a := Build(t, Spec{Mode: "a"})
	b := Build(t, Spec{Mode: "b"})
	root := dir
	for _, p := range []string{a, b} {
		if rel, err := filepath.Rel(root, p); err != nil || !filepath.IsLocal(rel) {
			t.Fatalf("built %s outside the one build directory %s", p, root)
		}
	}
	if err := CleanupAtExit()(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("%s is still there after the removal (stat: %v)", root, err)
	}
	// The cache went with it: a later Build builds again rather than returning a deleted path.
	if c := Build(t, Spec{Mode: "a"}); c == a {
		t.Fatalf("Build returned the removed path %s", c)
	} else if _, err := os.Stat(c); err != nil {
		t.Fatal(err)
	}
}
