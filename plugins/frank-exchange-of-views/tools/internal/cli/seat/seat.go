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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
needed and what you would have done with your role's 'friction' verb — a missing
capability is a finding about the tooling, and that channel is how it gets fixed.`

// RoleKey carries the seat's ROLE on the root command.
//
// It replaces roleOf's walk up the tree, and the walk is worth remembering. Role used to be read
// from a verb's POSITION — a verb's parent was its role node — which was true only while every
// verb's parent WAS its role. `show` became a group, so `blue show board`'s parent became `show`,
// and every projection read reported its role as the string "show". record.SittingOf switches on
// the role to decide what a seat owes; the switch matched nothing, and for weeks the only duty any
// seat received was the friction line that sits above it. Nothing reported it, because a switch
// that matches nothing falls through silently: absent duties and an honestly empty duty list are
// the same bytes.
//
// The role is not a fact about the tree. It is a fact about WHO WAS DISPATCHED — bound at
// `register` and read back from the record — and the tool derives the tree FROM it rather than
// recovering it from a path the seat had to type correctly.
const RoleKey = "feov.seat-role"

// SeatKey carries the resolved seat ID beside its role, for refusals that must name the party.
const SeatKey = "feov.seat-id"

// SetRole stamps the resolved identity on the root, where every verb can reach it.
//
// Read from HERE and not from the environment: the tests drive cobra in-process with SetArgs, so
// os.Args carries nothing, and a refusal that read the env would name an empty seat in exactly the
// runs that check it names the right one.
func SetRole(root *cobra.Command, role, seatID string) {
	if root.Annotations == nil {
		root.Annotations = map[string]string{}
	}
	root.Annotations[RoleKey] = role
	root.Annotations[SeatKey] = seatID
}

// DispatchedAs is the seat id this tree was built for.
func DispatchedAs(c *cobra.Command) string {
	if c == nil {
		return ""
	}
	return c.Root().Annotations[SeatKey]
}

// Context is what every verb needs and no verb should re-derive.
type Context struct {
	RunDir string
	SeatID string
	Role   string
	// Round is the seat's round as a FACT — injected by the dispatcher, not recovered from the
	// seat id (#348). -1 means unknown, which is NOT round 0: round 0 is synthesis, and
	// conflating the two is what produced the phantom-archive bug in #327.
	Round int
	// RunVia says which of the three paths supplied RunDir — the hook's injection, the seat's
	// own --run, or the tool's inference from the marker. `register` records it, because a run
	// whose seats all resolve by INFERENCE is a run the hook is not reaching, and nothing else
	// distinguishes that from a healthy one (#512).
	RunVia seatenv.RunSource
}

// Identity is what a record write needs to know about who is writing: the run, the seat, and the
// ROUND AS A FACT rather than a regex over the id. Every `record.Append` goes through this, so
// when the dispatcher starts injecting a round (#290) it lands in one place instead of 32.
//
// Role is not carried. See record.Event.Role: the party on an event is derived from the seat id,
// and this Context's Role answers a different question — which command group the verb sits under.
func (c Context) Identity() record.Identity {
	return record.Identity{RunDir: c.RunDir, SeatID: c.SeatID, Round: c.Round}
}

// Of reads the seat context from the inherited persistent flags, inferring the run
// directory when the flag is absent.
// It stays error-free by design: the resolution that CAN fail (an injected run directory
// disagreeing with a typed --run) is surfaced by Begin, which already has an error to return.
// Threading one through Of would have changed every caller for a case only Begin acts on.
// requireBound refuses an act by an agent that has not registered.
//
// THE ASYMMETRY IS THE MECHANISM. `register` is the one verb that may run unbound, because it is
// what CREATES the binding; every other verb is refused until it has. Without this the binding is
// advisory — a seat could skip register, keep typing --seat-id, and file events under any id the
// tree will build, which is the self-asserted identity this replaces.
//
// IT FIRES ONLY WHERE THE MECHANISM EXISTS. No agent handle means no hook, which means a test, an
// operator at a shell, or a harness with no PreToolUse — none of them can bind, and refusing them
// would be demanding a mechanism their environment does not have. The check is therefore keyed on
// the handle's PRESENCE, and its absence falls back to the flag exactly as before.
//
// COBRA ANSWERS --help WITHOUT REACHING RunE, so reading a surface is structurally exempt and needs
// no arm here: an unregistered seat can still discover what it is about to register for.
func requireBound(cmd *cobra.Command, s Context) error {
	agent := seatenv.AgentID()
	if agent == "" || cmd == nil || cmd.Name() == "register" {
		return nil
	}
	bound, found, err := record.SeatOfAgent(s.RunDir, agent)
	if err != nil {
		return err
	}
	if found && bound != "" {
		return nil
	}
	return feov.Errorf(feov.Validation,
		"`register` is your first act and it has not happened: nothing on the record binds this agent to a seat, "+
			"so %q is a claim rather than an identity. Run `register --seat-id %s` — it records who you are, and every "+
			"call after it resolves your seat from that binding instead of trusting the flag. This is the ONE call that "+
			"needs you to type your seat id.",
		s.SeatID, s.SeatID)
}

// BoundSeat resolves this process's agent handle to the seat it registered as, against a run.
//
// It is a closure rather than a value because BOTH call sites need it at a moment when the run
// directory has just been resolved and nothing has read the record yet — and because a nil one is
// the honest way to say "there is no run to look in", which is a different state from "this agent
// has not registered".
func BoundSeat(runDir string) func() (string, error) {
	if runDir == "" || seatenv.AgentID() == "" {
		return nil
	}
	return func() (string, error) {
		seat, _, err := record.SeatOfAgent(runDir, seatenv.AgentID())
		return seat, err
	}
}

func Of(cmd *cobra.Command) Context {
	runDir, _ := cmd.Flags().GetString(flags.Run)
	resolved, via, err := seatenv.ResolveWithSource(runDir, func() string { return InferRunDir("") })
	if err == nil {
		runDir = resolved
	}
	// Identity resolves the same way (#348): injected wins, a disagreeing flag is refused by
	// Begin, and the ROUND arrives as a field rather than being read back out of the id.
	seatID, _ := cmd.Flags().GetString(flags.SeatID)
	round := -1
	if id, rerr := seatenv.ResolveSeat(seatID, BoundSeat(runDir), record.RoundIn(runDir)); rerr == nil {
		seatID, round = id.ID, id.Round
	}
	return Context{RunDir: runDir, SeatID: seatID, Round: round, Role: roleOf(cmd), RunVia: via}
}

// roleOf answers WHICH SEAT is running this command, from the identity the engine injected.
//
// See RoleKey for what this replaced and what that cost.
func roleOf(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	return cmd.Root().Annotations[RoleKey]
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
// Records declares the event type a verb writes, when that is not the verb's own name.
//
// A verb and an event type stopped being 1:1 the moment the verbs that carried TWO CONTRACTS
// were split into the two verbs they already were — `lens verify` and `lens corroborate` both
// write a `verify`, `merge close` and `merge carry` both write a `close`. The record's
// required-field table is keyed by event type, so the lookup needs the event, not the word.
func Records(c *cobra.Command, eventType string) *cobra.Command {
	annotate(c, recordsKey, eventType)
	return c
}

// Supplied answers whether this verb fills the field itself, and why. The gates read it so the
// help check and the marking cannot disagree about the same flag.
func Supplied(c *cobra.Command, key string) string { return c.Annotations[suppliesKey+key] }

// Supplies marks a record field this verb fills ITSELF, with why.
//
// A REQUIRED FIELD IS NOT A REQUIRED FLAG, and conflating them produced help that contradicted
// itself in one line: `--status` read "proposed (put forward, undecided — the
// default)". Required, and also the default. A seat reading that supplies a value it did not
// need to choose, and the one state the field exists to express is the one it is least likely
// to type.
//
// Declared HERE, at the verb that does the supplying, rather than in a table keyed by verb name
// somewhere else: the fact and the code that makes it true stay in the same file, so one cannot
// move without the other being in front of you.
func Supplies(c *cobra.Command, key, why string) *cobra.Command {
	annotate(c, suppliesKey+key, why)
	return c
}

const (
	recordsKey  = "feov.records"
	suppliesKey = "feov.supplies."
)

func annotate(c *cobra.Command, key, val string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[key] = val
}

// satisfiedByAnyOf are payload keys a seat may supply through MORE THAN ONE flag. Cobra's
// per-flag required marking cannot express that — it would refuse the alternative — so these go
// through MarkFlagsOneRequired instead.
var satisfiedByAnyOf = map[string][]string{
	// The prose channel is one argument arriving three ways: inline, from a file, or from stdin
	// via `--reason-file -`. The file forms exist for prose too large to type.
	"reason": {flags.Reason, flags.ReasonFile},
	// mint copies --reason into `problem` when --problem is absent, and its help says so in the
	// same sentence it calls the field required.
	"problem": {flags.Problem, flags.Reason, flags.ReasonFile},
	// A line of inquiry's own statement and a manifest receipt are the verb's prose, so they
	// arrive through the same channel every other prose field does — including from a file.
	"line": {flags.Reason, flags.ReasonFile},
	"row":  {flags.Reason, flags.ReasonFile},
}

// markTree annotates every LEAF under a role. Verbs with subgroups are the ordinary shape now,
// and a loop over the role's direct children would have marked the group and left the verbs
// under it unmarked — help that says everything is optional, which is the silent half of exactly
// the defect this function exists to remove.
func markTree(c *cobra.Command) {
	if c.HasSubCommands() {
		for _, sub := range c.Commands() {
			markTree(sub)
		}
		return
	}
	if t := RecordType(c); t != "" {
		markRequired(c, t)
	}
}

// RecordType is the event type a verb writes, or "" for a command that writes no record.
//
// Exported because the gates that check the CLI against the record's tables have to spell a verb
// the same way the record does, and a verb name stopped being that spelling twice over: when the
// two-contract verbs were split (two verbs, one event type), and at the root, where the operator
// commands `verify` and `friction` share a word with seat verbs and write nothing at all. Only a
// command built by New carries the annotation, so the empty string is a fact, not a miss.
func RecordType(c *cobra.Command) string { return c.Annotations[recordsKey] }

func markRequired(c *cobra.Command, verb string) {
	for _, key := range record.RequiredFields[verb] {
		f := c.Flags().Lookup(flags.ForPayloadKey(key))
		if f == nil {
			// The field is set by the verb rather than typed by the seat. Nothing to
			// annotate, and nothing wrong: not every payload key has a flag.
			continue
		}
		if c.Annotations[suppliesKey+key] != "" {
			// Required of the RECORD, supplied by the VERB. Marking it would tell a seat to
			// type what it does not have to, and the annotation would sit beside the word
			// "default" in the same sentence.
			continue
		}
		// THE ONE WRITER OF THIS WORD. A dozen call sites also hand-wrote "REQUIRED — " at the
		// front of their own usage, so the rendered line read "REQUIRED — REQUIRED — what the
		// source ACTUALLY DID" on every flag that had both. Two copies of a fact the mechanism
		// already supplies, and the seat reading it twice cannot tell which one is authoritative.
		f.Usage = "REQUIRED — " + f.Usage

		// AND COBRA ENFORCES IT. This only ever rewrote the usage string, so the help said
		// REQUIRED and the parser accepted the command without it — the tool asserting a
		// constraint nothing held. Measured: `blue friction` and `blue revision`, no flags,
		// recorded empty events and took a seat from two outstanding duties to complete:true.
		//
		// The record layer refuses these too, and that is not a second reader of one rule: it
		// guards a different boundary, the one internal callers reach through record.Append
		// without a command line. Cobra's is the SEAT's boundary, and it is the one that can
		// refuse before an event exists and can say so in the help.
		if alts := satisfiedByAnyOf[key]; len(alts) > 0 {
			var present []string
			for _, a := range alts {
				if c.Flags().Lookup(a) != nil {
					present = append(present, a)
				}
			}
			if len(present) > 1 {
				c.MarkFlagsOneRequired(present...)
				continue
			}
		}
		_ = c.MarkFlagRequired(f.Name)
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
	// AND THE IDENTITY DISAGREEMENT, which Of's own comment said was refused here and which
	// nothing refused anywhere. Measured: with FEOV_SEAT=blue-respond-r1 injected, a call
	// passing --seat-id blue-respond-r9 was ACCEPTED and filed under r9 — a seat no dispatch
	// ever created, carrying its own register event and its own shard. Attribution is the one
	// fact a seat must not be able to get wrong; every found_by, estoppel and parity check
	// reads it, and this is the guarantee #348 shipped a message for and no code behind.
	seatFlag, _ := cmd.Flags().GetString(flags.SeatID)
	if _, err := seatenv.ResolveSeat(seatFlag, BoundSeat(Of(cmd).RunDir), record.RoundIn(Of(cmd).RunDir)); err != nil {
		return Of(cmd), err
	}
	s := Of(cmd)
	if s.RunDir == "" {
		return s, feov.Errorf(feov.MissingField, "--run <runDir> is required")
	}
	if s.SeatID == "" {
		return s, feov.Errorf(feov.MissingField, "--seat-id is required (state it once at `register`; after that it is bound and read back for you)")
	}
	if err := requireBound(cmd, s); err != nil {
		return s, err
	}
	// THE ROLE CHECK IS GONE BECAUSE THERE IS NOTHING LEFT TO RECONCILE. It compared the role in
	// the command PATH against the seat id, which was a real guard while a seat typed both: two
	// copies of one fact, and a seat could type them to disagree. The tree is built FROM the seat
	// id now, so s.Role is READ OUT OF that id — the check could only ever compare a value with
	// itself and pass. Keeping it would be the reconciliation of a fork that no longer exists.
	//
	// The half that still means something is that the seat is one the engine created, and it is
	// enforced where it can be: NewRootFor gives an unrecognised id no tree at all, so a verb
	// cannot be reached under an invented identity.
	if err := record.RequireDispatchedSeat(s.SeatID); err != nil {
		return s, err
	}
	if err := CheckFlagReferences(cmd, s.RunDir); err != nil {
		return s, err
	}
	return s, nil
}

// referenceChecker is what a typed flag implements when it knows what it is checked against.
// Declared here rather than imported so the flags package keeps no dependency on this one.
type referenceChecker interface {
	Check(runDir string) error
}

// CheckFlagReferences resolves every flag that carries an existence check.
//
// # Why HERE and not in a cobra hook
//
// A pflag.Value cannot do it: parsing is ONE left-to-right pass over argv with the parent's
// persistent flags merged in, so `close --id R1-1 --run <dir>` calls `--id`'s Set() before --run
// is bound. Measured. The check has to run after parse.
//
// The obvious after-parse seam is PersistentPreRunE, and it is a trap: a command defining its own
// SHADOWS the root's entirely, with no error and no warning — so a hook installed once at the
// root stops running the moment any verb grows one, silently. Measured too.
//
// `Begin` has neither problem. It is a plain function at the top of the one hook-free RunE that
// `New` wires for every seat verb, which is the shape this package already chose (see New): the
// run directory is resolved, nothing chains, and a failure renders through the same Emit as any
// other refusal. Coverage is DERIVED — a new verb is checked because it goes through Begin, not
// because someone remembered to add it to a list.
//
// The checks themselves are record.GapExists and friends — thin wrappers over the helpers
// `validate` uses. validate stays the enforcer, because it is the single write path every caller
// goes through; this is the earlier, better-placed copy of the same refusal, not a second one.
func CheckFlagReferences(cmd *cobra.Command, runDir string) error {
	var err error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		if c, ok := f.Value.(referenceChecker); ok {
			err = c.Check(runDir)
		}
	})
	return err
}

// taught marks an error the HANDLER already answered — diagnosed in the verb's own words, with
// whatever the seat needs next already in the sentence.
//
// It exists so the top level can tell those apart from the ones cobra raises BEFORE the handler
// runs: an unknown flag, a missing required flag, a flag group violated. Those say only which
// flag word was wrong, and the flag table is the answer to that — so the top level attaches the
// command's own help to them, and to nothing else.
//
// A FIELD, NOT A PATTERN. The alternative was matching cobra's message text ("required flag(s)
// ... not set"), which is a string shape from another module: it fails silently the day cobra
// rewords it, and a refusal that quietly stops teaching reads exactly like one that never needed
// to (see [[facts-are-fields]]).
type taught struct{ error }

func (t taught) Unwrap() error { return t.error }

// Taught reports whether a verb's handler already answered this error.
func Taught(err error) bool {
	var t taught
	return errors.As(err, &t)
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
		prefixed := taught{fmt.Errorf("%s: %w", role, err)}
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
// New builds a seat verb.
//
// TWO FIELDS, BECAUSE THEY ARE TWO JOBS, and they live in help/<name>.md rather than here. `## menu`
// is the line a seat reads in a listing, while it is deciding and has not chosen yet: what this
// verb is FOR, and when to reach for it rather than its neighbour. `## detail` is the page it reads
// once it has chosen: the flags, the sibling comparison, the measured history.
//
// IT WAS ONE STRING, SET INTO BOTH, and that is why the listing became unreadable. Everything a
// verb had to say went into Short — flag enumerations, sibling discriminators, the arguments from
// past runs — and cobra prints Short IN FULL in the parent's Available Commands block. `finding`
// alone ran to 961 characters there. A menu whose entries are each a paragraph is not a menu, and
// a seat facing one reads the first clause, which was the mechanics, and stops.
//
// The reverse cost was the same defect from the other side: Long was `help` again, so the page a
// seat opens after choosing told it nothing the listing had not already said. Both positions
// carried one string and neither did its own job.
func New(name string, run Handler) *cobra.Command { return NewKeyed(name, name, run) }

// NewKeyed is New for a verb whose help is not keyed by its own name — the shared verbs, where the
// same command means something different in each chair. `position` is a RED section for the merge
// and a BLUE one for blue, so those are two documents; `register` and `friction` mean exactly the
// same thing everywhere and are one.
//
// THEY WERE FOUR COPIES. Each role passed its own literal, under a comment saying the guidance
// differs by role — and register's and friction's were byte-identical at every call site but one,
// where the difference was the word "seat" against "sitting". Four hand-kept copies of one
// paragraph, which is the shape that goes stale in three places and nobody notices.
func NewKeyed(name, key string, run Handler) *cobra.Command {
	d := helpFor(key)
	c := &cobra.Command{
		Use:          name,
		Short:        d.menu,
		Long:         d.menu + "\n\n" + d.detail + "\n" + FrictionFooter,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
		Annotations:  map[string]string{recordsKey: name},
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

// ProseRequired adds the channel and makes it mandatory THROUGH EITHER SPELLING.
//
// A verb that reached for cobra's MarkFlagRequired got a rule naming one flag, which refuses the
// file form — so `spot-check`, whose reason is always owed, could not be discharged with a
// heredoc. One argument, one requirement, stated by the channel rather than by each verb.
func ProseRequired(c *cobra.Command) *cobra.Command {
	new(flags.Prose).RegisterRequired(c)
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

// GroupTitles are the two headings a seat's root help renders its surface under.
//
// WHY THE SPLIT EXISTS. Cobra lists every child under one "Available Commands:" heading, and a
// command holding subcommands looks exactly like one that does not — same line, same shape. So the
// root page of a seat surface silently understated itself: `motion` reads as a verb, and the three
// commands inside it are not on any page the seat has opened.
//
// The prompt used to carry the missing half, in a sentence appended to step 1 of the tree walk:
// "Some of those entries are GROUPS, holding commands this page does not show you." That is the
// tool describing its own output format to the reader, in the prompt, because the output would not
// describe itself — and it is exactly the shape this pass is removing. A page that cannot say what
// it is hiding needs fixing, not narrating.
const (
	GroupVerbs  = "verbs"
	GroupGroups = "groups"
)

// SplitGroups sorts a root's children into the two headings. Called AFTER every AddCommand, since
// it reads HasSubCommands to decide — a group assigned before its children are attached would
// classify itself as a leaf.
func SplitGroups(root *cobra.Command) {
	root.AddGroup(
		&cobra.Group{ID: GroupVerbs, Title: "Available Commands:"},
		&cobra.Group{ID: GroupGroups, Title: "Command groups — each holds commands THIS page does not list:"},
	)
	for _, sub := range root.Commands() {
		if sub.HasSubCommands() {
			sub.GroupID = GroupGroups
			continue
		}
		sub.GroupID = GroupVerbs
	}
	// `help` and `completion` are cobra's, added lazily at execute time; without these they land
	// under an "Additional Commands:" heading of their own, which reads as a third category of
	// seat surface rather than as the shell plumbing they are.
	root.SetHelpCommandGroupID(GroupVerbs)
	root.SetCompletionCommandGroupID(GroupVerbs)
}
