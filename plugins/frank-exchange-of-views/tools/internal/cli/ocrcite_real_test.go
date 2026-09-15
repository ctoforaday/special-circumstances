//go:build tessocr

package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE MISREAD THAT STARTED THIS, END TO END (#644, #986). Page 10 of IEEE 1012 reads "Annexes A
// through J" where the scan says "I". A citation of that span records page 10, red draws page 10
// and records the image it checked, and the assembled report prints the page in the note and the
// source once in the Bibliography. Real HTTP fetch, real renderer, real engine, over the real
// document served on loopback (this suite refuses outside hosts): env-gated, because it names a
// local copy of an 80-page PDF that is not checked in.
func TestOCRCiteOnRealScan(t *testing.T) {
	path := os.Getenv("FEOV_OCR_E2E")
	if path == "" {
		t.Skip("set FEOV_OCR_E2E=<path to IEEE 1012's PDF> to read it with the real engine")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	url := srv.URL + "/1012.pdf"
	const claim = "The standard's annexes are informative."
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+claim+"\n")
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", runDir, "--url", url, "--json")
	if err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	var sum struct {
		Sha256     string `json:"sha256"`
		OCRDerived bool   `json:"ocr_derived"`
		Pages      int    `json:"pages"`
	}
	if err := json.Unmarshal([]byte(out), &sum); err != nil || !sum.OCRDerived || sum.Pages != 80 {
		t.Fatalf("fetch summary = %+v (%v), want ocr_derived over 80 pages:\n%s", sum, err, out)
	}

	cite := func(span string) (string, error) {
		return run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", claim, "--url", url,
			"--title", "IEEE Std 1012-1998", "--source-text", "leaf", "--ocr-quote", span)
	}
	if out, err := cite("Annexes A through Q"); err == nil || !strings.Contains(err.Error(), "not in the reading at") {
		t.Fatalf("a span the reading does not hold was not refused: %v\n%s", err, out)
	}
	if out, err := cite("Annexes A through J"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	if got := c.GetPages(); len(got) != 1 || got[0] != 10 {
		t.Fatalf("pages = %v, want [10]", got)
	}
	t.Logf("cite %s: pages=%v engine=%s text_sha=%s", c.GetLabel(), c.GetPages(), c.GetOcrEngine(), c.GetOcrTextSha())

	if out, err := lensVerify(t, runDir, c.GetLabel()); err == nil || !strings.Contains(err.Error(), "render-page --sha "+sum.Sha256+" --page 10") {
		t.Fatalf("a verify naming no page was not refused with the render-page command: %v\n%s", err, out)
	}
	rp, err := run(t, "render-page", "--run", runDir, "--seat-id", lensSeat, "--sha", sum.Sha256, "--page", "10", "--json")
	if err != nil || !strings.Contains(rp, `"matches_reading":true`) || !strings.Contains(rp, `"dpi":300`) {
		t.Fatalf("render-page p10: %v\n%s", err, rp)
	}
	t.Logf("render-page: %s", strings.TrimSpace(rp))
	if out, err := lensVerify(t, runDir, c.GetLabel(), "--page", "10"); err != nil {
		t.Fatalf("verify --page 10: %v\n%s", err, out)
	}
	v := firstVerify(t, runDir)
	img, _ := os.ReadFile(fetchcache.PageImagePath(runtest.Open(t, runDir), sum.Sha256, 10))
	if v.GetPage() != 10 || len(v.GetPageRenderSha()) != 64 || v.GetPageRenderSha() != fetchcache.Sha(img) {
		t.Fatalf("verify page=%d page_render_sha=%q, want 10 and the image's sha", v.GetPage(), v.GetPageRenderSha())
	}
	t.Logf("verify: page=%d page_render_sha=%s reading_render_sha=%s", v.GetPage(), v.GetPageRenderSha(), v.GetReadingRenderSha())

	md := assembled(t, runDir)
	if !strings.Contains(md, "IEEE Std 1012-1998, PDF p. 10. "+url) {
		t.Errorf("the assembled note does not carry the page:\n%s", md)
	}
	bib := md[strings.Index(md, "## Bibliography"):]
	if strings.Count(bib, url) != 1 || strings.Contains(bib, "PDF p.") {
		t.Errorf("the Bibliography does not list the source once, page-less:\n%s", bib)
	}
	t.Logf("assembled tail:\n%s", md[max(0, len(md)-600):])
}
