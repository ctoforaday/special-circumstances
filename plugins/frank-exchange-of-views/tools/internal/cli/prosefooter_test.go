package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// EVERY VERB THAT TAKES PROSE SHOWS THE QUOTING RULE, on the page the seat reads before it writes.
//
// The rule is attached where the prose channel is registered (flags.Prose.Register), so this is a
// check that the attachment survives every verb's construction — a verb that set its Long AFTER
// registering would silently drop it, and the seat would never see why its backticks vanished.
// Asked of the rendered help, because that is what a seat reads; asked of every tree, because the
// surface is per-role.
func TestEveryProseVerbShowsTheQuotingRule(t *testing.T) {
	first := strings.SplitN(flags.ProseFooter, "\n", 2)[0]
	n := 0
	for path, c := range commandsByPath() {
		if flags.ProseOf(c) == nil {
			continue
		}
		n++
		if !strings.Contains(seat.HelpText(c), first) {
			t.Errorf("%s takes prose and its help does not carry the quoting rule", path)
		}
	}
	// A walk that found nothing would pass this test on every tree forever.
	if n < 20 {
		t.Fatalf("only %d verbs registered a prose channel — the walk is not seeing the surface", n)
	}
}

// THE FILE SPELLING IS GONE FROM EVERY TREE, not only from the vocabulary list.
//
// flags.All() is what the vocabulary gate compares against; a verb declaring the old flag by hand
// would still be a second way to write prose, with none of the rule that came with the first.
func TestNoVerbTakesTheFileSpellingOfProse(t *testing.T) {
	for path, c := range commandsByPath() {
		if c.Flags().Lookup("reason-file") != nil {
			t.Errorf("%s still declares --reason-file", path)
		}
	}
}
