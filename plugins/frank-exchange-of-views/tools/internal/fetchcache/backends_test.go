package fetchcache

import (
	"errors"
	"strings"
	"testing"
)

type fake func(string) (*Response, error)

func (f fake) Fetch(u string) (*Response, error) { return f(u) }

// CDX, NOT THE AVAILABILITY API — and the date bound is why. The measured run's load-bearing
// evidence was a capture 146 days BEFORE the preprint; "closest to now" can never produce that.
func TestACaptureIsChosenByDateNotByRecency(t *testing.T) {
	cs := []Capture{
		{Timestamp: "20190520000000", Original: "https://ex.org/p"},
		{Timestamp: "20191013000000", Original: "https://ex.org/p"},
		{Timestamp: "20251103000000", Original: "https://ex.org/p"},
	}
	// A priority question: what did this say BEFORE the preprint appeared?
	got, ok := PickCapture(cs, "20191012")
	if !ok || got.Timestamp != "20190520000000" {
		t.Errorf("bounded pick = %v (%v), want the 2019-05-20 capture", got.Timestamp, ok)
	}
	// Unbounded takes the EARLIEST, because first-visible is the question that needs asking.
	if got, _ := PickCapture(cs, ""); got.Timestamp != "20190520000000" {
		t.Errorf("unbounded pick = %v, want the earliest", got.Timestamp)
	}
	if _, ok := PickCapture(cs, "20180101"); ok {
		t.Error("a bound before every capture returned one anyway")
	}
	if u := cs[0].SnapshotURL(); !strings.Contains(u, "20190520000000id_/") {
		t.Errorf("snapshot url does not request the raw capture: %s", u)
	}
}

// AN ANSWERED "NO OPEN COPY" IS THE POINT, not a failure. Measured: Crossref, OpenAlex and
// Unpaywall all agree 10.5951/MT.82.1.0033 — the Savage record — has no open copy anywhere. That
// is a fact about the WORLD, where the run could only say "unreachable from this container".
func TestAnAnsweredNoOpenCopyIsAFindingNotAMiss(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "unpaywall") {
			return &Response{Body: []byte(`{"doi":"10.5951/MT.82.1.0033","is_oa":false,"best_oa_location":null,"oa_locations":[]}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://doi.org/10.5951/MT.82.1.0033", ViaOA, "")
	if att == nil {
		t.Fatal("an answered 'no open copy exists' returned nothing — indistinguishable from never having asked")
	}
	if att.TextRetrieved {
		t.Error("a no-open-copy answer claims text was retrieved")
	}
	for _, want := range []string{"NO OPEN COPY EXISTS", "not about this container"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("the finding does not distinguish world from container (missing %q): %s", want, att.Via)
		}
	}
}

// A METADATA ANSWER IS A CITATION, AND NOT A READING.
func TestMetadataIsARecordNotAReading(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "api.crossref.org") {
			return &Response{Body: []byte(`{"message":{"title":["Sharing Teaching Ideas"],` +
				`"container-title":["The Mathematics Teacher"],"page":"33-35","issued":{"date-parts":[[1989]]}}}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://doi.org/10.5951/MT.82.1.0033", ViaMetadata, "")
	if att == nil {
		t.Fatal("no metadata answer")
	}
	if att.TextRetrieved {
		t.Fatal("a bibliographic record claims to be the source's text")
	}
	for _, want := range []string{"The Mathematics Teacher", "33-35", "TEXT WAS NOT RETRIEVED"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("the record does not carry %q: %s", want, att.Via)
		}
	}
}

func TestIdentifiersAreLiftedOutOfURLs(t *testing.T) {
	if got := DOIOf("https://doi.org/10.5951/MT.82.1.0033"); got != "10.5951/MT.82.1.0033" {
		t.Errorf("doi = %q", got)
	}
	if got := DOIOf("https://example.org/a-page"); got != "" {
		t.Errorf("a non-article url yielded a doi: %q", got)
	}
	if got := ArxivIDOf("https://arxiv.org/abs/1910.06709v2"); got != "1910.06709v2" {
		t.Errorf("arxiv id = %q", got)
	}
	// THE OLD SCHEME IS THE MAJORITY OF THE CITED LITERATURE, not a legacy corner: 75 of the
	// 100 most-cited hep-th papers on INSPIRE carry one. Each of these resolves at arxiv.org
	// today (checked by hand), including the defunct archive `q-alg` and the subject class in
	// the case arxiv.org's own links use.
	for url, want := range map[string]string{
		"https://arxiv.org/abs/hep-th/9711200":               "hep-th/9711200",
		"https://arxiv.org/pdf/math.AG/0611800":              "math.ag/0611800",
		"https://arxiv.org/abs/cond-mat.stat-mech/0605194v3": "cond-mat.stat-mech/0605194v3",
		"https://arxiv.org/abs/q-alg/9705015":                "q-alg/9705015",
	} {
		if got := ArxivIDOf(url); got != want {
			t.Errorf("ArxivIDOf(%s) = %q, want %q", url, got, want)
		}
	}
	// AND IT STILL SAYS NO. A seven-digit run is an old-scheme id only after an archive name;
	// widening the pattern must not turn any path segment into an identifier the backend then
	// builds a bogus arxiv.org url from.
	for _, none := range []string{
		"https://example.org/files/1234567",
		"https://arxiv.org/abs/hep-th/97112",
		"https://arxiv.org/list/hep-th/9711200",
	} {
		if got := ArxivIDOf(none); got != "" {
			t.Errorf("ArxivIDOf(%s) = %q, want no identifier", none, got)
		}
	}
	// The LaTeX source is reachable without the retired MCP server, and the HTML rendering is
	// the form that rescues a modern paper too big for the fetch cap.
	_, _, e, h := ArxivURLs("1910.06709")
	if !strings.Contains(e, "/e-print/") {
		t.Errorf("no e-print url: %s", e)
	}
	if !strings.Contains(h, "/html/") {
		t.Errorf("no html url: %s", h)
	}
}

// A PDF OVER THE CAP IS NOT AN ABSENT PAPER. Measured on the live service: CLIP's PDF is 6.8 MB,
// over this tool's 5 MiB cap, and its HTML rendering is 874 KB. Before the rescue, a seat asking
// `--via arxiv` for one of the most-cited papers in machine learning read "that backend has no
// answer for this url" — the same sentence a url with no arXiv identifier in it gets.
func TestAnOversizePDFFallsBackToArxivsHTML(t *testing.T) {
	var asked []string
	f := fake(func(u string) (*Response, error) {
		asked = append(asked, u)
		if strings.Contains(u, "/pdf/") {
			return nil, errors.New("fetch: exceeds the 5242880-byte cap")
		}
		return &Response{Body: []byte("<html>the paper</html>"), ContentType: "text/html"}, nil
	})
	att := Recover(f, "https://arxiv.org/abs/2103.00020", ViaArxiv, "")
	if att == nil {
		t.Fatal("no answer for a paper arXiv serves")
	}
	if !att.TextRetrieved || string(att.Body) != "<html>the paper</html>" {
		t.Fatalf("the rescue returned %q, retrieved=%v", att.Body, att.TextRetrieved)
	}
	if att.Backend != ViaArxiv {
		t.Errorf("Backend = %q, want the backend that answered", att.Backend)
	}
	// THE SUBSTITUTION MUST BE VISIBLE. The HTML has no pages, so a citation of it cannot name
	// one, and the seat is told that where it will read it rather than inferring it from a
	// content type.
	for _, want := range []string{"HTML RENDERING", "NO PAGE NUMBERS", "exceeds the 5242880-byte cap"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("the substitution is not stated (missing %q): %s", want, att.Via)
		}
	}
	if len(asked) != 2 || !strings.Contains(asked[1], "/html/") {
		t.Errorf("urls tried = %v, want the pdf then the html", asked)
	}
}

// WHERE NEITHER FORM ANSWERS, SAY SO. arXiv has no HTML for a pre-2007 identifier (measured: 404
// on hep-th/9711200), so an oversize old paper exhausts both forms — and the seat must still hear
// the difference between "this run could not take it" and "there is no such paper".
func TestBothArxivFormsFailingIsStatedNotSilent(t *testing.T) {
	f := fake(func(string) (*Response, error) { return nil, errors.New("boom") })
	att := Recover(f, "https://arxiv.org/abs/hep-th/9711200", ViaArxiv, "")
	if att == nil {
		t.Fatal("a named backend that knows the paper exists returned nothing")
	}
	if att.TextRetrieved {
		t.Error("a stated failure claims text was retrieved")
	}
	for _, want := range []string{"COULD NOT TAKE IT", "NOT about the paper's existence", "e-print"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("the refusal does not separate this fetch from the world (missing %q): %s", want, att.Via)
		}
	}
}

// AND UNDER auto IT DECLINES INSTEAD, so the chain runs on. arxiv is FIRST in AutoOrder: a stated
// failure returned there would end the search at the one backend that could not answer, and the
// seat would never learn that an open copy sits one rung down. A fact about this fetch must not
// foreclose the other routes.
func TestAStatedArxivFailureDoesNotEndTheAutoChain(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "arxiv.org") {
			return nil, errors.New("boom")
		}
		if strings.Contains(u, "unpaywall") {
			return &Response{Body: []byte(`{"doi":"10.4310/ATMP.1998.v2.n2.a1","is_oa":true,"best_oa_location":{"url_for_pdf":"https://ex.org/open.pdf"},"oa_locations":[]}`)}, nil
		}
		if u == "https://ex.org/open.pdf" {
			return &Response{Body: []byte("%PDF-1.7 open copy"), ContentType: "application/pdf"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://arxiv.org/abs/hep-th/9711200?doi=10.4310/ATMP.1998.v2.n2.a1", ViaAuto, "")
	if att == nil {
		t.Fatal("the auto chain gave up")
	}
	if att.Backend != ViaOA {
		t.Fatalf("Backend = %q (via %q), want the chain to have carried on past arxiv to oa", att.Backend, att.Via)
	}
}

// A 200 THAT IS NOT A WORK IS NOT AN ANSWER. OpenAccessURL's second return licenses the caller's
// determinate world-claim — "no open copy exists anywhere the OA indexes know of" — and the
// struct it decodes into carries only `open_access`, which unmarshals from any json object at
// all. Without a discriminator an error envelope, a metering message or a moved response shape
// becomes a published fact about a paper the payload was never about.
func TestAnOpenAccessAnswerRequiresAWork(t *testing.T) {
	for _, tc := range []struct {
		name, openalex string
		wantAnswered   bool
	}{
		{"a work with no open copy", `{"id":"https://openalex.org/W1","open_access":{"is_oa":false,"oa_url":""}}`, true},
		{"an error envelope", `{"error":"Not found","message":"the doi is unknown"}`, false},
		{"a metering message", `{"detail":"daily budget exhausted"}`, false},
		{"an empty object", `{}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fake(func(u string) (*Response, error) {
				if strings.Contains(u, "openalex") {
					return &Response{Body: []byte(tc.openalex)}, nil
				}
				return nil, &Refusal{URL: u, Status: 403} // unpaywall silent, so openalex decides
			})
			_, answered := OpenAccessCandidates(f, "10.1234/x")
			if answered != tc.wantAnswered {
				t.Errorf("answered = %v, want %v — this value is what lets the caller say NO OPEN "+
					"COPY EXISTS anywhere, which is a claim about the world", answered, tc.wantAnswered)
			}
		})
	}
}

// AN ARCHIVE THAT CANNOT ANSWER IS NOT AN ARCHIVE WITH NOTHING IN IT. Measured live on
// 2026-09-22 while the source sweep was running: CDX served `<title>Internet Archive: Temporarily
// Offline</title>` as HTML with a 200. The unmarshal failed, CapturesFor returned no captures and
// no error, and a seat was told the url had never been archived — the outage and the honest zero
// were the same bytes, at exactly the moment an outage is most likely, which is under the load
// that caused it.
func TestAnArchiveOutageIsNotAnEmptyArchive(t *testing.T) {
	offline := []byte(`<html><head><title>Internet Archive: Temporarily Offline</title></head><body>…</body></html>`)
	f := fake(func(string) (*Response, error) {
		return &Response{Body: offline, ContentType: "text/html"}, nil
	})
	caps, err := CapturesFor(f, "https://ex.org/a")
	if err == nil {
		t.Fatal("an archive outage was reported as a clean lookup; a seat cannot tell it from 'never archived'")
	}
	if len(caps) != 0 {
		t.Errorf("captures = %v, want none", caps)
	}
	// NAMED, IT SAYS SO. The seat asked this backend a question and is owed the real answer.
	att := Recover(f, "https://ex.org/a", ViaArchive, "")
	if att == nil {
		t.Fatal("--via archive returned nothing for an archive that answered with an outage page")
	}
	if att.TextRetrieved {
		t.Error("an outage report claims text was retrieved")
	}
	for _, want := range []string{"THE ARCHIVE DID NOT ANSWER", "not be recorded as absent"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("the refusal does not separate silence from absence (missing %q): %s", want, att.Via)
		}
	}
}

// AND UNDER auto IT DECLINES, so one backend's outage cannot end the search before the rungs that
// would have answered.
func TestAnArchiveOutageDoesNotEndTheAutoChain(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "web.archive.org") {
			return &Response{Body: []byte("<html><title>Internet Archive: Temporarily Offline</title></html>"), ContentType: "text/html"}, nil
		}
		if strings.Contains(u, "unpaywall") {
			return &Response{Body: []byte(`{"doi":"10.4310/ATMP.1998.v2.n2.a1","is_oa":true,"best_oa_location":{"url_for_pdf":"https://ex.org/open.pdf"},"oa_locations":[]}`)}, nil
		}
		if u == "https://ex.org/open.pdf" {
			return &Response{Body: []byte("%PDF-1.7 open copy"), ContentType: "application/pdf"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaAuto, "")
	if att == nil {
		t.Fatal("the auto chain gave up when the archive was offline")
	}
	if att.Backend != ViaOA {
		t.Fatalf("Backend = %q (via %q), want the chain to have carried on past the archive outage", att.Backend, att.Via)
	}
}

// A WELL-FORMED EMPTY ANSWER IS STILL THE HONEST ZERO, and must not be dressed up as an outage.
func TestAnEmptyCaptureListIsNotAnError(t *testing.T) {
	f := fake(func(string) (*Response, error) {
		return &Response{Body: []byte(`[["timestamp","original","digest"]]`), ContentType: "application/json"}, nil
	})
	caps, err := CapturesFor(f, "https://ex.org/never-archived")
	if err != nil {
		t.Errorf("a url the archive has simply never captured reported an error: %v", err)
	}
	if len(caps) != 0 {
		t.Errorf("captures = %v, want none", caps)
	}
}

// THE INDEXES NOMINATE ONE LOCATION AND LIST SEVERAL. This backend read only the nomination, and
// measured over the sweep's own failures, 48% of the fetches that ended at a wall or a
// bibliographic record had a fetchable pdf_url in a list nothing touched.
func TestEveryListedOpenAccessLocationIsTried(t *testing.T) {
	var tried []string
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "unpaywall"):
			return &Response{Body: []byte(`{"doi":"10.1234/x","best_oa_location":null,
				"oa_locations":[{"url_for_pdf":"https://walled.example/a.pdf","url":"https://walled.example/a"}]}`)}, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"oa_url":null},
				"locations":[{"pdf_url":"https://repo.example/open.pdf","is_oa":true}]}`)}, nil
		}
		tried = append(tried, u)
		if strings.Contains(u, "walled") {
			return nil, &Refusal{URL: u, Status: 403}
		}
		return &Response{Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}, nil
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaOA, "")
	if att == nil || !att.TextRetrieved {
		t.Fatalf("the second listed location was never reached: %+v", att)
	}
	if len(tried) != 2 || !strings.Contains(tried[1], "repo.example") {
		t.Errorf("urls tried = %v, want the refused one then the one that works", tried)
	}
	// BOTH INDEXES CONTRIBUTE. Unpaywall's null best_oa_location used to short-circuit the whole
	// lookup, so OpenAlex's list was never consulted at all.
	if !strings.Contains(att.Via, "repo.example/open.pdf") {
		t.Errorf("the winning location is not named: %s", att.Via)
	}
}

// A PDF URL AN INDEX PUBLISHES IS NOT A COPY. Ten of fourteen such urls answered 403 from the
// very publisher platform that listed them, so a location counts only once its bytes arrive and
// pass the same wall test every other fetch passes.
func TestALocationThatAnswersWithAWallIsNotAccepted(t *testing.T) {
	wall := []byte(`<html><head><title>Just a moment...</title></head><body>checking</body></html>`)
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "unpaywall"):
			return &Response{Body: []byte(`{"doi":"10.1234/x","best_oa_location":{"url_for_pdf":"https://pub.example/a.pdf"}}`)}, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[]}`)}, nil
		}
		return &Response{Body: wall, ContentType: "text/html"}, nil
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaOA, "")
	if att == nil {
		t.Fatal("a named backend returned nothing")
	}
	if att.TextRetrieved {
		t.Error("a challenge page was accepted as the open-access copy")
	}
	// AND IT IS NOT REPORTED AS "no open copy exists" — the copy exists, we could not take it.
	for _, want := range []string{"COULD NOT TAKE IT", "NOT", "browser can open them"} {
		if !strings.Contains(att.Via, want) {
			t.Errorf("located-but-blocked reads as closed (missing %q): %s", want, att.Via)
		}
	}
	if !strings.Contains(string(att.Body), "pub.example") {
		t.Error("the urls a human could open are not carried")
	}
}

// AND A GENUINE ABSENCE IS STILL A DETERMINATE FINDING — now on two silences rather than one.
func TestNoLocationsFromEitherIndexIsStillAnAnsweredNo(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "unpaywall") {
			return &Response{Body: []byte(`{"doi":"10.1234/x","best_oa_location":null,"oa_locations":[]}`)}, nil
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"oa_url":null},"locations":[]}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaOA, "")
	if att == nil || !strings.Contains(att.Via, "NO OPEN COPY EXISTS") {
		t.Fatalf("a genuine absence stopped being a finding: %+v", att)
	}
}
