package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

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
		// WHAT THE LONG NO LONGER SAYS, AND WHERE IT LIVES. The share of cited PDFs that are scans
		// (one in four, 2026-08-23 corpus) is at the OCR call below; the engine (tesseract +
		// leptonica, statically linked) and its 300 DPI grayscale render are in
		// fetchcache/pageread.go; the openai.com 403 that shipped a question "open rather than
		// resolved" (#592) is at fetchcache/httpfetcher.go and egress_test.go. None of it changes
		// what a seat does, so the page carries only the instruction each one produced.
		Long: "fetch GETs --url once, caches the bytes at <run>/cache/<sha256>, and prints A SUMMARY NAMING THE FILES — never the document itself. Open what you need with Read, using an offset and a limit: a 67-page paper pasted into your context is waste whether legible or not. A PDF's text is extracted to <run>/cache/<sha256>.txt and named in the summary, so you need no PDF tooling.\n\n" +
			"A PDF WITH NO TEXT LAYER (a scan) is read by the LOCAL OCR ENGINE compiled into this binary — deterministic, reproducible, no model, credentials or network, seconds per document — and named in the summary as ocr_derived text keyed to the engine identity, so an audit can re-derive it byte for byte. Ruled tables are reconstructed into |-separated rows with confidence stats on the record; one that cannot place its marks falls back to plain text WITH the failure stated. A document over the render disk budget is refused rather than partly read.\n\n" +
			"Pages already read are never re-derived: each carries a receipt checked against its image hash, so a retry or a crash resumes cleanly. A later fetch of the same URL is served from cache, so every seat reads identical bytes. It writes no record event. A failure is a non-zero error (pick another source) and logs no friction itself.\n\n" +
			"AN UNREACHED SOURCE IS NOT EVIDENCE OF ABSENCE: behind an egress proxy, a host outside the allowlist answers 403 — the status an origin uses to refuse a client — so a failure can be a fact about THIS CONTAINER or about the SOURCE, and those are different findings. The refusal says which where it can; where you cannot tell, record the source as UNREACHABLE FROM HERE, not the question as unresolved.",
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
			via, _ := cmd.Flags().GetString(flags.Via)
			at, _ := cmd.Flags().GetString(flags.At)
			// A NAMED BACKEND IS AN INSTRUCTION, NOT A FALLBACK. `--via archive` means read the
			// archive whether or not the live source would answer — which is how a seat asks
			// "what did this say in 2019" about a page that is perfectly reachable today.
			// REFUSED AT THE SITE, because a near-miss must not take the other branch silently.
			// `--via archiv` falling through to a live fetch would answer a question the seat did
			// not ask and look like success.
			if via != "" && !slices.Contains(fetchcache.Vias(), via) {
				return feov.Errorf(feov.Validation, "fetch: unknown --via %q (%s)", via, strings.Join(fetchcache.Vias(), " | "))
			}
			// --at WITHOUT A BACKEND THAT READS IT WAS A SILENT NO-OP. It is consumed only on the
			// non-live path (Recover takes it); a live fetch read the flag and dropped it, so a
			// seat asking "what did this say in 2019" with --at alone got TODAY'S page and a
			// success message. The date is the whole question there, and losing it silently
			// answers a different one — the same shape the --via near-miss refusal above exists
			// to close, one flag over.
			if at != "" && (via == "" || via == fetchcache.ViaLive) {
				return feov.Errorf(feov.Validation,
					"fetch: --at %s bounds an ARCHIVE capture and nothing else reads it — pass --via archive, "+
						"or drop --at. Alone it would have fetched the live source and silently ignored your date", at)
			}
			if via != "" && via != fetchcache.ViaLive {
				att := fetchcache.Recover(fetchcache.Default, url, via, at)
				if att == nil {
					return fmt.Errorf("fetch --via %s: that backend has no answer for %s. The strategies answer "+
						"different questions: try `metadata` to learn whether the source exists at all, or `live` for the source itself", via, url)
				}
				entry, serr := fetchcache.Store(run, fetchcache.Entry{
					URL: url, ContentType: att.ContentType,
					RetrievedVia: att.Via, TextRetrieved: att.TextRetrieved,
				}, att.Body)
				if serr != nil {
					return serr
				}
				sv := summarize(run, entry, len(att.Body), false)
				if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(sv)
				}
				fmt.Fprint(cmd.OutOrStdout(), sv.render())
				return nil
			}
			entry, body, hit, err := fetchcache.Resolve(run, url, fetchcache.Default)
			if err != nil {
				// A bare read failure is operational, not a seat-input fault — no existing
				// coded category fits, and it is NOT a friction (only the DECISION to cite an
				// unreachable source is). CodeOf renders a plain error as "error" in --json.
				return fmt.Errorf("fetch: %v", err)
			}
			s := summarize(run, entry, len(body), hit)

			// A SCANNED DOCUMENT IS READ HERE, rather than handed back as a dead end (#644).
			//
			// The reading is LOCAL — the OCR engine statically linked into this binary — and it
			// is deliberately not left to the seat: the alternative was an INSTRUCTION telling a
			// seat that `ocr pages` and `ocr read` exist, which is the weakest carrier available
			// for a step the tool can simply take. It fires only where the extractor looked at a
			// PDF and found no text layer — one document in four in the 2026-08-23 corpus — and
			// it is bounded by the render disk budget and reused across fetches of the same
			// document.
			//
			// A READ FAILURE IS NEVER A FETCH FAILURE, exactly as an extraction failure is not:
			// the bytes are cached and the source is perfectly good, it is the READING that is
			// missing. So the reason travels on the summary and the command still exits 0 — this
			// is also how a binary built WITHOUT the engine states itself: ocr_reason carries the
			// engine-absent sentence instead of an empty reading. What must not happen is the
			// reason travelling nowhere — a document with no text and no stated cause reads
			// identically whether reading was refused, off, or broken.
			if applicableToOCR(entry) {
				switch on, _ := cmd.Flags().GetBool(flags.OCR); {
				case !on:
					s.OCRReason = fmt.Sprintf("automatic reading is off (--%s=false); read it deliberately "+
						"with `ocr pages --%s %s` then `ocr read --%s %s`",
						flags.OCR, flags.Sha, entry.Sha, flags.Sha, entry.Sha)
				default:
					rec, rerr := fetchcache.DefaultScanReader.ReadScanned(cmd.Context(), run, entry)
					if rerr != nil {
						s.OCRReason = rerr.Error()
					} else {
						s.applyReading(run, rec)
					}
				}
			}

			if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(s)
			}
			fmt.Fprint(cmd.OutOrStdout(), s.render())
			return nil
		},
	}
	c.Flags().String(flags.URL, "", "the http/https URL to read")
	// THE FIRST BACKQUOTED WORD IS COBRA'S PLACEHOLDER, so it is the shape (`backend`), not a value:
	// it rendered `--via archive`, which reads like a default.
	c.Flags().String(flags.Via, "", "reach the source through a named `backend`: "+
		strings.Join(fetchcache.Vias(), " | ")+". They answer DIFFERENT questions — live the URL itself (the same as omitting this); "+
		"archive what the page said on a date (right for web pages, usually the landing page for a subscription article); "+
		"oa whether a legal open copy exists; metadata only that the source exists and where (no text, and the honest answer "+
		"when there is none to get); arxiv the preprint's PDF or LaTeX source; eric the US education index's record, with its "+
		"full text where ERIC holds an authorised copy; auto tries "+strings.Join(fetchcache.AutoOrder(), ", ")+
		" in that order. A refused live fetch falls back through that same order")
	c.Flags().String(flags.At, "", "with --via archive: bound the capture to YYYYMMDD and take the latest at or before it. "+
		"A priority question needs the FIRST time something was visible, which the newest capture cannot answer")
	c.Flags().Bool(flags.OCR, true, "read a PDF that has no text layer with the local OCR engine; --ocr=false caches it unread")
	return c
}
