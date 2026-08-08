package blue

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/spf13/cobra"
	"strings"

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
			// --reason is the prose channel (Prose provides it); it lands under `reason`,
			// which validate requires — substance leaves the report only with its reason.
			p := seat.SetSame(cmd, record.NewPayload(), flags.Claim)
			if err := seat.SetReason(cmd, p, "reason"); err != nil {
				return nil, err
			}
			seat.Set(cmd, p, "superseded_by", flags.SupersededBy)

			// THE REMOVAL IS CHECKED, NOT TAKEN ON TRUST.
			//
			// This verb recorded whatever it was told. Nothing confirmed the claim had ever
			// been in the report, and nothing confirmed it had left — so "substance leaves only
			// through the retire verb", the rule this comment above calls "strictly stronger
			// than the prose rule it replaces", rested on the seat's word.
			//
			// A PHANTOM RETIRE IS WORSE THAN USELESS. The scorecard's additive-integrity
			// detector computes unrecorded_claim_loss as (drop in claim_count) MINUS (retire
			// events): a retirement of a claim that was never there subtracts from the
			// accounted side and CANCELS REAL LOSS, blinding the one detector built to catch
			// silent deletion.
			claim := seat.Str(cmd, flags.Claim)
			basis := record.RemovalAsserted
			if md, rerr := record.ReadBlueReport(s.RunDir); rerr == nil {
				if strings.Contains(string(md), claim) {
					return nil, feov.Errorf(feov.Conflict,
						"blue retire: %q is still in the report. Retiring is how a removal is EXPLAINED, not how it is performed — remove the text with `blue edit` first, then retire the claim to say why it went and what replaced it",
						claim)
				}
				// Absent now. Whether it was ever THERE is a different question, and the record
				// can answer it: a claim removed by a recorded edit appears in that edit's old
				// span. Absent from both is a retirement of something nobody can show existed.
				if record.ClaimAppearsInAnEdit(s.RunDir, claim) {
					basis = record.RemovalVerified
				}
			}
			p.Set("removal_basis", basis)

			if _, err := record.Append(s.RunDir, s.SeatID, "retire", p); err != nil {
				return nil, err
			}
			return retireResult{Claim: seat.Str(cmd, flags.Claim)}, nil
		}))

	c.Flags().String(flags.Claim, "", "the claim being removed, quoted from the report as it stood")
	c.Flags().String(flags.SupersededBy, "", "the claim that replaces it, when one does")
	return c
}

type retireResult struct {
	Claim string `json:"claim"`
}

func (r retireResult) Human() string { return "retired: " + r.Claim }
