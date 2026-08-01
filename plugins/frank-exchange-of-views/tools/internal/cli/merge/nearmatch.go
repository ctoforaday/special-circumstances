package merge

import (
	"fmt"
	"strconv"
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

func newNearMatch() *cobra.Command {
	c := seat.New("near-match",
		`screen a candidate against the board before minting: --candidate "<problem text>" [--location "<section>"] — returns the top `+strconv.Itoa(nearMatchTopN)+` gaps (open AND closed) by lexical overlap, so a near-duplicate surfaces as a reopen (mint --supersedes <id>) rather than a fresh gap. The tool SCREENS and ranks; you decide reopen-or-new. Read-only: it records nothing.`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			cand := strings.TrimSpace(seat.Str(cmd, flags.Candidate))
			if cand == "" {
				return nil, fmt.Errorf("near-match requires --candidate: the candidate gap's problem text to screen against the board")
			}
			b, err := record.BoardState(s.RunDir)
			if err != nil {
				return nil, err
			}
			matches := record.NearMatch(b, cand, seat.Str(cmd, flags.Location), nearMatchTopN)
			return nearMatchResult{Matches: matches}, nil
		})
	c.Flags().String(flags.Candidate, "", "the candidate gap's problem text to screen against the board")
	c.Flags().String(flags.Location, "", "the candidate's section/location, for the location-match bonus")
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
