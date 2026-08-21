package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// A VERB THAT TAKES --reason TAKES --reason-file, ON EVERY SEAT, WITHOUT EXCEPTION.
//
// # The helper exists and nothing checked that anyone used it
//
// seat.Prose registers the pair through flags.RegisterPayload, and its own doc comment says why:
// "so a verb cannot register one form and forget the other; `close` shipped exactly that leak (a
// private --file, no --text) before the channel was centralized here."
//
// The helper prevents the leak for its callers. Nothing prevented a verb from registering
// `--reason` by hand and skipping it — which `spot-check` and `outcome` both did, reproducing the
// exact defect the comment says was fixed. That is policy without mechanism: the rule is real, the
// enforcement covers only the code that opted in.
//
// # Why the pair matters more than it looks
//
// The file form is not a convenience. Measured in the 2026-07-18 run, before the channel existed:
// 68 commands carrying escaped quotes, 9 heredocs, and 37 that staged a temp file first — two of
// which failed because the staged file was not there. A verb with only `--reason` sends every seat
// with a paragraph back into fighting the shell.
//
// `outcome --reason` is the sharpest case. Its own help says that on a judged deadlock it "is the
// only evidence the determination ever had" — the single most consequential prose field in a run,
// and it had no file channel at all.
//
// # This gate ENUMERATES rather than listing
//
// Its predecessor was eight hand-written cases, and both offenders fell outside them. A list of
// verbs to check is a list that rots the moment a verb is added; the command tree is the only
// copy that cannot.
func TestEveryVerbTakingReasonAlsoTakesReasonFile(t *testing.T) {
	var checked int
	for role, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			if c.Flags().Lookup(flags.Reason) == nil {
				return
			}
			checked++
			if c.Flags().Lookup(flags.ReasonFile) == nil {
				t.Errorf("%s: `%s` registers --%s and not --%s.\n\n"+
					"seat.Prose registers the pair; a verb that registers one by hand reproduces the leak "+
					"that helper exists to prevent. A seat with a paragraph then has to fight the shell, "+
					"which is measured: 68 escaped-quote commands, 9 heredocs and 37 staged temp files in "+
					"one run before the channel was centralized.",
					role, strings.Join(path, " "), flags.Reason, flags.ReasonFile)
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
// Registering the pair is structural; filling one payload key from the resolver and another from
// the raw flag is behavioural, and `line-of-inquiry propose` did exactly that with both flags
// correctly registered. The obvious gate for it — drive both spellings through every prose verb
// and compare the record — does not scale: 51 commands take prose, most need board state to run,
// and the version that existed covered eight, with both hand-registered offenders outside them.
//
// TestNoVerbReadsTheProseFlagDirectly checks the invariant instead, at every site, with no
// fixture: prose comes from the resolver. It catches all three shapes this has taken, including
// propose's. TestBothProseChannelsProduceTheSameRecord stays as the end-to-end demonstration for
// the verbs that are cheap to drive — evidence that the invariant means what it claims, not the
// coverage mechanism.
