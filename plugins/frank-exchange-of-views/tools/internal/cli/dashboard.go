package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/dashboard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// newDashboard is the operator command that renders a run's live dashboard.html — ported from
// render-run-dashboard.mjs. Like verify/graph/setup/cost/scorecard it is not a seat verb;
// /research's monitoring step runs it (the record binary IS the tool, so no --bin). It reads
// the record IN-PROCESS plus the run's transcripts/telemetry/scorecards and writes
// <runDir>/dashboard.html. --watch regenerates every 15s, keyed to the .run-live marker.
func newDashboard() *cobra.Command {
	var watch bool
	var now int64
	var model, judgmentModel, maxRounds, lanes string
	c := &cobra.Command{
		Use:           "dashboard <runDir> <transcriptDir>",
		Short:         "render a run's live dashboard.html (operator; board/cost/seats/judiciary/scorecards)",
		Long:          "dashboard renders the run's own instruments — board-mass trend, open/close rates, live seats + completion ETA, cost-so-far, judiciary analytics, friction, and this run's scorecards — into <runDir>/dashboard.html (meta-refreshes every 20s). Reads the record in-process plus the Workflow transcripts. --watch regenerates every 15s until the .run-live marker is removed. Ported from render-run-dashboard.mjs.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				fmt.Fprintln(cmd.ErrOrStderr(), "usage: feov-record dashboard <runDir> <workflow-transcript-dir> [--watch] [--model M] [--judgment-model M] [--max-rounds N] [--lanes N]")
				os.Exit(1)
			}
			runDir, transcriptDir := args[0], args[1]
			cfg := dashboard.Config{Model: model, JudgmentModel: judgmentModel, MaxRounds: maxRounds, Lanes: lanes}
			clock := func() float64 {
				if now != 0 {
					return float64(now)
				}
				return float64(time.Now().UnixMilli())
			}
			out := filepath.Join(runDir, "dashboard.html")
			generate := func() error {
				html := dashboard.RenderHTML(dashboard.BuildModel(runDir, transcriptDir, cfg, clock()))
				return os.WriteFile(out, []byte(html), 0o644)
			}
			if err := generate(); err != nil {
				return err
			}
			if !watch {
				fmt.Fprintln(cmd.OutOrStdout(), "dashboard:", out, "(static snapshot — re-run or use --watch to refresh)")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "dashboard:", out, "(watching — regenerates every 15s until the run-live marker is gone)")
			// Marker-keyed lifetime: run-capture removes .claude/run-live.json when the run ends;
			// on its disappearance, one final render and exit.
			marker := filepath.Join(".claude", "run-live.json")
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				_ = generate()
				if _, err := os.Stat(marker); err != nil {
					_ = generate()
					fmt.Fprintln(cmd.OutOrStdout(), "run-live marker gone — final render written, watcher exiting")
					return nil
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&watch, flags.Watch, false, "regenerate every 15s until the run-live marker is removed")
	c.Flags().Int64Var(&now, flags.Now, 0, "inject the clock in unix-ms (default: real time) — for deterministic tests")
	c.Flags().StringVar(&model, flags.Model, "", "bulk-tier model, for the run-config header (else read from inputs/run-config.json)")
	c.Flags().StringVar(&judgmentModel, flags.JudgmentModel, "", "judgment-tier model, for the run-config header")
	c.Flags().StringVar(&maxRounds, flags.MaxRounds, "", "round ceiling, for the progress bar segmentation")
	c.Flags().StringVar(&lanes, flags.Lanes, "", "blue lane count, for the run-config header")
	return c
}
