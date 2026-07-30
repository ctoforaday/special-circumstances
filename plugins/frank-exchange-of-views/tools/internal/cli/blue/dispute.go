package blue

import (
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

	c := seat.New("dispute",
		`contest a grade through the accounted channel: --id <gap> --dimension `+record.MustEnum("dispute", "dimension").Spelling()+` --proposed <grade> --reason "..."`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			seat.SetSame(cmd, p, flags.Dimension)
			seat.SetGrade(p, "proposed", &proposed)
			// Flag word --reason (the one prose word), payload key stays evidence.
			reason, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			if reason != "" {
				p.Set("evidence", reason)
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute", p); err != nil {
				return nil, err
			}
			return disputeResult{GapID: seat.Str(cmd, flags.ID), Dimension: seat.Str(cmd, flags.Dimension)}, nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.Dimension, "", record.MustEnum("dispute", "dimension").Usage("the axis you contest"))
	c.Flags().Var(&proposed, flags.Proposed, flags.GradeUsage("the grade you say it should be"))
	return seat.Prose(c)
}

type disputeResult struct {
	GapID     string `json:"gap_id"`
	Dimension string `json:"dimension"`
}

func (r disputeResult) Human() string { return "dispute filed on " + r.GapID + "." + r.Dimension }
