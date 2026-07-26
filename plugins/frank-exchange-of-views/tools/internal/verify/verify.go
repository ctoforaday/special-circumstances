// Package verify cross-checks a run's RECORD against the invariants that must hold if the
// record is sound — replacing the ad-hoc grep/python forensics a human ran by eye. It is
// strictly READ-ONLY: it replays the board and asserts, it never writes an event. As the
// record becomes the only inter-agent channel (plans/record-only-channel.md), this is how a
// run's completeness and consistency become a checkable result rather than an eyeball.
//
// Every check here sees ONLY the record. Cross-channel checks (the workflow envelope's
// self-reported count vs the board's truth — the #83 class) are served by EXPOSING the
// authoritative numbers in Stats, so an external consumer has something to compare against;
// verify does not read the envelope.
package verify

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Check is one invariant's result. Violations lists the specific offenders (gap ids, refs)
// so a failure points at what to fix, not merely that something is wrong.
type Check struct {
	Name       string   `json:"name"`
	OK         bool     `json:"ok"`
	Detail     string   `json:"detail"`
	Violations []string `json:"violations,omitempty"`
}

func ok(name, detail string) Check { return Check{Name: name, OK: true, Detail: detail} }

func result(name, okDetail, badDetail string, violations []string) Check {
	if len(violations) == 0 {
		return ok(name, okDetail)
	}
	sort.Strings(violations)
	return Check{Name: name, OK: false, Detail: badDetail, Violations: violations}
}

// Run executes every invariant against the replayed board, in a stable order.
func Run(b *record.Board) []Check {
	return []Check{
		gapsDisposed(b),
		foundByResolves(b),
		dialecticRefsResolve(b),
		supersedesResolve(b),
		passClosesAllGaps(b),
		registerBeforeAppend(b),
	}
}

// Failed returns the checks that did not hold — the exit-code signal for a caller.
func Failed(checks []Check) []Check {
	var out []Check
	for _, c := range checks {
		if !c.OK {
			out = append(out, c)
		}
	}
	return out
}

// gapsDisposed: a gap is either OPEN or CLOSED WITH A RECORDED REASON. A closure carries its
// reason in one of two fields depending on WHO closed it: a merge `close` carries a
// `closure_class`, while a bench `opinion` that ends the gap carries a `disposition`
// (closed / rebuttal_sustained / risk_accepted — see replay.go benchClosesGap). Either is a
// decision; a closed gap with NEITHER is a torn closure — closed by the replay with no reason
// on the record.
func gapsDisposed(b *record.Board) Check {
	var bad []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Open {
			continue
		}
		if g.Closure == nil || (g.Closure.Str("closure_class") == "" && g.Closure.Str("disposition") == "") {
			bad = append(bad, id)
		}
	}
	return result("gaps-disposed",
		"every closed gap carries a reason (a closure class or a bench disposition)",
		"a gap is closed with no recorded reason (a torn closure, not a decision)", bad)
}

// foundByResolves: every gap's found_by credit names a finding or observation that actually
// exists on the record. Enforced at write time; verified here against the replayed set in case
// a shard was hand-edited or lost.
func foundByResolves(b *record.Board) Check {
	labels := map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "finding" || e.Type == "observe" {
			if l := e.Payload.Str("label"); l != "" {
				labels[l] = true
			}
		}
	}
	var bad []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Mint == nil {
			continue
		}
		for _, l := range g.Mint.StrList("found_by") {
			if !labels[l] {
				bad = append(bad, fmt.Sprintf("%s→%s", id, l))
			}
		}
	}
	return result("found-by-resolves",
		"every found_by credit names a real finding",
		"a gap credits a finding that no lens recorded (attribution to a finding nobody made)", bad)
}

// dialecticRefsResolve: every act that argues ABOUT a gap (closing, dispute, dispute-respond,
// opinion) names a gap that exists. A dialectic event pointing at a phantom gap is a thread
// with no anchor — it renders under nothing and audits nothing.
func dialecticRefsResolve(b *record.Board) Check {
	var bad []string
	for _, e := range b.Events {
		switch e.Type {
		case "closing", "dispute", "dispute-respond", "opinion":
			gid := e.Payload.Str("gap_id")
			if gid != "" && b.Gaps[gid] == nil {
				bad = append(bad, fmt.Sprintf("%s/%s→%s", e.SeatID, e.Type, gid))
			}
		}
	}
	return result("dialectic-refs-resolve",
		"every closing/dispute/opinion names a gap that exists",
		"a dialectic act argues about a gap that is not on the record", bad)
}

// supersedesResolve: a successor gap's lineage names ancestors that exist — the chain the
// docket detector follows must not dead-end at a phantom.
func supersedesResolve(b *record.Board) Check {
	var bad []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Mint == nil {
			continue
		}
		for _, anc := range g.Mint.StrList("supersedes") {
			if b.Gaps[anc] == nil {
				bad = append(bad, fmt.Sprintf("%s⊃%s", id, anc))
			}
		}
	}
	return result("supersedes-resolve",
		"every supersedes ancestor is a real gap",
		"a gap's lineage names an ancestor that is not on the record", bad)
}

// passClosesAllGaps: the #67 gate, verified after the fact. A PASS verdict with an open gap is
// a contradiction — the record says the run resolved everything, and it did not.
func passClosesAllGaps(b *record.Board) Check {
	verdict := ""
	for _, e := range b.Events {
		if e.Type == "outcome" {
			verdict = e.Payload.Str("verdict")
		}
	}
	if verdict != "PASS" {
		return ok("pass-closes-all-gaps", fmt.Sprintf("verdict is %s — gate not applicable", nonEmpty(verdict, "unrecorded")))
	}
	var open []string
	for _, id := range b.GapOrder {
		if g := b.Gaps[id]; g != nil && g.Open {
			open = append(open, id)
		}
	}
	return result("pass-closes-all-gaps",
		"PASS and no gap left open",
		"the verdict is PASS but gaps are still open (the #67 gate was violated)", open)
}

// registerBeforeAppend: a seat's FIRST event must be its register — an event from a seat that
// never registered is an identity the record cannot vouch for.
func registerBeforeAppend(b *record.Board) Check {
	seen := map[string]bool{}
	var bad []string
	for _, e := range b.Events {
		if seen[e.SeatID] {
			continue
		}
		seen[e.SeatID] = true
		if e.Type != "register" {
			bad = append(bad, fmt.Sprintf("%s (first event was %s)", e.SeatID, e.Type))
		}
	}
	return result("register-before-append",
		"every seat registered before it wrote",
		"a seat emitted an event before registering", bad)
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// Stats is the authoritative tally of a run's record — the numbers a human would otherwise
// grep for. GapsOpen is the ground truth an external consumer must match: the 2026-07-22 run's
// envelope self-reported 10 outstanding while GapsOpen was 2 (#83). Dialectic coverage
// (gaps_with_closing/dispute/opinion) is how "is the argument on the record?" becomes a number
// — the record-only migration's progress bar.
type Stats struct {
	Verdict          string         `json:"verdict"`
	Events           map[string]int `json:"events"`
	GapsTotal        int            `json:"gaps_total"`
	GapsOpen         int            `json:"gaps_open"`
	GapsClosed       int            `json:"gaps_closed"`
	Findings         int            `json:"findings"`
	FindingsMinted   int            `json:"findings_minted"`
	FindingsUnminted int            `json:"findings_unminted"`
	// Citations is the count of cite events on the record — the canonical source for
	// the envelope's citations_checked, which red used to self-report (fabricated on
	// haiku). Cite events are keyed by reference (deriveKey), so re-verifying a source
	// updates in place rather than adding: this counts DISTINCT sources verified, and
	// equals the citation-ledger projection's row count.
	Citations       int `json:"citations"`
	GapsWithClosing int `json:"gaps_with_closing"`
	GapsWithDispute int `json:"gaps_with_dispute"`
	GapsWithOpinion int `json:"gaps_with_opinion"`
}

// Compute tallies the record. Read-only, one replay.
func Compute(b *record.Board) Stats {
	s := Stats{Events: map[string]int{}}
	for _, e := range b.Events {
		s.Events[e.Type]++
		if e.Type == "outcome" {
			s.Verdict = e.Payload.Str("verdict")
		}
	}

	minted := map[string]bool{}
	for _, id := range b.GapOrder {
		if g := b.Gaps[id]; g != nil && g.Mint != nil {
			for _, l := range g.Mint.StrList("found_by") {
				minted[l] = true
			}
		}
	}
	findingLabels := map[string]bool{}
	withClosing, withDispute, withOpinion := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "finding" {
			if l := e.Payload.Str("label"); l != "" {
				findingLabels[l] = true
			}
		}
		if e.Type == "cite" {
			s.Citations++
		}
		if gid := e.Payload.Str("gap_id"); gid != "" {
			switch e.Type {
			case "closing":
				withClosing[gid] = true
			case "dispute":
				withDispute[gid] = true
			case "opinion":
				withOpinion[gid] = true
			}
		}
	}
	s.Findings = len(findingLabels)
	for l := range findingLabels {
		if minted[l] {
			s.FindingsMinted++
		} else {
			s.FindingsUnminted++
		}
	}
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil {
			continue
		}
		s.GapsTotal++
		if g.Open {
			s.GapsOpen++
		} else {
			s.GapsClosed++
		}
	}
	s.GapsWithClosing = len(withClosing)
	s.GapsWithDispute = len(withDispute)
	s.GapsWithOpinion = len(withOpinion)
	return s
}
