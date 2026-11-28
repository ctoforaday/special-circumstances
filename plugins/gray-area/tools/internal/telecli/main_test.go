package telecli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testConfigDir is the configuration directory every test in this package sees as
// $XDG_CONFIG_HOME (and %APPDATA%), and testRuntimeDir its $XDG_RUNTIME_DIR.
var testConfigDir, testRuntimeDir string

// TestMain points $XDG_CONFIG_HOME, %APPDATA% and $XDG_RUNTIME_DIR at directories of the suite's
// own before any test runs. The daemon's token file lives under the configuration directory, and
// with none there no test that builds its Env from Defaults() can wake the host's daemon; the
// runtime directory names the host's systemd user bus, which CurrentDaemon asks.
func TestMain(m *testing.M) {
	cfg, err := os.MkdirTemp("", "telecli-config-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rt, err := os.MkdirTemp("", "telecli-runtime-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testConfigDir, testRuntimeDir = cfg, rt
	os.Setenv("XDG_CONFIG_HOME", cfg)
	os.Setenv("APPDATA", cfg)
	os.Setenv("XDG_RUNTIME_DIR", rt)
	code := m.Run()
	os.RemoveAll(cfg)
	os.RemoveAll(rt)
	os.Exit(code)
}

func TestDefaultsNeverReachTheHostConfig(t *testing.T) {
	d := Defaults()
	for name, p := range map[string]string{"TokenFile": d.TokenFile, "DropInDir": d.DropInDir} {
		if p == "" || !strings.HasPrefix(p, testConfigDir+string(filepath.Separator)) {
			t.Errorf("Defaults().%s = %q, want a path under the suite's own %q", name, p, testConfigDir)
		}
	}
}
