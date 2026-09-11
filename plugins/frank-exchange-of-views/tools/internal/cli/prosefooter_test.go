package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY VERB THAT TAKES FREE TEXT SHOWS THE QUOTING RULE, on the page the seat reads before it
// writes.
//
// The rule is attached where a free-text flag is registered (flags.Text; the prose channel goes
// through it too), so this checks two things. The attachment survives every verb's construction —
// a verb that set its Long AFTER registering would silently drop it. And NO string flag escapes the
// classification: one registered the plain way must be flags.ClosedForm, so a new free-text flag
// added with c.Flags().String fails here instead of shipping without the rule. Asked of the rendered
// help, because that is what a seat reads; asked of every seat's tree and the operator's, because
// the surface is per-role.
func TestEveryFreeTextVerbShowsTheQuotingRule(t *testing.T) {
	first := strings.SplitN(flags.ProseFooter, "\n", 2)[0]
	surface := commandsByPath()
	walkSurface(NewRootFor(record.OperatorRole), func(path []string, c *cobra.Command) {
		surface["operator "+strings.Join(path, " ")] = c
	})
	n := 0
	for path, c := range surface {
		free := false
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			switch f.Value.Type() {
			case "string", "stringArray", "stringSlice":
			default:
				return // an enum, grade, list or number refuses a bad value at parse
			}
			switch {
			case flags.IsFreeText(f) && flags.ClosedForm(f.Name):
				t.Errorf("%s --%s is registered as free text and classified closed-form: one word, one meaning", path, f.Name)
			case flags.IsFreeText(f):
				free = true
			case !flags.ClosedForm(f.Name):
				t.Errorf("%s --%s is a string flag registered the plain way: register it with flags.Text if a seat "+
					"composes its value in words, or add it to flags.ClosedForm if it is an id, key, enum, path or URL", path, f.Name)
			}
		})
		if !free {
			continue
		}
		n++
		if !strings.Contains(seat.HelpText(c), first) {
			t.Errorf("%s takes free text and its help does not carry the quoting rule", path)
		}
	}
	// A walk that found nothing would pass this test on every tree forever.
	if n < 30 {
		t.Fatalf("only %d verbs take a free-text flag — the walk is not seeing the surface", n)
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
