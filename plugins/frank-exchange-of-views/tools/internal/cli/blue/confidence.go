package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// confidence: per-claim confidence, the calibration substrate.
//
// Calibration cannot be computed from prose, which is why "confidence
// self-graded" was mandated for a whole suite and practised five times in 1,892
// lines. A structured per-claim grade is the difference between a clause and a
// measurement.
func newConfidence() *cobra.Command {
	c := seat.New("confidence",
		"per-claim confidence (the calibration substrate): --claim <claim-label> --confidence high|medium|low",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			// The FLAG WORDS are --claim and --confidence, matching `lens cite`; the
			// PAYLOAD KEYS stay label/grade, which is the event schema and moves on its
			// own schedule. Set rather than SetSame is what lets the two differ.
			p := record.NewPayload()
			seat.Set(cmd, p, "label", flags.Claim)
			seat.Set(cmd, p, "grade", flags.Confidence)
			if _, err := record.Append(s.RunDir, s.SeatID, "confidence", p); err != nil {
				return nil, err
			}
			return confidenceResult{Claim: seat.Str(cmd, flags.Claim)}, nil
		})

	// --label meant "your lens-scoped finding label" on two other verbs and "the claim
	// this attaches to" here: one word, two intents, and a seat generalises from the one
	// it met first. This is the claim, so it is --claim.
	c.Flags().String(flags.Claim, "", "the claim this confidence attaches to")
	c.Flags().String(flags.Confidence, "", "high | medium | low — your confidence in the claim")
	return c
}

type confidenceResult struct {
	Claim string `json:"claim"`
}

func (r confidenceResult) Human() string { return "confidence recorded for " + r.Claim }
