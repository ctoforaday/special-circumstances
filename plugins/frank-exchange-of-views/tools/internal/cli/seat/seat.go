// Package seat carries what every verb of every role shares: the seat context,
// the preconditions, and the after-render.
//
// It is deliberately small. The first cut of this CLI grew a Verb struct, a
// []Verb table per role, a RoleCommand factory and an Args shim over pflag —
// a hand-rolled framework, built because cobra was adopted late for flag parsing
// rather than designed in. Cobra already owns every one of those jobs:
//
//	persistent flags     --run and --seat-id declared ONCE on the root and
//	                     inherited, rather than re-declared by sixteen verbs
//	PreRunE              preconditions where cobra runs preconditions
//	PostRunE             render-on-mutation as an after-hook, not an
//	                     if-statement a wrapper had to remember
//	flag usage strings   documentation at the declaration site, which is what
//	                     made the flagDocs map unnecessary
//
// What is left here is shared BEHAVIOUR, not structure: no command is defined in
// this package, and no role's contract is stated here. Each verb builds itself,
// in its own file, next to its own flags.
package seat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// FrictionFooter closes the loop the help opens. A seat that needs something the
// contract does not offer must not improvise around it — the gap in the tooling
// is itself a finding, and friction is the channel that carries it to the human
// who can retool the seat.
const FrictionFooter = `
If you need a verb or a flag that is not listed here, it does not exist for you:
do not improvise around it, and do not hand-write the artifact. Record what you
needed and what you would have done with the 'friction' verb — a missing
capability is a finding about the tooling, and that channel is how it gets fixed.`

// Context is what every verb needs and no verb should re-derive.
type Context struct {
	RunDir string
	SeatID string
	Role   string
}

// Of reads the seat context from the inherited persistent flags, inferring the run
// directory when the flag is absent.
func Of(cmd *cobra.Command, role string) Context {
	runDir, _ := cmd.Flags().GetString(flags.Run)
	if runDir == "" {
		runDir = InferRunDir("")
	}
	seatID, _ := cmd.Flags().GetString(flags.SeatID)
	return Context{RunDir: runDir, SeatID: seatID, Role: role}
}

// InferRunDir answers "which run am I in?" from the live-run marker instead of
// requiring every call to say so.
//
// The first live run measured 55 tool-call errors in 534 executions, and TEN of them
// were this one flag: a seat copies the engine's `register --run <dir> --seat-id <id>`
// line, then improvises later verbs and drops the flags. Shell state does not persist
// between tool calls, so the seat cannot export it once; there is no per-agent
// environment variable to carry it; and the engine is a sandboxed script that cannot
// set one. But the answer is already on disk — setup writes .claude/run-live.json with
// the runDir, and the hook guards already read it.
//
// An explicit --run always wins: inference is a fallback for the seat that forgot, not
// a new source of truth. The marker's runDir is project-relative, so it resolves
// against the directory holding .claude/, and an inferred directory that does not
// exist is discarded rather than passed on — a wrong answer here would attach a seat's
// events to the wrong run, which is worse than the error it replaces.
func InferRunDir(start string) string {
	dir := start
	if dir == "" {
		if p := os.Getenv("CLAUDE_PROJECT_DIR"); p != "" {
			dir = p
		} else if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	for i := 0; dir != "" && i < 12; i++ {
		marker := filepath.Join(dir, ".claude", "run-live.json")
		if b, err := os.ReadFile(marker); err == nil {
			var m struct {
				RunDir string `json:"runDir"`
			}
			if json.Unmarshal(b, &m) == nil && m.RunDir != "" {
				resolved := m.RunDir
				if !filepath.IsAbs(resolved) {
					resolved = filepath.Join(dir, resolved)
				}
				if st, err := os.Stat(resolved); err == nil && st.IsDir() {
					return resolved
				}
			}
			return "" // marker present but unusable: say nothing rather than guess
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Handler is a verb's work: everything cobra-shaped is handled around it.
type Handler func(Context, *cobra.Command) (string, error)

// New builds a verb command with the shared plumbing attached. The verb supplies
// its own name, contract text, flags and handler; it never restates the
// preconditions, the error prefix, or the render.
func New(role, name, help string, run Handler) *cobra.Command {
	c := &cobra.Command{
		Use:          name,
		Short:        help,
		Long:         help + "\n" + FrictionFooter,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			s := Of(cmd, role)
			if s.RunDir == "" {
				return fmt.Errorf("%s: --run <runDir> is required", role)
			}
			if s.SeatID == "" {
				return fmt.Errorf("%s: --seat-id is required (the engine assigns it; it is in your prompt)", role)
			}
			return record.CheckSeatRole(role, s.SeatID)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Set before the verb runs, so every record.Append in this process carries it.
			comment, cerr := flags.ReadComment(cmd, cmd.InOrStdin())
			if cerr != nil {
				return fmt.Errorf("%s: %w", role, cerr)
			}
			record.AmbientComment = comment
			out, err := run(Of(cmd, role), cmd)
			if err != nil {
				// The ROLE leads the message: a seat reading "close requires --id"
				// learns less than one reading "merge: close requires --id", because
				// the role names which contract it is being held to.
				return fmt.Errorf("%s: %w", role, err)
			}
			if out != "" {
				fmt.Println(out)
			}
			return nil
		},
	}
	// Render-on-mutation keeps projections current after every write. register is
	// exempt: it creates the seat rather than changing the board.
	if name != "register" {
		c.PostRunE = func(cmd *cobra.Command, _ []string) error {
			_, err := record.Render(Of(cmd, role).RunDir, "")
			return err
		}
	}
	// A UNIVERSAL FREE-TEXT FIELD, attached here so no verb can ship without one.
	//
	// The 2026-07-18 run's seats reached for --note, --detail, --target and --line and
	// were refused every time. That knowledge did not evaporate: it went into the
	// hand-written markdown, which is why the archive rendered from events was 7,527
	// bytes against the hand copy's 34,086. Every schema gap pushes evidence out of the
	// queryable channel and into the one nothing can query.
	//
	// --comment is the pressure valve. It is deliberately unstructured and deliberately
	// everywhere, and it doubles as the backlog: a note that keeps recurring names a
	// field the schema is missing, which is a better way to find them than guessing.
	flags.RegisterComment(c)

	return c
}

// Prose adds the shared payload channel to a verb that reads one.
func Prose(c *cobra.Command) *cobra.Command {
	c.Flags().String(flags.File, "", "read the prose payload from a file — ALWAYS use this over --text for anything above ~2KB")
	c.Flags().String(flags.Text, "", "the prose payload, inline (short values only)")
	return c
}

// Text resolves that channel: --file (read whole) or --text, else empty.
func Text(cmd *cobra.Command) (string, error) {
	if Given(cmd, flags.File) {
		b, err := os.ReadFile(Str(cmd, flags.File))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if Given(cmd, flags.Text) {
		return Str(cmd, flags.Text), nil
	}
	return "", nil
}

// Str reads a string flag.
func Str(cmd *cobra.Command, name string) string {
	v, err := cmd.Flags().GetString(name)
	if err != nil {
		return ""
	}
	return v
}

// Given answers whether the SEAT passed the flag. The absent/present distinction
// is load-bearing for the record format — a flag never passed must not appear in
// the event at all — and pflag's Changed is exactly that question.
func Given(cmd *cobra.Command, name string) bool {
	f := cmd.Flags().Lookup(name)
	return f != nil && f.Changed
}

// Set copies a flag into the payload only when the seat passed it.
func Set(cmd *cobra.Command, p *record.Payload, key, name string) *record.Payload {
	if Given(cmd, name) {
		p.Set(key, Str(cmd, name))
	}
	return p
}

// SetGrade writes a typed grade only when the seat passed it. The typed value
// carries the absent/present distinction itself, so this never consults the flag
// set — one fewer place for the two to disagree.
func SetGrade(p *record.Payload, key string, g *flags.GradeValue) *record.Payload {
	if g.Given() {
		p.Set(key, string(g.Grade))
	}
	return p
}

// SetList writes a comma-list field. These are ALWAYS present in the event, even
// empty: a gap with no ancestors records "supersedes": [], because an absent key
// would read as "lineage unknown" where the truth is "lineage none".
func SetList(p *record.Payload, key string, c *flags.CSV) *record.Payload {
	return p.Set(key, c.Value())
}

// SetSame is Set where the payload key and the flag name agree.
func SetSame(cmd *cobra.Command, p *record.Payload, names ...string) *record.Payload {
	for _, n := range names {
		Set(cmd, p, n, n)
	}
	return p
}
