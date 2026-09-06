package fetchcache

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A fetcher that refuses the live url and serves the archive.
type walledFetcher struct{ calls []string }

func (w *walledFetcher) Fetch(u string) (*Response, error) {
	w.calls = append(w.calls, u)
	switch {
	case strings.HasPrefix(u, availabilityAPI):
		return &Response{Body: []byte(`{"archived_snapshots":{"closest":{"available":true,` +
			`"url":"http://web.archive.org/web/20190520/https://ex.org/p","timestamp":"20190520000000"}}}`),
			ContentType: "application/json"}, nil
	case strings.Contains(u, "web.archive.org"):
		return &Response{Body: []byte("the paper, as it stood in 2019"), ContentType: "text/html"}, nil
	default:
		return nil, &Refusal{URL: u, Status: 403}
	}
}

// THE REFUSAL IS NOT THE END OF THE ATTEMPT.
//
// Every recovery in research/2026-09-02_quadratic-formula was a seat doing this BY HAND against
// the Wayback CDX — and it produced that run's single most load-bearing piece of evidence, the
// cut-the-knot snapshot of 20 May 2019 that refuted the originality claim. A capability a seat
// must rebuild by hand each time is one most seats will skip, so the tool does it.
func TestARefusedSourceIsRecoveredFromTheArchive(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	f := &walledFetcher{}

	e, body, hit, err := Resolve(run, "https://ex.org/p", f)
	if err != nil {
		t.Fatalf("a source with a live snapshot was reported unreachable: %v", err)
	}
	if hit {
		t.Error("a fresh recovery reported a cache hit")
	}
	if string(body) != "the paper, as it stood in 2019" {
		t.Errorf("the archive bytes did not come back: %q", body)
	}
	// PROVENANCE, OR THE RECOVERY IS A LIE. A snapshot is a different artifact from the live page
	// and a citation must be able to say so.
	if e.RetrievedVia == "" {
		t.Fatal("recovered from the archive and recorded nothing about where the bytes came from")
	}
	for _, want := range []string{"2019-05-20", "web.archive.org"} {
		if !strings.Contains(e.RetrievedVia, want) {
			t.Errorf("the provenance does not name %q: %q", want, e.RetrievedVia)
		}
	}
	// The refusal itself is still recorded: recovering the content does not erase the fact that
	// the live source refused this container.
	if e.HTTPStatus != 403 {
		t.Errorf("the live refusal was lost once the archive answered: status %d", e.HTTPStatus)
	}
}

// NO SNAPSHOT IS AN ORDINARY ANSWER. The original refusal must survive unchanged — dressing a
// miss up as a different failure would hide which of the two actually happened.
func TestNoSnapshotLeavesTheOriginalRefusal(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	_, _, _, err := Resolve(run, "https://ex.org/p", fetcherFunc(func(u string) (*Response, error) {
		if strings.HasPrefix(u, availabilityAPI) {
			return &Response{Body: []byte(`{"archived_snapshots":{}}`)}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	}))
	if err == nil {
		t.Fatal("a source with no snapshot anywhere was reported as fetched")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("the original refusal was replaced rather than kept: %v", err)
	}
}

type fetcherFunc func(string) (*Response, error)

func (f fetcherFunc) Fetch(u string) (*Response, error) { return f(u) }
