package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
)

// newScorecard is the operator command that prints a chair's scorecard for THIS run.
//
// It is NOT the seat's work list, and its Short used to call it "the seat's in-run self-read"
// three lines under a comment saying "Not a seat verb". That contradiction is how a third name
// for one thing gets built: scorecard measures how the run is GOING, `show` says what is LEFT.
//
// THE SEAT-FACING READ IS NOW `show scorecard`, and this command kept only the job it alone can
// do: read a chair OTHER than your own. --chair exists here because an operator is not a party
// and has no chair of its own; a seat has exactly one and passes nothing. The dispatch prompt
// used to send seats HERE, by telling them to override --seat-id to `operator` and find "the
// selector that names a chair" — four seats across three runs filed friction saying no such
// thing was on their surface, and the one that obeyed contradicted the rule that --seat-id
// selects your surface. Ported from
// scorecards.mjs's CLI. It reads the record IN-PROCESS (BoardState → board/findings/debate
// projections) instead of self-spawning `merge show`, plus the journal envelopes + telemetry;
// mid-run the envelope-derived rows read "not computed". Not a seat verb.
func newScorecard() *cobra.Command {
	c := &cobra.Command{
		Use:           "scorecard --run <dir> --chair blue|red|bench",
		Short:         "print a chair's scorecard for this run — OPERATOR ANALYTICS ACROSS CHAIRS. A seat reads its OWN chair with `show scorecard`, which takes no --chair because the chair is the seat it registered as",
		Long:          "scorecard computes the given chair's scorecard rows from the run's record (board/findings/debate, read in-process), its journal envelopes, and its board telemetry, and prints the markdown section — the same numbers the dashboard and the human see. Ported from scorecards.mjs. The envelope-derived rows read \"not computed\" until capture assembles the journal.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			run, rerr := seat.Of(cmd).RequireRun("scorecard")
			if rerr != nil {
				return rerr
			}
			chair, _ := cmd.Flags().GetString(flags.Chair)
			cards := map[string]bool{"blue": true, "red": true, "bench": true}
			// The run is no longer part of this test: RequireRun refused an unsupplied one
			// above, so `run.Dir() == ""` could only ever be false by the time it was read.
			// RETURNED, NOT EXITED. RunE's contract is that a refusal comes back as an error:
			// Execute renders it through EmitTopLevelError first, so a --json caller gets an
			// envelope naming the bad flag rather than a usage sentence on a channel whose whole
			// contract is that it is machine-readable (root.go says this in as many words). An
			// os.Exit here also skipped the signal guard's release and made this command
			// undrivable by any in-process test, which is how #716 was found.
			if !cards[chair] {
				return feov.Errorf(feov.Validation, "usage: %s scorecard --run <dir> --chair blue|red|bench", InvokedAs())
			}
			// A run with no record yet (BoardState errors) leaves board nil — the record-derived
			// rows then read "needs the tool", exactly as the JS did when the view spawn failed.
			var board *record.Board
			if b, err := record.BoardState(run); err == nil {
				board = b
			}
			rows := scorecard.Compute(run, scorecard.ReadResults(run), board)[chair]
			fmt.Fprint(cmd.OutOrStdout(), scorecard.RenderChair(chair, rows, "this run")+"\n")
			return nil
		},
	}
	c.Flags().String(flags.Chair, "", "which chair: blue|red|bench")
	return c
}
