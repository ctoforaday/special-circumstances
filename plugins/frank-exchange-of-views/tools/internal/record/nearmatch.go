package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"math"
	"sort"
)

// NEAR-MATCH: a lexical screen, not a decision.
//
// Before a lens mints a fresh gap it must ask "is this already on the board?" — a
// near-duplicate should reopen (mint --supersedes) the prior gap, not fork a second one.
// That screen used to be the chair reading the whole board and eyeballing it. Here it is a
// tool op: score the candidate's text against every gap's problem+location by token
// overlap, return the top few. The tool SCREENS — it ranks candidates and never decides;
// the seat reads the ranking and calls reopen-or-new. No store, no model, no embedding: a
// brute-force lexical pass over ≤~150 gaps is well within budget and stays deterministic.

// NearMatchJSON is one ranked screen result: which gap, how strong the overlap, where it
// sits, whether it is still open, and — where it is closed — WHO closed it.
//
// CLOSED_BY DECIDES THE ANSWER THIS SCREEN EXISTS TO INFORM, so it is on the row. The screen's
// whole job is reopen-or-new, and those are not the same question for the two kinds of closure:
// red may reopen its OWN closure on new evidence, while a bench ruling is ESTOPPED and re-raising
// it is relitigation rather than diligence — new evidence there is a lineage successor that names
// the ruled gap in `supersedes` and says what the ruling did not account for.
//
// Without it the screen returned `status: closed` for both and the seat had to go and find the
// distinction somewhere else, which is a fact arriving after the decision it governs.
type NearMatchJSON struct {
	ID       string  `json:"id"`
	Score    float64 `json:"score"`
	Location string  `json:"location"`
	Status   string  `json:"status"` // open | closed
	// ClosedBy is "bench" or "red" on a closed gap, and empty on an open one — absent because
	// there is nothing to say, not because it is unknown.
	ClosedBy string `json:"closed_by,omitempty"`
}

// locationBonus nudges a candidate that shares its section with a gap: two defects in the
// same location are likelier the same defect. Additive, capped with the base at 1.0.
const locationBonus = 0.15

func round2(f float64) float64 { return math.Round(f*100) / 100 }

// NearMatch scores a candidate (problem text, and optionally its location) against every
// gap on the board and returns the top-N by descending score, id-ascending on ties. Both
// OPEN and CLOSED gaps are scored — a reopen most often matches a closed gap. A gap with no
// overlap at all is omitted (score 0 is not a match).
func NearMatch(f Family, candidate, location string, topN int) []NearMatchJSON {
	candTokens := anchortext.Tokenize(candidate + " " + location)
	locTokens := anchortext.Tokenize(location)

	out := []NearMatchJSON{}
	for _, g := range f.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		gapLoc := g.Mint.GetLocation()
		score := anchortext.Jaccard(candTokens, anchortext.Tokenize(g.Mint.GetProblem()+" "+gapLoc))
		if score <= 0 {
			continue
		}
		if len(locTokens) > 0 && len(intersect(locTokens, anchortext.Tokenize(gapLoc))) > 0 {
			score += locationBonus
		}
		if score > 1 {
			score = 1
		}
		status, closedBy := "closed", "red"
		if g.Open {
			status, closedBy = "open", ""
		} else if g.ClosedByBench {
			closedBy = "bench"
		}
		out = append(out, NearMatchJSON{ID: g.ID, Score: round2(score), Location: gapLoc, Status: status, ClosedBy: closedBy})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID < out[j].ID
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}

func intersect(a, b map[string]bool) []string {
	var out []string
	for t := range a {
		if b[t] {
			out = append(out, t)
		}
	}
	return out
}
