package blue

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// engage: in the last epoch blue answers in, blue engages its next sitting this epoch on the gaps of
// its minted after its dispatch (plans/findings-fold.md X22a; record.requireEngage).
//
// The lenses sit before blue, so a gap of blue's minted after the chair's dispatch is on no row blue
// sits for. Where a later sitting of blue's follows it answers the gap; where none does, the gap
// waited past the run's last audit — universe-m31 lost its last audit that way. The row is a dispatch
// row because "who sits, on what" is what a dispatch row says, and every reader of "has this seat
// sat for what it was dispatched for" reads it: the exchange count, the position duty, the manifest,
// capture's dispatch parity.
//
// THE SET IS THE RECORD'S. The write refuses an id the sitting does not owe and a set that leaves
// one out — the seat states it, the record holds it to it.
func newEngage() *cobra.Command {
	ids := flags.IDList(flags.GapID)
	c := seat.New("engage", engageResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		gaps := ids.Value()
		if len(gaps) == 0 {
			return nil, feov.Errorf(feov.MissingField, "requires --%s: the gaps this sitting owes, as your work list's engaged.owed names them", flags.IDs)
		}
		for _, id := range gaps {
			if err := flags.GapID().Set(id); err != nil {
				return nil, feov.Errorf(feov.Validation, "--%s: %v", flags.IDs, err)
			}
		}
		body := &recordpb.Dispatch{SeatId: proto.String(s.Identity().SeatID), GapIds: gaps}
		if _, err := record.Append(s.Identity(), body, record.Check(func(l *record.Landing) error {
			head, err := l.ReportHead()
			if err != nil {
				return err
			}
			body.Pin = proto.Int64(head)
			return nil
		})); err != nil {
			return nil, err
		}
		return engageResult{GapIDs: gaps, Pin: body.GetPin()}, nil
	})
	c.Flags().Var(ids, flags.IDs, "the `gap-ids` of every gap this sitting owes, comma-separated — your work list's engaged.owed")
	seat.SaysRequired(c, flags.IDs)
	seat.Records(c, "dispatch")
	seat.Supplies(c, "pin", "the events.id of the report head, read off the record")
	seat.Supplies(c, "seat_id", "your own seat — blue engages its own next sitting")
	return c
}

type engageResult struct {
	GapIDs []string `json:"gap_ids"`
	Pin    int64    `json:"pin"`
}

func (r engageResult) Human() string {
	return "engaged your next sitting on " + strings.Join(r.GapIDs, ", ") + " — name them in your envelope's engage"
}
