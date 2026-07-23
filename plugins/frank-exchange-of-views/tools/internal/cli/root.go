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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/bench"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/merge"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Version is stamped on register events and answered by --version. The setup
// preflight compares it against the plugin manifest BEFORE the run-live marker
// is written, so a skewed binary fails at setup rather than mid-round.
//
// It tracks the binary's OBSERVABLE CONTRACT, not the plugin release: bump it when a run
// would behave differently against a stale binary — the events on disk, the projections
// render writes, or the verb/flag surface a prompt may call. It is deliberately decoupled
// from the plugin version, which says what shipped.
//
//	0.2.0  the 2026-07-19 schema work — events gained `ts` and replay orders by it,
//	       findings gained a tool-assigned `finding_id`, four flags were renamed with
//	       aliases deleted, cross-references enforced at write time.
//	0.3.0  render computes `realized_open` in the board telemetry (Phase B1) — a stale
//	       binary renders telemetry a downstream reader now expects.
//	0.4.0  the global `--json` flag: mutating verbs emit a structured result and errors,
//	       so a machine consumer a prompt drives needs the binary that speaks it.
//	0.5.0  the --json envelope carries `role` and, on failure, a `code` (feov.Error) a
//	       consumer branches on — a wire-shape change a stale binary would not produce.
//	0.6.0  `bench assemble` writes report.md in-tool (verbatim sections + board risk matrix);
//	       a prompt that drives it needs the binary that has the verb.
//	0.7.0  `bench outcome` records the terminal verdict as an event, and `bench assemble`
//	       drops --inputs entirely: the report is composed from the record (event log +
//	       board) and blue's audited report, so a driving prompt needs the inputs-free verb
//	       and the outcome verb it now depends on.
//	0.8.0  `merge verdict --as PASS` is REFUSED while any gap is still open — the gate is
//	       enforced at the tool, not trusted to the seat. A prompt that drives it sees a new
//	       failure it must handle (close the gaps or issue FAIL).
//	0.9.0  the prose vocabulary collapses to --reason (+ --reason-file); --file/--text/
//	       --comment/--basis are retired, and every claim/judgment act (mint, close,
//	       closing, dispute, dispute-respond, opinion, retire, halt, certify) now requires
//	       its prose. A prompt driving the old flags sees unknown-flag and new refusals.
//	0.10.0 `bench assemble` composes a materially different report.md: a top "## Read this
//	       first" orientation (open gaps ranked from the board + the bench's certify/halt
//	       voice promoted), the blue-report embed carries ONLY blue's non-composed remainder
//	       (its lifted and tool-owned sections are dropped, killing the lift-AND-embed
//	       duplication and the stale-verdict contradiction), and lens findings the merge did
//	       not mint are surfaced. A stale binary silently produces the duplicated, contradicted
//	       report the preflight must now refuse.
//	0.11.0 --gap-id is accepted as an alias for --id on every verb (the payload key is gap_id,
//	       the spelling a seat reaches for). A stale binary rejects --gap-id as an unknown flag,
//	       so a seat that uses the intuitive spelling fails against it.
//
// versionsync_test.go asserts this equals recordToolVersion in the plugin manifest, which
// is what setup preflights against. Without that test the two drift and the preflight
// compares a stale number to itself.
const Version = "0.11.0"

func init() { record.ToolVersion = Version }

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "feov-record",
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: `feov-record — the seat-side record runtime (the verb set IS the role boundary)

A lens structurally cannot mint or close a board gap: no such verb exists in its
namespace. Blue has no board verbs at all. The bench rules and never originates.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// The two flags EVERY verb needs, declared once and inherited. Persistent
	// flags are the mechanism the first cut of this CLI re-implemented by
	// re-declaring --run and --seat-id on all sixteen verbs.
	root.PersistentFlags().String(flags.Run, "", "the run directory (the engine passes it; it is in your prompt)")
	root.PersistentFlags().String(flags.SeatID, "", "your seat id, assigned by the engine and bound to this role's namespace")
	// --json makes every mutating verb emit a structured result and every failure a
	// structured error, so a machine consumer parses fields instead of prose. Reads keep
	// their own format contract (`show --view`), which is view-selected, not flag-selected.
	root.PersistentFlags().Bool(flags.JSON, false, "emit a structured JSON result (and structured errors) instead of human text")

	root.AddCommand(
		lens.NewCommand(),
		merge.NewCommand(),
		blue.NewCommand(),
		bench.NewCommand(),
	)
	return root
}

// Execute runs the CLI. Abort-safety is armed first: a seat killed mid-command
// must lose an event, never leave a torn one or a stuck lock.
func Execute() {
	defer record.InstallSignalGuard()()

	if err := newRoot().Execute(); err != nil {
		// Errors print bare, without cobra's usage dump: a validation refusal here
		// is a TEACHING message a seat reads and acts on, and burying it under a
		// flag listing is how it stops being read.
		fmt.Fprintf(os.Stderr, "feov-record: %v\n", err)
		os.Exit(2)
	}
}
