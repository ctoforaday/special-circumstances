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
// Regenerate (goldens and STATUS.md together):
//
//	go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"' \
//	  ./internal/fetchcache/ -run TestCorpusGoldens -update
const corpusRoot = "../tessocr/testdata/corpus"

// update regenerates the goldens and STATUS.md together: a golden without the meter beside it is
// half a record of the same run.
var update = flag.Bool("update", false, "rewrite the corpus goldens and STATUS.md from this run")

func TestCorpusGoldens(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	var results []corpus.Result
	for _, c := range cases {
		c := c
		t.Run(c.Slug, func(t *testing.T) {
			got, observed := readCorpusPage(t, c)
			results = append(results, corpus.Result{Slug: c.Slug, Failed: c.Prov.Expect.Check(observed)})

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
				t.Errorf("%s: the reading moved. Read the diff line by line, then regenerate with "+
					"-update.\n--- golden\n%s\n--- now\n%s", c.Slug, want, got)
			}
		})
	}
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
	rd, err := RenderPages(run, e.Sha, body, DefaultRenderDPI)
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
	fmt.Fprintf(&b, "%s · %s\n", c.Slug, DefaultPageEngine.Identity())
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
