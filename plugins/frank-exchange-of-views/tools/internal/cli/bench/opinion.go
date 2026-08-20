package bench

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// opinion: a ruling that has to BE an opinion.
//
// principle, tension and review-flag are required by the domain, and the refusal
// says "opinions, not dispositions" because that is the failure being prevented.
// Across runs 4 and 5 the bench ruled `carried` on 64 of 65 items — a router, not
// a judge — and a disposition with no stated principle is indistinguishable from
// a default. Requiring the reasoning is what makes the difference visible.
func newOpinion() *cobra.Command {
	c := seat.Prose(seat.New("opinion",
		`a ruling as an OPINION: --id R3-2 --as carried|closed|... --principle "..." --tension "correctness vs economy" --review-flag "why a human should look" --reason <rationale>`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			body := &recordpb.Opinion{
				GapId:       proto.String(seat.Str(cmd, flags.ID)),
				Disposition: proto.String(seat.Str(cmd, flags.As)),
				Principle:   proto.String(seat.Str(cmd, flags.Principle)),
				Tension:     proto.String(seat.Str(cmd, flags.Tension)),
				ReviewFlag:  proto.String(seat.Str(cmd, flags.ReviewFlag)),
				Rationale:   proto.String(text),
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return nil, err
			}
			return opinionResult{ID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
		}))

	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap being ruled on")
	enumhelp.Flag(c, flags.As, record.MustEnum("opinion", "disposition"), ("REQUIRED — your ruling AND the gap's fate. Every value ends the gap except `carried`, which defers it to a later round with a stated direction. One vocabulary with red's closure classes since #342"))
	c.Flags().String(flags.Principle, "", "the principle applied — a ruling is an OPINION, not a disposition")
	c.Flags().String(flags.Tension, "", "the values in tension (e.g. correctness vs economy)")
	c.Flags().String(flags.ReviewFlag, "", "why a human should, or should not, look at this")
	return c
}

type opinionResult struct {
	ID string `json:"id"`
	As string `json:"as"`
}

func (r opinionResult) Human() string { return "opinion recorded: " + r.ID + " " + r.As }
