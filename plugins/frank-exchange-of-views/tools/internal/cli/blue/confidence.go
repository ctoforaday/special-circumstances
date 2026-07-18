package blue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// confidence: per-claim confidence, the calibration substrate.
//
// Calibration cannot be computed from prose, which is why "confidence
// self-graded" was mandated for a whole suite and practised five times in 1,892
// lines. A structured per-claim grade is the difference between a clause and a
// measurement.
func newConfidence() *cobra.Command {
	c := seat.New(role, "confidence",
		"per-claim confidence (the calibration substrate): --label <claim-label> --grade high|medium|low",
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.SetSame(cmd, record.NewPayload(), "label", "grade")
			if _, err := record.Append(s.RunDir, s.SeatID, "confidence", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("confidence recorded for %s", seat.Str(cmd, "label")), nil
		})

	c.Flags().String("label", "", "the claim this confidence attaches to")
	c.Flags().String("grade", "", "high | medium | low — your confidence in the claim")
	return c
}
