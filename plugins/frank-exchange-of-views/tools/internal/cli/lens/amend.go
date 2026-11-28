package lens

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE AMEND (plans/findings-fold.md §III.B7, ruled 2026-11-26 "AMEND"): a lens restates its own open
// gap's wording — problem, required fix, acceptance check, prescription — as a new version on the
// gap's one id, which every reference keeps. Modelled on regrade at every layer: the record holds the
// minter's right and the gap's openness, stamps the resulting wording, and the Fresh guard rides with
// the write. The gap's class, subject, location and check kind are fixed at the mint, so a defect of
// another class or at another place is a new gap.
func newAmend() *cobra.Command {
	c := seat.NewKeyed("amend", "amend", amendResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		basis, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		gapID := seat.Str(cmd, flags.ID)
		// THE AUTHOR TAKES THE STATEMENT ON A GAP ABOUT ITS GAP (--accept G2): the tool writes what the
		// raiser states, whole, answering G2 (seat.GapStatementActs).
		if g2 := seat.Str(cmd, flags.Accept); g2 != "" {
			for _, f := range []string{flags.Problem, flags.RequiredFix, flags.AcceptanceCheck, flags.FixNew, flags.Answers} {
				if seat.Given(cmd, f) {
					return nil, feov.Errorf(feov.Validation, "--accept %s supplies the wording its raiser states and names the gap the amend answers — passing --%s as well would be a second source for one fact. Drop it, or drop --accept and amend by your own judgment", g2, f)
				}
			}
			acts, under, err := seat.GapStatementActs(s, g2, gapID, basis)
			if err != nil {
				return nil, err
			}
			if err := seat.WriteGapStatement(s, under, acts); err != nil {
				return nil, err
			}
			return amendResult{GapID: under, Accepts: g2}, nil
		}
		if err := seat.RequireOwed(cmd); err != nil {
			return nil, err
		}
		body := &recordpb.Amend{GapId: proto.String(gapID), Basis: proto.String(basis)}
		for _, f := range []struct {
			flag string
			to   **string
		}{{flags.Problem, &body.Problem}, {flags.RequiredFix, &body.RequiredFix}, {flags.AcceptanceCheck, &body.AcceptanceCheck}} {
			if seat.Given(cmd, f.flag) {
				*f.to = proto.String(seat.Str(cmd, f.flag))
			}
		}
		if g := seat.Str(cmd, flags.Answers); g != "" {
			body.Answers = proto.String(g)
		}
		var guards []record.Guard
		if seat.Given(cmd, flags.FixNew) {
			fixNew := seat.Str(cmd, flags.FixNew)
			// fix_basis is derived, as the mint's is: verified, because the guard below refuses the
			// whole write where blue's accept would not apply the pair.
			body.FixNew, body.FixBasis = proto.String(fixNew), proto.String("verified")
			guards = append(guards, verifiedAmendFix(s.SeatID, gapID, fixNew))
		}
		if err := writeAmend(s, body, guards...); err != nil {
			return nil, err
		}
		return amendResult{GapID: gapID}, nil
	})
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "your open gap whose wording this restates")
	seat.RequiredUnlessGiven(c, flags.ID, flags.Accept)
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Accept, "take the statement another lens's gap about your gap states in its raiser's closing: that gap's G- id. The tool writes the wording — and the grades, where it states them — on your gap exactly as stated, answering it, so you pass no wording, --id or --answers; under its docketing it is your concession, ending the docketing with no ruling. --reason is still your own argument")
	flags.Text(c, flags.Problem, "the gap's problem as it now stands, in the mint's sense; omit to keep the current one")
	flags.Text(c, flags.RequiredFix, "the gap's required fix as it now stands, in the mint's sense; omit to keep the current one")
	flags.Text(c, flags.AcceptanceCheck, "the gap's acceptance check as it now stands, in the mint's sense; omit to keep the current one")
	flags.Text(c, flags.FixNew, "the gap's prescription as it now stands: what its location, fixed at the mint, becomes — verified as a mint's replacement text is, against the report the amend lands on; omit to keep the current one")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap about this gap that your amend answers — another lens's gap on your gap's wording; omit for an amend that answers no gap")
	return seat.Prose(c)
}

// verifiedAmendFix is an amend's --fix-new checked as the mint's --new is, on the report the amend
// lands on, with the gap's location — fixed at its mint — as the quote: blue's accept must apply it
// (blue.PlanAccept on the location at the gap's token), and the prescribing lens's estoppel bar meets
// it, so no prescription over text the lens's own applied prescription wrote lands.
func verifiedAmendFix(seatID, gapID, fixNew string) record.Guard {
	return record.Check(func(l *record.Landing) error {
		loc, _, _, err := l.Prescription(gapID)
		if err != nil {
			return err
		}
		if err := requireNotEstopped(l, seatID, loc); err != nil {
			return err
		}
		if err := bluedoc.RefuseAuthoring(loc, fixNew); err != nil {
			return err
		}
		report, err := l.Report()
		if err != nil {
			return err
		}
		if _, err := blue.PlanAccept(report, gapID, loc, fixNew); err != nil {
			return feov.Wrap(feov.Validation, err,
				"--fix-new is no prescription blue's accept would apply to the report as it stands, over %s's location as its mint fixed it: %v. A gap's prescription is amendable only while its location stands in the report, word for word, ending at the gap's marker; the location itself is fixed at the mint, so a defect at another place is a new gap. Nothing was recorded", gapID, err)
		}
		return nil
	})
}

// writeAmend is the amend's write: the Fresh guard rides with it, and its refusal records the Stale
// that moves the seat's read point (seat.AppendFresh), as a regrade's does.
func writeAmend(s seat.Context, body *recordpb.Amend, guards ...record.Guard) error {
	return seat.AppendFresh(s, body.GetGapId(), func(fresh ...record.Guard) error {
		_, err := record.AppendActs(s.Identity(), record.Act{Body: body, Guards: append(fresh, guards...)})
		return err
	})
}

type amendResult struct {
	GapID   string `json:"gap_id"`
	Accepts string `json:"accepts,omitempty"`
}

func (r amendResult) Human() string {
	if r.Accepts != "" {
		return "amended " + r.GapID + ", taking " + r.Accepts + "'s statement"
	}
	return "amended " + r.GapID
}
