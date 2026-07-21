package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// finding: a lens-scoped observation graded for the merge to dispose.
//
// The label is LENS-scoped (L3-F1) on purpose. Stable R<round>-N ids are the
// merge's to assign, and letting lenses mint them produced four different
// "R5-1"s in one round of run 3.
func newFinding() *cobra.Command {
	var severity, likelihood, impact flags.GradeValue

	c := seat.Prose(seat.New("finding",
		`a lens-scoped finding: --label L3-F1 --severity <g> --likelihood <g> --impact <g> [--reason "..."] [--location "..."]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			p := seat.SetSame(cmd, record.NewPayload(), flags.Label)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetSame(cmd, p, flags.Location)
			p.Set("text", text)
			ev, err := record.Append(s.RunDir, s.SeatID, "finding", p)
			if err != nil {
				return nil, err
			}
			// The ID IS THE ANSWER, so it leads. A seat that is told only "recorded" has
			// to invent a way to refer to this later, and inventing is what produced
			// nine disposals of findings that were never recorded.
			return findingResult{FindingID: ev.Payload.Str("finding_id"), Label: seat.Str(cmd, flags.Label)}, nil
		}))

	c.Flags().String(flags.Label, "", "your lens-scoped finding label (L3-F1). Stable R<round>-N ids are the merge's to assign, never a lens's")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Location, "", "where the defect lives: a section heading plus a quoted sentence")
	return c
}

type findingResult struct {
	FindingID string `json:"finding_id"`
	Label     string `json:"label"`
}

func (r findingResult) Human() string {
	return "finding recorded: " + r.FindingID + " (" + r.Label + ") — the merge disposes it by ID, not by label"
}
