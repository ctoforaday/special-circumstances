package lens

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
)

// reproduce: red re-RUNS a proof instead of re-reading it.
//
// THIS IS THE STRONGEST AUDIT THE ENGINE HAS. Every other check ends in reading: a cited
// source is fetched and read, a claim is compared against prose, an anchor is located. All of
// them end with the auditor believing bytes someone else chose. A proof ends with red
// executing the same program and comparing the result — the one place the audit does not
// depend on trusting the audited.
//
// It is the direct answer to the smoke's R2-2, where blue asserted a test had happened and
// red could only observe that no evidence was shown. With a proof there is nothing to assert:
// the script is on disk, and either it still says what blue recorded or it does not.
//
// RECORDS NOTHING. Like near-match and claim-index it is a read — what red does with the
// result is a finding, in red's own words, through the finding verb.
func newReproduce() *cobra.Command {
	c := seat.New("reproduce",
		`re-RUN a proof blue recorded and compare: --id <sha256 of the proof> — read-only, records nothing (a disagreement is a finding you write yourself)`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			sha := seat.Str(cmd, flags.ID)
			if sha == "" {
				return nil, fmt.Errorf("lens reproduce requires --id: the sha256 of the proof to re-run (blue prove records it, and `show --view proofs` lists them)")
			}
			ok, got, want, err := proof.Reproduce(s.RunDir, sha)
			if err != nil {
				return nil, err
			}
			return reproduceResult{SHA: sha, Matches: ok, Got: got, Recorded: want}, nil
		})
	c.Flags().String(flags.ID, "", "REQUIRED — the sha256 of the recorded proof to re-run")
	return c
}

type reproduceResult struct {
	SHA      string `json:"sha256"`
	Matches  bool   `json:"matches"`
	Got      string `json:"got"`
	Recorded string `json:"recorded"`
}

func (r reproduceResult) Human() string {
	if r.Matches {
		return "proof " + r.SHA[:12] + " REPRODUCES — re-running it gives the recorded output byte for byte"
	}
	return "proof " + r.SHA[:12] + " DOES NOT REPRODUCE.\n  recorded: " + r.Recorded + "\n  now:      " + r.Got +
		"\n  Either the computation depends on something that moved, or the recorded output is not what the script produces. Both are findings; neither is recorded for you."
}
