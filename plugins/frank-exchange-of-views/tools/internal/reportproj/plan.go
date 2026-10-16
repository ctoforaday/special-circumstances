package reportproj

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
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
	TrailingOnly bool
}

func (e *NoChangeError) Error() string {
	msg := ErrNoChange.Error() + "."
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
// with `new` yields, the replacement it applied, and whether it took the literal span. It is the ONE
// decision `blue edit` makes before recording and `lens mint` makes before verifying a prescription,
// so red cannot prescribe what blue's edit would refuse or apply differently.
//
// The replacement applied is `new` with every anchor the transit check put back, and the caller
// records it, so replay splices the same bytes and needs no placement of its own.
//
// FIRST THE ORDINARY LOCATE, which trims the quote's trailing punctuation. Its result stands unless
// it misfires in one of the two ways the trim itself causes: the edit changes nothing, or it doubles
// a terminator at its seam (the old terminator standing beside the new — DoubledTerminator).
// Then, and only when the quote ENDS in punctuation the trim dropped, the quote AS WRITTEN is tried:
// it must occur byte-for-byte exactly once, carry its anchors unchanged, and change the report. If
// it does, exact is true and the caller records it — replay reads the field and never re-decides.
//
// A trimmed result that still doubles a terminator is RETURNED, not refused: that judgement is the
// caller's (blue refuses it; mint does not ask). Only a no-op is an error here, because a no-op is
// wrong for every caller — and an anchor the splice puts on a heading, which no placement may do.
func PlanSplice(report, old, new string) (next, applied string, exact bool, err error) {
	// An anchor on a heading in next that was on none in report is one this splice carried into a
	// heading or made a heading around. One already on a heading, an archived carry, is not this
	// edit's doing and refuses nothing, its heading rewritten or not.
	defer func() {
		was := headingAnchors(report)
		for _, id := range headingAnchors(next) {
			if err == nil && !slices.Contains(was, id) {
				next, applied, exact, err = "", "", false, fmt.Errorf("this edit puts %s on a section heading, and an anchor sits on prose, never on a heading: keep the anchor on its sentence in the section's text and write the heading without it", anchor.Token(id))
			}
		}
	}()
	start, end, err := bluedoc.LocateUniqueReplacing(report, old)
	if err != nil {
		return "", "", false, err
	}
	if applied, err = bluedoc.AnchorsTransitUnchanged(report[start:end], new); err != nil {
		// AN ANCHOR AT THE SPAN'S EDGE, TYPED INTO THE REPLACEMENT. The quote left it out, so it is
		// outside the span and stays where it stands; the flat "introduces" reads as a prohibition
		// to a seat that copied an anchor it can see against the sentence it is rewriting.
		var typed *bluedoc.ErrAnchorIntroduced
		if run, _ := bluedoc.AbuttingRun(report, end); errors.As(err, &typed) && strings.Contains(run, anchor.Token(typed.ID)) {
			err = fmt.Errorf("your replacement carries %s, and your quote stops just before it: that anchor stands at the edge of the span, outside it, and the edit leaves it exactly where it is — after your replacement, before the sentence's terminator. "+
				"Leave it out of the replacement. To put it somewhere else in the new text, quote it as `show report` prints it and write it where it goes", anchor.Token(typed.ID))
		}
		return "", "", false, err
	}
	next = ApplySplice(report, start, end, applied)
	if next != report && DoubledTerminator(report, next) == "" {
		return next, applied, false, nil
	}
	if endsInTrimmedPunct(old) {
		if ls, le, _, ok := bluedoc.LocateLiteral(report, old); ok {
			if lnew, err := bluedoc.AnchorsTransitUnchanged(report[ls:le], new); err == nil {
				if lit := ApplySplice(report, ls, le, lnew); lit != report && DoubledTerminator(report, lit) == "" {
					return lit, lnew, true, nil
				}
			}
		}
	}
	if next == report {
		return "", "", false, &NoChangeError{TrailingOnly: stripTail(old) == stripTail(new)}
	}
	return next, applied, false, nil
}

// headingAnchors is every anchor id on a heading line of s.
func headingAnchors(s string) (ids []string) {
	for _, b := range anchor.Blocks(s) {
		if b.Kind == anchor.Heading {
			ids = append(ids, anchor.IDs(s[b.Start:b.End])...)
		}
	}
	return ids
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

// DoubledTerminator names the punctuation run an edit leaves where a terminator lands beside another.
//
// MEASURED, and it is the shape this check exists for. In research/2026-09-02_quadratic-formula
// (blue-respond) red minted a punctuation repair with a `verified` fix basis, blue applied the
// text verbatim, and the site went from a doubled terminator `."."` to a TRIPLED one `."."."`.
// The same happened to two of blue's own edits in that sitting. All three were invisible until
// blue re-ran red's acceptance check against the shipped document — the verb exited 0 every time.
//
// It reads ONLY THE SEAMS the edit makes — where the changed text meets the text it left alone —
// because a terminator can land beside another only there. Punctuation inside the replacement is
// the author's prose: universe-m24's blue edit `For the query "is 91 a prime number", nine web
// pages come up` carries a `",` in a report holding no run that long, and is a rewording. The
// seams are where before and after stop agreeing, read from each end.
//
// A run at a seam doubles when BOTH sides of it carry a terminator (`!` against the report's `.`,
// `91:` against it) or both carry the same closing quote (`"goodbye"` against the report's `".`).
// A closing quote meeting a terminator is how a quoted phrase ends a sentence: the same sitting's
// `Nine web pages come up for the query "is 91 a prime number"` before the report's period leaves
// `".`, which is ordinary English.
//
// It reads the SHAPE OF THE SEAM rather than the document's longest run, so ordinary prose cannot
// trip it — a document may legitimately gain a "?!" or an ellipsis — and a run the document holds
// elsewhere cannot license it: in a report whose quoted phrase ends a sentence in `".`, a `!` landing
// beside a period is the same artifact as in one that quotes nothing. Anchors between two marks are
// not content between them, so the runs are read with every anchor out.
func DoubledTerminator(before, after string) string {
	if before == after {
		return ""
	}
	punct := func(b byte) bool { return strings.ContainsRune(anchortext.TrailingPunct, rune(b)) }
	// The changed span of after: [lo, hi), widened to whole anchor tokens so the parts strip cleanly.
	lo := 0
	for lo < len(before) && lo < len(after) && before[lo] == after[lo] {
		lo++
	}
	hi := len(after)
	for k := len(before); hi > lo && k > lo && before[k-1] == after[hi-1]; k-- {
		hi--
	}
	anchor.Each(after, func(start, end int, _ string) {
		if start < lo && lo < end {
			lo = start
		}
		if start < hi && hi < end {
			hi = end
		}
	})
	left, mid := claimcount.StripAnchors(after[:lo]), claimcount.StripAnchors(after[lo:hi])
	text := left + mid + claimcount.StripAnchors(after[hi:])
	doubles := func(a, b string) bool {
		terminates := func(s string) bool {
			return strings.ContainsFunc(s, func(r rune) bool { return r != '"' && r != '\'' })
		}
		return terminates(a) && terminates(b) || strings.ContainsRune(a, '"') && strings.ContainsRune(b, '"') || strings.ContainsRune(a, '\'') && strings.ContainsRune(b, '\'')
	}
	for _, at := range []int{len(left), len(left) + len(mid)} {
		i, j := at, at
		for i > 0 && punct(text[i-1]) {
			i--
		}
		for j < len(text) && punct(text[j]) {
			j++
		}
		if doubles(text[i:at], text[at:j]) {
			return text[i:j]
		}
	}
	return ""
}
