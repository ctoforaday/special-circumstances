package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/graph"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// newGraph renders a run's actual behaviour — replayed from its record — as a graph. Like
// verify, it is an operator tool (not a seat verb), so it lives at the root. It emits text: the
// tool speaks the graph, rendering is free downstream (mermaid in an artifact/PR, or DOT
// through graphviz). Read-only.
func newGraph() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:           "graph",
		Short:         "render a run's actual behaviour (seat flow + gap lifecycle) as mermaid or dot (read-only)",
		Long:          "graph replays <run>'s record and prints its shape — the seats that ran per round with their event tallies, and each gap's lifecycle coloured by state, with holes (a dispute nobody answered, a silent closure) highlighted. Default output is markdown with mermaid (renders in an artifact, a PR, or a phone); --format dot emits a graphviz digraph for a dense run.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			runDir := seat.Of(cmd).RunDir
			if runDir == "" {
				runDir = seat.InferRunDir("")
			}
			if runDir == "" {
				return feov.Errorf(feov.MissingField, "graph: --run <runDir> is required")
			}
			board, err := record.BoardState(runDir)
			if err != nil {
				return fmt.Errorf("graph: %w", err)
			}
			switch format {
			case "", "mermaid":
				fmt.Fprint(cmd.OutOrStdout(), graph.Mermaid(board))
			case "dot":
				fmt.Fprint(cmd.OutOrStdout(), graph.Dot(board))
			default:
				return feov.Errorf(feov.Validation, "graph: unknown --format %q (mermaid | dot)", format)
			}
			return nil
		},
	}
	c.Flags().StringVar(&format, flags.Format, "mermaid", "output format: mermaid (markdown, renders natively) | dot (graphviz)")
	return c
}
