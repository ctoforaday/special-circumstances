package merge

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// spot-check: re-verify archived closures, and say which.
//
// The duty exists because a closure index is only as good as the last time
// anyone looked. W1.8 keys the floor on the archive's state at round START —
// run 5's round-2 merge entered with an empty archive, so the old
// from-round-2 rule degraded into a seat attesting blocks it was about to write
// itself.
func newSpotCheck() *cobra.Command {
	var ids flags.CSV

	c := seat.New(role, "spot-check",
		`the round archive spot-check record (W1.8 duty): --ids R1-4,R2-7 [--notes "..."]`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := record.NewPayload()
			seat.SetList(p, "ids", &ids)
			seat.SetSame(cmd, p, "notes")
			if _, err := record.Append(s.RunDir, s.SeatID, "spot-check", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("spot-checked %s", strings.Join(ids.Value(), ", ")), nil
		})

	c.Flags().Var(&ids, "ids", "comma-separated archived closures you re-verified this round")
	c.Flags().String("notes", "", "what the spot-check found")
	return c
}
