package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// retire: the ONLY way substance leaves the report.
//
// "Never subtract substance" was a PROSE-level rule, and it was doing two jobs at
// once. It stopped run 3's real failure — blue quietly dropping content under
// repair pressure — but it also forbade rewriting, so the report could only grow:
// 1178 to 1668 lines in a single run, and every audit seat paid to re-read all of
// it, every round.
//
// Splitting the two jobs: prose may now be compacted, merged and reorganized
// freely, because a claim can only LEAVE through this verb. Deletion stops being
// something a rule forbids and becomes something the record shows — with what was
// removed, why, and what (if anything) replaced it.
//
// That is strictly stronger than the prose rule it replaces. The old rule could
// be broken silently by an edit; this one leaves a hole in the record that
// capture detects, because claim_count falling further than the retire events
// account for is arithmetic, not judgement.
func newRetire() *cobra.Command {
	c := seat.Prose(seat.New("retire",
		`remove a claim from the report, on the record: --claim "<the claim, quoted>" --reason "..." [--superseded-by "<the claim that replaces it>"]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return nil, err
			}
			p := seat.SetSame(cmd, record.NewPayload(), flags.Claim, flags.Reason)
			seat.Set(cmd, p, "superseded_by", flags.SupersededBy)
			if text != "" {
				p.Set("detail", text)
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "retire", p); err != nil {
				return nil, err
			}
			return retireResult{Claim: seat.Str(cmd, flags.Claim)}, nil
		}))

	c.Flags().String(flags.Claim, "", "the claim being removed, quoted from the report as it stood")
	c.Flags().String(flags.Reason, "", "why it goes: refuted, superseded, merged, or out of scope — a removal with no stated reason is the failure this verb exists to make visible")
	c.Flags().String(flags.SupersededBy, "", "the claim that replaces it, when one does")
	return c
}

type retireResult struct {
	Claim string `json:"claim"`
}

func (r retireResult) Human() string { return "retired: " + r.Claim }
