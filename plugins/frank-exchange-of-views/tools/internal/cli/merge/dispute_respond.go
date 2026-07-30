package merge

import (
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
	c := seat.New("dispute-respond",
		`red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --reason "..."`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			seat.Set(cmd, p, "response", flags.As)
			// Flag word --reason (the one prose word), payload key stays rationale.
			reason, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			if reason != "" {
				p.Set("rationale", reason)
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute-respond", p); err != nil {
				return nil, err
			}
			return disputeRespondResult{GapID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.As, "", record.MustEnum("dispute-respond", "response").Usage("your answer to blue's grade dispute"))
	return seat.Prose(c)
}

type disputeRespondResult struct {
	GapID string `json:"gap_id"`
	As    string `json:"as"`
}

func (r disputeRespondResult) Human() string { return "dispute on " + r.GapID + ": " + r.As }
