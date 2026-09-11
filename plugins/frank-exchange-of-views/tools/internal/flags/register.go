package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FreeTextAnnotation marks a flag registered through Text: its value is words a seat composes, so
// it passes through the shell as text and the quoting rule applies to it.
const FreeTextAnnotation = "feov_free_text"

// Text registers a FREE-TEXT string flag: the flag, the annotation that says what it is, and the
// quoting rule in the verb's help.
//
// THE RULE IS ATTACHED HERE, NOT TYPED ONTO PAGES. Every free-text flag passes through this
// function (the prose channel included, via Prose.Register), so this is the one place that can say
// "every verb that takes free text shows it" and be right by construction.
// TestEveryFreeTextVerbShowsTheQuotingRule holds it, and refuses a string flag registered the plain
// way unless its name is ClosedForm.
func Text(c *cobra.Command, name, desc string) { TextVar(c, new(string), name, desc) }

// TextVar is Text bound to a variable.
func TextVar(c *cobra.Command, p *string, name, desc string) {
	c.Flags().StringVar(p, name, "", desc)
	_ = c.Flags().SetAnnotation(name, FreeTextAnnotation, []string{"true"})
	if !strings.Contains(c.Long, ProseFooter) {
		c.Long = strings.TrimRight(c.Long, "\n") + "\n\n" + ProseFooter
	}
}

// IsFreeText reports whether a flag was registered through Text.
func IsFreeText(f *pflag.Flag) bool { return len(f.Annotations[FreeTextAnnotation]) > 0 }

// RegisterPayload attaches the prose payload channel.
//
// A verb that declared c.Flags().String(Reason, ...) by hand used to ship without the channel's
// other halves — `spot-check` and `outcome` both did. The mechanism is the Prose type, which owns
// the flag, its wording and the quoting rule in the verb's help. This is the thin adapter for the
// call sites that register through seat.Prose and read through seat.Reason.
func RegisterPayload(c *cobra.Command) { new(Prose).Register(c) }

// ReadPayload resolves a command's prose channel to one string.
//
// AN UNREGISTERED READ IS AN ERROR, NOT AN EMPTY STRING. This read the flag with GetString and
// discarded the error, so a verb that never registered it got "" and no complaint — and the write
// that followed was then refused for a field the seat believed it had supplied. That is the same
// defect the comment on Value() below documents for enum flags, in the function beside it.
func ReadPayload(c *cobra.Command) (string, error) {
	p := ProseOf(c)
	if p == nil {
		return "", fmt.Errorf("%s reads prose but never registered the --%s channel: "+
			"register it with flags.Prose.Register (seat.Prose) rather than declaring a flag by hand",
			c.CommandPath(), Reason)
	}
	return p.Read(), nil
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
