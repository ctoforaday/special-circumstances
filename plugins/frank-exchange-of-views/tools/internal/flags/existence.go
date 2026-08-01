package flags

import (
	"fmt"
	"slices"
	"strings"
)

// Existence is the leaf-check axis of a gap: was the defect VERIFIED present (checked
// at the leaf) or SUSPECTED (inferred)? It is orthogonal to likelihood, which grades
// the CONSEQUENCE. A two-value enum, parsed here so a bad value is refused where the
// mistake is made rather than landing an empty axis on the board (the board reads
// existence; before the write path this flag adds, nothing set it).
type ExistenceAxis string

const (
	Verified  ExistenceAxis = "verified"
	Suspected ExistenceAxis = "suspected"
)

// ExistenceAxes is the canonical set — the single membership authority.
var ExistenceAxes = []ExistenceAxis{Verified, Suspected}

// ExistenceValue is the pflag.Value implementation. Like GradeValue it is a POINTER so
// an unset flag stays distinguishable from a set one: an existence never passed must
// not appear in the event at all.
type ExistenceValue struct {
	Axis ExistenceAxis
	set  bool
}

func (v *ExistenceValue) String() string {
	if v == nil || !v.set {
		return ""
	}
	return string(v.Axis)
}

// Set parses and validates, naming what WOULD have worked — the seat's only teacher is
// the error text.
func (v *ExistenceValue) Set(s string) error {
	e := ExistenceAxis(strings.TrimSpace(s))
	if slices.Contains(ExistenceAxes, e) {
		v.Axis, v.set = e, true
		return nil
	}
	return fmt.Errorf("%q is not an existence axis — use one of: verified | suspected", s)
}

// Type is what pflag prints in usage: "--existence existence" tells the seat there is
// an enum behind it.
func (v *ExistenceValue) Type() string { return "existence" }

// IsSet reports whether the flag was given, so an unset existence stays out of the event.
func (v *ExistenceValue) IsSet() bool { return v != nil && v.set }
