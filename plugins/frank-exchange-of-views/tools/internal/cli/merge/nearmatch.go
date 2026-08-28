package merge

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// near-match: screen a candidate gap against the board BEFORE minting.
//
// The NEAR-MATCH RULE is the merge's, not the tool's: on a near-match against a closed gap,
// the candidate is a reopen (mint --supersedes), not a fresh gap. This op moves the SCREEN
// off the seat — instead of re-reading the whole board and eyeballing it, the merge asks the
// tool for the ranked matches and then decides. It is READ-ONLY (it records nothing) and it
// never decides: it ranks, the seat reopens-or-mints.
const nearMatchTopN = 5

// The help document renders this number rather than restating it: help/near-match.md names
// {{.NearMatchTopN}}, so the sentence a seat reads and the rank cut this file applies cannot
// disagree.
func init() { seat.HelpValues["NearMatchTopN"] = nearMatchTopN }

func newNearMatch() *cobra.Command {
	c := seat.New("near-match", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		cand := strings.TrimSpace(seat.Str(cmd, flags.Problem))
		b, err := record.BoardState(run)
		if err != nil {
			return nil, err
		}
		matches := record.NearMatch(b, cand, seat.Str(cmd, flags.Quote), nearMatchTopN)
		return nearMatchResult{Matches: matches}, nil
	})
	// THE SAME TWO WORDS `mint` TAKES. This verb screens what that one would file, and it spelled
	// the identical two facts --candidate and --location — so a seat learned one vocabulary to ask
	// the question and a different one to act on the answer.
	c.Flags().String(flags.Problem, "", "what is wrong, as you would state it in the mint — screened against the board for a near-duplicate")
	_ = c.MarkFlagRequired(flags.Problem)
	c.Flags().String(flags.Quote, "", flags.DescQuote+" — scored for a location-match bonus")
	return c
}

// (help text intentionally teaches WHEN and WHAT — a seat that reads only `merge near-match
// --help` learns to screen before minting and to reopen on a strong match.)

type nearMatchResult struct {
	Matches []record.NearMatchJSON `json:"matches"`
}

func (r nearMatchResult) Human() string {
	if len(r.Matches) == 0 {
		return "near-match: no gap on the board overlaps this candidate — it is genuinely new"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "near-match: %d candidate(s) (reopen with mint --supersedes <id>, or mint fresh if none is the same defect):", len(r.Matches))
	for _, m := range r.Matches {
		loc := m.Location
		if loc == "" {
			loc = "—"
		}
		fmt.Fprintf(&sb, "\n  %s [%s] score %.2f — %s", m.ID, m.Status, m.Score, loc)
	}
	return sb.String()
}
