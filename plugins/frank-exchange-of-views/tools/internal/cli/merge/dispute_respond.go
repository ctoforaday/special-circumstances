package merge

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dispute-respond: red's answer when blue contests a grade.
//
// The channel exists so disagreement is accounted rather than silent. Across two
// full runs it fired ZERO times — which is a finding about the channel, not
// evidence that nobody ever disagreed, and it is why the answer is recorded as an
// event rather than settled in prose.
func newDisputeRespond() *cobra.Command {
	c := seat.New("dispute-respond",
		`red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --reason "..."`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			// THE ANSWER NAMES THE DIMENSION IT ANSWERS.
			//
			// A dispute is filed against ONE grade — `blue dispute --id R1-1 --dimension
			// severity` — and blue may contest more than one on the same gap. The reply
			// recorded only the gap, so two disputes and one "rejected" left a record nobody
			// could resolve: which grade was refused? Demonstrated on a live run — severity and
			// impact both disputed, one answer naming neither.
			//
			// The ORCHESTRATOR always had it: debate.js matches red's answer to blue's dispute
			// on (gap_id, dimension) and reads the gap's grade at that dimension. It came off
			// the envelope, which evaporates with the session, while the durable record kept
			// the lossier half. The ephemeral channel was more precise than the permanent one.
			dim := seat.Str(cmd, flags.Dimension)
			if dim == "" {
				return nil, feov.Errorf(feov.MissingField,
					"dispute-respond: --dimension is required — blue disputes a single grade, and an answer that names only the gap cannot be matched to the grade it refuses when more than one is contested")
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			p.Set("dimension", dim)
			seat.Set(cmd, p, "response", flags.As)
			// Flag word --reason (the one prose word), payload key stays rationale.
			if err := seat.SetReason(cmd, p, "rationale"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "dispute-respond", p); err != nil {
				return nil, err
			}
			return disputeRespondResult{GapID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.Dimension, "", "REQUIRED — the grade dimension blue disputed, so the answer names what it refuses")
	c.Flags().String(flags.As, "", record.MustEnum("dispute-respond", "response").Usage("your answer to blue's grade dispute"))
	return seat.Prose(c)
}

type disputeRespondResult struct {
	GapID string `json:"gap_id"`
	As    string `json:"as"`
}

func (r disputeRespondResult) Human() string { return "dispute on " + r.GapID + ": " + r.As }
