//go:build windows

package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A BUNDLE FILE NAMED ./prog.exe RUNS FROM THE COPY, and is no toolchain. The program is the
// system's own command interpreter, copied into the bundle: a small .exe every Windows box holds.
func TestABundleExeRunsAsTheCommand(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe"))
	if err != nil {
		t.Skipf("no cmd.exe to copy: %v", err)
	}
	if len(body) > MaxBundle {
		t.Skipf("cmd.exe is %d bytes, over the bundle's limit", len(body))
	}
	p, err := ReadProgram(bundle(t, map[string]string{"prog.exe": string(body)}), []string{"./prog.exe", "/c", "echo ran from the copy"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	res := run(t, p)
	if !strings.Contains(res.Output, "ran from the copy") || res.Drift != "" {
		t.Errorf("output %q drift %q", res.Output, res.Drift)
	}
	if res.ToolchainPath != "" || res.ToolchainSha256 != "" {
		t.Errorf("a bundle file was named a toolchain: %s %s", res.ToolchainPath, res.ToolchainSha256)
	}
}

// ./run.sh ON WINDOWS is refused naming the start failure: the system starts no shell script directly.
func TestAShellScriptAsTheCommandIsRefusedOnWindows(t *testing.T) {
	p, err := ReadProgram(bundle(t, map[string]string{"run.sh": "#!/bin/sh\necho 7\n"}), []string{"./run.sh"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(t.TempDir(), p); err == nil || !strings.Contains(err.Error(), "cannot start ./run.sh") {
		t.Errorf("err = %v", err)
	}
}
