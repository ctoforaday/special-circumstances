package cli

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/dashboard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
	var serve int
	var model, judgmentModel, maxRounds, lanes string
	c := &cobra.Command{
		Use:           "dashboard <runDir> <transcriptDir>",
		Short:         "render a run's live dashboard.html (operator; board/cost/seats/judiciary/scorecards)",
		Long:          "dashboard renders the run's own instruments — board-mass trend, open/close rates, live seats + completion ETA, cost-so-far, judiciary analytics, friction, and this run's scorecards — into <runDir>/dashboard.html (meta-refreshes every 20s). Reads the record in-process plus the Workflow transcripts. --watch regenerates every 15s until the .run-live marker is removed. --serve <port> instead HOSTS it over HTTPS behind a per-run secret URL (rendered fresh per request, self-tears-down when the run ends) so it is reachable across the local network. Ported from render-run-dashboard.mjs.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// RETURNED, NOT EXITED — see newScorecard for the whole argument. The exit status
			// moves 1 -> 2 here, which is the point rather than a side effect: every other
			// refusal in this tool is 2, and 1 was this one site disagreeing.
			if len(args) < 2 {
				return feov.Errorf(feov.MissingField, "usage: %s dashboard <run-dir> <workflow-transcript-dir> [--watch] [--model M] [--judgment-model M] [--max-rounds N] [--lanes N]", InvokedAs())
			}
			transcriptDir := args[1]
			// Same reasoning as capture: a positional run directory nobody validated renders a
			// dashboard of zeros for a path that was simply mistyped.
			run, err := record.OpenRun(args[0])
			if err != nil {
				return err
			}
			cfg := dashboard.Config{Model: model, JudgmentModel: judgmentModel, MaxRounds: maxRounds, Lanes: lanes}
			clock := func() float64 {
				if now != 0 {
					return float64(now)
				}
				return float64(time.Now().UnixMilli())
			}
			// --serve hosts the dashboard over HTTPS as an ephemeral, capability-gated,
			// self-tearing-down server (see serveDashboard): each request renders FRESH from the
			// record in-process — the same JIT discipline the projections follow — behind a
			// per-run secret URL, and the page's own 20s meta-refresh keeps a browser current.
			if serve != 0 {
				return serveDashboard(cmd.OutOrStdout(), run, serve, func() string {
					return dashboard.RenderHTML(dashboard.BuildModel(run, transcriptDir, cfg, clock()))
				})
			}

			out := filepath.Join(run.Dir(), "dashboard.html")
			generate := func() error {
				html := dashboard.RenderHTML(dashboard.BuildModel(run, transcriptDir, cfg, clock()))
				return os.WriteFile(out, []byte(html), 0o644)
			}
			if err := generate(); err != nil {
				return err
			}
			if !watch {
				fmt.Fprintln(cmd.OutOrStdout(), "dashboard:", out, "(static snapshot — re-run or use --watch to refresh)")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "dashboard:", out, "(watching — regenerates every 15s until the run ends)")
			// TWO SIGNALS, because the marker alone was not enough (#270). `capture` is the only
			// thing that removes .claude/run-live.json, and capture is optional: a run that was
			// killed or never captured left this watcher regenerating a dead run forever. The
			// RECORD is the second signal and the more truthful one — once the bench records the
			// outcome, the run is over whatever the filesystem still says.
			marker := filepath.Join(".claude", "run-live.json")
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			// THE TICK IS THE SCHEDULE, NOT THE REASON. Every 15s this called BuildModel
			// unconditionally, which replays the record AND re-reads and re-parses every
			// agent-*.jsonl in the transcript directory — append-only files that GROW through the
			// run, so the cost climbs while the work stays the same. ~2s of CPU per tick on a long
			// run, ~1,200 ticks over five hours: ~40 minutes of CPU re-deriving a page that in most
			// windows nobody had given it a new fact for (#684 F15).
			//
			// The fingerprint answers "has anything BuildModel can read moved", and it fails
			// towards rendering: unmeasurable counts as changed (dashboard.Changed owns that
			// direction, so it is not a judgement made here). Skipping a tick is invisible to a
			// reader — the page it would have rewritten is byte-for-byte the one already on disk.
			seen := dashboard.Fingerprint(run.Dir(), transcriptDir)
			for range ticker.C {
				// `current` and not `now`: `now` is the injected-clock flag in this scope, and
				// shadowing it here would compile and read as the clock to everyone after.
				current := dashboard.Fingerprint(run.Dir(), transcriptDir)
				if dashboard.Changed(seen, current) {
					_ = generate()
					// RECORDED FROM BEFORE THE RENDER, deliberately. Re-measuring afterwards
					// would absorb anything a seat wrote WHILE the page was being built and call
					// it already-seen, so that write would wait for the next one to arrive
					// before it showed up. Carrying the earlier reading costs at most one
					// redundant render and cannot lose an update.
					seen = current
				}
				if ended, why := runHasEnded(marker, run); ended {
					_ = generate()
					fmt.Fprintln(cmd.OutOrStdout(), why+" — final render written, watcher exiting")
					return nil
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&watch, flags.Watch, false, "regenerate every 15s until the run-live marker is removed")
	// The old help said "over HTTP … UNAUTHENTICATED", which the code has never done: serveDashboard
	// runs ListenAndServeTLS with a TLS 1.2 floor and gates every request on a random per-run secret
	// in the URL path. Two co-resident statements of one contract disagreed — and on a SECURITY
	// property, so the wrong one is the kind of thing that stops a careful reader using the feature
	// (or, worse, makes a careless one think the exposure is already understood).
	c.Flags().IntVar(&serve, flags.Serve, 0, "host the dashboard over HTTPS on this port instead of writing a file — bound to 0.0.0.0 so it is reachable across the local network, gated by a per-run secret URL (self-signed cert, rendered fresh per request, self-tears-down when the run ends)")
	c.Flags().Int64Var(&now, flags.Now, 0, "inject the clock in unix-ms (default: real time) — for deterministic tests")
	c.Flags().StringVar(&model, flags.Model, "", "bulk-tier model, for the run-config header (else read from inputs/run-config.json)")
	c.Flags().StringVar(&judgmentModel, flags.JudgmentModel, "", "judgment-tier model, for the run-config header")
	c.Flags().StringVar(&maxRounds, flags.MaxRounds, "", "round ceiling, for the progress bar segmentation")
	c.Flags().StringVar(&lanes, flags.Lanes, "", "blue lane count, for the run-config header")
	return c
}
