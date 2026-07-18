package record

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The CLI layer, on cobra.
//
// WHY COBRA, having previously argued against it: the objection was contingent,
// not principled. While the mjs oracle was live, the differential gate compared
// CLI text byte-for-byte, so any change to the help renderer or error format
// would have diverged — and the oracle could not be edited to match, because an
// oracle you edit to pass your test is not an oracle. The oracle is retired, so
// that constraint is gone and what remains is a real defect cobra fixes:
//
//	UNKNOWN FLAGS WERE SILENTLY IGNORED. `--anchor-set` instead of `--anchor-seat`
//	produced a closure with no anchor and no complaint — a record-integrity bug in
//	the tool whose entire job is record integrity. pflag rejects unknown flags by
//	construction, and a bare `--severity` now fails asking for its argument instead
//	of becoming boolean true and failing the grade enum three layers later.
//
// The published CONTRACT is unchanged and is held by the help-contract goldens,
// which were recorded before this migration precisely so the rendering could
// change while the contract could not.

// Verb is one entry in a seat's contract. The verb SET is the role boundary: a
// missing verb is a missing capability, enforced by topology rather than by
// asking a seat to behave.
type Verb struct {
	Name string
	Help string
	// Flags this verb accepts, as payload-key:flag specs (see Args.Copy) or bare
	// flag names. Declaring them here is what makes unknown-flag rejection
	// possible, and it keeps the flag set and the payload shape in one place
	// rather than drifting apart.
	Flags []string
	Fn    func(runDir, seatID string, a *Args) (string, error)
}

// Args reads the parsed flags. It keeps the absent/present distinction the
// record format depends on: a flag the seat never passed must not appear in the
// event at all, and pflag's Changed is exactly that question.
type Args struct{ fs *pflag.FlagSet }

func (a *Args) Has(flag string) bool {
	f := a.fs.Lookup(flag)
	return f != nil && f.Changed
}

func (a *Args) Str(flag string) string {
	v, err := a.fs.GetString(flag)
	if err != nil {
		return ""
	}
	return v
}

// Val is what the oracle placed in a payload literal. Every flag is a string
// now: the old bare-flag-means-true path existed only because a hand-rolled
// parser could not tell a missing argument from an intentional boolean, and it
// only ever produced values that failed validation.
func (a *Args) Val(flag string) any { return a.Str(flag) }

// SetInto copies a flag into the payload only when present — the `undefined`
// drop rule that keeps absent flags out of the event entirely.
func (a *Args) SetInto(p *Payload, key, flag string) *Payload {
	if a.Has(flag) {
		p.Set(key, a.Val(flag))
	}
	return p
}

// Copy is the declarative form: a verb listing the payload it builds, in order.
// A spec is "name" when the payload key and flag agree, or "payload_key:flag"
// when they differ (the CLI spells flags in kebab-case, the record in snake).
//
//	a.Copy(p, "gap_id:id", "severity", "complexity_cost:cx", "basis")
func (a *Args) Copy(p *Payload, specs ...string) *Payload {
	for _, spec := range specs {
		key, flag, found := strings.Cut(spec, ":")
		if !found {
			flag = key
		}
		a.SetInto(p, key, flag)
	}
	return p
}

// List mirrors the oracle's list(): comma-separated, trimmed, empties removed,
// ALWAYS an array (the record carries [] rather than dropping the key).
func List(v string) []string {
	out := []string{}
	for _, s := range strings.Split(v, ",") {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// PayloadText resolves --file (read whole) or --text, else "".
func PayloadText(a *Args) (string, error) {
	if a.Has("file") {
		b, err := os.ReadFile(a.Str("file"))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if a.Has("text") {
		return a.Str("text"), nil
	}
	return "", nil
}

// flagOf reduces a Copy spec to the flag it names.
func flagOf(spec string) string {
	if _, flag, found := strings.Cut(spec, ":"); found {
		return flag
	}
	return spec
}

// RoleCommand builds a role's subcommand tree: its verbs, plus the render verb
// every seat may invoke, plus the --run/--seat-id flags they all require.
func RoleCommand(role, short string, verbs []Verb) *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   role,
		Short: short,
		Long:  short + "\n" + frictionFooter,
		// ArbitraryArgs so an unrecognised verb reaches RunE instead of cobra's
		// generic `unknown command "mint" for "feov-record lens"`. The difference
		// is the whole point of the role boundary: the seat must be told what it
		// CAN do, not merely that it did something wrong. The help-contract golden
		// caught this regression during the cobra migration, which is exactly the
		// job it was recorded for.
		Args: cobra.ArbitraryArgs,
		// Flag parsing off at the ROLE level only (subcommand resolution happens
		// first, so each verb still parses and validates its own flags strictly).
		// Without this, `lens mint --run x` reports `unknown flag: --run` — the
		// role command trying to parse flags it does not own — which buries the
		// real answer: the lens has no mint verb at all.
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			for _, a := range args {
				if a == "--help" || a == "-h" || a == "help" {
					return c.Help()
				}
			}
			if len(args) > 0 {
				return fmt.Errorf("verb %q is outside this seat's role (available: %s, render)", args[0], verbNames(verbs))
			}
			// A role invoked with no verb is a usage error, not a no-op: silently
			// succeeding would let a mis-scripted seat believe it recorded something.
			return fmt.Errorf("%s: a verb is required (%s, render)", role, verbNames(verbs))
		},
		SilenceUsage: true,
	}

	for _, v := range verbs {
		roleCmd.AddCommand(verbCommand(role, v))
	}
	roleCmd.AddCommand(renderCommand(role))
	return roleCmd
}

func verbCommand(role string, v Verb) *cobra.Command {
	c := &cobra.Command{
		Use:          v.Name,
		Short:        v.Help,
		Long:         v.Help + "\n" + frictionFooter,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
		Args:         cobra.NoArgs,
	}
	c.Flags().String("run", "", flagDoc(v.Name, "run"))
	c.Flags().String("seat-id", "", flagDoc(v.Name, "seat-id"))
	for _, spec := range v.Flags {
		f := flagOf(spec)
		// An undocumented flag makes the help stop being a contract, so a missing
		// gloss shows up in the output rather than as a silent blank column.
		doc := flagDoc(v.Name, f)
		if doc == "" {
			doc = "(undocumented — add it to flagDocs)"
		}
		c.Flags().String(f, "", doc)
	}
	// --file and --text are the prose payload channel; every verb that reads
	// PayloadText accepts them, and declaring them centrally keeps a verb from
	// accidentally omitting one.
	if c.Flags().Lookup("file") == nil {
		c.Flags().String("file", "", flagDoc(v.Name, "file"))
	}
	if c.Flags().Lookup("text") == nil {
		c.Flags().String("text", "", flagDoc(v.Name, "text"))
	}

	c.RunE = func(cmd *cobra.Command, _ []string) error {
		a := &Args{fs: cmd.Flags()}
		runDir := a.Str("run")
		if runDir == "" {
			return fmt.Errorf("%s: --run <runDir> is required", role)
		}
		seatID := a.Str("seat-id")
		if seatID == "" {
			return fmt.Errorf("%s: --seat-id is required (the engine assigns it; it is in your prompt)", role)
		}
		if err := checkSeatRole(role, seatID); err != nil {
			return err
		}
		result, err := v.Fn(runDir, seatID, a)
		if err != nil {
			// Carry the ROLE in the prefix. A seat reading "record: close requires
			// --id" learns less than one reading "merge: record: close requires
			// --id", because the role names which contract it is being held to —
			// and a lens seeing "merge" in its own error has learned something
			// more useful still.
			return fmt.Errorf("%s: %w", role, err)
		}
		// Render-on-mutation: projections stay current after every write.
		if v.Name != "register" {
			if _, rerr := Render(runDir, ""); rerr != nil {
				return rerr
			}
		}
		if result != "" {
			fmt.Println(result)
		}
		return nil
	}
	return c
}

func renderCommand(role string) *cobra.Command {
	c := &cobra.Command{
		Use:          "render",
		Short:        "read-only projection refresh (any seat may invoke)",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
	}
	c.Flags().String("run", "", "the run directory")
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		a := &Args{fs: cmd.Flags()}
		runDir := a.Str("run")
		if runDir == "" {
			return fmt.Errorf("--run <runDir> is required")
		}
		r, err := Render(runDir, "")
		if err != nil {
			return err
		}
		extra := ""
		if r.Anomalies > 0 {
			extra = fmt.Sprintf(", %d anomalies", r.Anomalies)
		}
		fmt.Printf("feov-record %s: rendered to %s (%d open, %d closed%s)\n", role, r.Out, r.Open, r.Closed, extra)
		return nil
	}
	return c
}

func verbNames(verbs []Verb) string {
	names := make([]string, len(verbs))
	for i, v := range verbs {
		names[i] = v.Name
	}
	return strings.Join(names, ", ")
}
