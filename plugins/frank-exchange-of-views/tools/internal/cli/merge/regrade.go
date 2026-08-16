package merge

import (
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

	c := seat.New("regrade",
		`same-id grade movement, recorded with its reason: --id R2-5 [--severity/--likelihood/--impact/--cx <grade>] --reason "..."`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetGrade(p, "complexity_cost", &cx)
			// Flag word --reason (the one prose word), payload key stays basis.
			if err := seat.SetReason(cmd, p, "basis"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.Identity(), "regrade", p); err != nil {
				return nil, err
			}
			return regradeResult{GapID: seat.Str(cmd, flags.ID)}, nil
		})

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what one grade meant before v2 split them")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().Var(&cx, flags.Complexity, "complexity_cost — what fixing it costs, on the same scale")
	return seat.Prose(c)
}

type regradeResult struct {
	GapID string `json:"gap_id"`
}

func (r regradeResult) Human() string { return "regraded " + r.GapID }
