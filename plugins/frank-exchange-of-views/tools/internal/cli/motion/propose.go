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

// A DOCKET MOTION IS A PROPOSED ACT. The filer files the regrade, close or supersede it wants on the
// gap, with that act's own flags; the gap's assignee may accept it (`motion docket accept`), else the
// bench accepts, rejects or remands it. So `file` under docket is a group, one verb per act, and the
// act's flags are the ones its own verb takes.

// proposeSays is what each proposal's page says, and the docket filing group's refusal of a bare call.
const proposeSays = "A DOCKET MOTION PROPOSES THE ACT YOU WANT ON A GAP: regrade its grades, close it, or supersede it with a corrected statement — each with that act's own flags, the gap named by --id and your argument in --reason. " +
	"Its assignee may accept it, and the act is then applied exactly as you filed it; otherwise the bench accepts it, rejects it — the gap stands as it is — or remands it. " +
	"A proposal is admitted once the gap has had one exchange, one at a time on a gap, never on your own gap (take the act yourself), and on a closed gap only as a supersede. " +
	"It is refused on a gap you filed your closing on in this sitting, which holds one closing per gap — file it in a later sitting — and on a gap a successor superseded, naming the successor: propose on the successor."

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
			enumhelp.Flag(c, flags.As, record.MustEnum("close", "closure_class"), "HOW the gap ended — the same words the bench's dispositions use. Not repaired_with_regression: a close naming the regression it caused is the assignee's own act")
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
	var st lens.Statement
	c.AddCommand(propose("supersede", "the corrected statement of the gap — the successor its assignee or the bench writes on accepting it",
		func(c *cobra.Command) {
			st.Register(c)
			// THE SUCCESSOR'S STATEMENT IS WHOLE AT THE FILING: the accept writes it as filed, so what a
			// mint cannot be written without is refused here, at the parse, with each flag's own usage.
			seat.Require(c, flags.Class, flags.Problem, flags.Fix, flags.Check, flags.CheckKind,
				flags.Severity, flags.Likelihood, flags.Impact)
			seat.Records(c, "mint")
			seat.Supplies(c, "gap_id", "the tool assigns the successor's id at filing, and names the motion's gap as its ancestor")
			seat.Supplies(c, "class_material", "stamped from the class's material default at the write path")
		},
		func(run record.Run, cmd *cobra.Command, gapID, reason string, m *recordpb.Motion) error {
			p, err := st.Fill(run, cmd, reason, "")
			if err != nil {
				return err
			}
			// THE SUCCESSOR'S ID IS ASSIGNED NOW, and the motion's gap is its ancestor: the accept
			// writes the stored mint as it is, so the successor is the gap's assignee's (its chain's).
			p.GapId = proto.String(record.NewID("gap"))
			p.Supersedes = proto.String(gapID)
			if n := seat.Str(cmd, flags.New); n != "" {
				p.FixNew = proto.String(n)
			}
			m.Act = &recordpb.Motion_Supersede{Supersede: p}
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
				_, err := record.Append(s.Identity(), m, append([]record.Guard{record.HeardInTime(subj)}, guards...)...)
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

// newAccept is the gap's assignee's answer to a proposed act on its gap: the act, applied exactly as
// filed, through the act's own write, after the ruling that says so — one transaction.
func newAccept() *cobra.Command {
	c := &cobra.Command{
		Use:   "accept",
		Short: "accept the act a docket motion proposes on your gap — the tool applies it as filed",
		Long: "accept the act a docket motion proposes on a gap you are the assignee of — the tool applies it exactly as filed.\n\n" +
			"ONE TRANSACTION: your answer to the motion, then the act — the regrade, the close, or the successor the proposal states, its id assigned at filing — written as your own act, through that act's own write and every check it meets. " +
			"A refusal of the act records nothing, and the motion stands for the bench. Read the proposal on your work list before you accept: a close is attested by the anchors your sitting read. " +
			"If you do not accept it, your closing on the gap is your argument, and the bench rules.",
		RunE: seat.HandlerRunE(func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			run, err := s.Run()
			if err != nil {
				return nil, err
			}
			id := seat.Str(cmd, flags.ID)
			if err := record.RequireSubjectMatches(run, recordpb.MotionSubject_MOTION_SUBJECT_DOCKET, id); err != nil {
				return nil, err
			}
			reason, err := prose(cmd, "why you accept the act as filed")
			if err != nil {
				return nil, err
			}
			ruling := record.Act{Body: &recordpb.MotionRule{
				MotionId: proto.String(id),
				Opinion:  proto.String(reason),
				Ruling:   &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{Accepted: proto.Bool(true)}},
			}}
			if err := applyAccepted(s, run, id, ruling); err != nil {
				return nil, err
			}
			return ruled{ID: id, Ruling: "accepted"}, nil
		}),
	}
	seat.Returns(c, ruled{})
	c.Flags().Var(flags.MotionID(), flags.ID, refHelp("docket"))
	seat.Require(c, flags.ID)
	seat.SaysRequired(seat.Prose(c), flags.Reason)
	seat.Records(c, "motion_rule")
	seat.FillsArm(c, "docket")
	return c
}

// applyAccepted writes ruling and then the act motion id proposes, through the act's own write, under
// s's identity: the assignee's accept, or the bench's accept of a regrade or a supersede. A motion that
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
		// The bench's regrade is the grade its ruling rules on, so it is written first.
		if s.Role == "bench" {
			_, err := record.AppendActs(s.Identity(), record.Act{Body: a}, ruling)
			return err
		}
		return seat.WriteRegrade(s, a, ruling)
	case *recordpb.Close:
		return lens.WriteClose(s, run, a, ruling)
	case *recordpb.Mint:
		fixNew := a.GetFixNew()
		a.FixNew, a.FixBasis = nil, proto.String("proposed")
		return lens.WriteMint(s, a, a.GetLocation(), fixNew, ruling)
	default:
		_, err := record.AppendActs(s.Identity(), ruling)
		return err
	}
}
