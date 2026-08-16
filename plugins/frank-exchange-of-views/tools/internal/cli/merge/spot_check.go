package merge

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// spot-check: re-verify archived closures, and say which.
//
// The duty exists because a closure index is only as good as the last time
// anyone looked. W1.8 keys the floor on the archive's state at round START —
// run 5's round-2 merge entered with an empty archive, so the old
// from-round-2 rule degraded into a seat attesting blocks it was about to write
// itself.
func newSpotCheck() *cobra.Command {
	var ids flags.CSV

	c := seat.New("spot-check",
		`the round archive spot-check record (W1.8 duty): --ids R1-4,R2-7 [--notes "..."] | --none --reason "..." when the archive was empty at round start`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			none := seat.Given(cmd, flags.None)
			if none && len(ids.Value()) > 0 {
				return nil, fmt.Errorf("--none and --ids are contradictory: either you sampled closures or there were none to sample")
			}
			if none && strings.TrimSpace(seat.Str(cmd, flags.Reason)) == "" {
				return nil, fmt.Errorf("--none requires --reason: an empty discharge that does not say WHY is indistinguishable from a skipped duty")
			}
			p := record.NewPayload()
			seat.SetList(p, "ids", &ids)
			seat.Set(cmd, p, "notes", flags.Notes)
			if none {
				p.Set("none", true)
				seat.SetSame(cmd, p, flags.Reason)
			}
			if _, err := record.Append(s.Identity(), "spot-check", p); err != nil {
				return nil, err
			}
			if none {
				return spotCheckResult{}, nil
			}
			return spotCheckResult{Sampled: ids.Value()}, nil
		})

	c.Flags().Var(&ids, flags.IDs, "comma-separated archived closures you re-verified this round")
	c.Flags().String(flags.Notes, "", "what the spot-check found")
	// AN HONESTLY-EMPTY ROUND IS A DISCHARGE, NOT A SKIP.
	//
	// This run's red-merge-r1 reported in friction that spot-check "cannot record an
	// honestly-empty round" because --ids requires a list. That is NOT what the tool
	// did: a bare spot-check was always accepted and recorded ids as an empty array —
	// TestSpotCheckIdsAreAlwaysAnArray pins exactly that, and caught the attempt to
	// make the bare form an error. The friction was self-report that did not survive
	// checking, the second such entry in this run.
	//
	// What the seat actually lacked was EXPRESSIVENESS: an empty array records that
	// nothing was sampled but not WHY, so a later audit cannot tell "the archive was
	// empty at round start, floor not applicable" from "the seat skipped its duty".
	// --none --reason says which, and the bare form keeps working.
	c.Flags().Bool(flags.None, false, "the archive was empty at round start, so there was nothing to sample (requires --reason)")
	c.Flags().String(flags.Reason, "", "why there was nothing to sample")
	return c
}

type spotCheckResult struct {
	Sampled []string `json:"sampled"`
}

func (r spotCheckResult) Human() string {
	if len(r.Sampled) == 0 {
		return "spot-check: nothing to sample, recorded with its reason"
	}
	return "spot-checked " + strings.Join(r.Sampled, ", ")
}
