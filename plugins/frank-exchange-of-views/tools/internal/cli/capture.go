package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
)

// newCapture is the operator command that runs /research's post-hoc capture step — ported from
// capture-research-run.mjs. Like setup/cost/scorecard/dashboard it is not a seat verb; /research's
// capture step spawns it (the record binary IS the tool — no --bin) to write run-record-audit.md,
// cost.md, the transcript tarball, the appended scorecards, and the precedent harvest, and to
// remove the run-live marker. The three record-backed audits read the record IN-PROCESS. Exit 2
// iff any audit FAILs (integrity findings are never smoothed over).
func newCapture() *cobra.Command {
	c := &cobra.Command{
		Use:           "capture <runDir> <transcriptDir>",
		Short:         "run /research's post-hoc capture auditor (operator; nine audits + cost.md/tarball/scorecards)",
		Long:          "capture is the mechanical half of /research's run-record step plus the mechanized post-hoc auditor: it copies the journal, packs the agent transcripts, writes cost.md, runs the nine integrity audits (recomputing counts from the git-tracked files and diffing them against the envelopes' self-reports), appends each chair's scorecard, harvests judicial rulings into law/proposed/, and removes the run-live marker — writing run-record-audit.md and exiting 2 on any audit FAIL. The record-backed audits read the record in-process. Ported from capture-research-run.mjs.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				fmt.Fprintln(cmd.ErrOrStderr(), "usage: feov-record capture <runDir> <workflow-transcript-dir>")
				os.Exit(1)
			}
			audits, report, exitFail, err := capture.Run(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), report)
			_ = audits
			if exitFail {
				os.Exit(2)
			}
			return nil
		},
	}
	return c
}
