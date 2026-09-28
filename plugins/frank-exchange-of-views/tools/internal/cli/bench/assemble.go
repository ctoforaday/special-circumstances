package bench

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// newAssemble builds `feov-record bench assemble` — the report assembler.
//
// It is a run-level operation like render, not a board act, so it is a plain command (no
// seat context, no envelope). It reads the record and blue/report.md and writes report.md:
// blue's synthesis surfaces (title, TL;DR, Catechism, foundations, analysis, open questions)
// are lifted verbatim; the verdict, risk matrix, expansions, alternatives, findings, and the
// debate transcript are composed from the event log. It takes NO inputs — everything it
// needs is in the record (the verdict via the terminal `bench outcome` event) or in blue's
// audited report. The seat's whole job is to run it; what it prints is the confirmation.
func newAssemble() *cobra.Command {
	c := &cobra.Command{
		Use:          "assemble",
		Short:        "assemble the run's documents for the human reader, from the record — blue's audited sections lifted verbatim, the rest composed; no inputs. It prints the verdict it stamped: the documents are the human's, and a seat reads the report with `show report`",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		// Resolved, so the injected run reaches this read as it does every write.
		run, rerr := seat.Of(cmd).RequireRun("bench assemble")
		if rerr != nil {
			return rerr
		}
		a, err := report.Assemble(run)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "feov-record bench: assembled %d documents for the human reader, the verdict stamped %s from the outcome on the record. "+
			"They are not for a seat to open — the report is read with `show report`.\n", a.Documents, a.Verdict)
		return nil
	}
	return c
}
