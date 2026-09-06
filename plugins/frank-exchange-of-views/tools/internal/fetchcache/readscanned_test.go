package fetchcache

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// storeScanned puts a text-less PDF in the cache and returns the entry a fetch would have
// written for it: a PDF the extractor looked at and found nothing in.
func storeScanned(t *testing.T, pages int) (record.Run, Entry) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	no := false
	e, err := Store(run, Entry{
		URL: "https://ex/scan.pdf", ContentType: "application/pdf",
		TextExtracted: &no, TextReason: "no text layer", Pages: pages,
	}, pdfWithNoTextLayer())
	if err != nil {
		t.Fatal(err)
	}
	return run, e
}

// THE WHOLE POINT OF THE COMPOSITION: one call, from a cached scan to a reading. Before it,
// a seat that fetched IEEE 1012 got a reason and had to already know two more verbs existed.
func TestReadScannedRendersAndReadsInOneCall(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "1012-1998\n", nil })
	withEngine(t, fe)
	run, e := storeScanned(t, 1)

	rec, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatalf("ReadScanned: %v", err)
	}
	if len(rec.Pages) != 1 {
		t.Fatalf("reading = %+v, want one page", rec.Pages)
	}
	if fe.calls != 1 {
		t.Errorf("engine ran %d times, want 1 — one page, one run", fe.calls)
	}
	if rec.DPI != tessocr.RenderDPI {
		t.Errorf("DPI = %d, want the engine's operative resolution %d", rec.DPI, tessocr.RenderDPI)
	}
	if rec.Engine != "fake@test" {
		t.Errorf("Engine = %q, want the identity of what read it", rec.Engine)
	}
	body, rerr := os.ReadFile(OCRTextPath(run, e.Sha))
	if rerr != nil || !strings.Contains(string(body), "1012-1998") {
		t.Errorf("assembled text = %q, %v", body, rerr)
	}
}

// A SECOND FETCH OF THE SAME DOCUMENT MUST NOT DERIVE AGAIN. Not for cost any more — for
// stability: replacing a record a seat may already have cited from is not a side effect a
// cache hit should have.
func TestReadScannedReusesAReadingOfTheSameImages(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "same", nil })
	withEngine(t, fe)
	run, e := storeScanned(t, 1)

	first, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatal(err)
	}
	after := fe.calls
	second, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatal(err)
	}
	if fe.calls != after {
		t.Errorf("the engine ran %d more times on the second pass; a reading of these exact "+
			"images already existed", fe.calls-after)
	}
	if second.TextSha != first.TextSha || second.ReadAt != first.ReadAt {
		t.Error("the second call returned a different reading rather than the stored one")
	}
}

// A STORED READING BY ANOTHER ENGINE IS RE-DERIVED, not served: the record's engine key is
// half of the re-derivation identity, and pins that bumped mean bytes that may differ.
func TestReadScannedRederivesWhenTheEngineChanged(t *testing.T) {
	old := &fakeEngine{id: "fake@old", perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: "old"}, nil
	}}
	withEngine(t, old)
	run, e := storeScanned(t, 1)
	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e); err != nil {
		t.Fatal(err)
	}

	renewed := &fakeEngine{id: "fake@new", perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: "new"}, nil
	}}
	withEngine(t, renewed)
	rec, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatal(err)
	}
	if renewed.calls == 0 {
		t.Error("a reading by another engine was served as this engine's")
	}
	if rec.Engine != "fake@new" {
		t.Errorf("Engine = %q, want the identity that derived it", rec.Engine)
	}
}

// THE DISK BUDGET REFUSES BEFORE ANYTHING IS RENDERED, from the page count the extractor
// already measured — a document that would blow the budget never costs the first raster.
func TestReadScannedRefusesAnOverBudgetDocumentWithoutRenderingIt(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "x", nil })
	withEngine(t, fe)
	run, e := storeScanned(t, MaxRenderBudgetBytes/approxPageBytesAt300DPI+1)

	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err == nil {
		t.Fatal("an over-budget document was read")
	}
	for _, want := range []string{"budget", "nothing was rendered"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	if fe.calls != 0 {
		t.Errorf("the engine ran %d times on a refused document", fe.calls)
	}
	// AND NO PIXELS WERE WRITTEN. A refusal that still rasterised would spend the disk it
	// exists to protect.
	if _, serr := os.Stat(PagesDir(run, e.Sha)); !os.IsNotExist(serr) {
		t.Errorf("a pages directory exists for a refused document: %v", serr)
	}
}

// AN ENGINE FAILURE IS AN ERROR, NOT AN EMPTY READING. fetch turns it into a stated
// reason; what it must never become is a document that looks read and says nothing.
func TestReadScannedPropagatesAnEngineFailure(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) {
		return "", errors.New("engine exploded mid-page")
	}))
	run, e := storeScanned(t, 1)
	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err == nil || !strings.Contains(err.Error(), "engine exploded") {
		t.Errorf("err = %v, want the engine's own message", err)
	}
}

func TestReadScannedRefusesANonPDF(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) { return "x", nil }))
	run := runtest.New(t, t.TempDir())
	no := false
	e, err := Store(run, Entry{URL: "https://ex/a", ContentType: "text/html", TextExtracted: &no}, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	_, rerr := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if rerr == nil || !strings.Contains(rerr.Error(), "text/html") {
		t.Errorf("err = %v, want a refusal naming the content type", rerr)
	}
}

// A READING MUST NOT OUTLIVE THE IMAGES IT DESCRIBES. Found by a mutation pass, not by a
// test: the reading record and the per-page text live inside the pages directory and were
// already wiped on a re-render, while the assembled text sat beside the document and
// survived — a reading of pixels nobody kept, attesting to image hashes whose record had
// just been deleted. It read exactly like a current reading.
func TestARerenderClearsTheReadingOfTheOldPixels(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) { return "old pixels", nil }))
	run, e := storeScanned(t, 1)
	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{OCRTextPath(run, e.Sha), PageTextPath(run, e.Sha, 1)} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("the reading did not write %s: %v", p, err)
		}
	}

	if _, err := RenderPages(run, e.Sha, pdfWithNoTextLayer(), 150); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{OCRTextPath(run, e.Sha), PageTextPath(run, e.Sha, 1)} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s survived a re-render: a reading of pixels nobody kept, with no "+
				"record left saying which engine read them (%v)", p, err)
		}
	}
	if _, had, err := ReadReadingRecord(run, e.Sha); err != nil || had {
		t.Errorf("ReadReadingRecord after a re-render = (%v, %v), want an honest absence", had, err)
	}
}

// THE AUTOMATIC PATH KEEPS THE TEXT, NOT THE PIXELS. Each page is rasterised, read, and
// released (#671); what lands on disk is the reading, its per-page texts, and a record
// whose RenderShas still name exactly which images were read. No page PNG and no render
// record — an operator who wants images has `ocr pages`.
func TestReadScannedPersistsNoImages(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) { return "text", nil }))
	run, e := storeScanned(t, 1)

	rec, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.RenderShas) != 1 || rec.RenderShas[0] == "" {
		t.Errorf("RenderShas = %v, want one hash per page read — the images' only identity", rec.RenderShas)
	}
	for _, p := range []string{PagePath(run, e.Sha, 1), renderRecordPath(run, e.Sha)} {
		if _, serr := os.Stat(p); !os.IsNotExist(serr) {
			t.Errorf("%s exists after an automatic read; the fetch path keeps text, never pixels (%v)", p, serr)
		}
	}
	for _, p := range []string{PageTextPath(run, e.Sha, 1), OCRTextPath(run, e.Sha)} {
		if _, serr := os.Stat(p); serr != nil {
			t.Errorf("the reading did not persist %s: %v", p, serr)
		}
	}
}

// A MALFORMED READING RECORD IS AN ERROR, NOT AN ABSENCE. Treated as absence, the path
// would silently replace a record a seat may already have cited from.
func TestReadScannedRefusesACorruptReadingRecord(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "x", nil })
	withEngine(t, fe)
	run, e := storeScanned(t, 1)
	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e); err != nil {
		t.Fatal(err)
	}
	calls := fe.calls
	if err := os.WriteFile(readingRecordPath(run, e.Sha), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e)
	if err == nil || !strings.Contains(err.Error(), "unreadable") {
		t.Errorf("err = %v, want the record named as unreadable", err)
	}
	if fe.calls != calls {
		t.Errorf("the engine ran %d more times over a corrupt record", fe.calls-calls)
	}
}

func TestSameRendersComparesOrderAndLength(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b []string
		want bool
	}{
		{"identical", []string{"a", "b"}, []string{"a", "b"}, true},
		{"reordered", []string{"a", "b"}, []string{"b", "a"}, false},
		{"a page added", []string{"a"}, []string{"a", "b"}, false},
		{"both empty", nil, nil, true},
	} {
		if got := SameRenders(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: SameRenders(%v,%v) = %v, want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}
