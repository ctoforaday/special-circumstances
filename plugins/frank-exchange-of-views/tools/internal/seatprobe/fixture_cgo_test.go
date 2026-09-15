//go:build tessocr && cgo

package seatprobe

import (
	"context"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BOARD'S CITATION IS ONE ITS OWN FIXTURE CAN PRODUCE. lens-ocr-verify quotes
// ScannedFixtureSpan from the reading of scanned-1p.pdf; this runs the REAL engine over that scan
// and requires the span on page 1. An engine or traineddata pin bump that changes what it reads is
// caught here, not in a probe run that reports a seat miss. CI anchors on this test's name.
func TestScannedFixtureSpanIsFoundOnPage1ByTheRealEngine(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	no := false
	e, err := fetchcache.Store(run, fetchcache.Entry{URL: "https://probe/scan.pdf", ContentType: "application/pdf", TextExtracted: &no, Pages: 1}, scannedFixture)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := fetchcache.RenderAndRead{}.ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatalf("the real engine did not read the fixture: %v", err)
	}
	loc, err := fetchcache.LocateSpan(run, rec, ScannedFixtureSpan)
	if err != nil {
		t.Fatalf("the real engine's reading does not hold the board's span: %v", err)
	}
	if len(loc.Pages) != 1 || loc.Pages[0] != 1 {
		t.Errorf("span found on pages %v, want [1]", loc.Pages)
	}
}
