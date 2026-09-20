package lens

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// render-page: draw ONE page of a cached PDF, so a citation of OCR text can be checked against
// the pixels instead of against the reading that may have misread them.
//
// ONE PAGE, NEVER THE DOCUMENT. The operator's `ocr pages` renders everything; a lens needs the
// page a citation's quote sits on, and a neighbour for context at most. It is READ-ONLY on the
// record, as near-match is: the verification that follows records which image was checked. The
// image goes beside the cache, outside the reading's directory, which a render or a re-read
// clears.
func newRenderPage() *cobra.Command {
	c := seat.New("render-page", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		sha := strings.TrimSpace(seat.Str(cmd, flags.Sha))
		page, _ := cmd.Flags().GetInt(flags.Page)
		entry, ok, err := fetchcache.LookupSha(run, sha)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, feov.Errorf(feov.Validation, "lens render-page: no cached document has sha %s in %s — take the sha from `show evidence` or the fetch summary", sha, fetchcache.Dir(run))
		}
		if fetchcache.MediaType(entry.ContentType) != "application/pdf" {
			return nil, feov.Errorf(feov.Validation, "lens render-page: only a PDF has pages, and sha %s is %s", sha, fetchcache.ContentTypeOrUnknown(entry.ContentType))
		}
		body, err := fetchcache.Read(run, sha)
		if err != nil {
			return nil, err
		}
		// AT THE RESOLUTION THE READING WAS MADE AT, NOT A CONSTANT. This verb exists so a lens can
		// check a cited page against the pixels the quote came from, and that check is answered by
		// comparing render shas. A scan is read at its own resolution (#1031) and now at its own
		// exact pixel count (#1105), so rendering here at a fixed 300 made matches_reading false on
		// every scanned page — the check reporting "these differ" for a document nobody had touched.
		dpi, derr := fetchcache.PageRenderDPI(run, sha, body, page)
		if derr != nil {
			return nil, derr
		}
		path, renderSha, _, err := fetchcache.RenderOnePage(run, sha, body, page, dpi)
		if errors.Is(err, fetchcache.ErrPageOutOfRange) {
			return nil, feov.Errorf(feov.Validation, "lens render-page: %v", err)
		}
		if err != nil {
			return nil, err
		}
		out := renderPageResult{Path: path, Page: page, DPI: dpi.Max(), RenderSha: renderSha, Reading: "none"}
		rec, had, err := fetchcache.ReadReadingRecord(run, sha)
		if err != nil {
			return nil, err
		}
		if had && page <= len(rec.RenderShas) {
			out.Reading = "present"
			out.PageText = fetchcache.PageTextPath(run, sha, page)
			out.ReadingRenderSha = rec.RenderShas[page-1]
			matches := rec.RenderShas[page-1] == renderSha
			out.MatchesReading = &matches
		}
		return out, nil
	})
	c.Flags().String(flags.Sha, "", "the cached document's sha256, as the evidence view lists it beside the citation")
	c.Flags().Int(flags.Page, 0, "the 1-based PDF page to draw — one of a citation's pages, as the evidence view lists them")
	_ = c.MarkFlagRequired(flags.Sha)
	_ = c.MarkFlagRequired(flags.Page)
	return c
}

type renderPageResult struct {
	Path      string  `json:"path"`
	Page      int     `json:"page"`
	DPI       float64 `json:"dpi"`
	RenderSha string  `json:"render_sha"`
	// Reading is present or none: whether the tool holds an OCR reading of this document.
	Reading          string `json:"reading"`
	PageText         string `json:"page_text,omitempty"`
	ReadingRenderSha string `json:"reading_render_sha,omitempty"`
	// MatchesReading says whether this image is the one the reading was made from. False is not a
	// refusal: the image still shows the page, drawn by a different renderer.
	MatchesReading *bool `json:"matches_reading,omitempty"`
}

func (r renderPageResult) Human() string {
	s := fmt.Sprintf("page %d drawn at %.2f DPI: %s (sha256 %s)", r.Page, r.DPI, r.Path, r.RenderSha)
	if r.Reading == "none" {
		return s + " — the tool holds no OCR reading of this document"
	}
	s += "\n  the reading of this page: " + r.PageText
	if r.MatchesReading != nil && !*r.MatchesReading {
		s += "\n  this image is not the render the reading was made from (" + r.ReadingRenderSha + ") — it still shows the page"
	}
	return s
}
