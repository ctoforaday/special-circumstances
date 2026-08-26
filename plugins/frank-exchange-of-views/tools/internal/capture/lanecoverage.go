package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RUN THAT LOST A LANE LOOKS EXACTLY LIKE A RUN THAT ASKED FOR FEWER.
//
// `lanes` is the breadth of round 0 — how many independent slices of the frontier get drafted
// before anything is merged. It is declared in inputs/run-config.json and, until this audit, read
// by nothing that could check it: outside `setup`, `dashboard` and `debatejs`, which render or
// pass the value, no capture audit, no verify invariant and nothing in `record` ever compared it
// to what happened. A run configured for three lanes whose third crashed, was never dispatched, or
// died before registering leaves a board with two `blue-lane-N` seats — byte-identical to a board
// from a run that asked for two.
//
// Found by the sibling sweep the rule-sweep gate forced on the served-model work (#589/#603): of
// every field run-config carries, `maxRounds`, `eventSchema`, `runDir` and the two model tiers are
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

var laneSeat = regexp.MustCompile(`^blue-lane-(\d+)$`)

// declaredLanes reads the lane count run-config declares.
//
// The three answers are kept apart: a declared count, a run that declared none (older runs, and
// `--lanes` is optional), and a config that could not be read at all. Only the first can be
// checked, and the other two must not read as agreement.
func declaredLanes(runDir string) (n int, declared bool) {
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		return 0, false
	}
	var cfg map[string]any
	if json.Unmarshal(b, &cfg) != nil {
		return 0, false
	}
	// setup writes it as a STRING (ptrOrNil over the flag), so that is what is read first; a
	// number is accepted too rather than silently missed if the writer ever changes.
	switch v := cfg["lanes"].(type) {
	case string:
		if v == "" {
			return 0, false
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil || n <= 0 {
			return 0, false
		}
		return n, true
	case float64:
		if v <= 0 {
			return 0, false
		}
		return int(v), true
	}
	return 0, false
}

// registeredLanes returns the lane INDEXES that registered, deduplicated and sorted.
//
// Indexes rather than a count, because which lane is missing is the actionable half: "lane 3 never
// registered" sends a reader to that dispatch, where "2 of 3" sends them to the whole round.
func registeredLanes(b *record.Board) []int {
	seen := map[int]bool{}
	for i := range b.Events {
		e := b.Events[i]
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		if m := laneSeat.FindStringSubmatch(e.GetSeatId()); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				seen[n] = true
			}
		}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// LaneCoverageAudit joins the declared lane count to the lane seats that actually registered.
func LaneCoverageAudit(runDir string) Audit {
	want, declared := declaredLanes(runDir)
	if !declared {
		return Audit{Check: "lane-coverage", Verdict: "SKIP",
			Detail: "this run declared no lane count in run-config, so there is nothing to hold its lanes to — NOT a run whose lanes were checked"}
	}
	board, err := record.BoardState(runDir)
	if err != nil || board == nil {
		return Audit{Check: "lane-coverage", Verdict: "SKIP",
			Detail: fmt.Sprintf("run-config declares %d lane(s) and the record could not be read, so none of them could be confirmed — NOT a run whose lanes all registered", want)}
	}
	got := registeredLanes(board)

	var missing, extra []string
	present := map[int]bool{}
	for _, n := range got {
		present[n] = true
		if n > want {
			extra = append(extra, "blue-lane-"+strconv.Itoa(n))
		}
	}
	for n := 1; n <= want; n++ {
		if !present[n] {
			missing = append(missing, "blue-lane-"+strconv.Itoa(n))
		}
	}

	detail := fmt.Sprintf("run-config declares %d lane(s); %d registered", want, len(got))
	switch {
	case len(extra) > 0:
		// Unambiguous: no legitimate dispatch produces a lane the config never asked for.
		return Audit{Check: "lane-coverage", Verdict: "FAIL",
			Detail: detail + fmt.Sprintf("; %s registered beyond the declared count, which no dispatch of this config could have produced — the engine and the run's own config disagree about how wide round 0 was",
				strings.Join(extra, ", "))}
	case len(missing) > 0:
		return Audit{Check: "lane-coverage", Verdict: "WARN",
			Detail: detail + fmt.Sprintf("; %s never registered. TWO READINGS, and this record cannot separate them: a lane that died before its first act, or an operator narrowing deliberately on a resume (laneFloorOverride), which is a debate.js argument run-config does not carry. Round 0's breadth was %d slices, not the %d the report will describe",
				strings.Join(missing, ", "), len(got), want)}
	default:
		return Audit{Check: "lane-coverage", Verdict: "PASS", Detail: detail + "; every declared lane took its seat"}
	}
}
