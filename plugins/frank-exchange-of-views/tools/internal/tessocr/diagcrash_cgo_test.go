//go:build tessocr && cgo

package tessocr

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The capture points file descriptor 2 at a scratch file it then deletes, so anything that
// dies inside the window takes its trace with it. These run in a child process because the
// first one kills it, and because the second one's failure mode is a parent left with
// stderr captured.
const diagChildEnv = "TESSOCR_DIAG_CHILD"

const crashMarker = "tessocr-diag-child-crashed"

// A fatal error inside the window — the shape of a signal in tesseract, which does not
// unwind the capturing goroutine — still reaches the real stderr. A panic on another
// goroutine stands in for the signal: it kills the process while the capture is open.
func TestACrashInsideTheCaptureReachesStderr(t *testing.T) {
	if os.Getenv(diagChildEnv) == "crash" {
		_, _ = captureDiagnostics(func() error {
			go panic(crashMarker)
			select {}
		})
		return
	}
	stderr, err := runDiagChild(t, "crash", "TestACrashInsideTheCaptureReachesStderr")
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("the child was meant to die inside the capture; it returned %v", err)
	}
	if !strings.Contains(stderr, crashMarker) {
		t.Fatalf("the child died inside the capture and its trace never reached stderr; stderr was:\n%s", stderr)
	}
}

// A panic in fn restores stderr before it propagates: a caller that recovers it can capture
// the next page, and what it prints afterwards is on the real stderr.
func TestARecoveredPanicLeavesStderrRestored(t *testing.T) {
	if os.Getenv(diagChildEnv) == "recover" {
		func() {
			defer func() { _ = recover() }()
			_, _ = captureDiagnostics(func() error { panic("in the window") })
		}()
		if _, err := captureDiagnostics(func() error { return nil }); err != nil {
			os.Stderr.WriteString("second capture refused: " + err.Error() + "\n")
			os.Exit(3)
		}
		os.Stderr.WriteString(crashMarker + "\n")
		return
	}
	stderr, err := runDiagChild(t, "recover", "TestARecoveredPanicLeavesStderrRestored")
	if err != nil {
		t.Fatalf("child failed (%v); stderr:\n%s", err, stderr)
	}
	if !strings.Contains(stderr, crashMarker) {
		t.Fatalf("after a recovered panic in the window, stderr no longer reaches the parent; stderr was:\n%s", stderr)
	}
}

func runDiagChild(t *testing.T, mode, test string) (string, error) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^"+test+"$", "-test.count=1")
	cmd.Env = append(os.Environ(), diagChildEnv+"="+mode)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err
}
