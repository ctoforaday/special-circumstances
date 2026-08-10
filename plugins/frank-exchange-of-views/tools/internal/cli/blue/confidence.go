package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
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

	// --label meant "your observation label" on observe and "the claim this attaches to"
	// here: one word, two intents, and a seat generalises from the one it met first. This
	// is the claim, so it is --claim. (finding no longer takes --label — the tool assigns
	// its L{role}-F{N} — so observe is the only other verb the word appears on.)
	c.Flags().String(flags.Claim, "", "the claim this confidence attaches to")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("confidence", "grade"), "your confidence in the claim")
	return c
}

type confidenceResult struct {
	Claim string `json:"claim"`
}

func (r confidenceResult) Human() string { return "confidence recorded for " + r.Claim }
