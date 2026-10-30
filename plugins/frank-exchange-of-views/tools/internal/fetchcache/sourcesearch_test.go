package fetchcache

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// m36Page is a cached page of the kind m36's seats searched: a definition of primality with 91's
// factorisation in it, a sentence the page repeats, and a typographic apostrophe.
const m36Page = "Prime numbers\n\nA prime number is a natural number greater than 1 that is not a product of two smaller natural numbers. " +
	"A natural number greater than 1 that is not prime is called a composite number. " +
	"For example, 91 is composite, because 91 = 7 × 13, and it is often mistaken for a prime.\n\n" +
	"Euler’s criterion is a test.\n\nRepeated line. Repeated line.\n"

func storeTextSource(t *testing.T, run record.Run, url, text string) Entry {
	t.Helper()
	yes := true
	e, err := Store(run, Entry{URL: url, ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+text+"</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if e.TextSha, err = StoreText(run, e.Sha, []byte(text)); err != nil {
		t.Fatal(err)
	}
	return e
}

// AN m36 KEYWORD-IN-CONTEXT PATTERN WORKS HERE, AND EACH HIT IS CITEABLE AS PRINTED. adversary's
// `.{0,50}composite.{0,30}` was refused by the seats' shell grep; RE2 takes it (its tail is
// shortened here so the two matches do not run into one, as `grep -o`'s would not). Every hit's
// source_span is located by the locator cite runs, at the sentence offsets the hit printed.
func TestAKeywordInContextSearchOfASourceReturnsCiteableSentences(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://ex/primes", m36Page)

	got, err := SearchSource(run, e, `.{0,50}composite.{0,20}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Selection.Matched != 2 || len(got.Hits) != 2 || !got.Selection.Complete || got.Selection.Omitted != 0 {
		t.Fatalf("want 2 hits, complete; got selection %+v, %d hits", got.Selection, len(got.Hits))
	}
	if !strings.Contains(got.Hits[1].Matched, "91 is composite") {
		t.Errorf("the second hit's matched text is %q, want the 91 sentence's", got.Hits[1].Matched)
	}
	for i, h := range got.Hits {
		if h.SourceSpan == "" {
			t.Fatalf("hit %d is not citeable: %s", i, h.NotCiteable)
		}
		if m36Page[h.Start:h.End] != h.Matched || h.TextSha != e.TextSha {
			t.Errorf("hit %d's offsets [%d,%d) or text sha do not name its matched bytes", i, h.Start, h.End)
		}
		loc, err := LocateSourceSpan(run, e, h.SourceSpan)
		if err != nil {
			t.Fatalf("hit %d's source_span %q is refused by the locator cite runs: %v", i, h.SourceSpan, err)
		}
		if loc.Start != h.SentenceStart || loc.End != h.SentenceEnd {
			t.Errorf("hit %d: cite locates its span at [%d,%d), the hit printed [%d,%d)", i, loc.Start, loc.End, h.SentenceStart, h.SentenceEnd)
		}
	}
	if want := "For example, 91 is composite, because 91 = 7 × 13, and it is often mistaken for a prime."; got.Hits[1].SourceSpan != want {
		t.Errorf("the 91 hit's span is %q, want the whole sentence %q", got.Hits[1].SourceSpan, want)
	}
}

// A LITERAL IS READ IN THE SOURCE'S NORMALISATION, and a sentence the source holds twice is a hit
// cite cannot take as typed, which the hit says.
func TestASourceSearchQuoteFoldsTypographyAndNamesAnUnciteableSentence(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://ex/primes", m36Page)

	got, err := SearchSource(run, e, "", "euler's criterion")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 1 || got.Hits[0].Matched != "Euler’s criterion" {
		t.Fatalf("a keyboard apostrophe should find the page's typographic one; got %+v", got.Hits)
	}
	rep, err := SearchSource(run, e, "repeated", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Hits) != 2 || rep.Hits[0].SourceSpan != "" || rep.Hits[0].NotCiteable == "" {
		t.Fatalf("a sentence the source holds twice is not citeable as printed; got %+v", rep.Hits)
	}
}

// THE LIST IS BOUNDED AND SAYS WHAT IT LEFT OUT: a list capped at N that prints N must say so.
func TestASourceSearchCapsItsHitsAndCountsTheRest(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://ex/many", strings.Repeat("Word here. ", SearchHitCap+7))
	got, err := SearchSource(run, e, "word", "")
	if err != nil {
		t.Fatal(err)
	}
	s := got.Selection
	if len(got.Hits) != SearchHitCap || s.Matched != SearchHitCap+7 || s.Shown != SearchHitCap || s.Omitted != 7 || s.Complete {
		t.Fatalf("want %d shown of %d with 7 omitted, incomplete; got %d hits, %+v", SearchHitCap, SearchHitCap+7, len(got.Hits), s)
	}
}

// AN EMPTY PATTERN, ONE THAT MATCHES EVERYWHERE, AND A SOURCE WITH NO TEXT ARE REFUSED.
func TestASourceSearchRefusesAnEmptyPatternAndATextlessSource(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://ex/primes", m36Page)
	for _, p := range [][2]string{{"", ""}, {"x*", ""}, {"", "   "}} {
		if _, err := SearchSource(run, e, p[0], p[1]); !errors.Is(err, ErrPatternMatchesEmpty) {
			t.Errorf("SearchSource(%q, %q) = %v, want ErrPatternMatchesEmpty", p[0], p[1], err)
		}
	}
	no := false
	img, err := Store(run, Entry{URL: "https://ex/fig.png", ContentType: "image/png", TextExtracted: &no}, []byte("\x89PNG"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SearchSource(run, img, "prime", ""); !errors.Is(err, ErrNoText) {
		t.Errorf("a source with no text: got %v, want ErrNoText", err)
	}
}
