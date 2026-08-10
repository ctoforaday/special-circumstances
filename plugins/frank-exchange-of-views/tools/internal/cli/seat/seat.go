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

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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
	// Round is the seat's round as a FACT — injected by the dispatcher, not recovered from the
	// seat id (#348). -1 means unknown, which is NOT round 0: round 0 is synthesis, and
	// conflating the two is what produced the phantom-archive bug in #327.
	Round int
}

// Of reads the seat context from the inherited persistent flags, inferring the run
// directory when the flag is absent.
// It stays error-free by design: the resolution that CAN fail (an injected run directory
// disagreeing with a typed --run) is surfaced by Begin, which already has an error to return.
// Threading one through Of would have changed every caller for a case only Begin acts on.
func Of(cmd *cobra.Command) Context {
	runDir, _ := cmd.Flags().GetString(flags.Run)
	resolved, err := seatenv.Resolve(runDir, func() string { return InferRunDir("") })
	if err == nil {
		runDir = resolved
	}
	// Identity resolves the same way (#348): injected wins, a disagreeing flag is refused by
	// Begin, and the ROUND arrives as a field rather than being read back out of the id.
	seatID, _ := cmd.Flags().GetString(flags.SeatID)
	round := -1
	if id, rerr := seatenv.ResolveSeat(seatID, record.RoundOf); rerr == nil {
		seatID, round = id.ID, id.Round
	}
	return Context{RunDir: runDir, SeatID: seatID, Round: round, Role: roleOf(cmd)}
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

// okEnvelope / errEnvelope are the --json wire shape. verb and role come from the command's
// position (not string-mashed into the message); the Result nests under "result" so the
// envelope stays fully typed end to end. An error carries a `code` a consumer switches on.
type okEnvelope struct {
	Verb   string `json:"verb"`
	Role   string `json:"role"`
	OK     bool   `json:"ok"`
	Result Result `json:"result,omitempty"`
}

type errEnvelope struct {
	Verb  string `json:"verb"`
	Role  string `json:"role"`
	OK    bool   `json:"ok"`
	Code  string `json:"code"`
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
// suppliedByTheVerb are (verb, payload key) pairs the RECORD requires and the VERB fills, with
// what fills them.
//
// A REQUIRED FIELD IS NOT A REQUIRED FLAG, and conflating them produced help that contradicted
// itself in one line: `blue avenue --status` read "REQUIRED — proposed (put forward, undecided —
// the default)". Required, and also the default. A seat reading that supplies a value it did not
// need to choose, and the one state the field exists to express — a direction PUT FORWARD and not
// yet resolved — is the one a seat is least likely to type.
//
// Caught by TestEveryRequiredFlagIsActuallyRefused, which ran the verb without the flag and it
// worked.
var suppliedByTheVerb = map[string]string{
	"avenue.status": "a fresh proposal defaults to `proposed` in the verb — the state the old shape could not express, and the one a seat would not think to type. A MOVE still requires it, and the move's own refusal says so",
}

func markRequired(c *cobra.Command, verb string) {
	for _, key := range record.RequiredFields[verb] {
		f := c.Flags().Lookup(flags.ForPayloadKey(key))
		if f == nil {
			// The field is set by the verb rather than typed by the seat. Nothing to
			// annotate, and nothing wrong: not every payload key has a flag.
			continue
		}
		if suppliedByTheVerb[verb+"."+key] != "" {
			// Required of the RECORD, supplied by the VERB. Marking it would tell a seat to
			// type what it does not have to, and the annotation would sit beside the word
			// "default" in the same sentence.
			continue
		}
		f.Usage = "REQUIRED — " + f.Usage
	}
}

// Begin runs a verb's preconditions and returns its seat context. It is a plain function
// called at the TOP of the one RunE — NOT a PreRunE hook — so there is no lifecycle chaining
// to reason about. It requires the run dir and seat id and holds the seat to its role.
func Begin(cmd *cobra.Command) (Context, error) {
	// The disagreement refusal fires FIRST — before "--run is required" and before the seat-id
	// checks — because a seat pointed at the wrong run must not get a message about anything
	// else. Of() swallows the error to stay signature-compatible; this is where it is honoured.
	flagRun, _ := cmd.Flags().GetString(flags.Run)
	if _, err := seatenv.Resolve(flagRun, nil); err != nil {
		return Of(cmd), err
	}
	s := Of(cmd)
	if s.RunDir == "" {
		return s, feov.Errorf(feov.MissingField, "--run <runDir> is required")
	}
	if s.SeatID == "" {
		return s, feov.Errorf(feov.MissingField, "--seat-id is required (the engine assigns it; it is in your prompt)")
	}
	if err := record.CheckSeatRole(s.Role, s.SeatID); err != nil {
		return s, err
	}
	return s, nil
}

// Emit renders a verb's outcome. It is the ONE place both a success and a failure are
// rendered, so they cannot disagree, and the ONE place a --json error's code is read — from
// the error's own type via feov.CodeOf, with no switch on concrete error types. verb and
// role come from the command's POSITION, not from string-mashing the message.
func Emit(cmd *cobra.Command, res Result, err error) error {
	role := roleOf(cmd)
	if err != nil {
		// The ROLE leads the human message ("merge: close requires --id") — it names which
		// contract is being held. Under --json the role is a field, the code is the machine
		// handle a consumer branches on, and the message stays the clean domain sentence.
		prefixed := fmt.Errorf("%s: %w", role, err)
		if jsonMode(cmd) {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(errEnvelope{
				Verb: cmd.Name(), Role: role, Code: feov.CodeOf(err), Error: prefixed.Error(),
			})
		}
		return prefixed
	}
	if jsonMode(cmd) {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(okEnvelope{
			Verb: cmd.Name(), Role: role, OK: true, Result: res,
		})
	}
	if res != nil {
		fmt.Fprintln(cmd.OutOrStdout(), res.Human())
	}
	return nil
}

// New builds a verb command. The verb supplies its name, contract text, flags and work; New
// wires the ONE hook-free RunE that shape holds: Begin (preconditions) -> the work -> Emit
// (render). No PreRunE, no PostRunE — nothing chains, and a precondition failure renders
// through the same Emit as any other error.
func New(name, help string, run Handler) *cobra.Command {
	c := &cobra.Command{
		Use:          name,
		Short:        help,
		Long:         help + "\n" + FrictionFooter,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := Begin(cmd)
			if err != nil {
				return Emit(cmd, nil, err)
			}
			res, werr := run(s, cmd)
			return Emit(cmd, res, werr)
		},
	}
	// Render-on-mutation is GONE (2026-07-19). It re-rendered every projection from the full
	// event log after every write — O(events) per mutation — to keep the markdown current
	// for the live dashboard. But `show` renders on read, the debate prompt renders
	// explicitly at the points it needs current files, and capture renders at the end; the
	// only cost of dropping it is that the dashboard can be seconds stale between renders,
	// which a monitor tolerates. Render is now a VISIBLE operation, not a hidden per-write tax.
	//
	// The universal --comment field is GONE too (2026-07-20). It was the pressure valve for
	// "what this verb has no field for", but a comment was never surfaced in any report, so
	// prose put there was lost to the reader — a junk drawer that taught seats to record
	// substance where nothing could read it. A missing field is now a friction entry (a
	// finding about the tooling) rather than text dropped into an unqueryable channel.

	return c
}

// Prose adds the shared prose payload channel — --reason / --reason-file — to a verb that
// reads one. It routes through flags.RegisterPayload rather than registering the flags by
// hand, so a verb cannot register one form and forget the other; `close` shipped exactly
// that leak (a private --file, no --text) before the channel was centralized here.
func Prose(c *cobra.Command) *cobra.Command {
	flags.RegisterPayload(c)
	return c
}

// Reason resolves that channel through flags.ReadPayload — the ONE resolver: --reason
// inline, --reason-file from disk, or `--reason-file -` from stdin, with both-given refused
// and the trailing newline a shell heredoc leaves behind trimmed off.
//
// Routing every verb through the one resolver is what gives them all stdin support for free.
// The gap it closed was measured in the 2026-07-18 run: 68 commands carrying escaped quotes,
// 9 heredocs, and 37 staging a temp file first — two of which failed because the staged file
// was not there. Prose into markdown costs nothing; prose through the tool meant fighting
// the shell.
func Reason(cmd *cobra.Command) (string, error) {
	return flags.ReadPayload(cmd, cmd.InOrStdin())
}

// Str reads a string flag.
func Str(cmd *cobra.Command, name string) string {
	// ONE IMPLEMENTATION, in flags.Value. This used to hand-roll the same fallback and
	// flags.Set hand-rolled the broken version, which is how the identical defect shipped twice.
	return flags.Value(cmd, name)
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

// SetReason reads the seat's prose (--reason / --reason-file / stdin) and, when non-empty, sets it
// under key. It folds the read+error-guard+conditional-set that every prose-bearing verb otherwise
// hand-rolls (evidence / basis / rationale / prose / reason) — close.go documents that a verb which
// opts out of the shared READ drifts by construction; the same holds for the SET.
func SetReason(cmd *cobra.Command, p *record.Payload, key string) error {
	r, err := Reason(cmd)
	if err != nil {
		return err
	}
	if r != "" {
		p.Set(key, r)
	}
	return nil
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
