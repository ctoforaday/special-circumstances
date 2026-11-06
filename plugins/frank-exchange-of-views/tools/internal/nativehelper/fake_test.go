package nativehelper

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

// fakeName is the helper name the tests' clients look up.
const fakeName = "fake-helper"

var (
	fakeRoot  string
	fakeMu    sync.Mutex
	fakeBuilt = map[string]string{}
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "nativehelper-fakes-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fakeRoot = dir
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// fake builds testdata/fakehelper with a digest, a protocol version and a misbehaviour, once per
// combination, and returns the directory holding it as fake-helper.
func fake(t *testing.T, digest string, version uint16, mode string) string {
	t.Helper()
	key := fmt.Sprintf("%s-%d-%s", digest, version, mode)
	fakeMu.Lock()
	defer fakeMu.Unlock()
	if dir, ok := fakeBuilt[key]; ok {
		return dir
	}
	dir := filepath.Join(fakeRoot, key)
	ldflags := fmt.Sprintf("-X main.digest=%s -X main.protocol=%d -X main.mode=%s", digest, version, mode)
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, exe(fakeName)), "-ldflags", ldflags, "./testdata/fakehelper")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building the fake helper: %v\n%s", err, out)
	}
	fakeBuilt[key] = dir
	return dir
}

// startLog points the fake's start log at a fresh file and returns a counter over it.
func startLog(t *testing.T) func() int {
	t.Helper()
	p := filepath.Join(t.TempDir(), "starts")
	t.Setenv("FAKEHELPER_START_LOG", p)
	return func() int {
		b, err := os.ReadFile(p)
		if err != nil {
			return 0
		}
		return len(strings.Fields(string(b)))
	}
}
