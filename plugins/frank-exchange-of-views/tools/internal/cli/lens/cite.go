package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// cite: one citation-ledger row.
//
// The ledger is what makes cross-round re-verification cheap — a claim verified
// HIGH stays verified unless its section changed, more than two rounds elapsed,
// or its source is volatile. The access date is not bookkeeping: it drives that
// staleness trigger, which is why it is a flag rather than something inferred.
func newCite() *cobra.Command {
	c := seat.New(role, "cite",
		`citation-ledger row (the cross-round re-fetch gate): --claim "..." --reference "..." --confidence high|medium|low --access-date YYYY-MM-DD`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.SetSame(cmd, record.NewPayload(), flags.Claim, flags.Reference, flags.Confidence)
			seat.Set(cmd, p, "access_date", flags.AccessDate)
			if _, err := record.Append(s.RunDir, s.SeatID, "cite", p); err != nil {
				return nil, err
			}
			return citeResult{Reference: seat.Str(cmd, flags.Reference)}, nil
		})

	c.Flags().String(flags.Claim, "", "the claim being verified, quoted from the report")
	c.Flags().String(flags.Reference, "", "the source the claim rests on")
	c.Flags().String(flags.Confidence, "", "high | medium | low — your confidence the source SUPPORTS the claim")
	c.Flags().String(flags.AccessDate, "", "YYYY-MM-DD you actually fetched it; drives the staleness re-fetch trigger")
	return c
}

type citeResult struct {
	Reference string `json:"reference"`
}

func (r citeResult) Human() string { return "citation recorded: " + r.Reference }
