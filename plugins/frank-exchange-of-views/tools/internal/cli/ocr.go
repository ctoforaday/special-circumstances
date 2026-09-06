package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ocrSummary is what `ocr pages` prints: where the images are, and the facts a seat needs to
// decide whether to read them.
//
// The same discipline fetch's summary has — every fact a field, nothing recovered from a
// path. `pages` is the rendered count and it is the LENGTH of the record's hash list, so the
// count and the images cannot disagree.
type ocrSummary struct {
	Sha       string `json:"sha"`
	PagesDir  string `json:"pages_dir"`
	Pages     int    `json:"pages"`
	DPI       int    `json:"dpi"`
	Renderer  string `json:"renderer"`
	FirstPage string `json:"first_page"`
	LastPage  string `json:"last_page"`
	// Reused is true when this render already existed at this resolution and nothing was
	// re-rendered. A seat re-running the verb should be able to tell.
	Reused bool `json:"reused"`
}

func (s ocrSummary) render() string {
	var b strings.Builder
	line := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	line("sha", s.Sha)
	line("pages_dir", s.PagesDir)
	line("pages", fmt.Sprint(s.Pages))
	line("dpi", fmt.Sprint(s.DPI))
	line("renderer", s.Renderer)
	line("first_page", s.FirstPage)
	line("last_page", s.LastPage)
	line("reused", fmt.Sprint(s.Reused))
	return b.String()
}

// newOCR is the group that turns a document with no text layer into something a seat can
// read (#644).
//
// A ROOT OPERATOR COMMAND, NOT A SEAT VERB — the same reasoning fetch states. Rasterising
// or OCRing a page is not an act on the record: it takes no seat identity and writes no
// event. What a seat later SAYS about the reading will be an act on the record, and a
// different verb.
//
// THE ENGINE IS IN THE BINARY NOW (plans/local-ocr.md). The 2026-08-29 survey that ruled
// every Go OCR option out — cgo Tesseract as a machine prerequisite, gogosseract
// panicking against this module's wazero — was answering a different question: whether to
// take a SYSTEM dependency. internal/tessocr takes none: tesseract + leptonica are built
// from hash-pinned sources and statically linked, traineddata embedded, so a release
// binary reads scans with no model, no credentials, no network, and nothing installed on
// any seat.
func newOCR() *cobra.Command {
	c := &cobra.Command{
		Use:           "ocr",
		Short:         "render a scanned document's pages and read them with the local OCR engine (#644)",
		Long:          "ocr operates on a document `fetch` has already cached. `ocr pages` rasterises the document and names the images; `ocr read` runs the local OCR engine over them and records the reading. A document whose text layer was already extracted is refused unless --force, because reading pixels re-derives, less accurately, what a file already says.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.AddCommand(newOCRPages())
	c.AddCommand(newOCRRead())
	return c
}

func newOCRPages() *cobra.Command {
	var sha string
	var dpi int
	var force bool

	c := &cobra.Command{
		Use:   "pages",
		Short: "rasterise every page of a cached PDF and name the images",
		Long: "pages renders each page of the cached document --sha to a GRAYSCALE PNG under <run>/cache/<sha>.pages/ and writes a render record beside them holding the resolution, one hash per page image, and the renderer's library@semver. The default 300 DPI is the OCR engine's operative resolution — its grid-detection and table-reconstruction constants are tuned there, and `ocr read` reads only renders at it; other resolutions exist for a human inspecting the pixels. It prints a summary naming the directory and the page range — never an image, and never the document. " +
			"A DOCUMENT THAT ALREADY HAS TEXT IS REFUSED. Where fetch extracted a text layer, its extraction is already on disk and reading the pixels instead re-derives it, less accurately. --force renders anyway, for the case where the text layer is present but wrong.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			entry, ok, err := fetchcache.LookupSha(run, sha)
			if err != nil {
				return err
			}
			if !ok {
				// NAMED, NOT A BARE MISS. A seat reaches this by mistyping a sha out of a fetch
				// summary, and "not found" without the cache location leaves it guessing whether
				// the document, the run, or the hash is wrong.
				return fmt.Errorf("no cached document has sha %s in %s — `fetch` it first, and take "+
					"the sha from that summary", sha, fetchcache.Dir(run))
			}
			if entry.ContentType != "application/pdf" {
				return fmt.Errorf("sha %s is %s, and only application/pdf renders to pages; its "+
					"content is at %s", sha, fetchcache.ContentTypeOrUnknown(entry.ContentType), fetchcache.Path(run, sha))
			}
			// THE GUARD THAT MAKES THIS VERB CHEAP TO USE WRONGLY. Rendering a document whose text
			// was already extracted sets up an OCR re-derivation, less accurate, of what is already
			// a file on disk. --force exists for the real case where a text layer is present but
			// wrong.
			if !force && entry.TextExtracted != nil && *entry.TextExtracted {
				return fmt.Errorf("sha %s already has an extracted text layer at %s — reading its "+
					"pixels would re-derive that text less accurately. Pass --force if the text "+
					"layer is present but wrong",
					sha, fetchcache.TextPath(run, sha))
			}

			// An existing render at this resolution is reused rather than repeated. A re-render at
			// a DIFFERENT resolution is a different rendering and replaces it.
			rec, have, err := fetchcache.ReadRenderRecord(run, sha)
			if err != nil {
				return err
			}
			reused := have && rec.DPI == dpi && rec.Pages() > 0
			if !reused {
				body, rerr := fetchcache.Read(run, sha)
				if rerr != nil {
					return fmt.Errorf("the index names sha %s but its content file is unreadable: %w", sha, rerr)
				}
				if rec, err = fetchcache.RenderPages(run, sha, body, dpi); err != nil {
					return err
				}
			}

			s := ocrSummary{
				Sha:       rec.Sha,
				PagesDir:  fetchcache.PagesDir(run, sha),
				Pages:     rec.Pages(),
				DPI:       rec.DPI,
				Renderer:  rec.Renderer,
				FirstPage: fetchcache.PagePath(run, sha, 1),
				LastPage:  fetchcache.PagePath(run, sha, rec.Pages()),
				Reused:    reused,
			}
			if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
				b, jerr := json.MarshalIndent(s, "", "  ")
				if jerr != nil {
					return jerr
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), s.render())
			return nil
		},
	}
	c.Flags().StringVar(&sha, flags.Sha, "", "sha256 of a document already in the run cache (from a fetch summary)")
	c.Flags().IntVar(&dpi, flags.DPI, fetchcache.DefaultRenderDPI,
		fmt.Sprintf("render resolution, %d–%d", fetchcache.MinRenderDPI, fetchcache.MaxRenderDPI))
	c.Flags().BoolVar(&force, flags.Force, false, "render even though a text layer was already extracted")
	_ = c.MarkFlagRequired(flags.Sha)
	return c
}

// ocrReadSummary is what `ocr read` prints: what was read, by what, and what the grid
// branch found.
type ocrReadSummary struct {
	Sha      string `json:"sha"`
	Engine   string `json:"engine"`
	Pages    int    `json:"pages"`
	DPI      int    `json:"dpi"`
	TextPath string `json:"text_path"`
	TextSha  string `json:"text_sha"`
	// OCRDerived is always true here and is printed anyway. It is the field that keeps text a
	// machine read off pixels distinguishable from text an author embedded, and a reader who
	// does not see it stated has to infer it from the verb that produced the file.
	OCRDerived bool `json:"ocr_derived"`
	// TablePages counts pages whose ruled grid the engine detected, present only when
	// nonzero — including on the reuse path, where the stored record's table facts must
	// still reach a summary-only reader. The per-page reconstruction stats live on the
	// reading record.
	TablePages int  `json:"table_pages,omitempty"`
	Reused     bool `json:"reused"`
}

func (s ocrReadSummary) render() string {
	var b strings.Builder
	line := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	line("sha", s.Sha)
	line("engine", s.Engine)
	line("pages", fmt.Sprint(s.Pages))
	line("dpi", fmt.Sprint(s.DPI))
	line("text_path", s.TextPath)
	line("text_sha", s.TextSha)
	line("ocr_derived", "true")
	if s.TablePages > 0 {
		line("table_pages", fmt.Sprint(s.TablePages))
	}
	line("reused", fmt.Sprint(s.Reused))
	return b.String()
}

// newOCRRead runs the local OCR engine over the rendered pages, and records the reading.
//
// FULLY LOCAL, LIKE EVERYTHING ELSE HERE NOW: fetch serves cached bytes, extraction runs
// PDFium in-process, rendering rasterises, and this reads with the engine compiled into
// the binary — no credentials, no network, no filter. It remains a separate verb even
// though fetch reads a scanned PDF on its own: fetch reads only what its own extractor
// found no text layer in, and this is how an operator reads pages that door will not open
// — a document already cached, a --force re-read over kept images.
//
// WHAT IT PRODUCES IS REPRODUCIBLE, AND THE RECORD SAYS HOW. #636 keyed an extraction to
// library@semver so an audit could re-run it and compare hashes; this record keys the
// reading to the engine identity (tesseract@x+leptonica@y) and the hashes of the exact
// images read, restoring that check for scans — same binary, same pixels, same bytes.
func newOCRRead() *cobra.Command {
	var sha string
	var force bool

	c := &cobra.Command{
		Use:   "read",
		Short: "read the rendered pages with the local OCR engine, and record the reading",
		Long: "read runs the OCR engine compiled into this binary (tesseract + leptonica, statically linked) over each page image rendered by `ocr pages`. The assembled text is written to <run>/cache/<sha>.ocr.txt and each page's own reading beside its image. " +
			"FULLY LOCAL AND DETERMINISTIC: no model, no credentials, no network. The record keys the reading to the engine identity and the exact image hashes, so an audit can re-derive it byte for byte. It reads only renders at the engine's operative resolution (the default `ocr pages` DPI); re-render if the images are at another. " +
			"A page with a ruled table is detected and reconstructed into |-separated rows, with confidence stats as fields on the record; a reconstruction that cannot place its marks falls back to plain text WITH the failure stated, never a plausible half-table. Pages already read are never derived twice — each carries a receipt validated against its image hash — and --force discards the receipts for a full re-read.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			rd, have, err := fetchcache.ReadRenderRecord(run, sha)
			if err != nil {
				return err
			}
			if !have {
				return fmt.Errorf("nothing is rendered for sha %s — run `ocr pages --sha %s` first; "+
					"this verb reads images, it does not make them", sha, sha)
			}

			// A READING OF THESE EXACT IMAGES ALREADY EXISTS. Re-deriving it is cheap now, but
			// silently replacing a record a seat may already have cited from is not a thing to
			// do as a side effect of re-running a verb — --force is the deliberate lever.
			if prev, had, rerr := fetchcache.ReadReadingRecord(run, sha); rerr == nil && had && !force {
				if fetchcache.SameRenders(prev.RenderShas, rd.PageShas) {
					return printOCRRead(cmd, readSummaryOf(run, sha, prev, true))
				}
			}

			// --force MEANS FORCE: discard the receipts so every page is derived again.
			// Without this the per-page reuse (#679's design, kept) would quietly turn
			// --force into a no-op, which is the operator's lever removed by an
			// optimisation they cannot see.
			if force {
				if err := fetchcache.ClearReceipts(run, sha); err != nil {
					return err
				}
			}
			rec, err := fetchcache.ReadRenderedPages(run, sha, rd)
			if err != nil {
				return err
			}
			return printOCRRead(cmd, readSummaryOf(run, sha, rec, false))
		},
	}
	c.Flags().StringVar(&sha, flags.Sha, "", "sha256 of a document whose pages ocr pages has already rendered")
	c.Flags().BoolVar(&force, flags.Force, false, "read again even though a reading of these exact images exists")
	_ = c.MarkFlagRequired(flags.Sha)
	return c
}

func readSummaryOf(run record.Run, sha string, r fetchcache.ReadingRecord, reused bool) ocrReadSummary {
	return ocrReadSummary{
		Sha: r.Sha, Engine: r.Engine, Pages: len(r.Pages), DPI: r.DPI,
		TextPath: fetchcache.OCRTextPath(run, sha), TextSha: r.TextSha,
		OCRDerived: true, TablePages: r.TablePages(), Reused: reused,
	}
}

func printOCRRead(cmd *cobra.Command, s ocrReadSummary) error {
	if jsonMode, _ := cmd.Flags().GetBool(flags.JSON); jsonMode {
		b, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	}
	fmt.Fprint(cmd.OutOrStdout(), s.render())
	return nil
}
