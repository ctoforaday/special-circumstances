package fetchcache

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// clockAt pins the record clock the fetch reads to stamp.
func clockAt(t *testing.T, stamp string) {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		t.Fatal(err)
	}
	was := record.Now
	record.Now = func() time.Time { return ts }
	t.Cleanup(func() { record.Now = was })
}

// A SOURCE WAS ACCESSED WHEN ITS COPY WAS FETCHED, and a later read served from the cache took no copy:
// the entry keeps the first fetch's time.
func TestTheFetchTimeIsWhenTheCopyArrivedAndAHitKeepsIt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	f := &stubFetcher{resp: map[string][]byte{"https://ex/1": []byte("the source")}}
	clockAt(t, "2030-01-01T23:59:00Z")
	first, _, _, err := Resolve(run, "https://ex/1", f)
	if err != nil {
		t.Fatal(err)
	}
	if first.FetchedAt != "2030-01-01T23:59:00Z" {
		t.Errorf("the fetch stamped %q, want the clock at the fetch, 2030-01-01T23:59:00Z", first.FetchedAt)
	}
	clockAt(t, "2030-01-02T10:00:00Z")
	again, _, hit, err := Resolve(run, "https://ex/1", f)
	if err != nil || !hit {
		t.Fatalf("second Resolve: hit %v, %v", hit, err)
	}
	if again.FetchedAt != first.FetchedAt {
		t.Errorf("a cache hit moved the fetch time from %q to %q", first.FetchedAt, again.FetchedAt)
	}
}

// A RECOVERED COPY WAS FETCHED WHEN ITS BACKEND ANSWERED: Recover stamps the attempt, and the entry
// built from it carries that time.
func TestARecoveredCopyCarriesTheTimeItsBackendAnswered(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "the paper's own text")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz"),
		locs:   `{"pdf_url":"https://repo.example/open.pdf","is_oa":true}`,
		bodies: map[string]*Response{"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}}}
	clockAt(t, "2030-03-04T05:06:07Z")
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Substitution != SubstitutionFullerCopy {
		t.Fatalf("the stored copy is %q, not the fuller copy this test is about", entry.Substitution)
	}
	if entry.FetchedAt != "2030-03-04T05:06:07Z" {
		t.Errorf("the recovered copy carries fetch time %q, want the clock when its backend answered", entry.FetchedAt)
	}
}

// AN ENTRY NO FETCH STAMPED IS REFUSED, and nothing reaches the index: a copy with no fetch time would
// cite with no accessed date and say nothing of why.
func TestStoreRefusesAnEntryWithNoFetchTime(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if _, err := Store(run, Entry{URL: "https://ex/unstamped"}, []byte("bytes")); err == nil {
		t.Fatal("Store accepted an entry with no fetch time")
	}
	if _, err := os.Stat(indexPath(run)); !os.IsNotExist(err) {
		t.Errorf("the refused entry reached the index: %v", err)
	}
}

// clockedSite answers from bodies, and each request it answers moves the record clock one minute on,
// so the time a stamp holds names the request whose bytes it describes.
type clockedSite struct {
	start  time.Time
	calls  int
	bodies map[string]*Response
}

func (c *clockedSite) Fetch(url string) (*Response, error) {
	c.calls++
	at := c.start.Add(time.Duration(c.calls) * time.Minute)
	record.Now = func() time.Time { return at }
	if r, ok := c.bodies[url]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("no response for %s", url)
}

// A FOLLOWED POINTER'S BYTES WERE FETCHED AT THE HOP, not at the landing page that named them.
func TestAFollowedPointersCopyCarriesTheHopsTime(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "the paper")
	clockAt(t, "2030-05-06T07:00:00Z")
	site := &clockedSite{start: record.Now(), bodies: map[string]*Response{
		"https://ex/abs":       {Body: []byte(`<html><head><meta name="citation_pdf_url" content="https://ex/paper.pdf"></head><body>abstract</body></html>`), ContentType: "text/html"},
		"https://ex/paper.pdf": {Body: []byte("%PDF-1.7 " + strings.Repeat("the paper's body ", 40)), ContentType: "application/pdf"},
	}}
	entry, _, _, err := Resolve(run, "https://ex/abs", site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.CopyURL != "https://ex/paper.pdf" {
		t.Fatalf("the stored copy is %q, not the followed pdf this test is about", entry.CopyURL)
	}
	if entry.FetchedAt != "2030-05-06T07:02:00Z" {
		t.Errorf("the followed copy carries fetch time %q, want the hop's, 2030-05-06T07:02:00Z (the landing page arrived at 07:01)", entry.FetchedAt)
	}
}
