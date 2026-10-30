package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// viewReader is the group, the seat and the arguments that read a view on runDir, for the tests
// that drive every view from the table. A bench projection is the judge's `inquest`; every other is
// the chair's `show` and is read bare — but `source` is on the trees of the seats that read sources
// alone, and reads one cached source for a pattern, so it is read as blue over a source this stores
// in the run's cache.
func viewReader(t *testing.T, runDir, view string) (group, who string, args []string) {
	t.Helper()
	switch {
	case seat.GroupOf(view) == "inquest":
		return "inquest", "judge", []string{view}
	case view == "source":
		e, err := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: "https://ex/view-reader", ContentType: "text/plain"},
			[]byte("A source the view tests read. It holds one word twice: word, and word.\n"))
		if err != nil {
			t.Fatal(err)
		}
		return "show", citeSeat, []string{view, "--sha", e.Sha, "--match", "word"}
	}
	return "show", "red-chair", []string{view}
}

// sourceSearch runs `show source` and decodes it.
func sourceSearch(t *testing.T, seatID, runDir string, args ...string) fetchcache.SourceSearch {
	t.Helper()
	out, err := run(t, append([]string{"show", "source", "--run", runDir, "--seat-id", seatID}, args...)...)
	if err != nil {
		t.Fatalf("show source %v: %v\n%s", args, err, out)
	}
	var s fetchcache.SourceSearch
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatalf("show source %v is not its documented JSON: %v\n%s", args, err, out)
	}
	return s
}

// fetchedSha fetches the Fermat page as blue and returns the sha256 its summary prints.
func fetchedSha(t *testing.T, runDir string) string {
	t.Helper()
	out, err := run(t, "fetch", "--run", runDir, "--seat-id", citeSeat, "--url", fermatURL, "--json")
	if err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	var s struct {
		Sha string `json:"sha256"`
	}
	if err := json.Unmarshal([]byte(out), &s); err != nil || s.Sha == "" {
		t.Fatalf("fetch summary carries no sha256: %v\n%s", err, out)
	}
	return s.Sha
}

// A HIT IS CITED AS PRINTED. Blue searches the source it fetched with the keyword-in-context
// pattern m36's seats could not run in their shell, cites the hit's source_span, and the citation
// records the span at the hit's sentence offsets; a lens then reaches the same source through the
// citation's anchor.
func TestAShowSourceHitIsCitedAtTheOffsetsItPrinted(t *testing.T) {
	runDir := fermatRun(t, "text/html")
	sha := fetchedSha(t, runDir)

	got := sourceSearch(t, citeSeat, runDir, "--sha", sha, "--match", `.{0,60}Fermat liar.{0,20}`)
	if got.Selection.Matched != 1 || len(got.Hits) != 1 {
		t.Fatalf("want one hit, got %+v", got)
	}
	h := got.Hits[0]
	if h.SourceSpan == "" {
		t.Fatalf("the hit is not citeable: %s", h.NotCiteable)
	}
	out, err := fermatCite(t, runDir, "--source-span", h.SourceSpan, "--json")
	if err != nil {
		t.Fatalf("cite of the hit's source_span is refused: %v\n%s", err, out)
	}
	c := firstCiteEvent(t, runDir)
	if int(c.GetSpanStart()) != h.SentenceStart || int(c.GetSpanEnd()) != h.SentenceEnd || c.GetSpanTextSha() != h.TextSha {
		t.Errorf("the citation records [%d,%d) in %s; the hit printed [%d,%d) in %s",
			c.GetSpanStart(), c.GetSpanEnd(), c.GetSpanTextSha(), h.SentenceStart, h.SentenceEnd, h.TextSha)
	}

	byAnchor := sourceSearch(t, lensSeat, runDir, "--anchor", c.GetLabel(), "--quote", "fermat liar")
	if byAnchor.Sha != sha || len(byAnchor.Hits) != 1 || byAnchor.Hits[0].Start != h.Start+strings.Index(h.Matched, "Fermat liar") {
		t.Errorf("a lens searching by the citation's anchor reads another source or another place: %+v", byAnchor)
	}
}

// A SOURCE SEARCH REFUSES WHAT NAMES NOTHING: an unknown source, no source, both sources, no
// pattern, a pattern that matches everywhere — and the seats that read no source do not hold it.
func TestShowSourceRefusesAnUnknownSourceAndAnEmptyPattern(t *testing.T) {
	runDir := fermatRun(t, "text/html")
	sha := fetchedSha(t, runDir)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--sha", strings.Repeat("0", 64), "--match", "liar"}, "no source with sha256"},
		{[]string{"--match", "liar"}, "name the source to search"},
		{[]string{"--sha", sha}, "name what to find"},
		{[]string{"--sha", sha, "--match", ""}, "name what to find"},
		{[]string{"--sha", sha, "--match", "x*"}, "matches the empty string"},
		{[]string{"--sha", sha, "--match", "a", "--quote", "a"}, "Pass ONE"},
	} {
		out, err := run(t, append([]string{"show", "source", "--run", runDir, "--seat-id", citeSeat}, tc.args...)...)
		if err == nil || !strings.Contains(out+err.Error(), tc.want) {
			t.Errorf("show source %v: want a refusal saying %q, got %v\n%s", tc.args, tc.want, err, out)
		}
	}
	if out, err := run(t, "show", "source", "--run", runDir, "--seat-id", "red-chair", "--sha", sha, "--match", "liar"); err == nil {
		t.Errorf("the chair reads no source, and its tree holds `show source`:\n%s", out)
	}
}
