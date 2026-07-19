package blue

import (
	"math"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
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
			if cc := seat.Str(cmd, flags.ClaimCount); cc != "" && cc != "0" {
				p.Set("claim_count", parseCount(cc))
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "revision", p); err != nil {
				return "", err
			}
			return "revision recorded — the round is on the record", nil
		}))

	c.Flags().String(flags.ClaimCount, "", "FOOTNOTED declarative claims in the report (a footnoted sentence counts once)")
	return c
}

// parseCount mirrors the oracle's Number(): an unparseable value becomes null in
// the event rather than a silent zero, because zero is a claim and null is not.
func parseCount(s string) any {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil
	}
	// DEFECT FIXED: NaN and ±Inf were allowed through. Go's ParseFloat ACCEPTS
	// "NaN", "Inf" and "+Inf", so `--claim-count NaN` produced a float64 the event
	// could not be serialized with: encoding/json refuses non-finite values, the
	// marshal failed, Append returned an error, and the revision was never
	// recorded at all. That is the worst possible outcome for this particular
	// verb — a revision missing from the record is exactly the run-5 failure the
	// type comment above describes, reintroduced through a flag value.
	//
	// It is also an oracle divergence: JSON.stringify(NaN) is `null` in the mjs
	// side, so the faithful serialization of a non-finite count is null, not a
	// write failure. Returning nil here yields that same null AND satisfies this
	// function's own stated contract — a value that is not a count becomes null,
	// never a silent zero.
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return int64(f)
	}
	return f
}
