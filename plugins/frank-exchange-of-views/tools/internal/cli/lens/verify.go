package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// verify: red records that it CHECKED a claim against its source, and how far it trusts it.
//
// SPLIT FROM `cite` (#341), and the reason is the sharpest instance of facts-are-fields in the
// shipped tool. Blue authoring a citation and red verifying one shared the `cite` event type,
// told apart at read time by:
//
//	func IsVerifiedCite(e Event) bool { return e.Type == "cite" && e.Payload.Str("label") == "" }
//
// The distinction was the ABSENCE OF A FIELD. A blue cite written without a label counted as
// red's audit volume — a number red reads as how much work it did — with no error and no signal.
// Two acts, two event types now, and nothing infers which is which.
//
// The ledger is what makes cross-round re-verification cheap — a claim verified HIGH stays
// verified unless its section changed, more than two rounds elapsed, or its source is volatile.
// The access date is not bookkeeping: it drives that staleness trigger, which is why it is a
// flag rather than something inferred.
func newVerify() *cobra.Command {
	c := seat.New("verify",
		`record that you CHECKED a claim against its source (the cross-round re-fetch gate): --claim "..." --reference "..." --trust high|medium|low --access-date YYYY-MM-DD`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.SetSame(cmd, record.NewPayload(), flags.Claim, flags.Reference, flags.Trust)
			seat.Set(cmd, p, "access_date", flags.AccessDate)
			if _, err := record.Append(s.RunDir, s.SeatID, "verify", p); err != nil {
				return nil, err
			}
			return verifyResult{Reference: seat.Str(cmd, flags.Reference)}, nil
		})

	c.Flags().String(flags.Claim, "", "the claim being verified, quoted from the report")
	c.Flags().String(flags.Reference, "", "the source the claim rests on")
	c.Flags().String(flags.Trust, "", record.MustEnum("verify", "trust").Usage("how far the source SUPPORTS the claim. Named --trust, not --confidence: blue's `confidence` is its self-grade on a claim, and one word for two questions is how the two came to share an event type"))
	c.Flags().String(flags.AccessDate, "", "YYYY-MM-DD you actually fetched it; drives the staleness re-fetch trigger")
	return c
}

type verifyResult struct {
	Reference string `json:"reference"`
}

func (r verifyResult) Human() string { return "source verified: " + r.Reference }
