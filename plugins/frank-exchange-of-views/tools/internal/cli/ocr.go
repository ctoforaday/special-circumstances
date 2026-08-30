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
// A ROOT OPERATOR COMMAND, NOT A SEAT VERB — the same reasoning fetch states. Rasterising a
// page is not an act on the record: it takes no seat identity and writes no event. What a
// seat later SAYS the pages contain will be an act on the record, and will be a different
// verb.
//
// THE ENGINE IS A SEAT, WHICH IS WHY THERE IS NO OCR LIBRARY HERE. Every Go OCR option is
// Tesseract through cgo (3 of the 6 pairs this repo ships from ubuntu-latest), a wrapper
// around a machine prerequisite, or `danlock/gogosseract` — CGo-free, and measured on
// 2026-08-29 to PANIC against the wazero this module pins (`emscripten_stack_get_current not
// exported`), green only at wazero ≤1.7.3 against go-pdfium v1.19.8's required v1.12.0, one
// release, dated 2023-11-04, and abandoned. PDFium already renders; a seat already reads.
// The library was the part that could be removed.
func newOCR() *cobra.Command {
	c := &cobra.Command{
		Use:           "ocr",
		Short:         "render a scanned document's pages so a seat can read them (#644)",
		Long:          "ocr operates on a document `fetch` has already cached. It does not itself read anything: `ocr pages` rasterises the document and names the images, and what those images SAY is a seat's reading, recorded separately. A document whose text layer was already extracted is refused unless --force, because re-reading text the record already holds spends a model to re-derive what a file already says.",
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
		Long: "pages renders each page of the cached document --sha to a PNG under <run>/cache/<sha>.pages/ and writes a render record beside them holding the resolution, one hash per page image, and the renderer's library@semver. It prints a summary naming the directory and the page range — never an image, and never the document. " +
			"A DOCUMENT THAT ALREADY HAS TEXT IS REFUSED. Where fetch extracted a text layer, its extraction is already on disk and reading the pixels instead spends a model to re-derive it, less accurately. --force renders anyway, for the case where the text layer is present but wrong.",
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
					"content is at %s", sha, contentTypeOrUnknown(entry.ContentType), fetchcache.Path(run, sha))
			}
			// THE GUARD THAT MAKES THIS VERB CHEAP TO USE WRONGLY. Rendering a document whose text
			// was already extracted spends a model to re-derive, less accurately, what is already a
			// file on disk. --force exists for the real case where a text layer is present but wrong.
			if !force && entry.TextExtracted != nil && *entry.TextExtracted {
				return fmt.Errorf("sha %s already has an extracted text layer at %s — reading its "+
					"pixels would re-derive that text less accurately and at a model's cost. Pass "+
					"--force if the text layer is present but wrong",
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

// contentTypeOrUnknown keeps an absent Content-Type from rendering as an empty string in the
// middle of a sentence, where it reads as a missing word rather than a missing measurement.
func contentTypeOrUnknown(ct string) string {
	if ct == "" {
		return "of no recorded content type"
	}
	return ct
}

// ocrReadSummary is what `ocr read` prints: what was read, by what, and where it disagreed.
type ocrReadSummary struct {
	Sha      string `json:"sha"`
	Model    string `json:"model"`
	Pages    int    `json:"pages"`
	DPI      int    `json:"dpi"`
	TextPath string `json:"text_path"`
	TextSha  string `json:"text_sha"`
	// OCRDerived is always true here and is printed anyway. It is the field that keeps text a
	// machine read off pixels distinguishable from text an author embedded, and a reader who
	// does not see it stated has to infer it from the verb that produced the file.
	OCRDerived bool `json:"ocr_derived"`
	// Divergences names the pages whose two readings disagreed. A COUNT ALONE WOULD BE USELESS:
	// the point of the check is to send a human to a specific page.
	Divergences []int `json:"divergences,omitempty"`
	InTokens    int64 `json:"input_tokens"`
	OutTokens   int64 `json:"output_tokens"`
	Reused      bool  `json:"reused"`
}

func (s ocrReadSummary) render() string {
	var b strings.Builder
	line := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	line("sha", s.Sha)
	line("model", s.Model)
	line("pages", fmt.Sprint(s.Pages))
	line("dpi", fmt.Sprint(s.DPI))
	line("text_path", s.TextPath)
	line("text_sha", s.TextSha)
	line("ocr_derived", "true")
	line("input_tokens", fmt.Sprint(s.InTokens))
	line("output_tokens", fmt.Sprint(s.OutTokens))
	line("reused", fmt.Sprint(s.Reused))
	if len(s.Divergences) > 0 {
		line("divergent_pages", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(s.Divergences)), ","), "[]"))
		line("divergence_note", "the two readings disagree on these pages; the text marks each in place "+
			"and both passes are kept beside the page image. Nothing picked a winner.")
	}
	return b.String()
}

// newOCRRead asks a model what the rendered pages say — twice — and records both answers.
//
// THIS IS THE ONE VERB IN THIS TOOL THAT CALLS OUT. Everything else here is local: fetch
// serves cached bytes, extraction runs PDFium in-process, rendering rasterises. This spends a
// model, per page, twice, and that is why it is a separate opt-in verb rather than part of
// fetch — a seat asking for a URL does not expect to spend money, and one document in four
// needs this.
//
// WHAT IT PRODUCES IS NOT REPRODUCIBLE, AND THE RECORD SAYS SO. #636 keyed an extraction to
// library@semver so an audit could re-run it and compare hashes. A model re-reading a page
// returns different bytes, so that check does not exist here. What replaces it is an
// attestation — which model, when, against which image hashes — plus the two-pass agreement.
// That is weaker, and it is stated rather than left to be discovered by a `reproduce` that
// fails mysteriously.
func newOCRRead() *cobra.Command {
	var sha, model string
	var force bool

	c := &cobra.Command{
		Use:   "read",
		Short: "ask a model what the rendered pages say, twice, and record both answers",
		Long: "read sends each page image rendered by `ocr pages` to a model and asks for a transcription — TWICE, independently. Where the two readings agree, the agreed text is written to <run>/cache/<sha>.ocr.txt. Where they DISAGREE, the text marks the page in place and keeps both passes beside the image: nothing picks a winner, because an uncorroborated reading in the position a citation is taken from is the failure this check exists to catch. " +
			"THIS VERB SPENDS A MODEL, per page, twice. It needs credentials (ANTHROPIC_API_KEY, or `ant auth login`) and it is the only verb here that calls out of the machine. " +
			"WHAT IT WRITES IS NOT REPRODUCIBLE: a re-read returns different bytes. The record carries an attestation — model, time, and the hashes of the images actually read — in place of the re-derivation an extraction supports.",
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

			// A READING ALREADY EXISTS, AND RE-READING COSTS REAL MONEY. It is also not idempotent:
			// a second reading returns different text, so silently redoing it would replace a
			// record a seat may already have cited from.
			if prev, had, rerr := fetchcache.ReadReadingRecord(run, sha); rerr == nil && had && !force {
				same := len(prev.RenderShas) == len(rd.PageShas)
				for i := range prev.RenderShas {
					if same && prev.RenderShas[i] != rd.PageShas[i] {
						same = false
					}
				}
				if same {
					s := readSummaryOf(run, sha, prev, true)
					return printOCRRead(cmd, s)
				}
			}

			rec, err := fetchcache.ReadRenderedPages(cmd.Context(), run, sha, model, rd)
			if err != nil {
				return err
			}
			return printOCRRead(cmd, readSummaryOf(run, sha, rec, false))
		},
	}
	c.Flags().StringVar(&sha, flags.Sha, "", "sha256 of a document whose pages ocr pages has already rendered")
	c.Flags().StringVar(&model, flags.Model, defaultReadModel, "the model that reads the pages")
	c.Flags().BoolVar(&force, flags.Force, false, "read again even though a reading of these exact images exists")
	_ = c.MarkFlagRequired(flags.Sha)
	return c
}

// defaultReadModel is the model that reads a page unless the operator names another.
//
// The most capable model is the default DELIBERATELY. This is transcription of a degraded
// scan feeding citation at the leaf, where a misread word becomes a quotation nobody can
// tell is wrong; the cheaper failure mode is silent. Choosing a lesser model to save tokens
// is the operator's call to make explicitly, not this tool's to make quietly.
const defaultReadModel = "claude-opus-5"

func readSummaryOf(run record.Run, sha string, r fetchcache.ReadingRecord, reused bool) ocrReadSummary {
	return ocrReadSummary{
		Sha: r.Sha, Model: r.Model, Pages: len(r.Pages), DPI: r.DPI,
		TextPath: fetchcache.OCRTextPath(run, sha), TextSha: r.TextSha,
		OCRDerived: true, Divergences: r.Divergences(),
		InTokens: r.InTokens, OutTokens: r.OutTok, Reused: reused,
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
