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
		quote := seat.Str(cmd, flags.Quote)
		if strings.TrimSpace(quote) == "" {
			return nil, fmt.Errorf("blue cite requires --quote: the EXACT sentence to anchor the citation at, verbatim from blue/report.md and nothing else")
		}
		url := seat.Str(cmd, flags.URL)
		if strings.TrimSpace(url) == "" {
			return nil, fmt.Errorf("blue cite requires --url: the source being cited (fetched once and cached; red re-reads the same bytes)")
		}
		title := seat.Str(cmd, flags.Title)
		if strings.TrimSpace(title) == "" {
			return nil, fmt.Errorf("blue cite requires --title: the source's name as it appears in the composed bibliography")
		}

		// Crash-retry idempotency: a prior cite under this --key returns its label, no
		// second fetch AND no second anchor (BEFORE any effect).
		key := seat.Str(cmd, flags.Key)
		if prior, err := record.ExistingCiteByKey(run, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			return citeResult{Label: prior, Idempotent: true}, nil
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
		}
		// THE DEFAULT IS THE WEAK CLAIM. A citation nobody has asserted a reading for is UNREAD:
		// the honest state costs the seat nothing, and only a stronger one is stated on purpose.
		read := recordpb.SourceTextRead_SOURCE_TEXT_READ_UNREAD
		if w := seat.Str(cmd, flags.SourceText); w != "" {
			v, known := record.SourceTextReadOf(w)
			if !known || v == recordpb.SourceTextRead_SOURCE_TEXT_READ_UNSPECIFIED {
				return nil, fmt.Errorf("blue cite: %q is not a reading this record can carry (leaf | summary_only | unread)", w)
			}
			read = v
		}
		body.SourceTextRead = &read
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: entry.Sha}, nil
	}))

	enumhelp.Flag(c, flags.SourceText, record.MustEnum("cite", "source_text_read"),
		"how much of the source you actually READ. Omitted records `unread` — the citation then rests on the source EXISTING, not on anything it says")
	c.Flags().String(flags.Quote, "", flags.DescQuote+". The invisible citation anchor is spliced here, so a mis-quote is rejected rather than guessed at")
	c.Flags().String(flags.URL, "", flags.DescURL)
	c.Flags().String(flags.Title, "", flags.DescTitle)
	c.Flags().String(flags.Key, "", flags.DescKey+"; the TOOL assigns the c-<hex> label")
	return c
}

type citeResult struct {
	Label      string `json:"label"`
	URL        string `json:"url,omitempty"`
	Sha256     string `json:"sha256,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r citeResult) Human() string {
	if r.Idempotent {
		return "cite " + r.Label + " (idempotent retry — existing anchor returned)"
	}
	return "citation recorded: " + r.Label + " — an invisible immortal anchor at the quote, woven into the bibliography at assembly (" + r.URL + ")"
}
