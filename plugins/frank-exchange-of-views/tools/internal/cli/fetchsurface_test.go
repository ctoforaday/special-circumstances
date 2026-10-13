package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// `fetch` IS ON THE SURFACE OF THE SEATS THAT READ SOURCES, AND ON NO OTHER SEAT'S.
//
// Blue reads what it cites and a lens reads those bytes to verify the citation. The chair reads the
// report, the board and the changes since a stale area's pin; the bench's ruling basis is confined
// to the closings, the transcript and the artifacts as the record renders them. Neither is handed a
// source, and every seat reads its whole surface at its opening act — so a page on their surfaces
// is read at every chair and bench sitting and called at none.
func TestFetchIsOnTheSurfaceOfTheSeatsThatReadSources(t *testing.T) {
	want := map[string]bool{"lens": true, "blue": true, "chair": false, "bench": false}
	for role, has := range want {
		got := false
		for _, c := range NewRootFor(record.SampleSeatOf(role)).Commands() {
			if c.Name() == "fetch" {
				got = true
			}
		}
		if got != has {
			t.Errorf("the %s surface holds `fetch`: %v, want %v", role, got, has)
		}
	}
}

// AND THE TREE IS WHAT REFUSES IT. `fetch` writes no event, so no write path can turn a seat away:
// the refusal is the wrong-seat one, it names the seats that hold the verb, and nothing reaches the
// run's source cache.
func TestASeatWithoutFetchIsRefusedAndCachesNothing(t *testing.T) {
	for _, seatID := range []string{"red-chair", "judge"} {
		dir := recordtest.TmpRun(t)
		withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://ex/a": []byte("<html>a source</html>")}, contentType: "text/html"})
		withExtractor(t, stubExtractor{})

		_, err := run(t, "fetch", "--seat-id", seatID, "--run", dir, "--url", "https://ex/a")
		if err == nil {
			t.Fatalf("%s ran `fetch` and was not refused", seatID)
		}
		for _, wantText := range []string{`"fetch" is not on your surface`, "blue and lens"} {
			if !strings.Contains(err.Error(), wantText) {
				t.Errorf("%s: the refusal does not say %q:\n%v", seatID, wantText, err)
			}
		}
		if cached, _ := filepath.Glob(filepath.Join(dir, "cache", "*")); len(cached) > 0 {
			t.Errorf("%s was refused `fetch` and the run cache holds %d file(s)", seatID, len(cached))
		}
	}
}
