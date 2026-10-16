package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// EVERY --id SAYS WHICH KIND OF ID IT WANTS (gblock: "fix --id or remove it where it can't be
// supported", and "some IDs on one flag and others on another make no sense").
//
// `--id` carries a gap, a motion, an inquiry and a proof digest depending on the verb. That is fine —
// internal/flags/idnamespace_test.go proves the shapes decide a value unambiguously — but only if each
// verb SAYS which kind it takes, because the payoff of one flag is the refusal: "M1 is a motion id and
// this verb takes a gap id" sends a seat to the right verb, where "invalid id" sends it to re-check
// its typing.
//
// A shaped value carries that: pflag prints its kind as the flag's type, and the refusal is built from
// the same word. Registered as a plain string, `--id` prints `string` and refuses nothing — five of
// the nine registrations were shaped and four were not, which is the uniformity this gate holds.
func TestEveryIDFlagDeclaresWhichKindItTakes(t *testing.T) {
	var plain []string
	for _, r := range AllRoots() {
		walkIDFlags(r, "", &plain)
	}
	sort.Strings(plain)
	if len(plain) > 0 {
		t.Errorf("%d --id flag(s) are registered as a plain string, so the help prints `string` and a "+
			"wrong-kind value is refused by a lookup rather than by its shape. Register through a shaped "+
			"value (flags.GapID(), flags.MotionID(), flags.InquiryID(), flags.SHA(), …) so the page names "+
			"the kind and the refusal can say what it got:\n  %s", len(plain), strings.Join(plain, "\n  "))
	}
}

// walkIDFlags collects every command path whose --id is not a shaped value.
//
// IT READS THE BUILT TREE, not the source. A grep for `flags.ID` finds registrations and misses which
// of them went through Var with a shape — the distinction this gate is about — and would also miss one
// added through a helper.
func walkIDFlags(c *cobra.Command, path string, out *[]string) {
	here := strings.TrimSpace(path + " " + c.Name())
	check := func(f *pflag.Flag) {
		if f == nil || f.Name != flags.ID {
			return
		}
		// A shaped value names its kind as the flag's type; a plain string flag's type is "string".
		if f.Value.Type() == "string" {
			*out = append(*out, here+" --id (prints type `string`)")
		}
	}
	check(c.Flags().Lookup(flags.ID))
	// Persistent flags are registered once on a parent and inherited, so a plain one there is a plain
	// one on every child — the largest version of this defect rather than an exception to it.
	if c.PersistentFlags() != nil {
		if f := c.PersistentFlags().Lookup(flags.ID); f != nil && f.Value.Type() == "string" {
			*out = append(*out, here+" --id (persistent, prints type `string`)")
		}
	}
	for _, sub := range c.Commands() {
		walkIDFlags(sub, here, out)
	}
}
