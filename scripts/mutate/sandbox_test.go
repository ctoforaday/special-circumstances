package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// module writes a minimal buildable module and returns its directory.
func module(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module sandboxfixture\n\ngo 1.25\n")
	write("x.go", "package sandboxfixture\n\nfunc eq(a, b int) bool { return a == b }\n")
	write("sub/y.go", "package sub\n\nfunc ok() bool { return true }\n")
	return dir
}

// THE ORIGINAL IS NEVER WRITTEN TO. This is the whole reason the sandbox exists, so it is
// asserted on the original's bytes rather than on any claim the sweep makes.
func TestTheSweepNeverTouchesTheOriginal(t *testing.T) {
	src := module(t)
	before := map[string][]byte{}
	for _, rel := range []string{"go.mod", "x.go", "sub/y.go"} {
		b, err := os.ReadFile(filepath.Join(src, rel))
		if err != nil {
			t.Fatal(err)
		}
		before[rel] = b
	}

	work, cleanup, err := sandbox(src)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// Mutate every file in the copy, as a sweep does, and never restore.
	for _, rel := range []string{"x.go", "sub/y.go"} {
		if err := os.WriteFile(filepath.Join(work, rel), []byte("package broken\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for rel, want := range before {
		got, err := os.ReadFile(filepath.Join(src, rel))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s in the ORIGINAL changed: %q", rel, got)
		}
	}
}

// The copy is the tree AS IT STANDS, uncommitted work included — the reason it is a copy
// rather than a `git worktree`, which would carry committed state and silently measure
// something other than the code under test.
func TestTheSandboxCarriesTheWorkingState(t *testing.T) {
	src := module(t)
	uncommitted := "package sandboxfixture\n\nfunc eq(a, b int) bool { return a != b } // edited, not committed\n"
	if err := os.WriteFile(filepath.Join(src, "x.go"), []byte(uncommitted), 0o644); err != nil {
		t.Fatal(err)
	}
	work, cleanup, err := sandbox(src)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	got, err := os.ReadFile(filepath.Join(work, "x.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != uncommitted {
		t.Errorf("the sandbox does not carry the working state:\n got %q\nwant %q", got, uncommitted)
	}
	if _, err := os.Stat(filepath.Join(work, "sub", "y.go")); err != nil {
		t.Errorf("nested packages must be copied too: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "go.mod")); err != nil {
		t.Errorf("without go.mod the copy is not a module: %v", err)
	}
}

// cleanup removes the whole scratch tree — an abandoned sandbox is harmless, but a tool that
// leaves one per run on a developer's machine is its own kind of litter.
func TestCleanupRemovesTheScratchTree(t *testing.T) {
	work, cleanup, err := sandbox(module(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(work); err != nil {
		t.Fatal(err)
	}
	cleanup()
	if _, err := os.Stat(work); !os.IsNotExist(err) {
		t.Errorf("the scratch tree survived cleanup: %v", err)
	}
}

// A directory that is not a module is refused by NAME, rather than producing a sweep that
// finds no files and reports a clean board.
func TestADirectoryWithoutGoModIsRefused(t *testing.T) {
	_, _, err := sandbox(t.TempDir())
	if err == nil {
		t.Fatal("a directory with no go.mod must be refused, not swept")
	}
	if !strings.Contains(err.Error(), "go.mod") {
		t.Errorf("the refusal must name what is missing: %v", err)
	}
}

// restoreFile's failure ends the sweep, because a file left mutated makes every LATER mutant
// in the run measure against two defects at once. Inside the sandbox this is measurement
// hygiene rather than data safety — but a wrong measurement reported as a right one is the
// failure this tool exists to prevent elsewhere.
func TestAFailedRestoreIsReportedNotDiscarded(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes through read-only permissions")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "x.go")
	if err := os.WriteFile(p, []byte("original\n"), 0o444); err != nil {
		t.Fatal(err)
	}
	err := restoreFile(p, []byte("original\n"))
	if err == nil {
		t.Fatal("a restore that could not write must not report success")
	}
	for _, want := range []string{"x.go", "carries this one"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q is missing %q", err, want)
		}
	}
}
