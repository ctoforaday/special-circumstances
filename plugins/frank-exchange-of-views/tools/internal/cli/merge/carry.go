package merge

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// carry stays with the chair: it RESTATES a closure the archive already holds (so the archive is
// not double-counted), which is a chair's account of the board and not a closure of a gap — the
// originator closes (plans/roundless.md §III.B.3); the chair carries.
func newCarry() *cobra.Command {
	c := seat.Records(seat.New("carry", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		p, err := seat.ClosurePayload(cmd)
		if err != nil {
			return nil, err
		}
		p.CarriedFrom = proto.String(seat.Str(cmd, flags.CarriedFrom))
		if _, err := record.Append(s.Identity(), p); err != nil {
			return nil, err
		}
		return seat.CloseResult{GapID: seat.Str(cmd, flags.ID), Class: recordpb.Word(p.GetClosureClass()), Carried: true}, nil
	}), "close")

	seat.ClosureFlags(c)
	c.Flags().String(flags.CarriedFrom, "", "the epoch (chair sitting) whose closure this restates")
	_ = c.MarkFlagRequired(flags.CarriedFrom)
	return seat.Prose(c)
}
