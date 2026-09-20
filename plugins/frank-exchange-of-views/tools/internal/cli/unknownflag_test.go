package cli

import (
	"strings"
	"testing"
)

// A SEAT THAT TYPED THE RECORD'S WORD IS TOLD THE FLAG'S, and only when that is actually true.
//
// Measured on the 2026-09-20 smoke: a lens typed `--complexity_cost` at `mint` TWELVE times, each
// costing a turn. That is the key the record stores and shows the seat in a gap's own JSON before
// it mints, so the seat used the vocabulary it had just been handed — the most defensible mistake
// available. The translation was already in internal/flags and was consulted only when the TOOL
// named a field, never when a SEAT did.
func TestAnUnknownFlagNamesTheFlagTheSeatMeant(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-lens-evidence", "--complexity_cost", "high")
	if err == nil {
		t.Fatal("mint --complexity_cost was accepted; it is not a flag")
	}
	msg := err.Error() + out
	// THE ASSERTION MUST NOT BE SATISFIED BY COBRA'S OWN MESSAGE, and the obvious one is:
	// "unknown flag: --complexity_cost" already contains the substring "--complexity", because the
	// wrong word has the right word as a PREFIX. A mutation that deleted the translation entirely
	// passed against that check. What is pinned instead is the sentence only the translation can
	// produce.
	for _, want := range []string{
		"the flag that sets it is --complexity.",
		"the record stores that field as `complexity_cost`",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q — a seat is left to guess a second time:\n%s", want, msg)
		}
	}
}

// AND IT DOES NOT INVENT ONE. A key with no entry gets cobra's plain refusal: a suggested spelling
// the parser rejects is worse than no suggestion, which is the defect the payload map was built to
// stop in the first place.
func TestAnUnknownFlagWithNoTranslationIsNotGivenOne(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "mint", "--run", runDir, "--seat-id", "red-lens-evidence", "--not_a_real_field", "x")
	if err == nil {
		t.Fatal("an invented flag was accepted")
	}
	if strings.Contains(err.Error()+out, "the flag that sets it is") {
		t.Errorf("a flag with no known translation was given one anyway:\n%s", err.Error()+out)
	}
}

// AND THE SUGGESTION IS SCOPED TO THE FAILING VERB. The payload map is GLOBAL while payload keys
// are not globally unique — its own comment says so — so a suggestion is made only where the verb
// that refused actually carries that flag. `log` has no --complexity, and must not be told it does.
func TestTheSuggestionIsNotMadeForAFlagTheVerbDoesNotHave(t *testing.T) {
	runDir := seatRun(t)

	out, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-evidence", "--complexity_cost", "high")
	if err == nil {
		t.Fatal("log --complexity_cost was accepted")
	}
	if strings.Contains(err.Error()+out, "the flag that sets it is --complexity") {
		t.Errorf("log was told to type --complexity, which it does not have:\n%s", err.Error()+out)
	}
}
