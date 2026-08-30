package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A minimal PDF with one image-only page — structurally what a 1998 Acrobat PDFWriter scan
// is, and the document class this whole verb exists for. Built rather than pasted so the
// absence of a text layer is legible in the source.
func scanPDF(t *testing.T) []byte {
	t.Helper()
	const hexPixels = "000000FFFFFF000000FFFFFF>"
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R " +
			"/Resources << /XObject << /Im1 5 0 R >> >> >>",
		"<< /Length 31 >>\nstream\nq 612 0 0 792 0 0 cm /Im1 Do Q\nendstream",
		"<< /Type /XObject /Subtype /Image /Width 2 /Height 2 /ColorSpace /DeviceRGB " +
			"/BitsPerComponent 8 /Filter /ASCIIHexDecode /Length 25 >>\nstream\n" + hexPixels + "\nendstream",
	}
	var out strings.Builder
	out.WriteString("%PDF-1.4\n")
	offs := make([]int, 0, len(objs))
	for i, o := range objs {
		offs = append(offs, out.Len())
		out.WriteString(strings.Join([]string{itoa(i + 1), " 0 obj\n", o, "\nendobj\n"}, ""))
	}
	xref := out.Len()
	out.WriteString("xref\n0 " + itoa(len(objs)+1) + "\n0000000000 65535 f \n")
	for _, off := range offs {
		out.WriteString(pad10(off) + " 00000 n \n")
	}
	out.WriteString("trailer\n<< /Size " + itoa(len(objs)+1) + " /Root 1 0 R >>\nstartxref\n" + itoa(xref) + "\n%%EOF\n")
	return []byte(out.String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func pad10(n int) string {
	s := itoa(n)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}

// cacheScan stores a scan in a fresh run's cache and returns the run dir and its sha.
func cacheScan(t *testing.T, textExtracted *bool, contentType string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	run := runtest.New(t, dir)
	e, err := fetchcache.Store(run, fetchcache.Entry{
		URL:           "https://ex/scan.pdf",
		ContentType:   contentType,
		TextExtracted: textExtracted,
	}, scanPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	return dir, e.Sha
}

func TestOCRPagesRendersAScanAndNamesTheImages(t *testing.T) {
	no := false
	dir, sha := cacheScan(t, &no, "application/pdf")

	out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "72", "--run", dir)
	if err != nil {
		t.Fatalf("ocr pages: %v\n%s", err, out)
	}
	for _, want := range []string{"pages: 1", "dpi: 72", "pages_dir:", "first_page:", "renderer:"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
	// NEVER AN IMAGE, NEVER THE DOCUMENT. The summary names files; a seat opens what it needs.
	if strings.Contains(out, "%PDF") || strings.Contains(out, "\x89PNG") {
		t.Error("the summary carried document or image bytes")
	}
}

// THE GUARD THIS VERB IS MOST LIKELY TO NEED. A document whose text was already extracted
// must not be re-read from pixels: it spends a model to re-derive, less accurately, a file
// that is already on disk.
func TestOCRPagesRefusesADocumentThatAlreadyHasText(t *testing.T) {
	yes := true
	dir, sha := cacheScan(t, &yes, "application/pdf")

	out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--run", dir)
	if err == nil {
		t.Fatalf("a document with an extracted text layer was rendered anyway:\n%s", out)
	}
	msg := err.Error()
	// The refusal must name the remedy and the file, or a seat cannot act on it.
	for _, want := range []string{"already has an extracted text layer", "--force", ".txt"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal = %q, want it to mention %q", msg, want)
		}
	}

	// --force is the escape for the real case: a text layer that is present but wrong.
	if out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "72", "--force", "--run", dir); err != nil {
		t.Errorf("--force did not lift the refusal: %v\n%s", err, out)
	}
}

// A CONTENT TYPE THAT DOES NOT RASTERISE IS NAMED, not silently skipped. An HTML page has no
// pages, and "nothing rendered" would read like a document with zero of them.
func TestOCRPagesRefusesANonPDF(t *testing.T) {
	no := false
	dir, sha := cacheScan(t, &no, "text/html")
	_, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--run", dir)
	if err == nil {
		t.Fatal("text/html was accepted for rasterisation")
	}
	if !strings.Contains(err.Error(), "text/html") || !strings.Contains(err.Error(), "application/pdf") {
		t.Errorf("refusal = %q, want it to name both what it got and what it needs", err)
	}
}

// AN UNKNOWN SHA NAMES THE CACHE. A seat reaches this by mistyping a hash out of a fetch
// summary, and a bare "not found" leaves it guessing whether the document, the run or the
// hash is wrong.
func TestOCRPagesRefusesAShaTheCacheDoesNotHold(t *testing.T) {
	dir := t.TempDir()
	_, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", "deadbeef", "--run", dir)
	if err == nil {
		t.Fatal("an uncached sha was accepted")
	}
	for _, want := range []string{"deadbeef", "fetch"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
}

// A SECOND RUN AT THE SAME RESOLUTION REUSES THE RENDER and says so — re-rasterising 80 pages
// because a seat ran the verb twice is waste it should be able to see it avoided.
func TestOCRPagesReusesAnExistingRenderAtTheSameResolution(t *testing.T) {
	no := false
	dir, sha := cacheScan(t, &no, "application/pdf")

	if out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "72", "--run", dir); err != nil {
		t.Fatalf("first render: %v\n%s", err, out)
	}
	out, err := run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "72", "--run", dir)
	if err != nil {
		t.Fatalf("second render: %v\n%s", err, out)
	}
	if !strings.Contains(out, "reused: true") {
		t.Errorf("a repeat render at the same dpi did not report reuse:\n%s", out)
	}
	// A DIFFERENT resolution is a different rendering and must not be served from the old one.
	out, err = run(t, "ocr", "pages", "--seat-id", "operator", "--sha", sha, "--dpi", "150", "--run", dir)
	if err != nil {
		t.Fatalf("re-render at a new dpi: %v\n%s", err, out)
	}
	if !strings.Contains(out, "reused: false") || !strings.Contains(out, "dpi: 150") {
		t.Errorf("a render at a new dpi was served from the old one:\n%s", out)
	}
}
