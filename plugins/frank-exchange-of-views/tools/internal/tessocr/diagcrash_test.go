package tessocr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

// In the helper model, a crash inside the helper is a child process exit (helper-crashed),
// which captures the helper's stderr and surfaces as an error rather than taking down the Go process.
func TestACrashInsideTheCaptureReachesStderr(t *testing.T) {
	nativehelper.Require(t, "feov-ocr")
	en, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer en.Close()

	// Terminate the helper process to simulate a crash.
	en.KillHelperForTest()

	png, err := os.ReadFile(filepath.Join("testdata", "gridcrop.png"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.PageTSV(png, PSMAuto)
	if err == nil {
		t.Fatal("expected error after helper process was killed")
	}
	if !errors.Is(err, nativehelper.ErrCrashed) {
		t.Fatalf("err = %v, want helper-crashed", err)
	}
}

// A panic in fn restores the diagnostics capture before it propagates: a caller that recovers
// it can capture the next page cleanly.
func TestARecoveredPanicLeavesStderrRestored(t *testing.T) {
	func() {
		defer func() { _ = recover() }()
		_, _ = captureDiagnostics(func() error { panic("in the window") })
	}()
	out, err := captureDiagnostics(func() error { return nil })
	if err != nil {
		t.Fatalf("second capture refused: %v", err)
	}
	if out != "" {
		t.Fatalf("second capture unexpectedly got %q", out)
	}
}
