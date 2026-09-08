package capture

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// DispatchParityAudit holds the workflow's RELAY to the record (plans/roundless.md §III.B.1). The
// chair's verb records who sits; the chair relays that into its envelope; the workflow dispatches
// what the envelope says — and the chair could drop or add a party on the way. So: within the
// window from the first dispatch to termination, every register between one chair sitting's
// dispatches D and the next's D+1 is either a party named in D or the seat that authors D+1 (the
// chair registering to run the verb is exempt BY RULE, not by allowlist), and every party named in
// D registered before D+1. The bookends register outside the window by position: the base phase
// before the first dispatch, judge-terminal and assemble after the last chair sitting.
//
// The verb records the truth; this catches a relay that departed from it. A mismatch is a FAIL.
func DispatchParityAudit(run record.Run) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
	}
	type reg struct {
		pos  int
		seat string
	}
	type group struct { // one chair sitting's dispatches
		first, last int
		parties     map[string]bool
	}
	var regs []reg
	var groups []*group
	var clk record.Clock
	lastChairSitting := -1
	for i, e := range fam.Events {
		w := clk.Advance(e)
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Register:
			regs = append(regs, reg{pos: i, seat: e.GetSeatId()})
		case *recordpb.Dispatch:
			if w.Sitting != lastChairSitting || len(groups) == 0 {
				groups = append(groups, &group{first: i, parties: map[string]bool{}})
				lastChairSitting = w.Sitting
			}
			g := groups[len(groups)-1]
			g.last = i
			g.parties[b.GetSeatId()] = true
		}
	}
	if len(groups) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "SKIP", Detail: "no dispatch on the record — a run before the chair dispatched, or one that never reached the debate"}
	}
	terminal := map[string]bool{"judge-terminal": true, "assemble": true}
	var strays, absent []string
	for k, g := range groups {
		end := len(fam.Events)
		if k+1 < len(groups) {
			end = groups[k+1].first
		}
		sat := map[string]bool{}
		for _, r := range regs {
			if r.pos <= g.last || r.pos >= end {
				continue
			}
			sat[r.seat] = true
			if g.parties[r.seat] || r.seat == "red-chair" || (k+1 == len(groups) && terminal[r.seat]) {
				continue
			}
			strays = append(strays, fmt.Sprintf("%s registered after dispatch %d and was not a party to it", r.seat, k+1))
		}
		for p := range g.parties {
			if !sat[p] {
				absent = append(absent, fmt.Sprintf("%s was named in dispatch %d and never registered before the next", p, k+1))
			}
		}
	}
	sort.Strings(strays)
	sort.Strings(absent)
	if len(strays)+len(absent) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "PASS", Detail: fmt.Sprintf("%d dispatch(es); every register between them was a named party or the chair, and every party registered", len(groups))}
	}
	return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: strings.Join(append(strays, absent...), "; ")}
}
