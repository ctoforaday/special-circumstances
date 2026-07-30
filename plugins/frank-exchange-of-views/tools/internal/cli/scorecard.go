package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

// newScorecard is the operator command that prints a chair's scorecard for THIS run — the
// in-run self-read a seat runs before its docket (debate.js's prompt calls it). Ported from
// scorecards.mjs's CLI. It reads the record IN-PROCESS (BoardState → board/findings/debate
// projections) instead of self-spawning `merge show`, plus the journal envelopes + telemetry;
// mid-run the envelope-derived rows read "not computed". Not a seat verb.
func newScorecard() *cobra.Command {
	c := &cobra.Command{
		Use:           "scorecard --run <dir> --chair blue|red|bench",
		Short:         "print a chair's scorecard for this run (operator; the seat's in-run self-read)",
		Long:          "scorecard computes the given chair's scorecard rows from the run's record (board/findings/debate, read in-process), its journal envelopes, and its board telemetry, and prints the markdown section — the same numbers the dashboard and the human see. Ported from scorecards.mjs. The envelope-derived rows read \"not computed\" until capture assembles the journal.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runDir, _ := cmd.Flags().GetString(flags.Run)
			chair, _ := cmd.Flags().GetString(flags.Chair)
			cards := map[string]bool{"blue": true, "red": true, "bench": true}
			if runDir == "" || !cards[chair] {
				fmt.Fprintln(cmd.ErrOrStderr(), "usage: feov-record scorecard --run <runDir> --chair blue|red|bench")
				os.Exit(2)
			}
			// A run with no record yet (BoardState errors) leaves board nil — the record-derived
			// rows then read "needs the tool", exactly as the JS did when the view spawn failed.
			var board *record.Board
			if b, err := record.BoardState(runDir); err == nil {
				board = b
			}
			rows := scorecard.Compute(runDir, scorecard.ReadResults(runDir), board)[chair]
			fmt.Fprint(cmd.OutOrStdout(), scorecard.RenderChair(chair, rows, "this run")+"\n")
			return nil
		},
	}
	c.Flags().String(flags.Chair, "", "which chair: blue|red|bench")
	return c
}
