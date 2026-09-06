package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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
	// VerdictAsserted: the record cannot decide it, so the seat's word stands. Today that is
	// exactly one case — a judged deadlock — and it is unfalsifiable for a nameable reason:
	// the bench's determination lives only in its envelope and leaves no independent trace.
	// See #289; when the bench's terminal call becomes a recorded act, this constant loses its
	// last user.
	VerdictAsserted = "asserted"
)

// DeriveVerdict computes a run's terminal verdict from the record alone.
//
// ok is false ONLY when the record genuinely cannot decide — which is the deadlock case, and
// is a finding rather than a defect in this function. Everything else is already recorded:
//
//	HALTED    a halt event exists — the bench ended the run on its own authority
//	VERIFIED  the merge recorded a PASS verdict
//	CEILING   the rounds on the record reached the ceiling in inputs/run-config.json
//
// The order matters: a halt outranks a pass, because a run stopped on safety or integrity
// grounds did not end by passing however clean the board looked when it stopped.
func DeriveVerdict(run Run) (verdict, why string, ok bool) {
	// Three questions, three queries — a halt on the record, a recorded PASS, and how far the
	// rounds got. The body-less-verdict rule holds by construction now: the PASS lives in the
	// round_verdict row, and an event with no body row is refused by the loader, not read as
	// anything.
	halted, err := recordHas(run, `SELECT 1 FROM "halt" LIMIT 1`)
	if err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	passed, err := recordHas(run, `SELECT 1 FROM "round_verdict" WHERE "verdict" = ? LIMIT 1`,
		recordpb.Word(recordpb.Verdict_VERDICT_PASS))
	if err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	var maxRound int
	if _, err := queryRow(run, []any{&maxRound}, `SELECT COALESCE(max("round"), 0) FROM "events"`); err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	switch {
	case halted:
		return "HALTED", "a halt event is on the record", true
	case passed:
		return "VERIFIED", "the merge recorded a PASS verdict", true
	}
	if ceiling := configuredMaxRounds(run); ceiling > 0 && maxRound >= ceiling {
		return "CEILING", "the record reaches round " + strconv.Itoa(maxRound) + " against a ceiling of " + strconv.Itoa(ceiling), true
	}
	// No pass, no halt, and the ceiling not reached: the run ended early, which the engine
	// only does on a judged deadlock. That judgement is not on the record — the bench's
	// determination lives in its envelope and nothing preserves it (#289) — so the tool
	// cannot confirm or refute the seat here, and says so rather than guessing.
	return "", "no pass, no halt, and the round ceiling was not reached — the run ended early, which only a judged deadlock does, and that determination is not on the record (#289)", false
}

// configuredMaxRounds reads the ceiling setup recorded; 0 when it is absent or unparseable,
// which degrades CEILING to underivable rather than inventing a bound.
func configuredMaxRounds(run Run) int {
	b, err := os.ReadFile(filepath.Join(run.Dir(), "inputs", "run-config.json"))
	if err != nil {
		return 0
	}
	var rc struct {
		MaxRounds *string `json:"maxRounds"`
	}
	if json.Unmarshal(b, &rc) != nil || rc.MaxRounds == nil {
		return 0
	}
	n, err := strconv.Atoi(*rc.MaxRounds)
	if err != nil {
		return 0
	}
	return n
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
