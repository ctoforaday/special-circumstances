package debatejs

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// positionDuty is the sentence that states the position duty in a dispatch.
const positionDuty = "YOUR POSITION ARGUES TO THE BENCH"

// closingDuty is the sentence that tells a party the plan holds a closing owed by how to argue the
// closing its work list names.
const closingDuty = "A CLOSING ARGUMENT IS OWED WHERE YOUR WORK LIST NAMES ONE"

// positionWord is the event's name as a word of its own — not `disposition`, not `composition`.
var positionWord = regexp.MustCompile(`(?i)\bpositions?\b`)

// returnedProse is a request for prose returned to the engine in place of an act: a synopsis, a line
// per hypothesis. No reader of the returned text exists, and what it would say the acts record.
var returnedProse = regexp.MustCompile(`(?i)synopsis|line per hypothesis`)

// THE ARGUMENT DUTIES ARE STATED ONLY TO THE SEAT THAT OWES THEM, AND ONCE. Every dispatch the engine
// makes is rendered — lanes, the frontier, the synthesizer, the chair, a lens, blue's response, the
// bench — over a plan with a docketed gap whose closings both parties owe.
//
//   - THE POSITION IS THE CHAIR'S (record.PositionSittings): its duty sentence is in the chair's
//     dispatch alone — counted over what the seat is handed, its agent definition, the skills it
//     preloads and the prompt — and every other dispatch holds the word nowhere.
//   - THE CLOSING IS EACH PARTY'S (record.closingBlockersOf), AND THE WORK LIST NAMES IT: a lens and
//     blue-respond the plan's closing_owed blockers name are told how to argue the closing their work
//     list names, and no other seat is told of one. No dispatch names the gap: the plan is read at
//     the epoch's start, and a lens that closes the gap before blue sits ends the duty (#1663 smoke,
//     blue ev530, ev608).
//   - NO DISPATCH ASKS FOR RETURNED PROSE in place of an act: nothing reads it.
func TestTheArgumentDutiesAreStatedOnlyToTheSeatThatOwesThem(t *testing.T) {
	blueParty := map[string]any{"seat_id": "blue-respond", "gap_ids": []any{}}
	owedBy := func(seat string) map[string]any {
		return map[string]any{"kind": string(record.BlockerClosingOwed), "subject": "G1", "owner": seat}
	}
	blockers := []any{owedBy("red-lens-logic"), owedBy("blue-respond"), map[string]any{"kind": "unruled_motion", "subject": "M1", "owner": "judge"}}
	docketBench := map[string]any{"seat_id": "judge", "gap_ids": []any{"G2"}}
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, blueParty, docketBench}, blockers)}, nil)

	seen := map[string]bool{}
	constitution := map[string]string{}
	for _, d := range ds {
		if d.SeatID == "" {
			continue // a re-prompt of a seat already sitting: it names no seat and restates no duty
		}
		seen[d.SeatID] = true
		typ := strings.TrimPrefix(d.AgentType, "frank-exchange-of-views:")
		if _, ok := constitution[typ]; !ok {
			path, err := repotree.Plugin("agents", typ+".md")
			if err != nil {
				t.Fatal(err)
			}
			text, err := repotree.ConstitutionText(path)
			if err != nil {
				t.Fatalf("%s: %v", d.SeatID, err)
			}
			constitution[typ] = text
		}
		owes := d.SeatID == "red-chair"
		if got := strings.Count(constitution[typ]+"\n"+d.Prompt, positionDuty); (got == 1) != owes || got > 1 {
			t.Errorf("%s (%s): the dispatch states the position duty %d time(s), and the seat owes it = %v", d.SeatID, d.Label, got, owes)
		}
		if !owes {
			if m := positionWord.FindStringIndex(d.Prompt); m != nil {
				lo, hi := max(0, m[0]-80), min(len(d.Prompt), m[1]+80)
				t.Errorf("%s (%s) owes no position and its dispatch speaks of one: …%s…", d.SeatID, d.Label, d.Prompt[lo:hi])
			}
		}
		party := d.SeatID == "red-lens-logic" || d.SeatID == "blue-respond"
		if got := strings.Count(d.Prompt, closingDuty); got != map[bool]int{true: 1}[party] {
			t.Errorf("%s (%s): the dispatch states the closing duty %d time(s), and the seat is a party owing it = %v", d.SeatID, d.Label, got, party)
		}
		if strings.Contains(d.Prompt, "OWED ON") || strings.Contains(d.Prompt, "closing owed on G1") {
			t.Errorf("%s (%s): the dispatch names a gap a closing is owed on — the work list is the one place that does", d.SeatID, d.Label)
		}
		if m := returnedProse.FindStringIndex(d.Prompt); m != nil {
			lo, hi := max(0, m[0]-80), min(len(d.Prompt), m[1]+80)
			t.Errorf("%s (%s) is asked for returned prose no reader reads: …%s…", d.SeatID, d.Label, d.Prompt[lo:hi])
		}
	}
	// THE DRIVE REACHED EVERY CLASS, or the loop above passed seats it never read.
	for _, seat := range []string{"blue-lane-1", "frontier", "blue-synthesize", "red-chair", "red-lens-logic", "blue-respond", "judge"} {
		if !seen[seat] {
			t.Errorf("the drive dispatched no %s: its prompt was not read (seats seen: %v)", seat, seen)
		}
	}
}
