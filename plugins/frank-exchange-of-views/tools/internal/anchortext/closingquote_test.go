package anchortext

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// m36Sentence is universe-m36's report line 65 after blue's edit ev333: a sentence ending on a quotation,
// its anchors standing after the closing mark.
const m36Sentence = `The same page uses the phrase "not a prime" once on each of five lines of the cached copy (a case-insensitive literal count), in three distinct answer texts, and each of those lines says that 91 is not prime<!--fx:F-d1701a96-->, for example "91 is not a Prime number"<!--gap:G-380a3523--><!--cite:C-3b49320f-->.`

// m36Location is K-f3354ca4's recorded location (computation ev335): the quote runs through the closing mark.
const m36Location = `The same page uses the phrase "not a prime" once on each of five lines of the cached copy (a case-insensitive literal count), in three distinct answer texts, and each of those lines says that 91 is not prime, for example "91 is not a Prime number"`

// AN ANCHOR NEVER SPLITS A QUOTATION ITS QUOTE CLOSED. m36 ev339: the trailing-punctuation trim took the
// closing mark off the computation's quote and the anchor landed inside the quotation, between "number" and the
// mark. Every anchor kind goes through the one placement, at the write and at replay.
func TestAnAnchorLandsPastTheClosingMarkItsQuoteCarries(t *testing.T) {
	report := "# Findings\n\n" + m36Sentence + "\n"
	for _, id := range []string{"K-f3354ca4", "G-00000001", "C-00000001"} {
		tok := anchor.Token(id)
		want := `"91 is not a Prime number"` + tok + `<!--gap:G-380a3523-->`
		got, err := Attach(report, id, m36Location)
		if err != nil {
			t.Fatalf("Attach(%s) refused the m36 location: %v", id, err)
		}
		if !strings.Contains(got, want) {
			t.Errorf("Attach(%s) placed the anchor elsewhere than past the closing mark:\n%s", id, got)
		}
		replayed, err := InsertAnchor([]byte(report), m36Location, tok)
		if err != nil {
			t.Fatalf("InsertAnchor(%s): %v", id, err)
		}
		if string(replayed) != got {
			t.Errorf("replay and the write place %s differently:\nreplay %s\nwrite  %s", id, replayed, got)
		}
	}
}

// The full stop keeps its place: an anchor sits before sentence punctuation the quote carried, and past
// a closing mark that comes before it.
func TestAnAnchorPassesTheClosingMarkButNotTheFullStop(t *testing.T) {
	report := "# H\n\nHe wrote \"it is done\". Then he left.\n\nShe wrote \"it is over.\" Then she left.\n\nThey called it 'final'. Nobody agreed.\n"
	for _, c := range []struct{ quote, want string }{
		{`They called it 'final'`, `'final'<!--computation:K-00000001-->.`},
		{`He wrote "it is done".`, `"it is done"<!--computation:K-00000001-->.`},
		{`He wrote "it is done"`, `"it is done"<!--computation:K-00000001-->.`},
		{`She wrote "it is over."`, `"it is over."<!--computation:K-00000001--> Then she`},
		{`Then he left.`, `Then he left<!--computation:K-00000001-->.`},
		// A quote that drops a full stop the report holds after the mark still names the mark.
		{`He wrote "it is done`, ""},
	} {
		got, err := Attach(report, "K-00000001", c.quote)
		if c.want == "" {
			if !errors.Is(err, ErrInQuotation) {
				t.Errorf("Attach(%q) = %v, want ErrInQuotation", c.quote, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Attach(%q): %v", c.quote, err)
		}
		if !strings.Contains(got, c.want) {
			t.Errorf("Attach(%q) does not hold %q:\n%s", c.quote, c.want, got)
		}
	}
	if at := ContentEnd(`He wrote "it is done".`); at != len(`He wrote "it is done"`) {
		t.Errorf("ContentEnd stops at %d, not past the closing mark where Attach places", at)
	}
}

// A QUOTE THAT STOPS INSIDE A QUOTATION IS REFUSED, not moved: the seat quoted the words and not the
// mark, and the tool does not carry its anchor past text it did not quote.
func TestAQuoteThatStopsBeforeAClosingMarkIsRefused(t *testing.T) {
	report := "# Findings\n\n" + m36Sentence + "\n\nShe called it “a settled question” in print.\n"
	for _, quote := range []string{
		strings.TrimSuffix(m36Location, `"`),
		strings.TrimSuffix(m36Location, `"`) + ".",
		"She called it “a settled question",
	} {
		_, err := Attach(report, "K-00000001", quote)
		if !errors.Is(err, ErrInQuotation) {
			t.Errorf("Attach(%q) = %v, want ErrInQuotation", quote, err)
		}
		if msg := Refusal(err).Error(); !strings.Contains(msg, "quote through the closing mark") {
			t.Errorf("the refusal does not teach the repair: %s", msg)
		}
	}
}

// An opening mark is not a closing one: a space in the quote's trailing run ends the walk, and a mark
// after a space opens a quotation rather than ending the quote inside one.
func TestAnOpeningMarkIsNeitherPassedNorRefused(t *testing.T) {
	report := "# H\n\nThe term is defined in \"Primes\" here. The students' work stands.\n"
	got, err := Attach(report, "K-00000001", `The term is defined in "`)
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if !strings.Contains(got, `defined in<!--computation:K-00000001--> "Primes"`) {
		t.Errorf("the anchor did not stop before the opening mark:\n%s", got)
	}
	// An apostrophe after a letter is not read as a closing quotation mark.
	if _, err := Attach(report, "K-00000002", "The students"); err != nil {
		t.Errorf("a quote ending before an apostrophe was refused: %v", err)
	}
}
