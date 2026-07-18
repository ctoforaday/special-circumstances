package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// regrade: same id, moved grade, stated reason.
//
// --basis is required by the domain rather than optional, because a grade that
// moves without a recorded reason is indistinguishable from a grade that drifted.
// E0.5b found exactly that: regrade history unrecoverable after the fact, so
// nobody could tell concession from re-scoping.
func newRegrade() *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue

	c := seat.New(role, "regrade",
		`same-id grade movement, recorded with its reason: --id R2-5 [--severity/--likelihood/--impact/--cx <grade>] --basis "..."`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", "id")
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetGrade(p, "complexity_cost", &cx)
			seat.SetSame(cmd, p, "basis")
			if _, err := record.Append(s.RunDir, s.SeatID, "regrade", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("regraded %s", seat.Str(cmd, "id")), nil
		})

	c.Flags().String("id", "", "the gap id")
	c.Flags().Var(&severity, "severity", flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, "likelihood", "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, "impact", "how bad the consequence is if it lands")
	c.Flags().Var(&cx, "cx", "complexity_cost — what fixing it costs, on the same scale")
	c.Flags().String("basis", "", "why the grade moved — grade movement is recorded with its reason")
	return c
}
