package difftest

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

var (
	fuzzSourceOnce sync.Once
	fuzzSourceSrv  string
)

// fuzzSourceURL is a loopback source `blue cite` can fetch through the run cache, as the release
// gate's sourceURL is: one server for the test binary, so every replay of a sequence fetches the
// same bytes from the same URL and the determinism comparison sees the same citation twice.
func fuzzSourceURL(path string) string {
	fuzzSourceOnce.Do(func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>a source the fuzz cites at " + r.URL.Path + "</body></html>"))
		}))
		fuzzSourceSrv = srv.URL // deliberately never closed: it lives for the test binary
	})
	return fuzzSourceSrv + path
}
