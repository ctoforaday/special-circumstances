package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// newCountClaims computes claim_count from blue's report deterministically. Like
// verify/graph it is a root, read-only operator command, not a seat verb — it
// takes no seat identity because counting a file is not an act on the record. It
// serves two readers: the blue agent runs it live and relays the number into its
// envelope (replacing a hand-count that diverged 2x between honest merges), and
// capture recomputes it independently to arm the retire-vs-drop detector without
// trusting the relayed figure. See internal/claimcount for the counting rule.
func newCountClaims() *cobra.Command {
	c := &cobra.Command{
		Use:           "count-claims",
		Short:         "count the FOOTNOTED declarative claims in blue's report (read-only)",
		Long:          "count-claims reads <run>/blue/report.md and prints its claim_count — the number of footnoted declarative claims, computed deterministically (see the rule in internal/claimcount). It writes nothing. Used live by blue to size its envelope figure and by capture to recompute it independently.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			sc := seat.Of(cmd)
			// seat.Of ALREADY inferred if nothing was supplied — it ends its own resolution with
			// InferRunDir. What used to sit here was a SECOND inference, reached only when the
			// first had produced nothing, which by then meant the run had been REFUSED rather
			// than omitted. Inferring there resolved quietly to a different run than the one the
			// seat was told it could not have. Run() returns that refusal instead.
			run, err := sc.Run()
			if err != nil {
				return err
			}
			runDir := run.Dir()
			path := filepath.Join(runDir, "blue", "report.md")
			md, err := os.ReadFile(path)
			if err != nil {
				return feov.Errorf(feov.MissingField, "count-claims: cannot read %s: %v", path, err)
			}
			n := claimcount.Count(string(md))

			if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					ClaimCount int `json:"claim_count"`
				}{ClaimCount: n})
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), n)
			return nil
		},
	}
	return c
}
