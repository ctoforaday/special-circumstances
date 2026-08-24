package bench

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE HELP CANNOT SAY SOMETHING THE VOCABULARY DOES NOT.
//
// The line this guards used to be a hand-written sentence: "every value ends the gap except
// `carried`". It was true when written and it is a COPY of a fact the record now holds on the
// values themselves, which is the shape that has already gone wrong here twice — the predicate
// that classified a deferring disposition as closing, and the four agent-facing surfaces that
// named dispositions the record refused.
//
// A seat's entire briefing on this flag is what `--help` prints, so a help text that disagrees
// with the record is not cosmetic: it instructs the bench to do something the write path will
// refuse, and a refused ruling in a terminal sitting is a docket with no legal value.
func TestTheOpinionHelpNamesTheDeferringWordsTheRecordHas(t *testing.T) {
	deferring := record.Names(record.DeferringDispositions)
	if len(deferring) == 0 {
		t.Fatal("no disposition defers — every gap the bench touches would end, and `bench opinion` could not carry a question to a later round at all. If that is intended, this test and dispositionHelp's zero case both need rewriting; it is not a state to discover from a passing suite")
	}

	help := dispositionHelp()
	for _, word := range deferring {
		if !strings.Contains(help, word) {
			t.Errorf("the vocabulary defers on %q and the help does not mention it — a seat reading this is told every value closes, so it will never carry a gap it meant to keep alive", word)
		}
	}
	// And the other direction: a closing word must not be described as deferring. This catches the
	// clause being built from the wrong side of the filter, which reads perfectly well and inverts
	// the whole meaning.
	for _, v := range record.ClosureClasses {
		if strings.Contains(help, "except "+v.Name) || strings.Contains(help, " and "+v.Name+",") {
			t.Errorf("the help lists %q among the words that do NOT close, and the record says it closes — a bench following this would leave settled questions on the board", v.Name)
		}
	}
}

// COBRA READS A BACKQUOTED WORD IN A FLAG'S USAGE AS THE VALUE PLACEHOLDER.
//
// That is not a style rule, it is a rendering behaviour with a measured effect: "except `carried`"
// made the flag render as `--as carried`, putting one deferring value where the TYPE belongs, on
// the line a seat skims before it reads anything else. The seat's first impression of a
// seven-value vocabulary was a single word — and the wrong one, since `carried` is the only value
// that does not do what the verb is for.
func TestTheDispositionHelpDoesNotHijackTheFlagPlaceholder(t *testing.T) {
	if strings.Contains(dispositionHelp(), "`") {
		t.Error("dispositionHelp contains a backquote; cobra will render whatever it wraps as the --as value placeholder instead of the flag's type")
	}
}
