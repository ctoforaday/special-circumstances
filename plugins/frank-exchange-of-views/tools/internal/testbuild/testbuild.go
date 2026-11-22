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
// # The prebuilt set
//
// Across PROCESSES ON DIFFERENT RUNNERS the build cache is not there to share: the hooks
// workflow's release-fuzz parts each run one test binary on a runner of their own. Their build
// job builds every command once (`go build -o <dir>/ ./cmd/...`, the default build this package
// runs) and names the directory in PrebuiltVar; Binary then returns the command from there and
// builds nothing. A command the directory lacks is refused rather than built, so a set that
// stopped holding everything the suite runs fails the test that needed it.
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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nonet"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"io/fs"
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
	// THE SUITE IS OFF THE PUBLIC INTERNET, for the same reason HOME is sandboxed just below:
	// a test that reaches something outside this process reports someone else's problem as its
	// own. See internal/nonet for the measured case.
	nonet.OnlyLoopback()
	ClearDispatch()
	restore, err := sandboxHome()
	if err != nil {
		fmt.Fprintln(os.Stderr, "testbuild:", err)
		os.Exit(1)
	}
	restoreClient, err := stubClient()
	if err != nil {
		if rerr := restore(); rerr != nil {
			fmt.Fprintln(os.Stderr, "testbuild:", rerr)
		}
		fmt.Fprintln(os.Stderr, "testbuild:", err)
		os.Exit(1)
	}
	code := m.Run()
	// A DIRECTORY THIS PACKAGE MADE AND COULD NOT REMOVE FAILS THE PACKAGE. `go test` prints a
	// passing package's stderr only under -v, so a printed warning on a green run is read by
	// nobody: the sandbox HOME leaked a read-only module cache on every run that way, hundreds
	// of directories and 30 GB, until the disk filled. A non-zero exit is the one report a
	// green run cannot hide.
	for _, undo := range []func() error{restoreClient, restore, Cleanup} {
		if err := undo(); err != nil {
			fmt.Fprintln(os.Stderr, "testbuild:", err)
			if code == 0 {
				code = 1
			}
		}
	}
	for _, check := range after {
		if err := check(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if code == 0 {
				code = 1
			}
		}
	}
	os.Exit(code)
}

// stubClient puts a `claude` first on PATH that prints a version and nothing else: `setup` refuses a
// client it cannot read, and CI has none. A test of that refusal injects its own Exec.
func stubClient() (func() error, error) {
	dir, err := os.MkdirTemp("", "sc-testclient-")
	if err != nil {
		return nil, fmt.Errorf("no directory for the stub client: %w", err)
	}
	name, body := "claude", "#!/bin/sh\necho '99.0.0 (Claude Code)'\n"
	if runtime.GOOS == "windows" {
		name, body = "claude.bat", "@echo 99.0.0 (Claude Code)\r\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("writing the stub client: %w", err)
	}
	old, had := os.LookupEnv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+old); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("putting the stub client on PATH: %w", err)
	}
	return func() error {
		if had {
			_ = os.Setenv("PATH", old)
		} else {
			_ = os.Unsetenv("PATH")
		}
		return removeTree(dir)
	}, nil
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
// the developer's. `chair verdict` checkpoints the run's whole records/ tree to
// ~/.cache/feov/run-mirror/<sha1(runDir)[:12]> on every epoch, keyed by the run directory —
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
// ClearDispatch removes the seat-dispatch variables from the suite's environment.
//
// THE SAME PRINCIPLE AS sandboxHome AND nonet, AND IT WAS MISSED. A seat is told which run it
// belongs to through FEOV_RUN, injected by the PreToolUse hook from the project's live-run marker.
// A developer's shell is a session like any other, so while ANY run is open anywhere in the
// project — another agent's, in another worktree — every `go test` inherits that run, the
// in-process CLI resolves it as the dispatched run, and every test that passes its own --run is
// refused for disagreeing with a stranger's.
//
// Measured on 2026-09-19: 299 subtests red in one worktree from a live run in another, with a
// message that reads like a test bug and is not one. Production behaviour is untouched — a real
// seat gets the variable from the hook, per call.
//
// Exported so recordtest.Main, the module's other shared TestMain, installs the same guard rather
// than keeping a second copy of this list.
func ClearDispatch() {
	for _, v := range []string{
		seatenv.Var, // FEOV_RUN — the dispatched run
		seatenv.VarHookVersion,
		seatenv.AgentVar, // FEOV_AGENT_ID — whose seat the caller is
		seatenv.TypeVar,
	} {
		_ = os.Unsetenv(v)
	}
}

func sandboxHome() (func() error, error) {
	h, err := os.MkdirTemp("", "sc-testhome-")
	if err != nil {
		return nil, fmt.Errorf("no temp HOME: %w", err)
	}
	unpin := keepGoCaches()
	key := homeKey()
	if err := os.Setenv(key, h); err != nil {
		unpin()
		_ = removeTree(h)
		return nil, fmt.Errorf("setting %s: %w", key, err)
	}
	return func() error {
		if realHomeSet {
			_ = os.Setenv(key, realHome)
		} else {
			_ = os.Unsetenv(key)
		}
		unpin()
		return removeTree(h)
	}, nil
}

// keepGoCaches pins GOMODCACHE and GOCACHE, for every `go` this process starts, to where they
// resolve NOW — before HOME moves — and returns the undo.
//
// THE SANDBOX IS FOR THE COMMAND UNDER TEST, NOT FOR THE TOOLCHAIN; buildEnv states the same
// for the builds this package runs, and this extends it to every other `go` a test starts
// (nativehelpertest builds its fake helper with the inherited environment). Unpinned, such a
// build resolves GOPATH to $HOME/go under the sandbox, downloads the module graph into
// $HOME/go/pkg/mod — which Go writes READ-ONLY — and the sandbox could not be removed: one
// directory per test binary, up to 430 MB each, 30 GB on 2026-11-22 before the disk filled.
//
// A value the caller already set is left alone, and a machine with no `go` on PATH is left as
// it was: a test that needs one says so itself. The prosthetic-conscience plugin's hooktest
// pins the same two variables for the same reason.
func keepGoCaches() (undo func()) {
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

// removeTree removes dir and everything under it, and says so when it cannot.
//
// os.RemoveAll alone fails on a tree holding a directory without write permission — the module
// cache Go writes is exactly that — and the error used to be discarded, so the directory stayed
// and nothing said so. Write permission is restored on every directory first, so a read-only
// tree a test left behind still goes; whatever survives that is returned.
func removeTree(dir string) error {
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
func Cleanup() error {
	if dir != "" {
		return removeTree(dir)
	}
	return nil
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

// PrebuiltVar names an absolute directory holding this module's commands, already built. Set, it
// replaces every build Binary would run; see the package comment.
const PrebuiltVar = "FEOV_TESTBUILD_PREBUILT"

// prebuilt is cmd/<name> from the prebuilt directory dir.
func prebuilt(dir, name string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("%s=%s is not an absolute path; each test binary runs in its own package directory, so a relative one names a different place in each", PrebuiltVar, dir)
	}
	path := filepath.Join(dir, ExeName(name))
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s=%s holds no %s, and a prebuilt set is never completed by a build: %w", PrebuiltVar, dir, ExeName(name), err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s=%s: %s is not a regular file", PrebuiltVar, dir, ExeName(name))
	}
	return path, nil
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
// With PrebuiltVar set, the path is the prebuilt command's and nothing is built.
// It fails the test rather than returning an error: a test that cannot build the binary it
// drives has nothing left to assert, and every previous copy of this helper called t.Fatalf
// for the same reason.
func Binary(t *testing.T, name string) string {
	t.Helper()

	mu.Lock()
	once, ok := cache[name]
	if !ok {
		once = &sync.Once{}
		cache[name] = once
	}
	mu.Unlock()

	once.Do(func() {
		b := &built{}
		if pre := os.Getenv(PrebuiltVar); pre != "" {
			b.path, b.err = prebuilt(pre, name)
			mu.Lock()
			result[name] = b
			mu.Unlock()
			return
		}
		dirOne.Do(func() { dir, dirErr = os.MkdirTemp("", "sc-testbuild-") })
		root, err := moduleRoot()
		if dirErr != nil {
			b.err = fmt.Errorf("no temp dir: %w", dirErr)
		} else if err != nil {
			b.err = err
		} else {
			// The platform convention, not a fixed spelling — see the package comment.
			b.path = filepath.Join(dir, ExeName(name))
			args := []string{"build", "-o", b.path, "./cmd/" + name}
			cmd := exec.Command("go", args...)
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
		t.Fatalf("testbuild: cmd/%s: %v\n%s", name, b.err, b.out)
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
