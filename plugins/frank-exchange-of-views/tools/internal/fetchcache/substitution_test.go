package fetchcache

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE URL'S OWN BYTES SAY SO. A live answer to the asked url carries no substitution and no copy
// url, and its record word is NONE — the statement, not a blank.
func TestALiveAnswerIsTheURLsOwnBytes(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	const u = "https://ex/page"
	f := fake(func(got string) (*Response, error) {
		if got == u {
			return &Response{Body: []byte("<html><body>" + strings.Repeat("a plain page of prose. ", 60) + "</body></html>"), ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: got, Status: 404}
	})
	e, _, _, err := Resolve(run, u, f)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if e.Substitution != "" || e.CopyURL != "" {
		t.Errorf("the url's own answer records a substitution %q at %q", e.Substitution, e.CopyURL)
	}
	if e.CopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE {
		t.Errorf("record word = %v, want NONE", e.CopySubstitution())
	}
}

// A REFUSED URL ANSWERED BY THE RECOVERY CHAIN records the route and the url the backend fetched —
// the arXiv pdf here, the snapshot below — carried from the fetch, never read out of the Via sentence.
func TestARecoveredCopyRecordsTheURLItWasFetchedFrom(t *testing.T) {
	t.Run("arxiv", func(t *testing.T) {
		run := runtest.New(t, t.TempDir())
		const asked = "https://arxiv.org/abs/2101.00001"
		f := fake(func(u string) (*Response, error) {
			if u == "https://arxiv.org/pdf/2101.00001" {
				return &Response{Body: []byte("%PDF-1.7 the preprint"), ContentType: "application/pdf"}, nil
			}
			return nil, &Refusal{URL: u, Status: 403}
		})
		e, _, _, err := Resolve(run, asked, f)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if e.Substitution != SubstitutionRecovered || e.CopyURL != "https://arxiv.org/pdf/2101.00001" {
			t.Errorf("substitution %q copy_url %q; want recovered from the arXiv pdf", e.Substitution, e.CopyURL)
		}
		if e.CopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED {
			t.Errorf("record word = %v, want RECOVERED", e.CopySubstitution())
		}
	})
	t.Run("archive", func(t *testing.T) {
		run := runtest.New(t, t.TempDir())
		cdx := `[["timestamp","original","digest"],["20190520000000","https://ex/a","D1"]]`
		var snapshot string
		f := fake(func(u string) (*Response, error) {
			switch {
			case strings.Contains(u, "cdx/search"):
				return &Response{Body: []byte(cdx), ContentType: "application/json"}, nil
			case strings.Contains(u, "web.archive.org/web/"):
				snapshot = u
				return &Response{Body: []byte("<html><body>" + strings.Repeat("the archived page. ", 60) + "</body></html>"), ContentType: "text/html"}, nil
			}
			if r, ok := emptyIndexAnswer(u); ok {
				return r, nil
			}
			return nil, &Refusal{URL: u, Status: 403}
		})
		e, _, _, err := Resolve(run, "https://ex/a", f)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if snapshot == "" || e.Substitution != SubstitutionRecovered || e.CopyURL != snapshot {
			t.Errorf("substitution %q copy_url %q; want recovered from the snapshot %q", e.Substitution, e.CopyURL, snapshot)
		}
	})
}

// A FULLER COPY THAT REPLACED AN ABSTRACT records the copy's url and the route, which is not a
// refusal's recovery: the asked url answered, and its answer was not the body.
func TestAFullerCopyRecordsTheURLItWasFetchedFrom(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "the paper's own text")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz"),
		locs:   `{"pdf_url":"https://repo.example/open.pdf","is_oa":true,"version":"acceptedVersion"}`,
		bodies: map[string]*Response{"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}}}
	e, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if e.Substitution != SubstitutionFullerCopy || e.CopyURL != "https://repo.example/open.pdf" {
		t.Errorf("substitution %q copy_url %q; want a fuller copy from the open pdf", e.Substitution, e.CopyURL)
	}
	if e.CopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY {
		t.Errorf("record word = %v, want FULLER_COPY", e.CopySubstitution())
	}
}

// A BODY THE TOOL WROTE ITSELF WAS READ FROM NO URL. "The archive did not answer" is a statement
// of what a backend did; naming the archive's url as where it was read would claim a document.
func TestAStatedAnswerCarriesNoCopyURL(t *testing.T) {
	f := fake(func(u string) (*Response, error) { return nil, &Refusal{URL: u, Status: 503} })
	att := Recover(f, "https://ex/a", ViaArchive, "")
	if att == nil || att.TextRetrieved {
		t.Fatalf("wanted the stated archive-outage answer, got %+v", att)
	}
	if att.URL != "" {
		t.Errorf("a body the tool wrote names a url it was read from: %q", att.URL)
	}
}

// EVERY ROUTE HAS A RECORD WORD, and a word nothing knows is UNSPECIFIED, which the write refuses.
func TestEachSubstitutionHasItsRecordWord(t *testing.T) {
	for word, want := range map[string]recordpb.CopySubstitution{
		"":                       recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE,
		SubstitutionFollowedLink: recordpb.CopySubstitution_COPY_SUBSTITUTION_FOLLOWED_LINK,
		SubstitutionRecovered:    recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED,
		SubstitutionFullerCopy:   recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY,
		SubstitutionNamedBackend: recordpb.CopySubstitution_COPY_SUBSTITUTION_NAMED_BACKEND,
		"guessed":                recordpb.CopySubstitution_COPY_SUBSTITUTION_UNSPECIFIED,
	} {
		if got := (Entry{Substitution: word}).CopySubstitution(); got != want {
			t.Errorf("%q → %v, want %v", word, got, want)
		}
	}
}
