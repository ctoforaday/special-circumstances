//go:build !tessocr || !cgo

package tessocr

import (
	"errors"
	"testing"
)

// Under the stub build the refusal must be the NAMED error on every entry point — a stub
// that returned ("", nil) anywhere would read as a blank page.
func TestStubRefusesLoudly(t *testing.T) {
	if _, err := New(); !errors.Is(err, ErrNotCompiledIn) {
		t.Errorf("New() = %v, want ErrNotCompiledIn", err)
	}
	var en Engine
	if _, err := en.PageText(nil); !errors.Is(err, ErrNotCompiledIn) {
		t.Errorf("PageText = %v, want ErrNotCompiledIn", err)
	}
	if _, err := en.PageTSV(nil, PSMAuto); !errors.Is(err, ErrNotCompiledIn) {
		t.Errorf("PageTSV = %v, want ErrNotCompiledIn", err)
	}
	if _, err := en.RotatedBand(nil, 0, 0, 1, 1); !errors.Is(err, ErrNotCompiledIn) {
		t.Errorf("RotatedBand = %v, want ErrNotCompiledIn", err)
	}
	if _, err := DetectGrid(nil, Grid300); !errors.Is(err, ErrNotCompiledIn) {
		t.Errorf("DetectGrid = %v, want ErrNotCompiledIn", err)
	}
}
