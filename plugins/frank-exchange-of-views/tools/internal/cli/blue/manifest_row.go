package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// manifest-row: the correctness manifest's receipt, one per repaired gap.
//
// The manifest is blue's self-audit, and this is what makes it auditable by
// anyone else: an unmanifested repair is unchecked by blue's OWN standard, which
// is a stronger thing to be able to say than "we think it was checked".
func newManifestRow() *cobra.Command {
	c := seat.Prose(seat.New("manifest-row",
		`one correctness-manifest receipt per repaired gap: --id R2-3 --row "figures recomputed; acceptance check run: pass; sites swept: S2,S4"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			row := seat.Str(cmd, flags.Row)
			if row == "" {
				row = text
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			p.Set("row", row)
			if _, err := record.Append(s.RunDir, s.SeatID, "manifest-row", p); err != nil {
				return nil, err
			}
			return manifestRowResult{GapID: seat.Str(cmd, flags.ID)}, nil
		}))

	c.Flags().String(flags.ID, "", "the gap id this receipt covers")
	c.Flags().String(flags.Row, "", "what you checked and what it showed, compressed to one line")
	return c
}

type manifestRowResult struct {
	GapID string `json:"gap_id"`
}

func (r manifestRowResult) Human() string { return "manifest row recorded for " + r.GapID }
