package record

import "errors"

// ToolLogged marks a refusal whose log entry the tool has ALREADY written — an estoppel, a proof
// that would not run, a source that could not be fetched. The refusal logger records every other
// refusal a seat receives; these carry a more specific entry of their own, and a second, generic
// one beside it would count the same event twice.
//
// A TYPE, NOT A PHRASE. The logger could have recognised these by their wording; a reworded
// message would then have been logged twice with nothing failing.
func ToolLogged(err error) error {
	if err == nil {
		return nil
	}
	return toolLogged{err}
}

// IsToolLogged reports whether err, or anything it wraps, was marked by ToolLogged.
func IsToolLogged(err error) bool {
	var t toolLogged
	return errors.As(err, &t)
}

type toolLogged struct{ error }

func (e toolLogged) Unwrap() error { return e.error }
