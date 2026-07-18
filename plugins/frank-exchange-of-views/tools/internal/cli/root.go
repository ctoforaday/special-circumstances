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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Version is stamped on register events and answered by --version. The setup
// preflight compares it against the plugin manifest BEFORE the run-live marker
// is written, so a skewed binary fails at setup rather than mid-round.
const Version = "0.1.0"

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
	root.PersistentFlags().String("run", "", "the run directory (the engine passes it; it is in your prompt)")
	root.PersistentFlags().String("seat-id", "", "your seat id, assigned by the engine and bound to this role's namespace")

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
