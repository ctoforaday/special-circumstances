package fetchcache

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN HTML PAGE'S TEXT IS WHAT A READER OF IT GETS. The sentence survives with its inline markup
// and entities resolved; scripts, styles, comments and the head do not; blocks break lines.
func TestAnHTMLPageIsReadAsItsText(t *testing.T) {
	page := []byte(`<!doctype html><html><head><title>Fermat primality test &ndash; Wikipedia</title>
<style>.x{color:red}</style><script>var trap = "SCRIPT TEXT";</script></head>
<body><!-- a comment nobody sees --><h1>Fermat primality test</h1>
<p>Carmichael numbers are <b>substantially</b> rarer than primes,&nbsp;yet there are enough of them.</p>
<p>For 91, 2<sup>90</sup> &equiv; 64 (mod 91).</p><ul><li>one</li><li>two</li></ul>
<table><tr><td>a</td><td>b</td></tr></table><noscript>enable JS</noscript></body></html>`)
	ex := DocumentExtractor{}.Extract("", "text/html; charset=utf-8", page)
	if !ex.Attempted || ex.ExtractorID != htmlExtractorID || ex.Reason != "" {
		t.Fatalf("extraction = %+v, want an attempted html extraction with no reason", ex)
	}
	if ex.Title != "Fermat primality test – Wikipedia" {
		t.Errorf("title = %q", ex.Title)
	}
	for _, want := range []string{
		"Carmichael numbers are substantially rarer than primes, yet there are enough of them.",
		"For 91, 2^90 ≡ 64 (mod 91).",
		"\n- one\n", "\n- two\n", "a b",
	} {
		if !strings.Contains(ex.Text, want) {
			t.Errorf("text lacks %q:\n%s", want, ex.Text)
		}
	}
	for _, gone := range []string{"SCRIPT TEXT", "color:red", "a comment nobody sees", "enable JS", "<", "&nbsp;", "Wikipedia"} {
		if strings.Contains(ex.Text, gone) {
			t.Errorf("text carries %q, which a reader of the page does not see:\n%s", gone, ex.Text)
		}
	}
}

// A PAGE WITH NO TEXT IS SAID TO HAVE NONE, with the reason — never an empty file.
func TestAnHTMLPageWithNoTextSaysSo(t *testing.T) {
	ex := extractHTML([]byte(`<html><body><div id="root"></div><script src="/app.js"></script></body></html>`))
	if !ex.Attempted || ex.Text != "" || !strings.Contains(ex.Reason, "no text outside its markup") {
		t.Errorf("extraction = %+v, want attempted, empty, with its reason", ex)
	}
}

// THE FETCH WRITES THE PAGE'S TEXT BESIDE ITS BYTES, through the same fold a PDF's goes through:
// the entry names the text's hash and extractor, the file is on disk, and the body is untouched.
func TestAFetchedPageHasItsTextBesideIt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	page := []byte(`<html><head><title>T</title></head><body><p>Trial division finds 7.</p></body></html>`)
	entry := EntryFor(run, "https://example.org/trial", &Attempt{ContentType: "text/html", Body: page, TextRetrieved: true}, SubstitutionNamedBackend)
	if entry.TextExtracted == nil || !*entry.TextExtracted || entry.Extractor != htmlExtractorID || entry.TextSha == "" {
		t.Fatalf("entry = %+v, want the page's text extracted and hashed", entry)
	}
	b, err := os.ReadFile(TextPath(run, entry.Sha))
	if err != nil {
		t.Fatal(err)
	}
	if Sha(b) != entry.TextSha || strings.TrimSpace(string(b)) != "Trial division finds 7." {
		t.Errorf("text on disk = %q (sha %s), entry says %s", b, Sha(b), entry.TextSha)
	}
}

// AND ON THE LIVE PATH, which extracts in its own place: a page Resolve fetches has its text too.
func TestALiveFetchedPageHasItsTextBesideIt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	page := "<html><body><p>" + strings.Repeat("Trial division finds 7. ", 60) + "</p></body></html>"
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "example.org") {
			return &Response{Body: []byte(page), ContentType: "text/html; charset=utf-8"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	e, _, _, err := Resolve(run, "https://example.org/trial", f)
	if err != nil {
		t.Fatal(err)
	}
	if e.TextExtracted == nil || !*e.TextExtracted || e.Extractor != htmlExtractorID {
		t.Fatalf("entry = %+v, want the page's text extracted", e)
	}
	if b, err := os.ReadFile(TextPath(run, e.Sha)); err != nil || !strings.HasPrefix(string(b), "Trial division finds 7.") {
		t.Errorf("text on disk = %q (%v)", b, err)
	}
}

// EVERY HAND-READ PAGE OF THE CORPUS YIELDS TEXT, and none of it is markup. The corpus holds real
// publisher, repository and wall pages; a regression that stripped a page to nothing, or left tags
// in it, shows here first.
func TestEveryCorpusPageYieldsTextWithoutMarkup(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "bodies", "*.html.gz"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus pages (%v)", err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(zr)
		if err != nil {
			t.Fatal(err)
		}
		text := htmlText(parseHTML(body))
		if len(text) < 100 {
			t.Errorf("%s: %d characters of text from %d bytes", filepath.Base(f), len(text), len(body))
		}
		for _, tell := range []string{"<div", "<script", "</p>", "<span"} {
			if strings.Contains(text, tell) {
				t.Errorf("%s: text carries markup %q", filepath.Base(f), tell)
			}
		}
	}
}
