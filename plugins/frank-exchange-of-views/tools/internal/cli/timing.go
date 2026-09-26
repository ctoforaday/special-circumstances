package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// newShowTiming is the fourth operator question: WHERE THE WALL CLOCK WENT.
//
// The two views behind it were built for an analysis that decomposed a run by hand and were never
// wired to a surface, so the engine could measure its own clock and had no way to show it. The cost
// surface renders per-seat totals, which is the one shape that cannot show a stall: fifty-nine fast
// turns beside one four-minute turn average out to nothing remarkable.
//
// JSON ONLY, and per turn as well as per bucket, because the reader's first question after "which
// bucket is big" is "which turns", and a summary that cannot be drilled into sends the asker back to
// the transcripts this exists to replace.
func newShowTiming() *cobra.Command {
	return &cobra.Command{
		Use: "timing",
		Short: "WHERE THE WALL CLOCK WENT: per seat, its turns bucketed by what they contained (thinking, a tool call, both, neither) " +
			"with the span and output tokens of each, and every turn's own span beside it. STRUCTURED JSON",
		Long: "timing answers which turns a run spent its wall clock in, and what those turns were doing.\n\n" +
			"THE BUCKETS ARE FACTS, NOT JUDGEMENTS. A threshold — what counts as a stall, what counts as healthy\n" +
			"generation — is chosen from a particular run, and freezing one here would apply it to every future run\n" +
			"by readers who never saw it chosen. So the split is by what each turn CONTAINED, and span_ms and\n" +
			"output_tokens come through per turn so an analysis applies its own cutoffs where they can be argued with.\n\n" +
			"A turn's span is the gap to the turn BEFORE it, so a seat's first turn has none: null, never zero.\n" +
			"Zero would read as instant, and any total summing it under-reports that seat by its opening turn.\n\n" +
			"measured says whether the per-turn measurement was TAKEN. False means the transcripts were never\n" +
			"read for this run, so there is nothing to decompose — a different fact from a run whose turns were\n" +
			"all fast, and one an empty answer on its own cannot tell you.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			t, err := record.SeatTimingOf(run)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(t, "", "  ")
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(append(b, '\n'))
			return nil
		},
	}
}
