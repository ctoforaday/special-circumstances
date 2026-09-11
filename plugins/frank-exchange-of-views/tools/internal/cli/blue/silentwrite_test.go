package blue

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE REPAIR THAT MADE THE DOCUMENT WORSE, REPRODUCED.
//
// A quote's TRAILING punctuation is trimmed before the span is located, so a quote naming a
// sentence stops SHORT of its terminator. Replacing that span with text carrying its own
// terminator leaves the original standing after it. Measured in
// research/2026-09-02_quadratic-formula (blue-respond): red minted a punctuation repair with a
// `verified` fix basis, blue applied it verbatim, and a doubled terminator became a TRIPLED one.
// The verb exited 0. It was invisible until the acceptance check was re-run against the document.
//
// The literal quote now stands in when it occurs as written (next test), so the refusal is what is
// left when it does not: here the report's spacing differs from the quote's, which the ordinary
// locate folds away and a byte-for-byte match cannot.
func TestAnEditThatWouldDoubleATerminatorIsRefused(t *testing.T) {
	report := `The count  is 27."` + "\n"
	_, _, err := validateEdit(report, `The count is 27."`, `The count is 28."`)
	if err == nil {
		t.Fatal("an edit that would leave a punctuation run the document did not have was accepted — " +
			"this is the shape that turned a doubled terminator into a tripled one, at exit 0")
	}
	// A REFUSAL THAT DOES NOT EXPLAIN SENDS THE SEAT IN A CIRCLE: the same text, refused again.
	for _, want := range []string{"trimmed", "--quote", "exactly once"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not tell the seat how to fix it (missing %q):\n%v", want, err)
		}
	}
}

// THE SAME REPAIR, QUOTED AS WRITTEN, LANDS. The quote occurs byte-for-byte once, so the literal
// span — terminator included — is replaced, and the edit says so for the record.
func TestAChangedTerminatorTakesTheLiteralSpan(t *testing.T) {
	report := `The count is 27."` + "\n"
	got, exact, err := validateEdit(report, `The count is 27."`, `The count is 28."`)
	if err != nil {
		t.Fatalf("a terminator repair quoted exactly as written was refused: %v", err)
	}
	if got != `The count is 28."`+"\n" || !exact {
		t.Errorf("got %q exact=%v, want the terminator replaced on the literal span", got, exact)
	}
	// `It works.` → `It works!` at the END of the document: there is no text after the terminator
	// to quote through, so the literal span is the only route there is.
	got, exact, err = validateEdit("Intro.\n\nIt works.", "It works.", "It works!")
	if err != nil || got != "Intro.\n\nIt works!" || !exact {
		t.Errorf("a terminator change at the end of the document: got %q exact=%v err=%v", got, exact, err)
	}
}

// A PUNCTUATION-ONLY REPAIR THAT USED TO LAND NOWHERE, REPRODUCED FROM A RUN.
//
// #861's arm-B rerun: blue quoted `on their own.).` meaning to drop the final period. The trim left
// that period outside the span, the span became itself, and the verb reported the edit recorded
// with the report byte-identical. It now takes the literal span — at the END of the document too,
// where quoting through the terminator into the next text is impossible because there is none.
func TestAPunctuationOnlyRepairTakesTheLiteralSpan(t *testing.T) {
	for _, tc := range []struct{ name, report, quote, want string }{
		{"end of document", "Intro.\n\non their own.).\n", "on their own.).", "Intro.\n\non their own.)\n"},
		{"end of document, no newline", "Intro.\n\non their own.).", "on their own.).", "Intro.\n\non their own.)"},
		{"mid-document", "on their own.).\n\nNext.\n", "on their own.).", "on their own.)\n\nNext.\n"},
		// Copied with its newline: the whitespace around a quote is not part of its span, so the
		// line keeps its break rather than being joined to the next.
		{"quote copied with its newline", "on their own.).\nNext.\n", "on their own.).\n", "on their own.)\nNext.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, exact, err := validateEdit(tc.report, tc.quote, "on their own.)")
			if err != nil {
				t.Fatalf("the repair was refused: %v", err)
			}
			if got != tc.want || !exact {
				t.Errorf("got %q exact=%v, want %q on the literal span", got, exact, tc.want)
			}
		})
	}
	// Quoting through the terminator still works, and needs no literal span.
	got, exact, err := validateEdit("on their own.).\n\nNext.\n", "on their own.).\n\nNext", "on their own.)\n\nNext")
	if err != nil || got != "on their own.)\n\nNext.\n" || exact {
		t.Errorf("quoting through the terminator: got %q exact=%v err=%v", got, exact, err)
	}
}

// AN EDIT THAT STILL CHANGES NOTHING IS REFUSED, and the refusal blames the trim only when the trim
// is the cause. A whitespace-only difference the locate folds away is a no-op with no punctuation
// in it; telling that seat about trailing punctuation sends it after a cause it does not have.
func TestAnEditThatChangesNothingIsRefused(t *testing.T) {
	_, _, err := validateEdit("on their own.\n", "on  their own", "on their own")
	if err == nil {
		t.Fatal("a whitespace-only no-op was accepted — the verb would say \"recorded\" over a document it did not change")
	}
	if !errors.Is(err, reportproj.ErrNoChange) || !strings.Contains(err.Error(), "changes nothing") {
		t.Errorf("the refusal is not the no-change refusal: %v", err)
	}
	if strings.Contains(err.Error(), "trimmed") {
		t.Errorf("a no-op with no punctuation in it was blamed on the punctuation trim: %v", err)
	}

	// The trim IS the cause here, and the literal quote cannot stand in — its trailing punctuation is
	// not the report's, so it is nowhere in the report as written — so the refusal names both.
	_, _, err = validateEdit("Intro.\n\non their own.).\n", "on their own.);", "on their own.)")
	if err == nil {
		t.Fatal("a punctuation-only no-op whose literal quote is not in the report was accepted")
	}
	for _, want := range []string{"changes nothing", "trimmed", "byte-for-byte"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name the cause and the route (missing %q):\n%v", want, err)
		}
	}
}

// AND ORDINARY PROSE STILL EDITS, on the ordinary span. The check compares punctuation RUNS rather
// than counts, so a document may gain a question mark or an ellipsis; what it refuses is a run
// LONGER than any the document already had, which is the signature of a terminator landing beside one.
func TestAnOrdinaryEditIsNotRefused(t *testing.T) {
	report := "The count is 27.\n"
	got, exact, err := validateEdit(report, "The count is 27.", "The count is 28.")
	if err != nil {
		t.Fatalf("an ordinary replacement was refused: %v", err)
	}
	if !strings.Contains(got, "28") || exact {
		t.Errorf("the replacement did not land on the ordinary span: %q exact=%v", got, exact)
	}
}
