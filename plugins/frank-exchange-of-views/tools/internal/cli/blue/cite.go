package blue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// cite: blue's ONLY mechanism for citing a source.
//
// Blue never hand-writes a footnote. `blue cite` fetches the source through the run cache
// (fetch-once, hash-verified — red re-reads the exact bytes), then splices an INVISIBLE,
// IMMORTAL "<!--cite:c-<hex>-->" anchor at the quoted sentence, exactly the way a lens
// finding is anchored. Assembly weaves those anchors into the visible [^N] footnotes and a
// composed ## Bibliography. Because the anchor is tool-inserted and the lockdown forbids
// removing it by a raw edit, the set of cite events is a strict bijection with the anchors
// in the document — the record shows exactly what the report references.
//
// A source that cannot be loaded is an UNUSABLE citation: the cite is REJECTED and the
// failure is auto-logged as friction (a bare `fetch` miss is only an error — but the
// DECISION to cite an unreachable source is a protocol event worth surfacing).
//
// Crash-safety is the append-only record's: the cite event IS the anchor (it carries the quote,
// and the report is replayed from the record), so there is no separate marker write to tear from
// the event — a crash before the append leaves nothing, and a --key retry is idempotent.
func newCite() *cobra.Command {
	c := seat.Prose(seat.New("cite", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// A CORRECTION RE-STATES THE CITE; it does not fetch again or mint a new label. The label,
		// url, hash, anchoring quote and access date are the corrected act's, and only the title and
		// the argument — the seat's own wording — may change. A flag left out is the act's own value;
		// one given that differs is refused by the correction as a change to a frozen field.
		target, err := s.CorrectionTarget()
		if err != nil {
			return nil, err
		}
		prior, correcting := target.(*recordpb.Cite)
		quote, url, title := seat.Str(cmd, flags.Quote), seat.Str(cmd, flags.URL), seat.Str(cmd, flags.Title)
		ocrQuote := seat.Str(cmd, flags.OCRQuote)
		if correcting {
			if !seat.Given(cmd, flags.OCRQuote) {
				ocrQuote = prior.GetOcrQuote()
			}
			if !seat.Given(cmd, flags.Quote) {
				quote = prior.GetLocation()
			}
			if !seat.Given(cmd, flags.URL) {
				url = prior.GetUrl()
			}
			if !seat.Given(cmd, flags.Title) {
				title = prior.GetTitle()
			}
		}
		if strings.TrimSpace(quote) == "" {
			return nil, fmt.Errorf("blue cite requires --quote: the EXACT sentence to anchor the citation at, verbatim from the report as `show report` serves it, and nothing else")
		}
		if strings.TrimSpace(url) == "" {
			return nil, fmt.Errorf("blue cite requires --url: the source being cited (fetched once and cached; red re-reads the same bytes)")
		}
		if strings.TrimSpace(title) == "" {
			return nil, fmt.Errorf("blue cite requires --title: the source's name as it appears in the composed bibliography")
		}

		// THE TITLE IS THE ONE SEAT-TYPED STRING THE REPORT PRINTS for a cite — in its note and its
		// Bibliography entry, "<title>. <url> (accessed <date>)" (report/assemble.go); the quote is text the
		// report already holds and the rest is the tool's. So the title meets the voice advisory.
		// ADVISORY, COMPUTED BEFORE ANY RETURN so an idempotent retry carries it; it refuses
		// nothing — a real source's own title can legitimately carry a tell.
		tells := spanVoiceTells(title)

		// The argument for the citation is resolved BEFORE the retry check, as `edit` and `prove`
		// resolve theirs: a retry is the same act, judged on the same inputs.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}

		// The replacement's marker is the original's: same label, same quote, and the render skips a
		// marker the text already holds, so the report carries one anchor wherever edits moved it.
		if correcting {
			body := proto.Clone(prior).(*recordpb.Cite)
			body.Location, body.Url, body.Title = proto.String(quote), proto.String(url), proto.String(title)
			// The pages stay the corrected act's: a correction never re-locates. A differing
			// --ocr-quote is set so the correction refuses it as a change to a frozen field.
			if seat.Given(cmd, flags.OCRQuote) {
				body.OcrQuote = proto.String(ocrQuote)
			}
			if seat.Given(cmd, flags.Reason) {
				body.Text = proto.String(why)
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return nil, err
			}
			return citeResult{Label: prior.GetLabel(), URL: prior.GetUrl(), Sha256: prior.GetSha256(), VoiceTells: tells}, nil
		}

		// Crash-retry idempotency: a prior cite under this --key returns its label, no
		// second fetch AND no second anchor (BEFORE any effect).
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingCiteByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			return citeResult{Label: prior, Idempotent: true, VoiceTells: tells}, nil
		}

		// Resolve the source through the run cache (fetch-once). A FAILURE is an unusable
		// citation: reject AND auto-emit a friction event (unlike a bare `fetch` miss).
		entry, _, _, err := fetchcache.Resolve(run, url, fetchcache.Default)
		if err != nil {
			msg := fmt.Sprintf("blue cite: could not load %s: %v — pick a reachable source or an archive.org snapshot", url, err)
			if _, ferr := record.Append(s.Identity(), &recordpb.Log{Text: proto.String(msg), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}); ferr != nil {
				return nil, ferr
			}
			return nil, errors.New(msg)
		}

		read := recordpb.SourceTextRead_SOURCE_TEXT_READ_UNREAD
		if w := seat.Str(cmd, flags.SourceText); w != "" {
			v, known := record.SourceTextReadOf(w)
			if !known || v == recordpb.SourceTextRead_SOURCE_TEXT_READ_UNSPECIFIED {
				return nil, fmt.Errorf("blue cite: %q is not a reading this record can carry (leaf | summary_only | unread)", w)
			}
			read = v
		}

		// WHERE THE CITED TEXT CAME FROM, and for OCR text, WHICH PAGE THE QUOTE SITS ON. The
		// binary finds the page from the span; a seat never types one. Refused before the label is
		// minted, so a refusal leaves nothing on the record.
		ocr, err := locateOCRQuote(run, entry, url, ocrQuote, read)
		if err != nil {
			return nil, err
		}

		// THE CITE EVENT IS THE ANCHOR. It carries the quote in Location, and reportproj.Render
		// re-places the invisible <!--cite:c-…--> marker at that quote on every read. There is no
		// file to splice, so there is no torn-splice window: the marker cannot exist without the
		// event that names it, and a crash before the append leaves nothing to adopt. Mint the label
		// (it forms the marker) and VALIDATE the placement against the current render — a mis-quote
		// or in-fence quote is refused now with the same message, and the validated bytes discarded.
		label := record.NewCitationID()
		marker := "<!--cite:" + label + "-->"
		current, err := reportproj.RenderFromRecord(run)
		if err != nil {
			return nil, err
		}
		if _, aerr := anchortext.InsertAnchor([]byte(current), quote, marker); aerr != nil {
			switch {
			case errors.Is(aerr, anchortext.ErrMisQuote):
				return nil, fmt.Errorf("blue cite: the quoted content was not found in report.md — quote the EXACT sentence you are citing (via --quote) — the whole string is matched, so a section heading prepended to it matches nothing")
			case errors.Is(aerr, anchortext.ErrInFence):
				return nil, fmt.Errorf("blue cite: the quote resolves inside a code fence — cite a prose sentence, not code")
			}
			return nil, aerr
		}

		// The anchor is recorded as the cite event. access_date is engine-supplied
		// from the record clock (pinned under the golden harness), not typed by the seat.
		body := &recordpb.Cite{
			Label:      proto.String(label),
			Url:        proto.String(url),
			Sha256:     proto.String(entry.Sha),
			Title:      proto.String(title),
			Location:   proto.String(quote),
			AccessDate: proto.String(record.Now().Format("2006-01-02")),
			CiteKey:    proto.String(seat.Str(cmd, flags.Key)),

			SourceTextOrigin: ocr.origin.Enum(),
			Pages:            ocr.pages(),
		}
		if len(ocr.loc.Pages) > 0 {
			body.OcrQuote = proto.String(ocrQuote)
			body.OcrEngine = proto.String(ocr.loc.Engine)
			body.OcrTextSha = proto.String(ocr.loc.TextSha)
		}
		// THE DEFAULT IS THE WEAK CLAIM. A citation nobody has asserted a reading for is UNREAD:
		// the honest state costs the seat nothing, and only a stronger one is stated on purpose.
		body.SourceTextRead = &read
		// THE ARGUMENT FOR THE CITATION goes on the record, in Cite.text — why this source backs
		// this sentence. It is not printed in the report (the note and the Bibliography print the title, and a
		// seat's argument there is exactly the run-voice the report refuses); the `evidence` view
		// shows it beside the source, where red decides what to verify. This flag was registered
		// and never read: a seat's reason was accepted and recorded nowhere. Set only when given,
		// so "no argument offered" stays distinct from an empty one.
		if seat.Given(cmd, flags.Reason) {
			body.Text = proto.String(why)
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: entry.Sha, Pages: ocr.pages(), VoiceTells: tells}, nil
	}))

	enumhelp.Flag(c, flags.SourceText, record.MustEnum("cite", "source_text_read"),
		"how much of the source you actually READ; omitted records `unread`")
	flags.Text(c, flags.Quote, flags.DescQuote+". A mis-quote is rejected rather than guessed at")
	c.Flags().String(flags.URL, "", flags.DescURL)
	flags.Text(c, flags.Title, flags.DescTitle)
	flags.Text(c, flags.OCRQuote, "for OCR-derived text: the span you quote, verbatim from the source's reading (not the report). The tool records the PDF page it sits on; required with --source-text leaf")
	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the c-<hex> label")
	return seat.Correctable(c)
}

type ocrCite struct {
	origin recordpb.SourceTextOrigin
	loc    fetchcache.SpanLocation
}

func (o ocrCite) pages() []int32 {
	var out []int32
	for _, p := range o.loc.Pages {
		out = append(out, int32(p))
	}
	return out
}

// locateOCRQuote stamps a cite's text origin and, for a quote from an OCR reading, the pages it
// sits on. A leaf citation of OCR text must name its span: a leaf reading is the claim that the
// source SAYS this, and OCR text can misread, so the page is what lets red check the pixels. On a
// source with no OCR reading the span is refused rather than dropped — a flag accepted and
// recorded nowhere reads as a page check that will never be owed.
func locateOCRQuote(run record.Run, entry fetchcache.Entry, url, span string, read recordpb.SourceTextRead) (ocrCite, error) {
	origin, rec, err := fetchcache.TextOrigin(run, entry)
	if err != nil {
		return ocrCite{}, err
	}
	out := ocrCite{origin: origin}
	given := strings.TrimSpace(span) != ""
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		if !given {
			if read == recordpb.SourceTextRead_SOURCE_TEXT_READ_LEAF {
				return ocrCite{}, fmt.Errorf("blue cite: a leaf citation of OCR-derived text names the span it quotes from the reading, with --%s, so the tool can record its page — the reading is at %s", flags.OCRQuote, fetchcache.OCRTextPath(run, entry.Sha))
			}
			return out, nil
		}
		loc, lerr := fetchcache.LocateSpan(run, rec, span)
		switch {
		case errors.Is(lerr, fetchcache.ErrReadingIncomplete):
			return ocrCite{}, fmt.Errorf("blue cite: %v — `fetch --url %s` re-derives it", lerr, url)
		case lerr != nil:
			return ocrCite{}, fmt.Errorf("blue cite: --%s %v", flags.OCRQuote, lerr)
		}
		out.loc = loc
		return out, nil
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE:
		if given {
			return ocrCite{}, fmt.Errorf("blue cite: --%s locates a span in the source's OCR reading, and %s has none — `fetch --url %s` reads a scan first", flags.OCRQuote, url, url)
		}
	default:
		if given {
			return ocrCite{}, fmt.Errorf("blue cite: --%s is for OCR-derived text, and this source's text is not OCR-derived, so there is no reading to locate a page in — quote the report sentence with --%s alone", flags.OCRQuote, flags.Quote)
		}
	}
	return out, nil
}

type citeResult struct {
	Label  string `json:"label"`
	URL    string `json:"url,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	// Pages are the PDF pages the tool found an OCR quote on; the report prints them at the marker.
	Pages      []int32 `json:"pages,omitempty"`
	Idempotent bool    `json:"idempotent,omitempty"`
	// VoiceTells is ADVICE on the title, and the citation is already recorded by the time it renders.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r citeResult) Human() string {
	note := voiceNote("the source's note and Bibliography entry", r.VoiceTells)
	if r.Idempotent {
		return "cite " + r.Label + " (idempotent retry — existing anchor returned)" + note
	}
	return "citation recorded: " + r.Label + " — an invisible immortal anchor at the quote, woven into the bibliography at assembly (" + r.URL + ")" + note
}
