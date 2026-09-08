// Package dashboard ports render-run-dashboard.mjs — the live run dashboard — to Go
// (#121 slice 4). It renders the run's own instruments into a single dashboard.html,
// byte-identical to the Node script under a shared injected clock (--now) so the port is
// differential-verifiable. This file holds the model types + the pure projectCompletion;
// render.go holds the HTML/SVG string building.
package dashboard

import (
	"math"
	"strconv"
)

func itoa(n int) string { return strconv.Itoa(n) }

// Seat is one debate seat's lifecycle + identity, as buildModel assembles it. StartedMs/
// EndedMs are *float64 so "not finite" (JS null) is distinct from a real 0 timestamp —
// projectCompletion's Number.isFinite guard depends on that distinction.
type Seat struct {
	AgentID string
	Done    bool
	Result  any
	Label   string
	Seat    string
	// Epoch and Sitting are the record's two windows at this agent's register (plans/roundless.md
	// §III.A.0), read off the register event the journal's agentId joins to — NOT parsed out of the
	// transcript head. Epoch is the chair's register count at that moment (the dispatch cycle the
	// sitting belongs to); Sitting is this seat's own register count (the ordinal the label
	// carries). Both are 0 when the record never bound the agent — a run whose hook did not fire,
	// or no record yet — and 0 is also the honest epoch of a bookend seat that sat before the
	// chair ever did.
	Epoch     int
	Sitting   int
	StartedMs *float64
	EndedMs   *float64
}

// Eta is projectCompletion's result.
type Eta struct {
	State           string
	LowMin          int
	HighMin         int
	PerEpochLowMin  int
	PerEpochHighMin int
	Basis           string
	Unmeasured      []string
}

// projectCompletion answers "how much longer?" from THIS run's own measured seat durations —
// a byte-faithful port of the JS. nowMs is injected (the CLI defaults it to the real clock)
// so the estimate is deterministic under the differential.
func projectCompletion(seats []Seat, nowMs float64) Eta {
	finite := func(p *float64) bool { return p != nil && !math.IsInf(*p, 0) && !math.IsNaN(*p) }

	var done, live []Seat
	for _, s := range seats {
		if finite(s.StartedMs) && finite(s.EndedMs) && *s.EndedMs > *s.StartedMs {
			done = append(done, s)
		}
	}
	for _, s := range seats {
		if !s.Done && finite(s.StartedMs) {
			live = append(live, s)
		}
	}

	assembleDone := false
	for _, s := range seats {
		if s.Seat == "assemble" && s.Done {
			assembleDone = true
			break
		}
	}
	if len(done) > 0 && assembleDone {
		return Eta{State: "complete", LowMin: 0, HighMin: 0, Basis: "assembly finished"}
	}

	// byClass: durations (minutes) grouped by seat, in first-seen order (order only matters
	// for the min/max span, which is order-independent).
	byClass := map[string][]float64{}
	byClassSeen := map[string]bool{}
	for _, s := range done {
		byClass[s.Seat] = append(byClass[s.Seat], (*s.EndedMs-*s.StartedMs)/60000)
		byClassSeen[s.Seat] = true
	}
	type span struct {
		lo, hi float64
		ok     bool
	}
	spanOf := func(seat string) span {
		v := byClass[seat]
		if len(v) == 0 {
			return span{}
		}
		lo, hi := v[0], v[0]
		for _, x := range v[1:] {
			if x < lo {
				lo = x
			}
			if x > hi {
				hi = x
			}
		}
		return span{lo, hi, true}
	}

	assembly := spanOf("assemble")
	if !assembly.ok {
		assembly = spanOf("blue-synthesize")
	}

	var lo, hi float64
	var missing []string
	for _, s := range live {
		sp := spanOf(s.Seat)
		elapsed := (nowMs - *s.StartedMs) / 60000
		if !sp.ok {
			missing = append(missing, s.Label)
			continue
		}
		lo += math.Max(0, sp.lo-elapsed)
		hi += math.Max(0, sp.hi-elapsed)
	}
	anyAssemble := false
	for _, s := range seats {
		if s.Seat == "assemble" {
			anyAssemble = true
			break
		}
	}
	if !anyAssemble {
		if assembly.ok {
			lo += assembly.lo
			hi += assembly.hi
		} else {
			missing = append(missing, "assembly")
		}
	}

	// One more epoch costs one more pass of the debate seats: lenses, chair, blue's response.
	epoch := struct{ lo, hi float64 }{}
	for _, c := range []string{"red-lens", "red-chair", "blue-respond"} {
		if sp := spanOf(c); sp.ok {
			epoch.lo += sp.lo
			epoch.hi += sp.hi
		}
	}

	state := "complete"
	if len(live) > 0 || !byClassSeen["assemble"] {
		state = "running"
	}
	return Eta{
		State:           state,
		LowMin:          int(math.Round(lo)),
		HighMin:         int(math.Round(hi)),
		PerEpochLowMin:  int(math.Round(epoch.lo)),
		PerEpochHighMin: int(math.Round(epoch.hi)),
		Basis:           itoa(len(done)) + " completed seat(s) in this run",
		Unmeasured:      missing,
	}
}
