// Package flags carries the CLI's non-trivial flag types: values that are more
// than a string and deserve parsing, validation, and an error message of their
// own rather than being checked three layers down.
//
// Grade is the case that motivated the package. It is an eight-value enum on a
// single scale, it appears on six flags across four verbs, and until now it was
// a bare string validated inside the domain at APPEND time — so a seat that
// mistyped a grade got its refusal after argument parsing, after the payload was
// built, and phrased as an internal field error: mint.severity="catastrophic"
// not in grade enum. Parsing it here moves the refusal to where the mistake was
// made and lets the message name the values that would have worked.
package flags

import (
	"fmt"
	"strings"
)

// Grade is a point on the shared grading scale. Order is severity-ascending and
// mirrors the domain's own enum, including the two off-scale values: realized
// (it already happened) and trivial (below the bar but real).
type Grade string

const (
	Low        Grade = "low"
	LowMedium  Grade = "low-medium"
	Medium     Grade = "medium"
	MediumHigh Grade = "medium-high"
	High       Grade = "high"
	Certain    Grade = "certain"
	Realized   Grade = "realized"
	Trivial    Grade = "trivial"
)

// Grades is the canonical order, used for both validation and the help text so
// the two cannot drift.
var Grades = []Grade{Low, LowMedium, Medium, MediumHigh, High, Certain, Realized, Trivial}

func GradeNames() []string {
	out := make([]string, len(Grades))
	for i, g := range Grades {
		out[i] = string(g)
	}
	return out
}

// GradeValue is the pflag.Value implementation. It is a POINTER to a Grade so an
// unset flag stays distinguishable from a set one — the record format requires
// that distinction (a flag never passed must not appear in the event at all),
// and pflag's Changed alone cannot carry it through a typed accessor.
type GradeValue struct {
	Grade Grade
	set   bool
}

func (v *GradeValue) String() string {
	if v == nil || !v.set {
		return ""
	}
	return string(v.Grade)
}

func (v *GradeValue) Set(s string) error {
	g := Grade(strings.TrimSpace(s))
	for _, known := range Grades {
		if g == known {
			v.Grade, v.set = g, true
			return nil
		}
	}
	// The message names what WOULD have worked. A refusal a seat can act on beats
	// one it will retry, and this is the seat's most common typo class.
	return fmt.Errorf("%q is not a grade — use one of: %s", s, strings.Join(GradeNames(), " | "))
}

// Type is what pflag prints in usage: "--severity grade" reads better than
// "--severity string" and tells the seat there is an enum behind it.
func (v *GradeValue) Type() string { return "grade" }

func (v *GradeValue) Given() bool { return v != nil && v.set }

// GradeUsage is the shared help text, generated from the canonical list.
func GradeUsage(what string) string {
	return what + ": " + strings.Join(GradeNames(), " | ")
}
