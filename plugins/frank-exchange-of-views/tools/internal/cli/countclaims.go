package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// newCountClaims counts the report's claims deterministically. Like verify/graph it is a root,
// read-only operator command, not a seat verb — it takes no seat identity because counting the
// report is not an act on the record. IT IS THE ONE DEFINITION OF THE COUNT: no seat relays the
// number and no envelope carries it, so a reader that wants it — the dashboard's tile, the mint
// budget — counts the report on the record through internal/claimcount, as this does.
func newCountClaims() *cobra.Command {
	c := &cobra.Command{
		Use:           "count-claims",
		Short:         "count the FOOTNOTED declarative claims in blue's report (read-only)",
		Long:          "count-claims prints the report's claim_count — the number of footnoted declarative claims, computed deterministically from the report on the record. It writes nothing, and it is the count's one definition: no seat relays the number.",
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
			md, err := reportproj.RenderFromRecord(run)
			if err != nil {
				return feov.Errorf(feov.MissingField, "count-claims: cannot read the report: %v", err)
			}
			n := claimcount.Count(md)

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
