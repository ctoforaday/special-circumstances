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
// research/2026-09-02_quadratic-formula (blue-respond): red minted a punctuation repair with a
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

// A PUNCTUATION-ONLY REPAIR THAT LANDS NOWHERE, REPRODUCED FROM A RUN.
//
// #861's arm-B rerun: blue quoted `on their own.).` meaning to drop the final period. The trim left
// that period outside the span, the span became itself, and the verb reported the edit recorded
// with the report byte-identical. The refusal has to name the route that works, and that route
// has to work — so both are asserted.
func TestAnEditThatChangesNothingIsRefused(t *testing.T) {
	report := "on their own.).\n\nNext.\n"
	_, err := validateEdit(report, "on their own.).", "on their own.)")
	if err == nil {
		t.Fatal("an edit whose planned report equals the report was accepted — the verb would say " +
			"\"recorded\" over a document it did not change")
	}
	for _, want := range []string{"changes nothing", "trimmed", "THROUGH"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name the cause and the route (missing %q):\n%v", want, err)
		}
	}
	got, err := validateEdit(report, "on their own.).\n\nNext", "on their own.)\n\nNext")
	if err != nil {
		t.Fatalf("the route the refusal names was itself refused: %v", err)
	}
	if got != "on their own.)\n\nNext.\n" {
		t.Errorf("quoting through the terminator did not remove it: %q", got)
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
