package seat

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	return &Refusal{Diagnosis: wrong, Help: HelpText(c), Path: c.CommandPath(), Lists: listsOf(c), Cause: cause}
}

// listsOf says what a command's page holds, for the pointer that names the page.
func listsOf(c *cobra.Command) string {
	switch {
	case c == c.Root():
		return "every command on this surface"
	case c.HasAvailableSubCommands():
		return "what it holds, each with what it is for"
	}
	return "what it takes, flag by flag"
}

// A FLAG REFUSAL IS A FEW LINES, NOT THE PAGE. A required flag left out and a flag the verb does
// not have are each answered by a handful of facts the command tree already holds — which flags
// are owed and what each is for; which verbs do take the flag that was typed, and what this verb
// takes instead — so the refusal states those and names the page. The whole page under either
// was 3–5 KB a call, and a seat that batches four calls on one guess reads it four times.
//
// Both forms carry the same text: there is no page for the JSON envelope to stand a pointer in for.

// RefuseMissingFlags is the refusal for required flags left out: the flags by name, each one's own
// usage line — what it is FOR, which the bare name does not say — and the page that lists the rest.
//
// THE MISSING SET IS READ FROM FLAG STATE, the way cobra's own check reads it — a flag carrying
// the required annotation that the call did not set — and never recovered from cobra's message.
// It returns nil when no required flag is missing, so a caller cannot render an empty list.
func RefuseMissingFlags(c *cobra.Command, cause error) error {
	owed := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
	var names []string
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if req := f.Annotations[cobra.BashCompOneRequiredFlag]; len(req) > 0 && req[0] == "true" && !f.Changed {
			owed.AddFlag(f)
			names = append(names, "--"+f.Name)
		}
	})
	if len(names) == 0 {
		return nil
	}
	return &Refusal{
		Diagnosis: "requires " + strings.Join(names, ", ") + "\n\n" + strings.TrimRight(owed.FlagUsagesWrapped(0), "\n"),
		Path:      c.CommandPath(), Lists: listsOf(c), Cause: cause,
	}
}

// RefuseUnknownFlag is the refusal for a flag the verb does not have. `said` is the parser's own
// sentence for it, and `name` the flag word as typed, without its dashes.
//
// IT ANSWERS THE MISTAKE SEATS ACTUALLY MAKE, which is carrying a flag across from a sibling verb:
// `--key` typed on a verb with no handle because three verbs beside it take one. So it says which
// verbs on THIS surface take the flag — read from the tree the refusal is raised in, which is the
// asking seat's own — and what this verb takes. It names; it does not accept: the call is refused
// and nothing ran, and no second spelling of any flag exists for it to have matched.
//
// `answered` says the diagnosis already names the flag the seat meant (the record's word for a
// field, translated to the flag that sets it), so the holders of the word as typed are left out.
func RefuseUnknownFlag(c *cobra.Command, name, said string, answered bool, cause error) error {
	var b strings.Builder
	b.WriteString(said)
	if !answered {
		if takers := TakersOf(c.Root(), name, c); len(takers) > 0 {
			verb := "take"
			if len(takers) == 1 {
				verb = "takes"
			}
			fmt.Fprintf(&b, " — not one of this verb's flags. On this surface %s %s it.", andList(quoted(takers)), verb)
		} else {
			b.WriteString(" — not one of this verb's flags, and no verb on this surface takes it.")
		}
	}
	// WHAT THIS VERB TAKES: every flag it parsed against, less the ones every command shares (the
	// root's) and its `--help`. A view under a group takes the group's flags, so they are listed.
	var own []string
	global := c.Root().PersistentFlags()
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" && !f.Hidden && global.Lookup(f.Name) == nil {
			own = append(own, "--"+f.Name)
		}
	})
	if len(own) == 0 {
		b.WriteString("\nThis verb takes no flags of its own.")
	} else {
		b.WriteString("\nThis verb's flags: " + strings.Join(own, ", ") + ".")
	}
	return &Refusal{Diagnosis: b.String(), Path: c.CommandPath(), Lists: listsOf(c), Cause: cause}
}

// declarers is the one walk that answers "which verbs take this flag": the runnable commands
// under root that DECLARE it — among their own flags, or as a group for everything beneath it,
// where the group is named once and stands for its members.
//
// IT READS THE DECLARED SETS AND MERGES NOTHING. cobra's LocalFlags and InheritedFlags fold a
// command's inherited flags into its own set as a side effect, and this runs while the tree is
// being built: a walk that changed what every later reader of a command's flags sees would be
// felt far from here.
func declarers(root *cobra.Command, flag string) map[*cobra.Command]*pflag.Flag {
	out := map[*cobra.Command]*pflag.Flag{}
	var walk func(cur *cobra.Command, inherited bool)
	walk = func(cur *cobra.Command, inherited bool) {
		for _, sub := range cur.Commands() {
			persistent := sub.PersistentFlags().Lookup(flag)
			f := persistent
			if f == nil && !inherited {
				f = sub.Flags().Lookup(flag)
			}
			if f != nil && !inherited && sub.Runnable() && sub.IsAvailableCommand() {
				out[sub] = f
			}
			walk(sub, inherited || persistent != nil)
		}
	}
	walk(root, root.PersistentFlags().Lookup(flag) != nil)
	return out
}

// TakersOf lists those commands by path, sorted, less `except` — the command asking.
func TakersOf(root *cobra.Command, flag string, except *cobra.Command) []string {
	var out []string
	for c := range declarers(root, flag) {
		if c != except {
			out = append(out, PathOf(c))
		}
	}
	sort.Strings(out)
	return out
}

// NameTakers closes a flag's usage, on every verb that takes it, with which verbs those are.
//
// A FLAG DESCRIBED WHERE IT EXISTS AND BOUNDED NOWHERE READS AS UNIVERSAL. `--key` says "your own
// stable handle … a repeat returns the first result", on three of blue's four report-writing
// verbs; a seat numbering its acts through it carried it to the fourth, whose page is silent
// about a handle. The bound is said where the generalisation forms — in the flag's own line.
//
// THE LIST IS THE TREE'S: the verbs under this root that declare the flag, so each seat reads its
// own surface's and a verb that gains or loses the flag changes every page's sentence with it.
// `noun` is what the flag carries, as the sentence names it ("a handle").
func NameTakers(root *cobra.Command, flag, noun string) {
	takers := TakersOf(root, flag, nil)
	if len(takers) == 0 {
		return
	}
	verb := "take"
	if len(takers) == 1 {
		verb = "takes"
	}
	// BARE NAMES: pflag reads the first backquoted word of a usage as the name of the flag's value.
	bound := fmt.Sprintf(". Only %s %s %s: no other verb on this surface has this flag", andList(takers), verb, noun)
	for _, f := range declarers(root, flag) {
		f.Usage += bound
	}
}

func quoted(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = "`" + n + "`"
	}
	return out
}

// andList joins names as prose: "a", "a and b", "a, b and c".
func andList(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}
