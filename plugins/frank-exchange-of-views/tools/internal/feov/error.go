// Package feov carries the tool's typed error: ONE type, an enum of codes, minted inline.
//
// Errors are polymorphic, but the polymorphism that matters here is by VALUE, not by Go
// type — a family of bespoke error types would be ceremony for a thing that varies only in
// its code. So there is one Error type and a Code enum. A consumer switches on the code; the
// --json edge reads it via errors.As with no type switch. The rich human sentence the domain
// already writes stays in the message; the code is the machine handle bolted alongside it.
//
// Not everything is minted. Most failures stay plain fmt.Errorf and fall through to the
// default "error" code at the edge. Only the domain faults a seat would branch on — a gap
// that does not exist, a seat acting out of role, a required field left off — earn a code.
// The enum below IS the taxonomy; it grows here and nowhere else.
//
// This is a LEAF package (imports only fmt/errors) so both record (which raises) and the CLI
// edge (which reads) can import it without a cycle.
package feov

import (
	"errors"
	"fmt"
)

// Code is the machine handle on a fault. The zero value is the empty string, which the edge
// treats the same as an unminted error: the generic "error".
type Code string

const (
	// NotFound: an id names nothing on the record — a gap, finding or observation that was
	// never minted (or was guessed rather than looked up).
	NotFound Code = "not_found"
	// MissingField: a required flag or payload field was omitted.
	MissingField Code = "missing_field"
	// RoleViolation: a seat acted outside its role — a seat-id bound to a role other than the
	// command's, or a verb a role does not own.
	RoleViolation Code = "role_violation"
	// Validation: a value was present but ill-formed — an unknown grade, a malformed anchor.
	Validation Code = "validation"
	// Conflict: the act contradicts the record's state — a lineage that does not hold, a
	// closure of something already closed.
	Conflict Code = "conflict"
)

// Error is the tool's coded error. Msg is the human sentence; Code is the machine handle;
// err is an optional wrapped cause so errors.Is/As walk through it.
type Error struct {
	Code Code
	Msg  string
	err  error
}

func (e *Error) Error() string { return e.Msg }

// Unwrap exposes the cause so errors.Is / errors.As see through this error to what it wraps.
func (e *Error) Unwrap() error { return e.err }

// Errorf mints a coded error inline — as cheap as fmt.Errorf, with a code.
func Errorf(code Code, format string, a ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, a...)}
}

// Wrap mints a coded error that carries a cause. The cause stays reachable through Unwrap,
// so a lower-level typed error is not flattened away by the code we add on top.
func Wrap(code Code, cause error, format string, a ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, a...), err: cause}
}

// CodeOf returns the code of the first *Error in err's chain, or "error" if there is none —
// the one lookup the --json edge does, with no switch on concrete types.
func CodeOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return string(e.Code)
	}
	return "error"
}
