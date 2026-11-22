package fuzz

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A PROPOSAL ON A GAP WHOSE LENS SITTING THE RECORD CANNOT CLOSE IS REFUSED NOT MEASURED, and only a
// sitting with no stop holds it there. The engine awaits every lens before blue sits, and a lens's stop
// is on the record when its agent returns, so a live blue filing finds the lens's audit closed and its
// own sitting open — which hides no exchange: the proposal lands. A lens sitting engaged on the gap that
// no stop closes — a register with no agent, or an agent whose stop was lost — stays open until the lens
// sits again, and no blue sitting after it ends a round: the sweep's dispute drive meets that shape when
// the lens never sits again (heldUnexchanged).
func TestAProposalWaitsOnALensSittingTheRecordCannotClose(t *testing.T) {
	for _, stop := range []bool{false, true} {
		name := "no stop: open until the lens sits again"
		if stop {
			name = "the lens's stop closes it"
		}
		t.Run(name, func(t *testing.T) {
			bin := buildBinary(t)
			runDir := recordtest.TmpRun(t)
			stageRun, err := record.NewRun(runDir)
			if err != nil {
				t.Fatal(err)
			}
			cast, lanes := record.CastFor(1)
			if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
				t.Fatal(err)
			}
			r := newRunner(bin, runDir, newLockedRand(7))
			r.forceVerified = true // a forced run leaves no drawn sitting open
			const sentence = "A § fuzz sentence to anchor findings."
			if err := r.stagePart(sentence + "\n"); err != nil {
				t.Fatalf("record the lane's part: %v", err)
			}
			r.chairPlans(t)
			lens := fuzzLensSeats[0]
			r.dispatchSits("lens", lens)
			r.mustExec(t, "--json", "--seat-id", lens, "show", "report")
			r.mustExec(t, "--json", "--seat-id", lens, "mint", "--quote", sentence,
				"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
				"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
			r.dispatchReturns(lens, false)
			gapID := r.openGaps()[0]
			// BLUE ANSWERS THE GAP, so its sitting ends no round: the lens's audit of the answer does.
			r.chairPlans(t)
			r.dispatchOpens("blue-respond")
			r.openUnderAgent = true
			r.sit("blue", "blue-respond")
			r.openUnderAgent = false
			r.do("edit", "blue-respond").set("--quote", "fuzz sentence").set("--new", "fuzz statement").
				set("--answers", gapID).set("--reason", "fuzz: blue's answer").admit()
			r.splitQuestion("blue-respond")
			r.coverQuestionParts("blue-respond")
			r.dispatchReturns("blue-respond", r.envelopeEnds("blue-respond"))
			if r.actErr != "" {
				t.Fatalf("blue's answer: %s", r.actErr)
			}
			// THE LENS SITS ENGAGED ON THE GAP to audit the answer: under an agent whose stop closes the
			// sitting, or by a register no stop closes.
			r.dispatchSits("chair", "red-chair")
			plan := r.mustExec(t, "--json", "dispatch", "next", "--seat-id", "red-chair")
			r.dispatchReturns("red-chair", false)
			if !strings.Contains(plan, `"seat_id":"`+lens+`","gap_ids":["`+gapID+`"]`) {
				t.Fatalf("the plan after blue's answer does not engage %s on %s: %s", lens, gapID, plan)
			}
			r.dispatchOpens(lens)
			if stop {
				r.openUnderAgent = true
				r.sit("lens", lens)
				r.openUnderAgent = false
			} else {
				r.mustExec(t, append([]string{"--json"}, registerArgs(lens, "")...)...)
			}
			r.dispatchReturns(lens, r.envelopeEnds(lens))

			r.chairPlans(t)
			r.dispatchOpens("blue-respond")
			r.openUnderAgent = true
			r.sit("blue", "blue-respond")
			r.openUnderAgent = false
			out, err := r.exec(withJSON(proposeCloseArgs("blue-respond", gapID, "not_a_defect", "fuzz: blue's dispute of "+gapID))...)
			switch {
			case stop && err != nil:
				t.Fatalf("blue's proposal after the lens's stop closed its sitting was refused: %s", firstLine(out))
			case !stop && err == nil:
				t.Fatalf("blue's proposal while the lens's sitting stands open was admitted: %s", firstLine(out))
			case !stop && !strings.Contains(out, "exchanges NOT MEASURED — 2 sitting(s)"):
				t.Fatalf("blue's proposal while the lens's sitting stands open was refused, and not as NOT MEASURED over the lens's sitting and blue's own: %s", firstLine(out))
			}
			if stop {
				return
			}
			r.dispatchReturns("blue-respond", false)
			// THE LENS SITS AGAIN, and its next opening closes the sitting that had no stop.
			r.dispatchOpens(lens)
			r.mustExec(t, append([]string{"--json"}, registerArgs(lens, "")...)...)
			r.dispatchReturns(lens, false)
			if !r.exchanged(gapID) {
				t.Fatalf("%s has no counted exchange once the lens sat again", gapID)
			}
		})
	}
}
