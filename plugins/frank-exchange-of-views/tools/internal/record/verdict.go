package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE VERDICT IS THE LAST BIG DERIVED-NOT-ASSERTED VIOLATION (#308).
//
// This engine computes the facts a seat would benefit from claiming, precisely so it cannot
// claim them: fix_basis is derived from whether the pair validates against the live report,
// proof_basis from running the script twice, applied_verbatim from comparing bytes — or, on
// `blue edit --accept`, from the tool having supplied those bytes itself. Each
// exists because a seat asked to self-report reports the flattering value.
//
// The terminal verdict — the single most consequential value the engine emits, the one the
// report stamps at the top and every audit and scorecard reads — was ASSERTED. debate.js
// computed it in JS, stated it in the assembler's prompt, and the seat typed it back. Nothing
// checked the two agreed, so `--as VERIFIED` over a board with open gaps would have been
// recorded and believed.

// VerdictBasis says how far a recorded verdict can be trusted, on the same footing as
// fix_basis and proof_basis.
const (
	// VerdictDerived: the tool computed it from the record and the seat's claim agreed.
	VerdictDerived = "derived"
	// VerdictAsserted: the record cannot decide it, so the seat's word stands. That is exactly
	// one case — UNVERIFIED, a run that stopped before the record reached a terminal state — and
	// it is underivable for a nameable reason: a workflow that ends leaves no event saying it
	// ended, so the record cannot tell a finished-early run from one still in flight. `bench
	// outcome` refuses any OTHER word over such a record.
	VerdictAsserted = "asserted"
)

// DeriveVerdict computes a run's terminal verdict from the record alone.
//
// ok is false ONLY when the record holds no terminal state — the run is still in flight, or it
// ended before reaching one (UNVERIFIED) — which is a finding rather than a defect in this
// function. Everything else is already recorded:
//
//	HALTED    a halt event exists — the bench ended the run on its own authority
//	VERIFIED  the chair recorded a PASS verdict
//	CEILING   the dispatch plan's ceiling: every open material gap is at its limit and ruled,
//	          or the run's epoch limit is reached — the why says which
//
// The order matters: a halt outranks a pass, because a run stopped on safety or integrity
// grounds did not end by passing however clean the board looked when it stopped.
func DeriveVerdict(run Run) (verdict, why string, ok bool) {
	halted, err := recordHas(run, `SELECT 1 FROM "halt" LIMIT 1`)
	if err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	passed, err := recordHas(run, `SELECT 1 FROM "gate" WHERE "verdict" = ? LIMIT 1`,
		recordpb.Word(recordpb.Verdict_VERDICT_PASS))
	if err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	switch {
	case halted:
		return "HALTED", "a halt event is on the record", true
	case passed:
		return "VERIFIED", "the chair recorded a PASS verdict", true
	}
	// CEILING IS THE DISPATCH PLAN'S (plans/roundless.md §III.B.2), for one of two reasons: every
	// open material gap is at impasse and has had its bench ruling — carried, since it is still
	// open — or the chair has sat for the run's last epoch under its epoch limit, a term setup
	// records. A record with no cast cannot reach it.
	if cast, err := CastOf(run); err == nil && cast != nil {
		plan, err := PlanDispatch(run)
		if err != nil {
			return "", "the record could not be read: " + err.Error(), false
		}
		switch {
		case plan.EpochLimitReached:
			return "CEILING", fmt.Sprintf("epoch limit %d reached — the run's term; the parties still ready were not dispatched and PASS is not permitted", plan.MaxEpochs), true
		case plan.Ceiling:
			return "CEILING", "every open material gap is at its limit and the bench has ruled on each — nobody is ready and PASS is not permitted", true
		}
	}
	return "", "no pass, no halt, and the board is not at its ceiling — the run ended before a terminal state was reached, and the record says so rather than guessing", false
}

// RunOutcomeOf is the seat's verdict word to the schema's value, and it lives beside DeriveVerdict
// for the reason GradeOf lives beside GradeStr: a conversion that exists in only one direction is
// how two vocabularies drift apart. The writer invents its own mapping, the reader keeps another,
// and nothing can see them disagree.
//
// The seat types `VERIFIED`; the schema spells `verified`. Case is presentation and is folded here
// rather than at each call site, because a caller that forgets returns the zero — and the zero is
// UNSPECIFIED, which would record a run as having no verdict at all rather than refusing the word.
// `false` means it is not a verdict; a caller must refuse rather than record the zero.
func RunOutcomeOf(word string) (recordpb.RunOutcome, bool) {
	vd, ok := recordpb.BySpelling(recordpb.RunOutcome(0).Descriptor(), strings.ToLower(strings.TrimSpace(word)))
	if !ok {
		return recordpb.RunOutcome_RUN_OUTCOME_UNSPECIFIED, false
	}
	return recordpb.RunOutcome(vd.Number()), true
}
