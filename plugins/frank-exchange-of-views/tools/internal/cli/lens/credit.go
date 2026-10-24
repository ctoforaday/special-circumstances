package lens

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// credit: answer a contradiction your finding raised with the open gap that already holds its defect
// (gblock, 2026-10-24). The gap's found_by gains the finding; no gap is minted.
//
// IT IS NOT A MINT, so it is not subject to the duplicate screen or the budget: the screen refuses
// the duplicate this replaces, and the gap credited was counted when it was minted. NOT STALE-GUARDED
// either: what it asserts is that the gap's problem is this contradiction's defect, and a gap's
// problem is fixed at its mint — its grade, blue's answer or another seat's finding does not change
// what the credit says. A gap that closed is refused at the write.
func newCredit() *cobra.Command {
	finding := flags.FindingID()
	c := seat.New("credit", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		body := &recordpb.Credit{
			GapId:   proto.String(seat.Str(cmd, flags.ID)),
			FoundBy: proto.String(seat.Str(cmd, flags.FoundBy)),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		answers, err := record.CreditAnswers(run, body.GetFoundBy())
		if err != nil {
			return nil, err
		}
		return creditResult{GapID: body.GetGapId(), FoundBy: body.GetFoundBy(), Answers: answers}, nil
	})
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the open gap that holds the defect your contradiction is")
	c.Flags().Var(finding, flags.FoundBy, "the id of your finding that raised the contradiction, as recording it returned it")
	seat.Require(c, flags.ID)
	seat.Require(c, flags.FoundBy)
	return c
}

type creditResult struct {
	GapID   string `json:"gap_id"`
	FoundBy string `json:"found_by"`
	// Answers is the contradiction the credit answers, by its id.
	Answers string `json:"answers"`
}

func (r creditResult) Human() string {
	return "credited " + r.GapID + " with " + r.FoundBy + " — it answers contradiction " + r.Answers
}
