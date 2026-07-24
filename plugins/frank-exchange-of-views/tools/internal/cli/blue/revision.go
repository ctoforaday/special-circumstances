package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// revision: the round-record event (W1.7).
//
// A revision that is not on the record did not happen as far as the debate is
// concerned. Run 5 shipped a round-3 report revision with neither a debate.md
// section nor a CHANGELOG entry: a lens misjudged the round state from the
// absence, and the judge had to reconstruct blue's whole position from red's
// ledger. This event is the machine-checkable half of that fix.
func newRevision() *cobra.Command {
	// claim_count is NO LONGER carried here. It was hand-typed via --claim-count and
	// stored on the revision event, a second source for a number that only mattered in
	// the envelope (which debate.js reads to size dispatch); revision itself is not on
	// the live path — the round narrative is a position event and the CHANGELOG is
	// hand-written. #70 moved the count to the deterministic `count-claims` command and
	// dropped this flag, so there is exactly one way the number is produced.
	return seat.Prose(seat.New("revision",
		"the round-record event (the CHANGELOG entry body via --reason) — singleton per seat-round; emit AFTER your report edits land",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			p := record.NewPayload().Set("text", text)
			if _, err := record.Append(s.RunDir, s.SeatID, "revision", p); err != nil {
				return nil, err
			}
			return seat.Msg{Message: "revision recorded — the round is on the record"}, nil
		}))
}
