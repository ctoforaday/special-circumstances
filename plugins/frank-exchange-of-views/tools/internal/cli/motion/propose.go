package motion

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A DOCKET MOTION IS A PROPOSED ACT. The filer files the regrade or close it wants on the
// gap, with that act's own flags; the gap's assignee may accept it, else the
// dispatch dockets it and the bench sides with a party or remands it. So `file` under docket is a group, one verb per act, and the
// act's flags are the ones its own verb takes.

// proposeSays is what each proposal's page says, and the docket filing group's refusal of a bare call.
const proposeSays = "A DOCKET MOTION PROPOSES THE ACT YOU WANT ON A GAP: regrade its grades or close it — each with that act's own flags, the gap named by --id and your argument in --reason. " +
	"Its assignee may accept it, and the act is then applied exactly as you filed it; otherwise the bench accepts it, rejects it — the gap stands as it is — or remands it. " +
	"A proposal is admitted once the gap has had one exchange, one at a time on a gap, never on your own gap (take the act yourself), and never on a closed gap. " +
	"It is refused on a gap you filed your closing on in this sitting, which holds one closing per gap — file it in a later sitting."

// newDocketFile is the docket subject's `file`: one verb per proposed act.
func newDocketFile() *cobra.Command {
	c := &cobra.Command{
		Use:   "file",
		Short: "propose the act you want on a gap — " + strings.Join(record.ProposedActs, " | ") + " — the tool assigns the motion's id",
		// NO PAGE OF ITS OWN: each act's page carries the proposal's rule, and the group's would
		// only repeat it in the manual (leftOutOfManual).
		SilenceUsage: true,
		// The act is a subcommand; a bare `file` is refused naming them, not parsed for flags.
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE: func(cmd *cobra.Command, args []string) error {
			return feov.Errorf(feov.Validation,
				"a docket motion names the act it proposes — %s — as its verb: `motion docket file regrade --id <gap>`, with that act's own flags. %s",
				strings.Join(record.ProposedActs, ", "), proposeSays)
		},
	}
	var grades struct{ severity, likelihood, impact, cx flags.GradeValue }
	c.AddCommand(propose("regrade", "the grades you hold the gap should carry",
		func(c *cobra.Command) {
			c.Flags().Var(&grades.severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
			c.Flags().Var(&grades.likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
			c.Flags().Var(&grades.impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
			c.Flags().Var(&grades.cx, flags.Complexity, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY)+", on the same scale")
			seat.Records(c, "regrade")
		},
		func(_ record.Run, cmd *cobra.Command, gapID, reason string, m *recordpb.Motion) error {
			m.Act = &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{
				GapId:          proto.String(gapID),
				Severity:       seat.GradeOrNil(&grades.severity),
				Likelihood:     seat.GradeOrNil(&grades.likelihood),
				Impact:         seat.GradeOrNil(&grades.impact),
				ComplexityCost: seat.GradeOrNil(&grades.cx),
				Basis:          proto.String(reason),
			}}
			return nil
		}))
	var ev lens.CloseEvidence
	c.AddCommand(propose("close", "how you hold the gap ended, verified as a close is",
		func(c *cobra.Command) {
			enumhelp.Flag(c, flags.As, record.MustEnum("close", "closure_class"), "HOW you hold the gap ended — one vocabulary with every closure on the record")
			seat.Require(c, flags.As)
			ev.Register(c)
			seat.Records(c, "close")
		},
		func(_ record.Run, cmd *cobra.Command, gapID, reason string, m *recordpb.Motion) error {
			class, ok := record.DispositionOf(seat.Str(cmd, flags.As))
			if !ok {
				return feov.Errorf(feov.Validation, "%q is not a disposition", seat.Str(cmd, flags.As))
			}
			cl := &recordpb.Close{GapId: proto.String(gapID), ClosureClass: &class, Prose: proto.String(reason)}
			ev.Fill(cmd, cl)
			m.Act = &recordpb.Motion_Close{Close: cl}
			return nil
		}))
	return c
}

// propose builds one act's filing verb: register mounts the act's flags, and fill sets the act on the
// motion from them.
func propose(act, says string, register func(*cobra.Command), fill func(run record.Run, cmd *cobra.Command, gapID, reason string, m *recordpb.Motion) error) *cobra.Command {
	c := &cobra.Command{
		Use:   act,
		Short: "propose a " + act + " of the gap — " + says,
		Long:  "propose a " + act + " of the gap — " + says + ".\n\n" + proposeSays + "\n\n" + finalSittingSays("docket") + "\n\n" + seat.StaleRule,
		RunE: seat.HandlerRunE(func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			run, err := s.Run()
			if err != nil {
				return nil, err
			}
			reason, err := prose(cmd, "your argument for the act you propose — what its assignee or the bench rules on")
			if err != nil {
				return nil, err
			}
			gapID := seat.Str(cmd, flags.ID)
			id := record.NewID("motion")
			m := &recordpb.Motion{
				MotionId: proto.String(id),
				Basis:    proto.String(reason),
				Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}},
			}
			if err := fill(run, cmd, gapID, reason, m); err != nil {
				return nil, err
			}
			subj := recordpb.MotionSubject_MOTION_SUBJECT_DOCKET
			// A PROPOSAL ON A GAP THAT MOVED SINCE THIS SITTING OPENED argues with what the gap was
			// (gblock ruling B): the refusal shows what changed, and the same filing lands once read.
			if err := seat.AppendFresh(s, gapID, func(guards ...record.Guard) error {
				guards = append([]record.Guard{record.HeardInTime(subj)}, guards...)
				_, err := record.Append(s.Identity(), m, guards...)
				return err
			}); err != nil {
				return nil, err
			}
			return filed{ID: id, Subject: "docket", Act: act}, nil
		}),
	}
	seat.Returns(c, filed{})
	c.Flags().Var(flags.GapID(), flags.ID, fileFlagHelp[flags.ID])
	seat.Require(c, flags.ID)
	seat.SaysRequired(seat.Prose(c), flags.Reason)
	register(c)
	seat.Records(c, "motion")
	seat.FillsArm(c, "docket")
	seat.Supplies(c, "motion_id", "the tool assigns it on filing; a ruling or an accept names it with --id")
	return c
}

// applyAccepted writes ruling and then the act motion id proposes, through the act's own write, under
// s's identity: the bench's application of the author's proposal. A motion that
// proposes no act writes the ruling alone, for the record to refuse in its own words.
func applyAccepted(s seat.Context, run record.Run, id string, ruling record.Act) error {
	p, ok, err := record.DocketProposalOf(run, id)
	if err != nil {
		return err
	}
	if !ok {
		return feov.Errorf(feov.Validation, "no docket motion %s is on the record", id)
	}
	switch a := record.StoredAct(p.Motion).(type) {
	case *recordpb.Regrade:
		// On the bench's side the stored regrade is written as the bench's, before the ruling that applies it.
		if s.Role == "bench" {
			_, err := record.AppendActs(s.Identity(), record.Act{Body: a}, ruling)
			return err
		}
		return seat.WriteRegrade(s, a, ruling)
	case *recordpb.Close:
		return lens.WriteClose(s, run, a, ruling)
	default:
		_, err := record.AppendActs(s.Identity(), ruling)
		return err
	}
}
