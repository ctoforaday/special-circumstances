package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// observe: a below-bar finding that still demands a FATE.
//
// The merge must dispose every observation. That is why this is a recorded event
// rather than an unrecorded note in a seat's prose: an observation nobody disposed
// shows in the render's undisposed footer, where a note that quietly evaporated would not.
func newObserve() *cobra.Command {
	c := seat.Prose(seat.New("observe",
		"a below-bar observation with a FATE (the merge must dispose it): --kind "+record.MustEnum("observe", "kind").Spelling()+" [--label ...] --reason",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			kind := seat.Str(cmd, flags.Kind)
			if kind == "" {
				kind = "note"
			}
			p := seat.SetSame(cmd, record.NewPayload().Set("kind", kind), flags.Label)
			p.Set("text", text)
			ev, err := record.Append(s.RunDir, s.SeatID, "observe", p)
			if err != nil {
				return nil, err
			}
			return observeResult{FindingID: ev.Payload.Str("finding_id"), Label: seat.Str(cmd, flags.Label)}, nil
		}))

	c.Flags().String(flags.Kind, "", record.MustEnum("observe", "kind").Usage("an observation's flavour; the merge must still dispose it"))
	c.Flags().String(flags.Label, "", "a stable local label, so the merge can name this observation when disposing it")
	return c
}

type observeResult struct {
	FindingID string `json:"finding_id"`
	Label     string `json:"label"`
}

func (r observeResult) Human() string {
	return "observation recorded: " + r.FindingID + " (" + r.Label + ") — awaiting merge disposition, which names it by ID"
}
