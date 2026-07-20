package merge

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
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
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.SetSame(cmd, record.NewPayload(), flags.Observation)
			seat.Set(cmd, p, "disposition", flags.As)
			seat.SetSame(cmd, p, flags.Into)
			if err := seat.SetLongForm(cmd, p, "reason", flags.Reason); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "dispose", p); err != nil {
				return nil, err
			}
			return disposeResult{Observation: seat.Str(cmd, flags.Observation), As: seat.Str(cmd, flags.As)}, nil
		})

	c.Flags().String(flags.Observation, "", "the lens observation being disposed, by label or key")
	c.Flags().String(flags.As, "", "minted-as | folded-into | declined | banked — the observation's fate")
	c.Flags().String(flags.Into, "", "the gap id it was folded or minted into")
	c.Flags().String(flags.Reason, "", "why, when the fate is declined or banked")
	return seat.Prose(c)
}

type disposeResult struct {
	Observation string `json:"observation"`
	As          string `json:"as"`
}

func (r disposeResult) Human() string { return "disposed " + r.Observation + ": " + r.As }
