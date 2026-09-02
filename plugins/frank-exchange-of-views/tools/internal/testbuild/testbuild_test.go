package testbuild

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// This package builds too, so it owes the same cleanup its consumers do — and it was one of
// the five directories the leak measurement found.
func TestMain(m *testing.M) { Main(m) }

// The binary must actually RUN, which is the assertion the .exe divergence needed and did
// not have. A builder that produced an unstartable file passed every test that only checked
// `go build` exited 0 — and on Windows that is exactly what an extensionless -o produces.
func TestTheBuiltBinaryRuns(t *testing.T) {
	bin := Binary(t, "feov-record")
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("the built binary did not run: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "feov-record") {
		t.Errorf("--help does not look like feov-record's: %s", out)
	}
	// The PLATFORM's name, in both directions. On Windows the extension is required or the
	// binary will not start; on Linux it must be absent or run-setup's preflight — which
	// derives the name itself — cannot find the binary it was pointed at.
	if got, want := filepath.Base(bin), ExeName("feov-record"); got != want {
		t.Errorf("built %q, want %q for %s", got, want, runtime.GOOS)
	}
}

// BUILT ONCE, not once per caller. The copies this package replaces linked the binary on
// every call, and two of them were called four times in one package.
func TestTheSecondCallDoesNotRebuild(t *testing.T) {
	first := Binary(t, "feov-record")
	info, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}
	second := Binary(t, "feov-record")
	if second != first {
		t.Errorf("second call returned %q, want the cached %q", second, first)
	}
	again, err := os.Stat(second)
	if err != nil {
		t.Fatal(err)
	}
	if !again.ModTime().Equal(info.ModTime()) {
		t.Errorf("the binary was rebuilt: %v -> %v", info.ModTime(), again.ModTime())
	}
}

// Concurrent callers get one build and the same path — `go test` runs test functions in one
// process and a parallel package would otherwise race the Once.
func TestConcurrentCallersShareOneBuild(t *testing.T) {
	const n = 8
	paths := make([]string, n)
	var wg sync.WaitGroup
	for i := range paths {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			paths[i] = Binary(t, "feov-record")
		}(i)
	}
	wg.Wait()
	for i, p := range paths {
		if p != paths[0] {
			t.Fatalf("caller %d got %q, caller 0 got %q", i, p, paths[0])
		}
	}
}

// The module is found by walking up to go.mod, NOT by counting ".." to a repo root. The
// helper that counted segments broke when its package moved; this asserts the property that
// made that impossible rather than the path it happens to produce.
func TestModuleRootIsFoundByGoMod(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Errorf("moduleRoot returned %q, which holds no go.mod: %v", root, err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "feov-record")); err != nil {
		t.Errorf("moduleRoot %q does not contain cmd/feov-record: %v", root, err)
	}
}

// THE SANDBOX IS LIVE, which is the half a static guard cannot see. TestEveryPackageThatBuilds
// CleansUp proves every consumer routes through Main; this proves what routing through it buys.
func TestTheTestBinaryHasItsOwnHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if realHomeSet && home == realHome {
		t.Fatalf("os.UserHomeDir() is still the developer's own %q — the sandbox did not take, "+
			"and every `merge verdict` this binary drives writes a mirror into it", home)
	}
	// Under TMPDIR, so it goes out with the temp filesystem rather than persisting.
	if rel, err := filepath.Rel(os.TempDir(), home); err != nil || strings.HasPrefix(rel, "..") {
		t.Errorf("the sandbox home %q is not under TMPDIR %q", home, os.TempDir())
	}
	// And it must be writable: the point is that the writes still HAPPEN, somewhere harmless.
	probe := filepath.Join(home, ".cache", "feov", "probe")
	if err := os.MkdirAll(probe, 0o755); err != nil {
		t.Errorf("the sandbox home is not writable: %v", err)
	}
}

// The TOOLCHAIN keeps the real home even while the command under test does not. A build that
// inherited the sandbox would re-resolve the module graph and relink from cold, once per test
// binary — which is the cost this package was written to remove.
func TestTheBuildKeepsTheRealHome(t *testing.T) {
	if !realHomeSet {
		t.Skip("no HOME in this environment, so there is nothing for the build to keep")
	}
	want := homeKey() + "=" + realHome
	var got []string
	for _, kv := range buildEnv() {
		if strings.HasPrefix(kv, homeKey()+"=") {
			got = append(got, kv)
		}
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("buildEnv carries %v for %s, want exactly [%q]", got, homeKey(), want)
	}
}
