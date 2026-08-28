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
		Use:   "fetch",
		Short: "cached, hash-verified web read (replaces WebFetch); serves both sides the same bytes",
		Long: "fetch GETs --url once, caches the bytes at <run>/cache/<sha256>, and prints them; a later fetch of the same URL is served from cache so every seat reads identical content. It writes no record event. A fetch failure is a non-zero error (pick another source); it does not itself log friction. " +
			"AN UNREACHED SOURCE IS NOT EVIDENCE OF ABSENCE. Where this session runs behind an egress proxy, a host outside its allowlist answers 403 — the same status an origin uses to refuse a client — so a failure can be a fact about THIS CONTAINER or a fact about the SOURCE, and the two are different findings. The refusal says so where it can. Measured 2026-08-23: openai.com 403s through the proxy, and a research question shipped `open rather than resolved` on that basis. Where you cannot tell which it was, record that the source was UNREACHABLE FROM HERE rather than that the question is unresolved.",
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
			url, _ := cmd.Flags().GetString(flags.URL)
			if url == "" {
				return feov.Errorf(feov.MissingField, "fetch: --url <url> is required")
			}
			sha, body, hit, err := fetchcache.Resolve(run, url, fetchcache.Default)
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
