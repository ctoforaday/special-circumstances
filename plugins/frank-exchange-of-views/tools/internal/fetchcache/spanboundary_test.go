package fetchcache

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// m31Fermat is the extraction m31's blue cited (C-4d61e3f8): a display formula breaks the
// definition into three blocks, and the third starts lowercase.
const m31Fermat = "If one wants to test whether p is prime, then we can pick random integers a. This congruence is unlikely to hold.\n\n" +
	"Any a such that\n\n" +
	"a n − 1 ≡ 1 ( mod n ) {\\displaystyle a^{n-1}\\equiv 1{\\pmod {n}}}\n\n" +
	"when n is composite is known as a Fermat liar. In this case n is called Fermat pseudoprime to base a.\n"

// EACH ROW IS A SPAN AND WHETHER THE BOUNDARY RULE ADMITS IT, found in its text by the rule's own
// locator so the start is the byte the span starts at.
func TestASpanStartsWhereTheSourceStartsASentenceOrAClause(t *testing.T) {
	for _, c := range []struct {
		name, text, span string
		admit            bool
	}{
		{"m31: after a formula, lowercase, drops the subject", m31Fermat, "when n is composite is known as a Fermat liar", false},
		{"m31: from the sentence's start", m31Fermat, "Any a such that", true},
		{"after a terminator", m31Fermat, "In this case n is called", true},
		{"after a terminator, mid-sentence", m31Fermat, "this case n is called", false},
		{"start of the text", m31Fermat, "If one wants", true},
		{"start of the text, lowercase", "it holds for every base.", "it holds", true},
		{"after a comma", m31Fermat, "then we can pick random integers a", true},
		{"mid-word", m31Fermat, "ermat liar", false},
		{"mid-clause", m31Fermat, "is known as a Fermat liar", false},
		{"after a semicolon", "Any divisor establishes n as composite; otherwise it is prime.", "otherwise it is prime", true},
		{"after a colon", "There are two cases: n is prime or composite.", "n is prime or composite", true},
		{"a comma inside a number is no clause mark", "The count is 3,215 primes below it.", "215 primes below it", false},
		{"after an em dash, closed", "Fermat’s test—the first one—holds.", "the first one", true},
		{"after an en dash, spaced", "Fermat’s test – the first one – holds.", "the first one", true},
		{"after a spaced keyboard hyphen", "Fermat's test - the first one - holds.", "the first one", true},
		{"after a doubled keyboard hyphen", "Fermat's test -- the first one -- holds.", "the first one", true},
		{"after a word hyphen", "It is a well-known test.", "known test", false},
		{"just inside a bracket", "A prime number (or a prime) is a natural number.", "or a prime", true},
		{"opening with the bracket", "A prime number (or a prime) is a natural number.", "(or a prime)", true},
		{"after a closing bracket", "Holds for a (mod n) when n is prime.", "when n is prime", false},
		{"just inside a curly quote", "It is called “the sieve” by all.", "the sieve", true},
		{"after a terminator and a closer", "He said “it holds.” It fails for 91.", "It fails for 91", true},
		{"after a terminator, an abbreviation's lowercase", "Use a base, e.g. two or three.", "two or three", false},
		{"after a blank line, uppercase", "Concept\n\nFermat's little theorem states it.", "Fermat's little theorem", true},
		{"after a single line break, lowercase", "it holds for every\nbase a coprime to n.", "base a coprime", false},
		{"after a markup tag, uppercase", "<feed><summary>We prove it.</summary></feed>", "We prove it", true},
		{"after a markup tag, lowercase", "<p>We prove <b>it</b> holds.</p>", "holds", false},
		{"after a greater-than sign", "It holds when a > b and fails else.", "b and fails", false},
		{"a digit after a terminator", "It is composite. 91 is 7 times 13.", "91 is 7 times 13", true},
		{"after a terminator and a reference mark", "It grows exponentially.[127] However, it is used.", "However", true},
		{"after an ellipsis, uppercase", "It holds for 10 bases… After that it fails.", "After that it fails", true},
		{"after a bracketed word that is no reference mark", "It is prime [as shown] and odd.", "and odd", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := fold(c.text, hyphenAsIs)
			i := strings.Index(f.s, fold(c.span, hyphenAsIs).s)
			if i < 0 {
				t.Fatalf("%q not in the text", c.span)
			}
			if got := spanStartsAtBoundary(c.text, f.at[i]); got != c.admit {
				t.Errorf("%q admitted %v, want %v", c.span, got, c.admit)
			}
		})
	}
}

// THROUGH THE VERB'S LOCATOR, the m31 fragment is refused with the sentence it sits in and the
// rule, and the same definition quoted from its sentence's start is placed.
func TestAMidClauseSpanIsRefusedWithTheSentenceItSitsIn(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	e, err := Store(run, Entry{URL: "https://ex/fermat-test", ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+m31Fermat+"</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if e.TextSha, err = StoreText(run, e.Sha, []byte(m31Fermat)); err != nil {
		t.Fatal(err)
	}
	_, err = LocateSourceSpan(run, e, "when n is composite is known as a Fermat liar")
	sentence := "Any a such that a n − 1 ≡ 1 ( mod n ) {\\displaystyle a^{n-1}\\equiv 1{\\pmod {n}}} when n is composite is known as a Fermat liar."
	if !errors.Is(err, ErrSpanMidClause) || !strings.Contains(err.Error(), sentence) || !strings.Contains(err.Error(), SpanBoundary) {
		t.Errorf("the m31 fragment: %v\nwant ErrSpanMidClause showing %q and the rule", err, sentence)
	}
	whole := "Any a such that a n − 1 ≡ 1 ( mod n ) {\\displaystyle a^{n-1}\\equiv 1{\\pmod {n}}} when n is composite is known as a Fermat liar"
	if loc, err := LocateSourceSpan(run, e, whole); err != nil || !strings.HasPrefix(loc.Matched, "Any a such that\n\n") {
		t.Errorf("the definition from its sentence's start: %+v, %v", loc, err)
	}
}

// AN OCR PAGE IS HELD TO THE SAME RULE, and its first words start a span whatever their case: a
// sentence the page before began cannot be quoted across the page break.
func TestAnOCRPagesSpanStartsAtABoundary(t *testing.T) {
	run, e := ocrSource(t, map[int]string{
		1: "Spot Welding 3.7 is the method used, when the\nsheet is thin.",
		2: "it is used for thin sheet; otherwise arc welding applies.",
	})
	if _, err := LocateSourceSpan(run, e, "otherwise arc welding applies"); err != nil {
		t.Errorf("after a semicolon on page 2: %v", err)
	}
	if _, err := LocateSourceSpan(run, e, "is used for thin sheet"); !errors.Is(err, ErrSpanMidClause) {
		t.Errorf("mid-clause on page 2: %v, want ErrSpanMidClause", err)
	}
	if _, err := LocateSourceSpan(run, e, "it is used for thin sheet"); err != nil {
		t.Errorf("a page's opening words, lowercase: %v — the part of the sentence on this page is quotable", err)
	}
	if _, err := LocateSourceSpan(run, e, "when the sheet is thin"); err != nil {
		t.Errorf("after a comma, across a line break: %v", err)
	}
}

// ocrSource is a two-page scan whose reading holds pages' texts, keyed by page.
func ocrSource(t *testing.T, pages map[int]string) (record.Run, Entry) {
	t.Helper()
	withEngine(t, &fakeEngine{perCall: func(n int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: pages[n]}, nil
	}})
	run := runtest.New(t, t.TempDir())
	sha := "scan-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("p1"), []byte("p2")}, tessocr.RenderDPI)
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}
	return run, Entry{Sha: sha, ContentType: "application/pdf"}
}

// THE SENTENCE A REFUSAL SHOWS runs from the nearest sentence start before the span to the first
// sentence end after it, and a blank line ends one as a terminator does: a heading or a formula
// block is not run into the next paragraph.
func TestTheRefusalsSentenceEndsAtATerminatorOrABlankLine(t *testing.T) {
	for _, c := range []struct{ text, span, want string }{
		{m31Fermat, "is known as a Fermat liar", "Any a such that a n − 1 ≡ 1 ( mod n ) {\\displaystyle a^{n-1}\\equiv 1{\\pmod {n}}} when n is composite is known as a Fermat liar."},
		{"Concept\n\nwhen n is odd it holds\n\nNext block.", "n is odd", "Concept when n is odd it holds"},
		{"It is prime. It has no factor but\n1 and itself. Done.", "no factor", "It has no factor but 1 and itself."},
	} {
		i := strings.Index(c.text, c.span)
		if got := sentenceAround(c.text, i, i+len(c.span)); got != c.want {
			t.Errorf("sentence around %q = %q, want %q", c.span, got, c.want)
		}
	}
}

// m32PrimeElement is the definition m32's blue cited (C-a36dce3f, C-c6aacaea): one sentence whose
// condition runs on past a comma into a second clause and a restating parenthesis.
const m32PrimeElement = "An element p of a commutative ring R is said to be prime if it is not the zero element or a unit, " +
	"and for all a, b in R, whenever p divides ab, p divides a or p divides b (that is, p ∣ a b ⟹ p ∣ a ∨ p ∣ b).\n"

// EACH ROW IS A SPAN AND WHETHER THE BOUNDARY RULE ADMITS ITS END, found in its text by the rule's
// own locator so the end is the byte after the span's last.
func TestASpanEndsWhereTheSourceEndsASentenceOrAClause(t *testing.T) {
	for _, c := range []struct {
		name, text, span string
		admit            bool
	}{
		{"before a terminator", m31Fermat, "In this case n is called Fermat pseudoprime to base a", true},
		{"with its terminator", m31Fermat, "In this case n is called Fermat pseudoprime to base a.", true},
		{"m31: the whole definition, before its terminator", m31Fermat, "when n is composite is known as a Fermat liar", true},
		{"before a comma", m31Fermat, "If one wants to test whether p is prime", true},
		{"with its comma", m31Fermat, "If one wants to test whether p is prime,", true},
		{"m32 C-a36dce3f: just before the comma of “, and for all a, b in R” — a comma is a span end (owner ruling 2026-10-24)", m32PrimeElement,
			"An element p of a commutative ring R is said to be prime if it is not the zero element or a unit", true},
		{"before a semicolon", "Any divisor establishes n as composite; otherwise it is prime.", "Any divisor establishes n as composite", true},
		{"before a colon", "There are two cases: n is prime or composite.", "There are two cases", true},
		{"before an em dash, closed", "Fermat’s test—the first one—holds.", "Fermat’s test", true},
		{"with its em dash", "Fermat’s test—the first one—holds.", "Fermat’s test—", true},
		{"before an en dash, spaced", "Fermat’s test – the first one – holds.", "Fermat’s test", true},
		{"before a spaced keyboard hyphen", "Fermat's test - the first one - holds.", "Fermat's test", true},
		{"before a doubled keyboard hyphen", "Fermat's test -- the first one -- holds.", "Fermat's test", true},
		{"just inside a closing bracket", "A prime number (or a prime) is a natural number.", "or a prime", true},
		{"closing with the bracket", "A prime number (or a prime) is a natural number.", "A prime number (or a prime)", true},
		{"just inside a curly quote", "It is called “the sieve” by all.", "the sieve", true},
		{"with a terminator and a closer", "He said “it holds.” It fails for 91.", "He said “it holds.”", true},
		{"before a reference mark", "It grows exponentially.[127] However, it is used.", "It grows exponentially", true},
		{"before a blank line", "Concept\n\nFermat's little theorem states it.", "Concept", true},
		{"before a markup tag", "<p>We prove it</p><p>Next.</p>", "We prove it", true},
		{"at the text's end", "it holds for every base", "it holds for every base", true},
		{"before an ellipsis and a capital", "It holds for 10 bases… After that it fails.", "It holds for 10 bases", true},
		{"mid-clause", m31Fermat, "If one wants to test", false},
		{"mid-word", m31Fermat, "If one wan", false},
		{"m31 C-2165a593: before the object of its preposition", "This method divides n by each integer from 2 up to the square root of n. Any such integer proves it.",
			"This method divides n by each integer from 2 up to the square root of", false},
		{"m31 C-16d602ea: cut before a trailing formula — refused, whatever the formula adds",
			"The most basic method of checking the primality of a given integer ⁠ n {\\displaystyle n} ⁠ is called trial division.",
			"The most basic method of checking the primality of a given integer", false},
		{"m32 C-c6aacaea: before a spaced opening bracket (owner ruling 2026-10-24)", m32PrimeElement, "whenever p divides ab, p divides a or p divides b", true},
		{"before a spaced opening square bracket", "It holds for every base [see Note 2] and fails else.", "It holds for every base", true},
		{"before an unspaced opening bracket", "It reads f(x) for each x.", "It reads f", false},
		{"a comma inside a number is no clause mark", "The count is 3,215 primes.", "The count is 3", false},
		{"a decimal point is no terminator", "Spot Welding 3.7 is used.", "Spot Welding 3", false},
		{"before a word hyphen", "It is a well-known test.", "It is a well", false},
		{"before an apostrophe", "Fermat's test holds.", "Fermat", false},
		{"before an abbreviation's lowercase", "Use a base, e.g. two or three.", "Use a base, e.g", false},
		{"with an abbreviation's terminator", "Use a base, e.g. two or three.", "Use a base, e.g.", false},
		{"before a single line break, lowercase", "it holds for every\nbase a coprime to n.", "it holds for every", false},
		{"before a greater-than sign", "It holds when a > b and fails else.", "It holds when a", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := fold(c.text, hyphenAsIs)
			want := fold(c.span, hyphenAsIs).s
			i := strings.Index(f.s, want)
			if i < 0 {
				t.Fatalf("%q not in the text", c.span)
			}
			if got := spanEndsAtBoundary(c.text, f.end[i+len(want)-1]); got != c.admit {
				t.Errorf("%q admitted %v, want %v", c.span, got, c.admit)
			}
		})
	}
}

// THROUGH THE VERB'S LOCATOR, a span cut before its clause ends is refused with the sentence it
// sits in, the side it is cut on and the rule, and the same span quoted to the sentence's end is
// placed. The tool does not widen the first into the second.
func TestASpanCutBeforeItsClauseEndsIsRefusedWithTheSentence(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	text := "Trial division\n\nThis method divides n by each integer from 2 up to the square root of n. Any such integer proves it.\n"
	e, err := Store(run, Entry{URL: "https://ex/trial", ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+text+"</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if e.TextSha, err = StoreText(run, e.Sha, []byte(text)); err != nil {
		t.Fatal(err)
	}
	_, err = LocateSourceSpan(run, e, "This method divides n by each integer from 2 up to the square root of")
	sentence := "This method divides n by each integer from 2 up to the square root of n."
	for _, want := range []string{"ends inside one", sentence, SpanBoundary} {
		if !errors.Is(err, ErrSpanMidClause) || !strings.Contains(err.Error(), want) {
			t.Errorf("the cut span: %v\nwant ErrSpanMidClause saying %q", err, want)
		}
	}
	if loc, err := LocateSourceSpan(run, e, "This method divides n by each integer from 2 up to the square root of n"); err != nil || loc.Matched != "This method divides n by each integer from 2 up to the square root of n" {
		t.Errorf("the span to its sentence's end: %+v, %v", loc, err)
	}
	if _, err := LocateSourceSpan(run, e, "by each integer from 2 up to the square root of"); !errors.Is(err, ErrSpanMidClause) || !strings.Contains(err.Error(), "starts and ends inside one") {
		t.Errorf("cut on both sides: %v", err)
	}
}

// AN OCR PAGE'S END IS A SPAN'S END whatever stands there: a sentence the next page finishes
// cannot be quoted across the page break, so the part of it on this page is quotable.
func TestAnOCRPagesSpanEndsAtABoundary(t *testing.T) {
	run, e := ocrSource(t, map[int]string{
		1: "The method used is spot welding when the",
		2: "sheet is thin; arc welding applies otherwise.",
	})
	if _, err := LocateSourceSpan(run, e, "The method used is spot welding when the"); err != nil {
		t.Errorf("to the page's end: %v — the part of the sentence on this page is quotable", err)
	}
	if _, err := LocateSourceSpan(run, e, "The method used is spot welding"); !errors.Is(err, ErrSpanMidClause) {
		t.Errorf("mid-clause on page 1: %v, want ErrSpanMidClause", err)
	}
	if _, err := LocateSourceSpan(run, e, "sheet is thin"); err != nil {
		t.Errorf("before a semicolon on page 2: %v", err)
	}
}

// THE EXAMPLE THE RULE'S TEXT SHOWS IS WHAT THE TOOL DOES: each span it calls accepted passes
// both boundaries, and each it calls refused fails at the side it names.
func TestSpanBoundarysExampleIsWhatTheToolDoes(t *testing.T) {
	for _, c := range []struct {
		span     string
		from, to bool
	}{
		{spanExampleWhole, true, true},
		{spanExampleClause, true, true},
		{spanExampleCutStart, false, true},
		{spanExampleCutEnd, true, false},
	} {
		i := strings.Index(spanExample, c.span)
		if i < 0 || !strings.Contains(SpanBoundary, "«"+c.span+"»") {
			t.Fatalf("%q is not in the example or not shown in SpanBoundary", c.span)
		}
		from, to := spanStartsAtBoundary(spanExample, i), spanEndsAtBoundary(spanExample, i+len(c.span))
		if from != c.from || to != c.to {
			t.Errorf("%q: starts at a boundary %v, ends at one %v; SpanBoundary says %v, %v", c.span, from, to, c.from, c.to)
		}
	}
}
