package cli

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
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
			// A SEAT THAT REACHED FOR ITS OWN VERB IS TOLD WHERE IT LIVES, not that a flag is
			// unknown.
			//
			// Both blue constitutions teach the write as `friction --reason "<what blocked
			// you>"` — no role in front of it. That form lands HERE, on the operator's read,
			// and cobra rejects it at PARSE time with `unknown flag: --reason`: exit 2, before
			// RunE, before the Long text one line above that says seats write `<role>
			// friction`. The seat is told which flag is wrong, never that it is at the wrong
			// address.
			//
			// The channel it could not reach is the one for reporting exactly this. Eighteen
			// probed sittings recorded no friction at all, read as seats declining a duty; a
			// seat that typed the taught form got a parse error and no second guess.
			//
			// So the flags are DECLARED (hidden) and refused here with the address. Declaring
			// them is what moves the failure from cobra's parser to a message that can teach —
			// the same reason the root answers an unknown verb with where it lives instead of
			// `unknown command`.
			if seatFlagsUsed(cmd) {
				return feov.Errorf(feov.Validation,
					"friction: this is the OPERATOR's read of the channel; it takes no --reason, --reason-file or --none. "+
						"To WRITE friction, name your role: `<role> friction --reason \"<the capability gap and what it blocked>\"`, "+
						"or `<role> friction --none --reason \"<what you reached for and found>\"` when nothing blocked you")
			}
			// A REFUSAL AND AN ABSENCE ARE DIFFERENT ANSWERS, and this read gave the same one to
			// both: it took the path off seat.Of, where no error is reachable, and told an
			// operator who HAD supplied --run that --run was required.
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			b, err := record.FrictionJSONBytes(run)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		},
	}
	// HIDDEN, because they are not this command's flags — they exist so the seat's mistake
	// reaches a message instead of the parser. Visible would advertise them as the read's own.
	for _, f := range []string{flags.Reason, flags.ReasonFile} {
		c.Flags().String(f, "", "not this command's — see the refusal")
		_ = c.Flags().MarkHidden(f)
	}
	c.Flags().Bool(flags.None, false, "not this command's — see the refusal")
	_ = c.Flags().MarkHidden(flags.None)
	return c
}

// seatFlagsUsed reports whether the caller passed a flag that belongs to the SEAT's friction
// verb. Keyed on Changed rather than on emptiness: `--reason ""` is a seat at the wrong address
// just as much as `--reason "x"` is, and an empty string is indistinguishable from unset.
func seatFlagsUsed(cmd *cobra.Command) bool {
	for _, f := range []string{flags.Reason, flags.ReasonFile, flags.None} {
		if cmd.Flags().Changed(f) {
			return true
		}
	}
	return false
}
