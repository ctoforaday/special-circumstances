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

import "strings"

// Classification is the seat a transcript head resolves to.
type Classification struct {
	Seat string
	// THERE IS NO ROUND HERE (plans/roundless.md §III.A.2). A seat's epoch and sitting are the
	// record's — read from the chair's registers — and a fact recovered from prompt wording that
	// the record already holds is the shape that migration removed. The head names the SEAT only.
}

// A seat is identified by the opening text of its prompt, so this table is coupled to
// debate.js's prompt wording — and bound to it: heads_bind_test.go classifies every rendered
// prompt golden the simulator keeps and demands the seat in the golden's name. The needles are
// tried in order, so a more specific head is listed before one that could sit in its preamble.
//
// ARCHIVED heads stay. A class read is a read of history as often as of a live run — cost and
// the dashboard read transcripts from run-archive/ — and an archived transcript carries the
// heading the engine wrote when it ran. The two generations cannot collide.
var needles = []struct {
	needle string
	seat   string
}{
	// The live heads (debate.js: lensPrompt, chairPrompt, bluePrompt, benchPrompt, the petition,
	// terminal and assembly sittings, synthesis, the lanes, the frontier).
	{"Red lens sitting, area", "red-lens"},
	{"Red chair, topic", "red-chair"},
	{"Blue response, topic", "blue-respond"},
	{"Adjudication, topic", "judge"},
	{"Terminal disposition", "judge-terminal"},
	{"Petition sitting", "judge-petition"},
	{"Blue synthesis", "blue-synthesize"},
	{"Blue lane", "blue-lane"},
	{"frontier hypotheses", "frontier"},
	{"Final assembly", "assemble"},
	// The heads the round-shaped engine wrote, kept for the archive.
	{"Red audit, round", "red-lens"},
	{"Red chair, round", "red-chair"},
	{"Blue response, round", "blue-respond"},
	{"Adjudication, round", "judge"},
	{"Terminal dispute disposition", "judge-terminal"},
}

// ClassifySeat resolves a prompt head to its seat. An unrecognized head is `other` —
// a visible bucket, never folded away, so a prompt-wording drift is spottable.
func ClassifySeat(head string) Classification {
	for _, n := range needles {
		if strings.Contains(head, n.needle) {
			return Classification{Seat: n.seat}
		}
	}
	return Classification{Seat: "other"}
}

// KnownSeats is every seat this table can name (each once, plus "other") — so a test can assert
// the consumers agree on the full set, not only the cases someone checked.
func KnownSeats() []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(needles)+1)
	for _, n := range needles {
		if !seen[n.seat] {
			seen[n.seat] = true
			out = append(out, n.seat)
		}
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
