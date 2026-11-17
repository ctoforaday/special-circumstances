package flags

import (
	"errors"
	"strings"
)

// OneAncestorRule is the flag's own statement of the rule, for its help line.
const OneAncestorRule = "one id: to fold several defects into one gap, supersede one and close the rest"

// OneAncestorRefusal is the refusal of a second ancestor on --supersedes, in either form: a comma in
// the value, or the flag given twice. A successor supersedes one gap, so a chain is a line and its
// assignee is its origin's; folding several defects into one gap is two acts the record has.
const OneAncestorRefusal = "a successor supersedes exactly one gap. To fold several defects into one gap, supersede one and close the rest — your own with `close`, another lens's by putting it before the bench (`motion docket file`). Nothing was recorded"

// AncestorValue is --supersedes: the one gap a mint replaces. It refuses a comma rather than taking
// the value as an id the shape then refuses, so a seat naming two gaps is told what to do with the
// second; a second occurrence of the flag is refused before the parse (FirstRepeat, Repeat.Said),
// in the same words, rather than kept last.
type AncestorValue struct{ *ShapedValue }

// AncestorID is the --supersedes flag's value: one gap id.
func AncestorID() *AncestorValue { return &AncestorValue{GapID()} }

func (v *AncestorValue) Set(s string) error {
	if strings.Contains(s, ",") {
		return errors.New(OneAncestorRefusal)
	}
	return v.ShapedValue.Set(s)
}
