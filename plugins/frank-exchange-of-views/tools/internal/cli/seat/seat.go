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
	"reflect"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/jsonshape"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// LogFooter closes the loop the help opens. A seat that needs something the
// contract does not offer must not improvise around it — the missing capability
// is itself a fact for the log, and `log` carries it to the operator who can
// retool the seat.
const LogFooter = `
If you need a verb or a flag that is not listed here, it does not exist for you:
do not improvise around it, and do not hand-write the artifact. Record what you
needed and what you would have done with 'log', as a request — a missing
capability is a fact about the tooling, and the log is how it reaches the operator
who can retool you.`

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
// GradeOrNil is SetGrade for a proto body: the typed value when the seat passed it, and nil when
// it did not. ABSENT AND ZERO ARE DIFFERENT and the distinction is the whole reason this returns a
// pointer — Grade's zero is UNSPECIFIED, so writing it for an omitted flag would record a grade the
// seat never gave, and every downstream reader would see a graded axis rather than an ungraded one.
//
// It refuses rather than guesses on a word the schema does not know: record.GradeOf returns false,
// and nil here means the field stays absent, which is what the pre-migration record did when the
// key was never set.
func GradeOrNil(g *flags.GradeValue) *recordpb.Grade {
	if !g.Given() {
		return nil
	}
	v, ok := record.GradeOf(string(g.Grade))
	if !ok {
		return nil
	}
	return &v
}

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
	// runDir is PRIVATE so no verb can read the path without meeting the refusal.
	//
	// It was exported, and two readers took it straight off seat.Of(cmd) where there was no
	// error to consult even in principle. One of them then did the thing this type's RunErr
	// comment predicted in as many words — answered "" with InferRunDir(""), swallowing the
	// refusal a layer down and resolving quietly to a different run. Run() is now the only way
	// in, and it hands back a record.Run that a mistyped path cannot construct.
	runDir string
	SeatID string
	Role   string
	// There is no epoch here. The epoch (chair sittings) and the sitting ordinal are windows the
	// record computes over its stored sittings at read time (events_w); a seat never carries
	// or stamps either. The epoch was once injected by the dispatcher and before that recovered
	// from the seat id by regex (#348).
	// RunVia says which of the three paths supplied the run — the hook's injection, the seat's
	// own --run, or the tool's inference from the marker. `register` records it, because a run
	// whose seats all resolve by INFERENCE is a run the hook is not reaching, and nothing else
	// distinguishes that from a healthy one (#512).
	RunVia seatenv.RunSource

	// RunErr is the resolution's REFUSAL, carried rather than encoded as an empty RunDir.
	//
	// MEASURED, on the first real exercise of the wrapper (#526). Of() returned the flag's own
	// value when resolution failed, so `show board --run "<typo>"` printed a board for a
	// directory nobody had dispatched — the 2026-08-05 failure this whole mechanism exists to
	// prevent, still live on every READ verb because only Begin honoured the refusal.
	//
	// Returning an empty runDir instead does not fix it, and the reason is worth keeping after
	// the fact: readers used to answer "" with a SECOND `runlive.InferRunDir("")` below Of()'s own,
	// so the refusal was swallowed one layer down and resolved quietly to the real run. Those
	// fallbacks are gone now — reachable only on a refusal, they were only ever wrong — but the
	// argument is what keeps them from coming back. "" already means "nothing supplied one";
	// making it also mean "you were refused" is the collapse the record keeps paying for — two
	// states, one byte, and the healthy one wins by default.
	RunErr error

	// seatErr is the identity resolution's refusal: a --seat-id disagreeing with the seat the record
	// binds this agent to. Carried for the same reason RunErr is — Of keeps the flag's value as
	// SeatID, so a reader that never consults this acts under an identity the record contradicts.
	// Run, RequireRun and Begin all refuse on it.
	seatErr error

	// bindErr is a binding LOOKUP that could not answer — a record the query cannot read. It is
	// not a disagreement: nothing contradicts the flag, the record simply could not say. A READ
	// proceeds under the flag's seat, as it would with no agent handle at all; a WRITE is refused
	// by Begin, because a write must be attributed to the registration, and here nothing could
	// read one.
	bindErr error

	// bound is the seat the record binds this agent to — its latest register, else its seat-naming
	// bracket — when the lookup answered one. A refusal of a disagreeing --seat-id is the bound
	// seat's act, and is logged under it (see RefusalSeat).
	bound string
}

// Run resolves the seat's run, or returns why it could not.
//
// One door, and it cannot be walked past. The pair this replaces — an exported RunDir beside
// a RunErr — could be half-read, and was: only Begin honoured the refusal, so every read verb
// took the path and reported on a run nobody had dispatched (#526).
//
// OpenRun rather than NewRun: a verb holding a Context is acting on a run that already exists,
// and the existence check is what turns a typo into a refusal instead of an empty board.
func (c Context) Run() (record.Run, error) {
	if c.RunErr != nil {
		return record.Run{}, c.RunErr
	}
	r, err := record.OpenRun(c.runDir)
	if err != nil {
		return record.Run{}, err
	}
	// AFTER the existence check: a run that is not there is the first thing to fix, and a binding
	// lookup on it answers nothing.
	if c.seatErr != nil {
		return record.Run{}, c.seatErr
	}
	return r, nil
}

// RefusalSeat is the run and the seat a refusal of this call is logged under.
//
// THE BINDING, NOT THE FLAG. A --seat-id that disagrees with the seat the record binds this agent
// to is refused, and Run refuses with it — so a logger reading the run through Run could never
// record that refusal at all. The act was the bound seat's, typing a wrong id, and the entry
// is attributed to it. Everything else is Run's order: the run's resolution, then its existence.
func (c Context) RefusalSeat() (record.Run, string, error) {
	if c.RunErr != nil {
		return record.Run{}, "", c.RunErr
	}
	r, err := record.OpenRun(c.runDir)
	if err != nil {
		return record.Run{}, "", err
	}
	if c.bound != "" {
		return r, c.bound, nil
	}
	return r, c.SeatID, nil
}

// handle resolves this context's run WITHOUT the existence check, for the internal best-effort
// reads that must work on a context whose run may not be on disk. Anything acting on the run
// wants Run(), which refuses. Unexported for exactly that reason.
func (c Context) handle() record.Run {
	r, _ := record.NewRun(c.runDir)
	return r
}

// Identity is what a record write needs to know about who is writing: the run and the seat.
// Every `record.Append` goes through this, so anything the envelope must carry lands in one
// place instead of 32. No clock rides on it: the record derives epoch and sitting on read.
//
// Role is not carried. See record.PartyOf: the party on an event is the role of its seat id,
// and this Context's Role answers a different question — which command group the verb sits under.
func (c Context) Identity() record.Identity {
	// The zero Run when this context was REFUSED. Append asserts Valid() and refuses rather
	// than writing, so a caller that skipped Run() still cannot record into an unresolved run —
	// which is what the old `RunDir: ""` did, into the working directory.
	return record.Identity{Run: c.handle(), SeatID: c.SeatID}
}

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
// AND WHERE THE TOOL CAN ESTABLISH THE BINDING ITSELF, IT DOES, INSTEAD OF REFUSING — see
// autoRegisters. The refusal was asking most seats to state a fact the tool already held, and the
// register it asked for is also what arms the per-sitting call cap, so an agent allowed to write
// without one ran uncapped.
//
// COBRA ANSWERS --help WITHOUT REACHING RunE, so reading a surface is structurally exempt and needs
// no arm here: an unregistered seat can still discover what it is about to register for.
func requireBound(cmd *cobra.Command, s Context) error {
	agent := seatenv.AgentID()
	if agent == "" || cmd == nil || cmd.Name() == "register" {
		return nil
	}
	// THE STRICT LOOKUP, not SeatOfAgent: the harness bracket answers who a seat IS, and that is a
	// read. What gates a write is whether a REGISTER exists, because a register opens the sitting the
	// per-call cap counts in. Accepting the bracket here made a bracket-bound seat uncapped.
	bound, found, err := record.RegisteredSeatOfAgent(s.handle(), agent)
	if err != nil {
		return err
	}
	if found && bound != "" {
		return nil
	}
	if autoRegisters(cmd) {
		return autoRegister(cmd, s)
	}
	return feov.Errorf(feov.Validation,
		"`register` is your first act and it has not happened: this agent has no register on the record, "+
			"so %q is a claim rather than an identity. Run `register --seat-id %s` — it records who you are, and every "+
			"call after it resolves your seat from the record instead of trusting the flag. This is the ONE call that "+
			"needs you to type your seat id.",
		s.SeatID, s.SeatID)
}

// autoRegisters says whether this surface's `register` can be performed FOR the seat, and it asks
// the surface rather than a list of roles — the fact lives once, where the difference is declared.
//
// A register carries the binding, and on one surface it carries one thing more, as a flag:
//
//	bench   --occasion        which of its four questions this sitting answers. Only its prompt knows.
//
// It can be supplied by nothing but the seat, so it cannot be silent. Every other surface's
// register is the binding and nothing else, which is a fact the tool already holds — so the seat
// spends a call telling the tool what the tool knows, and that call is what this removes.
//
// DERIVED, NEVER LISTED. A future flag on some surface's register means new information only the
// seat has, and the honest default for that is to stop registering it silently — which happens by
// itself here, with no second roster to keep in step.
func autoRegisters(cmd *cobra.Command) bool {
	for _, c := range cmd.Root().Commands() {
		if c.Name() != "register" {
			continue
		}
		bespoke := false
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "help" {
				bespoke = true
			}
		})
		return !bespoke
	}
	return false
}

// autoRegister registers this agent as the seat it says it is, silently, on its first act.
//
// SILENT IS THE POINT. The seat is not told, because there is nothing for it to decide: the
// alternative was a refusal telling it to run a command whose whole content is a fact the tool
// already had. Measured on universe-m11 and m12: 17 of one run's calls were registers, and 8 of 9
// sittings with nothing owed registered anyway — the tool's answer rather than the seat's.
//
// IT MUST DO EVERYTHING THE VERB DOES, which is why it calls the same handler rather than only the
// record write. The sitting's tool-call cap is armed BY register (sittingcap keys on its header), so
// a silent register that skipped that would buy back the call and lose the cap — the exact hole this
// change closes.
func autoRegister(cmd *cobra.Command, s Context) error {
	if _, err := register(s, cmd); err != nil {
		return fmt.Errorf("registering %s for its first act: %w", s.SeatID, err)
	}
	return nil
}

// BoundSeat resolves this process's agent handle to the seat the record binds it to — its latest
// register, else its seat-naming bracket — against a run.
//
// It is a closure rather than a value because BOTH call sites need it at a moment when the run
// directory has just been resolved and nothing has read the record yet — and because a nil one is
// the honest way to say "there is no run to look in", which is a different state from "the record
// binds this agent to no seat".
func BoundSeat(run record.Run) func() (string, error) {
	// `!run.Valid()` is the same test the empty string used to make, and now it cannot be
	// confused with a run whose path merely happens to be short: a Run is valid or it resolved
	// from nothing.
	if !run.Valid() || seatenv.AgentID() == "" {
		return nil
	}
	return func() (string, error) {
		seat, _, err := record.SeatOfAgent(run, seatenv.AgentID())
		return seat, err
	}
}

// Of reads the seat context from the inherited persistent flags, inferring the run
// directory when the flag is absent.
//
// It returns no error: the two resolutions that can fail — the run and the seat identity — are
// carried on the Context as RunErr and seatErr, and Run, RequireRun and Begin refuse on them. A
// binding lookup that could not answer is carried apart, as bindErr, and only Begin refuses on it.
//
// EVERY VERB THAT WRITES TO THE RUN — an event or a document — goes through Begin (see
// HandlerRunE), which also holds the agent to register-first. The one write that does not is the
// tool's own refusal-log entry, which is written BECAUSE Begin refused, and is attributed to the
// seat the record binds this agent to, through RefusalSeat.
func Of(cmd *cobra.Command) Context {
	runDir, _ := cmd.Flags().GetString(flags.Run)
	resolved, via, err := seatenv.ResolveWithSource(runDir, func() string {
		return runlive.InferRunDir("", runlive.Caller{ID: seatenv.AgentID(), Type: seatenv.AgentType()}).Dir
	})
	if err != nil {
		// NO RUN DIRECTORY LEAVES HERE. A caller holding one it was refused is a caller that
		// will use it, and every reader below this point trusts what it is handed.
		return Context{RunErr: err, Role: roleOf(cmd), RunVia: seatenv.RunUnresolved}
	}
	runDir = resolved
	// Identity resolves the same way (#348): the binding wins, and a disagreeing flag is carried
	// as seatErr rather than kept as the seat. There is no epoch here: the record computes the
	// epoch at each write.
	seatID, _ := cmd.Flags().GetString(flags.SeatID)
	// NewRun, not OpenRun: Of() must hand back a Context even for a run that is not on disk —
	// the existence check belongs at Run(), where a verb actually acts on the run.
	run, _ := record.NewRun(runDir)
	c := Context{runDir: runDir, Role: roleOf(cmd), RunVia: via}
	// THE LOOKUP ONCE, and its two failures apart. A binding that DISAGREES with the flag is a
	// refusal for every verb; a lookup that cannot answer (a record the query cannot read) is not,
	// for a read — it is refused at Begin, where a write would otherwise go unattributed.
	if lookup := BoundSeat(run); lookup != nil {
		c.bound, c.bindErr = lookup()
	}
	if c.bindErr != nil {
		c.SeatID = seatID
		return c
	}
	id, serr := seatenv.ResolveSeat(seatID, func() (string, error) { return c.bound, nil })
	if serr != nil {
		c.SeatID, c.seatErr = seatID, serr
		return c
	}
	c.SeatID = id.ID
	return c
}

// roleOf answers WHICH SEAT is running this command, from the identity the engine injected.
//
// See RoleKey for what this replaced and what that cost.
// RequireRun is the one way a verb turns a Context into a run directory it may act on.
//
// It answers the two failures separately, because their remedies are opposite: a REFUSED
// resolution is the seat's own contradicting --run (drop the flag), and an EMPTY one is nothing
// having supplied a run at all (the engine is not dispatching you). Collapsing them into
// "--run is required" tells a seat that DID pass --run to pass it, which is the message that
// sends it looking in the wrong place.
func (c Context) RequireRun() (record.Run, error) {
	if c.RunErr != nil {
		return record.Run{}, c.RunErr
	}
	if c.runDir == "" {
		return record.Run{}, feov.Errorf(feov.MissingField, "--run <runDir> is required")
	}
	// The message above is kept for the UNSUPPLIED case — the renderer names the invocation it
	// was raised on, which tells an operator which call to fix. Everything past it is the same
	// refusal Run() gives, for the same reason.
	return c.Run()
}

func roleOf(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	return cmd.Root().Annotations[RoleKey]
}

// Result is a verb's outcome: its OWN typed value, defined in the verb's own file, that
// renders itself as a human line. The value's json tags ARE the --json body; the verb name
// and the ok flag are supplied by the renderer below, so no result restates them. A handler
// returns a nil Result to say nothing.
type Result interface {
	Human() string
}

// Relayed is a result whose machine form IS the object the seat relays, and nothing else: --json
// prints Relay() alone, with no {verb, role, ok, result, standing} around it. A refusal keeps the
// refusal envelope.
//
// A COPY, NOT A COMPOSITION. The chair's envelope IS `dispatch next`'s plan. While the verb printed
// the plan as `result` inside the ok envelope, the chair lifted it out and renamed it, and about
// one sitting in five (m26-m30, 7 of 34) nested it and was refused. Printing exactly the relayed
// object makes "copy what the verb printed" and "return the envelope" one act.
type Relayed interface {
	Result
	Relay() any
}

// Msg is the shared result for verbs that only CONFIRM — they carry no command-specific
// data. A verb WITH data defines its own struct instead; "a bare confirmation" is the one
// outcome genuinely common across verbs, so it is the one shared Result type.
type Msg struct {
	Message string `json:"message"`
}

// Human renders Msg as itself.
func (m Msg) Human() string { return m.Message }

// okEnvelope is the --json wire shape of a success. verb and role come from the command's
// position (not string-mashed into the message); the Result nests under "result" so the
// envelope stays fully typed end to end. A refusal's shape is RefusalEnvelope.
type okEnvelope struct {
	Verb   string `json:"verb"`
	Role   string `json:"role"`
	OK     bool   `json:"ok"`
	Result Result `json:"result,omitempty"`
	// Standing is where the seat stands AFTER this act — absent on a read, which moved nothing.
	// See standing.go: it is what ends `show work` after an act, and it carries the verdict rather
	// than the list because the list is 4-5 KB and the verdict is the question.
	Standing *StandingJSON `json:"standing,omitempty"`
}

// readOutput is a read verb's --json success as its page states it: okEnvelope with the verb's
// own result type in the place of the interface, so the tree is generated from the two types that
// are marshalled.
func readOutput(result Result) string {
	env := reflect.TypeOf(okEnvelope{})
	fields := make([]reflect.StructField, 0, env.NumField())
	for i := 0; i < env.NumField(); i++ {
		f := env.Field(i)
		if f.Type == reflect.TypeOf((*Result)(nil)).Elem() {
			f.Type = reflect.TypeOf(result)
		}
		fields = append(fields, f)
	}
	tree := jsonshape.Tree(reflect.New(reflect.StructOf(fields)).Elem().Interface())
	if tree == "" {
		return ""
	}
	return "\n\nOUTPUT (JSON, with --json — the bare call is the human form): " + tree
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
// write a `verify`, `lens close` and `chair carry` both write a `close`. The record's
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
	armKey      = "feov.arm"
	suppliesKey = "feov.supplies."
)

func annotate(c *cobra.Command, key, val string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[key] = val
}

// satisfiedByAnyOf are FLAGS a seat may supply through more than one flag. Cobra's per-flag
// required marking cannot express that — it would refuse the alternative — so these go through
// MarkFlagsOneRequired instead.
//
// KEYED ON THE FLAG, NOT THE FIELD. It was keyed on payload keys, which worked while requiredness
// came from a hand-written table that spoke in the seat's words ("reason"). It is derived from the
// schema now and speaks in the schema's: a close stores `prose`, an opinion `rationale`, a halt
// `opinion`, a certification `statement`. Keyed on fields this map would need an entry per prose
// field and would go stale one field at a time; keyed on the flag it states the rule once, which
// is what the rule actually is — the prose channel is ONE argument arriving three ways.
var satisfiedByAnyOf = map[string][]string{
	// The prose channel is ONE spelling (flags.Prose says why the file forms went). Listed so a
	// prose FIELD is required through the flag a seat types rather than the key the schema stores.
	"reason": {flags.Reason},
	// An inquiry's question arrives through its own flag; a manifest receipt is the verb's prose,
	// so it arrives through the same channel every other prose field does.
	"question": {flags.Question},
	"row":      {flags.Reason},
}

// markTree annotates every LEAF under a role. Verbs with subgroups are the ordinary shape now,
// and a loop over the role's direct children would have marked the group and left the verbs
// under it unmarked — help that says everything is optional, which is the silent half of exactly
// the defect this function exists to remove.
// MarkTree is markTree for a command tree mounted outside the role verbs.
//
// THE MOTION GROUP WAS NEVER WALKED. seatVerbs applies markTree to every role verb — "a verb that
// had to remember to mark its own required flags is a verb that can forget, and the forgetting is
// silent" — but `motion` is mounted separately at root.go, so its leaves got none of it. It looked
// fine only because refHelp hand-wrote "REQUIRED — " into `--id`'s usage, which is the second copy
// markRequired exists to be the sole writer of.
func MarkTree(c *cobra.Command) { markTree(c) }

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
	teachAnswer(c)
}

// RecordType is the event type a verb writes, or "" for a command that writes no record.
//
// Exported because the gates that check the CLI against the record's tables have to spell a verb
// the same way the record does, and a verb name stopped being that spelling twice over: when the
// two-contract verbs were split (two verbs, one event type), and at the root, where the operator
// command `verify` shares a word with a seat verb and writes nothing at all. Only a
// command built by New carries the annotation, so the empty string is a fact, not a miss.
func RecordType(c *cobra.Command) string { return c.Annotations[recordsKey] }

// FillsArm declares which arm of its body's oneof a verb fills — `docket` on `motion docket rule`
// — so the fields that arm's message requires are marked REQUIRED on this verb's page and on no
// other verb writing the same record type.
func FillsArm(c *cobra.Command, arm string) *cobra.Command {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[armKey] = arm
	return c
}

// Arm is the oneof arm a verb declared it fills, or "".
func Arm(c *cobra.Command) string { return c.Annotations[armKey] }

func markRequired(c *cobra.Command, verb string) {
	// A declared arm the body does not have marks nothing and says nothing — a programming error,
	// refused at construction like a flag Require was never given.
	if arm := Arm(c); arm != "" && !record.HasArm(verb, arm) {
		panic("seat.FillsArm: `" + c.CommandPath() + "` declares the arm `" + arm + "`, and a " + verb + " has no such arm")
	}
	for _, rf := range record.RequiredFields(verb) {
		key := rf.Key
		f := c.Flags().Lookup(rf.Flag)
		if f == nil {
			// The field is set by the verb rather than typed by the seat. Nothing to
			// annotate, and nothing wrong: not every payload key has a flag.
			continue
		}
		if rf.Arm != "" {
			// REQUIRED OF ONE ARM, so marked only on the verb that fills it — and marked, not
			// enforced at the parse: the write refuses an arm's missing field with what the field
			// is FOR, after the handler has established the motion's subject, which a parse-time
			// refusal would pre-empt.
			if Arm(c) == rf.Arm {
				mark(f)
			}
			continue
		}
		if c.Annotations[suppliesKey+key] != "" {
			// Required of the RECORD, supplied by the VERB. Marking it would tell a seat to
			// type what it does not have to, and the annotation would sit beside the word
			// "default" in the same sentence.
			continue
		}
		// AND COBRA ENFORCES IT. This only ever rewrote the usage string, so the help said
		// REQUIRED and the parser accepted the command without it — the tool asserting a
		// constraint nothing held. Measured: two of blue's free-text verbs, run with no flags,
		// recorded empty events and took a seat from two outstanding duties to complete:true.
		//
		// The record layer refuses these too, and that is not a second reader of one rule: it
		// guards a different boundary, the one internal callers reach through record.Append
		// without a command line. Cobra's is the SEAT's boundary, and it is the one that can
		// refuse before an event exists and can say so in the help.
		if alts := satisfiedByAnyOf[rf.Flag]; len(alts) > 0 {
			var present []string
			for _, a := range alts {
				if c.Flags().Lookup(a) != nil {
					present = append(present, a)
				}
			}
			if len(present) > 1 {
				// The marker without the per-flag enforcement: the alternative satisfies it,
				// and the usage says so in the same sentence.
				mark(f)
				c.MarkFlagsOneRequired(present...)
				continue
			}
		}
		Require(c, f.Name)
	}
}

// requiredMarker is the ONE spelling of the word a seat reads, at the head of a usage line.
const requiredMarker = "REQUIRED — "

// RequiredAnnotation holds on the flag what the marker says in its usage: present when the flag is
// marked, its one value the condition the marker states ("unless --accept"), "" for a bare marker.
// The usage is what a seat reads; this is what the tool and its gates read, so a marker reworded in
// prose cannot drop a flag from a gate, and a usage that says REQUIRED with no annotation is a
// writer that bypassed the two below, which requiredmarker_test refuses.
const RequiredAnnotation = "feov_required"

// IsMarked says whether a flag carries the marker, conditional or not. It is the one reader of the
// marker: the writers below skip a flag it accepts, and the gates that hold the help to the
// enforcement (requiredmarker_test, helpcontract_test) read it through this.
func IsMarked(f *pflag.Flag) bool {
	_, ok := f.Annotations[RequiredAnnotation]
	return ok
}

// MarkedCondition is the condition a marker states — "unless --accept" — or "" for an
// unconditional one; ok is false when the flag carries no marker at all. A conditional marker is
// a promise about ONE case, so the gate that invokes a verb without a marked flag and demands a
// refusal holds the unconditional markers and leaves the conditional ones to the probes, which
// drive the condition.
func MarkedCondition(f *pflag.Flag) (cond string, ok bool) {
	v, ok := f.Annotations[RequiredAnnotation]
	if !ok {
		return "", false
	}
	return v[0], true
}

// mark writes the marker ONCE. Every writer of the word goes through here, so a flag the schema
// requires and a verb also hand-marks renders one marker rather than "REQUIRED — REQUIRED — …",
// whichever writer ran first.
func mark(f *pflag.Flag) {
	if !IsMarked(f) {
		f.Usage = requiredMarker + f.Usage
		setMarked(f, "")
	}
}

func setMarked(f *pflag.Flag, cond string) {
	if f.Annotations == nil {
		f.Annotations = map[string][]string{}
	}
	f.Annotations[RequiredAnnotation] = []string{cond}
}

// RequiredUnless marks a flag the verb refuses when it is absent EXCEPT under cond — "unless
// --accept" — which the marker states between the word and the dash: "REQUIRED unless --accept — …".
// Marked, not enforced at the parse: the handler holds the condition, and the probes in
// requiredmarker_test drive it. A flag the verb never registered, or one already marked, is a
// programming error.
func RequiredUnless(c *cobra.Command, name, cond string) *cobra.Command {
	f := c.Flags().Lookup(name)
	if f == nil {
		panic("seat.RequiredUnless: `" + c.Name() + "` has no flag --" + name)
	}
	if IsMarked(f) || !strings.HasPrefix(cond, "unless ") {
		panic("seat.RequiredUnless: `" + c.Name() + " --" + name + "` is already marked, or its condition does not open with \"unless \"")
	}
	f.Usage = "REQUIRED " + cond + " — " + f.Usage
	setMarked(f, cond)
	return c
}

// Require is what a verb calls for a flag it cannot run without: the marker at the head of the
// usage AND cobra's refusal when it is omitted, in one act, so the two cannot come apart. A dozen
// verbs wrote them as a pair — `"REQUIRED — " + usage` beside a bare MarkFlagRequired — and the
// pairs drifted both ways: `reproduce --id` carried the word with no refusal, and other
// flags the refusal with no word. A flag the verb never registered is a programming error, and is
// refused at construction rather than skipped.
func Require(c *cobra.Command, names ...string) *cobra.Command {
	for _, n := range names {
		f := c.Flags().Lookup(n)
		if f == nil {
			panic("seat.Require: `" + c.Name() + "` has no flag --" + n)
		}
		mark(f)
		_ = c.MarkFlagRequired(n)
	}
	return c
}

// Begin runs a verb's preconditions and returns its seat context. It is a plain function
// called at the TOP of the one RunE — NOT a PreRunE hook — so there is no lifecycle chaining
// to reason about. It requires the run dir and seat id and holds the seat to its role.
func Begin(cmd *cobra.Command) (Context, error) { return begin(cmd, true) }

// beginRead is Begin for a verb that records nothing (NewRead). It refuses everything Begin refuses
// about WHO is asking and WHERE — a read answering from another seat's surface or another run's
// record looks exactly like the right answer — and holds no agent to register-first: a seat looks
// before it registers, and a read that registered it would spend the register its first record
// makes, leaving its own `register` refused as a repeat.
func beginRead(cmd *cobra.Command) (Context, error) { return begin(cmd, false) }

func begin(cmd *cobra.Command, records bool) (Context, error) {
	// ONE RESOLUTION, Of's, refused in Run's order: the run's disagreement first — before "--run
	// is required" and before the seat-id checks, because a seat pointed at the wrong run must not
	// get a message about anything else — then the run's existence, then the identity.
	s := Of(cmd)
	if s.RunErr == nil && s.runDir == "" {
		return s, feov.Errorf(feov.MissingField, "--run <runDir> is required")
	}
	// AND THE IDENTITY DISAGREEMENT, which Of resolves and carries as seatErr, and which Run
	// refuses on after the existence check. Measured before anything refused it: with
	// FEOV_SEAT=blue-respond injected, a call passing --seat-id blue-respond was ACCEPTED and
	// filed under r9 — a seat no dispatch ever created, carrying its own register event and its
	// own shard. Attribution is the one fact a seat must not be able to get wrong; every found_by,
	// estoppel and parity check reads it, and this is the guarantee #348 shipped a message for and
	// no code behind.
	if _, err := s.Run(); err != nil {
		return s, err
	}
	// A lookup that could not answer is no refusal for a read, and is one here: a write must be
	// attributed to the registration, and nothing could read one.
	if s.bindErr != nil {
		return s, s.bindErr
	}
	if s.SeatID == "" {
		return s, feov.Errorf(feov.MissingField, "--seat-id is required (state it once at `register`; after that it is bound and read back for you)")
	}
	if err := requireOpenRecord(cmd, s); err != nil {
		return s, err
	}
	if records {
		if err := requireBound(cmd, s); err != nil {
			return s, err
		}
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
	if err := CheckFlagReferences(cmd, s.runDir); err != nil {
		return s, err
	}
	return s, nil
}

// referenceChecker is what a typed flag implements when it knows what it is checked against.
// Declared here rather than imported so the flags package keeps no dependency on this one.
type referenceChecker interface {
	Check(runDir, flag string) error
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
			err = c.Check(runDir, "--"+f.Name)
		}
	})
	return err
}

// taught marks an error the HANDLER already answered — diagnosed in the verb's own words, with
// whatever the seat needs next already in the sentence. The words are the handler's and nothing is
// added to them: the role and the command path are the renderer's, read from the command.
//
// It exists so the top level can tell those apart from the ones cobra raises BEFORE the handler
// runs: an unknown flag, a missing required flag, a bad value, a flag group violated. Those say
// only which flag word was wrong, and the command tree holds the answer — so the top level gives
// it to them, and to nothing else: a few lines for a flag left out or not the verb's, the
// command's own help for the rest.
//
// A FIELD, NOT A PATTERN. The alternative was matching cobra's message text, which is a string
// shape from another module: it fails silently the day cobra rewords it, and a refusal that
// quietly stops teaching reads exactly like one that never needed to (see [[facts-are-fields]]).
type taught struct{ error }

func (t taught) Unwrap() error { return t.error }

// Taught reports whether this error already answers itself: a verb's handler diagnosed it, or it
// is a Refusal carrying its own page.
func Taught(err error) bool {
	var t taught
	var r *Refusal
	return errors.As(err, &t) || errors.As(err, &r)
}

// Emit renders a verb's SUCCESS, and hands a failure back marked as the handler's own answer.
//
// A refusal is not written here. Every refusal leaves the tool through RenderRefusal, which reads
// the role and the verb from the command and the code from the error, so a handler's refusal and
// one cobra raised before the handler ran are the same shape by construction. The error is
// returned unchanged in its words: the role and the path are the renderer's to print.
func Emit(cmd *cobra.Command, res Result, err error) error {
	role := roleOf(cmd)
	if err != nil {
		return taught{err}
	}
	// WHERE THE SEAT NOW STANDS RIDES THE ACT. Computed once, here, for both forms — a seat reading
	// the human line and a consumer parsing the envelope are answering the same question and must not
	// be told different things.
	if r, ok := res.(Relayed); ok && jsonMode(cmd) {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(r.Relay())
	}
	standing := standingAfter(cmd)
	if jsonMode(cmd) {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(okEnvelope{
			Verb: cmd.Name(), Role: role, OK: true, Result: res, Standing: standing,
		})
	}
	if res != nil {
		fmt.Fprintln(cmd.OutOrStdout(), res.Human())
	}
	// AFTER the verb's own line, never instead of it: the act's own result is what the seat asked
	// for, and the standing is the context it would otherwise have spent a call on.
	if standing != nil {
		fmt.Fprintln(cmd.OutOrStdout(), standing.Human())
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
// same command means something different on each side. `position` is a RED section for the chair
// and a BLUE one for blue, so those are two documents; `register` and `log` mean exactly the same
// thing everywhere and are one.
//
// THEY WERE FOUR COPIES. Each role passed its own literal, under a comment saying the guidance
// differs by role — and register's and friction's were byte-identical at every call site but one,
// where the difference was the word "seat" against "sitting". Four hand-kept copies of one
// paragraph, which is the shape that goes stale in three places and nobody notices.
func NewKeyed(name, key string, run Handler) *cobra.Command {
	c := newVerb(name, key, "")
	annotate(c, recordsKey, name)
	c.RunE = HandlerRunE(run)
	return c
}

// NewRead is New for a verb that only reads: its page also states the --json success it emits,
// generated from the verb's result type, because a seat that does not know the key names guesses
// them — and a guessed key that misses reads as output that was never JSON.
//
// A READ RECORDS NOTHING, SO IT CARRIES NO RECORD TYPE AND REGISTERS NO ONE. Only an act that
// records registers a seat, as register's page says. A read that went through Begin registered
// its seat as a side effect, and the seat's own register later in the sitting was then refused as
// a repeat — universe-m33's voice lens met exactly that after a read.
func NewRead(name string, result Result, run Handler) *cobra.Command {
	c := newVerb(name, name, readOutput(result))
	c.RunE = readRunE(run)
	return c
}

func newVerb(name, key, output string) *cobra.Command {
	d := helpFor(key)
	c := &cobra.Command{
		Use:          name,
		Short:        d.menu,
		Long:         d.menu + "\n\n" + d.detail + output + "\n" + LogFooter,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // a validation refusal is a teaching message, not a usage dump
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
	// substance where nothing could read it. A missing field is now a log entry, filed as a
	// request, rather than text dropped into an unqueryable channel.

	return c
}

// HandlerRunE is the RunE every writing verb holds: Begin (identity, register-first, the dispatched seat,
// flag references) -> the work -> Emit. New wires it; a verb built as a raw cobra command — the
// motion verbs, whose help is composed per subject — sets it as its RunE, so its preconditions and
// its refusal's envelope are the same as every other verb's.
func HandlerRunE(run Handler) func(*cobra.Command, []string) error { return runE(Begin, run) }

// readRunE is HandlerRunE for a verb that records nothing: beginRead in place of Begin.
func readRunE(run Handler) func(*cobra.Command, []string) error { return runE(beginRead, run) }

func runE(begin func(*cobra.Command) (Context, error), run Handler) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		s, err := begin(cmd)
		if err != nil {
			return Emit(cmd, nil, err)
		}
		res, werr := run(s, cmd)
		return Emit(cmd, res, werr)
	}
}

// Prose adds the shared prose payload channel — --reason, and the quoting rule in the verb's help —
// to a verb that reads one. It routes through flags.RegisterPayload rather than registering the
// flag by hand, so a verb cannot take prose without the rule that keeps the shell out of it;
// `close` once shipped its own private prose flag before the channel was centralized here.
func Prose(c *cobra.Command) *cobra.Command {
	flags.RegisterPayload(c)
	if u, ok := ReasonIs[c.Name()]; ok {
		c.Flags().Lookup(flags.Reason).Usage = u
	}
	return c
}

// ReasonIs is what --reason IS on the verbs whose prose is the artifact itself rather than an
// argument about an act. flags.DescReason asks "why you graded, closed, ruled or edited", which
// misdescribes a fate's account, a receipt, a log entry or a holding. Keyed by verb name because
// Prose is where the flag is registered; TestReasonIsNamesOnlyProseVerbs refuses a key no prose
// verb carries, so a renamed verb cannot silently fall back to the generic text.
var ReasonIs = map[string]string{
	"record":       "the account of what settled this fate, in the SUBJECT's terms — printed in the report on the inquiry's row, beside its question",
	"review":       "your one read of the report: what it SAYS where it accounts for the inquiries this run investigated — quote it — and where its body falls short of what an inquiry claims; one text covers every inquiry",
	"finding":      "the finding itself — what is wrong and why, in one text: this flag carries it and no other flag states the problem. It is what the findings view prints and what the repair is re-audited against. " + flags.ReasonNotProcess,
	"cite":         "why this source backs the sentence — kept on the record and shown beside the source in the evidence view, never printed in the report",
	"prove":        "the proof note — printed in the report as this computation's footnote, so say what it SHOWS about the subject, not what happened while you ran it",
	"manifest-row": "the receipt: what you checked, and what checking it showed",
	"log":          "the entry: what you concluded about the tooling",
	"position":     "your sitting's argument — your THINKING, not your process. " + flags.ReasonNotProcess,
	"mint":         "your argument for the gap — why it is worth one, the case blue answers and a bench weighs; the report never prints it. Under unclassified it is REQUIRED: name the nearest class and what it does not cover. " + flags.ReasonNotProcess,
	"closing":      "your closing argument on this gap — your THINKING, not your process. " + flags.ReasonNotProcess,
	"spot-check":   "what reading the stale areas and sampling the closure archive found, a defect included — printed with the spot-check in the report's board and in the debate transcript. " + flags.ReasonNotProcess,
	"certify":      "what a human should re-examine after the run, and why. " + flags.ReasonNotProcess,
	"declare":      "the holding: how the whole record is to be READ. " + flags.ReasonNotProcess,
	"halt":         "the opinion: the boundary, and why continuing would itself be the harm. " + flags.ReasonNotProcess,
}

// ProseRequired registers the prose channel AND says so, for a verb whose argument is
// unconditional even though the FIELD's is not.
//
// `lens close` is the case. `Close.prose` cannot carry `required: true` — that refuses before
// validate runs and so refused a CARRY, which restates an argument an earlier epoch already made.
// Dropping the annotation fixed the carry and silently unmarked close: cobra stopped refusing and
// the help stopped saying REQUIRED, on a verb whose argument is the whole point. Marking it by
// hand at the verb restored the refusal and NOT the marker, which is worse than either — a
// requirement the parser holds and the help does not state is one a seat only meets by accident.
//
// Both halves here, so they cannot come apart again.
func ProseRequired(c *cobra.Command) *cobra.Command {
	return Require(Prose(c), flags.Reason)
}

// SaysRequired writes the REQUIRED marker on flags whose refusal DELIBERATELY lives in the verb's
// own handler ("requires --reason: …" in the handler's words for what the prose is FOR, said after
// whatever the handler establishes first) or in validate's field checks. It changes no
// enforcement; a flag whose refusal may be the parser's takes Require instead, which writes both
// halves. A field an ARM of the body requires is not written here: markRequired marks it from the
// schema on the verb that declares the arm (FillsArm).
//
// Measured before it existed: twenty-two flags across every seat surface refused when omitted
// and rendered with no marker, and the m13/m14 seats paid one call each to learn what the help
// could have told them (`position --reason`, `motion docket rule --reason`). markRequired writes the word from the schema and ProseRequired from the verb; this
// is the writer for the requirements the other two mechanisms cannot see, and
// TestHelpSaysRequiredExactlyWhereOmissionIsRefused holds every call here to the refusal it claims.
func SaysRequired(c *cobra.Command, names ...string) *cobra.Command {
	for _, n := range names {
		f := c.Flags().Lookup(n)
		if f == nil {
			panic("seat.SaysRequired: `" + c.Name() + "` has no flag --" + n)
		}
		mark(f)
	}
	return c
}

// Reason resolves that channel through flags.ReadPayload — the ONE resolver, which refuses a read
// of a channel the verb never registered and trims the newline a captured heredoc can leave.
//
// Prose into markdown costs nothing; prose through the tool meant fighting the shell. The 2026-07-18
// run measured 68 commands carrying escaped quotes and 37 staging a temp file, and the answer then
// was file and stdin spellings of this flag. The fight that mattered was substitution, not quoting,
// and the answer to it is one rule for every free-text flag — see flags.Prose.
func Reason(cmd *cobra.Command) (string, error) {
	return flags.ReadPayload(cmd)
}

// Str reads a string flag.
func Str(cmd *cobra.Command, name string) string {
	// ONE IMPLEMENTATION, in flags.Value.
	return flags.Value(cmd, name)
}

// OptStr is Str for a field whose ABSENCE is a different fact from its emptiness.
//
// `proto.String(Str(cmd, f))` sets the field whatever happens, so a flag the seat never passed
// lands in the record as the empty string — "the seat said nothing" written as "the seat said
// nothing at all". Every field on this schema is `optional` precisely to keep those apart, and the
// write path was throwing the distinction away on the way in.
//
// MEASURED, and the foreign keys are what found it: `lens close` set `successor` unconditionally,
// so an ordinary closure stored `successor = ”` — and once successor referenced `mint.gap_id`,
// every close in the tool failed with `FOREIGN KEY constraint failed`, because no gap is named ”.
// Before the constraint existed the same rows were written and simply read as a closure whose
// successor was the empty gap.
//
// It keys on Given (pflag's Changed), not on emptiness: a seat that passes `--reason ""` has said
// something — that there is nothing to say — and that is a fact the record keeps.
func OptStr(cmd *cobra.Command, name string) *string {
	if !Given(cmd, name) {
		return nil
	}
	v := Str(cmd, name)
	return &v
}

// Given answers whether the SEAT passed the flag. The absent/present distinction
// is load-bearing for the record format — a flag never passed must not appear in
// the event at all — and pflag's Changed is exactly that question.
func Given(cmd *cobra.Command, name string) bool {
	f := cmd.Flags().Lookup(name)
	return f != nil && f.Changed
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
	// `help` is cobra's, added lazily at execute time; without this it lands under an
	// "Additional Commands:" heading of its own, which reads as a third category of seat surface
	// rather than as the plumbing it is.
	root.SetHelpCommandGroupID(GroupVerbs)
}

// requireOpenRecord asks the record whether it takes this verb's act at all — the run was captured, the
// record holds an outcome, or an outcome is stamped outside the assembly sitting
// (record.RequireSeatMayWrite) — BEFORE the verb's own checks run, and before register-first: a verb
// that refuses on a check of its own first tells a seat after the outcome what else is wrong with
// an act the run no longer takes (`lens close` on a computation gap with no proof was told to wait
// for blue's proof). A verb that writes no event type has nothing to ask; the write asks again
// under its lock.
//
// A register states its occasion here, as its flag: the bench's register for the assemble occasion
// is the one register an outcome leaves open, so a crashed assembly sitting can resume.
func requireOpenRecord(cmd *cobra.Command, s Context) error {
	if cmd == nil {
		return nil
	}
	typ, ok := recordpb.EventTypeOf(RecordType(cmd))
	if !ok {
		return nil
	}
	var body proto.Message
	if typ == recordpb.EventType_EVENT_TYPE_REGISTER {
		reg := &recordpb.Register{}
		if f := cmd.Flags().Lookup(flags.Occasion); f != nil {
			if occ, known := record.OccasionOf(f.Value.String()); known {
				reg.Occasion = occ.Enum()
			}
		}
		body = reg
	}
	return record.RequireSeatMayWrite(s.handle(), s.SeatID, typ, body)
}
