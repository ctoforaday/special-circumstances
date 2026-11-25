package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// rawLockFixture writes files under a fresh root and returns it.
func rawLockFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// A RAW LOCK CALL IN SHIPPED CODE IS A FINDING, wherever it hides: a Windows-only file read on any
// platform, an aliased import, golang.org/x/sys/unix, a dot import, and shared/. A test file, a
// testdata file, gofrs/flock itself, and an unrelated selector named Flock are not.
func TestARawLockCallIsAFinding(t *testing.T) {
	const clean = "package lock\n\nimport \"github.com/gofrs/flock\"\n\nfunc L(p string) error { return flock.New(p).Lock() }\n"
	for _, c := range []struct {
		name  string
		files map[string]string
		want  []string // each a substring of one finding, in order; none means clean
	}{
		{name: "gofrs is clean", files: map[string]string{"plugins/a/tools/lock.go": clean}},
		{name: "syscall.Flock",
			files: map[string]string{"plugins/a/tools/lock.go": "package lock\n\nimport \"syscall\"\n\nfunc L(fd int) error { return syscall.Flock(fd, syscall.LOCK_EX) }\n"},
			want:  []string{"plugins/a/tools/lock.go:5 names syscall.Flock"}},
		{name: "a Windows-only file with an aliased import",
			files: map[string]string{"plugins/a/tools/lock_windows.go": "//go:build windows\n\npackage lock\n\nimport w \"golang.org/x/sys/windows\"\n\nfunc L(h w.Handle) error {\n\treturn w.LockFileEx(h, 0, 0, 1, 0, &w.Overlapped{})\n}\n\nfunc U(h w.Handle) error { return w.UnlockFileEx(h, 0, 1, 0, &w.Overlapped{}) }\n"},
			want:  []string{"lock_windows.go:8 names golang.org/x/sys/windows.LockFileEx", "lock_windows.go:11 names golang.org/x/sys/windows.UnlockFileEx"}},
		{name: "unix.Flock in shared",
			files: map[string]string{"shared/lock/lock.go": "package lock\n\nimport \"golang.org/x/sys/unix\"\n\nvar take = unix.Flock\n"},
			want:  []string{"shared/lock/lock.go:5 names golang.org/x/sys/unix.Flock"}},
		{name: "a dot import",
			files: map[string]string{"plugins/a/tools/lock.go": "package lock\n\nimport . \"syscall\"\n\nfunc L(fd int) error { return Flock(fd, LOCK_EX) }\n"},
			want:  []string{"lock.go:3 dot-imports syscall"}},
		{name: "tests, testdata and another package's Flock are not findings",
			files: map[string]string{
				"plugins/a/tools/lock.go":            clean,
				"plugins/a/tools/lock_test.go":       "package lock\n\nimport \"syscall\"\n\nvar take = syscall.Flock\n",
				"plugins/a/tools/testdata/raw.go":    "package raw\n\nimport \"syscall\"\n\nvar take = syscall.Flock\n",
				"plugins/a/tools/other/other.go":     "package other\n\ntype t struct{}\n\nfunc (t) Flock() {}\n\nvar syscall t\n\nfunc F() { syscall.Flock() }\n",
				"plugins/a/tools/flockpkg/flockp.go": "package flockpkg\n\nimport \"github.com/gofrs/flock\"\n\nvar Flock = flock.New\n",
			}},
		{name: "an empty tree is not a pass",
			files: map[string]string{"plugins/a/README.md": "no Go here\n"},
			want:  []string{"no Go file was read"}},
		{name: "a file that does not parse is a finding",
			files: map[string]string{"plugins/a/tools/broken.go": "package broken\n\nfunc {\n"},
			want:  []string{"plugins/a/tools/broken.go: does not parse", "no Go file was read"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := rawLockProblems(rawLockFixture(t, c.files))
			if len(got) != len(c.want) {
				t.Fatalf("got %d findings, want %d:\n%s", len(got), len(c.want), strings.Join(got, "\n"))
			}
			for i, w := range c.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("finding %d = %q, want it to contain %q", i, got[i], w)
				}
			}
		})
	}
}

// THE REPOSITORY HAS NO RAW LOCK CALL: every file lock in plugins/ and shared/ goes through
// gofrs/flock.
func TestTheTreeTakesNoRawLock(t *testing.T) {
	root, err := gitx.Root()
	if err != nil {
		t.Fatal(err)
	}
	if got := rawLockProblems(root); len(got) != 0 {
		t.Errorf("raw lock calls in the tree:\n%s", strings.Join(got, "\n"))
	}
}
