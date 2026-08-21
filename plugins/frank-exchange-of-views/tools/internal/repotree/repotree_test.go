package repotree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE EMPTY CASE IS THE ONLY CASE WORTH TESTING HERE, and it is what the callers could not test
// for themselves.
//
// Every gate this package serves asserts a NEGATIVE over a discovered set: no constitution names a
// verb, no literal names a retired surface, no golden carries a home directory. Those assertions
// are true of the empty set. So the contract is not "Glob finds files" — a caller can see that —
// it is "Glob REFUSES to hand back nothing", which is the branch no caller exercised because in a
// healthy tree it never runs.

func TestRootFindsTheRepositoryFromInsideIt(t *testing.T) {
	root, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"plugins", "scripts", "CLAUDE.md"} {
		if _, err := os.Stat(filepath.Join(root, want)); err != nil {
			t.Errorf("%s is not the repository root: %s is missing (%v)", root, want, err)
		}
	}
}

// A directory with no repository above it must be an ERROR, not a zero value that later joins into
// a path relative to nowhere.
func TestRootRefusesOutsideTheRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := Root(); err == nil {
		t.Fatal("Root() succeeded from a temp directory — a root found where there is none repoints " +
			"every caller at a tree that does not exist, and the negative assertions over it all pass")
	}
}

// Root must not depend on where the caller's own file lives. This is the second copy of the fact
// the package removes: callers used to count `..` from their package directory.
func TestRootIsTheSameFromAnyDepth(t *testing.T) {
	here, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(here, "plugins", "frank-exchange-of-views", "tools", "internal"))
	deeper, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	if deeper != here {
		t.Errorf("Root() from internal/ gave %q, from repotree/ gave %q", deeper, here)
	}
}

func TestGlobRefusesAPatternThatMatchesNothing(t *testing.T) {
	_, err := Glob("plugins", "frank-exchange-of-views", "agents-moved-away", "*.md")
	if err == nil {
		t.Fatal("Glob returned no error for a directory that does not exist — an empty slice is what " +
			"the caller then asserts a negative over, and the assertion holds")
	}
	if !strings.Contains(err.Error(), "agents-moved-away") {
		t.Errorf("the refusal does not name the pattern that missed: %v", err)
	}
}

// An EXISTING but EMPTY directory is the harder half: the path resolves, the glob is well-formed,
// and it still discovers nothing.
func TestGlobRefusesAnEmptyDirectory(t *testing.T) {
	root, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(root, "plugins", "frank-exchange-of-views", "tools", "internal", "repotree", "emptyfixture")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(empty) })
	if _, err := Glob("plugins", "frank-exchange-of-views", "tools", "internal", "repotree", "emptyfixture", "*.md"); err == nil {
		t.Fatal("Glob accepted a directory that exists and holds nothing — the path being valid is not " +
			"the question the callers are asking")
	}
}

func TestConstitutionsAreFoundAndAreMarkdown(t *testing.T) {
	paths, err := Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	// Four seat classes ship constitutions; the gates that read them assume every one is present.
	if len(paths) < 4 {
		t.Errorf("found %d constitution(s), expected at least 4: %v", len(paths), paths)
	}
	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			t.Errorf("%s is not markdown", p)
		}
		if !filepath.IsAbs(p) {
			t.Errorf("%s is relative — a caller that changes directory would lose it", p)
		}
	}
}

func TestGoSourcesRefusesADirectoryWithNoGo(t *testing.T) {
	if _, err := GoSources("plugins", "frank-exchange-of-views", "agents"); err == nil {
		t.Fatal("GoSources accepted a directory of markdown — a walk that reads no source reports " +
			"every file clean")
	}
}

func TestToolSourcesFindsSourceAndSkipsTests(t *testing.T) {
	paths, err := ToolSources()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 20 {
		t.Errorf("found %d source file(s) under internal/ — that is too few to be the whole tree", len(paths))
	}
	var thisFile string
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			t.Errorf("%s is a test file; the sweeps that use this list exempt tests on purpose", p)
		}
		if strings.HasSuffix(filepath.ToSlash(p), "internal/repotree/repotree.go") {
			thisFile = p
		}
	}
	if thisFile == "" {
		t.Error("the walk did not reach internal/repotree/repotree.go — it is not covering the tree it claims to")
	}
}

func TestDebateJSResolvesToTheShippedScript(t *testing.T) {
	p, err := DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("the orchestrator script is not where this package says it is: %v", err)
	}
	if !strings.Contains(string(b), "feov-record") {
		t.Errorf("%s does not name the tool — this is not the orchestrator", p)
	}
}
