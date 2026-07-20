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
func Of(cmd *cobra.Command) Context {
	runDir, _ := cmd.Flags().GetString(flags.Run)
	if runDir == "" {
		runDir = InferRunDir("")
	}
	seatID, _ := cmd.Flags().GetString(flags.SeatID)
	return Context{RunDir: runDir, SeatID: seatID, Role: roleOf(cmd)}
}

// roleOf reads the role from the command's POSITION in the tree: a verb's parent is its
// role node (Role sets the node's Use to the role name), so `merge mint`'s role is `merge`
// without anyone threading the string. Role is structure, not an argument.
func roleOf(cmd *cobra.Command) string {
	if p := cmd.Parent(); p != nil {
		return p.Name()
	}
	return ""
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

// Result is a verb's outcome: its OWN typed value, defined in the verb's own file, that
// renders itself as a human line. The value's json tags ARE the --json body; the verb name
// and the ok flag are supplied by the renderer below, so no result restates them. A handler
// returns a nil Result to say nothing.
type Result interface {
	Human() string
}

// Msg is the shared result for verbs that only CONFIRM — they carry no command-specific
// data. A verb WITH data defines its own struct instead; "a bare confirmation" is the one
// outcome genuinely common across verbs, so it is the one shared Result type.
type Msg struct {
	Message string `json:"message"`
}

// Human renders Msg as itself.
func (m Msg) Human() string { return m.Message }

// okEnvelope / errEnvelope are the --json wire shape. The Result nests under "result" so the
// envelope stays fully typed end to end — no map merging, no result restating verb/ok.
type okEnvelope struct {
	Verb   string `json:"verb"`
	OK     bool   `json:"ok"`
	Result Result `json:"result,omitempty"`
}

type errEnvelope struct {
	Verb  string `json:"verb"`
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func jsonMode(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool(flags.JSON)
	return b
}

// Handler is a verb's work: everything cobra-shaped is handled around it. It returns a
// Result (nil to say nothing); New renders it as human text or --json.
type Handler func(Context, *cobra.Command) (Result, error)

// markRequired annotates the flags a verb genuinely requires, reading record's single
// declaration so the help and the enforcement cannot disagree.
//
// A seat's contract is `--help`, and it was a flat alphabetical list: --check and --class
// (mandatory) sat indistinguishable from --comment and --found-by (optional). The only way
// to discover a requirement was to omit it and read the error — one round trip per missing
// flag, against a project whose binding constraint is wall clock.
//
// Applied AFTER the verb has registered its own flags, so a verb cannot forget to call it
// and no verb has to remember what it requires twice.
func markRequired(c *cobra.Command, verb string) {
	for _, key := range record.RequiredFields[verb] {
		f := c.Flags().Lookup(flags.ForPayloadKey(key))
		if f == nil {
			// The field is set by the verb rather than typed by the seat. Nothing to
			// annotate, and nothing wrong: not every payload key has a flag.
			continue
		}
		f.Usage = "REQUIRED — " + f.Usage
	}
}

// New builds a verb command with the shared plumbing attached. The verb supplies
// its own name, contract text, flags and handler; it never restates the
// preconditions, the error prefix, or the render.
func New(name, help string, run Handler) *cobra.Command {
	c := &cobra.Command{
		Use:          name,
		Short:        help,
		Long:         help + "\n" + FrictionFooter,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			s := Of(cmd)
			if s.RunDir == "" {
				return fmt.Errorf("%s: --run <runDir> is required", s.Role)
			}
			if s.SeatID == "" {
				return fmt.Errorf("%s: --seat-id is required (the engine assigns it; it is in your prompt)", s.Role)
			}
			return record.CheckSeatRole(s.Role, s.SeatID)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := Of(cmd)
			// Set before the verb runs, so every record.Append in this process carries it.
			comment, cerr := flags.ReadComment(cmd, cmd.InOrStdin())
			if cerr != nil {
				return fmt.Errorf("%s: %w", ctx.Role, cerr)
			}
			record.AmbientComment = comment
			res, err := run(ctx, cmd)
			if err != nil {
				// The ROLE leads the message: a seat reading "close requires --id"
				// learns less than one reading "merge: close requires --id", because
				// the role names which contract it is being held to. Under --json the
				// same message rides a structured error, so a machine consumer branches
				// on ok:false instead of parsing prose.
				prefixed := fmt.Errorf("%s: %w", ctx.Role, err)
				if jsonMode(cmd) {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(errEnvelope{Verb: cmd.Name(), Error: prefixed.Error()})
				}
				return prefixed
			}
			if jsonMode(cmd) {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(okEnvelope{Verb: cmd.Name(), OK: true, Result: res})
			}
			if res != nil {
				fmt.Fprintln(cmd.OutOrStdout(), res.Human())
			}
			return nil
		},
	}
	// Render-on-mutation is GONE (2026-07-19). It re-rendered every projection from the full
	// event log after every write — O(events) per mutation — to keep the markdown current
	// for the live dashboard. But `show` renders on read, the debate prompt renders
	// explicitly at the points it needs current files, and capture renders at the end; the
	// only cost of dropping it is that the dashboard can be seconds stale between renders,
	// which a monitor tolerates. Render is now a VISIBLE operation, not a hidden per-write tax.
	//
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
// Text resolves the payload channel through flags.ReadPayload — the ONE resolver.
//
// It used to do its own os.ReadFile, which made it a second reader of a concept the flags
// package already owned, and the two had drifted exactly as that always does: ReadPayload
// understood `--file -` as stdin, refused `--text` and `--file` together, and trimmed the
// trailing newline a shell heredoc leaves behind. This one did none of those, so every
// verb routed through it silently lacked stdin support while the capability sat one
// package away, written and tested.
//
// That gap was measured in the run: 68 commands carrying escaped quotes, 9 heredocs, and
// 37 staging a temp file first — two of which failed because the staged file was not there.
// Prose into markdown costs nothing; prose through the tool meant fighting the shell.
func Text(cmd *cobra.Command) (string, error) {
	return flags.ReadPayload(cmd, cmd.InOrStdin())
}

// SetLongForm fills a verb's own justification field from EITHER its named flag or the
// payload channel, and refuses both.
//
// The nine verbs with no payload channel were not a uniform gap. Six of them carry a field
// that is genuinely long-form — a dispute's --basis, a disposal's --reason, a spot-check's
// --notes — and those are exactly the values a seat had to inline, escape and quote,
// because the only alternative was the markdown. `cite` and `confidence` carry short
// values (a label, a grade) and want no payload; `verdict` is one word. Giving all nine
// the same channel would have been symmetry for its own sake.
//
// So --file/--text here is ANOTHER SPELLING of the named field, not a second field. Both
// given is refused rather than silently ranked, for the same reason ReadPayload refuses
// --text with --file: a seat that passes two should be told which one this verb would have
// ignored, not left to discover it in a projection three rounds later.
func SetLongForm(cmd *cobra.Command, p *record.Payload, key, flag string) error {
	named := Str(cmd, flag)
	payload, err := Text(cmd)
	if err != nil {
		return err
	}
	if named != "" && payload != "" {
		return fmt.Errorf("--%s and --file/--text are two spellings of this verb's %s: pass exactly one", flag, key)
	}
	if v := named + payload; v != "" {
		p.Set(key, v)
	}
	return nil
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
