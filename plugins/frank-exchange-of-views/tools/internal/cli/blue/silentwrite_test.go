package blue

import (
	"strings"
	"testing"
)

// THE REPAIR THAT MADE THE DOCUMENT WORSE, REPRODUCED.
//
// A quote's TRAILING punctuation is trimmed before the span is located, so `--old` naming a
// sentence stops SHORT of its terminator. Replacing that span with text carrying its own
// terminator leaves the original standing after it. Measured in
// research/2026-09-02_quadratic-formula (blue-respond-r2): red minted a punctuation repair with a
// `verified` fix basis, blue applied it verbatim, and a doubled terminator became a TRIPLED one.
// The verb exited 0. It was invisible until the acceptance check was re-run against the document.
func TestAnEditThatWouldDoubleATerminatorIsRefused(t *testing.T) {
	report := `The count is 27."` + "\n"
	// The seat means to replace the sentence and supplies its own terminator; the span located
	// stops before the report's, so the naive result is `The count is 28."."`.
	_, err := validateEdit(report, `The count is 27."`, `The count is 28."`)
	if err == nil {
		t.Fatal("an edit that would leave a punctuation run the document did not have was accepted — " +
			"this is the shape that turned a doubled terminator into a tripled one, at exit 0")
	}
	// A REFUSAL THAT DOES NOT EXPLAIN SENDS THE SEAT IN A CIRCLE: the same text, refused again.
	for _, want := range []string{"trimmed", "--old"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not tell the seat how to fix it (missing %q):\n%v", want, err)
		}
	}
}

// AND ORDINARY PROSE STILL EDITS. The check compares punctuation RUNS rather than counts, so a
// document may gain a question mark or an ellipsis; what it refuses is a run LONGER than any the
// document already had, which is the signature of a terminator landing beside one.
func TestAnOrdinaryEditIsNotRefused(t *testing.T) {
	report := "The count is 27.\n"
	got, err := validateEdit(report, "The count is 27.", "The count is 28.")
	if err != nil {
		t.Fatalf("an ordinary replacement was refused: %v", err)
	}
	if !strings.Contains(got, "28") {
		t.Errorf("the replacement did not land: %q", got)
	}
}
