//go:build !windows

package compute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// A BUNDLE FILE NAMED ./run.sh RUNS FROM THE COPY: the tool sets its executable bit, and it is no
// toolchain — its bytes are already the bundle's.
func TestABundleFileRunsAsTheCommand(t *testing.T) {
	p, err := ReadProgram(bundle(t, map[string]string{"run.sh": "#!/bin/sh\necho ran from the copy\n"}), []string{"./run.sh"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	res := run(t, p)
	if res.Output != "ran from the copy\n" || res.Drift != "" {
		t.Errorf("output %q drift %q", res.Output, res.Drift)
	}
	if res.ToolchainPath != "" || res.ToolchainSha256 != "" {
		t.Errorf("a bundle file was named a toolchain: %s %s", res.ToolchainPath, res.ToolchainSha256)
	}
}

// A BUNDLE FILE THE SYSTEM CANNOT START is refused, naming the failure: a text file with no #! line.
func TestABundleFileTheSystemCannotStartIsRefused(t *testing.T) {
	p, err := ReadProgram(bundle(t, map[string]string{"notes.txt": "9 = 3 x 3\n"}), []string{"./notes.txt"}, noRefusal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(t.TempDir(), p); err == nil || !strings.Contains(err.Error(), "cannot start ./notes.txt") {
		t.Errorf("err = %v", err)
	}
}

// A NON-REGULAR FILE has no bytes to carry.
func TestAPipeInTheBundleIsRefused(t *testing.T) {
	dir := bundle(t, map[string]string{"a": "a"})
	if err := syscall.Mkfifo(filepath.Join(dir, "p"), 0o644); err != nil {
		t.Skipf("no fifo here: %v", err)
	}
	if _, err := ReadProgram(dir, []string{"./a"}, noRefusal); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("err = %v", err)
	}
}

// THE MODES ARE THE TOOL'S, NOT THE UMASK'S. Under umask 077 and 022 alike, ./run.sh starts, and the
// second act on the bundle finds dir/ hashing to the bundle's sha.
func TestTheUmaskChangesNothing(t *testing.T) {
	for _, mask := range []int{0o077, 0o022} {
		t.Run(fmt.Sprintf("umask %03o", mask), func(t *testing.T) {
			old := syscall.Umask(mask)
			defer syscall.Umask(old)
			runDir := t.TempDir()
			p, err := ReadProgram(bundle(t, map[string]string{"run.sh": "#!/bin/sh\necho 7\n", "lib/x.txt": "x"}), []string{"./run.sh"}, noRefusal)
			if err != nil {
				t.Fatal(err)
			}
			for act := 0; act < 2; act++ {
				res, err := Run(runDir, p)
				if err != nil {
					t.Fatalf("act %d: %v", act, err)
				}
				if res.Output != "7\n" {
					t.Fatalf("act %d wrote %q", act, res.Output)
				}
				if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
					t.Fatalf("act %d: %v", act, err)
				}
			}
			files, err := Bundle(runDir, p.Sha)
			if err != nil {
				t.Fatal(err)
			}
			if got := BundleSha(p.Argv, Sums(files)); got != p.Sha {
				t.Errorf("dir/ hashes to %s, want %s", got, p.Sha)
			}
			if _, err := os.Stat(BundlePath(runDir, p.Sha)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
