package fetchcache

import (
	"errors"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"strings"
	"testing"
)

type fake func(string) (*Response, error)

// emptyIndexAnswer answers EVERY open-access index with a well-formed "I know of no copy".
//
// IT EXISTS BECAUSE ADDING A SOURCE KEPT BREAKING EVERY FAKE. Each index reports separately
// whether it ANSWERED, and one silence withdraws the determinate "no open copy exists anywhere"
// — correctly. But a fake that does not know about a newly added index refuses it, which reads
// as a silence, which turns unrelated tests red for the right reason in the wrong place. A test
// that means "the indexes all answered, and none had anything" should say that once.
//
// The second return says whether this url was an index lookup at all, so a caller can tell a
// candidate fetch from a directory query.
func isIndexLookup(u string) bool { _, ok := emptyIndexAnswer(u); return ok }

func emptyIndexAnswer(u string) (*Response, bool) {
	switch {
	case strings.Contains(u, "ebi.ac.uk"):
		return &Response{Body: []byte(`{"resultList":{"result":[]}}`)}, true
	case strings.Contains(u, "semanticscholar"):
		return &Response{Body: []byte(`{"paperId":"abc","openAccessPdf":null,"externalIds":{}}`)}, true
	case strings.Contains(u, "doaj.org"):
		return &Response{Body: []byte(`{"total":0,"results":[]}`)}, true
	// A real work record with no registered text-mining link. The DOI is the discriminator the
	// reader checks, so an empty object here would read as a FAILED lookup rather than an
	// answered "this publisher registered nothing".
	case strings.Contains(u, "api.crossref.org"):
		return &Response{Body: []byte(`{"message":{"DOI":"10.1234/x","link":[]}}`)}, true
	case strings.Contains(u, "pmc-oa-opendata"):
		return &Response{Body: []byte(`<ListBucketResult><KeyCount>0</KeyCount></ListBucketResult>`)}, true
	}
	return nil, false
}

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
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"is_oa":false,"oa_url":null},"locations":[]}`)}, nil
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
	// LOWERCASED. A doi's suffix is case-insensitive by the handbook and the indexes are not
	// consistently so: measured 2026-09-24, `10.1016/0021-9991(92)90240-y` resolves at OpenAlex
	// while Semantic Scholar wanted `…-Y`. Normalising asks every index the same question the
	// same way, which is what makes a missing answer mean something.
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
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"https://ex.org/open.pdf","is_oa":true}]}`)}, nil
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
			// EVERY OTHER INDEX ANSWERS CLEANLY, so `answered` turns on OpenAlex's payload alone
			// — which is what this test is about. The flag is a conjunction across the union, and
			// letting a second index be silent here would test the conjunction instead.
			f := fake(func(u string) (*Response, error) {
				if r, isIndex := emptyIndexAnswer(u); isIndex {
					return r, nil
				}
				if strings.Contains(u, "openalex") {
					return &Response{Body: []byte(tc.openalex)}, nil
				}
				return nil, &Refusal{URL: u, Status: 403}
			})
			_, _, _, answered := OpenAccessCandidates(f, "10.1234/x")
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
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"https://ex.org/open.pdf","is_oa":true}]}`)}, nil
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
		case isIndexLookup(u):
			r, _ := emptyIndexAnswer(u)
			return r, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"oa_url":null},
				"locations":[{"pdf_url":"https://walled.example/a.pdf","is_oa":true},
				             {"pdf_url":"https://repo.example/open.pdf","is_oa":true}]}`)}, nil
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
	// The refused candidate is asked of the web archive before the list moves on — that lookup is
	// the fallback for a copy an index asserted and a host would not hand over, so it belongs in
	// the sequence rather than being filtered out of this assertion.
	var docs []string
	for _, u := range tried {
		if !strings.Contains(u, "archive.org") {
			docs = append(docs, u)
		}
	}
	if len(docs) != 2 || !strings.Contains(docs[1], "repo.example") {
		t.Errorf("urls tried = %v, want the refused one then the one that works", tried)
	}
	if len(tried) == len(docs) {
		t.Error("a refused candidate was abandoned without asking the archive for a snapshot of it")
	}
	// THE LIST IS WHAT IS READ, not the index's nomination — `oa_url` here is null while two
	// locations carry a pdf, which is exactly the shape that used to yield "no open copy".
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
		case isIndexLookup(u):
			r, _ := emptyIndexAnswer(u)
			return r, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"https://pub.example/a.pdf","is_oa":true}]}`)}, nil
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

// AND A GENUINE ABSENCE IS STILL A DETERMINATE FINDING — but only when every index ANSWERED.
// Each one that fails to answer takes the world-claim with it, because "no open copy exists
// anywhere" cannot be assembled out of a timeout.
func TestNoLocationsFromEitherIndexIsStillAnAnsweredNo(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
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

// A FLAKY INDEX MUST NOT MANUFACTURE A FACT ABOUT THE WORLD. Europe PMC answers a bare nginx 503
// under load, and losing it loses the PubMed Central route — which is where the copies a
// publisher refuses actually live. Measured: one doi returned a 983 KB PDF on one call and "no
// open copy exists anywhere" on the next.
func TestAnIndexThatDidNotAnswerBlocksTheWorldClaim(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "ebi.ac.uk") {
			return nil, &Refusal{URL: u, Status: 503} // the shape EBI actually fails with
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"oa_url":null},"locations":[]}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	_, _, _, answered := OpenAccessCandidates(f, "10.1234/x")
	if answered {
		t.Error("an index that did not answer was counted as a silence — that licenses " +
			"'no open copy exists anywhere', which would be a claim about the world built from a timeout")
	}
	// AND THE CALLER DECLINES rather than publishing the claim.
	if att := Recover(f, "https://doi.org/10.1234/x", ViaOA, ""); att != nil && strings.Contains(att.Via, "NO OPEN COPY EXISTS") {
		t.Errorf("a 503 from one index became a determinate absence: %s", att.Via)
	}
}

// OPEN ACCESS IS TRIED BEFORE THE ARCHIVE, because one returns the document and the other returns
// a picture of the page you could not read.
//
// Measured on a Journal of Physics paper IOP disallows: the archive answered with a 220 KB
// IOPscience landing page — 17,031 characters of navigation and abstract, classified renderable
// and still not the paper — while open access returns the 633 KB arXiv PDF. Order decided which
// one a seat got.
func TestAutoPrefersTheDocumentOverASnapshotOfIt(t *testing.T) {
	order := AutoOrder()
	pos := map[string]int{}
	for i, b := range order {
		pos[b] = i
	}
	if pos[ViaOA] > pos[ViaArchive] {
		t.Errorf("auto order is %v — the archive is tried before open access, so a snapshot of a "+
			"paywalled landing page outranks the author's own copy", order)
	}
	if pos[ViaArxiv] != 0 {
		t.Errorf("auto order is %v — arxiv is the preprint itself and the cheapest to check", order)
	}
	if pos[ViaMetadata] != len(order)-1 {
		t.Errorf("auto order is %v — metadata is not the document and belongs last", order)
	}

	// AND THE CHAIN ACTUALLY FOLLOWS IT: an index with a copy wins over an archive that has a
	// capture, rather than the other way round.
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "cdx/search"):
			return &Response{Body: []byte(`[["timestamp","original","digest"],["20190520000000","https://ex/a","D1"]]`)}, nil
		case strings.Contains(u, "web.archive.org/web/"):
			return &Response{Body: []byte("<html><body>" + strings.Repeat("landing page. ", 100) + "</body></html>"), ContentType: "text/html"}, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"https://repo.example/open.pdf","is_oa":true}]}`)}, nil
		case u == "https://repo.example/open.pdf":
			return &Response{Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}, nil
		}
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaAuto, "")
	if att == nil {
		t.Fatal("the auto chain found nothing")
	}
	if att.Backend != ViaOA {
		t.Errorf("Backend = %q (via %q), want the open-access copy rather than the archived landing page",
			att.Backend, att.Via)
	}
}

// A RETRACTION REACHES THE RECORD WHICHEVER RUNG ANSWERED. Only the `oa` rung read the work
// record, so whether a seat heard that its paper was withdrawn depended on which backend happened
// to have a copy — and arxiv is FIRST in AutoOrder, so the common case was the silent one.
//
// The assertion runs through EntryFor rather than off the Attempt, because the Attempt carrying
// the fact and the index storing it are two different things and the second one is what a later
// reader sees.
func TestARetractionReachesTheRecordFromANonIndexBackend(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "arxiv.org/pdf"):
			return &Response{Body: []byte("%PDF-1.7 the preprint"), ContentType: "application/pdf"}, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","is_retracted":true,` +
				`"type":"article","open_access":{"oa_status":"green"}}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://arxiv.org/abs/2101.00001?doi=10.1234/retracted", ViaAuto, "")
	if att == nil || att.Backend != ViaArxiv {
		t.Fatalf("wanted the arxiv rung to answer, got %+v", att)
	}
	e := EntryFor(runtest.New(t, t.TempDir()), "https://arxiv.org/abs/2101.00001", att)
	if e.Work == nil || e.Work.Retracted == nil || !*e.Work.Retracted {
		t.Fatalf("the stored record does not say the work is retracted: %+v", e.Work)
	}
	if e.Work.OAStatus != "green" || e.Work.WorkType != "article" {
		t.Errorf("the rest of the work record did not travel with it: %+v", e.Work)
	}
}

// THE PUBLISHED COPY IS PREFERRED OVER A DRAFT OF IT. Both are pdfs and both answer, so nothing
// about the fetch distinguishes them; the index's `version` is the only thing that does, and
// before it was read the chain took whichever the index happened to list first. A quote from a
// submitted preprint attributed to the published paper is a misquotation that no later check
// catches.
func TestThePublishedCopyOutranksASubmittedOne(t *testing.T) {
	var got []string
	f := fake(func(u string) (*Response, error) {
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		if strings.Contains(u, "openalex") {
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[` +
				`{"pdf_url":"https://repo.example/draft.pdf","is_oa":true,"version":"submittedVersion"},` +
				`{"pdf_url":"https://publisher.example/final.pdf","is_oa":true,"version":"publishedVersion","license":"cc-by"}]}`)}, nil
		}
		if strings.HasSuffix(u, ".pdf") {
			got = append(got, u)
			return &Response{Body: []byte("%PDF-1.7 " + u), ContentType: "application/pdf"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaOA, "")
	if att == nil {
		t.Fatal("the oa rung gave up with two open pdfs listed")
	}
	if len(got) == 0 || got[0] != "https://publisher.example/final.pdf" {
		t.Fatalf("the first copy fetched was %v, want the published pdf — the draft was listed first", got)
	}
	if att.Version != "publishedVersion" || att.License != "cc-by" {
		t.Errorf("the copy's own version/licence did not reach the attempt: %q / %q", att.Version, att.License)
	}
	if !strings.Contains(att.Via, "the published version") {
		t.Errorf("the provenance sentence does not say which copy this is: %s", att.Via)
	}
}

// A DOI URL GOES STRAIGHT TO THE REGISTERED PAGE. doi.org publishes no robots.txt and no rate
// headers, so this tool's kind default applies to it — fifteen seconds per fetch, shared across
// processes. Measured on the source sweep, 2,564 of 2,599 urls were doi.org urls: the whole scan
// ran at one redirect service's floor. Crossref holds the same answer in a table.
func TestADOIUrlIsResolvedThroughCrossrefRatherThanTheResolver(t *testing.T) {
	var asked []string
	f := fake(func(u string) (*Response, error) {
		asked = append(asked, u)
		switch {
		case strings.Contains(u, "api.crossref.org"):
			return &Response{Body: []byte(`{"message":{"DOI":"10.1234/x","resource":{"primary":{"URL":"https://publisher.example/article/1"}}}}`)}, nil
		case u == "https://publisher.example/article/1":
			return &Response{Body: []byte("<html><body>the article</body></html>"), ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	if got := RegisteredTarget(f, "https://doi.org/10.1234/x"); got != "https://publisher.example/article/1" {
		t.Fatalf("RegisteredTarget = %q, want the publisher page", got)
	}
	for _, u := range asked {
		if strings.Contains(u, "doi.org") {
			t.Errorf("the resolver was asked anyway: %s", u)
		}
	}

	// A NON-DOI URL IS LEFT ALONE, and so is a doi whose record Crossref does not hold: the
	// caller follows the resolver's own redirect, which is what it did before this existed.
	if got := RegisteredTarget(f, "https://publisher.example/article/1"); got != "" {
		t.Errorf("a plain url was rewritten to %q", got)
	}
	silent := fake(func(u string) (*Response, error) { return &Response{Body: []byte(`{"error":"not found"}`)}, nil })
	if got := RegisteredTarget(silent, "https://doi.org/10.1234/x"); got != "" {
		t.Errorf("an error envelope was read as a target: %q", got)
	}
	// AND A TARGET THAT IS ITSELF A DOI URL IS NO SHORTCUT — it sends the next hop back to the
	// resolver, which is the loop this exists to leave.
	circular := fake(func(u string) (*Response, error) {
		return &Response{Body: []byte(`{"message":{"DOI":"10.1234/x","resource":{"primary":{"URL":"https://dx.doi.org/10.1234/x"}}}}`)}, nil
	})
	if got := RegisteredTarget(circular, "https://doi.org/10.1234/x"); got != "" {
		t.Errorf("a target back on the resolver was accepted: %q", got)
	}
}

// THE PUBLISHER'S OWN REGISTERED ROUTE TO FULL TEXT, and the two limits measured on it.
//
// Crossref's `link[]` carries urls a publisher registered for machine reading. 136 of 300 corpus
// works carry one, and 79 of those are works OpenAlex calls CLOSED — a sanctioned route on papers
// the open-access indexes have nothing for, which is the whole reason this source earns a place
// in the union.
func TestCrossrefTextMiningLinksAreTakenButNotEveryLink(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "api.crossref.org") {
			return &Response{Body: []byte(`{"message":{"DOI":"10.1234/x","link":[
				{"URL":"https://publisher.example/full.xml","content-type":"text/xml","intended-application":"text-mining"},
				{"URL":"https://publisher.example/full.pdf","content-type":"application/pdf","intended-application":"text-mining"},
				{"URL":"https://ithenticate.example/copy.pdf","content-type":"application/pdf","intended-application":"similarity-checking"},
				{"URL":"https://api.elsevier.com/content/article/PII:S1","content-type":"text/plain","intended-application":"text-mining"},
				{"URL":"https://api.wiley.com/onlinelibrary/tdm/v1/articles/10.1234","content-type":"application/pdf","intended-application":"text-mining"}
			]}}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	got, answered := crossrefTextMiningCandidates(f, "10.1234/x")
	if !answered {
		t.Fatal("a well-formed work record was read as an index failure")
	}
	want := []string{"https://publisher.example/full.xml", "https://publisher.example/full.pdf"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want exactly %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("candidate %d = %q, want %q", i, got[i], want[i])
		}
	}

	// A similarity-checking link is NOT a text-mining link. It is registered for plagiarism
	// services under separate agreements, and the intent on the link is the publisher's statement
	// of what it is offering — reading one as an invitation helps ourselves to a different
	// permission from the one that was given.
	for _, u := range got {
		if strings.Contains(u, "ithenticate") {
			t.Error("a similarity-checking link was taken as a text-mining one")
		}
		// And the two key-only hosts are skipped rather than merely allowed to fail. Measured
		// 2026-09-23: api.elsevier.com returned HTTP 400 five times of five, api.wiley.com
		// likewise, and between them they carry 133 of 136 registered links — so trying them
		// costs a request per work to be told no by a host that will always say no.
		if strings.Contains(u, "api.elsevier.com") || strings.Contains(u, "api.wiley.com") {
			t.Errorf("a host that answers only with an api key was tried anyway: %s", u)
		}
	}

	// AN ERROR ENVELOPE IS NOT AN ANSWER. `answered` false withdraws the union's determinate
	// "no open copy exists anywhere" claim, and a body with no DOI cannot support it.
	broken := fake(func(string) (*Response, error) { return &Response{Body: []byte(`{"error":"not found"}`)}, nil })
	if _, ok := crossrefTextMiningCandidates(broken, "10.1234/x"); ok {
		t.Error("an error envelope was read as an answered lookup")
	}
}

// THE INDEX THAT STATES THE TYPE OUTRANKS THE ONES THAT GUESS IT, and the candidate cap is why
// this is not cosmetic.
//
// Europe PMC serves its open pdfs at `/articles/PMC…?pdf=render` and SAYS SO — `documentStyle:
// pdf`, `availability: Free`. That field was discarded and the type re-derived from whether the
// url ends in `.pdf`, which this one does not. It therefore sorted behind every url another index
// merely listed under `pdf_url`, and the four-candidate budget ensured it was never reached.
//
// Measured over 120 works: of the ten where an index said a free copy existed and this tool
// returned none, three were exactly this url, and five wasted a slot on a
// `pubmed.ncbi.nlm.nih.gov` abstract page that OpenAlex had filed as a pdf.
func TestAStatedPDFOutranksAGuessedOne(t *testing.T) {
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "ebi.ac.uk"):
			return &Response{Body: []byte(`{"resultList":{"result":[{"pmcid":"","fullTextUrlList":{"fullTextUrl":[
				{"availability":"Free","documentStyle":"pdf","url":"https://europepmc.org/articles/PMC1?pdf=render"},
				{"availability":"Free","documentStyle":"html","url":"https://europepmc.org/articles/PMC1"}
			]}}]}}`)}, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[
				{"pdf_url":"https://pubmed.ncbi.nlm.nih.gov/9862982","is_oa":true,"version":"publishedVersion"},
				{"pdf_url":"https://www.ncbi.nlm.nih.gov/pmc/articles/148217","is_oa":true,"version":"submittedVersion"}
			]}`)}, nil
		}
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	locs, _, _, _ := OpenAccessCandidates(f, "10.1234/x")
	if len(locs) == 0 {
		t.Fatal("no candidates at all")
	}
	// THE API HOST COMES FIRST, ahead even of a pdf Europe PMC states it holds — because it
	// states that pdf on `europepmc.org`, which is behind a Cloudflare challenge in its entirety,
	// `/robots.txt` included. A stated location on a host that will not open the door loses to a
	// machine route that will.
	if !strings.Contains(locs[0], "ebi.ac.uk") {
		t.Errorf("first candidate is %q, want the api host that actually serves a machine", locs[0])
	}
	var seenRender bool
	for _, l := range locs {
		if strings.Contains(l, "pdf=render") {
			seenRender = true
		}
	}
	if !seenRender {
		t.Errorf("the stated pdf was dropped rather than ranked below a working route: %v", locs)
	}
	// AND THE ABSTRACT PAGES ARE STILL TRIED, as the landing pages they are — a page can name the
	// real pdf in its own citation_pdf_url, so demoting is right and dropping is not.
	joined := strings.Join(locs, " ")
	for _, want := range []string{"pubmed.ncbi.nlm.nih.gov/9862982", "ncbi.nlm.nih.gov/pmc/articles/148217"} {
		if !strings.Contains(joined, want) {
			t.Errorf("a demoted location was dropped rather than ranked lower: %v", locs)
		}
	}
	// THE CONTRACT IS THE ORDER, not a fixed position: every location typed as a pdf comes before
	// every one demoted to a landing page. With a four-candidate budget that is what decides
	// whether the copy that works is ever asked for.
	// The ordering contract, restated for what the union now holds: every route typed as a
	// document comes before every one demoted to a landing page.
	seenPage := false
	for _, l := range locs {
		isDoc := strings.Contains(l, "pdf=render") || strings.Contains(l, "ebi.ac.uk") ||
			strings.Contains(l, "pmc.ncbi.nlm.nih.gov")
		if isDoc && seenPage {
			t.Errorf("a document route sorts behind a landing page: %v", locs)
		}
		if !isDoc {
			seenPage = true
		}
	}
}

// EUROPE PMC'S OWN FULL TEXT, ON THE ONLY HOST THAT WILL HAND IT OVER.
//
// Measured 2026-09-24: `europepmc.org` is behind a Cloudflare managed challenge IN ITS ENTIRETY —
// not the article pages, the whole origin, `/robots.txt` included, which answers "Just a
// moment..." and a challenge script. Every `fullTextUrl` in an Europe PMC record points there,
// including the ones it labels `availability: Free`, and all of them 403.
//
// The REST host does not. `.../rest/<PMCID>/fullTextXML` returns JATS with a `<body>`.
//
// GATED ON isOpenAccess, NOT inEPMC. Three articles with `inEPMC: Y` and `hasPDF: Y` but no
// open-access flag returned HTTP 500 from that route, all three; three with `isOpenAccess: Y`
// returned 70 to 164 KB of text. The two fields answer different questions and only one predicts
// this route.
func TestEuropePMCFullTextComesFromTheAPIHostAndOnlyWhenOpen(t *testing.T) {
	record := func(openAccess string) []byte {
		return []byte(`{"resultList":{"result":[{"pmcid":"PMC148217","isOpenAccess":"` + openAccess + `",` +
			`"inEPMC":"Y","hasPDF":"Y","fullTextUrlList":{"fullTextUrl":[` +
			`{"availability":"Free","documentStyle":"pdf","url":"https://europepmc.org/articles/PMC148217?pdf=render"}]}}]}}`)
	}
	for _, tc := range []struct {
		name, openAccess string
		wantAPIRoute     bool
	}{
		{"open access: the api serves the text", "Y", true},
		{"in epmc but not open: it does not", "N", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fake(func(u string) (*Response, error) {
				if strings.Contains(u, "ebi.ac.uk") {
					return &Response{Body: record(tc.openAccess)}, nil
				}
				if strings.Contains(u, "pmc-oa-opendata") {
					return &Response{Body: []byte(`<ListBucketResult><KeyCount>0</KeyCount></ListBucketResult>`)}, nil
				}
				return nil, &Refusal{URL: u, Status: 403}
			})
			locs, ok := europePMCCandidates(f, "10.1234/x")
			if !ok {
				t.Fatal("a well-formed record read as an index failure")
			}
			const api = "https://www.ebi.ac.uk/europepmc/webservices/rest/PMC148217/fullTextXML"
			var got bool
			for _, l := range locs {
				if l.URL == api {
					got = true
					if !l.IsDocument {
						t.Error("the api full-text route is ranked as a landing page")
					}
				}
			}
			if got != tc.wantAPIRoute {
				t.Errorf("api route present = %v, want %v (locs %+v)", got, tc.wantAPIRoute, locs)
			}
			// THE WALLED URL IS STILL LISTED, deliberately. It is ranked, not denylisted: a host
			// that gates today may not tomorrow, and a hand-kept list of the walled is the shape
			// that rots. What changed is that something that works is tried first.
			var sawWalled bool
			for _, l := range locs {
				if strings.Contains(l.URL, "europepmc.org") {
					sawWalled = true
				}
			}
			if !sawWalled {
				t.Errorf("the advertised url was dropped rather than ranked below a working one: %+v", locs)
			}
		})
	}
}

// A RECOVERED PDF IS READ, NOT JUST STORED.
//
// MEASURED over 120 works: ten pdfs were retrieved, ALL TEN by the open-access chain, and all ten
// were stored with no page count, no text, no extractor id and no reason. Extraction was inline
// on the live path and EntryFor — which every backend answer goes through — never ran it. The
// chain that exists to find a readable copy found ten and read none.
//
// The reason matters as much as the text: OCR keys on an ATTEMPTED extraction that found none, so
// an unextracted pdf is not merely unread, it is invisible to the reader that would have read it.
func TestARecoveredPDFIsExtractedLikeALiveOne(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	att := &Attempt{
		Body:        []byte("%PDF-1.7 not a real document"),
		ContentType: "application/pdf",
		Via:         "open-access copy",
		Backend:     ViaOA,
	}
	e := EntryFor(run, "https://repo.example/paper.pdf", att)
	if e.TextExtracted == nil {
		t.Fatal("the extractor was never asked — 'not attempted' and 'no text found' are different " +
			"states, and this one leaves the document invisible to the OCR path")
	}
	if !*e.TextExtracted && e.TextReason == "" {
		t.Error("an extraction that found nothing recorded no reason, which is the silent zero")
	}
	if e.Extractor == "" {
		t.Error("no extractor id on an attempted extraction: nothing can re-run it")
	}
}

// A COPY AN INDEX ASSERTED, LOST TO A BOT WALL, RECOVERED FROM THE ARCHIVE.
//
// MEASURED 2026-09-24 on doi 10.1016/s0021-9258(19)52451-6 — Lowry et al. 1951, whose pdf carries
// "This is an Open Access article under the CC BY license" on its own first page. Unpaywall names
// the jbc.org pdf; jbc.org answers 403 to this client; the Wayback snapshot of that exact url is
// the same eleven-page Elsevier pdf, 823,098 bytes. An article we are licensed to read, lost to a
// door policy, recovered by asking somebody who kept a copy.
func TestARefusedOpenAccessCopyIsSoughtInTheArchive(t *testing.T) {
	const walled = "https://publisher.example/paper.pdf"
	const snapshot = "https://web.archive.org/web/20250812050548/" + walled
	pdf := "%PDF-1.7 " + strings.Repeat("the paper ", 3000)
	var asked []string
	f := fake(func(u string) (*Response, error) {
		asked = append(asked, u)
		switch {
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"` + walled + `","is_oa":true,"license":"cc-by"}]}`)}, nil
		case strings.Contains(u, "archive.org/wayback/available"):
			return &Response{Body: []byte(`{"archived_snapshots":{"closest":{"available":true,"status":"200",` +
				`"timestamp":"20250812050548","url":"` + snapshot + `"}}}`)}, nil
		case strings.Contains(u, "web.archive.org"):
			// The `if_` modifier asks for the stored bytes rather than the archive's framed replay.
			if !strings.Contains(u, "if_/") {
				t.Errorf("the snapshot was fetched without the raw-bytes modifier: %s", u)
			}
			return &Response{Body: []byte(pdf), ContentType: "application/pdf"}, nil
		case u == walled:
			return nil, &Refusal{URL: u, Status: 403}
		}
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	att := Recover(f, "https://doi.org/10.1234/x", ViaOA, "")
	if att == nil || !att.TextRetrieved {
		t.Fatalf("the archived copy was not taken: %+v", att)
	}
	if string(att.Body) != pdf {
		t.Errorf("wrong bytes: %d", len(att.Body))
	}
	if !strings.Contains(att.Via, "web archive") || !strings.Contains(att.Via, "refused this container") {
		t.Errorf("the provenance does not say where the bytes came from or why: %s", att.Via)
	}

	// A SNAPSHOT OF A REFUSAL IS NOT A COPY. The archive stores whatever it was served, including
	// the 403 page and the challenge interstitial; its own status field is the only thing telling
	// those from the document, and taking one would cache a wall as the paper.
	g := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"` + walled + `","is_oa":true}]}`)}, nil
		case strings.Contains(u, "archive.org/wayback/available"):
			return &Response{Body: []byte(`{"archived_snapshots":{"closest":{"available":true,"status":"403",` +
				`"timestamp":"20250812050548","url":"` + snapshot + `"}}}`)}, nil
		case strings.Contains(u, "web.archive.org"):
			t.Error("a snapshot the archive itself recorded as a 403 was fetched anyway")
			return nil, &Refusal{URL: u, Status: 403}
		}
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	if a := Recover(g, "https://doi.org/10.1234/x", ViaOA, ""); a != nil && a.TextRetrieved {
		t.Error("an archived refusal was served as the document")
	}
}

// A PMC IDENTIFIER IS WORTH MORE THAN THE HOST AN INDEX SPELLS IT WITH.
//
// MEASURED 2026-09-25 across 397 works: of the failures where an IP-indifferent host demonstrably
// holds an open copy, NINETEEN OF TWENTY-SEVEN were this one shape — OpenAlex naming
// `www.ncbi.nlm.nih.gov/pmc/articles/3929010`, the legacy numeric form on the browser host that
// answers a challenge. Demoting that host was right; throwing the identifier away with it was not.
func TestAPMCIdentifierIsExpandedIntoRoutesThatServeAMachine(t *testing.T) {
	for _, shape := range []string{
		"https://www.ncbi.nlm.nih.gov/pmc/articles/3929010", // legacy, numeric, no prefix
		"https://pmc.ncbi.nlm.nih.gov/articles/PMC3929010/", // canonical
		"http://europepmc.org/pmc/articles/PMC3929010",      // Europe PMC's spelling
	} {
		id, routes := PMCRoutes(shape)
		if id != "PMC3929010" {
			t.Errorf("%s -> id %q, want PMC3929010", shape, id)
		}
		if len(routes) == 0 || !strings.Contains(routes[0], "ebi.ac.uk") {
			t.Errorf("%s -> %v, want the api host first", shape, routes)
		}
	}
	// A url that is not PMC at all yields nothing, so nothing else is rewritten by accident.
	if id, _ := PMCRoutes("https://publisher.example/articles/12345"); id != "" {
		t.Errorf("a non-PMC url was read as one: %q", id)
	}

	// AND THE UNION EXPANDS IT. The index names only the walled browser url; the candidates must
	// carry the machine routes, and the bucket is asked because it answers anonymously.
	var asked []string
	f := fake(func(u string) (*Response, error) {
		asked = append(asked, u)
		switch {
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[` +
				`{"pdf_url":"https://www.ncbi.nlm.nih.gov/pmc/articles/3929010","is_oa":true}]}`)}, nil
		case strings.Contains(u, "pmc-oa-opendata"):
			return &Response{Body: []byte(`<ListBucketResult><KeyCount>0</KeyCount></ListBucketResult>`)}, nil
		}
		if r, isIndex := emptyIndexAnswer(u); isIndex {
			return r, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	locs, _, _, _ := OpenAccessCandidates(f, "10.1234/x")
	joined := strings.Join(locs, " ")
	for _, want := range []string{
		"ebi.ac.uk/europepmc/webservices/rest/PMC3929010/fullTextXML",
		"pmc.ncbi.nlm.nih.gov/articles/PMC3929010/",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("the union does not carry %q: %v", want, locs)
		}
	}
	var askedBucket bool
	for _, u := range asked {
		if strings.Contains(u, "pmc-oa-opendata") {
			askedBucket = true
		}
	}
	if !askedBucket {
		t.Error("the open-access bucket was never asked, though it is the route that needs no challenge")
	}
}
