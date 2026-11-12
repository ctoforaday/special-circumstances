package nativehelper

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
)

// fakeName is the helper name the tests' clients look up.
const fakeName = "fake-helper"

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// fake installs nativehelpertest's fake with a digest, a protocol version, a misbehaviour and the
// asset sha256 its handshake reports, and returns a fresh directory holding it as fake-helper.
func fake(t *testing.T, digest string, version uint16, mode string, assets ...string) string {
	t.Helper()
	dir := t.TempDir()
	nativehelpertest.Install(t, dir, fakeName, nativehelpertest.Spec{Digest: digest, Protocol: strconv.Itoa(int(version)), Mode: mode, Assets: assets})
	return dir
}

// startPIDs points the fake's start log at a fresh file and returns a reader of the pids in it.
func startPIDs(t *testing.T) func() []int {
	t.Helper()
	p := filepath.Join(t.TempDir(), "starts")
	t.Setenv("FAKEHELPER_START_LOG", p)
	return func() []int {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		var pids []int
		for _, f := range strings.Fields(string(b)) {
			pid, err := strconv.Atoi(f)
			if err != nil {
				t.Fatalf("the start log holds %q, not a pid", f)
			}
			pids = append(pids, pid)
		}
		return pids
	}
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
