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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/bench"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/merge"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/motion"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The CHANGELOG THAT WAS HERE IS GONE (#407). 70 entries, ~1000 lines, 96% of this file,
// with no reader but a person — and a person has `git log`, which cannot drift from the
// commits it describes. Verified before deleting rather than assumed: every entry's headline
// resolves to a real commit, and the bodies carry the same reasoning in more detail.
//
// Its one load-bearing part is kept and made reachable: the 48 "A stale binary …" sentences
// are now fields in record.capabilityDeltas, which `setup`'s preflight PRINTS when it refuses
// a skewed binary. The precedent is entry 0.68.0's own argument, which retired a hand-written
// changelog for duplicating a canonical record — stated in entry 67 of one.
//
//	`Append` and `RegisterSeat` stamped `Round: RoundOf(seatID)` — a regex over the seat
//	id, 0 on a miss — while the caller had already resolved the round as a field on
//	seat.Context and `Begin` had already refused an unresolvable seat. The fact was in
//	hand at the call site and thrown away one frame later, at 33 of them.
//
//	Both now take a `record.Identity{RunDir, SeatID, Round}`, built by
//	`seat.Context.Identity()`. NO VALUE CHANGES: every difftest golden is byte-identical,
//	which is the evidence rather than the hope — for any seat that reached a verb body,
//	`Context.Round` came through the same resolution this used to call. The point is the
//	SEAM: when the dispatcher injects a round (#290) it lands in one place.
//
//	ROLE IS NOT THREADED, and #396 asked for it. `seat.Context.Role` answers which command
//	GROUP a verb is mounted under; `Event.Role` is the PARTY that wrote it. They disagree —
//	`motion grade file` run by `blue-respond-r1` is party `blue`, command-group `grade` —
//	so threading it would relabel the author of every motion event. cli/seat's own
//	isRoleName says the two are "a different question with a different answer"; the
//	Event.Role doc now says so too, in place of its false claim to be "never re-derived".
//
//	An unknown round writes -1, not 0. Nothing produces one today; it becomes reachable
//	only when an injected identity conflicts with a typed --seat-id.
//
//	A stale binary stamps the round its seat id looks like.
//
// versionsync_test.go asserts this equals recordToolVersion in the plugin manifest, which
// is what setup preflights against. Without that test the two drift and the preflight
// compares a stale number to itself.
const Version = "0.71.0"

func init() { record.ToolVersion = Version }

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

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   InvokedAs(),
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: InvokedAs() + ` — the seat-side record runtime (the verb set IS the role boundary)

A lens structurally cannot mint or close a board gap: no such verb exists in its
namespace. Blue has no board verbs at all. The bench rules and never originates.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// The two flags EVERY verb needs, declared once and inherited. Persistent
	// flags are the mechanism the first cut of this CLI re-implemented by
	// re-declaring --run and --seat-id on all sixteen verbs.
	root.PersistentFlags().String(flags.Run, "", "the run directory. REQUIRED unless the engine injected it — it does in a real run, which is why you rarely type it. A value that DISAGREES with the run you were dispatched into is refused rather than obeyed")
	root.PersistentFlags().String(flags.SeatID, "", "your seat id, assigned by the engine and bound to this role's namespace")
	// --json makes every mutating verb emit a structured result and every failure a
	// structured error, so a machine consumer parses fields instead of prose. On READS the
	// format is primarily view-selected: board/findings/friction are JSON by name, the rest
	// are markdown. --json opts a markdown view into its structured form where one exists
	// (today only `show --view debate --json`); it is an error on a JSON-by-name view or a
	// markdown view with no JSON form, so there is exactly one way to reach each form.
	root.PersistentFlags().Bool(flags.JSON, false, "emit a structured JSON result (and structured errors) instead of human text")

	root.AddCommand(
		lens.NewCommand(),
		merge.NewCommand(),
		motion.NewCommand(),
		blue.NewCommand(),
		bench.NewCommand(),
		newVerify(),      // operator cross-check, not a seat role — read-only over the record
		newGraph(),       // operator: render a run's actual behaviour from the record
		newCountClaims(), // operator/blue: deterministic claim_count over blue/report.md
		newFriction(),    // operator: the friction channel — seats write it, the human reads it
		newFetch(),       // operator: cached, hash-verified web read (replaces WebFetch) — feeds the run source cache
		newSetup(),       // operator: build a research run's blackboard (ported from setup-research-run.mjs)
		newScorecard(),   // operator: a chair's in-run self-read scorecard (ported from scorecards.mjs)
		newDashboard(),   // operator: the live run dashboard.html (ported from render-run-dashboard.mjs)
		newCapture(),     // operator: the post-hoc capture auditor (ported from capture-research-run.mjs)
		newHook(),        // hook backend: the blue-report lockdown PreToolUse/PostToolUse gates (invoked by hooks.json)
	)
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
			return seat.RefuseAndTeach(cmd, "you named no command, so nothing ran. The commands below are the whole surface.")
		}
		return seat.RefuseAndTeach(cmd, fmt.Sprintf(
			"no command named %q exists. The commands below are the whole surface, and each one's own `--help` carries its verbs.", args[0]))
	}

	enumhelp.Install(root)
	return root
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
func refuseUnknownCommandFirst(root *cobra.Command, argv []string) error {
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

	name := argv[1]
	if name == "" || strings.HasPrefix(name, "-") {
		return nil // a flag, or help/version: cobra's own handling is right for those
	}
	for _, c := range root.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return nil
		}
	}
	return seat.RefuseAndTeach(root, fmt.Sprintf(
		"no command named %q exists. The commands below are the whole surface, and each one's own `--help` carries its verbs.", name))
}

// Execute runs the CLI. Abort-safety is armed first: a seat killed mid-command
// must lose an event, never leave a torn one or a stuck lock.
func Execute() {
	defer record.InstallSignalGuard()()

	root := newRoot()
	if err := refuseUnknownCommandFirst(root, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", InvokedAs(), err)
		os.Exit(2)
	}
	if err := root.Execute(); err != nil {
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
	_ = json.NewEncoder(w).Encode(struct {
		OK    bool   `json:"ok"`
		Code  string `json:"code"`
		Error string `json:"error"`
	}{OK: false, Code: string(feov.Validation), Error: err.Error()})
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
