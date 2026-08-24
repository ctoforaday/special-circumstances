package record

import (
	"regexp"
	"strconv"
)

var roundRe = regexp.MustCompile(`-r(\d+)`)

// roundlessShapes are the seat ids whose NAME cannot answer which round they act in.
//
// Split by whether the absence means round 0 or means "after the last one", because those are
// different facts and the old code answered 0 to both:
//
//	synthesis  frontier, blue-synthesize, blue-lane-<N> — dispatched BEFORE the round loop
//	           (debate.js: `let round = 0` sits after them), so round 0 is correct and exact.
//	terminal   judge-terminal, assemble — dispatched AFTER it. Round 0 is not merely imprecise
//	           here, it is backwards: a bench closure at run END rendered as a closure BEFORE
//	           round 1, which put a phantom entry in the archive and made the W1.8 spot-check
//	           floor demand samples from rounds whose seats had done nothing wrong (#327,
//	           found at 1 seed in 60, by luck).
var (
	synthesisSeats = map[string]bool{"frontier": true, "blue-synthesize": true}
	terminalSeats  = map[string]bool{"judge-terminal": true, "assemble": true}
)

var laneRe = regexp.MustCompile(`^blue-lane-\d+$`)

// RoundOf answers which round a seat acts in, AND whether its id answers that at all.
//
// THE SECOND RETURN IS THE WHOLE POINT. This used to return a bare int, and 0 meant both "round
// 0, which is synthesis and a real round with real events" and "this name says nothing about a
// round". Identity.Round has documented the right convention all along — "-1 means unknown, which
// is NOT round 0" — and nothing ever produced it, because the only producer was this function.
// The contract and its sole implementation disagreed, and the reader could not tell.
//
// AND THE REGEX IS NO LONGER A GUESS. It was, and the comment here said so for good reason: `-r(\d+)`
// over an arbitrary string is a hope about shape. As of the roster gate, a REGISTERED seat id has
// been refused unless it matches one of the engine's own patterns — `red-lens-r\d+-L\d+`,
// `red-merge-r\d+`, `judge-r\d+` — so the segment this reads is one the tool has already validated
// rather than one it is hoping to find. What changed was not this regex; it is what stands behind it.
//
// FEOV_ROUND IS GONE. It was read here first and set by nothing, anywhere, ever — the same
// readers-and-no-writer shape as FEOV_SEAT, and it is deleted for the same reason: an injected
// branch no run can reach is not a fallback, it is scenery.
func RoundOf(seatID string) (round int, known bool) {
	if m := roundRe.FindStringSubmatch(seatID); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n, true
		}
	}
	if synthesisSeats[seatID] || laneRe.MatchString(seatID) {
		return 0, true
	}
	return 0, false
}

// RoundIn resolves the round for a seat acting in a run, including the ones no name can answer.
//
// A TERMINAL SEAT'S ROUND IS A FACT ON THE RECORD, not one it has to be told. `judge-terminal` and
// `assemble` act after the last round, so "the last round" is what they are in — and the record
// already knows it, because every round's seats stamped theirs. That makes the answer a derivation
// from FIELDS rather than from a name, which is the distinction this whole exercise is about.
//
// UNKNOWN STAYS UNKNOWN when the record cannot answer either — an empty run, or a seat that is not
// part of the debate at all. Returning 0 there would rebuild the conflation one layer up.
func RoundIn(runDir string) func(string) int {
	return func(seatID string) int {
		if n, known := RoundOf(seatID); known {
			return n
		}
		if terminalSeats[seatID] {
			if n, ok := LastRoundOn(runDir); ok {
				return n
			}
		}
		return -1
	}
}

// LastRoundOn is the highest round any seat has recorded acting in.
//
// The bool is not decoration: a run with no rounds on it yet and a run whose record cannot be read
// both have no answer, and a caller that took 0 for either would be back where this started.
func LastRoundOn(runDir string) (int, bool) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return 0, false
	}
	max, found := 0, false
	for _, e := range m.Events {
		if r := int(e.GetRound()); r > max {
			max, found = r, true
		}
	}
	return max, found
}
