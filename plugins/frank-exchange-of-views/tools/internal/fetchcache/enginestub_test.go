//go:build !tessocr || !cgo

package fetchcache

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// A BINARY BUILT WITHOUT THE ENGINE MUST SAY SO, LOUDLY, THROUGH THE FUSED PATH. This
// file compiles only on the stub build — the exact complement of the engine's gate — and
// deliberately does NOT stub DefaultPageEngine: what is under test is the production
// wiring's answer when the engine is absent. An empty reading here would be the plausible
// zero [[facts-are-fields]] names, indistinguishable from a document of blank pages.
func TestAScannedReadOnAStubBuildRefusesLoudly(t *testing.T) {
	run, e := storeScanned(t, 1)
	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err == nil {
		t.Fatal("the engineless build produced a reading")
	}
	if !errors.Is(err, tessocr.ErrNotCompiledIn) {
		t.Errorf("err = %v, want it to wrap tessocr.ErrNotCompiledIn so callers can classify it", err)
	}
	// The sentence a seat meets in ocr_reason: what is missing, that release binaries have
	// it, and how a builder gets it.
	for _, want := range []string{"not compiled into this binary", "release binaries", "-tags tessocr"} {
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
