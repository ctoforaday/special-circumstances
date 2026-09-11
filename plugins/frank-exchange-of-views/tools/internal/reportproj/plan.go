package reportproj

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
)

// ErrNoChange is the sentinel for an edit whose planned report equals the report. `blue edit`
// refuses it, `lens mint` refuses to verify a prescription that is one, and `blue edit --accept`
// tells the two apart from a stale prescription with errors.Is.
var ErrNoChange = errors.New("this edit changes nothing — the report would read exactly as it does now")

// NoChangeError is ErrNoChange with what the refusal can say about WHY.
//
// TrailingOnly is set when the quote and the replacement differ only in trailing punctuation or
// whitespace — the one shape where the trim is the cause and naming it is the advice. Any other
// no-op (a whitespace-only difference the locate folds away, a seam the tidy repairs back, an
// anchor-only difference) gets the plain statement, because blaming the trim there would send the
// seat after a cause it does not have.
type NoChangeError struct {
	Verb         string
	TrailingOnly bool
}

func (e *NoChangeError) Error() string {
	msg := fmt.Sprintf("%s: %v.", e.Verb, ErrNoChange)
	if !e.TrailingOnly {
		return msg
	}
	// The literal quote failed to stand in. It cannot have been for occurring twice — a repeat as
	// written is a repeat to the ordinary locate too, refused as ambiguous before this — so what
	// is left is that it is not in the report byte-for-byte.
	return msg + " A quote's TRAILING punctuation is trimmed before the span is located, so a change confined to that punctuation " +
		"lands on the span without it. The quote as written, punctuation included, is used instead when it occurs byte-for-byte " +
		"exactly once in the report, and this one does not: quote it exactly as `show report` prints it, anchors and spacing included"
}

func (e *NoChangeError) Is(target error) bool { return target == ErrNoChange }

// PlanSplice locates `old` in report the way an edit does and returns the report that replacing it
// with `new` yields, and whether it took the literal span. It is the ONE decision `blue edit` makes
// before recording and `lens mint` makes before verifying a prescription, so red cannot prescribe
// what blue's edit would refuse or apply differently.
//
// FIRST THE ORDINARY LOCATE, which trims the quote's trailing punctuation. Its result stands unless
// it misfires in one of the two ways the trim itself causes: the edit changes nothing, or it leaves
// a punctuation run longer than any the report had (the old terminator standing beside the new).
// Then, and only when the quote ENDS in punctuation the trim dropped, the quote AS WRITTEN is tried:
// it must occur byte-for-byte exactly once, carry its anchors unchanged, and change the report. If
// it does, exact is true and the caller records it — replay reads the field and never re-decides.
//
// A trimmed result that still doubles a terminator is RETURNED, not refused: that judgement is the
// caller's (blue refuses it; mint does not ask). Only a no-op is an error here, because a no-op is
// wrong for every caller.
func PlanSplice(verb, report, old, new string) (next string, exact bool, err error) {
	start, end, err := bluedoc.LocateUniqueReplacing(verb, report, old)
	if err != nil {
		return "", false, err
	}
	if err := bluedoc.AnchorsTransitUnchanged(verb, report[start:end], new); err != nil {
		return "", false, err
	}
	next = ApplySplice(report, start, end, new)
	if next != report && DoubledTerminator(report, next) == "" {
		return next, false, nil
	}
	if endsInTrimmedPunct(old) {
		ls, le, _, ok := bluedoc.LocateLiteral(report, old)
		if ok && bluedoc.AnchorsTransitUnchanged(verb, report[ls:le], new) == nil {
			if lit := ApplySplice(report, ls, le, new); lit != report && DoubledTerminator(report, lit) == "" {
				return lit, true, nil
			}
		}
	}
	if next == report {
		return "", false, &NoChangeError{Verb: verb, TrailingOnly: stripTail(old) == stripTail(new)}
	}
	return next, false, nil
}

// endsInTrimmedPunct says whether the ordinary locate dropped punctuation off the end of this quote —
// the only case in which the literal span differs from the located one in the part being edited.
func endsInTrimmedPunct(q string) bool {
	body := strings.TrimRight(q, " \t\r\n")
	return strings.TrimRight(body, anchortext.TrailingPunct) != body
}

// stripTail is a quote with its trailing punctuation and whitespace gone — what the locate matches on.
func stripTail(s string) string {
	return strings.TrimRight(s, anchortext.TrailingPunct+" \t\r\n")
}

// DoubledTerminator names a punctuation run the edit would CREATE and the report did not have.
//
// MEASURED, and it is the shape this check exists for. In research/2026-09-02_quadratic-formula
// (blue-respond) red minted a punctuation repair with a `verified` fix basis, blue applied the
// text verbatim, and the site went from a doubled terminator `."."` to a TRIPLED one `."."."`.
// The same happened to two of blue's own edits in that sitting. All three were invisible until
// blue re-ran red's acceptance check against the shipped document — the verb exited 0 every time.
//
// It compares RUNS RATHER THAN COUNTS so ordinary prose cannot trip it: a document may legitimately
// gain a "?!" or an ellipsis. What it refuses is a run LONGER than any the document already had,
// which is the signature of a terminator landing beside one rather than on top of it.
func DoubledTerminator(before, after string) string {
	worst := func(s string) string {
		longest := ""
		for i := 0; i < len(s); {
			j := i
			for j < len(s) && strings.ContainsRune(anchortext.TrailingPunct, rune(s[j])) {
				j++
			}
			if j-i > len(longest) {
				longest = s[i:j]
			}
			if j == i {
				i++
			} else {
				i = j
			}
		}
		return longest
	}
	got := worst(after)
	if len(got) > len(worst(before)) {
		return got
	}
	return ""
}
