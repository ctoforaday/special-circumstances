// Package testbuild builds this module's commands ONCE per test binary.
//
// Eleven independent copies of "shell out to go build" had grown across this repository, and
// they disagreed about the two things such a helper has to get right.
//
// # The extension
//
// Three conventions were in use: `buildid.ExeName`, an unconditional ".exe", and a
// `runtime.GOOS == "windows"` conditional. Both failure directions are real and the tree has
// paid for each.
//
// Omitting it on WINDOWS produces a binary that cannot start — Go does not append .exe for
// an explicit -o filename — and the error was thrown away, which is how nine probe boards
// failed with a blank reason.
//
// Appending it on LINUX breaks the other half: `run-setup`'s preflight derives the binary's
// name from the platform, so a `feov-record.exe` sitting on Linux is a binary the production
// path cannot find. A first draft of this package spelled it ".exe" always, on the reasoning
// that one spelling cannot drift — and TestSetupCLIArgParsing refused it with "feov-record
// not runnable (no such file)". The extension is NOT inert when something other than the
// test resolves the name.
//
// So: the platform convention, spelled once — the same rule buildid.ExeName states for the
// sibling module.
//
// # The rebuild
//
// None of the copies cached, so a package calling its helper from four tests linked the
// binary four times. Linking is the expensive half of `go build` — compilation is already
// cached by the toolchain — and it is paid per call on the platform where process cost
// dominates the bill. Within one test binary the result is now built once and shared.
//
// Across PACKAGES it cannot be shared: each is its own process, and a cache that outlived
// the process would have to key on source state, which is what the build cache already does
// properly. So this removes the repeats it can see and claims nothing about the rest.
//
// # Finding the module
//
// By walking up from the working directory to the go.mod that owns it — not by locating the
// repository root. A test binary that depends on repo layout breaks when files move, which
// is a lesson this tree has already paid for: one helper counted ".." segments to reach the
// tools directory and stopped working when its package was relocated.
package testbuild

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// Main runs the test binary inside a sandbox: its own HOME, and the shared build directory
// removed at exit.
//
// Two leaks, one hook, because TestMain is the only place either can be closed — the HOME has
// to be in place before the first test spawns anything, and the build directory cannot be
// removed until the last one has stopped using it. sandboxHome documents the cache half.
//
// A consumer that calls Binary MUST route its tests through this. The directory is created
// once per test binary and shared by every call in it, so no single test owns it and
// t.Cleanup cannot remove it — the first test to finish would delete the path the next one
// is about to execute. The process exit is the only moment at which every caller is done,
// and TestMain is the only hook Go offers there.
//
// MEASURED, which is why this exists: one `go test ./...` of this module left five
// sc-testbuild directories holding one linked feov-record apiece — 186 MB, on a fully green
// run. TMPDIR is a 2 GB tmpfs both in CI (.github/workflows/hooks.yml sets
// TMPDIR=/dev/shm/tmp) and in the record-run plan's validation loop, and that plan already
// records what a full one looks like: `[build failed]` for every package in the module while
// `go build ./...` exits 0.
//
// TestEveryPackageThatBuildsCleansUp fails a package that forgets to call this.
// after are post-suite checks, run once every test's cleanup has fired and before the process
// exits. A non-nil error is printed and fails the package.
//
// It is variadic because TestMain is a package's ONE hook and packages need more than one thing
// from it: this package sandboxes a build directory, and a package that also opens records owes
// recordtest.CheckOrphanedHandles. Without somewhere to put the second, a package had to choose,
// and the one it dropped would be the one whose absence is silent. The hook is a func rather than
// an import so testbuild stays ignorant of the record layer.
func Main(m *testing.M, after ...func() error) {
	restore, err := sandboxHome()
	if err != nil {
		fmt.Fprintln(os.Stderr, "testbuild:", err)
		os.Exit(1)
	}
	code := m.Run()
	restore()
	for _, check := range after {
		if err := check(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if code == 0 {
				code = 1
			}
		}
	}
	Cleanup()
	os.Exit(code)
}

// homeKey is the environment variable os.UserHomeDir reads on this platform.
func homeKey() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}
	return "HOME"
}

// realHome is the developer's own home, captured at package initialisation — BEFORE Main can
// redirect it, and regardless of whether Main is ever called.
var realHome, realHomeSet = os.LookupEnv(homeKey())

// sandboxHome points HOME at a temporary directory for the life of this test binary and
// returns the undo.
//
// The command under test writes into the USER CACHE, and a test has no business writing into
// the developer's. `merge verdict` checkpoints the run's whole records/ tree to
// ~/.cache/feov/run-mirror/<sha1(runDir)[:12]> on every round, keyed by the run directory —
// and a test's run directory is a t.TempDir(), so every run mints a mirror under a key nothing
// will ever look up again. Measured before this: 10,008 orphaned mirrors, 182,665 files,
// 15.0 GB, of which one `go test ./...` of this module contributed 69 directories and 135 MB.
// The 30-day purge in internal/setup is reachable only from run-setup, so on a developer box
// it effectively never runs.
//
// Under a temporary HOME the same writes land inside TMPDIR and go out with it. Production
// behaviour is untouched, and the goldens still record a real run-mirror path for the
// harness's normalizer to match — which is why this is done here rather than by teaching
// verdict.go to skip the mirror for temp run directories.
func sandboxHome() (func(), error) {
	h, err := os.MkdirTemp("", "sc-testhome-")
	if err != nil {
		return nil, fmt.Errorf("no temp HOME: %w", err)
	}
	key := homeKey()
	if err := os.Setenv(key, h); err != nil {
		_ = os.RemoveAll(h)
		return nil, fmt.Errorf("setting %s: %w", key, err)
	}
	return func() {
		if realHomeSet {
			_ = os.Setenv(key, realHome)
		} else {
			_ = os.Unsetenv(key)
		}
		_ = os.RemoveAll(h)
	}, nil
}

// buildEnv is the environment for `go build`: this process's, with HOME put back to the real
// one.
//
// THE TOOLCHAIN RESOLVES ITS CACHES THROUGH HOME. GOPATH defaults to $HOME/go and the build
// cache to the user cache directory, so a build inheriting the sandboxed HOME would miss the
// module cache and the build cache entirely — re-resolving the module graph and relinking from
// cold, once per test binary, which is the cost this package exists to remove. The sandbox is
// for the COMMAND UNDER TEST; the toolchain that builds it belongs in the real cache.
func buildEnv() []string {
	env := os.Environ()
	key := homeKey() + "="
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, key) {
			out = append(out, kv)
		}
	}
	if realHomeSet {
		out = append(out, key+realHome)
	}
	return out
}

// Cleanup removes the shared build directory, if one was ever created.
//
// Called from Main after m.Run has returned, so every test goroutine that could read dir has
// finished — which is what makes the unsynchronized read here safe, and why this is not
// exported for use at any other moment.
func Cleanup() {
	if dir != "" {
		_ = os.RemoveAll(dir)
	}
}

// ExeName is the file name a built command must have on THIS platform.
//
// Required on Windows (an extensionless file will not start) and refused on Linux by any
// caller that resolves the name itself rather than being handed the path.
func ExeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// built records one command's build, so a failure is reported to every caller rather than
// only to the first one that raced into the Once.
type built struct {
	path string
	err  error
	out  []byte
}

var (
	mu     sync.Mutex
	cache  = map[string]*sync.Once{}
	result = map[string]*built{}
	dir    string
	dirErr error
	dirOne sync.Once
)

// Binary builds cmd/<name> for this module and returns the path to the executable.
//
// The build happens on the first call in this test binary; later calls return the same path.
// It fails the test rather than returning an error: a test that cannot build the binary it
// drives has nothing left to assert, and every previous copy of this helper called t.Fatalf
// for the same reason.
func Binary(t *testing.T, name string) string {
	t.Helper()

	dirOne.Do(func() { dir, dirErr = os.MkdirTemp("", "sc-testbuild-") })
	if dirErr != nil {
		t.Fatalf("testbuild: no temp dir: %v", dirErr)
	}

	mu.Lock()
	once, ok := cache[name]
	if !ok {
		once = &sync.Once{}
		cache[name] = once
	}
	mu.Unlock()

	once.Do(func() {
		b := &built{}
		root, err := moduleRoot()
		if err != nil {
			b.err = err
		} else {
			// The platform convention, not a fixed spelling — see the package comment.
			b.path = filepath.Join(dir, ExeName(name))
			cmd := exec.Command("go", "build", "-o", b.path, "./cmd/"+name)
			cmd.Dir = root
			cmd.Env = buildEnv()
			b.out, b.err = cmd.CombinedOutput()
		}
		mu.Lock()
		result[name] = b
		mu.Unlock()
	})

	mu.Lock()
	b := result[name]
	mu.Unlock()
	if b.err != nil {
		t.Fatalf("testbuild: building cmd/%s: %v\n%s", name, b.err, b.out)
	}
	return b.path
}

// moduleRoot walks up from the working directory to the directory holding go.mod.
func moduleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := wd; ; {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", &os.PathError{Op: "find go.mod above", Path: wd, Err: os.ErrNotExist}
		}
		d = parent
	}
}
