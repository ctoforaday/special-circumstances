package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dispose: give a lens observation its fate.
//
// Every observation gets one. An observation with no disposition is not "quietly
// agreed to be minor" — it is unaccounted for, and it appears in the render's
// undisposed footer saying so.
func newDispose() *cobra.Command {
	c := seat.New(role, "dispose",
		`every lens observation gets a fate: --observation <label-or-key> --as minted-as|folded-into|declined|banked [--into R2-4] [--reason "..."]`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.SetSame(cmd, record.NewPayload(), "observation")
			seat.Set(cmd, p, "disposition", "as")
			seat.SetSame(cmd, p, "into", "reason")
			if _, err := record.Append(s.RunDir, s.SeatID, "dispose", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("disposed %s: %s", seat.Str(cmd, "observation"), seat.Str(cmd, "as")), nil
		})

	c.Flags().String("observation", "", "the lens observation being disposed, by label or key")
	c.Flags().String("as", "", "minted-as | folded-into | declined | banked — the observation's fate")
	c.Flags().String("into", "", "the gap id it was folded or minted into")
	c.Flags().String("reason", "", "why, when the fate is declined or banked")
	return c
}
