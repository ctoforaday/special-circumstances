package fetchcache

import (
	"strings"
	"testing"
)

// EVERY RUNG OF D4's CHAIN IS LOAD-BEARING, and the corpus is why. Three of the four cited
// documents carried a meaningful /Title; little.pdf carried none; and NOT ONE of them was served
// with a Content-Disposition header — so a Disposition-first ordering falls through every time,
// and URL-tail-only turns the Auer paper into a document called "document".
func TestLabelPrefersTitleThenDispositionThenURLTail(t *testing.T) {
	for _, tc := range []struct {
		name, title, disposition, url, want string
	}{
		{
			name:  "the document's own title wins",
			title: "Active Learning Literature Survey",
			url:   "http://burrsettles.com/pub/settles.activelearning.pdf",
			want:  "Active Learning Literature Survey",
		},
		{
			name:        "a title beats a disposition that disagrees",
			title:       "Active Learning Literature Survey",
			disposition: `attachment; filename="settles.pdf"`,
			url:         "http://x/y.pdf",
			want:        "Active Learning Literature Survey",
		},
		{
			name:        "no title falls to the disposition",
			disposition: `attachment; filename="little-law-notes.pdf"`,
			url:         "http://www.columbia.edu/~ks20/stochastic-I/stochastic-I-LL.pdf",
			want:        "little-law-notes.pdf",
		},
		{
			// THE MEASURED CASE: little.pdf has no /Title and was served with no
			// Content-Disposition, so only the third rung answers.
			name: "no title and no disposition falls to the URL tail",
			url:  "http://www.columbia.edu/~ks20/stochastic-I/stochastic-I-LL.pdf",
			want: "stochastic-I-LL.pdf",
		},
		{
			// THE CASE THAT MAKES RUNG 1 WORTH HAVING. The Auer paper's URL ends in "/document".
			name: "a URL tail can be useless, and is still better than nothing",
			url:  "https://inria.hal.science/inria-00574987/document",
			want: "document",
		},
		{
			name: "a query string is not part of the name",
			url:  "https://ex/pub/paper.pdf?download=1&v=2",
			want: "paper.pdf",
		},
		{
			name: "a trailing slash does not produce an empty name",
			url:  "https://ex/papers/survey/",
			want: "survey",
		},
		{
			name:        "an unparseable disposition is skipped, not fatal",
			disposition: "attachment; filename=",
			url:         "https://ex/a/b.pdf",
			want:        "b.pdf",
		},
		{
			name:  "a title of only punctuation and spaces is not a title",
			title: "  ...  ",
			url:   "https://ex/real-name.pdf",
			want:  "real-name.pdf",
		},
	} {
		if got := Label(tc.title, tc.disposition, tc.url); got != tc.want {
			t.Errorf("%s: Label(%q,%q,%q) = %q, want %q", tc.name, tc.title, tc.disposition, tc.url, got, tc.want)
		}
	}
}

// A /Title IS WHATEVER AN AUTHOR PUT IN THE FILE. It reaches a seat's context and then,
// plausibly, a seat's shell. None of these may survive into a label: path separators (which
// would make the label look like a location), control characters (which could forge a line of
// the tool's own `key: value` output, or carry an ANSI escape), or a leading dot.
func TestLabelSanitizesHostileTitles(t *testing.T) {
	for _, tc := range []struct{ name, title, wantNot string }{
		{"path traversal", "../../etc/passwd", "/"},
		{"windows separators", `..\..\Windows\System32`, `\`},
		{"a newline forging an output line", "Real Title\ntext_extracted: true", "\n"},
		{"a carriage return", "Real Title\rmalicious", "\r"},
		{"an ANSI escape", "Real\x1b[31mTitle", "\x1b"},
		{"a NUL", "Real\x00Title", "\x00"},
		{"a drive-letter colon", "C:PleaseNo", ":"},
	} {
		got := Label(tc.title, "", "https://ex/fallback.pdf")
		if strings.Contains(got, tc.wantNot) {
			t.Errorf("%s: Label kept %q in %q", tc.name, tc.wantNot, got)
		}
		if got == "" {
			t.Errorf("%s: sanitizing emptied the label entirely; the URL tail should have answered", tc.name)
		}
	}
	// A hostile title that sanitizes to nothing must fall THROUGH the chain, not return "".
	if got := Label("///", "", "https://ex/kept.pdf"); got != "kept.pdf" {
		t.Errorf("a title that sanitizes to nothing did not fall through: got %q", got)
	}
}

// A LABEL IS BOUNDED. IEEE 1012's own /Title repeats itself twice before naming the standard,
// at 133 characters; an unbounded author string in a `key: value` line is a paragraph.
func TestLabelIsBounded(t *testing.T) {
	long := strings.Repeat("あ", 400) // runes, not bytes — a CJK title must not be cut to a third
	got := Label(long, "", "https://ex/x.pdf")
	if n := len([]rune(got)); n > maxLabelRunes {
		t.Errorf("label is %d runes, want at most %d", n, maxLabelRunes)
	}
	if !strings.HasPrefix(got, "あ") {
		t.Errorf("truncation mangled the leading runes: %q", got)
	}
}

// THE MEDIA TYPE IS THE FACT; THE PARAMETERS ARE NOT. Keeping "; charset=UTF-8" would give two
// spellings of one answer, and every consumer switches on the type alone.
func TestMediaTypeDropsParametersAndRefusesToGuess(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"application/pdf", "application/pdf"},
		{"application/pdf; charset=binary", "application/pdf"},
		{"TEXT/HTML; charset=UTF-8", "text/html"},
		{"  text/plain  ", "text/plain"},
		// An absent or unparseable header records NOT MEASURED, which the Entry carries as an
		// omitted field — never a guess at what the bytes probably are.
		{"", ""},
		{"not a media type at all", ""},
	} {
		if got := MediaType(tc.in); got != tc.want {
			t.Errorf("MediaType(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
