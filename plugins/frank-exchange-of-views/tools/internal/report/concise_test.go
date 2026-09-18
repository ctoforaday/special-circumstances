package report

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// THE SCAN SURFACE HAD NO TEST AT ALL, which is why both of its defects shipped and why a
// smoke run's risk matrix carried the cell "e.". Every row here is a cut that a reader saw
// and could not read; the invariants below are what "cut" is allowed to mean.
func TestConciseCutsAtABoundaryAndNeverMidWord(t *testing.T) {
	long := strings.Repeat("alpha ", 40) // no punctuation anywhere, so the fallback runs

	for _, tt := range []struct {
		name string
		in   string
		want string
	}{{
		name: "an abbreviation does not end a sentence",
		in:   "The parser mis-handles escapes, e.g. a backslash before a quote, and drops the rest of the field entirely.",
		want: "The parser mis-handles escapes, e.g. a backslash before a quote, and drops the rest of the field…",
	}, {
		name: "a decimal does not end a sentence",
		in:   "Throughput fell to 3.5 requests per second under the contended lock, which the budget does not survive.",
		want: "Throughput fell to 3.5 requests per second under the contended lock, which the budget does not…",
	}, {
		// The uppercase rule alone does NOT cover this: the character after the dot is a
		// capital, and only "no space follows" rejects it. Found by deleting the branch.
		name: "an uppercase-suffixed token is not a sentence end",
		in:   "The retry path was introduced in release 4.X and has never been exercised by the suite since.",
		want: "The retry path was introduced in release 4.X and has never been exercised by the suite since.",
	}, {
		name: "a real first sentence is taken whole",
		in:   "The lock is acquired twice on the error path. A second paragraph follows that nobody scanning needs.",
		want: "The lock is acquired twice on the error path.",
	}, {
		name: "a short field is returned untouched",
		in:   "Off-by-one in the header length",
		want: "Off-by-one in the header length",
	}, {
		name: "newlines collapse to one line",
		in:   "first line\n\tsecond line",
		want: "first line second line",
	}, {
		name: "no punctuation: cut at the last space, not mid-word",
		in:   long,
		want: strings.TrimSpace(strings.Repeat("alpha ", 16)) + "…", // 16*6=96; a 17th would cross 100
	}} {
		t.Run(tt.name, func(t *testing.T) {
			got := concise(tt.in)
			if got != tt.want {
				t.Errorf("concise():\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

// A CELL IS ALLOWED TO BE SHORT; IT IS NOT ALLOWED TO BE UNREADABLE. These hold for every
// input, and they are the properties the two shipped defects violated: "e." was a whole cell,
// and a byte cut through a multi-byte rune renders as U+FFFD.
func TestConciseInvariants(t *testing.T) {
	inputs := []string{
		"e.g. this",
		"i.e. the second reader of the same string",
		"A. B. Initial-heavy prose that runs on well past the limit and keeps running for a while yet.",
		strings.Repeat("verylongtokenwithnospaces", 20), // nothing to back up to
		strings.Repeat("日本語のテキストです", 30),                // every rune multi-byte
		"Ends exactly at the boundary" + strings.Repeat(" x", 60),
		"",
		"   ",
		".",
		"? ",
	}
	for _, in := range inputs {
		got := concise(in)
		if !utf8.ValidString(got) {
			t.Errorf("concise(%q) = %q: not valid UTF-8", in, got)
		}
		// The cut marker costs three bytes; nothing else may exceed the stated limit.
		if n := len(strings.TrimSuffix(got, "…")); n > conciseLimit {
			t.Errorf("concise(%q) = %q: %d bytes, over the stated limit of %d", in, got, n, conciseLimit)
		}
		// A cell that is ONLY an abbreviation's stub is the defect, restated as a property:
		// a truncated cell must carry more than a token and a period.
		if trimmed := strings.TrimSpace(got); len(in) > minSentence && len(trimmed) < 3 {
			t.Errorf("concise(%q) = %q: a cut cell that says nothing", in, got)
		}
	}
}

// The footer STATES the limit, so a reader meets the rule at the place the rule applies
// rather than inferring it from a row that looks unfinished.
func TestRiskMatrixFooterStatesTheLimit(t *testing.T) {
	if !strings.Contains(riskMatrixFooter(), "100") {
		t.Errorf("the matrix footer does not state its truncation limit:\n%s", riskMatrixFooter())
	}
}
