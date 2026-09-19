//go:build tessocr

package fetchcache

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/corpus"
)

// THE COMMITTED PAGES, READ BY THE REAL ENGINE THROUGH THE REAL FETCH PATH, PINNED BYTE FOR BYTE.
//
// These are scans that BEAT the reader, so most of these goldens record output that is WRONG. That
// is the point: the golden makes a change visible, and each page's `expect` block — transcribed
// from the page image by a human — says which direction is an improvement. STATUS.md is generated
// from those checks, so the meter cannot go on counting a page that has started to work.
//
// The harness is structurally offline. It serves the bytes it decompressed from page.pdf.gz over
// loopback and hands the reader THAT url; the record's source_url is provenance for a human and is
// never fetched. Nothing in this module refuses an outside host — offline-ness here is the absence
// of any other address, not a promise.
//
// THIS SUITE IS EXPENSIVE AND DOES NOT RIDE THE ORDINARY ONE. Eight scans at 12-24 s each is
// minutes of engine time, and CI pays for it on every leg that compiles it. It runs only where
// FEOV_OCR_CORPUS=1 is set — the `ocr-corpus` job, Linux only, since a Windows minute costs twice a
// Linux one and a macOS minute ten times, and these pages measure the engine, not the platform.
// The env gate is the repository's existing idiom for a costly gate (FEOV_RELEASE_GATE), and it
// keeps the file COMPILED everywhere, so it cannot rot behind a build tag.
//
// Run it, or regenerate (goldens and STATUS.md together):
//
//	FEOV_OCR_CORPUS=1 go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"' \
//	  ./internal/fetchcache/ -run TestCorpusGoldens [-update]
const corpusRoot = "../tessocr/testdata/corpus"

// update regenerates the goldens and STATUS.md together: a golden without the meter beside it is
// half a record of the same run.
var update = flag.Bool("update", false, "rewrite the corpus goldens and STATUS.md from this run")

func TestCorpusGoldens(t *testing.T) {
	if os.Getenv("FEOV_OCR_CORPUS") != "1" && !*update {
		t.Skip("the corpus reads eight scans through the real engine; set FEOV_OCR_CORPUS=1 (the " +
			"ocr-corpus job does) or pass -update")
	}
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var results []corpus.Result
	// The pages run in parallel, and the ceiling is stated rather than assumed: the shipped engine
	// serializes every read on one process-wide mutex (TessocrPageEngine.ReadPage), so what actually
	// overlaps here is the PDFium render, the fetch and the comparison. Wall-clock parallelism across
	// the whole corpus is a CI-level fan-out, not a -parallel flag.
	t.Run("pages", func(t *testing.T) {
		for _, c := range cases {
			c := c
			t.Run(c.Slug, func(t *testing.T) {
				t.Parallel()
				got, observed := readCorpusPage(t, c)
				mu.Lock()
				results = append(results, corpus.Result{Slug: c.Slug, Failed: c.Prov.Expect.Check(observed)})
				mu.Unlock()

				path := filepath.Join(c.Dir, corpus.FileGolden)
				if *update {
					if werr := os.WriteFile(path, []byte(got), 0o644); werr != nil {
						t.Fatal(werr)
					}
					return
				}
				want, rerr := os.ReadFile(path)
				if rerr != nil {
					t.Fatal(rerr)
				}
				if string(want) != got {
					t.Errorf("%s: the reading moved. Read the diff line by line, then regenerate "+
						"with -update.\n--- golden\n%s\n--- now\n%s", c.Slug, want, got)
				}
			})
		}
	})
	if t.Failed() {
		return
	}
	statusPath := filepath.Join(corpusRoot, corpus.FileStatus)
	status := corpus.Status(results)
	if *update {
		if werr := os.WriteFile(statusPath, []byte(status), 0o644); werr != nil {
			t.Fatal(werr)
		}
		return
	}
	have, rerr := os.ReadFile(statusPath)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(have) != status {
		t.Errorf("STATUS.md is stale — a page's expect block and the reading disagree with what it "+
			"records. Regenerate with -update.\n--- STATUS.md\n%s\n--- measured\n%s", have, status)
	}
}

// readCorpusPage runs ONE page through fetch, render and read, exactly as a citation would, and
// returns the golden text and what a human's expect block is checked against.
func readCorpusPage(t *testing.T, c corpus.Case) (string, corpus.Observed) {
	t.Helper()
	gz, err := os.ReadFile(filepath.Join(c.Dir, corpus.FilePage))
	if err != nil {
		t.Fatal(err)
	}
	page, err := corpus.Decompress(gz)
	if err != nil {
		t.Fatal(err)
	}
	if got := corpus.Sha(page); got != c.Prov.PageSha256 {
		t.Fatalf("page.pdf.gz decompresses to %s, the record says %s — the golden below would be "+
			"pinned against bytes nobody can identify", got, c.Prov.PageSha256)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(page)
	}))
	defer srv.Close()

	run := runtest.New(t, t.TempDir())
	url := srv.URL + "/" + c.Slug + ".pdf"
	e, body, _, err := Resolve(run, url, NewHTTPFetcher())
	if err != nil {
		t.Fatalf("fetching the page from loopback: %v", err)
	}
	if !strings.HasPrefix(e.URL, "http://127.0.0.1:") {
		t.Fatalf("the reader was handed %q — this harness serves only what it decompressed", e.URL)
	}
	// THE PAGE IS RENDERED THE WAY PRODUCTION RENDERS IT (#1031): at the scan's own resolution,
	// floored and capped. A fixed 300 here would have the corpus prove a path the reader no longer
	// takes — and these pages are natively 300, 350 and 501.
	native, nerr := NativeDPIs(run, body)
	if nerr != nil {
		t.Fatalf("native resolution: %v", nerr)
	}
	if len(native) != 1 {
		t.Fatalf("a corpus page is one page; this one measured %d", len(native))
	}
	rd, err := RenderPages(run, e.Sha, body, RenderDPIFor(native[0]))
	if err != nil {
		t.Fatalf("rendering: %v", err)
	}
	rec, err := ReadRenderedPages(run, e.Sha, rd)
	if err != nil {
		t.Fatalf("reading: %v", err)
	}
	if len(rec.Pages) != 1 {
		t.Fatalf("a corpus page is one page; this one read %d", len(rec.Pages))
	}
	p := rec.Pages[0]

	text, err := os.ReadFile(PageTextPath(run, e.Sha, 1))
	if err != nil {
		t.Fatal(err)
	}

	observed := corpus.Observed{Table: p.Table, Text: string(text)}
	if p.TextCells != nil {
		observed.Tables, observed.Rows, observed.Columns = p.TextCells.Tables, p.TextCells.Rows, p.TextCells.Columns
	}
	if p.Reconstruction != nil && p.TextCells == nil {
		observed.Rows, observed.Columns = p.Reconstruction.RowsFound, p.Reconstruction.SubColumnsFound
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s · %s · rendered at %d DPI (native %d)\n", c.Slug, DefaultPageEngine.Identity(), firstOf(rd.DPIRange()), native[0])
	fmt.Fprintf(&b, "table: %v · rotated: %v · grid intersections: %d · length: %d\n",
		p.Table, p.RotatedPage, p.GridIntersections, p.Length)
	if p.Reconstruction != nil {
		r := p.Reconstruction
		fmt.Fprintf(&b, "marks: %d columns, %d subcolumns, %d rows, %d of %d placed, psm disagreement %.2f\n",
			r.ColumnsFound, r.SubColumnsFound, r.RowsFound, r.MarksPlaced, r.MarksTotal, r.PSMDisagreement)
	}
	if p.TextCells != nil {
		tc := p.TextCells
		fmt.Fprintf(&b, "text cells: %d tables, %d bands, %d rows, %d columns, %d cells, %d words placed\n",
			tc.Tables, tc.Bands, tc.Rows, tc.Columns, tc.Cells, tc.WordsPlaced)
	}
	if p.ReconstructionFallback != "" {
		fmt.Fprintf(&b, "fell back: %s\n", p.ReconstructionFallback)
	}
	if p.TextCellFallback != "" {
		fmt.Fprintf(&b, "text cells fell back: %s\n", p.TextCellFallback)
	}
	b.WriteString("---\n")
	b.Write(text)
	return b.String(), observed
}

func firstOf(a, _ int) int { return a }
