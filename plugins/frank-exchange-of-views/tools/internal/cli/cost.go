package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
)

// newCost is the operator command that renders a run's cost.md from the Workflow transcripts —
// ported from cost-audit.mjs. Like verify/graph/setup it is not a seat verb; /research's capture
// step spawns it (via the record binary) to write cost.md. It reads transcripts + the run's
// run-config/telemetry, never the event record, and prints the report to stdout.
func newCost() *cobra.Command {
	c := &cobra.Command{
		Use:           "cost <transcriptDir> [runDir]",
		Short:         "render a run's cost.md from the Workflow transcripts (operator; per seat-round tokens + dollars)",
		Long:          "cost parses the Workflow harness's per-agent transcripts in <transcriptDir> and prints the cost.md report — per seat-round tokens and dollars (list-rate), totals, and cache share. With <runDir> it appends the #111 tier check (each seat's actual price-tier vs the tier its class was configured for, from inputs/run-config.json) and the per-round board-telemetry join. Ported from cost-audit.mjs.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				fmt.Fprintln(cmd.ErrOrStderr(), "usage: feov-record cost <workflow-transcript-dir> [runDir]")
				os.Exit(1)
			}
			runDir := ""
			if len(args) > 1 {
				runDir = args[1]
			}
			return cost.Report(args[0], runDir, cmd.OutOrStdout())
		},
	}
	return c
}
