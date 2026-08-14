package flags

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TYPED FLAGS FOR VALUES WITH A KNOWABLE SHAPE.
//
// `GradeValue` proved the pattern: a pflag.Value refuses a wrong value AT PARSE, before any RunE
// runs, with the help and the refusal generated from one list. Everything else that has a shape
// was registered as a bare string and checked — or not checked — somewhere downstream.
//
// A 2026-08-13 sweep of every flag in the tree found the gap is not uniform. Some values were
// cross-checked properly (`--found-by` against the findings on the record, `--supersedes` against
// the board, `--fix-old` against the report); others were accepted raw and read back by someone
// who assumed they meant something.
//
// # Shape here, existence at the write path — and the line is not arbitrary
//
// A pflag.Value sees ONE STRING. It does not know the run directory, so it cannot ask whether
// gap R3-7 exists, whether c-1a2b names a citation, or whether a quoted sentence appears in the
// report. Those are RECORD questions and they belong in record.validate, which is the single
// write path every caller goes through.
//
// What a flag type CAN do is refuse a value that could never be right whatever the record says:
// `R3` is not a gap id, `banana` is not an anchor, `last tuesday` is not a date. That is worth
// doing at the flag because the refusal arrives with the usage line attached, and because a
// malformed id reaching validate produces "no such gap R3" — which reads as a missing gap rather
// than a typo, and sends a seat looking for the wrong thing.
//
// So: shape is refused here, existence is refused there, and neither pretends to be the other.
// referencechecks_test.go asserts the second half — that every flag naming an entity is actually
// checked against the record — because a shape check that looked like a reference check would be
// the more dangerous half-measure.

// gapIDShape is R<round>-<n>, the id `MintGapID` assigns.
var gapIDShape = regexp.MustCompile(`^R\d+-\d+$`)

// anchorShape is the tool-inserted anchor id: f- a finding, c- a source, p- a computation. The
// prefix carries the class, which is why a bare hex string is not one.
var anchorShape = regexp.MustCompile(`^[fcp]-[0-9a-f]+$`)

// findingLabelShape is L<lens>-F<n>, the run-unique label the tool assigns a lens finding.
var findingLabelShape = regexp.MustCompile(`^L\d+-F\d+$`)

// motionIDShape is M<n>.
var motionIDShape = regexp.MustCompile(`^M\d+$`)

// shaShape is a sha256 in hex, the handle `blue prove` prints and `lens reproduce --id` takes.
var shaShape = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ShapedValue is a pflag.Value that refuses a value whose FORM is wrong, whatever the record
// holds. One type, parameterised by the shape, so a new kind of id is a table entry rather than
// another near-copy of GradeValue.
type ShapedValue struct {
	kind string // what pflag prints as the flag's type, and what the refusal names
	re   *regexp.Regexp
	hint string // what a right one looks like, in the seat's terms
	val  string
	set  bool
}

func (v *ShapedValue) String() string {
	if v == nil || !v.set {
		return ""
	}
	return v.val
}

func (v *ShapedValue) Set(s string) error {
	t := strings.TrimSpace(s)
	if v.re != nil && !v.re.MatchString(t) {
		// The message names what WOULD have worked and where to get one. A seat that mistypes an
		// id is a seat that does not have the id, and telling it the shape without telling it the
		// read leaves it to guess twice.
		return fmt.Errorf("%q is not %s — %s", s, v.kind, v.hint)
	}
	v.val, v.set = t, true
	return nil
}

func (v *ShapedValue) Type() string { return v.kind }

// GapID refuses anything that is not R<round>-<n>.
func GapID() *ShapedValue {
	return &ShapedValue{kind: "gap-id", re: gapIDShape,
		hint: "a gap id looks like R2-3 (round, then the number the mint returned); `show board` and `show worklist` list them"}
}

// AnchorID refuses anything that is not a tool-inserted anchor id.
func AnchorID() *ShapedValue {
	return &ShapedValue{kind: "anchor", re: anchorShape,
		hint: "an anchor is the id inside a `<!--cite:c-…-->`, `<!--fx:f-…-->` or `<!--proof:p-…-->` token in the report; `show evidence` and `show findings` resolve them"}
}

// FindingLabel refuses anything that is not L<lens>-F<n>.
func FindingLabel() *ShapedValue {
	return &ShapedValue{kind: "finding-label", re: findingLabelShape,
		hint: "a finding label looks like L1-F2 (the lens, then its finding number) and is ASSIGNED by `lens finding`; `show findings` lists them"}
}

// MotionID refuses anything that is not M<n>.
func MotionID() *ShapedValue {
	return &ShapedValue{kind: "motion-id", re: motionIDShape,
		hint: "a motion id looks like M1 and is assigned when the motion is filed; `show motions` lists them with what each one asks"}
}

// SHA refuses anything that is not a 64-character hex digest.
func SHA() *ShapedValue {
	return &ShapedValue{kind: "sha256", re: shaShape,
		hint: "a proof's sha256 is 64 hex characters; `show evidence` lists every proof with its anchor, its sha and whether anyone has re-run it"}
}

// DateValue refuses a date that is not YYYY-MM-DD, and one that is not a real day.
//
// `--access-date` drives the staleness re-fetch trigger: a claim verified at high confidence
// stays verified unless more than two rounds have elapsed or the recorded date suggests drift. A
// date nothing can parse silently disables that trigger for the row, which is the quiet failure
// this whole sweep is about — the reader gets a value, believes it, and computes nothing.
type DateValue struct {
	val string
	set bool
}

func (v *DateValue) String() string {
	if v == nil || !v.set {
		return ""
	}
	return v.val
}

func (v *DateValue) Set(s string) error {
	t := strings.TrimSpace(s)
	if _, err := time.Parse("2006-01-02", t); err != nil {
		return fmt.Errorf("%q is not a date — use YYYY-MM-DD, the day you actually fetched it (it drives the staleness re-fetch trigger, so a value nothing can parse turns that check off for this row)", s)
	}
	v.val, v.set = t, true
	return nil
}

func (v *DateValue) Type() string { return "YYYY-MM-DD" }
