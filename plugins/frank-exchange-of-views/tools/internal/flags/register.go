package flags

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// RegisterPayload attaches the prose payload in its inline and file forms.
//
// THE CLAIM THIS COMMENT USED TO MAKE WAS FALSE. It said the flags were "attached through this
// function rather than by hand so no verb can register one form and forget the other" — but
// nothing stopped a verb calling c.Flags().String(Reason, ...) itself, and `spot-check` and
// `outcome` both did, shipping without a file form at all. A convention is not a mechanism.
//
// The mechanism is the Prose type, which owns the pair. This is the thin adapter for the call
// sites that register through seat.Prose and read through seat.Reason.
func RegisterPayload(c *cobra.Command) { new(Prose).Register(c) }

// ReadPayload resolves a command's prose channel to one string.
//
// AN UNREGISTERED READ IS AN ERROR, NOT AN EMPTY STRING. This read both flags with GetString and
// discarded the errors, so a verb that never registered them got "" and no complaint — and the
// write that followed was then refused for a field the seat believed it had supplied. That is the
// same defect the comment on Value() below documents for enum flags, in the function beside it.
func ReadPayload(c *cobra.Command, stdin io.Reader) (string, error) {
	p := ProseOf(c)
	if p == nil {
		return "", fmt.Errorf("%s reads prose but never registered the --%s / --%s channel: "+
			"register it with flags.Prose.Register (seat.Prose) rather than declaring a flag by hand",
			c.CommandPath(), Reason, ReasonFile)
	}
	return p.Read(stdin)
}

// Set writes a flag's value under a payload key, ONLY when it is non-empty.
//
// Setting it unconditionally is a trap worth naming: required-field validation asks
// whether the key is PRESENT, so writing an empty string makes a missing flag look
// supplied and the check passes on nothing. That regression was introduced while
// renaming --gap-id to --id and caught by the bench's own required-fields test.
func Set[P interface{ Set(string, any) P }](p P, key string, c *cobra.Command, flag string) {
	if v := Value(c, flag); v != "" {
		p.Set(key, v)
	}
}

// Value reads a flag AS A STRING WHATEVER ITS TYPE, and that qualifier is the whole point.
//
// `GetString` returns ("", err) for any flag that is not a string flag, and every caller here
// discarded the error — so an enum flag, a grade flag, or anything else backed by a pflag.Value
// read back as UNSET. It cost this the same bug twice in one afternoon: `motion grade file`
// reported "--proposed is required" against a grade the seat had passed, and `bench opinion`
// reported "opinion requires --as (opinions, not dispositions)" against a disposition it had.
// Both times the flag was set, parsed and validated; only the READ was blind.
//
// A flag's own Value.String() is what it parsed, for every flag type there is. This is the one
// place that knows it, so the next flag with a custom type does not rediscover the failure.
func Value(c *cobra.Command, flag string) string {
	if v, err := c.Flags().GetString(flag); err == nil {
		return v
	}
	if f := c.Flags().Lookup(flag); f != nil {
		return f.Value.String()
	}
	return ""
}
