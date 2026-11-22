// Package nativehelpertest builds nativehelper's Go fake helper (testdata/fakehelper) under a real
// helper's name, for tests that drive a client or a built binary against a helper without the
// native build, or against one that misbehaves on purpose. It is the one builder of the fake: every
// test package that needs one goes through it. It needs Go, so no test using it belongs on
// native-run's must-pass list. It imports no nativehelper, so nativehelper's own tests use it too.
//
// # The build directory
//
// Every fake a test binary builds goes into ONE directory, made on the first Build and removed by
// the function CleanupAtExit returns, which the package's TestMain runs after m.Run (Run does both
// for a package with no other TestMain). A built fake is shared by every test in the binary, so no
// test owns it and t.Cleanup cannot remove it; process exit is the only moment every caller is
// done. Build refuses to run in a binary that never armed the removal, so a package added later
// cannot leak quietly: before this, every test binary left one nativehelpertest-* directory per
// spec in TMPDIR — 1,354 of them, 3.5 GB, on 2026-11-22.
package nativehelpertest

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

// Spec is what the fake is built to be: the digest and asset sha256s its handshake reports, the
// protocol version it speaks ("" for nativehelper.ProtocolVersion), and its misbehaviour
// (testdata/fakehelper's mode list; "" for none).
type Spec struct {
	Digest   string
	Protocol string
	Mode     string
	Assets   []string
}

var (
	mu    sync.Mutex
	built = map[string]string{}
	dir   string // the one build directory, "" until the first Build
	armed bool   // CleanupAtExit was called: this binary's TestMain removes dir
)

// CleanupAtExit arms Build for this test binary and returns the removal of everything it built,
// for TestMain to run once m.Run has returned — as a testbuild.Main post-suite hook, or directly.
// The removal fails when the directory survives, and the caller fails the package on it: `go test`
// prints a passing package's stderr only under -v, so a warning on a green run is read by nobody.
func CleanupAtExit() func() error {
	mu.Lock()
	armed = true
	mu.Unlock()
	return cleanup
}

func cleanup() error {
	mu.Lock()
	defer mu.Unlock()
	if dir == "" {
		return nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("nativehelpertest: could not remove %s: %w", dir, err)
	}
	dir, built = "", map[string]string{}
	return nil
}

// Run is TestMain for a package with nothing else to do there:
//
//	func TestMain(m *testing.M) { os.Exit(nativehelpertest.Run(m)) }
func Run(m *testing.M) int {
	remove := CleanupAtExit()
	code := m.Run()
	if err := remove(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			code = 1
		}
	}
	return code
}

// Build compiles the fake for spec once per test binary and returns the built file's path.
func Build(t testing.TB, spec Spec) string {
	t.Helper()
	key := strings.Join([]string{spec.Digest, spec.Protocol, spec.Mode, strings.Join(spec.Assets, ",")}, "\x00")
	mu.Lock()
	defer mu.Unlock()
	if src, ok := built[key]; ok {
		return src
	}
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("nativehelpertest: cannot find this package's source")
	}
	if !armed {
		t.Fatal("nativehelpertest: this test binary never called CleanupAtExit, so nothing would remove " +
			"what Build makes; route TestMain through nativehelpertest.Run, or pass " +
			"nativehelpertest.CleanupAtExit() to testbuild.Main")
	}
	if dir == "" {
		d, err := os.MkdirTemp("", "nativehelpertest-")
		if err != nil {
			t.Fatal(err)
		}
		dir = d
	}
	out, err := os.MkdirTemp(dir, "fake-")
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(out, exe("fakehelper"))
	ldflags := fmt.Sprintf("-X main.digest=%s -X main.mode=%s -X main.assets=%s", spec.Digest, spec.Mode, strings.Join(spec.Assets, ","))
	if spec.Protocol != "" {
		ldflags += " -X main.protocol=" + spec.Protocol
	}
	cmd := exec.Command("go", "build", "-o", src, "-ldflags", ldflags,
		filepath.Join(filepath.Dir(here), "..", "testdata", "fakehelper"))
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building the fake helper: %v\n%s", err, b)
	}
	built[key] = src
	return src
}

// Install writes the fake built for spec into dir as name[.exe] and returns its path.
func Install(t testing.TB, dir, name string, spec Spec) string {
	t.Helper()
	b, err := os.ReadFile(Build(t, spec))
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, exe(name))
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		t.Fatal(err)
	}
	return dst
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
