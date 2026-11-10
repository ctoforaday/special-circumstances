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
// its row "66 91" and could not.
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

// THE ROW m39 COULD NOT CITE IS NAMED BY ITS LINE: the search prints the line and the text it
// counts in, and the line locator resolves those to the row's words, placed where the source holds
// them. The span locator is unchanged on the same file: it infers nothing from the source's kind,
// so the row, which ends no clause, is still refused there.
func TestM39sRowIsCitedByItsLine(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := m39BFile(t)
	e := storePlainSource(t, run, "https://oeis.org/A002808/b002808.txt", body)

	got, err := SearchSource(run, e, "", "66 91")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 2 {
		t.Fatalf("want the row and the match across rows 9184-9185; got %d hits", len(got.Hits))
	}
	if h := got.Hits[0]; h.Line != 66 || h.LastLine != 66 || h.TextSha != e.Sha {
		t.Fatalf("the row's hit is at lines %d-%d of %s; want line 66 of the source's text", h.Line, h.LastLine, h.TextSha)
	}
	if h := got.Hits[1]; h.Line != 9184 || h.LastLine != 9185 {
		t.Errorf("the match across rows is at lines %d-%d, want 9184-9185", h.Line, h.LastLine)
	}

	h := got.Hits[0]
	loc, err := LocateSourceLines(run, e, h.TextSha, h.Line, h.LastLine)
	if err != nil {
		t.Fatalf("line 66: %v", err)
	}
	if loc.Matched != "66 91" || loc.Start != h.Start || loc.End != h.End || loc.TextSha != e.Sha ||
		!strings.HasSuffix(loc.Before, "65 90\n") || !strings.HasPrefix(loc.After, "\n67 92") {
		t.Errorf("line 66 resolves to %q at [%d,%d), before %q…; want the row where the search found it", loc.Matched, loc.Start, loc.End, loc.Before)
	}
	if loc, err := LocateSourceLines(run, e, e.Sha, 65, 66); err != nil || loc.Matched != "65 90\n66 91" {
		t.Errorf("lines 65-66: %q, %v; want both rows as the source holds them", loc.Matched, err)
	}
	if _, err := LocateSourceSpan(run, e, "66 91"); !errors.Is(err, ErrSpanMidClause) {
		t.Errorf("the row quoted as a span: %v, want refused at the prose boundary rule", err)
	}
}

// A LINE NUMBER POINTS INTO ONE TEXT, and only that text resolves it: a range past the text's end, a
// malformed range, a blank range and a hash the source does not hold are each refused.
func TestALineLocatorIsRefusedOutsideItsText(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storePlainSource(t, run, "https://ex/rows.txt", []byte("  65 90  \n\n66 91\r\n67 92\n"))

	if loc, err := LocateSourceLines(run, e, e.Sha, 1, 1); err != nil || loc.Matched != "65 90" {
		t.Errorf("an indented line: %q, %v; want its words, trimmed of blanks", loc.Matched, err)
	}
	if loc, err := LocateSourceLines(run, e, e.Sha, 3, 3); err != nil || loc.Matched != "66 91" {
		t.Errorf("a CRLF line: %q, %v; want it without its carriage return", loc.Matched, err)
	}
	if loc, err := LocateSourceLines(run, e, e.Sha, 4, 4); err != nil || loc.Matched != "67 92" {
		t.Errorf("the last line, ended by a line break: %q, %v", loc.Matched, err)
	}
	for _, c := range []struct {
		name        string
		sha         string
		first, last int
		want        error
	}{
		{"past the last line", e.Sha, 4, 5, ErrLinesOutside},
		{"a blank line", e.Sha, 2, 2, ErrLinesOutside},
		{"another text", strings.Repeat("0", 64), 1, 1, ErrTextGone},
	} {
		if _, err := LocateSourceLines(run, e, c.sha, c.first, c.last); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", c.name, err, c.want)
		}
	}
}

// ON AN OCR SOURCE, lines count in each page's reading, named by the page's hash, and the page goes
// on the location as a quoted span's does.
func TestALineLocatorCountsInOnePagesReading(t *testing.T) {
	run, e := ocrSource(t, map[int]string{1: "Table 1\n65 90\n", 2: "Table 2\n66 91\n"})
	rec, ok, err := ReadReadingRecord(run, e.Sha)
	if err != nil || !ok {
		t.Fatalf("reading: %v %v", ok, err)
	}
	loc, err := LocateSourceLines(run, e, rec.Pages[1].TextSha, 2, 2)
	if err != nil || loc.Matched != "66 91" || len(loc.Pages) != 1 || loc.Pages[0] != 2 || loc.Engine == "" || loc.ReadingSha != rec.TextSha {
		t.Errorf("page 2's line 2: %+v, %v; want its row on page 2 of the reading", loc, err)
	}
}

func TestParseLineRange(t *testing.T) {
	for in, want := range map[string][2]int{"66": {66, 66}, "65-66": {65, 66}, " 7 - 9 ": {7, 9}} {
		if a, z, err := ParseLineRange(in); err != nil || a != want[0] || z != want[1] {
			t.Errorf("%q = %d-%d, %v; want %v", in, a, z, err, want)
		}
	}
	for _, in := range []string{"", "0", "-3", "9-7", "a", "3-", "1-2-3"} {
		if _, _, err := ParseLineRange(in); !errors.Is(err, ErrLinesOutside) {
			t.Errorf("%q: %v, want refused", in, err)
		}
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
