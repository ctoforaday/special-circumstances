package fetchcache

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

// A RUN EXECUTED WITHOUT THE HELPER MUST SAY SO, LOUDLY, THROUGH THE FUSED PATH.
// Deliberately does NOT stub DefaultPageEngine: what is under test is the production
// wiring's answer when the helper is absent (NATIVE_HELPER_DIR=none). An empty reading
// here would be the plausible zero [[facts-are-fields]] names, indistinguishable from a
// document of blank pages.
func TestAScannedReadOnAStubBuildRefusesLoudly(t *testing.T) {
	t.Setenv("NATIVE_HELPER_DIR", "none")
	run, e := storeScanned(t, 1)
	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err == nil {
		t.Fatal("the engineless build produced a reading")
	}
	if !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("err = %v, want it to wrap nativehelper.ErrAbsent so callers can classify it", err)
	}
	// The sentence a seat meets in ocr_reason: what is missing.
	for _, want := range []string{"helper-absent: feov-ocr"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	// And nothing was recorded: the absence is honest, and a later engine-bearing binary
	// reads the document cleanly.
	if _, had, _ := ReadReadingRecord(run, e.Sha); had {
		t.Error("the engineless build left a reading record")
	}
}
