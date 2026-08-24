package blue

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"regexp"
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
		if prior, err := record.ExistingCiteByKey(s.RunDir, s.SeatID, key); err != nil {
			return nil, err
		} else if prior != "" {
			return citeResult{Label: prior, Idempotent: true}, nil
		}

		// Resolve the source through the run cache (fetch-once). A FAILURE is an unusable
		// citation: reject AND auto-emit a friction event (unlike a bare `fetch` miss).
		sha, _, _, err := fetchcache.Resolve(s.RunDir, url, fetchcache.Default)
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
		label := adoptTornCiteAnchor(s.RunDir, quote)
		if label == "" {
			// Mint the label UP FRONT: it forms the marker, so it must exist before the report
			// write. Append will not re-mint (the label is already in the payload).
			label = record.NewCitationID()
			marker := "<!--cite:" + label + "-->"

			// Splice the invisible anchor at the located quote UNDER THE LOCK, atomically, via
			// the shared anchor-insert (the same rule a finding is anchored by). Mis-quote or
			// in-fence -> reject; nothing is written and no cite is recorded.
			if err := record.MutateBlueReport(s.RunDir, func(old []byte) ([]byte, error) {
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
			Sha256:     proto.String(sha),
			Title:      proto.String(title),
			Location:   proto.String(quote),
			AccessDate: proto.String(record.Now().Format("2006-01-02")),
			CiteKey:    proto.String(seat.Str(cmd, flags.Key)),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return citeResult{Label: label, URL: url, Sha256: sha}, nil
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

// citeToken matches a citation anchor as minted; group 1 is the label.
var citeToken = regexp.MustCompile(`<!--cite:(c-[0-9a-f]+)-->`)

// adoptTornCiteAnchor returns the label of a citation anchor already sitting on the located
// quote that NO recorded event backs — the state a crash between splice and append leaves — or
// "" when the sentence carries no such orphan and a fresh splice is the right act.
//
// THE SCOPE IS THE SENTENCE'S OWN ANCHOR RUN, not the whole report: an orphan elsewhere on the
// page is somebody else's torn act, and adopting it here would attach this cite's source to a
// sentence it never read. The walk mirrors the splice's own geometry — InsertAnchor lands the
// token between the located span and its trailing punctuation, and abutting runs are ordinary
// (two lenses anchoring one sentence is a measured corpus shape) — so both token-runs and
// punctuation are stepped over, in either order.
//
// Every miss returns "": a report that cannot be read, a quote that does not locate, a run with
// no orphan. The caller then splices fresh, and whatever refusal that path produces is the one
// the seat should see.
func adoptTornCiteAnchor(runDir, quote string) string {
	rep, err := record.ReadBlueReport(runDir)
	if err != nil {
		return ""
	}
	report := string(rep)
	_, end, err := bluedoc.LocateUnique("blue cite", report, quote)
	if err != nil {
		return ""
	}
	recorded := map[string]bool{}
	if labels, err := record.CitationLabels(runDir); err == nil {
		for _, l := range labels {
			recorded[l] = true
		}
	} else {
		return "" // cannot tell orphan from backed; splice fresh rather than guess
	}
	j := end
	for j < len(report) {
		if next := anchor.SkipRun(report, j); next > j {
			for _, m := range citeToken.FindAllStringSubmatch(report[j:next], -1) {
				if !recorded[m[1]] {
					return m[1]
				}
			}
			j = next
			continue
		}
		if strings.ContainsRune(lens.TrailingPunct, rune(report[j])) {
			j++
			continue
		}
		break
	}
	return ""
}
