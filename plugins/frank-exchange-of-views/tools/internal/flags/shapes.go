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

// Checker resolves a flag's value against the record. It is what a shaped flag carries so the
// EXISTENCE half travels with the declaration instead of being remembered separately.
//
// The signature takes the run directory because that is the one thing a pflag.Value cannot see:
// parsing is a single pass in argv order, so `--id` may be Set before `--run` ever is. The check
// therefore runs after parse — see seat.CheckFlagReferences — and the flag's job is to CARRY it,
// not to run it.
type Checker func(runDir, value string) error

// ShapedValue is a pflag.Value that refuses a value whose FORM is wrong, whatever the record
// holds — and carries the check for whether the record actually has it.
//
// One type, parameterised by shape and checker, so a new kind of id is a constructor call rather
// than another near-copy of GradeValue plus an entry in a list somebody has to maintain.
type ShapedValue struct {
	kind  string // what pflag prints as the flag's type, and what the refusal names
	re    *regexp.Regexp
	hint  string // what a right one looks like, in the seat's terms
	check Checker
	val   string
	set   bool
}

// Check resolves this flag's value against the record, or returns nil when the flag was not
// passed or carries no checker. The value is checked ONLY when set: an absent optional reference
// is not a dangling one.
func (v *ShapedValue) Check(runDir string) error {
	if v == nil || !v.set || v.check == nil {
		return nil
	}
	return v.check(runDir, v.val)
}

// WithCheck attaches the existence check. Separate from the constructor so the shape constructors
// stay usable where only the form matters, and so the call site reads as one statement of what
// the flag is: `flags.GapID().WithCheck(record.GapExists)`.
func (v *ShapedValue) WithCheck(c Checker) *ShapedValue { v.check = c; return v }

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

// AnchorID refuses anything that is not a tool-inserted anchor id of any class.
func AnchorID() *ShapedValue {
	return &ShapedValue{kind: "anchor", re: anchorShape,
		hint: "an anchor is the id inside a `<!--cite:c-…-->`, `<!--fx:f-…-->` or `<!--proof:p-…-->` token in the report; `show evidence` and `show findings` resolve them"}
}

// citationAnchorShape is the CITATION class only.
var citationAnchorShape = regexp.MustCompile(`^c-[0-9a-f]+$`)

// CitationAnchor refuses an anchor of the wrong CLASS, not merely the wrong form.
//
// `blue prove --cites` and `lens verify --anchor` both name a source, and the general anchor
// shape would accept `f-…` (a finding) or `p-…` (a computation) — well-formed ids that cannot
// possibly be citations. The prefix carries the class precisely so a reader never has to guess
// which kind of thing an id is; a flag that accepts all three throws that away and defers the
// error to a record lookup whose message is about existence rather than kind.
func CitationAnchor() *ShapedValue {
	return &ShapedValue{kind: "citation-anchor", re: citationAnchorShape,
		hint: "a citation anchor is the c-<hex> inside a `<!--cite:c-…-->` token in the report — `f-` is a finding and `p-` is a computation, neither of which is a source; `show evidence` lists every citation by anchor"}
}

// avenueIDShape is A<n>, the id assigned when a line of inquiry is proposed.
var avenueIDShape = regexp.MustCompile(`^A\d+$`)

// AvenueID refuses anything that is not A<n>.
func AvenueID() *ShapedValue {
	return &ShapedValue{kind: "avenue-id", re: avenueIDShape,
		hint: "an avenue id looks like A1 and is ASSIGNED when you propose the line; `show lines-of-inquiry` lists every one with its fate"}
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
