package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// newFetch is the cached web read that REPLACES WebFetch for every seat.
//
// Like verify/count-claims it is a root operator command, not a seat verb: reading a URL
// is not an act on the record, so it takes no seat identity and writes no event. What it
// DOES own is the run's source cache — fetch a URL once, cache the bytes at
// <run>/cache/<sha256>, and serve every later read (blue evaluating a source, red
// re-verifying a cited one) those exact bytes. A second fetch of the same URL is a cache
// hit, so both sides reason about identical content and red never re-downloads what blue
// cited.
//
// A fetch FAILURE is a plain non-zero error: a bare read may legitimately miss and the
// caller just picks another source. It does NOT itself write a friction event — only the
// DECISION to cite an unreachable source does (blue cite), because that is the act that
// records an unusable citation.
func newFetch() *cobra.Command {
	c := &cobra.Command{
		Use:           "fetch",
		Short:         "cached, hash-verified web read (replaces WebFetch); serves both sides the same bytes",
		Long:          "fetch GETs --url once, caches the bytes at <run>/cache/<sha256>, and prints them; a later fetch of the same URL is served from cache so every seat reads identical content. It writes no record event. A fetch failure is a non-zero error (pick another source); it does not itself log friction.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			sc := seat.Of(cmd)
			// BEFORE the inference fallback, never after: falling back on a REFUSED resolution
			// resolves quietly to the real run and hides the contradiction the seat needs to see.
			if sc.RunErr != nil {
				return sc.RunErr
			}
			runDir := sc.RunDir
			if runDir == "" {
				runDir = seat.InferRunDir("")
			}
			if runDir == "" {
				return feov.Errorf(feov.MissingField, "fetch: --run <runDir> is required")
			}
			url, _ := cmd.Flags().GetString(flags.URL)
			if url == "" {
				return feov.Errorf(feov.MissingField, "fetch: --url <url> is required")
			}
			sha, body, hit, err := fetchcache.Resolve(runDir, url, fetchcache.Default)
			if err != nil {
				// A bare read failure is operational, not a seat-input fault — no existing
				// coded category fits, and it is NOT a friction (only the DECISION to cite an
				// unreachable source is). CodeOf renders a plain error as "error" in --json.
				return fmt.Errorf("fetch: %v", err)
			}
			if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					URL      string `json:"url"`
					Sha256   string `json:"sha256"`
					CacheHit bool   `json:"cache_hit"`
					Bytes    int    `json:"bytes"`
					Body     string `json:"body"`
				}{URL: url, Sha256: sha, CacheHit: hit, Bytes: len(body), Body: string(body)})
			}
			fmt.Fprint(cmd.OutOrStdout(), string(body))
			return nil
		},
	}
	c.Flags().String(flags.URL, "", "the http/https URL to read (fetched once, then served from the run cache)")
	return c
}
