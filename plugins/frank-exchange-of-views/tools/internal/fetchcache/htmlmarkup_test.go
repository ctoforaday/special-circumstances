package fetchcache

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// m37Number91 is an excerpt of Wikipedia's "91 (number)" as m37 cached it, markup as fetched: the
// list refusal ev839 stopped on, and the exponent defect logs ev57 and ev947 report.
const m37Number91 = `<html><body><ul>
<li id="mwVw">The only other ways to write 91 as a sum of distinct squares are:<br id="mwWA"/> <span class="nowrap" about="#mwt36" typeof="mw:Transclusion" id="mwWQ">91 = 1<sup>2</sup> + 4<sup>2</sup> + 5<sup>2</sup> + 7<sup>2</sup></span> and <span class="nowrap" about="#mwt37" typeof="mw:Transclusion" id="mwWg">91 = 1<sup>2</sup> + 3<sup>2</sup> + 9<sup>2</sup></span>.</li>
<li id="mwWw">the smallest <a rel="mw:WikiLink" href="https://en.wikipedia.org/wiki/Pseudoprime" title="Pseudoprime" id="mwXA">pseudoprime</a> satisfying the congruence <span class="nowrap" about="#mwt38" typeof="mw:Transclusion" id="mwXQ" data-mw='{"parts":[{"template":{"target":{"wt":"nowrap","href":"./Template:Nowrap"},"params":{"1":{"wt":"3&lt;sup>&apos;&apos;n&apos;&apos;&lt;/sup> ≡ 3 mod &apos;&apos;n&apos;&apos;"}},"i":0}}]}'>3<sup><i>n</i></sup> ≡ 3 mod <i>n</i></span>.<sup about="#mwt40" class="mw-ref reference" id="cite_ref-9" rel="dc:references" typeof="mw:Extension/ref"><a href="#cite_note-9" id="mwXg"><span class="mw-reflink-text" id="mwXw"><span class="cite-bracket" id="mwYA">[</span>9<span class="cite-bracket" id="mwYQ">]</span></span></a></sup></li>
<li id="mwYg">a <a rel="mw:WikiLink" href="https://en.wikipedia.org/wiki/Repdigit" title="Repdigit" id="mwYw">repdigit</a> in base 9 (111<sub id="mwZA">9</sub>).</li>
</ul></body></html>`

// m37Wikitext is the raw wikitext of the same lines, the source refusals ev263 and ev264 stopped on.
const m37Wikitext = `* The only other ways to write 91 as a sum of distinct squares are:<br> {{nowrap|1=91 = 1<sup>2</sup> + 3<sup>2</sup> + 9<sup>2</sup>}}.
* the smallest [[pseudoprime]] satisfying the congruence {{nowrap|1=3<sup>''n''</sup> ≡ 3 mod ''n''}}.<ref>Friedman, Erich.</ref>
* a [[repdigit]] in base 9 (111<sub>9</sub>).
** palindromic in bases 3 and 9
* a Riordan number
`

// THE EXPONENT IS THE MATHEMATICS. "3n ≡ 3 mod n" is false for 91 and "3^n ≡ 3 mod n" is true, so the
// text carries the script, and each list item is a line of its own with its reference mark as printed.
func TestAPagesScriptsAndListItemsStayInItsText(t *testing.T) {
	got := extractHTML([]byte(m37Number91)).Text
	want := "- The only other ways to write 91 as a sum of distinct squares are:\n" +
		"91 = 1^2 + 4^2 + 5^2 + 7^2 and 91 = 1^2 + 3^2 + 9^2.\n\n" +
		"- the smallest pseudoprime satisfying the congruence 3^n ≡ 3 mod n.[9]\n\n" +
		"- a repdigit in base 9 (111_9).\n"
	if got != want {
		t.Errorf("text:\n%q\nwant:\n%q", got, want)
	}
}

// EACH FORM MEASURED ON A REAL PAGE, and the two the brackets exist for. The sources are m37's
// cached Wikipedia and Friedman pages and the labelled Nature, PMC and ScienceDirect bodies.
func TestAScriptReadsByWhatItIs(t *testing.T) {
	for _, tc := range []struct{ name, html, want string }{
		{"one letter", `3<sup><i>n</i></sup>`, "3^n"},
		{"one number", `2<sup>90</sup>`, "2^90"},
		{"a compound exponent is bracketed", `a<sup>n−1</sup> ≡ 1`, "a^(n−1) ≡ 1"},
		{"a thing and its exponent stay apart", `2<sup>s</sup>d`, "2^sd"},
		{"letters and digits are compound", `a = a<sup>20d</sup>`, "a = a^(20d)"},
		{"a group's separator is compound", `10<sup>100,000</sup>`, "10^(100,000)"},
		{"a nested exponent is bracketed", `x<sup>y<sup>z</sup></sup>`, "x^(y^z)"},
		{"subscript", `10101<sub>3</sub>`, "10101_3"},
		{"compound subscript", `x<sub>n+1</sub>`, "x_(n+1)"},
		{"subscript before and after", `<sub>9</sub>C<sub>4</sub>`, "_9C_4"},
		{"chemical formula", `C<sub>11</sub>H<sub>24</sub>`, "C_11H_24"},
		{"ordinal", `the 10<sup>th</sup> Lucas number`, "the 10th Lucas number"},
		{"ordinals", `the 1<sup>st</sup>, 2<sup>nd</sup> and 3<sup>rd</sup>`, "the 1st, 2nd and 3rd"},
		{"wikipedia's mark", `prime.<sup class="mw-ref reference"><a href="#cite_note-9"><span>[</span>9<span>]</span></a></sup>`, "prime.[9]"},
		{"wikipedia's page pointer", `base a.<sup class="reference nowrap"><span>:&#8202;1402&#8202;</span></sup>`, "base a.: 1402"},
		{"a bracketed mark with no class", `prime.<sup>[9]</sup>`, "prime.[9]"},
		{"springer's linked mark", `Smith<sup class="u-js-hide"><a href="#Aff1">1</a></sup>, Jones`, "Smith[1], Jones"},
		{"two linked marks", `cells<sup><a href="/articles/x#ref-CR1">1</a>,<a href="/articles/x#ref-CR2">2</a></sup>.`, "cells[1][2]."},
		{"a link to another page is no mark", `x<sup><a href="/wiki/N">n</a></sup>`, "x^n"},
		{"text beside a link is no mark", `x<sup>2<a href="#n1">n</a></sup>`, "x^(2n)"},
		{"an unlinked number keeps its caret", `cells<sup>1</sup>.`, "cells^1."},
		{"a mark with no letter or digit", `Levy<sup>*</sup>, Keller<sup>†</sup>`, "Levy*, Keller†"},
		{"an empty script", `x<sup></sup><sub> </sub>y`, "xy"},
		{"a hidden script", `x<sup><script>var y</script></sup>y`, "xy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTML([]byte("<p>" + tc.html + "</p>")).Text; got != tc.want+"\n" {
				t.Errorf("%s = %q, want %q", tc.html, got, tc.want)
			}
		})
	}
}

// AN ITEM WITH NO TEXT HAS NO LINE, and a marker never stands alone. A block inside an item is the
// item's text, on the marker's line.
func TestAListItemIsOneMarkedUnit(t *testing.T) {
	for _, tc := range []struct{ name, html, want string }{
		{"empty", `<ul><li></li><li> &nbsp; </li><li>one</li></ul>`, "- one\n"},
		{"a block inside", `<ul><li><p>one</p></li><li><div>two</div></li></ul>`, "- one\n\n- two\n"},
		{"nested", `<ul><li>one<ul><li>two</li></ul></li></ul>`, "- one\n\n- two\n"},
		{"ordered", `<ol><li>one</li></ol>`, "- one\n"},
		{"hidden", `<ul><li><script>x</script></li><li>one</li></ul>`, "- one\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractHTML([]byte(tc.html)).Text; got != tc.want {
				t.Errorf("%s = %q, want %q", tc.html, got, tc.want)
			}
		})
	}
}

// THROUGH THE LOCATOR CITE RUNS, on m37's page: the item ev839 could not cite is cited from its
// lowercase first word, its exponent found as printed, and a search offers the item alone, not the
// item before it. The dropped form is no longer in the text.
func TestAListItemOfAPageIsCitedAndSearchedAsItself(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://en.wikipedia.org/wiki/91_(number)", extractHTML([]byte(m37Number91)).Text)

	item := "the smallest pseudoprime satisfying the congruence 3^n ≡ 3 mod n."
	loc, err := LocateSourceSpan(run, e, item)
	if err != nil {
		t.Fatalf("the list item: %v", err)
	}
	if loc.Matched != item {
		t.Errorf("matched %q", loc.Matched)
	}
	if _, err := LocateSourceSpan(run, e, "3n ≡ 3 mod n"); !errors.Is(err, ErrSpanAbsent) {
		t.Errorf("the dropped exponent: %v, want it absent", err)
	}
	if _, err := LocateSourceSpan(run, e, "pseudoprime satisfying the congruence 3^n ≡ 3 mod n."); !errors.Is(err, ErrSpanMidClause) {
		t.Errorf("a cut inside the item: %v, want it refused", err)
	}

	got, err := SearchSource(run, e, "", "3^n")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hits) != 1 {
		t.Fatalf("hits = %+v", got.Hits)
	}
	if h, want := got.Hits[0], item+"[9]"; h.SourceSpan != want || h.Sentence != want {
		t.Errorf("hit sentence %q, span %q (%s), want %q", h.Sentence, h.SourceSpan, h.NotCiteable, want)
	}
	if got, err := SearchSource(run, e, `3\^n`, ""); err != nil || len(got.Hits) != 1 {
		t.Errorf("a caret escaped in a regex: %+v, %v", got, err)
	}
}

// RAW WIKITEXT MARKS AN ITEM THE SAME WAY: the source ev263 and ev264 were refused on, whose items
// hold no blank line between them.
func TestAListMarkerOpeningALineBoundsASpan(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	e := storeTextSource(t, run, "https://en.wikipedia.org/w/index.php?title=91_(number)&action=raw", m37Wikitext)
	for _, span := range []string{
		"the smallest [[pseudoprime]] satisfying the congruence {{nowrap|1=3<sup>''n''</sup> ≡ 3 mod ''n''}}.",
		"palindromic in bases 3 and 9", // a run of markers opens it, and the next marker ends it
		"a Riordan number",
	} {
		if _, err := LocateSourceSpan(run, e, span); err != nil {
			t.Errorf("%q: %v", span, err)
		}
	}
	got, err := SearchSource(run, e, "", "palindromic")
	if err != nil || len(got.Hits) != 1 {
		t.Fatalf("search: %+v, %v", got, err)
	}
	if h := got.Hits[0]; h.SourceSpan != "palindromic in bases 3 and 9" {
		t.Errorf("an item with no terminator: sentence %q, span %q (%s)", h.Sentence, h.SourceSpan, h.NotCiteable)
	}
}

// THE MARKER IS WHAT SAYS IT, and only when it opens a line and a space follows: a product is no
// list, a word broken by a hyphen is no list, and a fragment after a formula, with no marker, still
// reads as the sentence it carries on.
func TestOnlyAMarkerOpeningALineIsOne(t *testing.T) {
	for _, tc := range []struct {
		name, text, span string
		ok               bool
	}{
		{"star at the text's start", "* the item is here.\n", "the item is here.", true},
		{"dash at the text's start", "- the item is here.\n", "the item is here.", true},
		{"bullet", "Heading\n• the item is here.\n", "the item is here.", true},
		{"indented", "Heading\n  * the item is here.\n", "the item is here.", true},
		{"a product mid-line", "It is 2 * the item count.\n", "the item count.", false},
		{"no space after it", "Heading\n*the item is here.\n", "the item is here.", false},
		{"a blank line alone", "Any a such that\n\nwhen n is composite is a liar.\n", "when n is composite is a liar.", false},
		{"ends before a marker", "* an item with no stop\n* the next\n", "an item with no stop", true},
		{"no marker follows", "An item with no stop\nthe next\n", "An item with no stop", false},
		{"a product on the next line does not end it", "An item with no stop\n*the next\n", "An item with no stop", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := strings.Index(tc.text, tc.span)
			got := spanAtBoundaries(tc.text, start, start+len(tc.span))
			if (got == nil) != tc.ok {
				t.Errorf("%q in %q: %v, want accepted=%v", tc.span, tc.text, got, tc.ok)
			}
		})
	}
}
