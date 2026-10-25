package lens

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// regrade: same id, moved grade, stated reason.
//
// --basis is required by the domain rather than optional, because a grade that
// moves without a recorded reason is indistinguishable from a grade that drifted.
// E0.5b found exactly that: regrade history unrecoverable after the fact, so
// nobody could tell concession from re-scoping.
func newRegrade() *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue

	c := seat.New("regrade", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		body := &recordpb.Regrade{GapId: proto.String(seat.Str(cmd, flags.ID))}
		body.Severity = seat.GradeOrNil(&severity)
		body.Likelihood = seat.GradeOrNil(&likelihood)
		body.Impact = seat.GradeOrNil(&impact)
		body.ComplexityCost = seat.GradeOrNil(&cx)
		// Flag word --reason (the one prose word), payload key stays basis.
		basis, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Basis = proto.String(basis)
		// A GRADE SET AGAINST A GAP THAT MOVED SINCE THIS SITTING OPENED grades what the gap was
		// (gblock ruling B): the refusal shows what changed, and the same regrade lands once read.
		if err := seat.AppendFresh(s, body.GetGapId(), func(guards ...record.Guard) error {
			_, err := record.Append(s.Identity(), body, guards...)
			return err
		}); err != nil {
			return nil, err
		}
		return regradeResult{GapID: seat.Str(cmd, flags.ID)}, nil
	})

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	// COBRA NAMES THE MISSING ID. Without this the only refusal was the record's foreign key —
	// `recording a regrade: constraint failed: FOREIGN KEY constraint failed` — which names no flag;
	// requireOriginator and requireGap both pass an empty id through on the note that the
	// missing-id refusal is the verb's own, and this verb had none.
	seat.Require(c, flags.ID)
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
	c.Flags().Var(&likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
	c.Flags().Var(&impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
	c.Flags().Var(&cx, flags.Complexity, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY)+", on the same scale")
	return seat.Prose(c)
}

type regradeResult struct {
	GapID string `json:"gap_id"`
}

func (r regradeResult) Human() string { return "regraded " + r.GapID }
