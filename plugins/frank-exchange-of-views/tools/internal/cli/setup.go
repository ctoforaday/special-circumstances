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
// launch fails here rather than mid-round.
//
// It owns its exit codes (no runDir → 1; a refused gate → 2), so it runs setup.Run and
// exits directly rather than returning an error through the root's uniform exit-2 path.
func newSetup() *cobra.Command {
	var (
		topic, model, judgmentModel string
		maxRounds, lanes            string
		binDir, memoryDir           string
		cites                       []string
	)
	c := &cobra.Command{
		Use:           "setup <runDir>",
		Short:         "build a research run's blackboard: skeleton, pins, memory mirrors, and the .run-live marker (operator; writes files)",
		Long:          "setup creates <runDir>'s blackboard skeleton (idempotent — pre-staged files are kept), pins the evidence base at HEAD, mirrors red's gap-pattern + law + scorecard memory into inputs/, writes the .run-live marker hook guards consult, and preflights the record binary. The four fail-fast gates (a runDir, both model tiers, valid cites, and — with --bin-dir — a matching record binary) run BEFORE any state is created, so a bad launch costs a re-run, not a seat mid-round.",
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
				RunDir:        runDir,
				Topic:         topic,
				Model:         model,
				JudgmentModel: judgmentModel,
				Cites:         cites,
				MaxRounds:     maxRounds,
				Lanes:         lanes,
				BinDir:        binDir,
				MemoryDir:     memoryDir,
				Cwd:           cwd,
				Home:          home,
				ProjectDir:    os.Getenv("CLAUDE_PROJECT_DIR"),
				ExpectVersion: Version,
			}
			if code := setup.Run(cfg, cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&topic, flags.Topic, "", "the run's research topic (goes in every stub header)")
	f.StringVar(&model, flags.Model, "", "REQUIRED — the bulk tier (frontier, blue lanes, red lenses, blue responses)")
	f.StringVar(&judgmentModel, flags.JudgmentModel, "", "REQUIRED — the judgment tier (blue-synthesize, red-merge, judge, assemble)")
	f.StringArrayVar(&cites, flags.Cite, nil, "a cited path, optionally pinned: <path>[@<commit>] (repeatable)")
	f.StringVar(&maxRounds, flags.MaxRounds, "", "the round ceiling (recorded in run-config.json for post-hoc readers)")
	f.StringVar(&lanes, flags.Lanes, "", "the frontier lane count (recorded in run-config.json)")
	f.StringVar(&binDir, flags.BinDir, "", "where the feov-record binary the SEATS will call lives (default: this executable's own directory); the version preflight always runs and always refuses on a miss")
	f.StringVar(&memoryDir, flags.MemoryDir, "", "override the gap-pattern memory source (default: promoted corpus, then raw accrual)")
	return c
}
