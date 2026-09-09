package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/setup"
)

// newSetup is the operator command that builds a research run's blackboard — the
// mechanical half of /research steps 1-3, ported from setup-research-run.mjs. Like
// verify/graph it is not a seat verb: no debating role runs it, so it lives at the
// root. It writes files (skeleton, pins, mirrors, the .run-live marker) and preflights
// the record binary BEFORE any run state exists, so a model-less or bad-cite
// launch fails here rather than mid-run.
//
// It owns its exit codes (no runDir → 1; a refused gate → 2), so it runs setup.Run and
// exits directly rather than returning an error through the root's uniform exit-2 path.
func newSetup() *cobra.Command {
	var (
		topic, model, judgmentModel string
		lanes                       string
		k, kMax, mintBudget         int
		convergenceFraction         float64
		lensAreas                   []string
		binDir, memoryDir           string
		runID, scriptPath           string
		cites                       []string
		allowSubstitution           bool
	)
	c := &cobra.Command{
		Use:           "setup <runDir>",
		Short:         "build a research run's blackboard: skeleton, pins, memory mirrors, and the .run-live marker (operator; writes files)",
		Long:          "setup creates <runDir>'s blackboard skeleton (idempotent — pre-staged files are kept), pins the evidence base at HEAD, mirrors red's gap-pattern + law + scorecard memory into inputs/, writes the .run-live marker hook guards consult, and preflights the record binary. The four fail-fast gates (a runDir, both model tiers, valid cites, and — with --bin-dir — a matching record binary) run BEFORE any state is created, so a bad launch costs a re-run, not a seat mid-sitting.",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			runDir := ""
			if len(args) > 0 {
				runDir = args[0]
			}
			cwd, _ := os.Getwd()
			home, _ := os.UserHomeDir()
			cfg := setup.Config{
				RunDir:              runDir,
				Topic:               topic,
				Model:               model,
				JudgmentModel:       judgmentModel,
				Cites:               cites,
				Lanes:               lanes,
				K:                   k,
				KMax:                kMax,
				MintBudget:          mintBudget,
				ConvergenceFraction: convergenceFraction,
				LensAreas:           lensAreas,
				BinDir:              binDir,
				MemoryDir:           memoryDir,
				RunID:               runID,
				ScriptPath:          scriptPath,
				AllowSubstitution:   allowSubstitution,
				Cwd:                 cwd,
				Home:                home,
				ProjectDir:          os.Getenv("CLAUDE_PROJECT_DIR"),
			}
			if code := setup.Run(cfg, cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&topic, flags.Topic, "", "the run's research topic (goes in every stub header)")
	f.StringVar(&model, flags.Model, "", "the bulk tier (frontier, blue lanes, red lenses, blue responses)")
	f.StringVar(&judgmentModel, flags.JudgmentModel, "", "the judgment tier (blue-synthesize, red-chair, judge, assemble)")
	f.StringArrayVar(&cites, flags.Cite, nil, "a cited path, optionally pinned: <path>[@<commit>] (repeatable)")
	f.StringVar(&lanes, flags.Lanes, "", "the frontier lane count (recorded in run-config.json)")
	f.IntVar(&k, flags.K, 0, "consecutive exchanges on one gap with no movement before it is at impasse (default 2; recorded in run-config.json)")
	f.IntVar(&kMax, flags.KMax, 0, "total exchanges on one gap before it is at impasse regardless of movement (default 6; a smoke run passes 2)")
	f.IntVar(&mintBudget, flags.MintBudget, 0, "gaps one lens may mint in the run, superseding mints included (default 5; a smoke run passes 1)")
	f.StringArrayVar(&lensAreas, flags.LensArea, nil, "a red lens area this run dispatches (repeatable; default evidence, logic, dark-side, voice) — the cast is written from these")
	f.Float64Var(&convergenceFraction, flags.ConvergenceFraction, 0, "fraction of the run's peak board mass below which a FAIL over a board with nothing material is refused (default 0.25)")
	f.StringVar(&binDir, flags.BinDir, "", "where the feov-record binary the SEATS will call lives (default: this executable's own directory); the version preflight always runs and always refuses on a miss")
	f.StringVar(&memoryDir, flags.MemoryDir, "", "override the gap-pattern memory source (default: promoted corpus, then raw accrual)")
	f.BoolVar(&allowSubstitution, flags.AllowSubstitution, false, "accept a run whose environment answers with a model other than the configured tier — recorded on the run, so every seat's register stops refusing it and the substitution stays visible on the record")
	// THE MARKER OUTLIVES THE WORKFLOW THAT WROTE IT, so it has to name how to continue.
	// A workflow killed by an idle SIGTERM never lifts .claude/run-live.json — it cannot, it is
	// gone — and a marker naming only a directory tells a later reader that something WAS running
	// here, not what to do about it. Optional because a launcher may not know them; absent stays
	// absent rather than becoming an empty string on the record.
	f.StringVar(&runID, flags.RunID, "", "the orchestrator's run id, recorded in the .run-live marker so a run found STALE can be resumed rather than only noticed")
	f.StringVar(&scriptPath, flags.ScriptPath, "", "the orchestrator script this run was launched from, recorded in the .run-live marker beside --run-id")
	return c
}
