package cli

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// newShowLog is the OPERATOR's read of the log, mounted under the operator's `show`.
//
// # Why it is not a seat projection
//
// Every seat WRITES the log — the verb is on all four roles, and writing to it is a duty of every
// sitting. Nobody reads it back from a seat: a missing capability is a report addressed to the
// human who can retool the seat, not material for the debate.
//
// # Why it is `show log`
//
// `log` is the seats' write verb. A read at the operator's root under the same word made one
// command name a write on four surfaces and a read on the fifth. Every operator read is under
// `show`, so `log` only ever writes.
func newShowLog() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "read the run's log (operator; every missing capability a seat reported, and every seat that reported none)",
		Long: "show log prints the entries seats filed in the log — missing capabilities, defects in the tooling, " +
			"impediments and requests — and, separately, the seats that explicitly said nothing blocked them. The two " +
			"counts are not interchangeable: no complaints and no nominal entries is a log nobody wrote.\n\n" +
			"Seats WRITE the log with their own `log --type <type> --reason \"...\"`, under their own --seat-id; " +
			"this is the read, and it is yours, not theirs.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// A REFUSAL AND AN ABSENCE ARE DIFFERENT ANSWERS: the run comes from seat.Of's Run,
			// which errors, so an operator who supplied --run is never told --run was required.
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			b, err := record.LogJSONBytes(run)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		},
	}
}
