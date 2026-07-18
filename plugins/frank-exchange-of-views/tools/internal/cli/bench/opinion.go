package bench

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
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
	c := seat.Prose(seat.New(role, "opinion",
		`a ruling as an OPINION: --gap-id R3-2 --disposition carried|closed|... --principle "..." --tension "correctness vs economy" --review-flag "why a human should look" [--file rationale]`,
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return "", err
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", "gap-id")
			seat.SetSame(cmd, p, "disposition", "principle", "tension")
			seat.Set(cmd, p, "review_flag", "review-flag")
			p.Set("rationale", text)
			if _, err := record.Append(s.RunDir, s.SeatID, "opinion", p); err != nil {
				return "", err
			}
			return fmt.Sprintf("opinion recorded: %s %s", seat.Str(cmd, "gap-id"), seat.Str(cmd, "disposition")), nil
		}))

	c.Flags().String("gap-id", "", "the gap being ruled on")
	c.Flags().String("disposition", "", "carried | closed | rebuttal_sustained | risk_accepted | routed_to_infrastructure | ...")
	c.Flags().String("principle", "", "the principle applied — a ruling is an OPINION, not a disposition")
	c.Flags().String("tension", "", "the values in tension (e.g. correctness vs economy)")
	c.Flags().String("review-flag", "", "why a human should, or should not, look at this")
	return c
}
