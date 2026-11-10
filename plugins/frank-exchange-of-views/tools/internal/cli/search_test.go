package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hybridsearch"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// --search LIVES ON THE TWO PROSE VIEWS AND NOWHERE ELSE. The report and the cached sources are long
// text a seat does not know the wording of; every other projection is a short list of record entries
// that --match and --quote select exhaustively, where a cut ranked list would hide an open gap.
func TestSearchLivesOnlyOnTheReportAndTheSources(t *testing.T) {
	found := 0
	for path, fs := range CommandFlags() {
		for _, f := range fs {
			if f != flags.Search {
				continue
			}
			// The key is the role and the command path: "lens show report".
			if !strings.HasSuffix(path, " show report") && !strings.HasSuffix(path, " show source") {
				t.Errorf("%s carries --search; only show report and show source rank", path)
				continue
			}
			found++
		}
	}
	if found < 2 {
		t.Errorf("--search is on %d command path(s); want it on show report and show source of the seats that read them", found)
	}
}

// --search IS REFUSED BESIDE AN EXHAUSTIVE SELECTOR, AND BESIDE THE ANCHOR READ: a seat that passes
// both would read a cut ranked list as every match, or a window as a ranking.
func TestSearchMutuallyExclusiveWithQuoteAndMatch(t *testing.T) {
	runDir, sha := searchRun(t)
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"report", "--search", "cost", "--quote", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"report", "--search", "cost", "--match", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"report", "--json", "--search", "cost", "--match", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"source", "--search", "cost", "--quote", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"source", "--search", "cost", "--match", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"source", "--sha", sha, "--search", "cost", "--quote", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"source", "--sha", sha, "--search", "cost", "--match", "cost"}, "--search cannot be combined with --quote or --match"},
		{[]string{"report", "--search", "  "}, "blank query"},
		{[]string{"source", "--search", ""}, "blank query"},
	}
	for _, c := range cases {
		out, err := run(t, append([]string{"show", "--run", runDir, "--seat-id", lensSeat}, c.args...)...)
		if err == nil || feov.CodeOf(err) != string(feov.Validation) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("show %s = %v (code %q), want a %s refusal saying %q\n%s", strings.Join(c.args, " "), err, feov.CodeOf(err), feov.Validation, c.want, out)
		}
	}
	// THE ANCHOR READ is one window, not a whole report to rank.
	cite := censusCiteAnchor(t, runDir)
	if _, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--anchor", cite, "--search", "cost"); err == nil ||
		feov.CodeOf(err) != string(feov.Validation) || !strings.Contains(err.Error(), "--anchor reads the report at one anchor") {
		t.Errorf("show report --anchor --search = %v, want the pair refused", err)
	}
	// NAMING NOTHING TO FIND is refused across sources as on one.
	for _, args := range [][]string{{"source"}, {"source", "--sha", sha}} {
		if _, err := run(t, append([]string{"show", "--run", runDir, "--seat-id", lensSeat}, args...)...); err == nil ||
			!strings.Contains(err.Error(), "name what to find: --search (ranked hybrid), --quote (exhaustive literal), or --match (exhaustive regex)") {
			t.Errorf("show %s = %v, want it told to name what to find", strings.Join(args, " "), err)
		}
	}
}

// A SELECTION ACROSS SOURCES IS EXHAUSTIVE, by regex as by literal: every match in every source with
// text, the total stated, the sources without text named — never a cut list.
func TestASelectionAcrossSourcesListsEveryMatch(t *testing.T) {
	runDir, sha := searchRun(t)
	for _, sel := range [][]string{{"--match", `cost\w*`}, {"--quote", "cost"}} {
		out, err := run(t, append([]string{"show", "source", "--run", runDir, "--seat-id", lensSeat}, sel...)...)
		if err != nil {
			t.Fatalf("show source %s: %v\n%s", sel[0], err, out)
		}
		var doc fetchcache.CrossSourceSearch
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out, `"search_scope"`) || doc.Selection.Matched != 2 || len(doc.Hits) != 2 || doc.Selection.SearchedSources != 1 {
			t.Errorf("show source %s across sources = %d hits, selection %+v; want both matches in the text source, stated exhaustively and unranked:\n%s", sel[0], len(doc.Hits), doc.Selection, out)
			continue
		}
		for _, h := range doc.Hits {
			if h.Sha != sha || h.Line == 0 {
				t.Errorf("show source %s: hit %+v does not name its source and line", sel[0], h)
			}
		}
	}
}

// A RANKED SEARCH SAYS IT IS RANKED AND CUT, AND WHAT IT SEARCHED, on the report and on the sources
// (one, and all), and each source hit carries what cite takes. The fake helper embeds by hashed
// token ids, so the ranking is real fusion over a deterministic vector, not the model's.
func TestRankedSearchStatesCutAndScope(t *testing.T) {
	useFakeEmbedder(t)
	runDir, sha := searchRun(t)
	// An image-only PDF holds no text to rank; the cross-source search names it rather than drop it.
	no := false
	if _, err := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: "https://census/scan.pdf", ContentType: "application/pdf", TextExtracted: &no}, []byte("%PDF-scan")); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--json", "--search", "moderate and rising cost")
	if err != nil {
		t.Fatalf("show report --search: %v\n%s", err, out)
	}
	var rep seat.ReportSearchJSON
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("show report --search --json is not its shape: %v\n%s", err, out)
	}
	if !rep.SearchScope.RankedAndCut || rep.SearchScope.Notice != hybridsearch.SearchNotice ||
		rep.SearchScope.Returned != len(rep.Chunks) || rep.SearchScope.SearchedChunks < len(rep.Chunks) || len(rep.Chunks) == 0 {
		t.Errorf("the report search's scope does not state a ranked, cut list of what it searched: %+v (%d chunks)", rep.SearchScope, len(rep.Chunks))
	}
	// The paragraph is ranked by its visible words; the anchors on it are listed, not left in its text.
	if top := rep.Chunks[0]; !strings.Contains(strings.ToLower(top.Text), "moderate and rising") || top.StartLine == 0 ||
		strings.Contains(top.Text, "<!--") || !slices.Contains(top.Anchors, censusCiteAnchor(t, runDir)) || top.Anchor != top.Anchors[0] {
		t.Errorf("the top-ranked paragraph is not the one holding the query's words with its anchors listed and stripped: %+v", top)
	}
	if bare, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--search", "moderate and rising cost"); err != nil ||
		!strings.Contains(bare, hybridsearch.SearchNotice) {
		t.Errorf("the bare report search does not carry the cut notice: %v\n%s", err, bare)
	}

	for _, scope := range [][]string{nil, {"--sha", sha}} {
		args := append([]string{"show", "source", "--run", runDir, "--seat-id", lensSeat, "--search", "moderate and rising cost"}, scope...)
		out, err := run(t, args...)
		if err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err, out)
		}
		var doc struct {
			Hits []struct {
				Sha        string `json:"sha256"`
				Line       int    `json:"line"`
				LastLine   int    `json:"last_line"`
				TextSha    string `json:"text_sha"`
				SourceSpan string `json:"source_span"`
			} `json:"hits"`
			Selection   json.RawMessage              `json:"selection"`
			SearchScope *fetchcache.CrossSourceScope `json:"search_scope"`
		}
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err, out)
		}
		s := doc.SearchScope
		if s == nil || !s.RankedAndCut || s.Notice != hybridsearch.SearchNotice || s.Returned != len(doc.Hits) || len(doc.Hits) == 0 {
			t.Errorf("%v: the scope does not state a ranked, cut list: %+v, %d hits\n%s", scope, s, len(doc.Hits), out)
			continue
		}
		if doc.Selection != nil {
			t.Errorf("%v: a ranked search printed a selection, which says every match is listed:\n%s", scope, out)
		}
		h := doc.Hits[0]
		if h.Line == 0 || h.LastLine < h.Line || h.TextSha == "" || h.SourceSpan == "" {
			t.Errorf("%v: the top hit does not carry what cite takes: %+v", scope, h)
		}
		if scope == nil {
			named := false
			for _, u := range s.UnsearchedSources {
				named = named || u.URL == "https://census/scan.pdf"
			}
			if h.Sha != sha || s.SearchedSources != 1 || !named {
				t.Errorf("across sources: hit %s, searched %d, unsearched %+v; want the text source searched and the scan named", h.Sha, s.SearchedSources, s.UnsearchedSources)
			}
		}
	}
}

// A RANKED SEARCH OF THE REPORT IS NO WHOLE READ: it serves a lens a cut list of paragraphs, so it
// records no read and does not move the lens's pin. The bare read beside it does, which is what
// makes the zero a measurement.
func TestSearchDoesNotAdvanceServeReportPin(t *testing.T) {
	useFakeEmbedder(t)
	runDir, _ := searchRun(t)
	registerSeatOnce(t, runDir, lensSeat)
	reads := func() int {
		m, err := record.MergedEvents(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, e := range m.Events {
			if e.GetType() == recordpb.EventType_EVENT_TYPE_REPORT_READ && e.GetSeatId() == lensSeat {
				n++
			}
		}
		return n
	}
	before := reads()
	for _, args := range [][]string{{"--search", "rising cost"}, {"--json", "--search", "rising cost"}} {
		if out, err := run(t, append([]string{"show", "report", "--run", runDir, "--seat-id", lensSeat}, args...)...); err != nil {
			t.Fatalf("show report %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	if after := reads(); after != before {
		t.Errorf("show report --search recorded %d whole read(s); a ranked search is no whole read", after-before)
	}
	if out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatalf("the bare read: %v\n%s", err, out)
	}
	if reads() != before+1 {
		t.Errorf("the bare read recorded no whole read, so the count above measures nothing")
	}
}

// searchSourceText is the text source searchRun caches: two paragraphs, so a ranking has an order.
const searchSourceText = "Costs in the region are moderate and rising, as the series shows for each year since the survey began.\n\n" +
	"The harbour was dredged twice in the decade. Nothing in this paragraph concerns what anything costs.\n"

// searchRun is censusRun with a text source cached beside the census citation's, whose fetched
// body carries no content type and so no text; it returns the text source's sha256.
func searchRun(t *testing.T) (string, string) {
	t.Helper()
	runDir := censusRun(t)
	e, err := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: "https://census/text", ContentType: "text/plain"}, []byte(searchSourceText))
	if err != nil {
		t.Fatal(err)
	}
	return runDir, e.Sha
}

// censusCiteAnchor is the anchor id of the citation censusRun records.
func censusCiteAnchor(t *testing.T, runDir string) string {
	t.Helper()
	out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	i := strings.Index(out, "C-")
	if i < 0 || len(out) < i+10 {
		t.Fatalf("the evidence view names no citation:\n%s", out)
	}
	return out[i : i+10]
}

var (
	fakeEmbedOnce sync.Once
	fakeEmbedDir  string
	fakeEmbedErr  error
)

// useFakeEmbedder points the tool at a fake feov-embed (nativehelper/testdata/fakehelper) built with
// the pins this binary carries, so a --search runs end to end without the native build.
func useFakeEmbedder(t *testing.T) {
	t.Helper()
	fakeEmbedOnce.Do(func() {
		dir, err := os.MkdirTemp("", "feov-fake-embed-")
		if err != nil {
			fakeEmbedErr = err
			return
		}
		exe := vecembed.HelperName
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		ldflags := fmt.Sprintf("-X main.digest=%s -X main.assets=%s", nativehelper.HelperDigest(vecembed.HelperName),
			strings.Join(nativehelper.HelperAssets(vecembed.HelperName), ","))
		cmd := exec.Command("go", "build", "-o", filepath.Join(dir, exe), "-ldflags", ldflags, "../nativehelper/testdata/fakehelper")
		if msg, err := cmd.CombinedOutput(); err != nil {
			fakeEmbedErr = fmt.Errorf("building the fake helper: %v\n%s", err, msg)
			return
		}
		fakeEmbedDir = dir
	})
	if fakeEmbedErr != nil {
		t.Fatal(fakeEmbedErr)
	}
	t.Setenv(nativehelper.EnvDir, fakeEmbedDir)
}
