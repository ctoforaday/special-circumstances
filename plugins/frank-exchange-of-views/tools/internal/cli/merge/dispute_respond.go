package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dispute-respond: red's answer when blue contests a grade.
//
// The channel exists so disagreement is accounted rather than silent. Across two
// full runs it fired ZERO times — which is a finding about the channel, not
// evidence that nobody ever disagreed, and it is why the answer is recorded as an
// event rather than settled in prose.
func newDisputeRespond() *cobra.Command {
	c := seat.New(role, "dispute-respond",
		`red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --basis "..."`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			seat.Set(cmd, p, "response", flags.As)
			// Flag word --basis (the grounds you argue from, as on dispute/regrade/
			// petition); payload key stays rationale.
			seat.Set(cmd, p, "rationale", flags.Basis)
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute-respond", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("dispute on %s: %s", seat.Str(cmd, flags.ID), seat.Str(cmd, flags.As)), nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.As, "", "accepted | rejected — your answer to blue's grade dispute")
	c.Flags().String(flags.Basis, "", "the grounds for the answer — why blue's proposed grade is accepted or refused")
	return c
}
