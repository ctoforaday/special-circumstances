package capture

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RUN THAT LOST A LANE LOOKS EXACTLY LIKE A RUN THAT ASKED FOR FEWER.
//
// `lanes` is the breadth of synthesis — how many independent slices of the frontier get drafted
// before anything is merged. It is declared in inputs/run-config.json and, until this audit, read
// by nothing that could check it: outside `setup`, `dashboard` and `debatejs`, which render or
// pass the value, no capture audit, no verify invariant and nothing in `record` ever compared it
// to what happened. A run configured for three lanes whose third crashed, was never dispatched, or
// died before registering leaves a board with two `blue-lane-N` seats — byte-identical to a board
// from a run that asked for two.
//
// Found by the sibling sweep the rule-sweep gate forced on the served-model work (#589/#603): of
// every field run-config carries, `k`, `kMax`, `mintBudget`, `eventSchema`, `runDir` and the two model tiers are
// each reconciled against something the run actually did. `lanes` was the one that was not.
//
// # Why a shortfall WARNS and an excess FAILS
//
// They are not equally ambiguous. More lane seats than declared is impossible under any legitimate
// dispatch — the engine would have run something the config never asked for. FEWER has two causes
// this audit cannot tell apart: a lane that died, and an operator deliberately narrowing on a
// resume, which is what `laneFloorOverride` exists for. That override is a debate.js argument and
// is NOT recorded in run-config, so the record cannot say which happened.
//
// Reporting a deliberate reduction as a failure would train a reader to ignore the line, which is
// how a gate stops being read. So the shortfall is named, counted, and its two readings are stated
// — and the remedy is named too, because "this cannot be distinguished" is a defect report about
// the record, not a permanent property.

// laneSeatsThatSat is the lane seat IDS that opened a sitting, deduplicated.
//
// IDS RATHER THAN INDEXES, AND THAT IS THE WHOLE CHANGE. This read the index back out of the seat
// id by pattern, which made the audit the last reader in the tree recovering a fact from a name.
// The ids are GENERATED from the declared lane count (record.LaneSeatIDs), so membership is a
// comparison and the answer a reader wants — WHICH lane is missing — is a name already in hand.
func laneSeatsThatSat(evs []*record.Event, lanes []string) map[string]bool {
	want := map[string]bool{}
	for _, id := range lanes {
		want[id] = false
	}
	sat := map[string]bool{}
	for i := range evs {
		if seat, opens := recordpb.SeatOpeningSitting(evs[i]); opens {
			if _, isLane := want[seat]; isLane {
				sat[seat] = true
			}
		}
	}
	return sat
}

// LaneCoverageAudit joins the declared lane count to the lane seats that actually registered.
func LaneCoverageAudit(run record.Run) Audit {
	want, declared := record.LanesDeclared(run)
	if !declared {
		return Audit{Check: "lane-coverage", Verdict: "SKIP",
			Detail: "this run declared no lane count in run-config, so there is nothing to hold its lanes to — NOT a run whose lanes were checked"}
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "lane-coverage", Verdict: "SKIP",
			Detail: fmt.Sprintf("run-config declares %d lane(s) and the record could not be read, so none of them could be confirmed — NOT a run whose lanes all registered", want)}
	}
	// A LANE THAT WAS BRACKETED BUT NEVER REGISTERED STILL SAT. Since #1089 a seat woken with
	// nothing to do need not register, so an audit that counted registers would report a lane that
	// worked as absent. SeatOpeningSitting is the one predicate that answers what opened a sitting.
	// THE CAST NAMES THE LANES; run-config says how many were ASKED FOR. Both are read, because the
	// two disagreeing is exactly what this audit exists to report.
	lanes := record.LaneSeatsOf(run)
	if len(lanes) == 0 {
		return Audit{Check: "lane-coverage", Verdict: "SKIP",
			Detail: fmt.Sprintf("run-config declares %d lane(s) and the record's cast names none, so there is nothing to hold them to — NOT a run whose lanes were checked", want)}
	}
	sat := laneSeatsThatSat(fam.Events, lanes)

	var missing []string
	for _, id := range lanes {
		if !sat[id] {
			missing = append(missing, id)
		}
	}
	// AN EXTRA LANE IS A CAST QUESTION NOW, NOT AN ARITHMETIC ONE. It used to be "an index above
	// the declared count"; it is a lane seat the run's own cast does not name, which is the same
	// defect said without recovering a number from a name.
	// AN EXTRA LANE IS A COUNT DISAGREEMENT NOW, NOT AN ARITHMETIC ONE OVER NAMES. The cast names
	// the lanes; run-config says how many the operator asked for. More named than asked for is the
	// engine and the run's own config disagreeing about how wide synthesis was.
	var extra []string
	if len(lanes) > want {
		extra = append(extra, lanes[want:]...)
	}
	got := make([]string, 0, len(sat))
	for _, id := range lanes {
		if sat[id] {
			got = append(got, id)
		}
	}

	detail := fmt.Sprintf("run-config declares %d lane(s); %d registered", want, len(got))
	switch {
	case len(extra) > 0:
		// Unambiguous: no legitimate dispatch produces a lane the config never asked for.
		return Audit{Check: "lane-coverage", Verdict: "FAIL",
			Detail: detail + fmt.Sprintf("; %s registered beyond the declared count, which no dispatch of this config could have produced — the engine and the run's own config disagree about how wide synthesis was",
				strings.Join(extra, ", "))}
	case len(missing) > 0:
		return Audit{Check: "lane-coverage", Verdict: "WARN",
			Detail: detail + fmt.Sprintf("; %s never registered. TWO READINGS, and this record cannot separate them: a lane that died before its first act, or an operator narrowing deliberately on a resume (laneFloorOverride), which is a debate.js argument run-config does not carry. Synthesis was %d slices wide, not the %d the report will describe",
				strings.Join(missing, ", "), len(got), want)}
	default:
		return Audit{Check: "lane-coverage", Verdict: "PASS", Detail: detail + "; every declared lane took its seat"}
	}
}
