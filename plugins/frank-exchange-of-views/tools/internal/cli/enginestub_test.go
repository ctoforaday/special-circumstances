//go:build !tessocr || !cgo

package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE FUSED PATH ON AN ENGINELESS BINARY, ALL THE WAY TO THE SUMMARY. This file compiles
// only on the stub build, and unlike every other test here it stubs NEITHER the scan
// reader NOR the page engine: a dev-built `go build` binary (no -tags tessocr) fetching a
// scan must exit 0 with ocr_reason carrying the engine-absent sentence — never an empty
// reading, and never a failed fetch. The real rasteriser runs, so this also proves the
// refusal fires per the wiring and not from a shortcut.
func TestFetchOnAStubBuildStatesTheEngineIsAbsent(t *testing.T) {
	dir := recordtest.TmpRun(t)
	withFetcher(t, &fakeFetcher{
		resp:        map[string][]byte{"https://ex/scan.pdf": scanPDF(t)},
		contentType: "application/pdf",
	})
	withExtractor(t, stubExtractor{out: scannedPDF(1)})

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/scan.pdf")
	if err != nil {
		t.Fatalf("an engineless binary failed the fetch itself: %v\n%s", err, out)
	}
	for _, want := range []string{
		"ocr_reason:",
		"not compiled into this binary",
		"release binaries",
		"-tags tessocr",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary is missing %q — the engineless case must be LOUD. got:\n%s", want, out)
		}
	}
	// And the absence stayed honest: no ocr-derived text was claimed.
	if strings.Contains(out, "ocr_derived: true") {
		t.Errorf("an engineless binary claimed a reading:\n%s", out)
	}
}
