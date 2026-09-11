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
// D SAT for it — record.DispatchGroup's Sat, the one "has this seat sat" predicate — before D+1.
// The bookends register outside the window by position: the base phase before the first dispatch,
// judge-terminal and assemble after the last chair sitting.
//
// A CHAIR SITTING'S DISPATCH IS record.DispatchGroups', NOT THE CLOCK'S CHAIR-SITTING COUNT. This
// audit grouped by the count, and a warm chair registers once per run: B3–B6 each read as a single
// chair sitting whose last dispatch was the run's last, so every register before it was discarded
// and every party of "dispatch 1" read as never registered — the same FAIL on every warm run,
// whether or not the parties sat.
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
	var regs []reg
	for i, e := range fam.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
			regs = append(regs, reg{pos: i, seat: e.GetSeatId()})
		}
	}
	groups := record.DispatchGroups(fam.Events)
	if len(groups) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "SKIP", Detail: "no dispatch on the record — a run before the chair dispatched, or one that never reached the debate"}
	}
	terminal := map[string]bool{"judge-terminal": true, "assemble": true}
	var strays, absent []string
	for k, g := range groups {
		end := len(fam.Events)
		if k+1 < len(groups) {
			end = groups[k+1].First
		}
		party := map[string]bool{}
		for _, p := range g.Parties {
			party[p] = true
		}
		for _, r := range regs {
			if r.pos <= g.Last || r.pos >= end {
				continue
			}
			if party[r.seat] || r.seat == "red-chair" || (k+1 == len(groups) && terminal[r.seat]) {
				continue
			}
			strays = append(strays, fmt.Sprintf("%s registered after dispatch %d and was not a party to it", r.seat, k+1))
		}
		for _, p := range g.Parties {
			if at, sat := g.Sat[p]; !sat || at >= end {
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
