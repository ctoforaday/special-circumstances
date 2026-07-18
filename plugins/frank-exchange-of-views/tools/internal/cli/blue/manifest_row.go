package blue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// manifest-row: the correctness manifest's receipt, one per repaired gap.
//
// The manifest is blue's self-audit, and this is what makes it auditable by
// anyone else: an unmanifested repair is unchecked by blue's OWN standard, which
// is a stronger thing to be able to say than "we think it was checked".
func newManifestRow() *cobra.Command {
	c := seat.Prose(seat.New(role, "manifest-row",
		`one correctness-manifest receipt per repaired gap: --id R2-3 --row "figures recomputed; acceptance check run: pass; sites swept: S2,S4"`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return "", err
			}
			row := seat.Str(cmd, "row")
			if row == "" {
				row = text
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", "id")
			p.Set("row", row)
			if _, err := record.Append(s.RunDir, s.SeatID, "manifest-row", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("manifest row recorded for %s", seat.Str(cmd, "id")), nil
		}))

	c.Flags().String("id", "", "the gap id this receipt covers")
	c.Flags().String("row", "", "what you checked and what it showed, compressed to one line")
	return c
}
