package chair

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// spot-check: name the stale areas read before a PASS, and the archived closures re-verified.
//
// It is owed before a PASS and at no other time: the PASS gate refuses a verdict until a spot-check
// in the chair's sitting names every stale area (record.BlockerStaleArea). The archived closures
// it names are the chair's sample of a closure index that is only as good as the last time anyone
// looked.
func newSpotCheck() *cobra.Command {
	var ids, areas flags.CSV

	c := seat.New("spot-check", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		// THE CHANNEL, NOT THE FLAG. seat.Reason is the one resolver: it refuses a read of a
		// channel the verb never registered and trims what a captured heredoc leaves behind.
		// Reading the flag directly skips both — this verb once shipped doing exactly that, with
		// no file form at all while every other verb had one.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body := &recordpb.SpotCheck{
			Ids:    ids.Value(),
			Areas:  areas.Value(),
			Reason: proto.String(why),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return spotCheckResult{Sampled: ids.Value()}, nil
	})

	c.Flags().Var(&ids, flags.IDs, "comma-separated archived closures you re-verified this sitting")
	c.Flags().Var(&areas, flags.Areas, "comma-separated lens seats from the plan's stale_areas whose area you read the changes against this sitting — a PASS is refused until a spot-check this sitting names every stale area; what you found goes in --reason")
	// ONE PROSE CHANNEL, ALWAYS REQUIRED: what the spot-check found, or why it names nothing — the
	// board says how much the archive held, and the reason says why none of it was re-read.
	// REQUIRED, THROUGH EITHER SPELLING, and stated once by the channel itself rather than by a
	// group this verb has to remember to declare. MarkFlagRequired names one flag and would refuse
	// the file form, which is what the hand-registered --reason here used to do.
	seat.ProseRequired(c)
	// The verb's name is not its event's word, so the event is declared.
	seat.Records(c, "spot_check")
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
