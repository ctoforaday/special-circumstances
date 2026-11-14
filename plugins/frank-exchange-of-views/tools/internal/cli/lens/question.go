package lens

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// question check: the logic lens's judgement that a question part's latest coverage stands as the
// report's account of that part of the question (ruled by gblock 2026-11-11, E). A positive judgement
// only, as `inquiry check` is: a part whose coverage the lens finds misstated gets a gap of class
// coverage_misstated about it, and a part of the question no part covers gets one about the question.
// The record refuses the check of a part with no coverage, a wanting part, one whose coverage or gap
// landed after the seat's sitting opened, one already checked, and any seat but the owner's
// (record/questionpart.go).
func newQuestion() *cobra.Command {
	c := &cobra.Command{
		Use:          "question",
		Short:        "judge a question part's coverage — `check` records that it stands as the report's account of that part of the question",
		SilenceUsage: true,
	}
	c.AddCommand(newQuestionCheck())
	return c
}

func newQuestionCheck() *cobra.Command {
	c := seat.Records(seat.NewKeyed("check", "question-check", questionCheckResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		id := seat.Str(cmd, flags.ID)
		if _, err := record.Append(s.Identity(), &recordpb.QuestionPartCheck{PartId: proto.String(id)}); err != nil {
			return nil, err
		}
		return questionCheckResult{ID: id}, nil
	}), "question_part_check")
	c.Flags().Var(flags.QuestionPartID().WithCheck(record.QuestionPartExists), flags.ID, "the question part you judge — its `S-` id, which `show question` lists with its text, its coverage and the reason for it")
	seat.Require(c, flags.ID)
	return c
}

type questionCheckResult struct {
	ID string `json:"part_id"`
}

func (r questionCheckResult) Human() string { return "question part " + r.ID + " checked" }
