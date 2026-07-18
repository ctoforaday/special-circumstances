package blue

import (
	"math"
	"strconv"
	"strings"

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
	c := seat.Prose(seat.New(role, "revision",
		"the round-record event (the CHANGELOG entry body via --file) — singleton per seat-round; emit AFTER your report edits land",
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return "", err
			}
			p := record.NewPayload().Set("text", text)
			// Absent and zero both drop: a claim_count of 0 is not a count, it is a
			// seat that did not report one.
			if cc := seat.Str(cmd, "claim-count"); cc != "" && cc != "0" {
				p.Set("claim_count", parseCount(cc))
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "revision", p); err != nil {
				return "", err
			}
			return "revision recorded — the round is on the record", nil
		}))

	c.Flags().String("claim-count", "", "FOOTNOTED declarative claims in the report (a footnoted sentence counts once)")
	return c
}

// parseCount mirrors the oracle's Number(): an unparseable value becomes null in
// the event rather than a silent zero, because zero is a claim and null is not.
func parseCount(s string) any {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return int64(f)
	}
	return f
}
