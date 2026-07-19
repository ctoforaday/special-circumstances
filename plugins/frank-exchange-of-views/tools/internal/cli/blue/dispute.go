package blue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dispute: contest a grade through the accounted channel.
//
// Blue may argue with red's grading — but through a recorded event with evidence
// attached, not by quietly repairing to a different standard. The channel has
// fired zero times across two full runs, which is itself the finding: blue
// contested nothing, and that is only visible because the channel exists.
func newDispute() *cobra.Command {
	var proposed flags.GradeValue

	c := seat.New(role, "dispute",
		`contest a grade through the accounted channel: --id <gap> --dimension severity|likelihood|impact|complexity_cost --proposed <grade> --basis "..."`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", "id")
			seat.SetSame(cmd, p, "dimension")
			seat.SetGrade(p, "proposed", &proposed)
			flags.SetEither(p, "evidence", cmd, flags.Basis, "evidence")
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("dispute filed on %s.%s", seat.Str(cmd, "id"), seat.Str(cmd, "dimension")), nil
		})

	c.Flags().String("id", "", "the gap id")
	c.Flags().String("dimension", "", "severity | likelihood | impact | complexity_cost — the axis you contest")
	c.Flags().Var(&proposed, "proposed", flags.GradeUsage("the grade you say it should be"))
	flags.AliasString(c, flags.Basis, "evidence", "why, citing the exact section")
	return c
}
