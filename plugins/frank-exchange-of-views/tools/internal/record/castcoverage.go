package record

import (
	"sort"
	"strings"
)

// THE DIMENSION NOBODY SAT IS A FACT THE RECORD ALREADY HOLDS, AND NOTHING SUBTRACTED.
//
// `setup` writes the cast before any seat registers, so which lens areas this run seats is settled
// up front and is on the record. The full roster is DefaultCastAreas. The areas that never sat are
// one set difference away — and no reader computed it. `DeriveVerdict` reads the cast only to ask
// whether one EXISTS, so it can tell CEILING from a run that ended early.
//
// What that cost, measured (#877): a run whose cast was narrowed to evidence and logic shipped a
// VERIFIED report carrying 19 voice tells. `setup` accepted the narrowing silently, capture audited
// nothing about it, and the stamp said VERIFIED — which a reader takes to mean red looked and found
// nothing material, when for voice nobody looked at all.
//
// IT IS THE CLASS THIS SYSTEM ALREADY GRADES `always` MATERIAL. `verification-scope-blindspot` is
// "less was checked than the report implies", and a lens roster that silently drops a dimension is
// that defect committed by the run itself, one level above the report its lenses audit.
//
// Derived and never stored: a second copy of "which areas sat" would be a fact the cast already
// carries, free to drift from it. This is the subtraction, done where it is read.
//
// hasCast SEPARATES TWO ZEROES, and the first draft did not. "No cast" and "a complete cast" both
// have nothing unseated, so both returned an empty list and a caller could not tell a whole audit
// from a run that never opened — the same shape this function exists to remove from the verdict.
// InCast draws the same distinction for the same reason.
func UnseatedAreas(run Run) (unseated []string, hasCast bool, err error) {
	cast, err := CastOf(run)
	if err != nil {
		return nil, false, err
	}
	// NO CAST IS NOT FULL COVERAGE, and it is not this function's question either. A record with no
	// cast has not narrowed anything — it has not started. Callers that must refuse on a missing
	// cast already do (register, the dispatch verb); answering "nothing is unseated" here would
	// hand them a clean bill for a run that never opened.
	if cast == nil {
		return nil, false, nil
	}
	seated := map[string]bool{}
	for _, s := range cast {
		if a := strings.TrimPrefix(s, "red-lens-"); a != s {
			seated[a] = true
		}
	}
	var out []string
	for _, a := range DefaultCastAreas {
		if !seated[strings.TrimPrefix(a, "red-lens-")] {
			out = append(out, strings.TrimPrefix(a, "red-lens-"))
		}
	}
	sort.Strings(out)
	return out, true, nil
}

// CoverageNote is the sentence a terminal verdict carries when a dimension never sat, and the empty
// string when the cast was whole.
//
// SEPARATION, NEVER DELETION — the rule the report's own voice work settled on. The narrowing may be
// entirely right: a product question can have no figure for `computation` to re-derive, and the
// skill tells an operator to narrow with a reason. What is not right is a verdict that reads the
// same whether every dimension sat or two of them did not. This states the limit on the answer
// rather than arguing about the run, which is the same shape as "the search reached only
// English-language sources".
func CoverageNote(unseated []string) string {
	if len(unseated) == 0 {
		return ""
	}
	noun, verb := "dimension", "was"
	if len(unseated) > 1 {
		noun, verb = "dimensions", "were"
	}
	return "no lens sat for " + strings.Join(unseated, ", ") + " — " +
		"that audit " + noun + " " + verb + " never performed on this run, so the verdict is silent about it rather than clear of it"
}
