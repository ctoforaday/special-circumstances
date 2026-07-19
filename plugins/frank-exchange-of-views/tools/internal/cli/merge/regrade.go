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
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetGrade(p, "complexity_cost", &cx)
			if err := seat.SetLongForm(cmd, p, "basis", flags.Basis); err != nil {
				return "", err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "regrade", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("regraded %s", seat.Str(cmd, flags.ID)), nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().Var(&cx, flags.Complexity, "complexity_cost — what fixing it costs, on the same scale")
	c.Flags().String(flags.Basis, "", "why the grade moved — grade movement is recorded with its reason")
	return seat.Prose(c)
}
