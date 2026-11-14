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
// basis from running the program twice, applied_verbatim from comparing bytes — or, on
// `blue edit --accept`, from the tool having supplied those bytes itself. Each
// exists because a seat asked to self-report reports the flattering value.
//
// The terminal verdict — the single most consequential value the engine emits, the one the
// report stamps at the top and every audit and scorecard reads — was ASSERTED. debate.js
// computed it in JS, stated it in the assembler's prompt, and the seat typed it back. Nothing
// checked the two agreed, so `--as VERIFIED` over a board with open gaps would have been
// recorded and believed.

// VerdictBasis says how far a recorded verdict can be trusted, on the same footing as
// fix_basis and a computation's basis.
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
//	HALTED     a halt event exists — the bench ended the run on its own authority
//	VERIFIED   the chair recorded a PASS verdict, and every question part is covered
//	QUALIFIED  the chair recorded a PASS verdict, and a question part's coverage is not covered — its
//	           sources were not obtained, or it is excluded from the report's answer
//	CEILING    the dispatch plan's ceiling: the run's epoch limit is reached with parties ready
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
	// parts is every question part, read where the chair recorded a PASS: which parts are not covered
	// decides VERIFIED or QUALIFIED.
	parts []QuestionPartRow
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
	if r.passed && !r.halted {
		r.parts, err = questionPartsAt(q)
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
		// A PASS OVER A PART NOT COVERED IS QUALIFIED (ruled by gblock 2026-11-12, F3). The gate refuses
		// the PASS while the question has no part or a part has no coverage or no standing check, so
		// every part a live PASS stands over is checked. An archived record holds no part, and its PASS
		// derives VERIFIED as it was stamped (F10: archived outcomes are carried as history).
		if why := notCoveredWhy(r.parts); why != "" {
			return "QUALIFIED", "the chair recorded a PASS verdict, and " + why, true, nil
		}
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
		// no terminal state over a citation or computation nobody has read against the sentence the report
		// carries, or one whose check failed, or over an inquiry still open or unchecked, and the reason
		// is where a reader of the outcome learns which. No other blocker kind is named: the reason
		// states why the run stopped, and the board states what else is open.
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

// StandingClause is one kind of blocker that holds a PASS by its own state, and the clause the outcome
// reason names it with.
type StandingClause struct {
	Kind   BlockerKind
	Clause string
}

// StandingClauses is the ONE table of the kinds a run that ends names in its reason — the citation or
// computation nobody checked, the failed check, the lens's reading debt, the inquiry open or
// unchecked, the question with no part split out, the question part with no coverage or no standing check
// — in kind-table order, each with its clause; a kind about the run as a whole names no subject. anchorsStanding folds it for the record's
// derived reason, and enginegen renders it into debate.js's generated vocabulary as STANDING_CLAUSES,
// which the engine's termination reason folds: two carriers, one declaration, so the record's reason
// and the engine's cannot name different kinds.
func StandingClauses() []StandingClause {
	return []StandingClause{
		{BlockerUnchecked, "unchecked"},
		{BlockerFailedCheck, "on a failed check"},
		{BlockerReadingDebt, "on a reading debt"},
		{BlockerOpenInquiry, "inquiries open"},
		{BlockerUncheckedInquiry, "inquiries unchecked"},
		{BlockerQuestionUnsplit, "the question with no part split out"},
		{BlockerOpenPart, "question parts with no coverage"},
		{BlockerUncheckedPart, "question parts unchecked"},
	}
}

// notCoveredWhy names each question part whose latest coverage is not covered, with its coverage —
// "" where every part is covered or there is none.
func notCoveredWhy(parts []QuestionPartRow) string {
	var named []string
	for _, p := range parts {
		if p.Coverage != "" && !p.Covered() {
			named = append(named, fmt.Sprintf("%s (%s)", p.ID, p.Coverage))
		}
	}
	if len(named) == 0 {
		return ""
	}
	return "question part(s) not covered: " + strings.Join(named, ", ")
}

// anchorsStanding is the outcome reason's clause naming each subject that holds the PASS by its own
// state — one clause per kind of StandingClauses, its subjects deduplicated, in table order — or ""
// where none does.
func anchorsStanding(bs []PlanBlocker) string {
	out := ""
	for _, c := range StandingClauses() {
		var subjects []string
		for _, b := range bs {
			if b.Kind == c.Kind && !slices.Contains(subjects, b.Subject) {
				subjects = append(subjects, b.Subject)
			}
		}
		switch {
		case len(subjects) == 1 && subjects[0] == "":
			// A KIND ABOUT THE RUN AS A WHOLE names no subject, and its clause stands alone.
			out += "; " + c.Clause
		case len(subjects) > 0:
			out += "; " + c.Clause + ": " + strings.Join(subjects, ", ")
		}
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
