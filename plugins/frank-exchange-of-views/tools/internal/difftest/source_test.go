package difftest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

var (
	fuzzSourceOnce sync.Once
	fuzzSourceSrv  string
	// fuzzSourceHits counts requests that reached the fuzz source over the network. A replay
	// serves its sources from the run cache, so a hit is a seed that missed.
	fuzzSourceHits atomic.Int64
)

// fuzzSourceBody is what the fuzz source holds at path: the bytes the run cache is seeded with.
func fuzzSourceBody(path string) []byte {
	return []byte("<html><body>a source the fuzz cites at " + path + "</body></html>")
}

// fuzzSourceURL is the url `blue cite` names in a generated sequence. One server for the test
// binary, so the url — and with it every command line — is the same in both replays.
//
// THE SERVER IS A TRIPWIRE, NOT THE SOURCE. When the cite fetched it live, a request that waited
// out the fetcher's 15-second header timeout logged a defect where the other replay logged a
// citation, and the determinism comparison failed on how the runner scheduled two processes
// rather than on the record (CI run 40176716191, seq12). seedFuzzSources puts the bytes in each
// run's cache before the sequence runs, so cite resolves them without a request, and replay
// fails if one arrives anyway.
func fuzzSourceURL(path string) string {
	fuzzSourceOnce.Do(func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fuzzSourceHits.Add(1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(fuzzSourceBody(r.URL.Path))
		}))
		fuzzSourceSrv = srv.URL // deliberately never closed: it lives for the test binary
	})
	return fuzzSourceSrv + path
}

// fuzzSourceFetcher serves the fuzz source's bytes in-process: no socket, no timeout, no clock.
type fuzzSourceFetcher struct{}

func (fuzzSourceFetcher) Fetch(url string) (*fetchcache.Response, error) {
	return &fetchcache.Response{Body: fuzzSourceBody(strings.TrimPrefix(url, fuzzSourceSrv)),
		ContentType: "text/html", FinalURL: url}, nil
}

// seedFuzzSources resolves every fuzz-source url the commands name into runDir's cache, through
// the fetchcache.Resolve that cite itself calls, so cite's Resolve is a cache hit.
func seedFuzzSources(t *testing.T, runDir string, cmds []cmd) {
	t.Helper()
	if fuzzSourceSrv == "" {
		return // no generated command named the fuzz source
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cmds {
		for i := 0; i+1 < len(c.args); i++ {
			if c.args[i] != "--url" || !strings.HasPrefix(c.args[i+1], fuzzSourceSrv) {
				continue
			}
			if _, _, _, err := fetchcache.Resolve(run, c.args[i+1], fuzzSourceFetcher{}); err != nil {
				t.Fatalf("seeding the run cache with %s: %v", c.args[i+1], err)
			}
		}
	}
}
