package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

// newScorecard is the operator command that prints this run's scorecards.
//
// It is NOT the seat's work list: scorecard measures how the run is GOING, `show` says what is
// LEFT. The seat-facing read is `show scorecard`, which prints the one card the seat is measured
// on and takes no selector.
//
// ALL THREE BY DEFAULT. An operator is not a party and has no card of its own, so the question
// it asks is "how is the run going", and the answer is every card. Making it name one before it
// could see any was a selector with no default the command could not have inferred. --card
// narrows to one, and an unknown value is refused by name.
//
// It reads the record IN-PROCESS (the board, findings and debate projections), plus the journal
// envelopes and telemetry; mid-run the envelope-derived rows read "not computed". Not a seat verb.
func newScorecard() *cobra.Command {
	c := &cobra.Command{
		Use:   "scorecard --run <dir> [--card red|blue|bench]",
		Short: "print this run's scorecards — red, blue and bench, or one with --card. OPERATOR ANALYTICS ACROSS CARDS. A seat reads its own with `show scorecard`, which takes no selector",
		Long: "scorecard prints this run's scorecards — red, blue and bench, in that order, each under its own `# <card> scorecard` heading — or only the one --card names. " +
			"A card's rows are computed from the run's record (the board, findings and debate projections, read in-process), its journal envelopes and its board telemetry: the same numbers the dashboard and the human see. " +
			"The envelope-derived rows read \"not computed\" until capture assembles the journal. " +
			"OPERATOR ANALYTICS ACROSS CARDS: a seat reads its own with `show scorecard`, which takes no selector because the seat it registered as decides its card. " +
			"A record that cannot be read is refused before any card is printed.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			run, rerr := seat.Of(cmd).RequireRun("scorecard")
			if rerr != nil {
				return rerr
			}
			card, _ := cmd.Flags().GetString(flags.Card)
			cards := scorecard.Cards
			if cmd.Flags().Changed(flags.Card) {
				if !isCard(card) {
					// RETURNED, NOT EXITED: Execute renders a refusal through EmitTopLevelError, so
					// a --json caller gets an envelope naming the bad flag.
					return feov.Errorf(feov.Validation, "--%s %q is not a scorecard: the cards are %s (omit --%s to print all three). usage: %s scorecard --run <dir> [--%s red|blue|bench]",
						flags.Card, card, cardList(), flags.Card, InvokedAs(), flags.Card)
				}
				cards = []string{card}
			}
			// AN EMPTY RECORD AND AN UNREADABLE ONE ARE NOT THE SAME RUN. Discarding FamilyOf's
			// error rendered every row `_not computed_ — no findings on the record yet` over a
			// record holding forty of them, and exited 0 — a plausible zero that is HARVESTED
			// into feov-memory (#743), so it outlives the run that could not produce it. An empty
			// run stays legal: its projections read zero events with no error.
			f, ferr := record.FamilyOf(run)
			if ferr != nil {
				return ferr
			}
			rows := scorecard.Compute(run, scorecard.ReadResults(run), &f)
			w := cmd.OutOrStdout()
			for i, c := range cards {
				if i > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "# %s scorecard\n\n%s\n", c, scorecard.RenderCard(rows[c], "this run"))
			}
			return nil
		},
	}
	c.Flags().String(flags.Card, "", "which card: red|blue|bench (omit to print all three)")
	return c
}

func isCard(s string) bool {
	for _, c := range scorecard.Cards {
		if s == c {
			return true
		}
	}
	return false
}

// cardList is the three cards as a sentence names them: "red, blue and bench".
func cardList() string {
	cs := scorecard.Cards
	return strings.Join(cs[:len(cs)-1], ", ") + " and " + cs[len(cs)-1]
}
