//go:build tessocr

package fetchcache

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/corpus"
)

// collapsingEngine is the real engine with its recognition thrown away: every fact the pipeline
// measured stands, and the text comes back empty — a read that produced nothing, on pixels that
// carry print.
type collapsingEngine struct{ real *TessocrPageEngine }

func (c collapsingEngine) Identity() string { return c.real.Identity() }

func (c collapsingEngine) ReadPage(png []byte, dpi int) (tessocr.PageResult, error) {
	res, err := c.real.ReadPage(png, dpi)
	res.Text = ""
	return res, err
}

// THE ACCEPTANCE OF #1029, ON THE REAL ENGINE AND THE REAL PAGES. nbs602-blank is SP 602's blank
// p57; nbs602-contents is the corpus's sparsest page of print, the one a collapse is hardest to
// tell from blank. Both write the empty text. Their receipts must still differ in what the READ
// found, and the blank page's must say it is blank: no text, and ink an order of magnitude under
// the sparsest printed page's (measured 0.078% against 1.5%, a ratio of 19).
//
// It reads two pages and does not ride the corpus gate, so it runs wherever the engine is built.
func TestABlankPageAndACollapsedReadAreDifferentReceiptsOnTheRealEngine(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	bySlug := map[string]corpus.Case{}
	for _, c := range cases {
		bySlug[c.Slug] = c
	}

	receipt := func(slug string, eng PageEngine) map[string]any {
		t.Helper()
		c, ok := bySlug[slug]
		if !ok {
			t.Fatalf("the corpus has no page %s", slug)
		}
		withEngine(t, eng)
		run, sha, rd, _ := renderCorpusPage(t, c)
		if _, err := ReadRenderedPages(run, sha, rd); err != nil {
			t.Fatalf("%s: %v", slug, err)
		}
		return readFacts(t, run, sha)
	}
	eng := &TessocrPageEngine{}
	blank := receipt("nbs602-blank", eng)
	collapsed := receipt("nbs602-contents", collapsingEngine{real: eng})

	if blank["length"] != 0.0 || blank["text_sha"] != collapsed["text_sha"] || blank["length"] != collapsed["length"] {
		t.Fatalf("both pages must have written the empty text: blank %v, collapsed %v", blank, collapsed)
	}
	if fmt.Sprint(blank) == fmt.Sprint(collapsed) {
		t.Fatalf("a blank page and a collapsed read wrote the same receipt: %v", blank)
	}
	bi, bok := blank["ink_fraction"].(float64)
	ci, cok := collapsed["ink_fraction"].(float64)
	if !bok || !cok {
		t.Fatalf("a receipt without an ink fraction: blank %v, collapsed %v", blank, collapsed)
	}
	if bi*10 > ci {
		t.Errorf("blank ink %.5f against collapsed %.5f — the blank page no longer measures an order of "+
			"magnitude under the sparsest page of print, so its receipt has stopped saying it is blank", bi, ci)
	}
	t.Logf("blank page: ink %.5f; collapsed read of the contents page: ink %.5f", bi, ci)
}

// readFacts is page 1's receipt without the page's identity — its number, its pixels, its
// resolution and the time — leaving what the READ found.
func readFacts(t *testing.T, run record.Run, sha string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(receiptPath(run, sha, 1))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"page", "render_sha", "dpi", "read_at"} {
		delete(m, k)
	}
	return m
}
