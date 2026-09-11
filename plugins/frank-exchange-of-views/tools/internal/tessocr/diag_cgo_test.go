//go:build tessocr && cgo

package tessocr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tesseract's diagnostics reach the result, per page, and not stderr; and a grid page keeps
// the word boxes its reconstruction was built from. The fixture is a real 300-DPI grid crop
// of IEEE 1012, on which tesseract states its resolution estimate once per pass (#644: the
// estimate exists nowhere else). A uniform blank page will not do — tesseract returns before
// saying anything. The second read proves the capture is per page, not cumulative, and that
// the first restored stderr: a capture still open refuses to start (stage 4).
func TestReadPageCapturesDiagnosticsAndEvidence(t *testing.T) {
	en, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer en.Close()
	png, err := os.ReadFile(filepath.Join("testdata", "gridcrop.png"))
	if err != nil {
		t.Fatal(err)
	}

	first, err := en.ReadPage(png, Grid300)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Table {
		t.Fatal("gridcrop.png no longer fires the grid detector; the evidence half of this test needs a grid page")
	}
	if !strings.Contains(first.Diagnostics, "Estimating resolution as") {
		t.Errorf("diagnostics = %q, want tesseract's resolution estimate", first.Diagnostics)
	}
	if first.Evidence.PortraitTSV == "" || first.Evidence.SparseTSV == "" {
		t.Errorf("grid page evidence missing: portrait %d bytes, sparse %d bytes",
			len(first.Evidence.PortraitTSV), len(first.Evidence.SparseTSV))
	}

	second, err := en.ReadPage(png, Grid300)
	if err != nil {
		t.Fatalf("second capture: %v", err)
	}
	if second.Diagnostics != first.Diagnostics {
		t.Errorf("the same page captured %q then %q — the capture is not per page", first.Diagnostics, second.Diagnostics)
	}
}
