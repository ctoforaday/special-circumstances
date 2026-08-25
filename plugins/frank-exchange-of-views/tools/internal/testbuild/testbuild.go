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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

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
