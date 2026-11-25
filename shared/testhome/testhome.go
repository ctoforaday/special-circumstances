// Package testhome points a test binary's home directory at a temporary one and removes it at exit,
// loudly. It is test-only and is the one copy of that sandbox for every module in this repository.
//
// # Why one copy
//
// Twelve hand-written copies had grown, and the directory leaked out of most of them, three ways:
//
//   - `defer os.RemoveAll(dir)` followed by `os.Exit(m.Run())`: os.Exit runs no deferred call, so
//     the directory was never removed at all.
//   - A `go` the tests started resolved GOPATH, and with it the module cache, under the sandboxed
//     HOME. Go writes the module cache read-only, so RemoveAll failed on it.
//   - The error from RemoveAll was discarded, so nothing said either had happened.
//
// On 2026-11-22 that was tens of gigabytes of sc-testhome-*, gray-area-capture-test-* and
// pc-hookfailures-test-* directories in TMPDIR, and the disk filled twice.
//
// # What this does instead
//
// New pins GOMODCACHE and GOCACHE to where they resolve before HOME moves, so no toolchain state
// lands in the sandbox: the sandbox is for the command under test, not for the toolchain. It also
// switches Go telemetry off in the sandbox's config directory, so no `go` leaves a detached child
// writing there after the removal. Its undo
// removes the tree with RemoveTree, which restores write permission when a plain RemoveAll fails,
// and returns what survives. Run fails the package on that error, because `go test` prints a
// passing package's stderr only under -v: a warning on a green run is read by nobody.
package testhome

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Keys are the variables a home-relative path resolves from: os.UserHomeDir reads HOME on Unix and
// USERPROFILE on Windows, and XDG_STATE_HOME outranks both for ~/.local/state.
var Keys = []string{"HOME", "USERPROFILE", "XDG_STATE_HOME"}

// New makes a temporary directory named with prefix, points each of keys at it (Keys when none are
// given) and returns it with the undo: the keys and the Go caches put back as they were, and the
// directory removed. The undo's error is the directory surviving.
func New(prefix string, keys ...string) (dir string, undo func() error, err error) {
	if len(keys) == 0 {
		keys = Keys
	}
	dir, err = os.MkdirTemp("", prefix)
	if err != nil {
		return "", nil, fmt.Errorf("testhome: no temporary home: %w", err)
	}
	unpin := pinGoCaches()
	type prior struct {
		key, value string
		had        bool
	}
	var set []prior
	restore := func() {
		for _, p := range set {
			if p.had {
				_ = os.Setenv(p.key, p.value)
			} else {
				_ = os.Unsetenv(p.key)
			}
		}
		unpin()
	}
	for _, k := range keys {
		v, had := os.LookupEnv(k)
		if err := os.Setenv(k, dir); err != nil {
			restore()
			_ = RemoveTree(dir)
			return "", nil, fmt.Errorf("testhome: setting %s: %w", k, err)
		}
		set = append(set, prior{k, v, had})
	}
	if err := telemetryOff(dir); err != nil {
		restore()
		_ = RemoveTree(dir)
		return "", nil, err
	}
	return dir, func() error {
		restore()
		return RemoveTree(dir)
	}, nil
}

// Run is a whole TestMain body: m (a *testing.M) runs under a home from New, then each of after
// runs, and the returned code is non-zero when a test failed, the home could not be removed, or a
// check failed.
//
//	func TestMain(m *testing.M) { os.Exit(testhome.Run(m, "my-package-test-")) }
//
// m is taken as the one method Run calls, so this package does not import testing and a command
// may use RemoveTree.
func Run(m interface{ Run() int }, prefix string, after ...func() error) int {
	_, undo, err := New(prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	code := m.Run()
	for _, f := range append([]func() error{undo}, after...) {
		if err := f(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if code == 0 {
				code = 1
			}
		}
	}
	return code
}

// pinGoCaches sets GOMODCACHE and GOCACHE, for every `go` this process starts, to where they
// resolve now, and returns the undo. A value the caller already set is left alone, and a machine
// with no `go` on PATH is left as it was: a test that needs one says so itself.
func pinGoCaches() (undo func()) {
	out, err := exec.Command("go", "env", "GOMODCACHE", "GOCACHE").Output()
	if err != nil {
		return func() {}
	}
	paths := strings.Split(strings.TrimSpace(string(out)), "\n")
	var set []string
	for i, k := range []string{"GOMODCACHE", "GOCACHE"} {
		if i >= len(paths) || strings.TrimSpace(paths[i]) == "" || os.Getenv(k) != "" {
			continue
		}
		if os.Setenv(k, strings.TrimSpace(paths[i])) == nil {
			set = append(set, k)
		}
	}
	return func() {
		for _, k := range set {
			_ = os.Unsetenv(k)
		}
	}
}

// telemetryOff switches Go telemetry off in the user config directory, when the redirect moved it
// into dir.
//
// A `go` command whose config directory holds no telemetry mode starts a DETACHED child that
// writes go/telemetry/ there after the parent exits. Under a temporary home that child outlives
// the removal and re-creates the directory: measured 2026-11-22, ten homes each removed right
// after one `go env`, all ten back five seconds later. A mode of "off" starts no child and writes
// nothing. A config directory the redirect did not move (XDG_CONFIG_HOME set) is the caller's.
func telemetryOff(dir string) error {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil
	}
	if rel, err := filepath.Rel(dir, cfg); err != nil || !filepath.IsLocal(rel) {
		return nil
	}
	tel := filepath.Join(cfg, "go", "telemetry")
	if err := os.MkdirAll(tel, 0o755); err != nil {
		return fmt.Errorf("testhome: switching Go telemetry off: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tel, "mode"), []byte("off"), 0o644); err != nil {
		return fmt.Errorf("testhome: switching Go telemetry off: %w", err)
	}
	return nil
}

// RemoveTree removes dir and everything under it, and says so when it cannot. A plain RemoveAll
// fails on a directory without write permission, which is how Go writes its module cache, so on
// failure write permission is restored on every directory and the removal tried again; whatever
// survives that is the error.
func RemoveTree(dir string) error {
	if err := os.RemoveAll(dir); err == nil {
		return nil
	}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			if info, ierr := d.Info(); ierr == nil {
				_ = os.Chmod(p, info.Mode().Perm()|0o700)
			}
		}
		return nil
	})
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("could not remove %s: %w", dir, err)
	}
	return nil
}
