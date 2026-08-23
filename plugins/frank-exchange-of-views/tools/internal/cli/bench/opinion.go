package bench

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// opinion: a ruling that has to BE an opinion.
//
// principle, tension and review-flag are required by the domain, and the refusal
// says "opinions, not dispositions" because that is the failure being prevented.
// Across runs 4 and 5 the bench ruled `carried` on 64 of 65 items — a router, not
// a judge — and a disposition with no stated principle is indistinguishable from
// a default. Requiring the reasoning is what makes the difference visible.
func newOpinion() *cobra.Command {
	c := seat.Prose(seat.New("opinion", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		p := record.NewPayload()
		flags.Set(p, "gap_id", cmd, flags.ID)
		flags.Set(p, "disposition", cmd, flags.As)
		seat.SetSame(cmd, p, flags.Principle, flags.Tension)
		seat.Set(cmd, p, "review_flag", flags.ReviewFlag)
		seat.Set(cmd, p, "settled", flags.Settled)
		// ONE QUESTION, TWO ASSERTABLE ANSWERS — the `friction --none` shape.
		//
		// "What would reopen this?" has a real answer of "nothing", and that answer must be
		// POSITIVE. Left as an empty --reopens-on it would be indistinguishable from a bench
		// that skipped the field, which is the same defect the friction channel had for
		// eighteen consecutive sittings until --none made the empty case sayable.
		//
		// Refused together rather than silently preferring one: a ruling that says both "this
		// is final" and "here is what would undo it" has not decided, and picking for it would
		// record a decision nobody made.
		final, _ := cmd.Flags().GetBool(flags.Final)
		if final && seat.Given(cmd, flags.ReopensOn) {
			return nil, feov.Errorf(feov.Validation,
				"opinion: --final says nothing would reopen this and --reopens-on names what would. "+
					"They are opposite answers to one question; pass exactly one")
		}
		if final {
			p.Set("final", true)
		} else {
			seat.Set(cmd, p, "reopens_on", flags.ReopensOn)
		}
		p.Set("reason", text)
		if _, err := record.Append(s.Identity(), "opinion", p); err != nil {
			return nil, err
		}
		return opinionResult{ID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
	}))

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap being ruled on")
	enumhelp.Flag(c, flags.As, record.MustEnum("opinion", "disposition"), ("your ruling AND the gap's fate. Every value ends the gap except `carried`, which defers it to a later round with a stated direction. One vocabulary with red's closure classes since #342"))
	c.Flags().String(flags.Principle, "", "the principle applied — a ruling is an OPINION, not a disposition")
	c.Flags().String(flags.Tension, "", "the values in tension (e.g. correctness vs economy)")
	c.Flags().String(flags.ReviewFlag, "", "why a human should, or should not, look at this")
	c.Flags().String(flags.Settled, "",
		"THE PROPOSITION NOW BARRED, as one sentence — not the gap id, and not the disposition. "+
			"A finding is a claim, its evidence and its demand; a fate says which of the three fell only "+
			"to whoever wrote it. Say what the losing party may no longer assert")
	c.Flags().String(flags.ReopensOn, "",
		"what would change this outcome — the evidence or condition that would make it worth raising again. "+
			"If nothing would, pass --final instead: that is a different answer, not a missing one")
	c.Flags().Bool(flags.Final, false,
		"nothing would reopen this: settled on the merits, and more evidence cannot change it. "+
			"The assertable form of an empty --reopens-on, so a decided question is distinguishable from a skipped field")
	return c
}

type opinionResult struct {
	ID string `json:"id"`
	As string `json:"as"`
}

func (r opinionResult) Human() string { return "opinion recorded: " + r.ID + " " + r.As }
