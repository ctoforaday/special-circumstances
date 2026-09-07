package seatprobe

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// ceilings are the largest a seat's dispatched prompt may be, per seat class.
//
// # Why a ratchet and not a target
//
// The prompts spent months growing. Every measured defect arrived as a paragraph, every paragraph
// was individually justified, and nothing ever counted the total — so the seats were reading
// 24,000 characters of instruction against a tool whose own --help said much of it again, in the
// flag's usage line, at the moment of use. The subtraction that cut them roughly in half is only
// as durable as something that NOTICES the next paragraph, because each one will be as justified
// as the last.
//
// So this fails on GROWTH and says nothing about shrinkage. It is not a quality measure: a short
// prompt is not a good one, and nothing here can tell whether what is left is the right half. What
// it can tell is that a prompt got bigger, which is the direction the failure comes from.
//
// # Raising one is a decision, not a formality
//
// If a ceiling has to move, the question to answer first is whether the paragraph belongs in the
// TOOL: does it teach a verb, a flag, a refusal, a rendering, or an output format? All of those
// are the help's, where the seat reads them while choosing rather than three thousand characters
// earlier. What belongs here is the JOB — the standard, the discipline, the measured indictment,
// the situation this round is in.
//
// Recorded 2026-08-21, immediately after the subtraction, with ~10% headroom over the measured
// size. The numbers before it: blue-respond 24,356 · red-merge 24,193 · red-lens 13,931 ·
// judge 11,861.
var ceilings = map[string]int{
	"blue-respond-r1":      15200,
	"red-merge-r1":         13000,
	"red-lens-r1-evidence": 8700,
	"judge-r2":             7800,
}

func TestNoSeatPromptGrowsPastItsCeiling(t *testing.T) {
	got := map[string]int{}
	for _, b := range Boards() {
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("%s: %v", b.Name, err)
			continue
		}
		if n := len(d.Prompt); n > got[d.SeatID] {
			got[d.SeatID] = n
		}
	}
	if len(got) == 0 {
		t.Fatal("no board produced a prompt — a size gate that measures nothing passes every time")
	}

	var table strings.Builder
	seats := make([]string, 0, len(got))
	for s := range got {
		seats = append(seats, s)
	}
	sort.Strings(seats)
	for _, s := range seats {
		fmt.Fprintf(&table, "  %-18s %6d (ceiling %d)\n", s, got[s], ceilings[s])
	}
	t.Log("dispatched prompt sizes:\n" + table.String())

	for _, s := range seats {
		limit, ok := ceilings[s]
		if !ok {
			t.Errorf("seat %s has no recorded ceiling (%d chars) — a seat the ratchet does not cover can grow without anyone seeing it", s, got[s])
			continue
		}
		if got[s] > limit {
			t.Errorf("%s is %d characters, past its ceiling of %d.\n\n"+
				"Before raising it: does the new text teach a VERB, a FLAG, a REFUSAL, or a RENDERING? "+
				"That belongs in the tool's --help, where the seat reads it while choosing. "+
				"The prompt carries the JOB — the standard, the discipline, and this round's situation.",
				s, got[s], limit)
		}
	}
	// A ceiling naming a seat production does not dispatch is a gate on nothing.
	for s := range ceilings {
		if _, ok := got[s]; !ok {
			t.Errorf("ceiling recorded for %s, which no board dispatches — the entry guards nothing", s)
		}
	}
}
