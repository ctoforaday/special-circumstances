//go:build tessocr && cgo && linux

package libcprobe

import (
	"os"
	"os/exec"
	"testing"
)

const libcProbeEnv = "TESSOCR_LIBC_PROBE"

// TestLibcScansStopAtTheTerminator runs each bounded scan in a child process, because an
// overread is a SIGSEGV that would take the whole test binary down with no test named.
// Under Zig 0.16's libc every one of these faults; the pinned toolchain must not (#1598).
func TestLibcScansStopAtTheTerminator(t *testing.T) {
	if which := os.Getenv(libcProbeEnv); which != "" {
		os.Exit(libcProbe(which))
	}
	for _, which := range libcProbeCalls {
		t.Run(which, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestLibcScansStopAtTheTerminator$")
			cmd.Env = append(os.Environ(), libcProbeEnv+"="+which)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s over a string ending at an unmapped page did not return cleanly (%v): "+
					"the linked libc reads past the terminator\n%s", which, err, out)
			}
		})
	}
}
