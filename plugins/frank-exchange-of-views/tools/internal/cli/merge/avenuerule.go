package merge

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// avenue-rule: red's fate for a direction blue proposed.
//
// RED AUDITS AND RULES; IT NEVER PROPOSES. Directing research is what a gap's required_fix
// already does — "research X, enumerate Y" — and giving red a second way to say the same
// thing is exactly the aliasing internal/flags exists to prevent. So blue proposes the
// directions and red says which are worth the run's time.
//
// WHY IT EXISTS AT ALL. Measured over six runs: blue rejected 18 of its own 86 avenues, and
// red rejected NONE — because red had no verb to. Every rejection in the corpus is blue
// marking its own homework, and the steelman duty red does carry ("attack the REASONS for
// declining") can only argue about rejections blue already made. It cannot say "this one you
// are pursuing is not worth it".
//
// THE THREE FATES answer the two questions gb named — is it out of scope, or is it too thin:
//
//	endorsed      worth the run's time; pursue it
//	out-of-scope  a real question, but not THIS run's question
//	too-thin      in scope, but the hypothesis does not carry enough to be worth the budget
//
// A ruling is not a command. Blue may contest it through `blue dispute`, the same accounted
// channel it contests a grade through, and the burden of citations falls THERE rather than
// on every proposal — ~14 avenues a run makes up-front sourcing a fetch storm before any
// research has happened.
func newAvenueRule() *cobra.Command {
	c := seat.New("avenue-rule",
		`rule on a direction blue proposed: --id A1 --ruling `+record.MustEnum("avenue-rule", "ruling").Spelling()+` --reason "<why>"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			id := seat.Str(cmd, flags.ID)
			p := record.NewPayload().Set("avenue_id", id)
			seat.Set(cmd, p, "ruling", flags.Ruling)
			if err := seat.SetReason(cmd, p, "reason"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "avenue-rule", p); err != nil {
				return nil, err
			}
			return ruleAvenueResult{ID: id, Ruling: seat.Str(cmd, flags.Ruling)}, nil
		})

	c.Flags().String(flags.ID, "", "REQUIRED — the avenue id blue proposed (A1, A2 …)")
	c.Flags().String(flags.Ruling, "", record.MustEnum("avenue-rule", "ruling").Usage(
		"endorsed (worth the run's time) | out-of-scope (a real question, but not THIS run's) | too-thin (in scope, but the hypothesis does not carry its budget)"))
	return seat.Prose(c)
}

type ruleAvenueResult struct {
	ID     string `json:"avenue_id"`
	Ruling string `json:"ruling"`
}

func (r ruleAvenueResult) Human() string { return "avenue " + r.ID + " ruled " + r.Ruling }
