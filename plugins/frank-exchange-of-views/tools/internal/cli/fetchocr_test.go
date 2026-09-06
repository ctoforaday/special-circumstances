package cli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// NOTHING HERE RENDERS A PAGE OR NEEDS THE ENGINE'S C STACK. What is under test at this
// layer is fetch's PLUMBING — when the automatic read fires, what the summary then says,
// and what it says when the read did not happen — and paying four seconds of WebAssembly
// compilation for that would make the test one nobody runs. RenderAndRead's own behaviour
// is tested in internal/fetchcache/readscanned_test.go against the real rasteriser.
type stubScanReader struct {
	rec   fetchcache.ReadingRecord
	err   error
	calls int
}

func (s *stubScanReader) ReadScanned(_ context.Context, _ record.Run, e fetchcache.Entry) (fetchcache.ReadingRecord, error) {
	s.calls++
	if s.err != nil {
		return fetchcache.ReadingRecord{}, s.err
	}
	r := s.rec
	r.Sha = e.Sha
	if r.Engine == "" {
		r.Engine = "fake@test"
	}
	if r.DPI == 0 {
		r.DPI = tessocr.RenderDPI
	}
	return r, nil
}

func withScanReader(t *testing.T, r fetchcache.ScanReader) {
	t.Helper()
	prev := fetchcache.DefaultScanReader
	fetchcache.DefaultScanReader = r
	t.Cleanup(func() { fetchcache.DefaultScanReader = prev })
}

// scannedPDF is what the extractor reports for the case this whole feature exists for: a PDF
// it opened, counted, and found not one glyph of text layer in.
func scannedPDF(pages int) fetchcache.Extraction {
	return fetchcache.Extraction{
		Attempted: true, Pages: pages, ExtractorID: "pdfium@test",
		Reason: "the pdf has no text layer (likely a scan)",
	}
}

func fetchScanned(t *testing.T, sr fetchcache.ScanReader, ex fetchcache.Extraction, args ...string) (string, error) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	withFetcher(t, &fakeFetcher{
		resp:        map[string][]byte{"https://ex/scan.pdf": []byte("%PDF-fake")},
		contentType: "application/pdf",
	})
	withExtractor(t, stubExtractor{out: ex})
	withScanReader(t, sr)
	return run(t, append([]string{"fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/scan.pdf"}, args...)...)
}

// THE ASK, AND THE WHOLE POINT OF THE CHANGE: a seat fetches a scanned standard and gets its
// text back, without knowing that two more verbs exist — read locally, keyed to the engine.
func TestFetchReadsAScannedPDFAutomatically(t *testing.T) {
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		ReadAt: time.Now(), TextSha: "textsha",
		Pages: []fetchcache.PageReading{{Page: 1, TextSha: "p1", Length: 42}},
	}}
	out, err := fetchScanned(t, sr, scannedPDF(1))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if sr.calls != 1 {
		t.Fatalf("the scanned document was read %d times, want once", sr.calls)
	}
	for _, want := range []string{
		"text_extracted: true",
		"ocr_derived: true",
		".ocr.txt",
		"text_sha256: textsha",
		"engine: fake@test",
		"dpi: 300",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary is missing %q. got:\n%s", want, out)
		}
	}
	// THE EXTRACTION'S REASON MUST NOT SURVIVE ALONGSIDE THE TEXT. "no text layer" printed
	// beside a text_path is two answers to one question, and a seat has no way to pick.
	if strings.Contains(out, "text_extracted: false") || strings.Contains(out, "text_reason:") {
		t.Errorf("the summary still says there is no text after reading it:\n%s", out)
	}
	// AND NO MODEL VOCABULARY SURVIVES: the reading is local, and a summary that still
	// spoke of models or tokens would be a carrier of the old concept.
	for _, gone := range []string{"model:", "input_tokens", "output_tokens", "ocr_blocked_pages"} {
		if strings.Contains(out, gone) {
			t.Errorf("summary still carries %q, which left with the API:\n%s", gone, out)
		}
	}
}

// THE FAILURE THAT MUST NOT LOOK LIKE AN UNREADABLE SOURCE. The engine absent from this
// binary, the render budget, a broken rasteriser — each is a fact about THE READING, and
// the document is still a perfectly good cached source. The fetch therefore succeeds and
// the cause is stated.
func TestFetchSurvivesAFailedReadAndSaysWhy(t *testing.T) {
	sr := &stubScanReader{err: errors.New("the OCR engine is not compiled into this binary; release binaries carry it")}
	out, err := fetchScanned(t, sr, scannedPDF(2))
	if err != nil {
		t.Fatalf("a failed reading failed the whole fetch: %v", err)
	}
	if !strings.Contains(out, "ocr_reason: the OCR engine is not compiled into this binary") {
		t.Errorf("the summary does not carry the reason the pages went unread. got:\n%s", out)
	}
	// And it still says the document is cached and where, or the seat has lost the source.
	if !strings.Contains(out, "sha256:") || !strings.Contains(out, "text_extracted: false") {
		t.Errorf("summary lost the document's own facts:\n%s", out)
	}
}

// --ocr=false IS A STATED REASON, NOT A SILENCE. A seat shown only "no text layer" concludes
// the source is unreadable; it must be told that reading it is available and was declined.
func TestFetchOCROffReadsNothingAndNamesTheVerbs(t *testing.T) {
	sr := &stubScanReader{}
	out, err := fetchScanned(t, sr, scannedPDF(2), "--ocr=false")
	if err != nil {
		t.Fatal(err)
	}
	if sr.calls != 0 {
		t.Errorf("--ocr=false still read the document %d times", sr.calls)
	}
	for _, want := range []string{"automatic reading is off", "ocr pages --sha", "ocr read --sha"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary is missing %q. got:\n%s", want, out)
		}
	}
}

// THE GUARD ON THE ENGINE'S SCOPE. The engine is for a scanned standard; it is not for an
// HTML page or a PDF whose text is already on disk, and the three-state pointer is what
// tells them apart.
func TestFetchReadsOnlyAPDFWithNoTextLayer(t *testing.T) {
	for _, tc := range []struct {
		name string
		ex   fetchcache.Extraction
	}{
		{"an html page nothing extracted", fetchcache.Extraction{}},
		{"a pdf that already has text", fetchcache.Extraction{
			Attempted: true, Text: "real text", Pages: 4, ExtractorID: "pdfium@test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sr := &stubScanReader{}
			if _, err := fetchScanned(t, sr, tc.ex); err != nil {
				t.Fatal(err)
			}
			if sr.calls != 0 {
				t.Errorf("the engine was run on %s (%d calls)", tc.name, sr.calls)
			}
		})
	}
}

// THE SUMMARY IS ONE SET OF FACTS IN TWO RENDERINGS, and --json is the one machines read.
func TestFetchJSONCarriesTheReadingsFacts(t *testing.T) {
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		TextSha: "textsha",
		Pages:   []fetchcache.PageReading{{Page: 1, TextSha: "p1", Length: 9}},
	}}
	out, err := fetchScanned(t, sr, scannedPDF(1), "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got fetchSummary
	if jerr := json.Unmarshal([]byte(out), &got); jerr != nil {
		t.Fatalf("fetch --json did not emit valid JSON: %v\n%s", jerr, out)
	}
	if got.TextExtracted == nil || !*got.TextExtracted || !got.OCRDerived {
		t.Errorf("json does not report ocr-derived text: %+v", got)
	}
	if !strings.HasSuffix(got.TextPath, ".ocr.txt") {
		t.Errorf("text_path = %q, want the reading's own file", got.TextPath)
	}
	if got.Engine != "fake@test" {
		t.Errorf("engine = %q, want the reading record's identity key", got.Engine)
	}
	// THE EXTRACTION'S ANSWER MUST NOT SURVIVE THE READING. Both of these are invisible in the
	// human rendering — it prints neither field once there is text — and both would be plainly
	// contradictory to the machine consumer that reads this JSON: a text_reason saying there is
	// no text beside a text_path, and an extractor id naming a library that produced none of it.
	if got.TextReason != "" {
		t.Errorf("text_reason = %q survives beside the text it says does not exist", got.TextReason)
	}
	if got.Extractor != "" {
		t.Errorf("extractor = %q is credited with a reading the OCR engine produced", got.Extractor)
	}
}

// A READING WITH TABLE PAGES ANNOUNCES THEM IN THE SUMMARY: a summary-only reader learns
// that some pages are reconstructed grids — whose confidence stats live on the reading
// record — before opening the text, in both renders.
func TestFetchSummaryCountsTablePages(t *testing.T) {
	recon := &tessocr.Stats{ColumnsFound: 11, RowsFound: 35, MarksTotal: 191, MarksPlaced: 189}
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		ReadAt: time.Now(), TextSha: "textsha",
		Pages: []fetchcache.PageReading{
			{Page: 1, TextSha: "p1", Length: 42},
			{Page: 2, TextSha: "p2", Length: 12, Table: true, GridIntersections: 432, Reconstruction: recon},
			{Page: 3, TextSha: "p3", Length: 40},
		},
	}}
	out, err := fetchScanned(t, sr, scannedPDF(3))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !strings.Contains(out, "table_pages: 1") {
		t.Errorf("summary does not count the table pages:\n%s", out)
	}

	jsonOut, err := fetchScanned(t, &stubScanReader{rec: sr.rec}, scannedPDF(3), "--json")
	if err != nil {
		t.Fatalf("fetch --json: %v", err)
	}
	if !strings.Contains(jsonOut, `"table_pages":1`) {
		t.Errorf("json summary does not carry the count:\n%s", jsonOut)
	}
}

// And the absent case stays absent: zero tables must not print a zero, which would make
// every prose reading's render carry a field that only means something when it doesn't.
func TestFetchSummaryOmitsTablePagesWhenProseOnly(t *testing.T) {
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		ReadAt: time.Now(), TextSha: "textsha",
		Pages: []fetchcache.PageReading{{Page: 1, TextSha: "p1", Length: 42}},
	}}
	out, err := fetchScanned(t, sr, scannedPDF(1))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "table_pages") {
		t.Errorf("a prose-only reading printed a table count:\n%s", out)
	}
}

// --at ALONE WAS A SILENT NO-OP, AND THE DATE IS THE WHOLE QUESTION.
//
// `--at` bounds an ARCHIVE capture and is read on the non-live path only. On a live fetch the
// flag was parsed and dropped: a seat asking what a page said in 2019 got TODAY'S page back with
// a success message and nothing anywhere saying its date had been ignored. That is the same shape
// the `--via` near-miss refusal exists to close — `--via archiv` must not quietly become a live
// fetch — one flag over, and it survived because the wrong answer looks exactly like the right one.
//
// THE SUCCESS CASE IS NOT ASSERTED HERE, deliberately: it needs a real archive backend, which is
// the same reason the release-gate sweep drives this refusal rather than the capture path.
func TestFetchRefusesAnArchiveDateWithNoArchiveBackend(t *testing.T) {
	for _, tc := range []struct{ name, via string }{
		{"no --via at all", ""},
		{"--via live, which reads no captures", "live"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"--ocr=false", "--at", "20190101"}
			if tc.via != "" {
				args = append(args, "--via", tc.via)
			}
			out, err := fetchScanned(t, &stubScanReader{}, fetchcache.Extraction{}, args...)
			if err == nil {
				t.Fatalf("--at was accepted and silently ignored; the fetch reported success:\n%s", out)
			}
			// The refusal has to name the way out, not just the fault: a seat that wanted the
			// 2019 capture needs to be told which flag gets it.
			if !strings.Contains(err.Error(), "--via archive") {
				t.Errorf("the refusal does not say how to actually bound a capture: %v", err)
			}
			if !strings.Contains(err.Error(), "20190101") {
				t.Errorf("the refusal does not quote the date it is declining to use: %v", err)
			}
		})
	}
}

// AND IT STILL FETCHES WITHOUT THE FLAG — the anti-vacuity half. A refusal that fired on every
// live fetch would satisfy the test above and break the verb.
func TestALiveFetchWithNoDateIsUnaffected(t *testing.T) {
	if _, err := fetchScanned(t, &stubScanReader{}, fetchcache.Extraction{}, "--ocr=false", "--via", "live"); err != nil {
		t.Fatalf("an ordinary --via live fetch was refused: %v", err)
	}
}
