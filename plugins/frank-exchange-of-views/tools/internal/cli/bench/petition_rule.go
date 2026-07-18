package bench

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// petition-rule: grant or deny a petition.
//
// A halt is deliberately NOT a value of --as. Ending the run is a different kind
// of act from granting relief, and giving it its own verb means it can never be
// reached by a typo in an enum.
func newPetitionRule() *cobra.Command {
	c := seat.Prose(seat.New(role, "petition-rule",
		"rule on a petition: --petitioner <seatId> --class <class> --as granted|denied --file <opinion> (a halt is its own verb)",
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return "", err
			}
			p := seat.SetSame(cmd, record.NewPayload(), "petitioner", "class")
			seat.Set(cmd, p, "ruling", "as")
			p.Set("opinion", text)
			if _, err := record.Append(s.RunDir, s.SeatID, "petition-rule", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("petition %s (%s)", seat.Str(cmd, "as"), seat.Str(cmd, "petitioner")), nil
		}))

	c.Flags().String("petitioner", "", "the seat that filed the petition")
	c.Flags().String("class", "", "ethical | safety | integrity | constitutional")
	c.Flags().String("as", "", "granted | denied — a halt is its own verb")
	return c
}
