package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CITATION OF OCR TEXT CARRIES ITS PAGE, AND RED CHECKS THAT PAGE'S PIXELS (#986). These drive
// the real command tree over a real rasterised scan, with the OCR engine faked so each page reads
// as text the test states: the page a quote is found on is derived by the tool, never typed.

const (
	ocrURL    = "https://ex/standard.pdf"
	ocrClaim  = "The standard defines integrity levels."
	ocrReport = "# Findings\n\n" + ocrClaim + "\n"
)

// scanPDFPages is scanPDF with n image-only pages sharing one image.
//
// THE PAGE IS 72x90 POINTS, NOT US LETTER, AND THAT IS THE POINT. These tests ask which PAGE a
// quote sits on; nothing here reads a pixel. A 612x792 page renders 2550x3300 px per page at the
// reader's 300 DPI, 74x the pixels, and the engine that would look at them is stubbed out. The
// fixture is sized to what is under test.
func scanPDFPages(n int) []byte {
	kids := make([]string, n)
	objs := []string{"<< /Type /Catalog /Pages 2 0 R >>", ""}
	for i := 0; i < n; i++ {
		kids[i] = itoa(5+i) + " 0 R"
	}
	objs[1] = "<< /Type /Pages /Kids [" + strings.Join(kids, " ") + "] /Count " + itoa(n) + " >>"
	objs = append(objs,
		"<< /Length 29 >>\nstream\nq 72 0 0 90 0 0 cm /Im1 Do Q\nendstream",
		"<< /Type /XObject /Subtype /Image /Width 2 /Height 2 /ColorSpace /DeviceRGB "+
			"/BitsPerComponent 8 /Filter /ASCIIHexDecode /Length 25 >>\nstream\n000000FFFFFF000000FFFFFF>\nendstream")
	for i := 0; i < n; i++ {
		objs = append(objs, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 72 90] /Contents 3 0 R "+
			"/Resources << /XObject << /Im1 4 0 R >> >> >>")
	}
	var out strings.Builder
	out.WriteString("%PDF-1.4\n")
	offs := make([]int, 0, len(objs))
	for i, o := range objs {
		offs = append(offs, out.Len())
		out.WriteString(itoa(i+1) + " 0 obj\n" + o + "\nendobj\n")
	}
	xref := out.Len()
	out.WriteString("xref\n0 " + itoa(len(objs)+1) + "\n0000000000 65535 f \n")
	for _, off := range offs {
		out.WriteString(pad10(off) + " 00000 n \n")
	}
	out.WriteString("trailer\n<< /Size " + itoa(len(objs)+1) + " /Root 1 0 R >>\nstartxref\n" + itoa(xref) + "\n%%EOF\n")
	return []byte(out.String())
}

// ocrRun is a run holding the report, a registered blue and lens, and a scanned PDF fetched and
// read — page n reading as pages[n-1]. It returns the run dir and the document's sha.
func ocrRun(t *testing.T, pages []string, fetchArgs ...string) (string, string) {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, ocrReport)
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)
	body := scanPDFPages(len(pages))
	withFetcher(t, &fakeFetcher{
		resp:        map[string][]byte{ocrURL: body, "https://ex/page.html": []byte("<html>integrity levels</html>")},
		contentType: "application/pdf",
	})
	withExtractor(t, stubExtractor{out: scannedPDF(len(pages))})
	withCLIEngine(t, func(n int) (string, error) { return pages[n-1], nil })
	if out, err := run(t, append([]string{"fetch", "--seat-id", "operator", "--run", runDir, "--url", ocrURL}, fetchArgs...)...); err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	return runDir, fetchcache.Sha(body)
}

func ocrCite(t *testing.T, runDir string, args ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", ocrClaim, "--url", ocrURL, "--title", "IEEE 1012"}, args...)...)
}

func mustRefuse(t *testing.T, err error, out string, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("accepted; want a refusal naming %q. out:\n%s", want, out)
	}
	for _, w := range want {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("refusal does not say %q:\n%v", w, err)
		}
	}
}

var threePages = []string{
	"Annex A through G\nfollow the scope.",
	"The integrity levels are four.\nAnnexes A through J are informa-\ntive.",
	"Clause 5 covers\nthe life cycle.",
}

func TestOCRCiteRecordsThePageItsSpanSitsOn(t *testing.T) {
	runDir, sha := ocrRun(t, threePages)
	if out, err := ocrCite(t, runDir, "--source-text", "leaf", "--ocr-quote", "Annexes A through J"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	rec, _, err := fetchcache.ReadReadingRecord(runtest.Open(t, runDir), sha)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.GetPages(); len(got) != 1 || got[0] != 2 {
		t.Errorf("pages = %v, want [2]", got)
	}
	if c.GetSourceTextOrigin() != recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR {
		t.Errorf("origin = %v, want ocr", c.GetSourceTextOrigin())
	}
	if c.GetOcrEngine() != rec.Engine || c.GetOcrTextSha() != rec.TextSha || c.GetOcrQuote() != "Annexes A through J" {
		t.Errorf("pins = %q %q %q, want the reading's %q %q and the span", c.GetOcrEngine(), c.GetOcrTextSha(), c.GetOcrQuote(), rec.Engine, rec.TextSha)
	}
}

func TestOCRCiteRecordsEveryPageThatHoldsTheSpan(t *testing.T) {
	runDir, _ := ocrRun(t, []string{"integrity levels", "nothing", "the integrity\nlevels"})
	if out, err := ocrCite(t, runDir, "--ocr-quote", "integrity levels"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if got := firstCiteEvent(t, runDir).GetPages(); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Errorf("pages = %v, want [1 3]", got)
	}
}

func TestOCRCiteRefusesASpanItCannotPlaceOnOnePage(t *testing.T) {
	runDir, sha := ocrRun(t, threePages)
	cases := []struct {
		name, span string
		want       []string
	}{
		{"straddle", "follow the scope. The integrity levels", []string{"runs from page 1 onto page 2", "quote the part on one page"}},
		{"hyphen", "are informative", []string{`found on page 2 only as "informa-"`, "quote it as the reading has it"}},
		{"absent", "Annexes A through Q", []string{"not in the reading at", fetchcache.OCRTextPath(runtest.Open(t, runDir), sha)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ocrCite(t, runDir, "--ocr-quote", tc.span)
			mustRefuse(t, err, out, tc.want...)
		})
	}
	if c := firstCiteEvent(t, runDir); c != nil {
		t.Errorf("a refused cite left an event: %v", c)
	}
}

func TestOCRCiteAtTheLeafRequiresTheSpan(t *testing.T) {
	runDir, _ := ocrRun(t, threePages)
	out, err := ocrCite(t, runDir, "--source-text", "leaf")
	mustRefuse(t, err, out, "a leaf citation of OCR-derived text names the span", "--ocr-quote")

	if out, err := ocrCite(t, runDir, "--source-text", "summary_only"); err != nil {
		t.Fatalf("a summary_only cite of OCR text with no span was refused: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	if c.GetSourceTextOrigin() != recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR || len(c.GetPages()) != 0 || c.OcrQuote != nil {
		t.Errorf("summary_only cite = origin %v pages %v quote %v; want ocr, no pages, no quote", c.GetSourceTextOrigin(), c.GetPages(), c.OcrQuote)
	}
}

func TestOCRCiteRefusesTheSpanWhereThereIsNoReading(t *testing.T) {
	t.Run("embedded", func(t *testing.T) {
		runDir, _ := ocrRun(t, threePages)
		withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://ex/page.html": []byte("<html>integrity levels</html>")}, contentType: "text/html"})
		withExtractor(t, stubExtractor{})
		out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", ocrClaim,
			"--url", "https://ex/page.html", "--title", "Page", "--ocr-quote", "integrity levels")
		mustRefuse(t, err, out, "this source's text is not OCR-derived")

		if out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", ocrClaim,
			"--url", "https://ex/page.html", "--title", "Page", "--source-text", "leaf"); err != nil {
			t.Fatalf("cite of HTML: %v\n%s", err, out)
		}
		if c := firstCiteEvent(t, runDir); c.GetSourceTextOrigin() != recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED || len(c.GetPages()) != 0 {
			t.Errorf("HTML cite = origin %v pages %v; want embedded, no pages", c.GetSourceTextOrigin(), c.GetPages())
		}
	})
	t.Run("unread scan", func(t *testing.T) {
		runDir, _ := ocrRun(t, threePages, "--ocr=false")
		out, err := ocrCite(t, runDir, "--ocr-quote", "Annexes A through J")
		mustRefuse(t, err, out, "has none", "`fetch --url "+ocrURL+"` reads a scan first")
	})
	t.Run("incomplete reading", func(t *testing.T) {
		runDir, sha := ocrRun(t, threePages)
		if err := os.Remove(fetchcache.PageTextPath(runtest.Open(t, runDir), sha, 3)); err != nil {
			t.Fatal(err)
		}
		out, err := ocrCite(t, runDir, "--ocr-quote", "Annexes A through J")
		mustRefuse(t, err, out, "incomplete on disk", "`fetch --url "+ocrURL+"` re-derives it")
	})
}

// pagedCite records a leaf cite of the p2 span and returns its label.
func pagedCite(t *testing.T, runDir string) string {
	t.Helper()
	if out, err := ocrCite(t, runDir, "--source-text", "leaf", "--ocr-quote", "Annexes A through J"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	return firstCiteEvent(t, runDir).GetLabel()
}

func lensVerify(t *testing.T, runDir, anchor string, args ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--quote", ocrClaim, "--as", "supports", "--confidence", "high", "--reason", "the image reads I, not J"}, args...)...)
}

func firstVerify(t *testing.T, runDir string) *recordpb.Verify {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok {
			return v
		}
	}
	return nil
}

func TestVerifyOfAPagedCitationNamesTheCheckedPageImage(t *testing.T) {
	runDir, sha := ocrRun(t, threePages)
	label := pagedCite(t, runDir)

	out, err := lensVerify(t, runDir, label)
	mustRefuse(t, err, out, "found on PDF p. 2", "render-page --sha "+sha+" --page 2")
	out, err = lensVerify(t, runDir, label, "--page", "3")
	mustRefuse(t, err, out, "--page 3 is not a page citation")
	out, err = lensVerify(t, runDir, label, "--page", "2")
	mustRefuse(t, err, out, "no image of page 2 is on disk", "render-page --sha "+sha+" --page 2")
	if v := firstVerify(t, runDir); v != nil {
		t.Fatalf("a refused verify left an event: %v", v)
	}

	rp, err := run(t, "render-page", "--run", runDir, "--seat-id", lensSeat, "--sha", sha, "--page", "2", "--json")
	if err != nil {
		t.Fatalf("render-page: %v\n%s", err, rp)
	}
	if !strings.Contains(rp, `"matches_reading":true`) {
		t.Errorf("render-page of the page the reading was made from does not match it:\n%s", rp)
	}
	if out, err := lensVerify(t, runDir, label, "--page", "2"); err != nil {
		t.Fatalf("verify --page 2 after render-page: %v\n%s", err, out)
	}
	img, err := os.ReadFile(fetchcache.PageImagePath(runtest.Open(t, runDir), sha, 2))
	if err != nil {
		t.Fatal(err)
	}
	rec, _, _ := fetchcache.ReadReadingRecord(runtest.Open(t, runDir), sha)
	v := firstVerify(t, runDir)
	if v.GetPage() != 2 || v.GetPageRenderSha() != fetchcache.Sha(img) || v.GetReadingRenderSha() != rec.RenderShas[1] {
		t.Errorf("verify stamps page=%d page_render_sha=%q reading_render_sha=%q; want 2, %q, %q",
			v.GetPage(), v.GetPageRenderSha(), v.GetReadingRenderSha(), fetchcache.Sha(img), rec.RenderShas[1])
	}
}

func TestVerifyOfAPagedCitationGoneFromTheCacheIsRefused(t *testing.T) {
	runDir, sha := ocrRun(t, threePages)
	label := pagedCite(t, runDir)
	if _, err := run(t, "render-page", "--run", runDir, "--seat-id", lensSeat, "--sha", sha, "--page", "2"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(fetchcache.PagesDir(runtest.Open(t, runDir), sha)); err != nil {
		t.Fatal(err)
	}
	out, err := lensVerify(t, runDir, label, "--page", "2")
	mustRefuse(t, err, out, "gone from the cache", "`fetch --url "+ocrURL+"` re-derives it")
}

func TestVerifyUnreachableNeedsNoPage(t *testing.T) {
	runDir, _ := ocrRun(t, threePages)
	label := pagedCite(t, runDir)
	if out, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", label,
		"--quote", ocrClaim, "--as", "unreachable", "--confidence", "high", "--reason", "the page would not render"); err != nil {
		t.Fatalf("an unreachable verify with no page was refused: %v\n%s", err, out)
	}
	if v := firstVerify(t, runDir); v.Page != nil {
		t.Errorf("unreachable verify recorded page %d", v.GetPage())
	}
}

func TestVerifyRefusesAPageOnACitationWithoutPages(t *testing.T) {
	runDir, _ := ocrRun(t, threePages)
	if out, err := ocrCite(t, runDir, "--source-text", "summary_only"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	label := firstCiteEvent(t, runDir).GetLabel()
	out, err := lensVerify(t, runDir, label, "--page", "2")
	mustRefuse(t, err, out, "has no pages")
	if out, err := lensVerify(t, runDir, label); err != nil {
		t.Fatalf("verify of a pageless cite with no --page: %v\n%s", err, out)
	}
}

func TestRenderPageRefusesWhatItCannotDraw(t *testing.T) {
	runDir, sha := ocrRun(t, threePages)
	lens := func(args ...string) (string, error) {
		return run(t, append([]string{"render-page", "--run", runDir, "--seat-id", lensSeat}, args...)...)
	}
	out, err := lens("--sha", strings.Repeat("0", 64), "--page", "1")
	mustRefuse(t, err, out, "no cached document has sha")
	for _, p := range []string{"0", "4"} {
		out, err = lens("--sha", sha, "--page", p)
		mustRefuse(t, err, out, "page "+p+" of a 3-page document")
	}

	html, serr := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: "https://ex/h", ContentType: "text/html"}, []byte("<html>x</html>"))
	if serr != nil {
		t.Fatal(serr)
	}
	out, err = lens("--sha", html.Sha, "--page", "1")
	mustRefuse(t, err, out, "only a PDF has pages")
}

func TestRenderPageSaysWhenThereIsNoReading(t *testing.T) {
	runDir, sha := ocrRun(t, threePages, "--ocr=false")
	out, err := run(t, "render-page", "--run", runDir, "--seat-id", lensSeat, "--sha", sha, "--page", "1", "--json")
	if err != nil {
		t.Fatalf("render-page: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"reading":"none"`) || strings.Contains(out, "reading_render_sha") {
		t.Errorf("render-page with no reading must say reading none and print no reading sha:\n%s", out)
	}
}

// A CORRECTION NEVER RE-LOCATES (D7a). The quote's page identifies where it sits, as the url does:
// a corrected title keeps the pages, and a different span is refused as a frozen field.
func TestACorrectedOCRCiteKeepsItsPages(t *testing.T) {
	runDir, _ := ocrRun(t, threePages)
	k := correctionKeyOf(t, runDir, citeSeat, []string{"cite", "--quote", ocrClaim, "--url", ocrURL,
		"--title", "IEEE 1012, as read in this run", "--source-text", "leaf", "--ocr-quote", "Annexes A through J"})
	must(t, runDir, "cite", "--seat-id", citeSeat, "--title", "IEEE 1012",
		"--corrects", k, "--correction-why", "the title narrated the run")
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Title != "IEEE 1012" || len(srcs[0].Pages) != 1 || srcs[0].Pages[0] != 2 {
		t.Fatalf("sources = %+v, want one with the corrected title and pages [2]", srcs)
	}

	k2 := correctionKeyOf(t, runDir, citeSeat, []string{"cite", "--quote", ocrClaim, "--url", ocrURL,
		"--title", "Again", "--ocr-quote", "integrity levels"})
	_, err = run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--title", "Again",
		"--ocr-quote", "The integrity levels are four.", "--corrects", k2, "--correction-why", "wrong span")
	if err == nil || !strings.Contains(err.Error(), "may change only the seat's own wording") {
		t.Fatalf("a correction moving the OCR quote was not refused as a frozen field: %v", err)
	}
}
