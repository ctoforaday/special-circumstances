package record

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// NEAR-MATCH: a lexical screen, not a decision.
//
// Before the merge mints a fresh gap it must ask "is this already on the board?" — a
// near-duplicate should reopen (mint --supersedes) the prior gap, not fork a second one.
// That screen used to be the merge reading the whole board and eyeballing it. Here it is a
// tool op: score the candidate's text against every gap's problem+location by token
// overlap, return the top few. The tool SCREENS — it ranks candidates and never decides;
// the seat reads the ranking and calls reopen-or-new. No store, no model, no embedding: a
// brute-force lexical pass over ≤~150 gaps is well within budget and stays deterministic.

// NearMatchJSON is one ranked screen result: which gap, how strong the overlap, where it
// sits, and whether it is still open.
type NearMatchJSON struct {
	ID       string  `json:"id"`
	Score    float64 `json:"score"`
	Location string  `json:"location,omitempty"`
	Status   string  `json:"status"` // open | closed
}

// locationBonus nudges a candidate that shares its section with a gap: two defects in the
// same location are likelier the same defect. Additive, capped with the base at 1.0.
const locationBonus = 0.15

// tokenize lowercases and splits on any non-alphanumeric rune, dropping empties and
// single-character tokens (punctuation noise, stray letters). Deterministic, unicode-aware.
func tokenize(s string) map[string]bool {
	out := map[string]bool{}
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		if len([]rune(f)) > 1 {
			out[f] = true
		}
	}
	return out
}

// jaccard is the overlap of two token sets: shared over union, 0..1. Empty on either side
// is 0 (nothing to match), so a candidate or gap with no usable tokens simply does not rank.
func jaccard(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	shared := 0
	for t := range a {
		if b[t] {
			shared++
		}
	}
	union := len(a) + len(b) - shared
	if union == 0 {
		return 0
	}
	return float64(shared) / float64(union)
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }

// NearMatch scores a candidate (problem text, and optionally its location) against every
// gap on the board and returns the top-N by descending score, id-ascending on ties. Both
// OPEN and CLOSED gaps are scored — a reopen most often matches a closed gap. A gap with no
// overlap at all is omitted (score 0 is not a match).
func NearMatch(b *Board, candidate, location string, topN int) []NearMatchJSON {
	candTokens := tokenize(candidate + " " + location)
	locTokens := tokenize(location)

	out := []NearMatchJSON{}
	for _, id := range b.GapOrder {
		g, ok := b.Gaps[id]
		if !ok || g.Mint == nil {
			continue
		}
		gapLoc := g.Mint.Str("location")
		score := jaccard(candTokens, tokenize(g.Mint.Str("problem")+" "+gapLoc))
		if score <= 0 {
			continue
		}
		if len(locTokens) > 0 && len(intersect(locTokens, tokenize(gapLoc))) > 0 {
			score += locationBonus
		}
		if score > 1 {
			score = 1
		}
		status := "closed"
		if g.Open {
			status = "open"
		}
		out = append(out, NearMatchJSON{ID: g.ID, Score: round2(score), Location: gapLoc, Status: status})
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
