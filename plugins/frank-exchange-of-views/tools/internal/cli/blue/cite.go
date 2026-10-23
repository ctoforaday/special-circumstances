package blue

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
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
		quote, url, title := seat.Location(cmd), seat.Str(cmd, flags.URL), seat.Str(cmd, flags.Title)
		span := seat.Str(cmd, flags.SourceSpan)
		if strings.TrimSpace(quote) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --quote: the EXACT sentence to anchor the citation at, verbatim from the report as `show report` serves it, and nothing else")
		}
		if strings.TrimSpace(url) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --url: the source being cited (fetched once and cached; red re-reads the same bytes)")
		}
		if strings.TrimSpace(title) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --title: the source's name as it appears in the composed bibliography")
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

		// Crash-retry idempotency: a prior cite under this --key returns its label, no
		// second fetch AND no second anchor (BEFORE any effect). A cite on the record without its
		// anchor — one migrate brought forward from an archive whose run crashed between the two — is finished by the retry, at the location it stored.
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingCiteByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != nil {
			// A KEY IS ONE CITATION'S NAME: a retry under it cites the same source, under the same
			// title, at the same span, for the same sentence and with the same argument, or it is a
			// second citation reusing the name.
			retry := &recordpb.Cite{SourceSpan: proto.String(span)}
			citeFields(cmd, retry, why)
			if err := seat.RepeatsKeyed(key, "citation", prior.GetLabel(), prior, retry, citeAsserted...); err != nil {
				return nil, err
			}
			if err := seat.PlaceOwed(s, prior.GetLabel(), key); err != nil {
				return nil, err
			}
			return citeResult{Label: prior.GetLabel(), URL: prior.GetUrl(), Sha256: prior.GetSha256(), Pages: prior.GetPages(), Origin: recordpb.Word(prior.GetSourceTextOrigin()), Idempotent: true, VoiceTells: tells}, nil
		}

		// Resolve the source through the run cache (fetch-once). A FAILURE is an unusable
		// citation: reject AND log a defect (unlike a bare `fetch` miss).
		entry, _, _, err := fetchcache.Resolve(run, url, fetchcache.Default)
		if err != nil {
			// THE CAUSE STAYS BENEATH IT, so the refusal carries the code the failure came with: a
			// source that would not load is not the seat's ill-formed input.
			unloaded := fmt.Errorf("could not load %s: %w — pick a reachable source or an archive.org snapshot", url, err)
			msg := unloaded.Error()
			if _, ferr := record.Append(s.Identity(), &recordpb.Log{Text: proto.String(msg), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}); ferr != nil {
				return nil, ferr
			}
			return nil, record.ToolLogged(unloaded)
		}

		// WHERE THE CITED TEXT CAME FROM, WHERE THE SPAN SITS IN IT, and what the source says
		// around it. The binary finds the span; a seat never types an offset or a page. Refused
		// before the label is minted, so a refusal leaves nothing on the record.
		at, err := locateSourceSpan(run, entry, url, span)
		if err != nil {
			return nil, err
		}

		// THE ANCHOR EVENT BELOW IS THE MARKER: it carries the quote, and reportproj.Render re-places
		// the invisible citation anchor at that quote on every read. There is no file to splice. Mint
		// the label (it forms the marker); the placement is validated against the report the Cite and
		// its Anchor land on — a mis-quote or in-fence quote is refused and nothing recorded.
		label := record.NewID("citation")

		// access_date is engine-supplied from the record clock (pinned under the golden harness),
		// not typed by the seat.
		body := &recordpb.Cite{
			Label:      proto.String(label),
			Sha256:     proto.String(entry.Sha),
			AccessDate: proto.String(record.Now().Format("2006-01-02")),

			SourceTextOrigin: at.origin.Enum(),
			Pages:            at.pages(),

			// WHAT THE LITERATURE SAYS ABOUT THE WORK, stamped here for the same reason the
			// origin above is: the seat cannot be asked to remember it. `fetch` prints the
			// retraction in a paragraph addressed to whoever ran it, and a paragraph is not a
			// carrier — the citation reaches the reader through the record, and until this line
			// existed a retracted source made a footnote indistinguishable from a sound one.
			WorkStatus: record.WorkStatusOf(entry.Retraction()).Enum(),
			// AND WHICH PART OF THE WORK THE COPY IS — a summary_only citation of an abstract and
			// one of a paper read the same until this was on the record.
			SourceCompleteness: entry.SourceCompleteness().Enum(),
			// AND WHOSE BYTES WERE READ. The seat typed --url; the cache may have answered it with
			// the full text the page pointed to, a copy recovered after a refusal, a fuller copy, or
			// a named backend's answer. sha256 above is those bytes' hash, and without this the
			// citation said it read the typed url.
			CopySubstitution: entry.CopySubstitution().Enum(),
		}
		if entry.Substitution != "" && entry.CopyURL != "" {
			body.CopyUrl = proto.String(entry.CopyURL)
		}
		if at.located {
			body.SourceSpan = proto.String(span)
			body.SpanStart = proto.Int32(int32(at.loc.Start))
			body.SpanEnd = proto.Int32(int32(at.loc.End))
			body.SpanTextSha = proto.String(at.loc.TextSha)
			body.SourceBefore = proto.String(at.loc.Before)
			body.SourceMatched = proto.String(at.loc.Matched)
			body.SourceAfter = proto.String(at.loc.After)
		}
		if len(at.loc.Pages) > 0 {
			body.OcrEngine = proto.String(at.loc.Engine)
			body.OcrTextSha = proto.String(at.loc.ReadingSha)
		}
		citeFields(cmd, body, why)
		if err := seat.AppendPlaced(s, body, label, quote, citeRefusal); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: entry.Sha, Pages: at.pages(), Origin: recordpb.Word(at.origin), VoiceTells: tells}, nil
	}))

	flags.Text(c, flags.Quote, flags.DescQuote+". A mis-quote is rejected rather than guessed at")
	c.Flags().String(flags.URL, "", flags.DescURL)
	flags.Text(c, flags.Title, flags.DescTitle)
	seat.Require(c, flags.Quote, flags.URL, flags.Title)
	flags.Text(c, flags.SourceSpan, "required for every source whose text the tool holds: the words you rely on, copied from the source's cached text (not the report, and not your own wording of it). "+
		"The tool finds them there once — "+fetchcache.SpanNormalisation+" — and refuses one that cuts into a clause: "+fetchcache.SpanBoundary+". "+
		"It records where, the source's text around them, and for OCR-derived text the PDF page. A lens reads that text, not your reason, against the sentence. "+
		"A source the tool holds no text for is cited without it")
	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL mints the citation's id. A repeat under the handle with another --quote, --url, --title, --source-span or --reason is refused: a handle names one citation")
	return c
}

// citeFields sets every field of a cite that the SEAT types.
//
// THE ARGUMENT FOR THE CITATION goes on the record, in Cite.text — why this source backs this
// sentence. It is not printed in the report (the note and the Bibliography print the title, and a
// seat's argument there is exactly the run-voice the report refuses); the `evidence` view shows it
// beside the source, where red decides what to verify. Set only when given, so "no argument
// offered" stays distinct from an empty one.
// citeAsserted is what a citation states, and so what a retry under its key must state again: the
// source and the span relied on in it.
var citeAsserted = []seat.Asserted{
	{Flag: flags.Quote, Field: "location"},
	{Flag: flags.URL, Field: "url"},
	{Flag: flags.Title, Field: "title"},
	{Flag: flags.SourceSpan, Field: "source_span"},
	{Flag: flags.Reason, Field: "text"},
}

func citeFields(cmd *cobra.Command, body *recordpb.Cite, why string) {
	body.Location = proto.String(seat.Location(cmd))
	body.Url = proto.String(seat.Str(cmd, flags.URL))
	body.Title = proto.String(seat.Str(cmd, flags.Title))
	body.CiteKey = proto.String(seat.Str(cmd, flags.Key))
	if seat.Given(cmd, flags.Reason) {
		body.Text = proto.String(why)
	}
}

// citeRefusal is the refusal `cite` gives for a quote Attach will not place, on a first call and on
// the retry that places the anchor a first call left owed.
func citeRefusal(err error) error {
	switch {
	case errors.Is(err, anchortext.ErrMisQuote):
		return feov.Errorf(feov.Validation, "the quoted content was not found in report.md — quote the EXACT sentence you are citing (via --quote) — the whole string is matched, so a section heading prepended to it matches nothing")
	case errors.Is(err, anchortext.ErrInFence):
		return feov.Errorf(feov.Validation, "the quote resolves inside a code fence — cite a prose sentence, not code")
	}
	return anchortext.Refusal(err)
}

type spanAt struct {
	origin  recordpb.SourceTextOrigin
	located bool
	loc     fetchcache.SpanLocation
}

func (a spanAt) pages() []int32 {
	var out []int32
	for _, p := range a.loc.Pages {
		out = append(out, int32(p))
	}
	return out
}

// locateSourceSpan stamps a cite's text origin and finds its span in the source's cached text. A
// citation is the claim that the source SAYS this, so a source whose text the tool holds is cited
// at words found in it: the record then carries the source's own text at the citation, and a
// lens reads that rather than blue's account of it. On a source with no text the span is refused
// rather than dropped — a flag accepted and recorded nowhere reads as a span nobody found.
func locateSourceSpan(run record.Run, entry fetchcache.Entry, url, span string) (spanAt, error) {
	origin, _, err := fetchcache.TextOrigin(run, entry)
	if err != nil {
		return spanAt{}, err
	}
	out := spanAt{origin: origin}
	given := strings.TrimSpace(span) != ""
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		if !given {
			return spanAt{}, feov.Errorf(feov.Validation, "the tool holds this source's text, so a citation of it names the words it relies on with --%s, copied from that text — it is at %s",
				flags.SourceSpan, fetchcache.SourceTextPath(run, entry, origin))
		}
		loc, lerr := fetchcache.LocateSourceSpan(run, entry, span)
		switch {
		case errors.Is(lerr, fetchcache.ErrReadingIncomplete):
			return spanAt{}, feov.Errorf(feov.Validation, "%v — `fetch --url %s` re-derives it", lerr, url)
		case lerr != nil:
			return spanAt{}, feov.Errorf(feov.Validation, "--%s %v", flags.SourceSpan, lerr)
		}
		out.located, out.loc = true, loc
	default:
		if given {
			return spanAt{}, feov.Errorf(feov.Validation, "--%s is found in the source's text, and the tool holds none for %s (%s) — cite it without one, or `fetch --url %s` to read a scan first",
				flags.SourceSpan, url, recordpb.Word(origin), url)
		}
	}
	return out, nil
}

type citeResult struct {
	Label  string `json:"label"`
	URL    string `json:"url,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	// Pages are the PDF pages the tool found an OCR span on; the report prints them at the marker.
	Pages []int32 `json:"pages,omitempty"`
	// Origin is where the source's text came from; "none" says the citation carries no span
	// because the tool holds no text to have found one in.
	Origin     string `json:"source_text_origin"`
	Idempotent bool   `json:"idempotent,omitempty"`
	// VoiceTells is ADVICE on the title, and the citation is already recorded by the time it renders.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r citeResult) Human() string {
	note := reportvoice.Note("the source's note and Bibliography entry", r.VoiceTells)
	if r.Idempotent {
		return "cite " + r.Label + " (idempotent retry — existing anchor returned)" + note
	}
	span := ""
	if r.Origin == recordpb.Word(recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE) {
		span = ". It carries no source span: the tool holds no text for this source"
	}
	return "citation recorded: " + r.Label + " — an invisible anchor at the quote, woven into the bibliography at assembly (" + r.URL + ")" + span + note
}
