package seat

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
// THE LENS'S ACT: the raiser moves or keeps its gap's grade, or takes the author's proposed regrade
// (--accept M). The bench files no regrade; the
// tool writes the bench's regrade as the statement a ruling applies (a held regrade proposal's grades,
// or the restore at grade movement), through the same write path (record.requireRegradeWithinBenchLock).
func Regrade(key string) *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue
	accept := flags.ProposalID()

	c := NewKeyed("regrade", key, regradeResult{}, func(s Context, cmd *cobra.Command) (Result, error) {
		// Flag word --reason (the one prose word), payload key stays basis.
		basis, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		gapID := Str(cmd, flags.ID)
		var body *recordpb.Regrade
		if m := accept.String(); m != "" {
			// THE RAISER TAKES THE AUTHOR'S PROPOSAL BY ITS OWN ACT: the tool supplies the grades the
			// motion stores, so what lands is what was proposed, and the act carries accepts = M.
			// --reason is still the raiser's own argument.
			if cmd.Flags().Changed(flags.Severity) || cmd.Flags().Changed(flags.Likelihood) || cmd.Flags().Changed(flags.Impact) || cmd.Flags().Changed(flags.Complexity) {
				return nil, feov.Errorf(feov.Validation, "--accept %s supplies the grades motion %s proposes — passing a grade as well would be a second source for one fact. Drop the grades, or drop --accept and regrade by your own judgment", m, m)
			}
			stored, err := AcceptedAct(s, m, gapID)
			if err != nil {
				return nil, err
			}
			rg, ok := stored.(*recordpb.Regrade)
			if !ok {
				return nil, feov.Errorf(feov.Validation, "motion %s proposes a close, which `close --accept %s` takes. Nothing was recorded", m, m)
			}
			body = rg
			body.Accepts = proto.String(m)
			gapID = rg.GetGapId()
		} else {
			if err := RequireOwed(cmd); err != nil {
				return nil, err
			}
			body = &recordpb.Regrade{GapId: proto.String(gapID)}
			body.Severity = GradeOrNil(&severity)
			body.Likelihood = GradeOrNil(&likelihood)
			body.Impact = GradeOrNil(&impact)
			body.ComplexityCost = GradeOrNil(&cx)
			if g := Str(cmd, flags.Answers); g != "" {
				body.Answers = proto.String(g)
			}
		}
		body.Basis = proto.String(basis)
		// A GRADE SET AGAINST A GAP THAT MOVED SINCE THIS SITTING OPENED grades what the gap was
		// (gblock ruling B): the refusal shows what changed, and the same regrade lands once read.
		if err := WriteRegrade(s, body); err != nil {
			return nil, err
		}
		return regradeResult{GapID: gapID, Accepts: accept.String()}, nil
	})

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	// COBRA NAMES THE MISSING ID unless --accept names the proposal, which names its gap: the handler
	// refuses a regrade with neither.
	RequiredUnlessGiven(c, flags.ID, flags.Accept)
	c.Flags().Var(accept, flags.Accept, "take the author's proposed regrade on your gap EXACTLY as filed: the M- id its filing returned. The tool supplies the grades and the act answers the proposal; under a docketing it is your concession, ending the docketing and the proposal with no ruling. --reason is still your own argument")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap about this gap that your regrade answers — another lens's gap on your gap's grades; omit for a regrade that answers no gap")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
	c.Flags().Var(&likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
	c.Flags().Var(&impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
	c.Flags().Var(&cx, flags.Complexity, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY)+", on the same scale")
	return Prose(c)
}

// WriteRegrade is the regrade's write, whoever takes it: `regrade` itself, or the bench's ruling
// applying a proposed regrade, with before — the ruling — written ahead of it in the same
// transaction. The Fresh guard rides with the regrade, and its refusal records the Stale that moves
// the seat's read point (AppendFresh).
func WriteRegrade(s Context, body *recordpb.Regrade, before ...record.Act) error {
	return AppendFresh(s, body.GetGapId(), func(guards ...record.Guard) error {
		_, err := record.AppendActs(s.Identity(), append(before, record.Act{Body: body, Guards: guards})...)
		return err
	})
}

type regradeResult struct {
	GapID   string `json:"gap_id"`
	Accepts string `json:"accepts,omitempty"`
}

func (r regradeResult) Human() string {
	if r.Accepts != "" {
		return "regraded " + r.GapID + ", taking " + r.Accepts
	}
	return "regraded " + r.GapID
}
