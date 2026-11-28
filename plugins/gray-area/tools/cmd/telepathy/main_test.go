package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// THE BINARY ITSELF, not the package it hands over to.
//
// internal/telecli's suite drives the command tree in-process, which is where the behaviour is
// tested. What that cannot see is `main`: a tree that works perfectly while os.Exit is wired to
// the wrong value, or while Execute is never called at all, passes every one of those tests. This
// runs the built artifact and reads its exit status the way a hook and a shell do.
func TestTheBuiltBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary; skipped under -short")
	}
	// .exe on Windows, or exec refuses to run a file it built a moment earlier — "executable file
	// not found in %PATH%" for an absolute path that exists, which reads like anything but a
	// missing suffix.
	name := "telepathy"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	// A store that cannot exist, so no case below can reach the caller's real catalogue.
	noStore := filepath.Join(t.TempDir(), "absent.db")
	// A configuration and a runtime directory of the test's own: the daemon's token file lives under
	// $XDG_CONFIG_HOME, and $XDG_RUNTIME_DIR names the host's systemd user bus, so a checkout-built
	// binary never finds the host's token or asks the host's user manager.
	configDir, runtimeDir := t.TempDir(), t.TempDir()

	for _, tc := range []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string // on stdout
		wantErr  string // on stderr
	}{
		{"version", []string{"--version"}, 0, "telepathy", ""},
		{"help", []string{"--help"}, 0, "what every agent on this box", ""},
		// Bare, with no verb: cobra prints the help rather than doing something. Nothing is a
		// sensible default here — every verb answers a different question.
		{"no arguments", nil, 0, "Available Commands", ""},
		{"unknown verb", []string{"telepathise"}, 2, "", "unknown command"},
		{"missing argument", []string{"touched"}, 2, "", "accepts 1 arg"},
		{"a missing store is worded", []string{"--store", noStore, "touched", "x.go"}, 1, "", "no store at"},
		// agy-provision asks agy's install whether it knows agy-wake by these two: a binary that
		// has the verb answers its help with 0, one that lacks it refuses.
		{"agy-wake's help", []string{"agy-wake", "--help"}, 0, "Starts a turn", ""},
		{"a verb it lacks, asked for help", []string{"telepathise", "--help"}, 2, "", "unknown command"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(bin, tc.args...)
			cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir, "XDG_RUNTIME_DIR="+runtimeDir)
			var out, errOut strings.Builder
			cmd.Stdout, cmd.Stderr = &out, &errOut
			err := cmd.Run()
			code := 0
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else if err != nil {
				t.Fatalf("running the binary: %v", err)
			}
			if code != tc.wantCode {
				t.Errorf("exit %d, want %d\nstdout: %s\nstderr: %s", code, tc.wantCode, out.String(), errOut.String())
			}
			if tc.wantOut != "" && !strings.Contains(out.String(), tc.wantOut) {
				t.Errorf("stdout does not contain %q:\n%s", tc.wantOut, out.String())
			}
			if tc.wantErr != "" && !strings.Contains(errOut.String(), tc.wantErr) {
				t.Errorf("stderr does not contain %q:\n%s", tc.wantErr, errOut.String())
			}
			// A refusal belongs on stderr and a result on stdout. Getting this backwards makes
			// `telepathy sql ... > out.txt` write an error message into a data file.
			if tc.wantCode != 0 && strings.TrimSpace(out.String()) != "" {
				t.Errorf("a failure printed to stdout: %q", out.String())
			}
		})
	}
}
