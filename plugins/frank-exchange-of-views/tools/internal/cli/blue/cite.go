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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// cite: blue's ONLY mechanism for citing a source.
//
// Blue never hand-writes a footnote. `blue cite` fetches the source through the run cache
// (fetch-once, hash-verified — red re-reads the exact bytes), then places an INVISIBLE
// "<!--cite:c-<hex>-->" anchor at the quoted sentence, exactly the way every kind of anchor is
// placed. Assembly weaves those anchors into the visible [^N] footnotes and a composed
// ## Bibliography. Because the anchor is tool-inserted and leaves only with a retire that names
// it, the record shows exactly what the report references.
//
// A source that cannot be loaded is an UNUSABLE citation: the cite is REJECTED and the
// failure is logged by the tool as a defect (a bare `fetch` miss is only an error — but the
// DECISION to cite an unreachable source is a protocol event worth surfacing).
//
// Crash-safety is the append-only record's: the cite and the Anchor that places its marker are two
// appends, and the report is replayed from the record — a crash before them leaves nothing, and a
// --key retry is idempotent and places the Anchor a crash between them left owed.
func newCite() *cobra.Command {
	c := seat.Prose(seat.New("cite", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// A CORRECTION RE-STATES THE CITE; it does not fetch again or mint a new label. The label,
		// the hash, the access date, the text origin, the pages and what the fetch learned about the
		// work are the corrected act's, because the tool assigned or computed them. Everything the
		// seat types comes from this command line exactly as a new cite takes it (citeFields): a
		// flag left out is left out, and the correction refuses it as a change wherever the act
		// recorded a value. Only the title and the argument — the seat's own wording — may differ.
		target, err := s.CorrectionTarget()
		if err != nil {
			return nil, err
		}
		prior, correcting := target.(*recordpb.Cite)
		quote, url, title := seat.Location(cmd), seat.Str(cmd, flags.URL), seat.Str(cmd, flags.Title)
		ocrQuote := seat.Str(cmd, flags.OCRQuote)
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

		read, err := sourceTextRead(cmd)
		if err != nil {
			return nil, err
		}

		// The replacement's marker is the original's: it re-carries the label and appends no Anchor,
		// so the report carries the one anchor the original placed, wherever edits moved it.
		if correcting {
			body := &recordpb.Cite{
				Label:              held(prior.Label),
				Sha256:             held(prior.Sha256),
				AccessDate:         held(prior.AccessDate),
				SourceTextOrigin:   held(prior.SourceTextOrigin),
				Pages:              append([]int32(nil), prior.GetPages()...),
				OcrEngine:          held(prior.OcrEngine),
				OcrTextSha:         held(prior.OcrTextSha),
				WorkStatus:         held(prior.WorkStatus),
				SourceCompleteness: held(prior.SourceCompleteness),
			}
			citeFields(cmd, body, read, why)
			// The pages stay the corrected act's: a correction never re-locates. The span is the
			// seat's, so it is set exactly when it is given — left out against an act that holds
			// one, or given against an act that holds none, it is a change the correction refuses.
			if seat.Given(cmd, flags.OCRQuote) {
				body.OcrQuote = proto.String(ocrQuote)
			}
			if _, err := record.Append(s.Identity(), body); err != nil {
				return nil, err
			}
			return citeResult{Label: prior.GetLabel(), URL: prior.GetUrl(), Sha256: prior.GetSha256(), VoiceTells: tells}, nil
		}

		// Crash-retry idempotency: a prior cite under this --key returns its label, no
		// second fetch AND no second anchor (BEFORE any effect). The cite and its anchor are two
		// appends; a retry finishes the pair at the location the cite stored.
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingCiteByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			if err := seat.PlaceOwed(s, run, "blue cite", prior, key); err != nil {
				return nil, err
			}
			return citeResult{Label: prior, Idempotent: true, VoiceTells: tells}, nil
		}

		// Resolve the source through the run cache (fetch-once). A FAILURE is an unusable
		// citation: reject AND log a defect (unlike a bare `fetch` miss).
		entry, _, _, err := fetchcache.Resolve(run, url, fetchcache.Default)
		if err != nil {
			msg := fmt.Sprintf("blue cite: could not load %s: %v — pick a reachable source or an archive.org snapshot", url, err)
			if _, ferr := record.Append(s.Identity(), &recordpb.Log{Text: proto.String(msg), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}); ferr != nil {
				return nil, ferr
			}
			return nil, record.ToolLogged(errors.New(msg))
		}

		// A LEAF READING OF AN ABSTRACT IS A CLAIM TO HAVE READ THE STUDY. The fetch has already
		// looked for a copy carrying the body; where the copy it kept is known not to be one, the
		// reading the record can carry is `summary_only` — which the vocabulary defines to include
		// an abstract — or `unread` for a page that is not the work at all.
		if read == recordpb.SourceTextRead_SOURCE_TEXT_READ_LEAF && entry.WithoutBody() {
			return nil, fmt.Errorf("blue cite: the copy of %s this run holds is %s — %s. A leaf reading claims "+
				"you read the work itself; cite it with --%s summary_only for an abstract, or unread for a page "+
				"that is not the work, or cite a source whose text you read", url, entry.Completeness, entry.CompletenessReason, flags.SourceText)
		}

		// WHERE THE CITED TEXT CAME FROM, and for OCR text, WHICH PAGE THE QUOTE SITS ON. The
		// binary finds the page from the span; a seat never types one. Refused before the label is
		// minted, so a refusal leaves nothing on the record.
		ocr, err := locateOCRQuote(run, entry, url, ocrQuote, read)
		if err != nil {
			return nil, err
		}

		// THE ANCHOR EVENT BELOW IS THE MARKER: it carries the quote, and reportproj.Render re-places
		// the invisible citation anchor at that quote on every read. There is no file to splice, so
		// a crash before the appends leaves nothing to adopt. Mint the label (it forms the marker)
		// and VALIDATE the placement against the current render — a mis-quote or in-fence quote is
		// refused now, and the validated bytes discarded.
		label := record.NewID("citation")
		if err := seat.Places(run, label, quote, citeRefusal); err != nil {
			return nil, err
		}

		// access_date is engine-supplied from the record clock (pinned under the golden harness),
		// not typed by the seat.
		body := &recordpb.Cite{
			Label:      proto.String(label),
			Sha256:     proto.String(entry.Sha),
			AccessDate: proto.String(record.Now().Format("2006-01-02")),

			SourceTextOrigin: ocr.origin.Enum(),
			Pages:            ocr.pages(),

			// WHAT THE LITERATURE SAYS ABOUT THE WORK, stamped here for the same reason the
			// origin above is: the seat cannot be asked to remember it. `fetch` prints the
			// retraction in a paragraph addressed to whoever ran it, and a paragraph is not a
			// carrier — the citation reaches the reader through the record, and until this line
			// existed a retracted source made a footnote indistinguishable from a sound one.
			WorkStatus: record.WorkStatusOf(entry.Retraction()).Enum(),
			// AND WHICH PART OF THE WORK THE COPY IS — a summary_only citation of an abstract and
			// one of a paper read the same until this was on the record.
			SourceCompleteness: entry.SourceCompleteness().Enum(),
		}
		if len(ocr.loc.Pages) > 0 {
			body.OcrQuote = proto.String(ocrQuote)
			body.OcrEngine = proto.String(ocr.loc.Engine)
			body.OcrTextSha = proto.String(ocr.loc.TextSha)
		}
		citeFields(cmd, body, read, why)
		if err := seat.AppendPlaced(s, body, label, quote); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: entry.Sha, Pages: ocr.pages(), VoiceTells: tells}, nil
	}))

	enumhelp.Flag(c, flags.SourceText, record.MustEnum("cite", "source_text_read"),
		"how much of the source you actually READ; omitted records `unread`. A leaf reading is refused on a copy fetch recorded as the work's abstract, or as not the work")
	flags.Text(c, flags.Quote, flags.DescQuote+". A mis-quote is rejected rather than guessed at")
	c.Flags().String(flags.URL, "", flags.DescURL)
	flags.Text(c, flags.Title, flags.DescTitle)
	seat.Require(c, flags.Quote, flags.URL, flags.Title)
	flags.Text(c, flags.OCRQuote, "for OCR-derived text: the span you quote, verbatim from the source's reading (not the report). The tool records the PDF page it sits on; required with --source-text leaf")
	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL mints the citation's id")
	return seat.Correctable(c)
}

// sourceTextRead is the reading the seat asserts for the source. THE DEFAULT IS THE WEAK CLAIM: a
// citation nobody has asserted a reading for is UNREAD — the honest state costs the seat nothing,
// and only a stronger one is stated on purpose.
func sourceTextRead(cmd *cobra.Command) (recordpb.SourceTextRead, error) {
	w := seat.Str(cmd, flags.SourceText)
	if w == "" {
		return recordpb.SourceTextRead_SOURCE_TEXT_READ_UNREAD, nil
	}
	v, known := record.SourceTextReadOf(w)
	if !known || v == recordpb.SourceTextRead_SOURCE_TEXT_READ_UNSPECIFIED {
		return v, fmt.Errorf("blue cite: %q is not a reading this record can carry (leaf | summary_only | unread)", w)
	}
	return v, nil
}

// citeFields sets every field of a cite that the SEAT types, for a new cite and for its correction
// alike — one builder, so a correction re-states the act with the presence the act itself has.
//
// THE ARGUMENT FOR THE CITATION goes on the record, in Cite.text — why this source backs this
// sentence. It is not printed in the report (the note and the Bibliography print the title, and a
// seat's argument there is exactly the run-voice the report refuses); the `evidence` view shows it
// beside the source, where red decides what to verify. Set only when given, so "no argument
// offered" stays distinct from an empty one.
func citeFields(cmd *cobra.Command, body *recordpb.Cite, read recordpb.SourceTextRead, why string) {
	body.Location = proto.String(seat.Location(cmd))
	body.Url = proto.String(seat.Str(cmd, flags.URL))
	body.Title = proto.String(seat.Str(cmd, flags.Title))
	body.CiteKey = proto.String(seat.Str(cmd, flags.Key))
	body.SourceTextRead = read.Enum()
	if seat.Given(cmd, flags.Reason) {
		body.Text = proto.String(why)
	}
}

// citeRefusal is the refusal `cite` gives for a quote Attach will not place, on a first call and on
// the retry that places the anchor a first call left owed.
func citeRefusal(err error) error {
	switch {
	case errors.Is(err, anchortext.ErrMisQuote):
		return fmt.Errorf("blue cite: the quoted content was not found in report.md — quote the EXACT sentence you are citing (via --quote) — the whole string is matched, so a section heading prepended to it matches nothing")
	case errors.Is(err, anchortext.ErrInFence):
		return fmt.Errorf("blue cite: the quote resolves inside a code fence — cite a prose sentence, not code")
	}
	return anchortext.Refusal("blue cite", err)
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
	note := reportvoice.Note("the source's note and Bibliography entry", r.VoiceTells)
	if r.Idempotent {
		return "cite " + r.Label + " (idempotent retry — existing anchor returned)" + note
	}
	return "citation recorded: " + r.Label + " — an invisible anchor at the quote, woven into the bibliography at assembly (" + r.URL + ")" + note
}
