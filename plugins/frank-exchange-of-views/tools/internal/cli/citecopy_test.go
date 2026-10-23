package cli

import (
	"encoding/json"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// fetchFunc answers each url from a function, so a test can refuse one url with a status and let
// a backend answer another.
type fetchFunc func(string) (*fetchcache.Response, error)

func (f fetchFunc) Fetch(u string) (*fetchcache.Response, error) { return f(u) }

// evidenceCopy is the one source's copy fields as `show evidence` renders them.
func evidenceCopy(t *testing.T, runDir string) (substitution string, copyURL *string) {
	t.Helper()
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Sources []map[string]json.RawMessage `json:"sources"`
	}
	if err := json.Unmarshal(b, &raw); err != nil || len(raw.Sources) != 1 {
		t.Fatalf("evidence: %v\n%s", err, b)
	}
	if err := json.Unmarshal(raw.Sources[0]["copy_substitution"], &substitution); err != nil {
		t.Fatalf("copy_substitution is not rendered: %v\n%s", err, b)
	}
	if v, ok := raw.Sources[0]["copy_url"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			t.Fatal(err)
		}
		copyURL = &s
	}
	return substitution, copyURL
}

// THE CITATION RECORDS WHOSE BYTES IT READ. For every route by which the cache answers a typed url
// with bytes that are not the url's own, the Cite carries the route and the url the copy was read
// from, and `show evidence` shows both to the lens. The url's own bytes are stated as `none`, with
// no copy url — a statement, not a blank.
func TestACiteRecordsTheCopyItRead(t *testing.T) {
	for _, tc := range []struct {
		substitution string
		copyURL      string
		want         recordpb.CopySubstitution
	}{
		{"", "", recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE},
		{fetchcache.SubstitutionFollowedLink, "https://en.example/full", recordpb.CopySubstitution_COPY_SUBSTITUTION_FOLLOWED_LINK},
		{fetchcache.SubstitutionRecovered, "https://web.archive.org/web/2019id_/" + fermatURL, recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED},
		{fetchcache.SubstitutionFullerCopy, "https://repo.example/open.html", recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY},
		{fetchcache.SubstitutionNamedBackend, "https://arxiv.org/html/2101.00001", recordpb.CopySubstitution_COPY_SUBSTITUTION_NAMED_BACKEND},
	} {
		t.Run(recordpb.Word(tc.want), func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, fermatReport)
			registerBlue(t, runDir)
			// Nothing is fetched: the entry is in the cache, as each route leaves it.
			withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
				t.Fatalf("cite fetched %s past a cached entry", u)
				return nil, nil
			}))
			run := runtest.Open(t, runDir)
			entry := fetchcache.EntryFor(run, fermatURL, &fetchcache.Attempt{Body: []byte(fermatPage), ContentType: "text/html",
				TextRetrieved: true, URL: tc.copyURL}, tc.substitution)
			if _, err := fetchcache.Store(run, entry, []byte(fermatPage)); err != nil {
				t.Fatal(err)
			}
			if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
				t.Fatalf("cite: %v\n%s", err, out)
			}
			c := firstCiteEvent(t, runDir)
			if c.GetCopySubstitution() != tc.want {
				t.Errorf("copy_substitution = %v, want %v", c.GetCopySubstitution(), tc.want)
			}
			if c.GetCopyUrl() != tc.copyURL || (tc.copyURL == "") != (c.CopyUrl == nil) {
				t.Errorf("copy_url = %v, want %q", c.CopyUrl, tc.copyURL)
			}
			if c.GetUrl() != fermatURL {
				t.Errorf("the typed url was not kept: %q", c.GetUrl())
			}
			word, copyURL := evidenceCopy(t, runDir)
			if word != recordpb.Word(tc.want) {
				t.Errorf("evidence copy_substitution = %q, want %q", word, recordpb.Word(tc.want))
			}
			if (copyURL == nil) != (tc.copyURL == "") || (copyURL != nil && *copyURL != tc.copyURL) {
				t.Errorf("evidence copy_url = %v, want %q", copyURL, tc.copyURL)
			}
		})
	}
}

// THE CALL CHAIN, NOT THE STAMP: a typed url that refuses is answered by the recovery chain inside
// cite's own resolve, and the citation names the copy that chain read.
func TestACiteOfARefusedURLRecordsTheRecoveredCopy(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, fermatReport)
	registerBlue(t, runDir)
	const asked = "https://arxiv.org/abs/2101.00001"
	const pdf = "https://arxiv.org/pdf/2101.00001"
	withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
		if u == pdf {
			return &fetchcache.Response{Body: []byte(fermatPage), ContentType: "text/html"}, nil
		}
		return nil, &fetchcache.Refusal{URL: u, Status: 403}
	}))
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", fermatClaim,
		"--url", asked, "--title", "Fermat pseudoprime", "--reason", fermatReason, "--source-span", fermatSpan); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	if c.GetCopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED || c.GetCopyUrl() != pdf {
		t.Errorf("copy_substitution %v copy_url %q; want recovered from %s", c.GetCopySubstitution(), c.GetCopyUrl(), pdf)
	}
}

// A NAMED BACKEND'S ANSWER IS CACHED UNDER THE TYPED URL, so a later cite of that url reads it — and
// says so.
func TestACiteAfterANamedBackendFetchRecordsTheBackendsCopy(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, fermatReport)
	registerBlue(t, runDir)
	const asked = "https://arxiv.org/abs/2101.00001"
	const pdf = "https://arxiv.org/pdf/2101.00001"
	withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
		if u == pdf {
			return &fetchcache.Response{Body: []byte(fermatPage), ContentType: "text/html"}, nil
		}
		return nil, &fetchcache.Refusal{URL: u, Status: 404}
	}))
	if out, err := run(t, "fetch", "--run", runDir, "--seat-id", citeSeat, "--url", asked, "--via", fetchcache.ViaArxiv); err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", fermatClaim,
		"--url", asked, "--title", "Fermat pseudoprime", "--reason", fermatReason, "--source-span", fermatSpan); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	if c.GetCopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_NAMED_BACKEND || c.GetCopyUrl() != pdf {
		t.Errorf("copy_substitution %v copy_url %q; want named_backend from %s", c.GetCopySubstitution(), c.GetCopyUrl(), pdf)
	}
}
