package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// newMigrate is the operator's replay-as-migration: an old record's events re-driven, in
// order, through THIS binary's write path into a fresh sibling run. The output is a record
// this binary could have written — validated per act, current schema and views — and the
// original is never modified. See plans/replay-migration.md.
func newMigrate() *cobra.Command {
	var from, to string
	var acceptLoss []string
	c := &cobra.Command{
		Use:   "migrate --from <oldRunDir> --to <newRunDir>",
		Short: "replay an old record through this binary's write path into a fresh run (the source is never modified)",
		Long: "migrate reads an old run's record — including a vocabulary this binary's readers refuse — translates " +
			"each event through the authored registry, and replays them in order through the current write path under " +
			"the ORIGINAL clock. Every act is validated exactly as a seat's would be, so the output is a record this " +
			"binary could have written; run channels (inputs, proofs) copy verbatim, and inputs/migration.json records " +
			"what was translated, refused, and accepted as loss. A refusal exits non-zero and is the tool's OUTPUT — " +
			"there is no flag that skips validation.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if from == "" || to == "" {
				return fmt.Errorf("migrate: --from and --to are both required — the source run, and the fresh directory the migration writes")
			}
			accept := map[string]bool{}
			for _, w := range acceptLoss {
				accept[w] = true
			}
			m, err := migrate.Migrate(from, to, migrate.Entries(), migrate.Options{AcceptLoss: accept})
			jsonMode, _ := cmd.Flags().GetBool(flags.JSON)
			if jsonMode && m != nil {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(m)
			} else if m != nil {
				printManifest(cmd, m)
			}
			return err
		},
	}
	c.Flags().StringVar(&from, flags.MigrateFrom, "", "the old run directory (its records/record.db file set is read from a copy)")
	c.Flags().StringVar(&to, flags.MigrateTo, "", "the fresh run directory to write — refused if it already holds anything")
	c.Flags().StringSliceVar(&acceptLoss, flags.AcceptLoss, nil,
		"accept a registry-documented loss for exactly this word (repeatable); the acceptance lands in the manifest")
	return c
}

func printManifest(cmd *cobra.Command, m *migrate.Manifest) {
	out := cmd.OutOrStdout()
	in, outN := 0, 0
	for _, n := range m.In {
		in += n
	}
	for _, n := range m.Out {
		outN += n
	}
	fmt.Fprintf(out, "migrated %s\n  %d event(s) in -> %d out; manifest at inputs/%s\n", m.SourcePath, in, outN, migrate.ManifestName)
	words := make([]string, 0, len(m.In))
	for w := range m.In {
		words = append(words, w)
	}
	sort.Strings(words)
	for _, w := range words {
		if m.Out[w] != m.In[w] {
			fmt.Fprintf(out, "  %s: %d -> %d\n", w, m.In[w], m.Out[w])
		}
	}
	for w, why := range m.Accepted {
		fmt.Fprintf(out, "  accepted loss of %q: %s\n", w, why)
	}
	if len(m.Refusals) > 0 {
		fmt.Fprintf(out, "  %d REFUSAL(S) — each is in the manifest, and each is a decision to make, not a row to ignore\n", len(m.Refusals))
	}
}
