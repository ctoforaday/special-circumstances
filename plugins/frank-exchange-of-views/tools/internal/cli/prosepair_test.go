package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// A VERB THAT TAKES --reason TAKES IT THROUGH THE CHANNEL, ON EVERY SEAT, WITHOUT EXCEPTION.
//
// # The helper exists and nothing checked that anyone used it
//
// seat.Prose registers the channel through flags.RegisterPayload, which attaches the flag, its
// wording and the quoting rule in the verb's help, and records the binding flags.ReadPayload reads.
//
// The helper does that for its callers. Nothing prevented a verb from registering `--reason` by
// hand and skipping it — which `spot-check` and `outcome` both did. That is policy without
// mechanism: the rule is real, the enforcement covers only the code that opted in. A hand-registered
// --reason now has no channel behind it at all: seat.Reason refuses the read as unregistered, and
// the verb's help carries no quoting rule.
//
// `outcome --reason` is the sharpest case. Its own help says that on a judged deadlock it "is the
// only evidence the determination ever had" — the single most consequential prose field in a run.
//
// # This gate ENUMERATES rather than listing
//
// Its predecessor was eight hand-written cases, and both offenders fell outside them. A list of
// verbs to check is a list that rots the moment a verb is added; the command tree is the only
// copy that cannot.
//
// A HIDDEN --reason is exempt: the operator `log` read declares one on purpose, hidden, so a seat
// that passes it reaches a refusal naming its friction verb instead of a parse error. It is not
// that command's channel and must not become one.
func TestEveryVerbTakingReasonRegistersTheChannel(t *testing.T) {
	var checked int
	for role, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			f := c.Flags().Lookup(flags.Reason)
			if f == nil || f.Hidden {
				return
			}
			checked++
			if flags.ProseOf(c) == nil {
				t.Errorf("%s: `%s` registers --%s by hand, not through seat.Prose.\n\n"+
					"seat.Reason refuses to read a channel the verb never registered, and the verb's help "+
					"lacks the quoting rule that keeps the shell out of the seat's prose.",
					role, strings.Join(path, " "), flags.Reason)
			}
		})
	}
	if checked == 0 {
		t.Fatal("no command registers --reason — a walk that finds nothing reports every surface clean")
	}
	t.Logf("%d command(s) take prose", checked)
}

// THE BEHAVIOURAL HALF IS THE SOURCE SCAN, NOT A CASE TABLE.
//
// Registering the channel is structural; filling one payload key from the resolver and another
// from the raw flag is behavioural, and `line-of-inquiry propose` did exactly that with its flags
// correctly registered. Driving every prose verb and comparing the record does not scale: 51
// commands take prose, most need board state to run, and the version that existed covered eight,
// with both hand-registered offenders outside them.
//
// TestNoVerbReadsTheProseFlagDirectly checks the invariant instead, at every site, with no
// fixture: prose comes from the resolver. TestProseLandsInEveryFieldThatReadsTheChannel stays as
// the end-to-end demonstration for the verbs that are cheap to drive — evidence that the invariant
// means what it claims, not the coverage mechanism.
