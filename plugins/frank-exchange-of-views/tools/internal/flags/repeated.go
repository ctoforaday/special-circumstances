package flags

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Repeat is one flag a call stated twice, with both values as typed.
type Repeat struct {
	Flag        *pflag.Flag
	First, Then string
}

// Said is the refusal's text. It names both values because either could be the one the seat meant,
// and it keeps neither. THE VALUES ARE ON THEIR OWN LINES: the refusal log keeps a diagnosis's first
// line, and names a flag a seat passed without the words it passed.
func (r *Repeat) Said() string {
	do := "pass it once, with the value you mean"
	if _, list := r.Flag.Value.(*CSV); list {
		do = "pass one comma-separated list"
	}
	return fmt.Sprintf("--%s was given twice and takes one value, so neither was kept and nothing was recorded: %s. It was given:\n  %q\n  %q",
		r.Flag.Name, do, r.First, r.Then)
}

// FirstRepeat is the first flag args state twice, or nil. args are what cobra's Find leaves for c.
//
// pflag KEEPS THE LAST of a repeated flag and says nothing, so the value before it is dropped:
// measured on universe-m37, a lens passed `mint --problem` twice and the record holds the second. A
// flag's Value cannot refuse for itself without every type carrying the same check, and wrapping
// each Value hides its type from the readers that assert it, and from pflag's own help rendering.
//
// SO THE SAME PARSER WALKS args FIRST, with a callback that sets nothing: ParseAll hands it each
// flag it reads, which makes `--x=1`, `--x 1`, a boolean with no value, a shorthand and the `--` that
// ends the flags the parser's own reading, with no second grammar here. Unknown flags are let past
// because the parse that follows refuses them in their own terms, and any other parse error is
// that parse's too.
//
// A FLAG THAT TAKES A SECOND VALUE SAYS SO BY ITS TYPE: pflag's list types implement SliceValue and
// append on every occurrence. CSV does not: it holds one comma-separated list.
func FirstRepeat(c *cobra.Command, args []string) *Repeat {
	// cobra adds --help during its own parse; a flag the scan does not hold stops it at that word.
	c.InitDefaultHelpFlag()
	scan := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
	scan.SetOutput(io.Discard)
	scan.Usage = func() {}
	scan.ParseErrorsAllowlist.UnknownFlags = true
	scan.AddFlagSet(c.LocalFlags())
	scan.AddFlagSet(c.InheritedFlags())
	var found *Repeat
	stop := errors.New("repeated")
	seen := map[string]string{}
	_ = scan.ParseAll(args, func(f *pflag.Flag, value string) error {
		if _, appends := f.Value.(pflag.SliceValue); appends {
			return nil
		}
		first, again := seen[f.Name]
		if !again {
			seen[f.Name] = value
			return nil
		}
		found = &Repeat{Flag: f, First: first, Then: value}
		return stop
	})
	return found
}
