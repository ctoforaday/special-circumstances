// Package seatclass recovers a seat's identity from the head of the prompt the engine gave
// it, and maps each seat to its tier CLASS (bulk / judgment). ONE table — a Go port of
// seat-classify.mjs, which itself exists because two JS copies of this classifier had drifted
// (cost-audit lacked the terminal-disposition case, misattributing that seat's spend). The Go
// cost port reads from here for the same reason: one source per fact, or the copies drift.
package seatclass

import (
	"regexp"
	"strconv"
	"strings"
)

// Classification is the seat and round a transcript head resolves to.
type Classification struct {
	Seat  string
	Round int
}

// A seat is identified by the opening text of its prompt, so this table is coupled to
// debate.js's prompt wording. ROUNDED needles carry the round; matched FIRST because a
// round-0 marker can appear in a round-bearing prompt's preamble.
var rounded = []struct {
	re   *regexp.Regexp
	seat string
}{
	{regexp.MustCompile(`Red audit, round (\d+)`), "red-lens"},
	{regexp.MustCompile(`Red merge, round (\d+)`), "red-merge"},
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

// ClassifySeat resolves a prompt head to its seat and round. An unrecognized head is `other`
// (round 0) — a visible bucket, never folded away, so a prompt-wording drift is spottable.
func ClassifySeat(head string) Classification {
	for _, r := range rounded {
		if m := r.re.FindStringSubmatch(head); m != nil {
			n, _ := strconv.Atoi(m[1])
			return Classification{Seat: r.seat, Round: n}
		}
	}
	for _, u := range unrounded {
		if strings.Contains(head, u.needle) {
			return Classification{Seat: u.seat, Round: 0}
		}
	}
	return Classification{Seat: "other", Round: 0}
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
	"red-merge":       "judgment",
	"judge":           "judgment",
	"judge-petition":  "judgment",
	"judge-terminal":  "judgment",
	"assemble":        "judgment",
}

// ClassOf returns a seat's tier class, or "" for other/unknown seats (not tier-bound).
func ClassOf(seat string) string { return SeatClass[seat] }
