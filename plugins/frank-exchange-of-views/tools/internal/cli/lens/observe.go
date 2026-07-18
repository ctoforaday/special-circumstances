package lens

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// observe: a below-bar finding that still demands a FATE.
//
// The merge must dispose every observation. That is why this is a recorded event
// rather than prose in a candidate file: an observation nobody disposed shows in
// the render's undisposed footer, where a note that quietly evaporated would not.
func newObserve() *cobra.Command {
	c := seat.Prose(seat.New(role, "observe",
		"a below-bar observation with a FATE (the merge must dispose it): --kind note|checked-held [--label ...] --text|--file",
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return "", err
			}
			kind := seat.Str(cmd, "kind")
			if kind == "" {
				kind = "note"
			}
			p := seat.SetSame(cmd, record.NewPayload().Set("kind", kind), "label")
			p.Set("text", text)
			ev, err := record.Append(s.RunDir, s.SeatID, "observe", p)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("observation recorded (%s) — awaiting merge disposition", ev.Key), nil
		}))

	c.Flags().String("kind", "", "note | checked-held — an observation's flavour; the merge must still dispose it")
	c.Flags().String("label", "", "a stable local label, so the merge can name this observation when disposing it")
	return c
}
