package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// THE VERDICT IS THE LAST BIG DERIVED-NOT-ASSERTED VIOLATION (#308).
//
// This engine computes the facts a seat would benefit from claiming, precisely so it cannot
// claim them: fix_basis is derived from whether the pair validates against the live report,
// proof_basis from running the script twice, applied_verbatim from comparing bytes. Each
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
func DeriveVerdict(runDir string) (verdict, why string, ok bool) {
	b, err := BoardState(runDir)
	if err != nil {
		return "", "the record could not be read: " + err.Error(), false
	}
	maxRound := 0
	passed := false
	halted := false
	for _, e := range b.Events {
		if e.Round > maxRound {
			maxRound = e.Round
		}
		switch e.Type {
		case "halt":
			halted = true
		case "verdict":
			if e.Payload.Str("verdict") == "PASS" {
				passed = true
			}
		}
	}
	switch {
	case halted:
		return "HALTED", "a halt event is on the record", true
	case passed:
		return "VERIFIED", "the merge recorded a PASS verdict", true
	}
	if ceiling := configuredMaxRounds(runDir); ceiling > 0 && maxRound >= ceiling {
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
func configuredMaxRounds(runDir string) int {
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
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
