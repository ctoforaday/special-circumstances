package tessocr

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

// Under the no-helper configuration (NATIVE_HELPER_DIR=none), the refusal must be
// helper-absent on every entry point — a missing helper that returned ("", nil) anywhere
// would read as a blank page.
func TestNoHelperRefusesLoudly(t *testing.T) {
	t.Setenv("NATIVE_HELPER_DIR", "none")
	resetDefaultClient()

	if _, err := New(); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("New() = %v, want ErrAbsent", err)
	}
	var en Engine
	dummyPNG := []byte("fake-png-data")
	if _, err := en.PageTSV(dummyPNG, PSMAuto); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("PageTSV = %v, want ErrAbsent", err)
	}
	if _, err := en.RotatedBand(dummyPNG, 0, 0, 1, 1); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("RotatedBand = %v, want ErrAbsent", err)
	}
	if _, err := DetectGrid(dummyPNG, Grid300); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("DetectGrid = %v, want ErrAbsent", err)
	}
	if _, err := InkFraction(dummyPNG); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("InkFraction = %v, want ErrAbsent", err)
	}
	if _, err := en.ReadPage(dummyPNG, Grid300); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("ReadPage = %v, want ErrAbsent", err)
	}
	if _, err := GridLines(dummyPNG, Grid300); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("GridLines = %v, want ErrAbsent", err)
	}
	if _, err := RepairedRules(dummyPNG, Grid300); !errors.Is(err, nativehelper.ErrAbsent) {
		t.Errorf("RepairedRules = %v, want ErrAbsent", err)
	}

	_, err := New()
	if !strings.Contains(err.Error(), "helper-absent: feov-ocr") {
		t.Errorf("err = %q, want helper-absent sentence", err)
	}
}
