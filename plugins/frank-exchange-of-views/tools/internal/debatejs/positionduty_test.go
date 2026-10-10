package debatejs

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// positionDuty is the sentence that states the position duty in a dispatch.
const positionDuty = "YOUR POSITION ARGUES TO THE BENCH"

// positionWord is the event's name as a word of its own — not `disposition`, not `composition`.
var positionWord = regexp.MustCompile(`(?i)\bpositions?\b`)

// THE POSITION DUTY IS STATED ONLY TO A SEAT THAT OWES ONE. The prompts are prose, so they cannot
// call record.SeatOwesPosition; this holds them to it. Every dispatch the engine makes is rendered —
// lanes, the frontier, the synthesizer, the chair, a lens, blue's response, the bench — and the duty
// sentence is in exactly those whose seat the predicate names.
//
// AND A SEAT THAT OWES NONE IS NOT TOLD OF ONE AT ALL. The shared record clause once said "your
// position and closings in the transcript" to every seat; across seventeen archived runs lanes, the
// frontier and the synthesizer filed 48 positions nothing held them to. A dispatch to a seat that
// owes none holds the word nowhere, so the verb's refusal is a backstop and not the first thing the
// seat learns about it.
func TestThePositionDutyIsStatedOnlyToASeatThatOwesOne(t *testing.T) {
	blueParty := map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}}
	docketBench := map[string]any{"seat_id": "judge", "gap_ids": []any{"G1"}, "occasions": []any{"docket"}}
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, blueParty, docketBench}, []any{})}, nil)

	seen := map[string]bool{}
	for _, d := range ds {
		if d.SeatID == "" {
			continue // a re-prompt of a seat already sitting: it names no seat and restates no duty
		}
		seen[d.SeatID] = true
		owes := record.SeatOwesPosition(d.SeatID)
		if got := strings.Count(d.Prompt, positionDuty); (got == 1) != owes || got > 1 {
			t.Errorf("%s (%s): the dispatch states the position duty %d time(s), and SeatOwesPosition is %v", d.SeatID, d.Label, got, owes)
		}
		if !owes {
			if m := positionWord.FindStringIndex(d.Prompt); m != nil {
				lo, hi := max(0, m[0]-80), min(len(d.Prompt), m[1]+80)
				t.Errorf("%s (%s) owes no position and its dispatch speaks of one: …%s…", d.SeatID, d.Label, d.Prompt[lo:hi])
			}
		}
	}
	// THE DRIVE REACHED EVERY CLASS, or the loop above passed seats it never read.
	for _, seat := range []string{"blue-lane-1", "frontier", "blue-synthesize", "red-chair", "red-lens-logic", "blue-respond", "judge"} {
		if !seen[seat] {
			t.Errorf("the drive dispatched no %s: its prompt was not read (seats seen: %v)", seat, seen)
		}
	}
}
