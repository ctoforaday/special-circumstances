package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
		`red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --rationale "..."`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", "id")
			seat.Set(cmd, p, "response", "as")
			seat.SetSame(cmd, p, "rationale")
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute-respond", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("dispute on %s: %s", seat.Str(cmd, "id"), seat.Str(cmd, "as")), nil
		})

	c.Flags().String("id", "", "the gap id")
	c.Flags().String("as", "", "accepted | rejected — your answer to blue's grade dispute")
	c.Flags().String("rationale", "", "the reasoning behind the answer")
	return c
}
