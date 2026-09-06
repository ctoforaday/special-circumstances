package fetchcache

import (
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
			return &Response{Body: []byte(`{"is_oa":false,"best_oa_location":null}`)}, nil
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
	// The LaTeX source is reachable without the retired MCP server.
	if _, _, e := ArxivURLs("1910.06709"); !strings.Contains(e, "/e-print/") {
		t.Errorf("no e-print url: %s", e)
	}
}
