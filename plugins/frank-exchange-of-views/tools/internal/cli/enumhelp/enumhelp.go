// Package enumhelp puts an enum's per-value MEANINGS into `--help`, which is the one thing no
// flag package does.
//
// # Why a template and not a library
//
// enumflag (github.com/thediveo/enumflag/v2) is the parser: type-safe values, case handling,
// slice-valued flags, and shell completion, in one implementation instead of a hand-rolled
// pflag.Value per set. Its `Help[E]` type looks like the answer and is not — it feeds COMPLETION
// only, and never reaches `--help`.
//
// Cobra will not close the gap either, and the reason is structural rather than an omission: its
// help is rich about COMMANDS and FLAGS, and an enum's values are neither. They are a value-space
// inside one flag, which cobra has no model for. So every vocabulary this system expresses as flag
// values had nowhere to put its semantics, and ours went into source comments — written for a Go
// reader and delivered to no seat.
//
// A `describe` command was the other candidate and was rejected: it splits the answer from the
// question and costs a round trip, and a seat that must run a second command to understand the
// first will guess instead.
//
// # What it does
//
// A verb registers its enum-valued flags here. The usage template then renders an "ENUMERATED
// VALUES" section under the flags, one line per value with what that value is FOR. The flag's own
// usage line stays short.
//
// The section is also where facts that cannot live on a command go — which value demands which
// other flag, most usefully. `repaired_with_regression` REQUIRES `--successor` and nothing in
// cobra's model can say so at the flag level.
package enumhelp

import (
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/thediveo/enumflag/v2"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

type entry struct {
	flag   string
	values []record.EnumValue
}

// Flag registers an enum-valued flag whose value the caller reads back through seat.Str rather
// than through a bound variable, which is how every verb in this tree already reads its flags.
//
// The target is allocated here and deliberately dropped: seat.Str falls through to the flag's own
// Value.String() for any non-string flag type, so the read path is unchanged and no verb has to
// grow a variable to hold something it immediately puts into a payload.
func Flag(c *cobra.Command, name string, e record.EnumField, usage string) {
	// THE REFUSAL CARRIES THE FIELD'S `Why`, NOT THE FLAG'S USAGE. Why says what the mistype
	// WOULD HAVE DONE — "a PASS is checked against the open board by exact match, so any other
	// spelling skips the check entirely and records an unadjudicated pass" — and that sentence is
	// the entire reason the set is closed. Substituting the usage line ("the seat's terminal
	// act") loses it, which the error-catalogue golden caught the moment it was tried.
	v := enumflag.NewWithoutDefault(new(string), typeName(name, e.Values), record.Identifiers(e.Values), enumflag.EnumCaseSensitive)
	//
	// NO SHELL COMPLETION IS REGISTERED. No surface mounts cobra's completion commands, so nothing
	// could ask for it, and cobra keeps every registered completion in a process-global map keyed
	// by *pflag.Flag — each tree built would leave its enum flags pinned there for good.
	c.Flags().Var(&speaking{Value: v, values: e.Values, why: e.Why}, name, usage)
}

// speaking wraps enumflag's Value so the PARSING is the library's and the REFUSAL is ours.
//
// enumflag rejects with `must be 'FAIL', 'PASS'` — correct, and it teaches nothing. This project's
// own sentence names the near-miss ("it differs from \"PASS\" only in case"), says what the
// mistype would have done, and prints the menu. That is not decoration: the refusal is where a
// seat meets the design, and a probe measured a seat learning its whole surface from what the
// tool printed back at it after a failure.
//
// Wrapping rather than replacing keeps enumflag's parsing and slice handling, so this
// is a message layer and never a second definition of what is legal.
type speaking struct {
	pflag.Value
	values []record.EnumValue
	why    string
}

func (s *speaking) Set(v string) error {
	if err := s.Value.Set(v); err != nil {
		return record.Refuse(v, s.values, s.why)
	}
	return nil
}

// typeName is what cobra shows as the flag's argument placeholder. `--as as` says nothing; the
// point is to name the VALUE SPACE, and the enumerated-values section below carries the rest.
func typeName(flag string, values []record.EnumValue) string {
	if len(values) <= 3 {
		return strings.Join(record.Names(values), "|")
	}
	return flag + "-value"
}

// Section renders the enumerated-values block for a command, or "" if it declares none.
func Section(c *cobra.Command) string {
	entries := declared(c)
	if len(entries) == 0 {
		return ""
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].flag < entries[j].flag })

	var b strings.Builder
	b.WriteString("Enumerated values:\n")
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  --" + e.flag + "\n")
		b.WriteString(record.Menu(e.values))
	}
	return strings.TrimRight(b.String(), "\n")
}

// usageTemplate is cobra's default with one section added.
//
// IT IS COBRA'S CURRENT DEFAULT, NOT THE ONE IT FORKED FROM. The copy here was taken before cobra
// grew command groups, so it rendered every child under a single "Available Commands:" heading and
// dropped .Groups on the floor — silently, because a template that ignores a field looks exactly
// like a tree that has no groups. Assigning a GroupID had no visible effect at all until this
// branch came back, which is the shape of every fork of someone else's default.
//
// Kept as a whole template rather than assembled, because cobra's is a single string and a
// partial override would silently drop whatever it does not reproduce — the failure mode being
// that `--help` quietly stops printing the global flags and nobody notices for a year.
const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if enumSection .}}

{{enumSection .}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var installOnce sync.Once

// Install puts the template on the root. Cobra resolves a command's template by walking to its
// parent when unset, so one call covers the tree.
func Install(root *cobra.Command) {
	installOnce.Do(func() {
		cobra.AddTemplateFunc("enumSection", Section)
	})
	root.SetUsageTemplate(usageTemplate)
}

// declared reads the enum flags off the command's own flag set. The values live on the flag
// itself, so nothing outside the tree holds a command: a side table keyed by *cobra.Command
// pinned every tree a test binary built, and internal/cli's suite peaked at 5 GB resident.
func declared(c *cobra.Command) []entry {
	var out []entry
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if s, ok := f.Value.(*speaking); ok {
			out = append(out, entry{flag: f.Name, values: s.values})
		}
	})
	return out
}

// Registered reports the flags documented for a command path — for the gate that asserts every
// set-shaped flag in the tree has its values described somewhere a seat reads.
func Registered(c *cobra.Command) map[string][]record.EnumValue {
	out := map[string][]record.EnumValue{}
	for _, e := range declared(c) {
		out[e.flag] = e.values
	}
	return out
}
