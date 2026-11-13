// Package nativehelpertest builds nativehelper's Go fake helper (testdata/fakehelper) under a real
// helper's name, for tests that drive a client or a built binary against a helper without the
// native build, or against one that misbehaves on purpose. It is the one builder of the fake: every
// test package that needs one goes through it. It needs Go, so no test using it belongs on
// native-run's must-pass list. It imports no nativehelper, so nativehelper's own tests use it too.
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
)

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
	out, err := os.MkdirTemp("", "nativehelpertest-")
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
