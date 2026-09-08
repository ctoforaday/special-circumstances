// Package seatclass recovers a seat's identity from the head of the prompt the engine gave
// it, and maps each seat to its tier CLASS (bulk / judgment). ONE table — it began as a Go port
// of seat-classify.mjs, which itself existed because two JS copies of this classifier had drifted
// (cost-audit lacked the terminal-disposition case, misattributing that seat's spend).
//
// seat-classify.mjs is now DELETED (#121 slice 5): this is the SOLE canonical seat→class map.
// No copy lives in debate.js — it never read one, it hardcodes `...bulk`/`...judgment` at each
// dispatch — and dispatch_bind_test.go reads debate.js SOURCE and asserts every dispatch spreads
// the tier ClassOf reports (with no dead key), so the map stays bound to the actual dispatches
// from the side that owns it. The drift hole the old free-floating JS oracle left is closed.
package seatclass

import (
	"regexp"
	"strings"
)

// Classification is the seat a transcript head resolves to.
type Classification struct {
	Seat string
	// THERE IS NO ROUND HERE ANY MORE (plans/roundless.md §III.A.2). The head still carries one
	// while debate.js still says "round N", and the needles below still match on it, but the
	// number was read by nothing once cost and the dashboard took a seat's epoch and sitting from
	// the record's registers — and a fact recovered from prompt wording that the record already
	// holds is the shape this migration exists to remove.
}

// A seat is identified by the opening text of its prompt, so this table is coupled to
// debate.js's prompt wording. ROUNDED needles carry the round; matched FIRST because a
// round-0 marker can appear in a round-bearing prompt's preamble.
var rounded = []struct {
	re   *regexp.Regexp
	seat string
}{
	{regexp.MustCompile(`Red audit, round (\d+)`), "red-lens"},
	{regexp.MustCompile(`Red chair, round (\d+)`), "red-chair"},
	// The heading an ARCHIVED transcript carries. A class read is a read of history as often as
	// of a live run, and the two headings cannot collide.
	{regexp.MustCompile(`Blue response, round (\d+)`), "blue-respond"},
	{regexp.MustCompile(`Adjudication, round (\d+)`), "judge"},
}

var unrounded = []struct {
	needle string
	seat   string
}{
	{"Terminal dispute disposition", "judge-terminal"},
	{"Petition sitting", "judge-petition"},
	{"Blue synthesis", "blue-synthesize"},
	{"Blue lane", "blue-lane"},
	{"frontier hypotheses", "frontier"},
	{"Final assembly", "assemble"},
}

// ClassifySeat resolves a prompt head to its seat. An unrecognized head is `other` —
// a visible bucket, never folded away, so a prompt-wording drift is spottable.
func ClassifySeat(head string) Classification {
	for _, r := range rounded {
		if m := r.re.FindStringSubmatch(head); m != nil {
			return Classification{Seat: r.seat}
		}
	}
	for _, u := range unrounded {
		if strings.Contains(head, u.needle) {
			return Classification{Seat: u.seat}
		}
	}
	return Classification{Seat: "other"}
}

// KnownSeats is every seat this table can name (rounded + unrounded + "other") — so a test
// can assert the consumers agree on the full set, not only the cases someone checked.
func KnownSeats() []string {
	out := make([]string, 0, len(rounded)+len(unrounded)+1)
	for _, r := range rounded {
		out = append(out, r.seat)
	}
	for _, u := range unrounded {
		out = append(out, u.seat)
	}
	return append(out, "other")
}

// SeatClass is the ONE seat→tier-class map. debate.js dispatches each seat as bulk or
// judgment; the tier guard needs that split as data. bulk = high-volume seats (lenses, lanes,
// responses, frontier); judgment = reasoning seats (synthesis, merge, the bench, assembly).
var SeatClass = map[string]string{
	"frontier":        "bulk",
	"blue-lane":       "bulk",
	"red-lens":        "bulk",
	"blue-respond":    "bulk",
	"blue-synthesize": "judgment",
	"red-chair":       "judgment",
	"judge":           "judgment",
	"judge-petition":  "judgment",
	"judge-terminal":  "judgment",
	"assemble":        "judgment",
}

// ClassOf returns a seat's tier class, or "" for other/unknown seats (not tier-bound).
func ClassOf(seat string) string { return SeatClass[seat] }
