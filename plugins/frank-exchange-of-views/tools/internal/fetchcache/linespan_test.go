package fetchcache

import (
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// m39BFile is the OEIS A002808 term file m39's seats fetched (sha256 19961e39…): 17,737 rows of
// "index value", no sentence punctuation anywhere. Requests ev199, ev282 and ev345 asked to cite
// its row "66 91", and the locator found no sentence edge.
func m39BFile(t *testing.T) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "oeis-b002808.txt.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if got := Sha(b); !strings.HasPrefix(got, "19961e39") {
		t.Fatalf("the fixture hashes to %s, not the bytes m39 fetched", got)
	}
	return b
}

// storePlainSource caches body as a text type a reader reads as-is, as a fetch of a .txt does.
func storePlainSource(t *testing.T, run record.Run, url string, body []byte) Entry {
	t.Helper()
	e, err := Store(run, Entry{URL: url, ContentType: "text/plain; charset=utf-8"}, body)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// A ROW OF A SOURCE THAT ENDS NO SENTENCE IS CITED WHOLE, and so is a run of rows; a span that
// cuts into a row is refused as one cut into a clause is, naming the line it sits in.
func TestARowOfM39sTermFileIsCitedWhole(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := m39BFile(t)
	e := storePlainSource(t, run, "https://oeis.org/A002808/b002808.txt", body)

	loc, err := LocateSourceSpan(run, e, "66 91")
	if err != nil {
		t.Fatalf("the row m39 could not cite is refused: %v", err)
	}
	if loc.Matched != "66 91" || body[loc.Start-1] != '\n' || body[loc.End] != '\n' {
		t.Errorf("the row is placed at [%d,%d) as %q, not as its own line", loc.Start, loc.End, loc.Matched)
	}
	if _, err := LocateSourceSpan(run, e, "65 90 66 91"); err != nil {
		t.Errorf("two whole rows are refused: %v", err)
	}
	_, err = LocateSourceSpan(run, e, "65 90 66")
	if !errors.Is(err, ErrSpanMidClause) || !strings.Contains(err.Error(), "the source's line it sits in reads «65 90 66 91»") {
		t.Errorf("a span ending inside a row: %v, want a mid-clause refusal naming the lines", err)
	}

	// The search lists every match — the row, and the one across rows "9184 10466" and "9185 10467"
	// — each with the lines it sits in.
	got, err := SearchSource(run, e, "", "66 91")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 2 {
		t.Fatalf("want the row and the match inside another; got %d hits", len(got.Hits))
	}
	if h := got.Hits[1]; h.SourceSpan != "9184 10466 9185 10467" {
		t.Errorf("the match across two rows offers %q, want those rows", h.SourceSpan)
	}
	if h := got.Hits[0]; h.SourceSpan != "66 91" || h.Sentence != "66 91" || h.NotCiteable != "" ||
		h.SentenceStart != loc.Start || h.SentenceEnd != loc.End {
		t.Errorf("the search's hit is %+v; want the row, citeable, at cite's offsets [%d,%d)", h, loc.Start, loc.End)
	}
}

// ONE SENTENCE END ANYWHERE IN THE SOURCE KEEPS THE PROSE RULE FOR ALL OF IT, on every page: a row
// of a table inside prose is not a unit, and the search says why it is not citeable.
func TestASourceThatEndsASentenceKeepsTheProseRuleForItsRows(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	rows := strings.Repeat("1 4\n2 6\n3 8\n", 60)
	e := storePlainSource(t, run, "https://ex/table.txt", []byte("The composite numbers follow.\n"+rows+"65 90\n66 91\n67 92\n"+rows))

	if _, err := LocateSourceSpan(run, e, "66 91"); !errors.Is(err, ErrSpanMidClause) ||
		!strings.Contains(err.Error(), "the source's sentence it sits in") {
		t.Errorf("a row of a prose source: %v, want refused at its sentence", err)
	}
	got, err := SearchSource(run, e, "", "66 91")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 1 || got.Hits[0].SourceSpan != "" || !strings.Contains(got.Hits[0].NotCiteable, "its line breaks are no boundary") {
		t.Errorf("the hit is %+v, want not citeable with the reason", got.Hits)
	}

	pages := []spanText{{page: 1, text: "Table 2\n66 91\n"}, {page: 2, text: "It is composite. 91 = 7 × 13\n"}}
	if edgesOf(pages) != sentenceEdges || edgesOf(pages[:1]) != lineEdges {
		t.Error("a sentence end on one page does not decide the edges of the whole source")
	}

	long := storePlainSource(t, run, "https://ex/long.txt", []byte("1 "+strings.Repeat("9", 2*spanReach)+"\n2 4\n"))
	got, err = SearchSource(run, long, "", "1 99")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 1 || !strings.Contains(got.Hits[0].NotCiteable, "no line edge") {
		t.Errorf("a row longer than the search looks: %+v, want not citeable for want of a line edge", got.Hits)
	}
}

// WHAT COUNTS AS A SENTENCE END is the scanner's terminator rule: a decimal point is none, a full
// stop at the text's very end is one.
func TestHoldsSentenceEndReadsTheTerminatorRule(t *testing.T) {
	for _, c := range []struct {
		text string
		want bool
	}{
		{"1 4\n2 6\n", false},
		{"1 3.14\n2 2.71\n", false},
		{"n a(n)\n1 4\n", false},
		{"It is composite.\n1 4\n", true},
		{"1 4\nDone.", true},
		{"1 4\nIt is “composite.” \n", true},
	} {
		if got := holdsSentenceEnd(c.text); got != c.want {
			t.Errorf("holdsSentenceEnd(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}

// BY LINE, a line's start and end are boundaries beside every prose boundary, and nothing else is.
func TestALineBreakIsABoundaryOnlyByLine(t *testing.T) {
	for _, c := range []struct {
		name, text, span string
		lines, prose     bool
	}{
		{"a row", "65 90\n66 91\n67 92\n", "66 91", true, false},
		{"two rows", "65 90\n66 91\n67 92\n", "65 90\n66 91", true, false},
		{"an indented row", "  65 90\n  66 91\n", "66 91", true, false},
		{"a CRLF row", "65 90\r\n66 91\r\n67 92\r\n", "66 91", true, false},
		{"the last row", "65 90\n66 91", "66 91", true, false},
		{"starts inside a row", "65 90\n66 91\n67 92\n", "6 91", false, false},
		{"ends inside a row", "65 90\n66 91\n67 92\n", "66 9", false, false},
		{"a clause of a row", "a, b\nc, d\n", "d", true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			i := strings.Index(c.text, c.span)
			if got := spanAtBoundaries(c.text, i, i+len(c.span), lineEdges) == nil; got != c.lines {
				t.Errorf("by line: accepted %v, want %v", got, c.lines)
			}
			if got := spanAtBoundaries(c.text, i, i+len(c.span), sentenceEdges) == nil; got != c.prose {
				t.Errorf("by sentence: accepted %v, want %v", got, c.prose)
			}
		})
	}
	text := "65 90\n66 91\n67 92\n"
	i := strings.Index(text, "6 9")
	if got := sentenceAround(text, i, i+3, lineEdges); got != "66 91" {
		t.Errorf("the line around a cut row is %q, want the row", got)
	}
}

// A PLACE COUNTS ONLY WHERE THE SPAN STARTS AND ENDS AT A BOUNDARY: a span held twice, once cut into
// a clause, is cited at the other; held twice and cut both times, it is refused as cut.
func TestASpanIsRepeatedOnlyAtBoundaries(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	once := storePlainSource(t, run, "https://ex/once.txt", []byte("It is slow, the test fails. Then the test fails again.\n"))
	loc, err := LocateSourceSpan(run, once, "the test fails")
	if err != nil || loc.Start != len("It is slow, ") {
		t.Errorf("a span placed at a boundary once: %+v, %v; want it cited at the clause", loc, err)
	}
	cut := storePlainSource(t, run, "https://ex/cut.txt", []byte("So the test fails again. Then the test fails again.\n"))
	if _, err := LocateSourceSpan(run, cut, "the test fails"); !errors.Is(err, ErrSpanMidClause) || !strings.Contains(err.Error(), "2 times") {
		t.Errorf("a span cut at both its places: %v, want refused as cut, naming both", err)
	}
	twice := storePlainSource(t, run, "https://ex/twice.txt", []byte("It fails, the test fails. Now, the test fails.\n"))
	if _, err := LocateSourceSpan(run, twice, "the test fails"); !errors.Is(err, ErrSpanRepeated) {
		t.Errorf("a span placed at a boundary twice: %v, want refused as repeated", err)
	}
}
