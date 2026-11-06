package cli

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
)

const (
	bfileURL   = "https://oeis.org/A002808/b002808.txt"
	bfileClaim = "The composite-number list includes 91 as its 66th term."
)

// bfileRun is a run whose report holds bfileClaim, with m39's A002808 term file — the bytes its
// seats fetched — answering bfileURL as plain text. It returns the run and the file's sha256,
// which is the text a plain-text source's lines count in.
func bfileRun(t *testing.T) (string, string) {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "fetchcache", "testdata", "oeis-b002808.txt.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+bfileClaim+"\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{bfileURL: body}, contentType: "text/plain"})
	return runDir, fetchcache.Sha(body)
}

func bfileCite(t *testing.T, runDir string, args ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"cite", "--run", runDir, "--seat-id", citeSeat, "--quote", bfileClaim,
		"--url", bfileURL, "--title", "OEIS A002808: The composite numbers, table of n, a(n)"}, args...)...)
}

// m39's ROW IS CITED BY ITS LINE. The line the search printed resolves to the row's words, and the
// citation holds them in the one shape every citation holds: the words as its span, the source's
// bytes and the text around them — which the result echoes, so blue sees what it cited.
func TestARowIsCitedByTheLineTheSearchPrinted(t *testing.T) {
	runDir, sha := bfileRun(t)
	if out, err := run(t, "fetch", "--run", runDir, "--seat-id", citeSeat, "--url", bfileURL); err != nil {
		t.Fatalf("fetch: %v\n%s", err, out)
	}
	hits := sourceSearch(t, citeSeat, runDir, "--sha", sha, "--quote", "66 91").Hits
	if len(hits) == 0 || hits[0].Line != 66 || hits[0].TextSha != sha {
		t.Fatalf("the search's first hit is %+v; want line 66 of the file", hits)
	}
	out, err := bfileCite(t, runDir, "--source-lines", "66", "--text-sha", hits[0].TextSha, "--json")
	if err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	var res struct {
		Result struct {
			Matched string `json:"source_matched"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil || res.Result.Matched != "66 91" {
		t.Errorf("the cite echoes %q (%v), want the row:\n%s", res.Result.Matched, err, out)
	}
	c := firstCiteEvent(t, runDir)
	if c == nil || c.GetSourceSpan() != "66 91" || c.GetSourceMatched() != "66 91" || c.GetSpanTextSha() != sha ||
		int(c.GetSpanStart()) != hits[0].Start || !strings.HasSuffix(c.GetSourceBefore(), "65 90\n") {
		t.Errorf("the citation holds %+v; want the row, placed where the search found it", c)
	}
}

// A CITATION NAMES ITS WORDS ONE WAY. Both locators, neither, a line locator with no text named,
// lines past the text's end and a text the source does not hold are each refused, and refused
// before anything is written.
func TestACitesLocatorIsRefusedUnlessItNamesOneTextOnce(t *testing.T) {
	runDir, sha := bfileRun(t)
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"both locators", []string{"--source-span", "66 91", "--source-lines", "66", "--text-sha", sha}, "pass one"},
		{"neither", nil, "--source-lines"},
		{"lines with no text", []string{"--source-lines", "66"}, "requires --text-sha"},
		{"a text with no lines", []string{"--text-sha", sha}, "no --source-lines was given"},
		{"past the last line", []string{"--source-lines", "17744-17745", "--text-sha", sha}, "the text has 17744 lines"},
		{"a malformed range", []string{"--source-lines", "67-66", "--text-sha", sha}, "not a line or a range"},
		{"another text", []string{"--source-lines", "66", "--text-sha", strings.Repeat("0", 64)}, "holds no text with that hash"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, err := bfileCite(t, runDir, c.args...)
			mustRefuse(t, err, out, c.want)
			if ev := firstCiteEvent(t, runDir); ev != nil {
				t.Errorf("a refused cite left an event: %v", ev)
			}
		})
	}
}

// A KEYED RETRY IS JUDGED ON THE WORDS ITS LOCATOR NAMES: the same row named again, by line or
// quoted, is the same citation; another row under the key is refused.
func TestAKeyedRetryOfALineCiteIsJudgedOnItsWords(t *testing.T) {
	runDir, sha := bfileRun(t)
	if out, err := bfileCite(t, runDir, "--key", "row", "--source-lines", "66", "--text-sha", sha); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if out, err := bfileCite(t, runDir, "--key", "row", "--source-lines", "66", "--text-sha", sha); err != nil || !strings.Contains(out, "66 91") {
		t.Errorf("the same line again: %v\n%s", err, out)
	}
	if out, err := bfileCite(t, runDir, "--key", "row", "--source-span", "66 91"); err != nil {
		t.Errorf("the same words quoted: %v\n%s", err, out)
	}
	out, err := bfileCite(t, runDir, "--key", "row", "--source-lines", "67", "--text-sha", sha)
	mustRefuse(t, err, out, "already names citation")
}
