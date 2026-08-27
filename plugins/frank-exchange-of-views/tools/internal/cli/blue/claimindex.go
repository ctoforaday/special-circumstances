package blue

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// newClaimIndex enumerates where each footnoted claim appears in blue's report —
// the claim-index. Like count-claims it is READ-ONLY and takes no seat identity:
// reading a file is not an act on the record. It is a blue verb (not a root command)
// because the report is blue's and blue is its only reader, discovered through
// `blue --help`.
//
// It exists to make propagation cheap. When blue corrects a footnoted claim it must
// reach EVERY site that states it; the index answers "where is claim [^L7]?" from a
// single query instead of re-reading the whole report to hunt the sites. It does NOT
// replace the report-wide string/figure sweep: an unfootnoted restatement of a
// corrected figure is invisible to a marker index, so a corrected number still needs
// the grep. The index is the footnoted-site locator; the sweep is the completeness
// backstop.
func newClaimIndex() *cobra.Command {
	return &cobra.Command{
		Use:   "claim-index",
		Short: "locate every site of each FOOTNOTED claim in your report (read-only)",
		Long: "claim-index reads <run>/blue/report.md and prints, per footnote label, every site " +
			"that claim appears: {label, occurrences:[{heading, line}]}. WHEN to use " +
			"it: you are correcting a claim and must propagate the fix to ALL its sites — query the " +
			"label for its occurrences instead of re-reading the whole report to find them. WHAT it " +
			"is NOT: a replacement for the report-wide string/figure sweep — an unfootnoted " +
			"restatement (a bare corrected FIGURE) is invisible to a footnote-marker index, so still " +
			"grep the corrected strings report-wide. It records nothing. JSON only.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// THE REFUSAL IS NOT AN ABSENCE, and this verb used to treat it as one.
			//
			// It read the path straight off seat.Of — where there is no error to consult even in
			// principle — and answered "" by inferring from the marker. So a run that had been
			// REFUSED fell through to inference and resolved quietly to a different one: the
			// exact swallow seat.Context's own comment predicts, "" meaning both "nobody supplied
			// a run" and "you were refused", with the healthy reading winning by default.
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			runDir := run.Dir()
			path := filepath.Join(runDir, "blue", "report.md")
			md, err := os.ReadFile(path)
			if err != nil {
				return feov.Errorf(feov.MissingField, "claim-index: cannot read %s: %v", path, err)
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(struct {
				Claims []claimcount.LabelOccurrences `json:"claims"`
			}{Claims: claimcount.Index(string(md))})
		},
	}
}
