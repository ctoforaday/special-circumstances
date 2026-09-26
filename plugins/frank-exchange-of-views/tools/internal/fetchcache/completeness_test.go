package fetchcache

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE CORPUS IS REAL PAGES THIS TOOL RECORDED AS DOCUMENTS, each read by hand before any rule was
// written (2026-09-25). `truth` is that reading — FULL, ABSTRACT, or OTHER — and it is fixed; the
// rules answer to it, never the other way round. The pages are kept whole and gzipped because a
// marker's worth is in NOT appearing on the other pages, and a trimmed page cannot show that.
type bodyLabel struct {
	File  string `json:"file"`
	URL   string `json:"url"`
	DOI   string `json:"doi"`
	Truth string `json:"truth"`
	Read  string `json:"read"`
}

func loadBody(t *testing.T, file string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "bodies", file))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestCompletenessOnHandReadPages holds every rule to the ground truth in two ways. It must never
// OVERCLAIM — `full` on a page read as an abstract is a seat quoting a summary as the study. And
// each page's exact verdict is pinned, so a deleted marker or a widened wall fails by name rather
// than drifting every page to `unverified`, which would pass the overclaim check vacuously.
func TestCompletenessOnHandReadPages(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "bodies", "labels.json"))
	if err != nil {
		t.Fatal(err)
	}
	var labels []bodyLabel
	if err := json.Unmarshal(raw, &labels); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"nature-full.html.gz":                  CompletenessFull,
		"pmc-full.html.gz":                     CompletenessFull,
		"nature-paywalled.html.gz":             CompletenessAbstract,
		"nature-pre1997-pdf-canonical.html.gz": CompletenessAbstract,
		"springer-link-paywalled.html.gz":      CompletenessAbstract,
		"highwire-extract-view.html.gz":        CompletenessAbstract,
		"pmc-scanned-images-only.html.gz":      CompletenessAbstract,
		"springer-book-landing.html.gz":        CompletenessNotTheWork,
		"choice-login-wall.html.gz":            CompletenessNotTheWork,
		// No marker speaks for these, and that is the honest answer: each is kept as the fallback
		// and a fuller copy is looked for. A rule that moves one of them needs a hand-read page
		// from its platform first.
		"springer-link-free-body-only-in-pdf.html.gz":   CompletenessUnverified,
		"sciencedirect-oa-abstract-via-wayback.html.gz": CompletenessUnverified,
		"osti-repository-record.html.gz":                CompletenessUnverified,
		"bepress-repository-record.html.gz":             CompletenessUnverified,
	}
	if len(labels) != len(want) {
		t.Fatalf("labels.json lists %d pages and this test pins %d: a page nobody pins is a page nobody scores", len(labels), len(want))
	}
	allowed := map[string]string{"FULL": CompletenessFull, "ABSTRACT": CompletenessAbstract, "OTHER": CompletenessNotTheWork}
	for _, l := range labels {
		t.Run(l.File, func(t *testing.T) {
			got, reason := Completeness("text/html", loadBody(t, l.File), true)
			if got != CompletenessUnverified && got != allowed[l.Truth] {
				t.Fatalf("OVERCLAIM: %s called %q (%s), and the hand reading says %s: %s", l.URL, got, reason, l.Truth, l.Read)
			}
			if got != want[l.File] {
				t.Errorf("%s: got %q (%s), pinned %q", l.File, got, reason, want[l.File])
			}
			if got != "" && reason == "" {
				t.Errorf("%s: a verdict with no reason is a verdict a reader cannot check", l.File)
			}
		})
	}
}

// The wall corpus's DOCUMENTS are abstract pages that are legitimately documents — a proceedings
// page, arXiv's abstract page. They are not walls, and they are not the paper's body either.
func TestAbstractPagesThatAreNotWallsAreNeverFull(t *testing.T) {
	for _, file := range []string{"neurips.html", "cvf.html", "pmlr.html", "acl.html", "arxiv-control.html"} {
		body, err := os.ReadFile(filepath.Join("testdata", "walls", file))
		if err != nil {
			t.Fatal(err)
		}
		if got, reason := Completeness("text/html", body, true); got == CompletenessFull {
			t.Errorf("%s called full (%s): it is an abstract page", file, reason)
		}
	}
}

// NOT ASKED is its own answer. An ordinary web page reached without a doi, declaring no scholarly
// metadata, has no abstract to mistake for a body; answering `unverified` there would send every
// such page looking for a better copy that does not exist.
func TestCompletenessIsNotAskedOfAnOrdinaryPage(t *testing.T) {
	page := []byte(`<html><head><title>Release notes</title></head><body><p>` +
		string(bytes.Repeat([]byte("words "), 400)) + `</p><form><input type="password"></form></body></html>`)
	if got, _ := Completeness("text/html", page, false); got != "" {
		t.Errorf("an ordinary page got %q; the question does not apply to it", got)
	}
	if got, _ := Completeness("application/pdf", []byte("%PDF-1.7"), true); got != "" {
		t.Errorf("a pdf got %q from the html classifier", got)
	}
	if got, _ := Completeness("text/html", page, true); got != CompletenessNotTheWork {
		t.Errorf("a sign-in page reached for a doi got %q", got)
	}
}

// THE TWO RULES THAT GUARD EACH OTHER, tested apart. On a real paywalled page the wall answers
// first, so neither the reference-list cut nor the order of the checks is ever reached — and
// either could be deleted with the corpus green. Each is exercised here by removing the other
// from the same real bytes.
func TestTheReferenceCutAndTheWallOrderEachHoldAlone(t *testing.T) {
	page := loadBody(t, "nature-paywalled.html.gz")
	// nature.com numbers its supplementary-information section in the body series (Sec10) and
	// places it after the references. Without its wall, that section alone must not make the
	// page read as the paper.
	noWall := bytes.ReplaceAll(page, []byte("preview of subscription content"), []byte("-"))
	noWall = bytes.ReplaceAll(noWall, []byte("Preview of subscription content"), []byte("-"))
	if got, reason := Completeness("text/html", noWall, true); got == CompletenessFull {
		t.Errorf("a section numbered after the reference list was counted as body: %s", reason)
	}
	// And with the reference list gone, so that stray section IS counted, the wall still wins: a
	// page printing the paywall is not the body, whatever else it carries.
	noRefs := bytes.ReplaceAll(page, []byte(`id="Bib1"`), []byte(`id="x"`))
	if got, reason := Completeness("text/html", noRefs, true); got != CompletenessAbstract {
		t.Errorf("a page printing its paywall got %q (%s)", got, reason)
	}
}

// THE CALL SITE, NOT THE FUNCTION: a live fetch of a real paywalled page, through Resolve, must
// land on the record as an abstract whose text was not retrieved — and say why.
func TestResolveRecordsAnAbstractAsAnAbstract(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	page := loadBody(t, "nature-paywalled.html.gz")
	const asked = "https://doi.org/10.1038/nature06964"
	f := fake(func(u string) (*Response, error) {
		if u == asked {
			return &Response{Body: page, ContentType: "text/html; charset=utf-8"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	entry, _, _, err := Resolve(run, asked, f)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Completeness != CompletenessAbstract || entry.CompletenessReason == "" {
		t.Errorf("completeness = %q (%q), want abstract with its reason", entry.Completeness, entry.CompletenessReason)
	}
	if !strings.Contains(entry.TextRetrievedReason, "ABSTRACT") {
		t.Errorf("the record does not say why the text was not retrieved: %q", entry.TextRetrievedReason)
	}
	// And it survives the round trip through the index, which is the only thing that outlives the run.
	got, _, ok, lerr := Lookup(run, asked)
	if lerr != nil || !ok || got.Completeness != CompletenessAbstract {
		t.Errorf("the index lost the verdict: ok=%v completeness=%q err=%v", ok, got.Completeness, lerr)
	}
}

// AND THE RECOVERY PATH, which is where an abstract most often arrives: the archive's snapshot of
// a subscription article is its landing page. The live url is refused, every other rung declines,
// and the archive answers with the real paywalled page.
func TestARecoveredAbstractDoesNotClaimItsTextWasRetrieved(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	page := loadBody(t, "springer-link-paywalled.html.gz")
	const asked = "https://doi.org/10.1023/a:1016212804288"
	cdx := `[["timestamp","original","digest"],["20190520000000","` + asked + `","D1"]]`
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "cdx/search"):
			return &Response{Body: []byte(cdx), ContentType: "application/json"}, nil
		case strings.Contains(u, "web.archive.org/web/"):
			return &Response{Body: page, ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	prev := DefaultExtractor
	DefaultExtractor = fixedExtractor{Extraction{}}
	t.Cleanup(func() { DefaultExtractor = prev })

	entry, _, _, err := Resolve(run, asked, f)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Backend != ViaArchive {
		t.Fatalf("the test did not reach the archive rung (backend %q)", entry.Backend)
	}
	if entry.Completeness != CompletenessAbstract {
		t.Errorf("completeness = %q, want abstract", entry.Completeness)
	}
	if entry.TextRetrieved || !strings.Contains(entry.TextRetrievedReason, "ABSTRACT") {
		t.Errorf("a recovered abstract claims its text: text_retrieved=%v reason=%q", entry.TextRetrieved, entry.TextRetrievedReason)
	}
}

// A DOI IS WHAT MAKES A BARE PAGE A QUESTION. Choice's sign-in page declares no scholarly metadata,
// so only the url it was reached by says a work was asked for — and without that, it is an
// ordinary page the question does not apply to.
func TestTheDOIAskedForMakesABarePageAQuestion(t *testing.T) {
	page := loadBody(t, "choice-login-wall.html.gz")
	viaDOI := Entry{URL: "https://doi.org/10.5860/choice.34-3310", ContentType: "text/html"}
	Classify(&viaDOI, page)
	if viaDOI.Completeness != CompletenessNotTheWork {
		t.Errorf("a sign-in page reached for a doi got %q", viaDOI.Completeness)
	}
	plain := Entry{URL: "https://www.choice360.org/login", ContentType: "text/html"}
	Classify(&plain, page)
	if plain.Completeness != "" {
		t.Errorf("the same page reached without a doi got %q; nothing asked it for a work", plain.Completeness)
	}
}
