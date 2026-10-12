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

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Grade is a point on the shared grading scale, as the word a seat types. The scale — its words
// and their order — is the record's Grade enum (recordpb.GradeWords); this package holds no list
// of them, so the flag, its help and its refusal cannot name a grade the write path does not.
type Grade string

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
	g := strings.TrimSpace(s)
	if _, ok := recordpb.BySpelling(recordpb.Grade(0).Descriptor(), g); ok {
		v.Grade, v.set = Grade(g), true
		return nil
	}
	// The message names what WOULD have worked. A refusal a seat can act on beats
	// one it will retry, and this is the seat's most common typo class.
	return fmt.Errorf("%q is not a grade — use one of: %s", s, strings.Join(recordpb.GradeWords(), " | "))
}

// Type is what pflag prints in usage: "--severity grade" reads better than
// "--severity string" and tells the seat there is an enum behind it.
func (v *GradeValue) Type() string { return "grade" }

func (v *GradeValue) Given() bool { return v != nil && v.set }

// GradeUsage is the shared help text: what the flag grades, then the scale in the record's order.
func GradeUsage(what string) string {
	return what + ": " + strings.Join(recordpb.GradeWords(), " | ")
}
