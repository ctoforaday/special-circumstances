package bench

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// halt: the safety boundary.
//
// Its own verb rather than a disposition value, because ending the run is not a
// grade of ruling — and capture relays the written opinion VERBATIM, never
// smoothed, so a halt reaches the human in the words the bench chose.
func newHalt() *cobra.Command {
	return seat.Prose(seat.New("halt",
		"the safety boundary: --reason <written opinion — capture relays it like a FAIL, never smoothed>",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "halt", record.NewPayload().Set("opinion", text)); err != nil {
				return nil, err
			}
			return seat.Msg{Message: "JUDICIAL HALT recorded — capture relays this verbatim"}, nil
		}))
}
