package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
// function. A record that could not be read is the error, with ok false and no verdict beside
// it: it is never folded into "no terminal state", because a caller keying a liveness verdict or
// a server's lifetime to that answer would take a busy record for a run in flight. Everything
// else is already recorded:
//
//	HALTED    a halt event exists — the bench ended the run on its own authority
//	VERIFIED  the chair recorded a PASS verdict
//	CEILING   the dispatch plan's ceiling: the run's epoch limit is reached with parties ready
//
// The order matters: a halt outranks a pass, because a run stopped on safety or integrity
// grounds did not end by passing however clean the board looked when it stopped.
//
// ONE READ TRANSACTION, as a narrowed view reads (readSnapshot): the halt, the PASS, the cast and the
// dispatch plan are asked of one snapshot, so a halt and a PASS landing between two reads cannot
// derive VERIFIED over a record that holds the halt that outranks it. The transaction reads
// (verdictReadsAt); the verdict, and the plan it may fold, are derived after it closes (verdictOf).
func DeriveVerdict(run Run) (verdict, why string, ok bool, err error) {
	var r verdictReads
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = verdictReadsAt(q)
		return err
	}); err != nil {
		return "", "", false, err
	}
	return verdictOf(filesOf(run), r)
}

// verdictReads is every answer the derived verdict takes off the record, read on one snapshot.
// plan is read only where the verdict can turn on it: no halt, no PASS, and a cast.
type verdictReads struct {
	halted, passed bool
	cast           []string
	plan           *planReads
}

// verdictReadsAt reads the verdict's answers off q.
func verdictReadsAt(q recordsql.Querier) (verdictReads, error) {
	var r verdictReads
	var err error
	if r.halted, err = recordHasAt(q, `SELECT 1 FROM "halt" LIMIT 1`); err != nil {
		return r, err
	}
	if r.passed, err = recordHasAt(q, `SELECT 1 FROM "gate" WHERE "verdict" = ? LIMIT 1`,
		recordpb.Word(recordpb.Verdict_VERDICT_PASS)); err != nil {
		return r, err
	}
	if r.cast, err = castAt(q); err != nil {
		return r, err
	}
	if r.halted || r.passed || r.cast == nil {
		return r, nil
	}
	plan, err := planReadsAt(q)
	if err != nil {
		return r, err
	}
	r.plan = &plan
	return r, nil
}

// verdictOf derives the verdict from what verdictReadsAt read. It asks the record nothing; files is
// for the run's terms, read only when the verdict turns on the plan.
func verdictOf(files runFiles, r verdictReads) (verdict, why string, ok bool, err error) {
	halted, passed, cast := r.halted, r.passed, r.cast
	switch {
	case halted:
		return "HALTED", "a halt event is on the record", true, nil
	case passed:
		return "VERIFIED", "the chair recorded a PASS verdict", true, nil
	}
	// CEILING IS THE DISPATCH PLAN'S (plans/roundless.md §III.B.2): the chair has sat for the run's
	// last epoch under its epoch limit, a term setup records. A gap at its own limit is closed by the
	// bench, never left open (requireRemandWithinLimit). A record with no cast cannot reach it.
	standing := ""
	if cast != nil {
		params, err := files.params()
		if err != nil {
			return "", "", false, err
		}
		plan := foldPlan(params, *r.plan)
		// A RUN THAT ENDS OVER AN UNCHECKED OR FAILED CHECK NAMES IT. The run ends on its term or with
		// no terminal state over a citation or proof nobody has read against the sentence the report
		// carries, or one whose check failed, and the reason is where a reader of the outcome
		// learns which. No other blocker kind is named: the reason states why the run stopped, and the
		// board states what else is open.
		standing = anchorsStanding(plan.Blockers)
		// THE LIMIT IS READ AFTER THE EXIT TOO. The plan asks who is ready now; the terminal bench
		// has ruled since the chair's last sitting, and where the bench was all the limit left
		// undispatched, its ruling leaves nobody ready (epochLimitHeldTheBench).
		if plan.EpochLimitReached || epochLimitHeldTheBench(params, *r.plan) {
			return "CEILING", fmt.Sprintf("epoch limit %d reached — the run's term; the parties ready at the chair's last sitting were not dispatched and PASS is not permitted", plan.MaxEpochs) + standing, true, nil
		}
	}
	return "", "no pass, no halt, and the board is not at its ceiling — the run ended before a terminal state was reached, and the record says so rather than guessing" + standing, false, nil
}

// anchorsStanding is the outcome reason's clause naming each citation or proof that holds the PASS by
// its own state — unchecked, or on a failed check — or "" where none does.
func anchorsStanding(bs []PlanBlocker) string {
	var unchecked, failed []string
	for _, b := range bs {
		switch b.Kind {
		case BlockerUnchecked:
			unchecked = append(unchecked, b.Subject)
		case BlockerFailedCheck:
			if !slices.Contains(failed, b.Subject) {
				failed = append(failed, b.Subject)
			}
		}
	}
	out := ""
	if len(unchecked) > 0 {
		out += "; unchecked: " + strings.Join(unchecked, ", ")
	}
	if len(failed) > 0 {
		out += "; on a failed check: " + strings.Join(failed, ", ")
	}
	return out
}

// RunOutcomeOf is the seat's verdict word to the schema's value, and it lives beside DeriveVerdict
// for the reason GradeOf lives beside GradeStr: a conversion that exists in only one direction is
// how two vocabularies drift apart. The writer invents its own mapping, the reader keeps another,
// and nothing can see them disagree.
//
// The seat types `VERIFIED`; the schema spells `verified`. The surface's word is the schema's in
// capitals (loud), and it is the only one admitted: `verified` or ` VERIFIED ` is refused rather
// than folded. The conversion lives here rather than at each call site, because a caller that
// forgets returns the zero — and the zero is UNSPECIFIED, which would record a run as having no
// verdict at all rather than refusing the word. `false` means it is not a verdict; a caller must
// refuse rather than record the zero.
func RunOutcomeOf(word string) (recordpb.RunOutcome, bool) {
	if word != strings.ToUpper(word) {
		return recordpb.RunOutcome_RUN_OUTCOME_UNSPECIFIED, false
	}
	vd, ok := recordpb.BySpelling(recordpb.RunOutcome(0).Descriptor(), strings.ToLower(word))
	if !ok {
		return recordpb.RunOutcome_RUN_OUTCOME_UNSPECIFIED, false
	}
	return recordpb.RunOutcome(vd.Number()), true
}
