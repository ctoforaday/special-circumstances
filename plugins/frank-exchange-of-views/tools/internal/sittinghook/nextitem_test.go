package sittinghook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
)

// THE HELD DELIVERY CROSSES THE PROCESS BOUNDARY AS AN EXIT STATUS: a writer exiting
// hookgate.ExitDeliveryHeld is hookgate.ErrDeliveryHeld here, carrying its diagnostics; any other
// non-zero exit is an error that is not it.
func TestNextItemReadsTheHeldStatusAsHeld(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stub writer is a shell script")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	stub := filepath.Join(filepath.Dir(self), writerFileName())
	t.Cleanup(func() { _ = os.Remove(stub) })
	writer := func(status int) {
		t.Helper()
		script := fmt.Sprintf("#!/bin/sh\necho 'feov-sitting-write: held (x.lock)' >&2\nexit %d\n", status)
		if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
			t.Skipf("cannot place a writer beside the test binary: %v", err)
		}
	}

	writer(hookgate.ExitDeliveryHeld)
	text, err := NextItem(t.TempDir(), "a1", "frank-exchange-of-views:red-lens-logic", "claude")
	if !errors.Is(err, hookgate.ErrDeliveryHeld) || text != "" {
		t.Fatalf("a writer exiting %d gave %q, %v — want nothing and hookgate.ErrDeliveryHeld", hookgate.ExitDeliveryHeld, text, err)
	}
	if !strings.Contains(err.Error(), "x.lock") {
		t.Errorf("the held error dropped the writer's diagnostics: %v", err)
	}

	writer(1)
	if _, err := NextItem(t.TempDir(), "a1", "frank-exchange-of-views:red-lens-logic", "claude"); err == nil || errors.Is(err, hookgate.ErrDeliveryHeld) {
		t.Errorf("a writer exiting 1 gave %v — want an error that is not ErrDeliveryHeld", err)
	}
}
