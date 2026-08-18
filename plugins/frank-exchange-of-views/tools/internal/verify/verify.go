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
	Name string `json:"name"`
	OK   bool   `json:"ok"`
	// NA marks a check that DID NOT APPLY, as distinct from one that held. Both have OK
	// true — deliberately, so the exit code and every existing consumer of `ok` are
	// unchanged — but they are not the same fact, and rendering them the same way is how a
	// dead gate hides.
	//
	// MEASURED: `pass-closes-all-gaps` read `outcome` events for the word PASS, which that
	// vocabulary cannot hold, so it took its inapplicable branch on EVERY run ever recorded
	// and printed "[ok  ] … gate not applicable" — a line that reads like a considered
	// judgement. A check that can only ever be inapplicable is indistinguishable from one
	// that holds, unless the two are marked differently (#411).
	//
	// An inapplicable check is NEVER a failure: a run that halts legitimately leaves the
	// PASS gate inapplicable, and that is correct, not a defect.
	NA         bool     `json:"na"`
	Detail     string   `json:"detail"`
	Violations []string `json:"violations"`
}

func ok(name, detail string) Check { return Check{Name: name, OK: true, Detail: detail} }

// notApplicable records that the check did not fire, and WHY. The reason is required by
// construction — an unexplained "did not apply" is the same unreadable zero the state exists
// to remove.
func notApplicable(name, why string) Check {
	return Check{Name: name, OK: true, NA: true, Detail: why}
}

// Applicable reports whether the check actually examined anything. Kept as a method so a
// renderer asks the Check rather than re-deriving the rule from its Detail string.
func (c Check) Applicable() bool { return !c.NA }

// Status is the word a renderer prints: "ok", "n/a" or "FAIL". One definition, so the report
// section and the CLI cannot drift into disagreeing about what a result means.
func (c Check) Status() string {
	switch {
	case c.NA:
		return "n/a"
	case c.OK:
		return "ok"
	default:
		return "FAIL"
	}
}

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
		archiveSpotCheckFloor(b),
	}
}

// archiveSpotCheckFloor: W1.8, enforced at last from the board rather than reported by the seat.
//
// The duty was born from a real defect — run 5's round-2 spot-check "had degraded to same-seat
// self-attestation" — and its fix shipped as an envelope self-report, which was deleted in 2026
// for comparing numbers the merge made up. Nothing replaced it. The `merge spot-check` verb
// carried a receipt that NOTHING READ, so the fix for a self-attestation defect was a better
// place to write the self-attestation.
//
// Two teeth, both against replayed state no seat can author: a round that entered with a
// non-empty archive and recorded no sample, and a round that CLAIMED an empty archive the board
// says was not empty. The second is the direct heir of the run-5 degeneracy.
func archiveSpotCheckFloor(b *record.Board) Check {
	_, debt, falseEmpty := record.SpotCheckAudit(b)
	var violations []string
	for _, round := range debt {
		violations = append(violations, fmt.Sprintf("round %d: the merge sat with archived closures available and recorded no spot-check", round))
	}
	for _, sc := range falseEmpty {
		violations = append(violations, fmt.Sprintf("round %d (%s): discharged with --none (%q) while the board shows %d archived closure(s) at round start",
			sc.Round, sc.SeatID, sc.NoneReason, sc.Archived))
	}
	return result("archive-spot-check-floor",
		"every round that entered with a non-empty archive sampled it",
		"the archive spot-check floor was not met — a closure index is only as good as the last time anyone looked, and these rounds did not look",
		violations)
}

// NotApplicable returns the checks that did not fire. Separate from Failed because they are
// separate facts: Failed drives the exit code, this drives what a reader is told. A run where
// every check is n/a exits 0 and has verified nothing, and only this can say so.
func NotApplicable(checks []Check) []Check {
	var out []Check
	for _, c := range checks {
		if c.NA {
			out = append(out, c)
		}
	}
	return out
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
		if e.Type == "finding" {
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

// passVerdict is the word this gate switches on, and the EVENT TYPE that can carry it. Both
// are read back out of enums.go rather than written here, because writing them here is the
// defect this constant exists to prevent.
//
// The check read `e.Type == "outcome"`. An outcome's vocabulary is
// VERIFIED|CEILING|HALTED|UNVERIFIED — validated at the write, so it can never hold "PASS".
// Two event types share the payload key `verdict` with DISJOINT vocabularies, so the wrong
// one is a plain type error that Go cannot see: the comparison never matched, the check took
// the "gate not applicable" branch on every run ever recorded, and said so in words that read
// like a considered judgement — "verdict is VERIFIED — gate not applicable".
//
// Severity, stated honestly rather than inflated: the LIVE gate works. record.Append refuses
// `merge verdict --as PASS` while any gap is open ("1 gap(s) still OPEN: R1-1"), so the
// contradiction cannot arise through the tool. What was lost is the after-the-fact half — the
// one that exists for a record assembled some OTHER way: a hand-edited shard, a legacy run, or
// a live gate that itself regressed. That is precisely the case a verifier is for, and it was
// the case that never ran.
//
// TestPassVerdictIsAWordItsEventTypeCanActuallyCarry asserts the pair against the schema, so
// re-pointing this at the wrong event type — or renaming the word — FAILS instead of going
// quietly not-applicable. That is the part that generalises; the fix below is one instance.
const (
	passVerdictType = "verdict"
	passVerdictWord = "PASS"
)

// passClosesAllGaps: the #67 gate, verified after the fact. A PASS verdict with an open gap is
// a contradiction — the record says the run resolved everything, and it did not.
func passClosesAllGaps(b *record.Board) Check {
	verdict := ""
	for _, e := range b.Events {
		if e.Type == passVerdictType {
			verdict = e.Payload.Str("verdict")
		}
	}
	if verdict != passVerdictWord {
		return notApplicable("pass-closes-all-gaps", fmt.Sprintf("the verdict is %s, so there is no PASS to contradict", nonEmpty(verdict, "unrecorded")))
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
		if e.Type == "verify" { // red's verifications only — blue's authored cites are not audit volume (#341)
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
