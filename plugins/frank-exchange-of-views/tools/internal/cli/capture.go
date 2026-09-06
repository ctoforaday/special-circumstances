package cli

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
			// RETURNED, NOT EXITED — see newScorecard. Same 1 -> 2 move, same reason.
			if len(args) < 2 {
				return feov.Errorf(feov.MissingField, "usage: %s capture <run-dir> <workflow-transcript-dir>", InvokedAs())
			}
			// A POSITIONAL PATH IS THE EASIEST ONE TO GET WRONG — nothing injects it and nothing
			// checks it. Opened rather than passed through, so a mistyped directory is refused
			// instead of audited as a run that recorded nothing.
			run, err := record.OpenRun(args[0])
			if err != nil {
				return err
			}
			audits, report, exitFail, err := capture.Run(run, args[1], time.Now())
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
