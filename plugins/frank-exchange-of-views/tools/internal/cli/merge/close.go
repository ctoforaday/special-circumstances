package merge

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// close: retire a gap, WITH the evidence that it is retired.
//
// The anchor triple (who verified, with what, against what) is required because
// E0.5a found unanchored closures unauditable after the fact — the record said a
// thing was checked and could not say by whom, how, or against what. --carried-
// from is the honest alternative: this closure is not a fresh act, it is last
// round's, restated.
func newClose() *cobra.Command {
	c := seat.New(role, "close",
		`close a gap WITH its verification anchor: --id R2-3 --as closed|closed_with_regression|... (--anchor-seat L1 --anchor-tool "git show" --anchor-target "7bc501e:path" | --carried-from <round>) [--successor R3-1] [--file f]`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			class := seat.Str(cmd, flags.As)
			if class == "" {
				class = "closed"
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			p.Set("closure_class", class)
			seat.Set(cmd, p, "anchor_seat", flags.AnchorSeat)
			seat.Set(cmd, p, "anchor_tool", flags.AnchorTool)
			seat.Set(cmd, p, "anchor_target", flags.AnchorTarget)
			seat.Set(cmd, p, "carried_from", flags.CarriedFrom)
			seat.SetSame(cmd, p, flags.Successor)
			f := seat.Str(cmd, flags.File)
			if f != "" {
				b, err := os.ReadFile(f)
				if err != nil {
					return "", err
				}
				p.Set("prose", string(b))
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "close", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("closed %s (%s)", seat.Str(cmd, flags.ID), class), nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.As, "", "closed | closed_with_regression | ... — the closure class")
	c.Flags().String(flags.AnchorSeat, "", "WHO verified the closure (the seat)")
	c.Flags().String(flags.AnchorTool, "", "WITH WHAT it was verified (the tool or command)")
	c.Flags().String(flags.AnchorTarget, "", "AGAINST WHAT — the exact file, ref or URL read")
	c.Flags().String(flags.CarriedFrom, "", "the round this closure was carried from, when it is not a fresh act")
	c.Flags().String(flags.Successor, "", "the gap id carrying the unresolved remainder forward")
	c.Flags().String(flags.File, "", "the full closure record, from a file")
	return c
}
