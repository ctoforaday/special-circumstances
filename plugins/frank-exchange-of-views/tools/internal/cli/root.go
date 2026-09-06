// Package cli assembles feov-record: the root command, the flags every seat
// shares, and the four role trees.
//
// The seat-side record runtime. ONE binary whose role subcommands carry bespoke
// verb sets, because the verb set IS the role boundary — a lens has no mint verb
// to call, blue has no board verbs at all, and the bench rules without
// originating. Seat identity is bound to its namespace, so the boundary is
// enforced rather than merely described.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/bench"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/merge"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/motion"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// InvokedAs is the name this binary was actually run under, and every place the tool names
// ITSELF uses it: the usage line, the refusals, the "did you mean" lists.
//
// It was the constant "feov-record", which is a claim about the filesystem the tool is in no
// position to make. The seat probe caught it — the harness builds the binary as fxr.exe, a seat
// typed a command wrong, and the refusal came back speaking about a `feov-record` that does not
// exist anywhere on that machine. A seat has exactly one way to learn the surface (the refusal)
// and the refusal was naming a different program.
//
// THE BASENAME, NOT THE FULL PATH, and the choice is measured rather than assumed. The full path
// would make every example directly runnable, which is tempting — but the golden help contracts
// are captured from a binary built into a temp directory, so a full path would make them
// machine-dependent, and across nine probed seats not one ever typed a bare name copied out of
// help: they all use the absolute path their prompt gives them. The copy-paste win is worth
// nothing here and the determinism is worth a great deal.
func InvokedAs() string {
	n := filepath.Base(os.Args[0])
	if ext := filepath.Ext(n); strings.EqualFold(ext, ".exe") {
		n = n[:len(n)-len(ext)]
	}
	if n == "" || n == "." || n == string(filepath.Separator) {
		return "feov-record" // argv[0] can be empty; a name is better than none
	}
	return n
}

// seatRoles maps a resolved role to its surface and its one-line description. Built here rather
// than in each package so the root can answer "which seat am I" once.
func seatVerbs(role string) ([]*cobra.Command, string) {
	switch role {
	case "lens":
		return lens.Verbs(), "red lens seats — findings, source verification and corroboration, proof re-runs. Cannot mint or close a gap: that is the merge's."
	case "merge":
		return merge.Verbs(), "the red merge seat — the board's only writer."
	case "blue":
		return blue.Verbs(), "blue seats — revisions, manifest rows, directions. No board verbs at all."
	case "bench":
		return bench.Verbs(), "the bench — opinions, halt, certification. Never originates."
	}
	return nil, ""
}

// RoleOfSeat resolves the dispatched role from a seat id, or "" for no seat.
func RoleOfSeat(seatID string) string {
	if strings.TrimSpace(seatID) == "" {
		return ""
	}
	for _, r := range []string{"lens", "merge", "blue", "bench", record.OperatorRole} {
		if record.CheckSeatRole(r, seatID) == nil {
			return r
		}
	}
	return ""
}

// noSeatNote explains an OPERATOR surface to something that expected a seat's.
//
// THIS IS THE COST OF SCOPING THE TREE, PAID RATHER THAN HIDDEN. A seat verb is now absent from a
// process with no identity, so "mint" answers "no such command" where it used to answer "--seat-id
// is required". That is the failure requireRuler's own comment named — under a scoped surface "not
// yours" reads as "does not exist", and a seat handed an unavailable verb logs friction and works
// around it, losing the capability for the run. The surface cannot carry the verb (it does not know
// which seat's), so the REFUSAL has to carry the reason.
func noSeatNote(seatID string) string {
	if strings.TrimSpace(seatID) == "" {
		return "\n\n--seat-id IS REQUIRED HERE, because nothing identified you: this agent has not " +
			"registered, so the record cannot say which seat it holds, and no --seat-id was given. The " +
			"surface is scoped to whoever is asking, so there is no tree to show until you say who that " +
			"is — a seat id, or `" + record.OperatorRole + "`. You type it ONCE, at `register`; after that " +
			"the binding is on the record and every later call resolves it for you."
	}
	if RoleOfSeat(seatID) == "" {
		return "\n\n" + seatID + " does not belong to any role namespace, so it selects no surface. The engine " +
			"assigns seat ids; a hand-invented one identifies nobody, and a record written under it would be " +
			"attributed to an identity no dispatch created."
	}
	return ""
}

// newRoot builds the tree for whoever is running it — see NewRootFor.
func newRoot() *cobra.Command {
	return NewRootFor(dispatchedSeat())
}

// dispatchedSeat answers who is running this process, BEFORE cobra parses anything.
//
// The run directory is resolved the way the seat's own verbs resolve it — injected FEOV_RUN
// first, then the live-run marker — because that is what the hook actually supplies, and a
// pre-parse scan of os.Args for --run would be a second flag parser guessing at the same fact.
// Where no run resolves there is no record to ask, so the answer is the flag or nothing.
func dispatchedSeat() string {
	runDir, err := seatenv.Resolve("", func() string { return runlive.InferRunDir("") })
	if err != nil {
		return seatenv.Dispatched(nil)
	}
	// NewRun, matching seat.Of: this asks whether an agent has registered, and must answer for a
	// run that is not yet on disk rather than refusing before the question is put.
	run, rerr := record.NewRun(runDir)
	if rerr != nil {
		return seatenv.Dispatched(nil)
	}
	return seatenv.Dispatched(seat.BoundSeat(run))
}

// NewRootFor builds THIS SEAT'S surface at the root, or the operator's when no seat was dispatched.
//
// THE ROLE LEVEL IS GONE. It used to be four groups under the root, so a merge seat typed
// `feov-record merge mint` — restating a fact the tool could already look up, with CheckSeatRole
// reconciling the two copies and refusing a disagreement. The tool's own refusal message said the
// principle it was breaking: you do not retype your identity. A merge seat now runs `mint`; a lens
// seat running `mint` gets an unknown command, which is the same boundary drawn by the only fact
// that decides it.
//
// The operator tree is the one with no seat: setup, capture, dashboard and the rest have no role
// to scope to, and a human running them is not a seat.
func NewRootFor(seatID string) *cobra.Command {
	root := &cobra.Command{
		Use:   InvokedAs(),
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: InvokedAs() + ` — the seat-side record runtime (the verb set IS the role boundary)

A lens structurally cannot mint or close a board gap: no such verb exists in its
namespace. Blue has no board verbs at all. The bench rules and never originates.`,
		Version:       buildid.Revision(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// The two flags EVERY verb needs, declared once and inherited. Persistent
	// flags are the mechanism the first cut of this CLI re-implemented by
	// re-declaring --run and --seat-id on all sixteen verbs.
	// --schema PRINTS ONE INTEGER AND EXITS. setup compares it against the plugin's
	// requirements.json, and giving it a flag of its own means nothing has to recover a number
	// from --version's prose — the string-shaped hope this repository keeps deleting.
	root.PersistentFlags().Bool(flags.Schema, false, "print the event-schema epoch this binary writes, and exit")

	root.PersistentFlags().String(flags.Run, "", "the run directory. The PreToolUse hook injects it in a real run, which is why you rarely type it. A value that DISAGREES with the run you were dispatched into is refused rather than obeyed")
	root.PersistentFlags().String(flags.SeatID, "", "your seat id, as the dispatch prompt states it (SEAT_ID). You pass it ONCE, at `register`, which binds it to you on the record; after that every call resolves it for you and typing it is optional. It SELECTS this surface — the verbs listed are the ones your seat may run — and a value disagreeing with what you registered as is refused rather than obeyed")
	// --json makes every mutating verb emit a structured result and every failure a
	// structured error, so a machine consumer parses fields instead of prose. On READS the
	// format is primarily view-selected: board/findings/work/motions/telemetry/evidence are JSON
	// by name (seat.JSONByNameViews is the set; each one says so in its own --help), the rest
	// are markdown. --json opts a markdown view into its structured form where one exists
	// (today only `show debate --json`); it is an error on a JSON-by-name view or a
	// markdown view with no JSON form, so there is exactly one way to reach each form.
	root.PersistentFlags().Bool(flags.JSON, false, "emit a structured JSON result (and structured errors) instead of human text")

	// A SEAT IS NOT AN OPERATOR, and the two sets are disjoint by construction rather than by
	// anyone remembering. Promoting the seat's verbs to the root put them beside the operator
	// commands, and two names collide outright: `friction` is a seat's write verb AND the
	// operator's read of the channel; `verify` is a lens's citation adjudication AND the
	// operator's whole-record cross-check. One tree cannot hold both meanings, and picking a
	// winner would hand some seat a verb that does the other thing.
	//
	// So the split is total. `fetch` and `count-claims` cross it because seats genuinely run them
	// — a lens reads blue's cached source bytes, and blue's claim_count is defined as what
	// count-claims prints — and neither name collides.
	// NO IDENTITY IS NOT A MODE. It used to fall through to the operator surface, which made
	// "nobody said who I am" a way of selecting a tree — and a tree nobody selects is one nobody
	// can be refused from. The surface is scoped to whoever is asking, so there is nothing to show
	// until that is answered; `--seat-id operator` is how an operator answers it.
	switch role := RoleOfSeat(seatID); {
	case role == "":
		// NO VERBS FOR AN IDENTITY NOBODY RECOGNISES, and that covers two cases with one rule:
		// nothing was given, or what was given belongs to no namespace. An invented seat id used
		// to fall through to the OPERATOR surface, which is a claim about the caller that nothing
		// checked — it is not the operator, it is unidentified. noSeatNote says which of the two
		// it met. refuseUnknownCommandFirst answers the identity question before cobra parses
		// anything, so `mint --check c` is not answered with "unknown flag: --check" — a complaint
		// about the third word of a command that could not have run. `--version` and `--help` are
		// left to cobra: they are answerable without knowing who is asking, and the setup
		// preflight runs `--version` against a binary it has not identified itself to.
	case role != "" && role != record.OperatorRole:
		verbs, short := seatVerbs(role)
		seat.SetRole(root, role, seatID)
		root.Short = short
		// THE FRICTION FOOTER MOVED WITH THE VERBS, not away with the group. It hung on the role
		// group's Long, and deleting the group would have dropped it silently — a seat would stop
		// being told that a capability it cannot find is a finding about the tooling rather than
		// something to work around. Caught by TestRoleHelpCarriesTheFrictionFooter.
		root.Long = InvokedAs() + " — " + short + "\n" + seat.FrictionFooter
		root.AddCommand(verbs...)
		root.AddCommand(motion.NewCommandFor(role))
		root.AddCommand(newFetch())       // a lens reads the EXACT bytes blue read, from the run cache
		root.AddCommand(newCountClaims()) // blue's claim_count is defined as what this prints
		// AFTER every AddCommand: the split reads HasSubCommands, so a group registered before
		// its children were attached would file itself under the leaves.
		seat.SplitGroups(root)
	default:
		root.AddCommand(
			newVerify(),          // operator cross-check, not a seat role — read-only over the record
			newGraph(),           // operator: render a run's actual behaviour from the record
			newShowDiagnostics(), // operator: did the seats find their surface — exposure and use, as fields
			newCountClaims(),     // operator/blue: deterministic claim_count over blue/report.md
			newLog(),             // operator: the friction channel — seats write it, the human reads it
			newFetch(),           // operator: cached, hash-verified web read — feeds the run source cache
			newOCR(),             // operator: render a scan's pages so a seat can read what has no text layer
			newSetup(),           // operator: build a research run's blackboard
			newScorecard(),       // operator: a chair's in-run self-read scorecard
			newDashboard(),       // operator: the live run dashboard.html
			newCapture(),         // operator: the post-hoc capture auditor
		)
	}

	// THE HELP TEMPLATE CARRIES WHAT COBRA HAS NO MODEL FOR: an enum's per-value meanings.
	// Cobra's help is rich about commands and flags; a flag's VALUE-SPACE is neither, so every
	// vocabulary here had nowhere to put its semantics and they went into source comments.
	// AN UNKNOWN COMMAND PRINTS THE SURFACE. Cobra's default answers `unknown command "x" for
	// "feov-record"` and stops — no list, no pointer, nothing. That is the one moment a seat is
	// definitively looking for what exists, and it was the one moment the tool said least.
	//
	// Taking ArbitraryArgs routes the miss into RunE instead of cobra's built-in message, so the
	// help goes out first and the refusal follows it.
	root.Args = cobra.ArbitraryArgs
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return seat.RefuseAndTeach(cmd, "you named no command, so nothing ran. The commands below are the whole surface."+noSeatNote(seatID))
		}
		return seat.RefuseAndTeach(cmd, unknownCommandRefusal(cmd.Root(), args[0])+noSeatNote(seatID))
	}

	enumhelp.Install(root)
	return root
}

// whereItLives finds the full paths a bare verb name is reachable at, so a refusal can say WHERE
// rather than only NO.
//
// MEASURED 2026-08-17, and it is the sharpest evidence this file has for the slip it already
// documents. A red-merge seat, holding a work list duty that named `inquiry-support`, typed
//
//	feov-record inquiry-support --help
//
// and was told `no command named "inquiry-support" exists`. It believed that — reasonably, the
// message is unambiguous — went hunting, and settled on `motion inquiry rule`, which is a DIFFERENT
// ACT: whether the direction is worth the run's time, not whether the report still carries it. The
// seat did not lose a turn; it landed on the wrong verb with confidence.
//
// The refusal was false in the way that matters. `inquiry-support` exists, on the merge seat, one
// word away. The role-level refusal already says this well — "it exists, on the blue seat, but not
// for you. That is a wrong-seat error rather than a missing capability" — and the ROOT level, which
// is where a seat dropping the prefix actually lands, said the opposite.
//
// It searches one level under each role group, which is where seat verbs live. `motion <subject>
// <verb>` is deliberately not searched: `rule` and `file` are generic enough that pointing at all
// three subjects would be noise rather than an answer.
func whereItLives(_ *cobra.Command, name string) []string {
	var out []string
	for role, r := range AllRoots() {
		for _, c := range r.Commands() {
			if c.Name() == name || c.HasAlias(name) {
				out = append(out, role)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

func unknownCommandRefusal(root *cobra.Command, name string) string {
	// A VERB THAT EXISTS SOMEWHERE ELSE IS NOT A MISSING CAPABILITY, and the difference decides
	// what the seat does next. It used to end "run it with its role in front", which was the fix
	// while the tree had a role level; that advice is now wrong, and wrong advice at a refusal is
	// worse than none — it sends a seat to type something that cannot work.
	if at := whereItLives(root, name); len(at) > 0 {
		return fmt.Sprintf("%q is not on your surface — it is the %s seat's verb. That is a "+
			"wrong-SEAT error, not a missing capability, so do not work around it: if the act is "+
			"yours to perform, you are dispatched as the wrong seat, and if it is not, the seat that "+
			"holds it is named here. The commands below are everything you can run.",
			name, strings.Join(at, " and "))
	}
	return fmt.Sprintf(
		"no command named %q exists. The commands below are the whole surface, and each one's own `--help` carries its verbs.", name)
}

// refuseUnknownCommandFirst answers the WRONG COMMAND before cobra can answer the wrong flag.
//
// Cobra parses flags before it decides a command is unknown, so `show --view board` reported
// `unknown flag: --view` — naming the one thing the caller had right. The flag is fine; it just
// belongs to `blue show`, `merge show`, `lens show` or `bench show`. A seat reading that goes
// hunting through view names and never meets the refusal that would have taught it the role
// prefix, which inverts the whole point of a teaching refusal.
//
// MEASURED, and it is the single most common failure a seat hits: across nine probed seats,
// 9 of 21 non-zero exits were this message, and SIX OF NINE SEATS produced it on their FIRST
// tool call — the moment a seat is orienting and least able to spare a wasted turn. `show` is
// per-role while the concept ("show the board") carries no role, so dropping the prefix is the
// natural slip rather than a careless one.
//
// IT INSPECTS os.Args[1] AND NOTHING ELSE. Scanning for the first non-flag token would misread
// a flag's VALUE as a command name (`--model haiku ...` would nominate "haiku"), and matching on
// cobra's error text would make this depend on the shape of a string another library owns —
// which is the failure mode this repo keeps finding. Position 1 cannot be a flag value, because
// nothing precedes it.
// seatID is passed rather than read from the environment: the tests drive cobra in-process with
// an args slice that never reaches os.Args, so a check reading the env would refuse every call.
func refuseUnknownCommandFirst(root *cobra.Command, argv []string, seatID string) error {
	if len(argv) < 2 {
		return nil
	}
	// COBRA BUILDS PART OF ITS OWN SURFACE DURING Execute, and refusing before that showed the
	// seat a SMALLER surface than the one that exists: `help`, `completion`, `--help` and
	// `--version` were all missing from the list. The golden caught it — a refusal that lists
	// less than the tool has is a worse lie than the flag error this replaces, because it reads
	// as authoritative.
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	root.InitDefaultHelpFlag()
	root.InitDefaultVersionFlag()

	// THE COMMAND IS THE FIRST NON-FLAG WORD, not argv[1]. `--json mint …` put a flag there, so
	// the pre-check saw one and stood aside, and cobra then complained about `--class` — a flag on
	// a command it had not found. The root's own flags are a knowable set: two take a value, the
	// rest are booleans.
	name := ""
	for i := 1; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--run" || a == "--seat-id":
			i++ // and its value
		case strings.HasPrefix(a, "-"):
			// --json, --help, --version, and any --flag=value form: no separate value to skip.
		default:
			name = a
		}
		if name != "" {
			break
		}
	}
	if name == "" {
		return nil // flags only: cobra's own handling is right for --help and --version
	}
	// NOBODY SAID WHO IS ASKING, so there is no surface to look this up in. Answered here rather
	// than after parsing, because the flags belong to a command that could not have been found.
	if RoleOfSeat(seatID) == "" {
		return seat.RefuseAndTeach(root, "a command was named and the surface is scoped to whoever is asking."+noSeatNote(seatID))
	}
	for _, c := range root.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return nil
		}
	}
	// A WRONG-SEAT ERROR KEEPS ITS CODE. `--json` consumers branch on the KIND of failure, and
	// this one is a role violation however the tree came to lack the verb; flattening it to a
	// plain message makes "you are the wrong seat" indistinguishable from a typo.
	taught := seat.RefuseAndTeach(root, unknownCommandRefusal(root, name)+noSeatNote(seatID))
	if len(whereItLives(root, name)) > 0 {
		return feov.Errorf(feov.RoleViolation, "%s", taught.Error())
	}
	return taught
}

// Execute runs the CLI. Abort-safety is armed first: a seat killed mid-command
// must lose an event, never leave a torn one or a stuck lock.
func Execute() {
	defer record.InstallSignalGuard()()

	root := newRoot()
	if err := refuseUnknownCommandFirst(root, os.Args, dispatchedSeat()); err != nil {
		// THROUGH THE SAME EMITTER AS EVERY OTHER TOP-LEVEL REFUSAL. This branch printed a bare
		// sentence, so `--json` callers got prose from the one path that answers before cobra —
		// a channel whose entire contract is that it is machine-readable.
		if !EmitTopLevelError(os.Stdout, os.Args, err) {
			fmt.Fprintf(os.Stderr, "%s: %v\n", InvokedAs(), err)
		}
		os.Exit(2)
	}
	if err := ExecuteRoot(root); err != nil {
		// A --json CALLER GETS JSON, INCLUDING WHEN THE FLAGS THEMSELVES ARE REFUSED.
		//
		// `seat.Emit` renders every refusal a HANDLER produces as a structured envelope. A
		// refusal from flag PARSING never reaches it: pflag rejects the value and cobra returns
		// the error straight out of Execute, so `--json merge mint --severity banana` printed a
		// bare sentence on a channel whose entire contract is that it is machine-readable. A
		// consumer parsing that channel gets a JSON error and cannot see which flag was wrong.
		//
		// It has been true since `GradeValue` shipped and nobody noticed, because the seats read
		// the human line and the engine only parses the SUCCESS path. It surfaced when the id
		// flags became typed: a case that used to fail inside the handler began failing at parse,
		// and the same input changed channel shape without changing meaning.
		//
		// So the envelope is rendered here too. The verb and role are unknown at this point —
		// parsing failed before cobra resolved which command owns the flag — and the envelope
		// says so with empty fields rather than guessing a name from argv.
		if EmitTopLevelError(os.Stdout, os.Args, err) {
			os.Exit(2)
		}
		// Errors print bare, without cobra's usage dump: a validation refusal here
		// is a TEACHING message a seat reads and acts on, and burying it under a
		// flag listing is how it stops being read.
		fmt.Fprintf(os.Stderr, "%s: %v\n", InvokedAs(), err)
		os.Exit(2)
	}
}

// ExecuteRoot runs the tree and attaches the failing verb's own help to the refusals COBRA
// raises — an unknown flag, a missing required flag, a flag group violated.
//
// Those refusals name a flag word and stop: "required flag(s) \"verified-by\" not set" tells a
// seat which word it missed and nothing about what the word wants or what else the verb takes.
// The flag table is the answer, and cobra already renders it — so the diagnosis leads and cobra's
// own help follows, which is the shape every other refusal in this tool has.
//
// It is scoped to SEAT VERBS and to errors the handler did not already answer. A handler's
// refusal is a teaching message already (seat.Taught marks them), and an operator command's
// refusals were never flag-shaped.
//
// Exported for the same reason EmitTopLevelError is: the test harness drives the tree directly,
// and a fix living only in Execute() would be invisible to every test.
func ExecuteRoot(root *cobra.Command) error {
	// BEFORE COBRA DISPATCHES ANYTHING. --schema is not a verb and must answer even when the
	// argv would otherwise be refused: setup runs it against a binary it does not yet trust,
	// and a preflight that cannot get an answer out of a working binary is worse than no
	// preflight. One integer on stdout, nothing else, so the caller parses a number and not a
	// sentence.
	for _, a := range os.Args[1:] {
		if a == "--"+flags.Schema {
			fmt.Fprintln(root.OutOrStdout(), record.EventSchema)
			return nil
		}
	}
	cmd, err := root.ExecuteC()
	if err == nil || cmd == nil || seat.Taught(err) || seat.RecordType(cmd) == "" {
		return err
	}
	return seat.RefuseAndTeach(cmd, err.Error())
}

// EmitTopLevelError renders an error that never reached seat.Emit, and reports whether it did.
//
// Exported because the TEST HARNESS must reproduce the binary here. It drives root.Execute()
// directly and returns the error, so a fix living only inside Execute() would be invisible to
// every test — the harness would go on measuring a shape the binary does not produce, which is
// the same defect one layer up from the one this fixes.
func EmitTopLevelError(w io.Writer, argv []string, err error) bool {
	if !jsonRequested(argv) {
		return false
	}
	// THE CODE IS THE ERROR'S OWN WHERE IT HAS ONE. Validation was hardcoded because everything
	// reaching here was a parse refusal; a wrong-SEAT refusal now arrives here too, and it carries
	// role_violation. A consumer branching on the kind of failure would otherwise be told a
	// dispatch error is a typo.
	code := feov.Validation
	var coded *feov.Error
	if errors.As(err, &coded) && coded.Code != "" {
		code = coded.Code
	}
	_ = json.NewEncoder(w).Encode(struct {
		OK    bool   `json:"ok"`
		Code  string `json:"code"`
		Error string `json:"error"`
	}{OK: false, Code: string(code), Error: err.Error()})
	return true
}

// jsonRequested reads --json off the RAW ARGV.
//
// It cannot come from the parsed flag set: this runs on the path where parsing FAILED, so the
// flag may never have been bound. Scanning argv is the only source that survives that, and the
// flag is a plain boolean with no value form to confuse it.
func jsonRequested(argv []string) bool {
	for _, a := range argv {
		if a == "--json" || a == "--json=true" {
			return true
		}
	}
	return false
}
