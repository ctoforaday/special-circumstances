package fetchcache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// NOTHING HERE CALLS A MODEL. DefaultPageReader is a variable for exactly this reason: a test
// that needed credentials would be a test nobody can run, and one that spent money on every
// `go test ./...` would be worse. The stub counts its calls, because after #659 the call COUNT
// is itself a load-bearing claim — one page is one call, and a change that quietly reinstates a
// second pass doubles every bill this tool produces.
type stubReader struct {
	// perCall returns the text for the nth call (1-based), so a test can make successive pages
	// differ and can prove which call produced which page.
	perCall func(n int) (string, error)
	calls   int
	in, out int64
}

func (s *stubReader) ReadPage(_ context.Context, _ string, _ []byte) (string, int64, int64, error) {
	s.calls++
	text, err := s.perCall(s.calls)
	return text, s.in, s.out, err
}

func withReader(t *testing.T, r PageReader) {
	t.Helper()
	prev := DefaultPageReader
	DefaultPageReader = r
	t.Cleanup(func() { DefaultPageReader = prev })
}

// renderOnePage puts a rendered single-page document in a run and returns it.
func renderOnePage(t *testing.T) (record.Run, string, RenderRecord) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	body := pdfWithNoTextLayer()
	sha := Sha(body)
	rd, err := RenderPages(run, sha, body, MinRenderDPI)
	if err != nil {
		t.Fatal(err)
	}
	return run, sha, rd
}

// ONE PAGE IS ONE MODEL CALL, and this is the assertion #659 turns on.
//
// The verb read every page twice until the cost of the second pass was weighed against
// evidence nobody had gathered. A regression to two passes would not fail any other test here
// — the text, the record and the attestation would all still be right — it would simply double
// the price of every document silently. So the call count is asserted directly.
func TestAPageCostsExactlyOneModelCall(t *testing.T) {
	sr := &stubReader{in: 1200, out: 40, perCall: func(int) (string, error) {
		return "Hello leaf verification\n", nil
	}}
	withReader(t, sr)
	run, sha, rd := renderOnePage(t)

	got, err := ReadRenderedPages(context.Background(), run, sha, "test-model", rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	if len(got.Pages) != 1 {
		t.Fatalf("Pages = %d, want 1", len(got.Pages))
	}
	if sr.calls != 1 {
		t.Errorf("the model was called %d times for one page, want 1 — a second pass doubles "+
			"the cost of every document and #659 ruled one until an error rate justifies more", sr.calls)
	}
	if got.InTokens != 1200 || got.OutTok != 40 {
		t.Errorf("tokens = %d in / %d out, want one call's worth (1200/40)", got.InTokens, got.OutTok)
	}

	// THE ATTESTATION, which is what replaces reproducibility here. It states provenance —
	// which model read which images when — and NOT accuracy: one reading corroborates nothing.
	if got.Model != "test-model" || got.ReadAt.IsZero() || got.DPI != rd.DPI {
		t.Errorf("attestation incomplete: model=%q readAt=%v dpi=%d", got.Model, got.ReadAt, got.DPI)
	}
	if len(got.RenderShas) != 1 || got.RenderShas[0] != rd.PageShas[0] {
		t.Error("the reading does not name the exact images it read — a re-render would go unnoticed")
	}

	body, err := os.ReadFile(OCRTextPath(run, sha))
	if err != nil {
		t.Fatalf("assembled text: %v", err)
	}
	if !strings.Contains(string(body), "Hello leaf verification") {
		t.Errorf("assembled text = %q", body)
	}
	if Sha(body) != got.TextSha {
		t.Error("text_sha does not hash the text on disk")
	}

	// The page's own transcription sits beside its image, and the record's hash is of that text.
	pageText, err := os.ReadFile(PageTextPath(run, sha, 1))
	if err != nil {
		t.Fatalf("page text was not kept: %v — a citation lands on a page, and a reader "+
			"checking it against the scan wants that page's text beside that page's image", err)
	}
	if Sha(pageText) != got.Pages[0].TextSha {
		t.Error("the page's recorded text_sha does not hash the page text on disk")
	}
	if got.Pages[0].Length != len([]rune(string(pageText))) {
		t.Errorf("Length = %d, want the transcription's rune count %d",
			got.Pages[0].Length, len([]rune(string(pageText))))
	}
}

// EVERY PAGE IS READ, IN ORDER, AND THE ASSEMBLY IS THE PAGES. A loop that skipped a page or
// reused one page's text for another would leave the record looking complete.
func TestEveryPageIsReadOnceAndAssembledInOrder(t *testing.T) {
	sr := &stubReader{perCall: func(n int) (string, error) {
		return fmt.Sprintf("page %d body", n), nil
	}}
	withReader(t, sr)
	run := runtest.New(t, t.TempDir())
	body := multiPagePDFWithNoTextLayer(3)
	sha := Sha(body)
	rd, err := RenderPages(run, sha, body, MinRenderDPI)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	if sr.calls != 3 || len(got.Pages) != 3 {
		t.Fatalf("%d calls for %d recorded pages, want 3 and 3 — one call a page", sr.calls, len(got.Pages))
	}
	text, err := os.ReadFile(OCRTextPath(run, sha))
	if err != nil {
		t.Fatal(err)
	}
	if want := "page 1 body\n\npage 2 body\n\npage 3 body"; string(text) != want {
		t.Errorf("assembled text = %q, want %q", text, want)
	}
	for i := 1; i <= 3; i++ {
		if got.Pages[i-1].Page != i {
			t.Errorf("Pages[%d].Page = %d, want %d", i-1, got.Pages[i-1].Page, i)
		}
	}
}

// A READER THAT FAILED MUST NOT LOOK LIKE A BLANK PAGE. An empty string is a legitimate
// reading — the prompt tells the model to return nothing for a blank page — so a failure has
// to arrive as an error or the two are indistinguishable.
func TestReadingPropagatesAReaderFailureRatherThanRecordingAnEmptyPage(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) {
		return "", errors.New("credentials refused (401)")
	}})
	run, sha, rd := renderOnePage(t)
	_, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err == nil {
		t.Fatal("a failing reader produced a reading record")
	}
	if !strings.Contains(err.Error(), "credentials refused") {
		t.Errorf("err = %v, want the reader's own message", err)
	}
	// And nothing was recorded, so a retry is clean.
	if _, ok, _ := ReadReadingRecord(run, sha); ok {
		t.Error("a reading record was written despite the failure")
	}
}

// A BLANK PAGE IS A READING, and must be recorded as one rather than as a failure.
func TestReadingAcceptsAGenuinelyBlankPage(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "", nil }})
	run, sha, rd := renderOnePage(t)
	got, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatalf("a blank page was treated as a failure: %v", err)
	}
	if got.Pages[0].Length != 0 {
		t.Errorf("Length = %d, want 0 for a blank page", got.Pages[0].Length)
	}
}

// THE STORED TEXT IS TIDIED AT ITS EDGES AND UNTOUCHED IN ITS MIDDLE.
//
// Line endings and trailing whitespace are noise in an artifact about to be cited from.
// Paragraph structure is not noise — it is how the page reads, and flattening it would be
// this normalisation editing the content instead of tidying it.
func TestStoredTextSettlesLineEndingsButKeepsParagraphStructure(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) {
		return "  line one   \r\n\r\nline two\n\n  ", nil
	}})
	run, sha, rd := renderOnePage(t)
	if _, err := ReadRenderedPages(context.Background(), run, sha, "m", rd); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(PageTextPath(run, sha, 1))
	if err != nil {
		t.Fatal(err)
	}
	if want := "line one\n\nline two"; string(got) != want {
		t.Errorf("stored text = %q, want %q — CRLF and trailing space settled, the paragraph "+
			"break between them kept", got, want)
	}
}

// THE CAP REFUSES; IT DOES NOT TRUNCATE. Reading the first N pages and recording it as the
// document's reading would be a false statement nothing downstream could detect.
func TestReadingRefusesADocumentOverTheCapRatherThanReadingPartOfIt(t *testing.T) {
	sr := &stubReader{perCall: func(int) (string, error) { return "x", nil }}
	withReader(t, sr)
	run := runtest.New(t, t.TempDir())
	over := make([]string, MaxReadPages+1)
	for i := range over {
		over[i] = "sha"
	}
	_, err := ReadRenderedPages(context.Background(), run, "s", "m",
		RenderRecord{Sha: "s", DPI: 200, PageShas: over})
	if err == nil {
		t.Fatal("a document over the cap was read")
	}
	for _, want := range []string{"cap", "model call"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	// AND IT REFUSED BEFORE SPENDING ANYTHING. A cap that reads the document and then declines
	// to record it has already paid the bill it exists to prevent.
	if sr.calls != 0 {
		t.Errorf("the model was called %d times for a document the cap refused", sr.calls)
	}
}

// THE CAP IS DENOMINATED IN MODEL CALLS. It was 50 pages when a page cost two calls; one pass
// makes the same 100-call budget 100 pages. This pins the reasoning so that halving the cost
// again, or restoring a second pass, has to face the arithmetic rather than leave a number
// nobody can account for.
func TestThePageCapSpendsTheSameCallBudgetItAlwaysDid(t *testing.T) {
	const budgetInCalls = 100
	if MaxReadPages != budgetInCalls {
		t.Errorf("MaxReadPages = %d, want %d — one model call a page, and the cap is that "+
			"call budget. Change this deliberately, with the budget restated.",
			MaxReadPages, budgetInCalls)
	}
}

func TestReadReadingRecordReportsAbsenceAndRefusesAMisfiledRecord(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if _, ok, err := ReadReadingRecord(run, "never-read"); err != nil || ok {
		t.Errorf("ReadReadingRecord on an unread document = (%v, %v), want (false, nil)", ok, err)
	}
	if err := os.MkdirAll(PagesDir(run, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(readingRecordPath(run, "x"), []byte(`{"sha":"someone-else"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadReadingRecord(run, "x"); err == nil || !strings.Contains(err.Error(), "names document") {
		t.Errorf("a record naming another document was accepted: %v", err)
	}
}

// AN OVER-LIMIT IMAGE IS REFUSED BY NAME, NOT BY THE API. `ocr pages` accepts --dpi up to 400,
// and a letter page of scan-like noise measures 12.69 MB of PNG at that resolution — 16.92 MB
// once base64-encoded, past the API's 10 MB per-image ceiling. Without this the operator gets
// an opaque request-too-large from inside the SDK, with nothing connecting it to the flag that
// caused it.
//
// The rule is a pure function of a length precisely so this test needs no model: asserting it
// through ReadPage would mean handing ReadPage a legal image and letting it call out.
func TestAnImageOverTheAPILimitIsRefusedBeforeTheCall(t *testing.T) {
	err := imageWithinAPILimit(MaxImageBytes + 1)
	if err == nil {
		t.Fatal("an image past the per-image limit was accepted")
	}
	for _, want := range []string{"per-image limit", "--dpi 200", "base64"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	// THE LIMIT ITSELF IS USABLE. A guard that refused its own boundary would be a different
	// bug wearing this one's clothes.
	if err := imageWithinAPILimit(MaxImageBytes); err != nil {
		t.Errorf("an image at exactly the limit was refused: %v", err)
	}
	// And the binary ceiling really does leave room once base64 inflates it by 4/3.
	if MaxImageBytes*4/3 >= 10<<20 {
		t.Errorf("MaxImageBytes=%d base64-encodes to %d, past the API's 10 MB limit",
			MaxImageBytes, MaxImageBytes*4/3)
	}
}
