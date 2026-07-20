package bench

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// newAssemble builds `feov-record bench assemble` — the report assembler moved into the tool.
//
// It is a run-level operation like render, not a board act, so it is a plain command (no
// seat context, no envelope): the sections are copied verbatim, the risk matrix is generated
// from the board, and the verdict is stamped, all mechanically. The assemble seat runs this
// and supplies only its judgment — the TL;DR and the synopsis — through --inputs.
func newAssemble() *cobra.Command {
	c := &cobra.Command{
		Use:          "assemble",
		Short:        "assemble <run>/report.md by union — verbatim sections + board risk matrix + verdict stamp; the seat hands in only the TL;DR and synopsis via --inputs",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.Flags().String(flags.Inputs, "", "path to a JSON file of the seat's inputs (verdict, tldr, synopsis, open_questions, ...)")
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		runDir := seat.Str(cmd, flags.Run)
		if runDir == "" {
			return fmt.Errorf("bench assemble: --run <runDir> is required")
		}
		var in report.Inputs
		if path := seat.Str(cmd, flags.Inputs); path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("bench assemble: reading --inputs: %w", err)
			}
			if err := json.Unmarshal(b, &in); err != nil {
				return fmt.Errorf("bench assemble: parsing --inputs JSON: %w", err)
			}
		}
		path, err := report.Assemble(runDir, in)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "feov-record bench: assembled %s (verdict %s)\n", path, in.Verdict)
		return nil
	}
	return c
}
