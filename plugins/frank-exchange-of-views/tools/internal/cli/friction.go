package cli

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// newFriction is the OPERATOR's read of the friction channel.
//
// # Why it is not a seat projection
//
// Every seat WRITES friction — the verb is on all four roles, and closing the channel is a duty
// of every sitting. Nobody reads it back from a seat: it was in the `show` menu and named in no
// prompt or constitution, because a capability gap is a report addressed to the human who can
// retool the seat, not material for the debate.
//
// Leaving it in the seat menu made it look like part of the exchange. Taking it out and stopping
// there would have been worse: `/research` tells the operator to read friction at capture, and
// removing the only CLI path would have left that instruction pointing at nothing — trading a
// misplaced view for a broken one. So it moves here, beside `verify`, `graph` and `scorecard`,
// which are the other reads that exist for the human rather than the run.
func newFriction() *cobra.Command {
	c := &cobra.Command{
		Use:   "friction",
		Short: "read the run's friction channel (operator; every capability gap a seat reported, and every seat that reported none)",
		Long: "friction prints the capability and protocol complaints seats recorded, and — separately — the seats that " +
			"explicitly said nothing blocked them. The two counts are not interchangeable: an empty complaint list with " +
			"no attestations is a channel nobody used, which is what eighteen recorded sittings turned out to be.\n\n" +
			"Seats WRITE friction with `<role> friction`; this is the read, and it is yours, not theirs.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runDir := seat.Of(cmd).RunDir
			if runDir == "" {
				return feov.Errorf(feov.MissingField, "friction: --run <runDir> is required")
			}
			b, err := record.FrictionJSONBytes(runDir)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		},
	}
	return c
}
