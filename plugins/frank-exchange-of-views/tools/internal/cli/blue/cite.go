package blue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// Crash-safety mirrors lens finding: the marker write is the atomic commit and the cite
// event follows, so a crash between them leaves an ORPHAN anchor the assembly surfaces as a
// dangling citation — never a wedge, never a phantom event.
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
		if prior, err := record.ExistingCiteByKey(run.Dir(), s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			return citeResult{Label: prior, Idempotent: true}, nil
		}

		// Resolve the source through the run cache (fetch-once). A FAILURE is an unusable
		// citation: reject AND auto-emit a friction event (unlike a bare `fetch` miss).
		entry, _, _, err := fetchcache.Resolve(run.Dir(), url, fetchcache.Default)
		if err != nil {
			msg := fmt.Sprintf("blue cite: could not load %s: %v — pick a reachable source or an archive.org snapshot", url, err)
			if _, ferr := record.Append(s.Identity(), &recordpb.Friction{Text: proto.String(msg)}); ferr != nil {
				return nil, ferr
			}
			return nil, errors.New(msg)
		}

		// A TORN SPLICE IS ADOPTED, NOT DOUBLED. The splice below and the event append after it
		// are two acts with no transaction over them, so a crash between them leaves an anchor
		// on this very sentence that no event backs — and retries are this record's ordinary
		// weather (mint, cite and prove all carry crash-retry keys). The retry used to look for
		// a prior EVENT, find none, and splice a second marker beside the orphan: one sentence,
		// two tokens, one of them immortal and backing nothing. If the located quote already
		// carries a citation anchor the record has never heard of, that anchor IS this cite's
		// first attempt, and the retry finishes the interrupted act instead of starting a rival.
		label := adoptTornCiteAnchor(run.Dir(), quote)
		if label == "" {
			// Mint the label UP FRONT: it forms the marker, so it must exist before the report
			// write. Append will not re-mint (the label is already in the payload).
			label = record.NewCitationID()
			marker := "<!--cite:" + label + "-->"

			// Splice the invisible anchor at the located quote UNDER THE LOCK, atomically, via
			// the shared anchor-insert (the same rule a finding is anchored by). Mis-quote or
			// in-fence -> reject; nothing is written and no cite is recorded.
			if err := record.MutateBlueReport(run.Dir(), func(old []byte) ([]byte, error) {
				next, aerr := lens.InsertAnchor(old, quote, marker)
				switch {
				case errors.Is(aerr, lens.ErrMisQuote):
					return nil, fmt.Errorf("blue cite: the quoted content was not found in report.md — quote the EXACT sentence you are citing (via --quote) — the whole string is matched, so a section heading prepended to it matches nothing")
				case errors.Is(aerr, lens.ErrInFence):
					return nil, fmt.Errorf("blue cite: the quote resolves inside a code fence — cite a prose sentence, not code")
				}
				return next, aerr
			}); err != nil {
				return nil, err
			}
		}

		// The anchor is committed; the cite event follows. access_date is engine-supplied
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
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: entry.Sha}, nil
	}))

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

// adoptTornCiteAnchor returns the label of a citation anchor already on the located quote that
// no recorded event backs — a torn splice — or "" for the ordinary fresh path. The walk itself is
// lens.OrphanAnchorAt, shared with `lens finding` and `blue prove`, which carry the same
// two-act crash window; only the recorded set is this verb's.
func adoptTornCiteAnchor(runDir, quote string) string {
	rep, err := record.ReadBlueReport(runDir)
	if err != nil {
		return ""
	}
	labels, err := record.CitationLabels(runDir)
	if err != nil {
		return "" // cannot tell an orphan from a backed anchor; splice fresh rather than guess
	}
	recorded := map[string]bool{}
	for _, l := range labels {
		recorded[l] = true
	}
	return lens.OrphanAnchorAt(string(rep), quote, "cite", func(id string) bool { return recorded[id] })
}
