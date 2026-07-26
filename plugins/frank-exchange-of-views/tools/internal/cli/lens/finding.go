package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// finding: a lens's graded observation, for the merge to dispose.
//
// The label is TOOL-assigned — L{role}-F{N}, run-unique per role, the role read
// from the seat id. A lens no longer invents it: hand-numbered labels collided
// (four L5-F1s in one round of run 3), and the label is now the identity a gap's
// found_by names, so it must be unambiguous run-wide. The lens passes a stable
// local --key (its own F1/F2) purely as a crash-retry handle; a retry returns the
// existing label rather than minting a duplicate (the mint --key pattern).
func newFinding() *cobra.Command {
	var severity, likelihood, impact flags.GradeValue

	c := seat.Prose(seat.New("finding",
		`a lens finding (the tool assigns the L{role}-F{N} label): --key <your local F1> --severity <g> --likelihood <g> --impact <g> [--reason "..."] [--location "..."]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			// Crash-retry idempotency: a prior finding under this --key returns its
			// label, no second event (mirrors ExistingMintByKey, BEFORE Append).
			key := seat.Str(cmd, flags.Key)
			if prior, err := record.ExistingFindingByKey(s.RunDir, s.SeatID, key); err != nil {
				return nil, err
			} else if prior != "" {
				return findingResult{Label: prior, Idempotent: true}, nil
			}
			label, err := record.NextFindingLabel(s.RunDir, s.SeatID)
			if err != nil {
				return nil, err
			}

			p := record.NewPayload().Set("label", label)
			seat.Set(cmd, p, "finding_key", flags.Key)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetSame(cmd, p, flags.Location)
			p.Set("text", text)
			ev, err := record.Append(s.RunDir, s.SeatID, "finding", p)
			if err != nil {
				return nil, err
			}
			// The LABEL leads: it is the run-unique identity a gap's found_by names.
			return findingResult{Label: label, FindingID: ev.Payload.Str("finding_id")}, nil
		}))

	c.Flags().String(flags.Key, "", "a stable local handle (your own F1, F2 …) making a retried finding idempotent; the TOOL assigns the run-unique label L{role}-F{N}")
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().String(flags.Location, "", "where the defect lives: a section heading plus a quoted sentence")
	return c
}

type findingResult struct {
	Label      string `json:"label"`
	FindingID  string `json:"finding_id,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r findingResult) Human() string {
	if r.Idempotent {
		return "finding " + r.Label + " (idempotent retry — existing label returned)"
	}
	return "finding recorded: " + r.Label + " — the run-unique label a gap's found_by names (id " + r.FindingID + ")"
}
