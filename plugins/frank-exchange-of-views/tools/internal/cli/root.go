// Package cli assembles feov-record: the root command, the flags every seat
// shares, and each role's verb set.
//
// The seat-side record runtime. ONE binary whose root carries the verb set of the
// seat asking, because the verb set IS the role boundary — a lens mints, regrades
// and closes its own gaps and no other seat has those verbs, blue has no board
// verbs at all, and the bench rules without originating. A seat's role is its
// roster row (seatclass.RoleOfSeat), so the boundary is enforced rather than
// merely described.
package cli

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/bench"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/chair"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
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

// seatVerbs maps a resolved role to its surface and its one-line description. Built here rather
// than in each package so the root can answer "which seat am I" once.
func seatVerbs(role string) ([]*cobra.Command, string) {
	switch role {
	case "lens":
		return lens.Verbs(), "red lens seats — findings, source verification and corroboration, computation re-runs; mints its own gaps against its budget and, as the originator, regrades and closes them."
	case "chair":
		return chair.Verbs(), "the red chair — runs the debate: dispatches, judges the board, carries, spot-checks."
	case "blue":
		return blue.Verbs(), "blue seats — the report and its edits, manifest rows, inquiries. No board verbs at all."
	case "bench":
		// "opinions" NAMED A VERB THE BENCH NO LONGER HAS. `bench opinion` is retired (#681) and
		// the disposition is a docket motion's ruling, so the line offered a capability the
		// surface below it does not list — the exact mismatch a seat reads as "the tool is
		// missing something" and works around.
		return bench.Verbs(), "the bench — rulings, halt, certification. Never originates."
	}
	return nil, ""
}

// noSeatNote explains an OPERATOR surface to something that expected a seat's.
//
// THIS IS THE COST OF SCOPING THE TREE, PAID RATHER THAN HIDDEN. A seat verb is now absent from a
// process with no identity, so "mint" answers "no such command" where it used to answer "--seat-id
// is required". Under a scoped surface "not yours" reads as "does not exist", and a seat handed an
// unavailable verb logs it as a request and works around it, losing the capability for the run. The
// surface cannot carry the verb (it does not know
// which seat's), so the REFUSAL has to carry the reason.
func noSeatNote(seatID string) string {
	if strings.TrimSpace(seatID) == "" {
		return "\n\n--seat-id IS REQUIRED HERE, because nothing identified you: this agent has not " +
			"registered, so the record cannot say which seat it holds, and no --seat-id was given. The " +
			"surface is scoped to whoever is asking, so there is no tree to show until you say who that " +
			"is — a seat id, or `" + record.OperatorRole + "`. You type it ONCE, at `register`; after that " +
			"the binding is on the record and every later call resolves it for you."
	}
	if seatclass.RoleOfSeat(seatID) == "" {
		return "\n\n" + seatID + " does not belong to any role namespace, so it selects no surface. The engine " +
			"assigns seat ids; a hand-invented one identifies nobody, and a record written under it would be " +
			"attributed to an identity no dispatch created."
	}
	return ""
}

// readsSources is seat.ReadsSources: the one statement of which roles read a source, shared with
// the `show source` projection.
func readsSources(role string) bool { return seat.ReadsSources(role) }

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
	runDir, err := seatenv.Resolve("", func() string {
		return runlive.InferRunDir("", runlive.Caller{ID: seatenv.AgentID(), Type: seatenv.AgentType(), Session: seatenv.Session()}).Dir
	})
	if err != nil {
		return seatenv.Dispatched(nil)
	}
	// NewRun, matching seat.Of: this asks which seat the record binds this agent to, and must answer for a
	// run that is not yet on disk rather than refusing before the question is put.
	run, rerr := record.NewRun(runDir)
	if rerr != nil {
		return seatenv.Dispatched(nil)
	}
	return seatenv.Dispatched(seat.BoundSeat(run))
}

// NewRootFor builds THIS SEAT'S surface at the root, or the operator's when no seat was dispatched.
//
// THERE IS NO ROLE LEVEL. A seat does not retype its identity as a command word: its role is read
// from its seat id, and the root IS that role's surface. A lens seat runs `mint`; a chair seat
// running `mint` is told the verb is the lens seat's — the boundary drawn by the only fact that
// decides it.
//
// The operator tree is the one with no seat: setup, capture, dashboard and the rest have no role
// to scope to, and a human running them is not a seat.
func NewRootFor(seatID string) *cobra.Command {
	root := &cobra.Command{
		Use:   InvokedAs(),
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: InvokedAs() + ` — the seat-side record runtime (the verb set IS the role boundary)

A lens mints its own gaps and, as the originator, regrades and closes them; no
other seat has a mint, regrade or close verb. Blue has no board verbs at all. The
bench rules and never originates.`,
		Version:       buildid.Revision(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// COBRA'S SHELL-COMPLETION GENERATOR IS NOT A SEAT'S VERB, and it was on every seat's surface:
	// the group plus one command per shell, five paths a seat can name and none it can use. No agent
	// generates a zsh completion script.
	//
	// THE COST IS PAID AT THE OPENING ACT. A seat is told to read its WHOLE surface before deciding
	// what to do — that instruction is the fix for seats opening 6 of 51 group pages — so every path
	// that cannot be an act is read by every seat in every sitting, and agentgen inlines the same
	// listing into all eleven agent definitions.
	root.CompletionOptions.DisableDefaultCmd = true

	// The two flags EVERY verb needs, declared once and inherited. Persistent
	// flags are the mechanism the first cut of this CLI re-implemented by
	// re-declaring --run and --seat-id on all sixteen verbs.
	// --schema PRINTS ONE INTEGER AND EXITS. setup compares it against the plugin's
	// requirements.json, and giving it a flag of its own means nothing has to recover a number
	// from --version's prose — the string-shaped hope this repository keeps deleting.
	root.PersistentFlags().Bool(flags.Schema, false, "print the event-schema epoch this binary writes, and exit")

	root.PersistentFlags().String(flags.Run, "", "the run directory — the PreToolUse hook injects it in a real run, so you rarely type it. A value that DISAGREES with the run you were dispatched into is refused")
	// NO BACKTICKS AROUND `register`: cobra takes the first backquoted word as the flag's
	// placeholder, and it rendered `--seat-id register`.
	root.PersistentFlags().String(flags.SeatID, "", "your seat id, as the dispatch prompt states it (SEAT_ID). Pass it ONCE, at register, which binds it to you on the record; every later call resolves it, so typing it is optional. It SELECTS this surface (the verbs listed are the ones your seat may run); a value disagreeing with the seat the record binds you to is refused")
	// --json makes every mutating verb emit a structured result and every failure a
	// structured error, so a machine consumer parses fields instead of prose. On READS the
	// format is primarily view-selected: board/findings/work/motions/telemetry/evidence are JSON
	// by name (seat.JSONByNameViews is the set; each one says so in its own --help), the rest
	// are markdown. --json opts a markdown view into its structured form where one exists
	// (today only `show debate --json`); on a JSON-by-name view it is accepted and changes
	// nothing (the same bytes are not a second form), and on a markdown view with no JSON
	// form it is an error, because the tool cannot give what was asked for.
	root.PersistentFlags().Bool(flags.JSON, false, "emit a structured JSON result (and structured errors) instead of human text")

	// A SEAT IS NOT AN OPERATOR, and the two sets are disjoint by construction rather than by
	// anyone remembering. Promoting the seat's verbs to the root put them beside the operator
	// commands, and a name collides outright: `verify` is a lens's citation adjudication AND the
	// operator's whole-record cross-check. One tree cannot hold both meanings, and picking a
	// winner would hand some seat a verb that does the other thing. The operator's read of the
	// log is `ops log` for the same reason: `log` only ever writes.
	//
	// So the split is total. `fetch` crosses it because the seats that read sources genuinely run
	// it (readsSources) — a lens reads blue's cached source bytes — and the name does not collide. `count-claims` does not cross:
	// no seat is asked for the report's claim count, and a reader that wants it counts the report.
	// NO IDENTITY IS NOT A MODE. It used to fall through to the operator surface, which made
	// "nobody said who I am" a way of selecting a tree — and a tree nobody selects is one nobody
	// can be refused from. The surface is scoped to whoever is asking, so there is nothing to show
	// until that is answered; `--seat-id operator` is how an operator answers it.
	switch role := seatclass.RoleOfSeat(seatID); {
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
		// THE LOG FOOTER MOVED WITH THE VERBS, not away with the group. It hung on the role
		// group's Long, and deleting the group would have dropped it silently — a seat would stop
		// being told that a capability it cannot find is a fact about the tooling rather than
		// something to work around. Caught by TestRoleHelpCarriesTheLogFooter.
		root.Long = InvokedAs() + " — " + short + "\n" + seat.LogFooter
		root.AddCommand(verbs...)
		root.AddCommand(motion.NewCommandFor(role))
		if readsSources(role) {
			root.AddCommand(newFetch()) // a lens reads the EXACT bytes blue read, from the run cache
		}
		// `manual` IS NOT ON A SEAT'S SURFACE, because a seat is already holding what it prints.
		//
		// scripts/agentgen writes every command's help into the seat's own constitution, so the
		// document arrives with the system prompt and costs no turns. Leaving the verb here left an
		// affordance with no job: measured on the 2026-09-20 run, the chair ran `manual` twice while
		// its constitution carried the same text, and piped both through `head` — ending up with
		// LESS of its surface than it was already given. Removing the need did not remove the
		// reaching; removing the verb does.
		//
		// It stays on the operator surface with --for, which is how a generator renders a seat's
		// surface without being that seat.
		// AFTER every AddCommand: the split reads HasSubCommands, so a group registered before
		// its children were attached would file itself under the leaves.
		seat.SplitGroups(root)
	default:
		root.AddCommand(
			newVerify(),          // operator cross-check, not a seat role — read-only over the record
			newGraph(),           // operator: render a run's actual behaviour from the record
			newShowDiagnostics(), // operator: did the seats find their surface — exposure and use, as fields
			newCountClaims(),     // operator: the report's claim count, counted from the report on the record
			newFetch(),           // operator: cached, hash-verified web read — feeds the run source cache
			newOCR(),             // operator: render a scan's pages so a seat can read what has no text layer
			newSetup(),           // operator: build a research run's blackboard
			newScorecard(),       // operator: this run's scorecards — every card, or one with --card
			newDashboard(),       // operator: the live run dashboard.html
			newCapture(),         // operator: the post-hoc capture auditor
			newEngineStopped(),   // operator: the engine's own stop, written to the log
			newTakeOver(),        // operator: a resuming session becomes the lead of its run
			newManual(seatID),    // every surface: every command's own help, in one call
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
			return seat.RefuseAndTeach(cmd, feov.MissingField, "you named no command, so nothing ran."+noSeatNote(seatID))
		}
		return seat.RefuseAndTeach(cmd, feov.Validation, unknownCommandRefusal(cmd.Root(), args[0])+noSeatNote(seatID))
	}

	teachUnknownSubcommand(root)
	// THE HANDLE'S OWN LINE SAYS WHICH VERBS TAKE ONE, from this tree: a handle described on three
	// verbs and bounded on none was carried to a fourth.
	seat.NameTakers(root, flags.Key, "a handle")
	enumhelp.Install(root)
	return root
}

// whereItLives finds the full paths a bare verb name is reachable at, so a refusal can say WHERE
// rather than only NO.
//
// A refusal for a bare seat verb names the seat that holds it — "it exists, on the blue seat, but
// not for you. That is a wrong-seat error rather than a missing capability" — and never says the
// verb does not exist: a seat told so believes it, goes hunting, and settles confidently on a
// different act.
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

// teachUnknownSubcommand gives every verb GROUP the refusal the root already gives: an act that is
// not in the group is REFUSED, and where another seat's surface holds it, that seat is named.
//
// A GROUP WITH NO RunE ANSWERED AN UNKNOWN SUBCOMMAND WITH ITS OWN HELP AND EXIT 0. Cobra parses the
// word as a positional argument to a command that cannot run, prints that command's help, and
// reports success — so a seat typing an act it does not hold was told, in effect, that it had done
// it. `inquiry` is a group on blue's surface (propose, record) AND the chair's (review), so each
// seat can type the other's act under a group it does hold; `motion <no-such-subject>` is the same
// case.
//
// THE HOLDER IS DERIVED, NEVER LISTED — whereSubcommandLives asks every role's real tree, the way
// whereItLives does one level up, so a subcommand moved between surfaces is named correctly with no
// second table to keep in step. It is read when the refusal is RENDERED, not when the tree is built,
// because AllRoots builds every tree through this function.
func teachUnknownSubcommand(root *cobra.Command) {
	for _, g := range root.Commands() {
		if !g.HasSubCommands() || g.RunE != nil || g.Run != nil {
			continue
		}
		group := g
		group.Args = cobra.ArbitraryArgs
		// THE ACT'S FLAGS ARE NOT THE GROUP'S, and cobra parses flags before it runs this: a seat
		// that types another seat's act types that act's flags with it, and the parse would answer
		// "unknown flag: --id" — a problem with a flag, about a command that is not the seat's at
		// all. The group has no flags of its own, so whatever lands here is already a refusal; the
		// only question is which one the seat is told.
		group.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
		group.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			sub := args[0]
			if at := whereSubcommandLives(group.Name(), sub); len(at) > 0 {
				// THE ACT IS NAMED AS ITS HOLDER'S. Quoted bare, `inquiry review` reads to blue as a
				// command it can type; it is the chair's, and the sentence says so where the path is.
				holders := make([]string, len(at))
				for i, a := range at {
					holders[i] = "the " + a + "'s"
				}
				return seat.RefuseAndTeach(cmd, feov.RoleViolation, fmt.Sprintf("This act is %s `%s %s`, and it is not on your surface: you are %s. "+
					"That is a wrong-SEAT error, not a missing capability, so do not work around it: if the act is yours to "+
					"perform, you are dispatched as the wrong seat, and if it is not, the seat that holds it is named here.",
					strings.Join(holders, " and "), group.Name(), sub, seat.DispatchedAs(cmd)))
			}
			var have []string
			for _, s := range group.Commands() {
				if s.IsAvailableCommand() {
					have = append(have, s.Name())
				}
			}
			return seat.RefuseAndTeach(cmd, feov.Validation, fmt.Sprintf("no `%s` act exists here, on any seat's surface — yours has %s.",
				sub, strings.Join(have, ", ")))
		}
	}
}

// whereSubcommandLives is whereItLives one level down: which roles hold `<group> <sub>`.
func whereSubcommandLives(group, sub string) []string {
	var out []string
	for role, r := range AllRoots() {
		for _, c := range r.Commands() {
			if c.Name() != group {
				continue
			}
			for _, s := range c.Commands() {
				if s.Name() == sub || s.HasAlias(sub) {
					out = append(out, role)
					break
				}
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
			"holds it is named here.",
			name, strings.Join(at, " and "))
	}
	return fmt.Sprintf(
		"no command named %q exists, on your surface or on any other seat's, so there is no seat to switch to.", name)
}

// refuseUnknownCommandFirst answers the WRONG COMMAND before cobra can answer the wrong flag.
//
// Cobra parses flags before it decides a command is unknown, so `show --view board` reported
// `unknown flag: --view` — naming the one thing the caller had right. The flag is fine; it just
// belongs to `blue show`, `chair show`, `lens show` or `bench show`. A seat reading that goes
// hunting through view names and never meets the refusal that would have taught it the role
// prefix, which inverts the whole point of a teaching refusal.
//
// MEASURED, and it is the single most common failure a seat hits: across nine probed seats,
// 9 of 21 non-zero exits were this message, and SIX OF NINE SEATS produced it on their FIRST
// tool call — the moment a seat is orienting and least able to spare a wasted turn. `show` is
// per-role while the concept ("show the board") carries no role, so dropping the prefix is the
// natural slip rather than a careless one.
//
// IT READS THE ARGV BY THE ROOT'S FLAGS, never cobra's error text, which would make this depend on
// the shape of a string another library owns — the failure mode this repo keeps finding. A root
// flag's value is skipped by what the parser knows of the flag; a verb's flag before the verb has
// no known arity here, and is named in the refusal rather than guessed past (below).
// seatID is passed rather than read from the environment: the tests drive cobra in-process with
// an args slice that never reaches os.Args, so a check reading the env would refuse every call.
func refuseUnknownCommandFirst(root *cobra.Command, argv []string, seatID string) error {
	if len(argv) < 2 {
		return nil
	}
	// COBRA BUILDS PART OF ITS OWN SURFACE DURING Execute, and refusing before that showed the
	// seat a SMALLER surface than the one that exists: `help`, `--help` and `--version` were all
	// missing from the list. The golden caught it — a refusal that lists
	// less than the tool has is a worse lie than the flag error this replaces, because it reads
	// as authoritative.
	root.InitDefaultHelpCmd()
	root.InitDefaultHelpFlag()
	root.InitDefaultVersionFlag()

	// THE COMMAND IS THE FIRST NON-FLAG WORD, not argv[1]. `--json mint …` put a flag there, so
	// the pre-check saw one and stood aside, and cobra then complained about `--class` — a flag on
	// a command it had not found. The root's own flags are read off the root, so a flag that takes a
	// value has it skipped by what the parser knows of it.
	//
	// A FLAG THE ROOT DOES NOT DECLARE IS A VERB'S, typed before the verb, and its arity is unknown
	// until a verb is found — so its next word is NOT skipped, and is read as the name. `before`
	// keeps that flag, so the refusal below can say what happened: `--key adv-f-tldr finding` read
	// "adv-f-tldr" as a command and was told no command of that name exists.
	name, before := "", ""
	for i := 1; i < len(argv) && name == ""; i++ {
		a := argv[i]
		if !strings.HasPrefix(a, "-") {
			name = a
			continue
		}
		flag, _, inline := strings.Cut(strings.TrimLeft(a, "-"), "=")
		f := root.Flags().Lookup(flag)
		if f == nil {
			f = root.PersistentFlags().Lookup(flag)
		}
		if f == nil && !strings.HasPrefix(a, "--") && len(flag) == 1 {
			f = root.Flags().ShorthandLookup(flag)
		}
		switch {
		case f != nil:
			before = ""
			if !inline && f.NoOptDefVal == "" {
				i++ // and its value
			}
		case inline:
			before = "" // --flag=value: no separate value to read as the name
		default:
			before = flag
		}
	}
	if name == "" {
		return nil // flags only: cobra's own handling is right for --help and --version
	}
	// NOBODY SAID WHO IS ASKING, so there is no surface to look this up in. Answered here rather
	// than after parsing, because the flags belong to a command that could not have been found.
	if seatclass.RoleOfSeat(seatID) == "" {
		return seat.RefuseAndTeach(root, feov.MissingField, "a command was named and the surface is scoped to whoever is asking."+noSeatNote(seatID))
	}
	for _, c := range root.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return nil
		}
	}
	if before != "" {
		if takers := seat.TakersOf(root, before, nil); len(takers) > 0 {
			return seat.RefuseFlagBeforeVerb(root, before, name, takers)
		}
	}
	// A WRONG-SEAT ERROR KEEPS ITS CODE. `--json` consumers branch on the KIND of failure, and
	// this one is a role violation however the tree came to lack the verb; flattening it to a
	// plain message makes "you are the wrong seat" indistinguishable from a typo.
	wrong := unknownCommandRefusal(root, name) + noSeatNote(seatID)
	if len(whereItLives(root, name)) > 0 {
		return seat.RefuseAndTeach(root, feov.RoleViolation, wrong)
	}
	return seat.RefuseAndTeach(root, feov.Validation, wrong)
}

// Execute runs the CLI.
func Execute() {
	if marker := os.Getenv(seatenv.VarComputeRerun); marker != "" {
		refuseInsideComputeRerun(marker, os.Args[1:])
		os.Exit(2)
	}
	if _, err := Invoke(newRoot(), os.Args[1:], dispatchedSeat()); err != nil {
		os.Exit(2)
	}
}

// Invoke is everything the binary does with an argv short of exiting: the refusal that answers a
// command nobody holds before cobra parses a flag, the tree, and — for a refusal from either — the
// ONE renderer, handed the command that resolved. It returns that command with the error.
//
// Exported because a TEST HARNESS must reproduce the binary. One that drove the tree and rendered
// for itself would measure a shape the binary does not produce; the in-process drivers (this
// package's `run`, the release gate's, and `manual`'s) call this instead, so there is no second
// copy of the sequence to keep in step.
//
// The refusal goes to the root's own writers — the process's stdout and stderr unless a harness
// set others.
func Invoke(root *cobra.Command, argv []string, seatID string) (*cobra.Command, error) {
	root.SetArgs(argv)
	leaf, err := root, refuseUnknownCommandNoted(root, append([]string{InvokedAs()}, argv...), seatID)
	if err == nil {
		leaf, err = ExecuteRoot(root)
	}
	if err != nil {
		seat.RenderRefusal(root.OutOrStdout(), root.ErrOrStderr(), leaf, argv, err)
	}
	return leaf, err
}

// ExecuteRoot runs the tree and returns the command that resolved with its error. To the refusals
// COBRA raises — an unknown flag, a missing required flag, a bad value, wrong arguments, a flag
// group violated — it gives the code their producer implies and the answer the command tree holds.
//
// Cobra's own sentences name a flag word and stop: which word was missed, and nothing about what
// the word wants or what else the verb takes. Each is answered in a few lines, on every command
// under the root: required flags left out get each missing flag's own usage line
// (seat.RefuseMissingFlags); a flag the verb does not have gets the verbs on this surface that do
// take it beside this verb's own flags (seat.RefuseUnknownFlag); a value a flag does not take gets
// the flag, what it was given and what it accepts (seat.RefuseBadValue); a word with no flag in
// front of it gets this verb's flags (seat.RefuseStrayArgs); and a flag group violated gets the
// usage line of each flag in it (seat.RefuseFlagGroup). Each names the verb's `--help` and prints
// none of it.
//
// A refusal of cobra's that none of those reads gets the verb's whole page on a SEAT VERB, the
// diagnosis leading. A handler's refusal is a teaching message already (seat.Taught marks them).
//
// THE REFUSAL IS LOGGED AS RETURNED, once its diagnosis is final: the log entry and the rendered
// forms read the same sentence.
func ExecuteRoot(root *cobra.Command) (*cobra.Command, error) {
	// BEFORE COBRA DISPATCHES ANYTHING. --schema is not a verb and must answer even when the
	// argv would otherwise be refused: setup runs it against a binary it does not yet trust,
	// and a preflight that cannot get an answer out of a working binary is worse than no
	// preflight. One integer on stdout, nothing else, so the caller parses a number and not a
	// sentence.
	for _, a := range os.Args[1:] {
		if a == "--"+flags.Schema {
			fmt.Fprintln(root.OutOrStdout(), record.EventSchema)
			return root, nil
		}
	}
	// A FLAG THAT DOES NOT PARSE IS ILL-FORMED INPUT, and cobra calls this for parse errors and for
	// nothing else — so the code is set where the producer is known, not guessed from the text.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return feov.Wrap(feov.Validation, err, "%s", err.Error())
	})
	cmd, err := root.ExecuteC()
	if cmd == nil {
		cmd = root
	}
	if err == nil {
		return cmd, nil
	}
	if !seat.Taught(err) {
		err = answerCobraRefusal(cmd, err)
	}
	noteRefusal(cmd, err)
	return cmd, err
}

// answerCobraRefusal gives a refusal cobra raised its code and its answer, by producer.
func answerCobraRefusal(cmd *cobra.Command, err error) error {
	// THE PRODUCERS ARE ASKED BEFORE THE CODE IS GIVEN: coding rewrites one of their sentences.
	missing := isMissingRequired(cmd, err)
	stray := isRefusal(err, cobra.NoArgs(cmd, cmd.Flags().Args()))
	group := isRefusal(err, cmd.ValidateFlagGroups())
	err = codeCobraRefusal(cmd, err)
	var unknown *pflag.NotExistError
	var invalid *pflag.InvalidValueError
	switch {
	case cmd == cmd.Root():
		// No verb resolved, so there is no verb's flags to answer with.
		return err
	case missing:
		if short := seat.RefuseMissingFlags(cmd, err); short != nil {
			return short
		}
	case errors.As(err, &unknown):
		said, answered := translateUnknownFlag(cmd, unknown.GetSpecifiedName(), unknown.Error())
		return seat.RefuseUnknownFlag(cmd, unknown.GetSpecifiedName(), said, answered, err)
	case errors.As(err, &invalid):
		// THE FLAG AND THE VALUE ARE THE PARSER'S FIELDS, and the sentence is the flag's own Value's
		// — except a standard-library Value's, which is the library's parse error and names
		// neither what the flag takes nor the flag.
		f, said := invalid.GetFlag(), invalid.Unwrap().Error()
		var library *strconv.NumError
		if errors.As(invalid.Unwrap(), &library) {
			said = ""
		}
		_, enumerated := enumhelp.Registered(cmd)[f.Name]
		return seat.RefuseBadValue(cmd, f, invalid.GetValue(), said, enumerated, err)
	case stray:
		return seat.RefuseStrayArgs(cmd, err.Error(), err)
	case group:
		if short := seat.RefuseFlagGroup(cmd, err.Error(), err); short != nil {
			return short
		}
	}
	if seat.RecordType(cmd) != "" {
		return seat.TeachCause(cmd, err.Error(), err)
	}
	return err
}

// isMissingRequired says whether an error is cobra's refusal of required flags left out.
//
// THE PRODUCER IS ASKED AGAIN, not recognised by its wording: cobra's check is re-run on the
// resolved command and its message compared whole with the one returned — the same function over
// the same flag state — so no shape of another module's text is assumed.
func isMissingRequired(cmd *cobra.Command, err error) bool {
	again := cmd.ValidateRequiredFlags()
	return again != nil && again.Error() == err.Error()
}

// codeCobraRefusal gives a refusal cobra raised AFTER parsing the code its producer implies: a
// required flag left out is a missing field; a flag group violated, or arguments the command does
// not take, are ill-formed input.
//
// Each of cobra's checks is asked again, as isMissingRequired asks the first, and a refusal none
// of them produced keeps the code it came with. An error nothing coded stays "error": a network
// failure is not the caller's ill-formed input, and nothing here says it is.
func codeCobraRefusal(cmd *cobra.Command, err error) error {
	var coded *feov.Error
	if errors.As(err, &coded) {
		return err
	}
	same := func(e error) bool { return isRefusal(err, e) }
	switch {
	case isMissingRequired(cmd, err):
		return feov.Wrap(feov.MissingField, err, "%s", err.Error())
	case same(cobra.NoArgs(cmd, cmd.Flags().Args())):
		// COBRA'S SENTENCE HERE NAMES THE COMMAND — `unknown command "x" for "<tool> <path>"` — and
		// calls an argument a command. The renderer prints the path, so this says what is wrong.
		return feov.Wrap(feov.Validation, err, "takes no arguments; got %s", quotedArgs(cmd.Flags().Args()))
	case same(cmd.ValidateFlagGroups()), same(cmd.ValidateArgs(cmd.Flags().Args())):
		return feov.Wrap(feov.Validation, err, "%s", err.Error())
	}
	return err
}

// isRefusal says whether `again` — one of cobra's checks, re-run on the resolved command — is
// the refusal `err`: the same function over the same state, compared whole.
func isRefusal(err, again error) bool { return again != nil && again.Error() == err.Error() }

// quotedArgs is an argument list as the seat typed it, each word quoted.
func quotedArgs(args []string) string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = strconv.Quote(a)
	}
	return strings.Join(out, " ")
}

// translateUnknownFlag answers the seat's ACTUAL mistake when it typed the record's word for a
// field instead of the flag's.
//
// THE THREE VOCABULARIES ARE REAL AND THE SEAT IS TOLD SO. Its own contract says "the projection
// names, this prompt's words for a concept, and the command that writes it are three different
// vocabularies and they do not always agree". A seat that reads a gap's JSON, sees
// `complexity_cost`, and types `--complexity_cost` has used the vocabulary the record just showed
// it — the most defensible mistake available. It cost twelve turns in one run.
//
// IT SUGGESTS, IT DOES NOT ACCEPT. The no-aliases rule is right: one word per concept, and adding
// `--complexity_cost` as a second spelling is the fork this repository refuses. Naming the correct
// flag in the refusal respects the rule; accepting both would not.
//
// A key with no entry falls through untouched, because flags.FlagForPayloadKey reports the miss
// instead of hyphenating a guess — a suggested spelling the parser rejects is worse than none.
//
// `typed` is the flag word as the parser read it (pflag's NotExistError carries it as a field) and
// `msg` the parser's sentence; translated says whether the sentence now names the flag to type.
func translateUnknownFlag(cmd *cobra.Command, typed, msg string) (said string, translated bool) {
	name, ok := flags.FlagForPayloadKey(typed)
	if !ok || name == typed {
		return msg, false
	}
	// THE MAP IS GLOBAL AND PAYLOAD KEYS ARE NOT, which its own comment says plainly. `prose` maps
	// to --reason, and a verb with no --reason would be told to type one. The suggestion is
	// therefore made only when the FAILING VERB actually has that flag — which turns a global
	// table into a per-verb answer without a second table to keep.
	if cmd == nil || cmd.Flags().Lookup(name) == nil {
		return msg, false
	}
	return fmt.Sprintf("%s — the record stores that field as `%s`, and the flag that sets it is --%s. "+
		"The field name and the flag word are not always the same; the flag is what you type.",
		msg, typed, name), true
}
