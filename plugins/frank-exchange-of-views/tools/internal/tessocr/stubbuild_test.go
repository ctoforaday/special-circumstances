package tessocr

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

// The default-build contract as a test: with no helper (NATIVE_HELPER_DIR=none),
// the engine refuses loudly with helper-absent.
func TestDefaultBuildRefusesWithoutHelper(t *testing.T) {
	t.Setenv("NATIVE_HELPER_DIR", "none")
	resetDefaultClient()

	en, err := New()
	if err == nil {
		en.Close()
		t.Fatal("New() succeeded without a helper")
	}
	if !errors.Is(err, nativehelper.ErrAbsent) {
		t.Fatalf("err = %v, want ErrAbsent", err)
	}
	if !strings.Contains(err.Error(), "helper-absent: feov-ocr") {
		t.Fatalf("err = %q, want helper-absent sentence", err)
	}
}
