package fetchcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// filterErr fabricates the content filter's refusal in the shape production sees: a typed
// *anthropic.Error with status 400 whose Error() carries the exact sentence. The SDK type's
// Error() dereferences Request and Response, so both are stubbed; the sentence rides a
// wrapper because the SDK keeps the raw body unexported — the classified VALUE is
// structurally identical either way (typed 400, sentence in the message).
func filterErr() error {
	u, _ := url.Parse("https://api.anthropic.com/v1/messages")
	apiErr := &anthropic.Error{
		StatusCode: 400,
		Request:    &http.Request{Method: "POST", URL: u},
		Response:   &http.Response{StatusCode: 400},
		RequestID:  "req_test",
	}
	return fmt.Errorf(`%w {"type":"error","error":{"type":"invalid_request_error","message":"Output blocked by content filtering policy"}}`, apiErr)
}

// fakeRender writes a render a test controls page by page: arbitrary bytes stand in for
// images (nothing in the read loop decodes them; the receipt key is their hash), and the
// render record binds them exactly as RenderPages would.
func fakeRender(t *testing.T, run record.Run, sha string, pages [][]byte, dpi int) RenderRecord {
	t.Helper()
	dir := PagesDir(run, sha)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rec := RenderRecord{Sha: sha, DPI: dpi, RenderedAt: time.Now().UTC(), Renderer: "test"}
	for i, b := range pages {
		if err := os.WriteFile(PagePath(run, sha, i+1), b, 0o644); err != nil {
			t.Fatal(err)
		}
		rec.PageShas = append(rec.PageShas, Sha(b))
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(renderRecordPath(run, sha), append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return rec
}

func threePageRender(t *testing.T) (record.Run, string, RenderRecord) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("page-one"), []byte("page-two"), []byte("page-three")}, 72)
	return run, sha, rd
}

// The classifier's both directions, pinned because it is a string-shaped key whose failure
// direction must stay CLOSED: the real shape classifies; a reworded message, an untyped
// error carrying the sentence, and a typed 400 without the sentence all fall through to
// the fatal path.
func TestBlockedByPolicyClassifiesBothDirections(t *testing.T) {
	if !blockedByPolicy(filterErr()) {
		t.Error("the production-shaped refusal was not classified")
	}
	u, _ := url.Parse("https://api.anthropic.com/v1/messages")
	reworded := fmt.Errorf("%w: output withheld by safety system", &anthropic.Error{
		StatusCode: 400, Request: &http.Request{Method: "POST", URL: u}, Response: &http.Response{StatusCode: 400},
	})
	if blockedByPolicy(reworded) {
		t.Error("a reworded message classified — the key must fail CLOSED, not fuzzy-match")
	}
	if blockedByPolicy(errors.New("Output blocked by content filtering policy")) {
		t.Error("an untyped error carrying the sentence classified — the type is half the key")
	}
}

// THE HOLE IS NAMED IN EVERY REGISTER: marker in the page text and the assembled document,
// blocked fields on the record, and the other pages read normally around it.
func TestABlockedPageBecomesANamedHole(t *testing.T) {
	sr := &stubReader{in: 10, out: 5, perCall: func(n int) (string, error) {
		if n == 2 {
			return "", filterErr()
		}
		return fmt.Sprintf("text of call %d", n), nil
	}}
	withReader(t, sr)
	run, sha, rd := threePageRender(t)

	rec, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatalf("a blocked page failed the read: %v", err)
	}
	if len(rec.Pages) != 3 || rec.BlockedPages() != 1 {
		t.Fatalf("pages = %d blocked = %d, want 3 and 1", len(rec.Pages), rec.BlockedPages())
	}
	p2 := rec.Pages[1]
	if !p2.Blocked || !strings.Contains(p2.BlockedReason, "content filtering") || !strings.Contains(p2.BlockedReason, "req_test") {
		t.Errorf("page 2 = %+v, want blocked with the API's verbatim reason incl. request id", p2)
	}
	body, _ := os.ReadFile(OCRTextPath(run, sha))
	if !strings.Contains(string(body), "[page 2: output blocked by content filtering policy]") {
		t.Errorf("assembled text carries no marker at the hole:\n%s", body)
	}
	if !strings.Contains(string(body), "text of call 1") || !strings.Contains(string(body), "text of call 3") {
		t.Errorf("the readable pages did not survive around the hole:\n%s", body)
	}
}

// A failure that is NOT the filter keeps the old semantics exactly: fatal, no record —
// but the pages already read leave receipts, which is what the next test spends.
func TestANonFilterErrorStaysFatal(t *testing.T) {
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 2 {
			return "", errors.New("credentials refused (401)")
		}
		return "ok", nil
	}})
	run, sha, rd := threePageRender(t)

	_, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err == nil || !strings.Contains(err.Error(), "credentials refused") {
		t.Fatalf("err = %v, want the reader's own fatal message", err)
	}
	if _, had, _ := ReadReadingRecord(run, sha); had {
		t.Error("a fatal read left a reading record — the failure that looks like success")
	}
	if _, ok, rerr := readReceipt(run, sha, 1); rerr != nil || !ok {
		t.Errorf("page 1's receipt did not survive the fatal error (%v) — resume has nothing to spend", rerr)
	}
}

// THE PAGE-79 PROBLEM, CLOSED: a re-run after a mid-document failure pays only for the
// pages that were never read.
func TestResumeSpendsOnlyUnreadPages(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 2 {
			return "", errors.New("transient network failure")
		}
		return "first attempt", nil
	}})
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err == nil {
		t.Fatal("the seeded failure did not fail")
	}

	sr := &stubReader{perCall: func(int) (string, error) { return "second attempt", nil }}
	withReader(t, sr)
	rec, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatal(err)
	}
	if sr.calls != 2 {
		t.Errorf("the resume spent %d calls, want 2 — page 1 was already paid for", sr.calls)
	}
	if got, _ := os.ReadFile(PageTextPath(run, sha, 1)); string(got) != "first attempt" {
		t.Errorf("page 1 = %q, want the first attempt's text reused, not re-read", got)
	}
	if len(rec.Pages) != 3 {
		t.Errorf("record has %d pages, want 3", len(rec.Pages))
	}
}

// A DETERMINISTIC REFUSAL IS A FACT ALREADY PAID FOR. With the reading record gone (the
// crash case), a re-run reconstructs the whole reading from receipts — blocked page
// included — for zero model calls.
func TestABlockedReceiptIsReusedWithoutACall(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 2 {
			return "", filterErr()
		}
		return "text", nil
	}})
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(readingRecordPath(run, sha)); err != nil {
		t.Fatal(err)
	}

	sr := &stubReader{perCall: func(int) (string, error) { return "must not be asked", nil }}
	withReader(t, sr)
	rec, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatal(err)
	}
	if sr.calls != 0 {
		t.Errorf("reconstruction spent %d calls, want 0 — every page had a receipt", sr.calls)
	}
	if rec.BlockedPages() != 1 || !rec.Pages[1].Blocked {
		t.Errorf("the blocked fact did not survive reconstruction: %+v", rec.Pages)
	}
}

// A malformed receipt is an ERROR, never an absence — and so is a receipt whose text has
// been altered underneath it: both would otherwise re-spend or mis-attest silently.
func TestACorruptReceiptOrTextRefuses(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "x", nil }})
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(receiptPath(run, sha, 2), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err == nil || !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("malformed receipt = %v, want a refusal naming it", err)
	}

	// Restore the receipt, corrupt the text instead.
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err == nil {
		t.Fatal("expected the malformed receipt to still refuse before restoration")
	}
	if err := os.Remove(receiptPath(run, sha, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatalf("re-reading page 2 after receipt removal: %v", err)
	}
	if err := os.WriteFile(PageTextPath(run, sha, 2), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Errorf("tampered text under a receipt = %v, want the pair named corrupt", err)
	}
}

// THE OLD WHOLESALE CLEAR'S GUARANTEE, RESTATED AS A TEST: different pixels mismatch every
// receipt, so nothing rendered at one resolution is ever served as a reading of another.
func TestChangedPixelsReplaceEveryReceipt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("a@72"), []byte("b@72")}, 72)
	sr := &stubReader{perCall: func(int) (string, error) { return "low", nil }}
	withReader(t, sr)
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatal(err)
	}

	rd2 := fakeRender(t, run, sha, [][]byte{[]byte("a@150"), []byte("b@150")}, 150)
	sr2 := &stubReader{perCall: func(int) (string, error) { return "high", nil }}
	withReader(t, sr2)
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd2); err != nil {
		t.Fatal(err)
	}
	if sr2.calls != 2 {
		t.Errorf("re-render spent %d calls, want 2 — no 72-DPI receipt may serve 150-DPI pixels", sr2.calls)
	}
}

// --force MEANS FORCE: receipts discarded, every page re-read, blocked pages included.
func TestClearReceiptsMakesForceAFullReRead(t *testing.T) {
	run, sha, rd := threePageRender(t)
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 2 {
			return "", filterErr()
		}
		return "first", nil
	}})
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatal(err)
	}
	if err := ClearReceipts(run, sha); err != nil {
		t.Fatal(err)
	}
	sr := &stubReader{perCall: func(int) (string, error) { return "unblocked now", nil }}
	withReader(t, sr)
	rec, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatal(err)
	}
	if sr.calls != 3 {
		t.Errorf("--force spent %d calls, want 3 — the blocked page must be re-attempted", sr.calls)
	}
	if rec.BlockedPages() != 0 {
		t.Errorf("a lifted block survived --force: %+v", rec.Pages)
	}
}
