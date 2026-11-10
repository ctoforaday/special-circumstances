// Package nativehelpertest builds nativehelper's Go fake helper under a real helper's name, for
// tests outside nativehelper that drive a client or a built binary against a helper that
// misbehaves on purpose. It needs Go, so no test using it belongs on native-run's must-pass list.
package nativehelpertest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	mu    sync.Mutex
	built = map[string]string{}
)

// Fake writes the fake helper into dir as name[.exe], built to report digest at its handshake and
// to misbehave as mode says (testdata/fakehelper's mode list), and returns its path. Each
// digest and mode builds once per test binary.
func Fake(t testing.TB, dir, name, digest, mode string) string {
	t.Helper()
	exe := name
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	key := digest + "\x00" + mode
	mu.Lock()
	src, ok := built[key]
	if !ok {
		_, here, _, ok := runtime.Caller(0)
		if !ok {
			mu.Unlock()
			t.Fatal("nativehelpertest: cannot find this package's source")
		}
		out, err := os.MkdirTemp("", "nativehelpertest-")
		if err != nil {
			mu.Unlock()
			t.Fatal(err)
		}
		src = filepath.Join(out, exe)
		cmd := exec.Command("go", "build", "-o", src,
			"-ldflags", fmt.Sprintf("-X main.digest=%s -X main.mode=%s", digest, mode),
			filepath.Join(filepath.Dir(here), "..", "testdata", "fakehelper"))
		if b, err := cmd.CombinedOutput(); err != nil {
			mu.Unlock()
			t.Fatalf("building the fake helper: %v\n%s", err, b)
		}
		built[key] = src
	}
	mu.Unlock()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, exe)
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		t.Fatal(err)
	}
	return dst
}
