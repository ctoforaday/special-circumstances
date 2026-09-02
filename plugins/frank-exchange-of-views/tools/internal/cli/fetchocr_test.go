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
)

// NOTHING HERE RENDERS A PAGE OR CALLS A MODEL. What is under test at this layer is fetch's
// PLUMBING — when the automatic read fires, what the summary then says, and what it says when
// the read did not happen — and paying four seconds of WebAssembly compilation plus a model
// for that would make the test one nobody runs. RenderAndRead's own behaviour is tested in
// internal/fetchcache/readscanned_test.go against the real rasteriser.
type stubScanReader struct {
	rec   fetchcache.ReadingRecord
	err   error
	calls int
}

func (s *stubScanReader) ReadScanned(_ context.Context, _ record.Run, e fetchcache.Entry, model string, dpi int) (fetchcache.ReadingRecord, error) {
	s.calls++
	if s.err != nil {
		return fetchcache.ReadingRecord{}, s.err
	}
	r := s.rec
	r.Sha, r.Model, r.DPI = e.Sha, model, dpi
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
// text back, without knowing that two more verbs exist.
func TestFetchReadsAScannedPDFAutomatically(t *testing.T) {
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		ReadAt: time.Now(), TextSha: "textsha", DPI: 200, InTokens: 12000, OutTok: 400,
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
		"model: " + defaultReadModel,
		"dpi: 200",
		"input_tokens: 12000",
		"output_tokens: 400",
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
}

// THE FAILURE THAT MUST NOT LOOK LIKE AN UNREADABLE SOURCE. Credentials refused, the cap, a
// broken rasteriser — each is a fact about THE READING, and the document is still a perfectly
// good cached source. The fetch therefore succeeds and the cause is stated.
func TestFetchSurvivesAFailedReadAndSaysWhy(t *testing.T) {
	sr := &stubScanReader{err: errors.New("credentials refused (401): set ANTHROPIC_API_KEY")}
	out, err := fetchScanned(t, sr, scannedPDF(2))
	if err != nil {
		t.Fatalf("a failed reading failed the whole fetch: %v", err)
	}
	if !strings.Contains(out, "ocr_reason: credentials refused (401)") {
		t.Errorf("the summary does not carry the reason the pages went unread. got:\n%s", out)
	}
	// And it still says the document is cached and where, or the seat has lost the source.
	if !strings.Contains(out, "sha256:") || !strings.Contains(out, "text_extracted: false") {
		t.Errorf("summary lost the document's own facts:\n%s", out)
	}
}

// --ocr=false IS A STATED REASON, NOT A SILENCE. A seat shown only "no text layer" concludes
// the source is unreadable; it must be told that reading it is available and was declined.
func TestFetchOCROffSpendsNothingAndNamesTheVerbs(t *testing.T) {
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

// THE GUARD ON SPEND. A model is worth a scanned standard; it is not worth an HTML page or a
// PDF whose text is already on disk, and the three-state pointer is what tells them apart.
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
				t.Errorf("a model was spent on %s (%d calls)", tc.name, sr.calls)
			}
		})
	}
}

// THE SUMMARY IS ONE SET OF FACTS IN TWO RENDERINGS, and --json is the one machines read.
func TestFetchJSONCarriesTheReadingsFacts(t *testing.T) {
	sr := &stubScanReader{rec: fetchcache.ReadingRecord{
		TextSha: "textsha", InTokens: 7, OutTok: 3,
		Pages: []fetchcache.PageReading{{Page: 1, TextSha: "p1", Length: 9}},
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
	if got.InTokens != 7 || got.OutTokens != 3 {
		t.Errorf("tokens = %d/%d, want 7/3", got.InTokens, got.OutTokens)
	}
	// THE EXTRACTION'S ANSWER MUST NOT SURVIVE THE READING. Both of these are invisible in the
	// human rendering — it prints neither field once there is text — and both would be plainly
	// contradictory to the machine consumer that reads this JSON: a text_reason saying there is
	// no text beside a text_path, and an extractor id naming a library that produced none of it.
	if got.TextReason != "" {
		t.Errorf("text_reason = %q survives beside the text it says does not exist", got.TextReason)
	}
	if got.Extractor != "" {
		t.Errorf("extractor = %q is credited with a transcription a model produced", got.Extractor)
	}
}
