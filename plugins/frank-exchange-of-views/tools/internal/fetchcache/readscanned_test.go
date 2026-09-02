package fetchcache

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
	sr := &stubReader{in: 900, out: 30, perCall: func(int) (string, error) { return "1012-1998\n", nil }}
	withReader(t, sr)
	run, e := storeScanned(t, 1)

	rec, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "test-model", MinRenderDPI)
	if err != nil {
		t.Fatalf("ReadScanned: %v", err)
	}
	if len(rec.Pages) != 1 {
		t.Fatalf("reading = %+v, want one page", rec.Pages)
	}
	if sr.calls != 1 {
		t.Errorf("model called %d times, want 1 — one page, one call", sr.calls)
	}
	if rec.DPI != MinRenderDPI {
		t.Errorf("DPI = %d, want the resolution it was told to render at", rec.DPI)
	}
	body, rerr := os.ReadFile(OCRTextPath(run, e.Sha))
	if rerr != nil || !strings.Contains(string(body), "1012-1998") {
		t.Errorf("assembled text = %q, %v", body, rerr)
	}
}

// A SECOND FETCH OF THE SAME DOCUMENT MUST NOT PAY AGAIN. The reading is not idempotent —
// a re-read returns different bytes — so redoing it would also replace a record a seat may
// already have cited from.
func TestReadScannedReusesAReadingOfTheSameImages(t *testing.T) {
	sr := &stubReader{perCall: func(int) (string, error) { return "same", nil }}
	withReader(t, sr)
	run, e := storeScanned(t, 1)

	first, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", MinRenderDPI)
	if err != nil {
		t.Fatal(err)
	}
	after := sr.calls
	second, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", MinRenderDPI)
	if err != nil {
		t.Fatal(err)
	}
	if sr.calls != after {
		t.Errorf("the model was called %d more times on the second pass; a reading of these exact "+
			"images already existed", sr.calls-after)
	}
	if second.TextSha != first.TextSha || second.ReadAt != first.ReadAt {
		t.Error("the second call returned a different reading rather than the stored one")
	}
}

// THE CAP REFUSES BEFORE ANYTHING IS RENDERED, and that is the difference between this and
// calling the two verbs in sequence. ReadRenderedPages also refuses over the cap — but by
// then a 500-page scan has been rasterised onto a disk allowance this repository has already
// exhausted once.
func TestReadScannedRefusesAnOverCapDocumentWithoutRenderingIt(t *testing.T) {
	sr := &stubReader{perCall: func(int) (string, error) { return "x", nil }}
	withReader(t, sr)
	run, e := storeScanned(t, MaxReadPages+1)

	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", MinRenderDPI)
	if err == nil {
		t.Fatal("an over-cap document was read")
	}
	for _, want := range []string{"cap", "ocr pages", "nothing was rendered"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	if sr.calls != 0 {
		t.Errorf("the model was called %d times on a refused document", sr.calls)
	}
	// AND NO PIXELS WERE WRITTEN. A refusal that still rasterised would spend the disk it
	// exists to protect.
	if _, serr := os.Stat(PagesDir(run, e.Sha)); !os.IsNotExist(serr) {
		t.Errorf("a pages directory exists for a refused document: %v", serr)
	}
}

// A READER FAILURE IS AN ERROR, NOT AN EMPTY READING. fetch turns it into a stated reason;
// what it must never become is a document that looks read and says nothing.
func TestReadScannedPropagatesAReaderFailure(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) {
		return "", errors.New("credentials refused (401)")
	}})
	run, e := storeScanned(t, 1)
	_, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", MinRenderDPI)
	if err == nil || !strings.Contains(err.Error(), "credentials refused") {
		t.Errorf("err = %v, want the reader's own message", err)
	}
}

func TestReadScannedRefusesANonPDF(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "x", nil }})
	run := runtest.New(t, t.TempDir())
	no := false
	e, err := Store(run, Entry{URL: "https://ex/a", ContentType: "text/html", TextExtracted: &no}, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	_, rerr := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", MinRenderDPI)
	if rerr == nil || !strings.Contains(rerr.Error(), "text/html") {
		t.Errorf("err = %v, want a refusal naming the content type", rerr)
	}
}

// A RE-RENDER AT ANOTHER RESOLUTION IS A DIFFERENT READING. The reuse guard compares image
// hashes rather than counts, so 200 DPI pixels are never served a 72 DPI transcription.
func TestReadScannedRereadsWhenTheImagesChanged(t *testing.T) {
	sr := &stubReader{perCall: func(int) (string, error) { return "x", nil }}
	withReader(t, sr)
	run, e := storeScanned(t, 1)

	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", 72); err != nil {
		t.Fatal(err)
	}
	after := sr.calls
	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", 150); err != nil {
		t.Fatal(err)
	}
	if sr.calls == after {
		t.Error("a re-render at a different resolution reused the old reading — the transcription " +
			"would describe pixels nobody read")
	}
}

// A TRANSCRIPTION MUST NOT OUTLIVE THE IMAGES IT DESCRIBES. Found by a mutation pass, not by
// a test: the reading record and the per-page text live inside the pages directory and were
// already wiped on a re-render, while the assembled text sat beside the document and survived
// — a transcription of pixels nobody kept, attesting to image hashes whose record had just
// been deleted. It read exactly like a current reading.
func TestARerenderClearsTheReadingOfTheOldPixels(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "old pixels", nil }})
	run, e := storeScanned(t, 1)
	if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e, "m", 72); err != nil {
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
			t.Errorf("%s survived a re-render: a transcription of pixels nobody kept, with no "+
				"record left saying which model read them (%v)", p, err)
		}
	}
	if _, had, err := ReadReadingRecord(run, e.Sha); err != nil || had {
		t.Errorf("ReadReadingRecord after a re-render = (%v, %v), want an honest absence", had, err)
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
