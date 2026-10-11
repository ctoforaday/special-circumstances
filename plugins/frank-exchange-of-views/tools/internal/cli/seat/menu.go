package seat

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// A REFUSAL SAYS WHAT WENT WRONG FIRST, THEN TEACHES THE SURFACE. Never a bare list, never help
// with no diagnosis in front of it.
//
// # Why the refusal is the teaching channel at all
//
// MEASURED across nine probe sittings: seats do not learn this tool from `--help`. Every one of
// them read it once or twice in twenty to forty tool calls. They learn it from REFUSALS — a seat
// that guesses `blue avenue` is told the verb does not exist, and that is the moment it
// finds out what it has.
//
// So the refusal was the primary channel and the LOSSY one. `blue --help` gives every verb with
// what it does and its flags; the refusal gave fourteen bare names. A seat learns WHAT EXISTS and
// never WHAT IT IS FOR, which is precisely the measured behaviour: seats use the verbs they can
// name, in the ways they already assumed.
//
// # Why the diagnosis comes first
//
// The first draft printed cobra's help and then the error, because that is the order cobra
// produces them in. That is backwards for a reader: a page of help arrives with no statement of
// what went wrong, and the seat has to infer its own mistake from a list of things it did not do.
// The whole message is built here instead, diagnosis first, so a seat never has to guess what it
// got wrong before it can use what follows.
//
// # And the mechanism was the cause of the lossy half
//
// The two texts differed because the refusal NEVER TOUCHED COBRA'S HELP. `SilenceUsage: true`
// suppresses the real usage on error — correct for a verb's own validation failure, which should
// not dump a page — and the role command then hand-built a `names` slice beside it. Two renderings
// of one fact, and the hand-rolled one had no way to carry a description because nothing generated
// it from the tree. This embeds cobra's own, which carries the verb descriptions AND the
// enumerated-value menus, from the one source that cannot drift.
//
// This is the same defect for the third time, in each of the three places a seat learns:
//
//	--view names with no descriptions   (the views table's `desc` field was read NOWHERE)
//	enum values with no meanings        (closed | defect_accepted | … under one shared sentence)
//	verb names with no meanings         (here — in the channel that turns out to be primary)

// HelpText renders a command's full help to a string, so it can be embedded in a message rather
// than printed around one.
func HelpText(c *cobra.Command) string {
	var b bytes.Buffer
	out, errOut := c.OutOrStdout(), c.ErrOrStderr()
	c.SetOut(&b)
	c.SetErr(&b)
	_ = c.Help()
	c.SetOut(out)
	c.SetErr(errOut)
	return strings.TrimRight(b.String(), "\n")
}

// RefuseAndTeach is the one shape: what went wrong, then what exists.
//
// `wrong` must be a complete diagnosis on its own — a seat that read only the first line should
// already know what it did, and it refers to no text after it: the page follows it in the human
// form and is named, not printed, in the JSON one. The help is what the seat needs NEXT, not what
// tells it there was a problem. `c` is the command whose page answers the refusal, which is the
// failing leaf for a flag and the group or the root for a command that was not found under it.
//
// The code is the refusal's KIND, stated where the refusal is raised: a consumer branches on it,
// and nothing downstream supplies one for a refusal that came without.
func RefuseAndTeach(c *cobra.Command, code feov.Code, wrong string) error {
	return TeachCause(c, wrong, feov.Errorf(code, "%s", wrong))
}

// TeachCause is RefuseAndTeach over an error that already exists — cobra's own, coded by its
// producer — which stays reachable beneath the refusal, so its code is the refusal's.
func TeachCause(c *cobra.Command, wrong string, cause error) error {
	lists := "what it takes, flag by flag"
	switch {
	case c == c.Root():
		lists = "every command on this surface"
	case c.HasAvailableSubCommands():
		lists = "what it holds, each with what it is for"
	}
	return &Refusal{Diagnosis: wrong, Help: HelpText(c), Path: c.CommandPath(), Lists: lists, Cause: cause}
}
