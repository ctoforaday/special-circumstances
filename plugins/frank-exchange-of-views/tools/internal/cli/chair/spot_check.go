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

// spot-check: name the stale areas judged before a PASS, and the archived closures re-verified. The
// write stamps what changed behind each area's pin (record.stampSpotCheckChanges) and the verb prints
// it back; the chair's prose is its judgement of those changes, never their count.
//
// It is owed before a PASS and at no other time: the PASS gate refuses a verdict until a spot-check
// in the chair's sitting names every stale area (record.BlockerStaleArea). The archived closures
// it names are the chair's sample of a closure index that is only as good as the last time anyone
// looked.
func newSpotCheck() *cobra.Command {
	ids := flags.IDList(flags.GapID)
	var areas flags.CSV

	c := seat.New("spot-check", spotCheckResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
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
		// THE ECHO IS THE STAMP. The write put each area's pin and the changes behind it on the body;
		// the chair is told what the record holds there, never asked to say it.
		return spotCheckResultOf(body), nil
	})

	c.Flags().Var(ids, flags.IDs, "comma-separated gap ids (G-…) of the closed gaps whose closures you re-verified this sitting")
	c.Flags().Var(&areas, flags.Areas, "comma-separated lens seats from the plan's stale_areas whose changes since the pin you judged against the area's duties this sitting — the tool states each pin and the changes behind it; a PASS is refused until a spot-check this sitting names every stale area; what you found goes in --reason")
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
	Areas   []spotCheckAreaJSON `json:"areas"`
	Sampled []string            `json:"sampled"`
}

// spotCheckAreaJSON is one area as the write stamped it: the pin, and the event id of each change to
// the report after it, with their count.
type spotCheckAreaJSON struct {
	SeatID  string  `json:"seat_id"`
	Pin     int64   `json:"pin"`
	Count   int     `json:"count"`
	Changes []int64 `json:"changes"`
}

// spotCheckResultOf is the echo of a written spot-check, read off the body the write stamped.
func spotCheckResultOf(body *recordpb.SpotCheck) spotCheckResult {
	r := spotCheckResult{Sampled: body.GetIds()}
	for _, a := range record.SpotCheckAreasOf(body) {
		r.Areas = append(r.Areas, spotCheckAreaJSON{SeatID: a.SeatID, Pin: a.Pin, Count: len(a.Changes), Changes: a.Changes})
	}
	return r
}

// Human states the act the chair recorded and what the record holds behind each area's pin, and
// nothing else about the board. A spot-check names what the chair chose to name, so a call naming no
// closure says the chair named none — the archive may hold closures it did not re-read, and the
// board, not this line, says how many.
func (r spotCheckResult) Human() string {
	var named []string
	if len(r.Areas) > 0 {
		areas := make([]record.SpotCheckArea, 0, len(r.Areas))
		for _, a := range r.Areas {
			areas = append(areas, record.SpotCheckArea{SeatID: a.SeatID, Pin: a.Pin, Changes: a.Changes})
		}
		named = append(named, "read stale area(s) "+record.DescribeAreas(areas))
	}
	if len(r.Sampled) > 0 {
		named = append(named, "re-verified "+strings.Join(r.Sampled, ", "))
	} else {
		named = append(named, "named no archived closure")
	}
	return "spot-check recorded with its reason: " + strings.Join(named, "; ")
}
