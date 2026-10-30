package seat

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Regrade: same id, moved grade, stated reason.
//
// --basis is required by the domain rather than optional, because a grade that
// moves without a recorded reason is indistinguishable from a grade that drifted.
// E0.5b found exactly that: regrade history unrecoverable after the fact, so
// nobody could tell concession from re-scoping.
//
// TWO SEATS FILE IT, AND EACH PAGE TEACHES ITS OWN: the minting lens moves or keeps its gap's grade
// (`regrade`), and the bench sets the grade its docket ruling rules on (`regrade-bench`). One act and
// one write path (record.requireRegradeWithinBenchLock), keyed by the help it renders.
func Regrade(key string) *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue

	c := NewKeyed("regrade", key, regradeResult{}, func(s Context, cmd *cobra.Command) (Result, error) {
		body := &recordpb.Regrade{GapId: proto.String(Str(cmd, flags.ID))}
		body.Severity = GradeOrNil(&severity)
		body.Likelihood = GradeOrNil(&likelihood)
		body.Impact = GradeOrNil(&impact)
		body.ComplexityCost = GradeOrNil(&cx)
		// Flag word --reason (the one prose word), payload key stays basis.
		basis, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Basis = proto.String(basis)
		// A GRADE SET AGAINST A GAP THAT MOVED SINCE THIS SITTING OPENED grades what the gap was
		// (gblock ruling B): the refusal shows what changed, and the same regrade lands once read.
		if err := AppendFresh(s, body.GetGapId(), func(guards ...record.Guard) error {
			_, err := record.Append(s.Identity(), body, guards...)
			return err
		}); err != nil {
			return nil, err
		}
		return regradeResult{GapID: Str(cmd, flags.ID)}, nil
	})

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	// COBRA NAMES THE MISSING ID. Without this the only refusal was the record's foreign key —
	// `recording a regrade: constraint failed: FOREIGN KEY constraint failed` — which names no flag;
	// requireOriginator and requireGap both pass an empty id through on the note that the
	// missing-id refusal is the verb's own, and this verb had none.
	Require(c, flags.ID)
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
	c.Flags().Var(&likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
	c.Flags().Var(&impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
	c.Flags().Var(&cx, flags.Complexity, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY)+", on the same scale")
	return Prose(c)
}

type regradeResult struct {
	GapID string `json:"gap_id"`
}

func (r regradeResult) Human() string { return "regraded " + r.GapID }
