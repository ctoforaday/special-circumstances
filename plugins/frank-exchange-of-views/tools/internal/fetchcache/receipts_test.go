package fetchcache

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// The receipt tests, ported from the blocked-page suite whose feature left with the API
// (plans/local-ocr.md §II). The receipts survive as the per-page provenance rows the
// record projects from: their resume purpose retired with the model spend, but the
// render-sha validation — never serve a reading for pixels it did not come from — is
// theirs, and it outlives the dollars.

// THE PAGE-79 PROBLEM, STILL CLOSED: a re-run after a mid-document failure re-derives
// only the pages that were never read. The stake is no longer money — it is that a
// re-derived page would silently replace text a seat may already have cited from.
func TestResumeRederivesOnlyUnreadPages(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withEngine(t, textEngine(func(n int) (string, error) {
		if n == 2 {
			return "", errors.New("transient engine failure")
		}
		return "first attempt", nil
	}))
	if _, err := ReadRenderedPages(run, sha, rd); err == nil {
		t.Fatal("the seeded failure did not fail")
	}
	// Page 1's receipt survived the fatal error — that is what resume spends.
	if _, ok, rerr := readReceipt(run, sha, 1); rerr != nil || !ok {
		t.Fatalf("page 1's receipt did not survive the fatal error (%v)", rerr)
	}
	// And no record was left: a fatal read must not look like a reading.
	if _, had, _ := ReadReadingRecord(run, sha); had {
		t.Fatal("a fatal read left a reading record — the failure that looks like success")
	}

	fe := textEngine(func(int) (string, error) { return "second attempt", nil })
	withEngine(t, fe)
	rec, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatal(err)
	}
	if fe.calls != 2 {
		t.Errorf("the resume ran the engine %d times, want 2 — page 1 already had a valid receipt", fe.calls)
	}
	if got, _ := os.ReadFile(PageTextPath(run, sha, 1)); string(got) != "first attempt" {
		t.Errorf("page 1 = %q, want the first attempt's text reused, not re-derived", got)
	}
	if len(rec.Pages) != 3 {
		t.Errorf("record has %d pages, want 3", len(rec.Pages))
	}
}

// A malformed receipt is an ERROR, never an absence — and so is a receipt whose text has
// been altered underneath it: both would otherwise re-derive or mis-attest silently.
func TestACorruptReceiptOrTextRefuses(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withEngine(t, textEngine(func(int) (string, error) { return "x", nil }))
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(receiptPath(run, sha, 2), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(run, sha, rd); err == nil || !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("malformed receipt = %v, want a refusal naming it", err)
	}

	// Remove the bad receipt (page 2 re-reads), then corrupt the text under a good one.
	if err := os.Remove(receiptPath(run, sha, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatalf("re-reading page 2 after receipt removal: %v", err)
	}
	if err := os.WriteFile(PageTextPath(run, sha, 2), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(run, sha, rd); err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Errorf("tampered text under a receipt = %v, want the pair named corrupt", err)
	}
}

// THE OLD WHOLESALE CLEAR'S GUARANTEE, KEPT: different pixels mismatch every receipt, so
// nothing rendered at one resolution is ever served as a reading of another. The receipt
// key is the render sha, and a re-render changes every one.
func TestChangedPixelsReplaceEveryReceipt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("a@old"), []byte("b@old")}, tessocr.RenderDPI)
	withEngine(t, textEngine(func(int) (string, error) { return "old pixels", nil }))
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}

	rd2 := fakeRender(t, run, sha, [][]byte{[]byte("a@new"), []byte("b@new")}, tessocr.RenderDPI)
	fe := textEngine(func(int) (string, error) { return "new pixels", nil })
	withEngine(t, fe)
	if _, err := ReadRenderedPages(run, sha, rd2); err != nil {
		t.Fatal(err)
	}
	if fe.calls != 2 {
		t.Errorf("re-render re-derived %d pages, want 2 — no old receipt may serve new pixels", fe.calls)
	}
}

// AN ENGINE CHANGE IS A DIFFERENT DERIVATION. A receipt is reusable only by the identity
// that wrote it: pins bump, output changes, and a reading served across that boundary
// would attest to an engine that never produced it.
func TestAReceiptFromAnotherEngineIsNotReused(t *testing.T) {
	run, sha, rd := onePageRender(t)
	old := &fakeEngine{id: "fake@old", perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: "old engine"}, nil
	}}
	withEngine(t, old)
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}

	renewed := &fakeEngine{id: "fake@new", perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: "new engine"}, nil
	}}
	withEngine(t, renewed)
	rec, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatal(err)
	}
	if renewed.calls != 1 {
		t.Errorf("the new engine ran %d times, want 1 — the old engine's receipt must not serve", renewed.calls)
	}
	if rec.Engine != "fake@new" {
		t.Errorf("record engine = %q, want the identity that actually derived it", rec.Engine)
	}
	if got, _ := os.ReadFile(PageTextPath(run, sha, 1)); string(got) != "new engine" {
		t.Errorf("page text = %q, want the new derivation", got)
	}
}

// --force MEANS FORCE: receipts discarded, every page derived again.
func TestClearReceiptsMakesForceAFullReRead(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withEngine(t, textEngine(func(int) (string, error) { return "first", nil }))
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}
	if err := ClearReceipts(run, sha); err != nil {
		t.Fatal(err)
	}
	fe := textEngine(func(int) (string, error) { return "again", nil })
	withEngine(t, fe)
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}
	if fe.calls != 3 {
		t.Errorf("--force derived %d pages, want all 3", fe.calls)
	}
}
