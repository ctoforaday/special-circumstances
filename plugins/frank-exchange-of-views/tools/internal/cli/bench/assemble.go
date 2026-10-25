package bench

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// newAssemble builds `feov-record bench assemble` — the report assembler.
//
// It is a run-level operation like render, not a board act, so it writes no event and renders no
// envelope — but it WRITES the run's documents, so it goes through seat.Begin like every writer,
// and then asks the record that this is the bench's assemble sitting (record.RequireAssemblySitting):
// the bench seat the agent registered as, on the run it was dispatched into. It reads the record and blue/report.md and writes report.md:
// blue's text (title, TL;DR, Catechism, the report's own sections, open questions)
// ships verbatim; the verdict, risk matrix, expansions, alternatives, findings, and the
// debate transcript are composed from the event log. It takes NO inputs — everything it
// needs is in the record (the verdict via the terminal `bench outcome` event) or in blue's
// audited report. The seat's whole job is to run it; what it prints is the confirmation.
func newAssemble() *cobra.Command {
	c := &cobra.Command{
		Use:          "assemble",
		Short:        "assemble the run's documents for the human reader, from the record, in the sitting you registered for the assemble occasion and no other — the report's audited text as it is written, the rest composed; no inputs. It prints the verdict it stamped: the documents are the human's, and a seat reads the report with `show report`",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		s, berr := seat.Begin(cmd)
		if berr != nil {
			return berr
		}
		run, rerr := s.Run()
		if rerr != nil {
			return rerr
		}
		if err := record.RequireAssemblySitting(run, s.SeatID); err != nil {
			return err
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
