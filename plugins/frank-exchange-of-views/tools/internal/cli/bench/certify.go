package bench

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// certify: the run-end statement of what a human should re-examine.
//
// The bench keeps no memory between runs; what it publishes here IS its
// continuity, which is why the certification is an event rather than a closing
// remark in prose that capture might or might not carry forward.
func newCertify() *cobra.Command {
	return seat.Prose(seat.New("certify",
		`the run-end certification statement ("what I would want a human to re-examine"): --reason <statement>`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			if _, err := record.Append(s.Identity(), &recordpb.Certify{Statement: proto.String(text)}); err != nil {
				return nil, err
			}
			return seat.Msg{Message: "certification recorded"}, nil
		}))
}
