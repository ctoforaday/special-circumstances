package fetchcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// NOTHING HERE NEEDS THE C STACK. DefaultPageEngine is a variable for exactly this reason:
// the read loops' plumbing — receipts, records, staleness, assembly — is pure Go, and only
// internal/tessocr's own suite exercises the real engine. The fake counts its calls,
// because a page derived twice where a receipt should have served it is a real defect even
// now that the second derivation costs seconds instead of dollars: it can replace text a
// seat already cited from.
type fakeEngine struct {
	// perCall returns the result for the nth call (1-based), so a test can make successive
	// pages differ and can prove which call produced which page.
	perCall func(n int) (tessocr.PageResult, error)
	calls   int
	id      string
}

func (f *fakeEngine) Identity() string {
	if f.id == "" {
		return "fake@test"
	}
	return f.id
}

func (f *fakeEngine) ReadPage(_ []byte) (tessocr.PageResult, error) {
	f.calls++
	return f.perCall(f.calls)
}

// textEngine is the common case: a fake whose pages are prose.
func textEngine(perCall func(n int) (string, error)) *fakeEngine {
	return &fakeEngine{perCall: func(n int) (tessocr.PageResult, error) {
		text, err := perCall(n)
		return tessocr.PageResult{Text: text}, err
	}}
}

func withEngine(t *testing.T, e PageEngine) {
	t.Helper()
	prev := DefaultPageEngine
	DefaultPageEngine = e
	t.Cleanup(func() { DefaultPageEngine = prev })
}

// fakeRender writes a render a test controls page by page: arbitrary bytes stand in for
// images (nothing in the read loop decodes them; the receipt key is their hash), and the
// render record binds them exactly as RenderPages would. DPI defaults matter: the read
// path refuses anything but the engine's operative resolution, so these fixtures render
// at it.
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

func onePageRender(t *testing.T) (record.Run, string, RenderRecord) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("page-one")}, tessocr.RenderDPI)
	return run, sha, rd
}

func threePageRender(t *testing.T) (record.Run, string, RenderRecord) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("page-one"), []byte("page-two"), []byte("page-three")}, tessocr.RenderDPI)
	return run, sha, rd
}

// THE RECORD SUPPORTS RE-DERIVATION, and this pins what it must carry to do so: the
// engine identity, the resolution, and the hash of every image read. A field missing here
// is a `reproduce` that cannot run.
func TestAPageIsReadOnceAndTheRecordKeysTheEngine(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "Hello leaf verification\n", nil })
	withEngine(t, fe)
	run, sha, rd := onePageRender(t)

	got, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	if len(got.Pages) != 1 {
		t.Fatalf("Pages = %d, want 1", len(got.Pages))
	}
	if fe.calls != 1 {
		t.Errorf("the engine ran %d times for one page, want 1", fe.calls)
	}

	// THE RE-DERIVATION KEY: engine identity, DPI, and the exact images read.
	if got.Engine != "fake@test" || got.ReadAt.IsZero() || got.DPI != rd.DPI {
		t.Errorf("re-derivation key incomplete: engine=%q readAt=%v dpi=%d", got.Engine, got.ReadAt, got.DPI)
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

	// The page's own reading sits beside its image, and the record's hash is of that text.
	pageText, err := os.ReadFile(PageTextPath(run, sha, 1))
	if err != nil {
		t.Fatalf("page text was not kept: %v — a citation lands on a page, and a reader "+
			"checking it against the scan wants that page's text beside that page's image", err)
	}
	if Sha(pageText) != got.Pages[0].TextSha {
		t.Error("the page's recorded text_sha does not hash the page text on disk")
	}
	if got.Pages[0].Length != len([]rune(string(pageText))) {
		t.Errorf("Length = %d, want the reading's rune count %d",
			got.Pages[0].Length, len([]rune(string(pageText))))
	}
}

// EVERY PAGE IS READ, IN ORDER, AND THE ASSEMBLY IS THE PAGES. A loop that skipped a page
// or reused one page's text for another would leave the record looking complete.
func TestEveryPageIsReadOnceAndAssembledInOrder(t *testing.T) {
	fe := textEngine(func(n int) (string, error) { return fmt.Sprintf("page %d body", n), nil })
	withEngine(t, fe)
	run, sha, rd := threePageRender(t)

	got, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	if fe.calls != 3 || len(got.Pages) != 3 {
		t.Fatalf("%d engine runs for %d recorded pages, want 3 and 3 — one run a page", fe.calls, len(got.Pages))
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

// AN ENGINE THAT FAILED MUST NOT LOOK LIKE A BLANK PAGE. An empty string is a legitimate
// reading — a blank page reads as nothing — so a failure has to arrive as an error or the
// two are indistinguishable. This is also the path the stub build's engine-absent refusal
// travels.
func TestReadingPropagatesAnEngineFailureRatherThanRecordingAnEmptyPage(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) {
		return "", errors.New("engine exploded mid-page")
	}))
	run, sha, rd := onePageRender(t)
	_, err := ReadRenderedPages(run, sha, rd)
	if err == nil {
		t.Fatal("a failing engine produced a reading record")
	}
	if !strings.Contains(err.Error(), "engine exploded") {
		t.Errorf("err = %v, want the engine's own message", err)
	}
	// And nothing was recorded, so a retry is clean.
	if _, ok, _ := ReadReadingRecord(run, sha); ok {
		t.Error("a reading record was written despite the failure")
	}
}

// A BLANK PAGE IS A READING, and must be recorded as one rather than as a failure.
func TestReadingAcceptsAGenuinelyBlankPage(t *testing.T) {
	withEngine(t, textEngine(func(int) (string, error) { return "", nil }))
	run, sha, rd := onePageRender(t)
	got, err := ReadRenderedPages(run, sha, rd)
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
	withEngine(t, textEngine(func(int) (string, error) {
		return "  line one   \r\n\r\nline two\n\n  ", nil
	}))
	run, sha, rd := onePageRender(t)
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
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

// THE ENGINE'S CONSTANTS ARE PER-DPI FACTS, so a render at another resolution is refused
// by name rather than read with thresholds tuned for different pixels — that failure
// would be a quiet misdetection, not an error.
func TestReadingRefusesARenderAtAnotherResolution(t *testing.T) {
	fe := textEngine(func(int) (string, error) { return "x", nil })
	withEngine(t, fe)
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("page-one")}, 200)

	_, err := ReadRenderedPages(run, sha, rd)
	if err == nil {
		t.Fatal("a 200-DPI render was read with 300-DPI constants")
	}
	for _, want := range []string{"tuned", "re-render", "300"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	if fe.calls != 0 {
		t.Errorf("the engine ran %d times over a refused render", fe.calls)
	}
}

// THE BUDGET REFUSES; IT DOES NOT TRUNCATE — and it is denominated in DISK, which is what
// remains bounded now that a page is local compute rather than a model call. Both
// directions of the boundary are pinned, with the arithmetic restated so a change has to
// face it.
func TestRenderBudgetRefusesByStatedArithmetic(t *testing.T) {
	pagesAtBudget := MaxRenderBudgetBytes / approxPageBytesAt300DPI // ~2995 at 1 GB / 350 KB
	if err := renderWithinDiskBudget(pagesAtBudget, tessocr.RenderDPI); err != nil {
		t.Errorf("a document at the budget was refused: %v", err)
	}
	err := renderWithinDiskBudget(pagesAtBudget+1, tessocr.RenderDPI)
	if err == nil {
		t.Fatal("a document over the disk budget was accepted")
	}
	for _, want := range []string{"budget", "nothing was rendered", "KB a page"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	// The estimate scales with the square of the resolution: at 400 DPI the same budget
	// holds (400/300)² ≈ 1.78× fewer pages.
	if err := renderWithinDiskBudget(pagesAtBudget, 400); err == nil {
		t.Error("the 300-DPI page count fit at 400 DPI — the resolution-squared scaling is gone")
	}
	// And the 534-page CJK chart that motivated the OLD cap now fits: it was always a
	// disk question wearing a billing cap.
	if err := renderWithinDiskBudget(534, tessocr.RenderDPI); err != nil {
		t.Errorf("534 pages at 300 DPI (~187 MB) was refused: %v", err)
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

// THE GRID BRANCH'S FACTS ARE FIELDS ON THE PAGE ROW — table, rotation, the detector's
// intersection count, the reconstruction stats, and the stated fallback — because
// confidence inferred from the emitted table's shape is exactly what the plan forbids.
func TestTableAndReconstructionFactsLandOnTheRecord(t *testing.T) {
	recon := &tessocr.Stats{
		ColumnsFound: 11, SubColumnsFound: 11, RowsFound: 35, HeaderNamesFound: 11,
		MarksTotal: 191, MarksPlaced: 189, MarksUnplaced: 2, PSMDisagreement: 0.1,
	}
	fe := &fakeEngine{perCall: func(n int) (tessocr.PageResult, error) {
		switch n {
		case 1:
			return tessocr.PageResult{Text: "prose page"}, nil
		case 2:
			return tessocr.PageResult{
				Text: "| corner | H |\n|---|---|\n| row | X |", Table: true, RotatedPage: true,
				Grid:           tessocr.GridStats{HPix: 20000, VPix: 9000, Intersections: 432},
				Reconstruction: recon,
			}, nil
		default:
			return tessocr.PageResult{
				Text: "fallback prose", Table: true,
				Grid:     tessocr.GridStats{HPix: 16000, VPix: 5000, Intersections: 120},
				Fallback: "grid detected but the TSV held no mark tokens; page kept as plain text",
			}, nil
		}
	}}
	withEngine(t, fe)
	run, sha, rd := threePageRender(t)

	rec, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatal(err)
	}
	if rec.TablePages() != 2 {
		t.Errorf("TablePages = %d, want 2", rec.TablePages())
	}
	p1, p2, p3 := rec.Pages[0], rec.Pages[1], rec.Pages[2]
	if p1.Table || p1.Reconstruction != nil || p1.GridIntersections != 0 {
		t.Errorf("the prose page carries grid facts: %+v", p1)
	}
	if !p2.Table || !p2.RotatedPage || p2.GridIntersections != 432 {
		t.Errorf("page 2 = %+v, want table+rotated with the measured intersections", p2)
	}
	if p2.Reconstruction == nil || p2.Reconstruction.MarksPlaced != 189 {
		t.Fatalf("page 2 reconstruction = %+v, want the engine's stats", p2.Reconstruction)
	}
	if got := p2.Reconstruction.ExpectedIntersections(); got != 432 {
		t.Errorf("ExpectedIntersections = %d, want (35+1)x(11+1) = 432 — the denominator "+
			"readers compare grid_intersections against", got)
	}
	if !p3.Table || p3.ReconstructionFallback == "" || !strings.Contains(p3.ReconstructionFallback, "plain text") {
		t.Errorf("page 3 = %+v, want the fallback STATED on the record", p3)
	}

	// AND THEY SURVIVE A RESUME FROM RECEIPTS ALONE: with the record gone, a re-run
	// reconstructs every page row — grid facts included — without running the engine.
	if err := os.Remove(readingRecordPath(run, sha)); err != nil {
		t.Fatal(err)
	}
	poisoned := &fakeEngine{perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{}, errors.New("must not be asked")
	}}
	withEngine(t, poisoned)
	again, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatal(err)
	}
	if poisoned.calls != 0 {
		t.Errorf("reconstruction from receipts ran the engine %d times, want 0", poisoned.calls)
	}
	q2 := again.Pages[1]
	if !q2.Table || !q2.RotatedPage || q2.GridIntersections != 432 || q2.Reconstruction == nil ||
		q2.Reconstruction.MarksPlaced != 189 {
		t.Errorf("grid facts did not survive the receipt round-trip: %+v", q2)
	}
	if again.Pages[2].ReconstructionFallback == "" {
		t.Error("the stated fallback did not survive the receipt round-trip")
	}
}
