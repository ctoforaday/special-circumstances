package seatprobe

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
)

// THE FIXTURE IS A SCAN: one page and no text layer, or fetch extracts its text and nothing is
// ever read by the engine — the board would cite embedded text and demand a page check it can
// never owe.
func TestScannedFixtureHasNoTextLayer(t *testing.T) {
	ex := fetchcache.PDFExtractor{}.Extract(t.TempDir(), "application/pdf", scannedFixture)
	if !ex.Attempted || ex.Text != "" || ex.Pages != 1 {
		t.Fatalf("scanned-1p.pdf: attempted=%v pages=%d text=%q; want an attempted, one-page extraction with no text", ex.Attempted, ex.Pages, ex.Text)
	}
}
