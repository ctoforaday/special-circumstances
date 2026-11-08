package seat

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// ReadsSources reports whether a role reads a SOURCE, and so carries `fetch` and `show source`.
//
// A lane reads what it researches, blue reads what it cites, and a lens reads the same cached bytes
// to read a citation's source, or a source it found, against the sentence. The chair and the bench are handed no source to read: the
// chair reads the report, the board and the changes since a stale area's pin, and the bench's
// ruling basis is confined to the closings, the transcript and the artifacts as the record renders
// them. A verb on a surface is a thing its seat reaches for, and every seat reads its whole
// surface at its opening act, so a read no duty asks for is cost on every sitting.
//
// THE TREE IS THE GATE. Neither read writes an event, so there is no write path to refuse them on:
// a seat whose tree does not mount one is answered by the wrong-seat refusal, which names the seats
// that hold it.
func ReadsSources(role string) bool {
	return role == "lens" || role == "blue" || role == seatclass.Lane.Role
}

// sourceView is the projection that searches one cached source's text.
const sourceView = "source"

// A SOURCE IS SEARCHED HERE, NOT IN THE SHELL. m36's seats searched cache/<sha>.txt with keyword-in-
// context patterns (`.{0,80}word.{0,80}`), which the seats' grep refuses, nine calls in seven seats;
// and a hit the shell finds says nothing about whether cite takes the words around it.
var sourceLong = "SEARCH ONE CACHED SOURCE'S TEXT OR ALL CACHED SOURCES — the text a citation of it is found in: the extraction of a PDF or an HTML page, the bytes of a text type, or the OCR reading of a scan, page by page. It reads; it records nothing and refuses no act.\n\n" +
	"SCOPE TO ONE SOURCE with --sha (its sha256, as the fetch summary and `show evidence` print it) or --anchor (a citation's id), or omit both to search across all cached sources in the run. WHAT TO FIND with --search (ranked top-K hybrid BM25 + dense vector retrieval) or --quote (exhaustive literal text). Both match case-insensitively, against the text as the citation locator reads it: every run of whitespace, line breaks included, is one space, and curly quotes, en and em dashes, an ellipsis and the ligatures read as their keyboard forms — so a phrase broken across lines is found, and a ' you type matches the page's ’. A page's text is what you search and quote: " + fetchcache.TextMarkup + ".\n\n" +
	"EACH HIT carries `matched`, the source's bytes; `start` and `end`, their byte offsets in the text whose sha256 is `text_sha` (on an OCR source, the reading of `page`) — the coordinates a citation records; and `sentence`, the source's sentence the match sits in, between `sentence_start` and `sentence_end`. `source_span` is that sentence when blue's `cite --source-span` takes it as printed — the source holds it once and it starts and ends at a boundary — and empty otherwise, with `not_citeable` saying why: quote a whole clause of it instead.\n\n" +
	"A ranked --search returns a cut top-K list with `search_scope` describing what was searched and which sources had no text; absence proves nothing. An exhaustive --quote returns every match with `selection` stating the total.\n\n" +
	"An unknown source, a source the tool holds no text for, an empty pattern, and mutually exclusive flag combinations are refused."

// addSourceFlags declares what `show source` reads: the source scoping and the pattern.
func addSourceFlags(c *cobra.Command) {
	c.Flags().String(flags.Sha, "", "the cached source's `sha256`, as the fetch summary and the evidence view print it. Pass this or --anchor, or omit both to search across all cached sources")
	c.Flags().Var(flags.CitationAnchor(), flags.Anchor, "the citation whose source to search, by its anchor `id`. Pass this or --sha, or omit both to search across all cached sources")
	c.Flags().Lookup(flags.Quote).Usage = "find every place the source's text holds this `text` LITERALLY (case-insensitive, in the source's normalisation). Every match is returned; nothing is ranked or cut. Pass this or --search"
	flags.Text(c, flags.Search, "ranked top-K hybrid (BM25 + dense vector) retrieval over cached sources. The list is cut and absence proves nothing; for an exhaustive literal match set, use --quote. Pass this or --quote")
}

// showSource prints a search of one cached source's text or all cached sources.
func showSource(cmd *cobra.Command, run record.Run) error {
	sha, at := Str(cmd, flags.Sha), Str(cmd, flags.Anchor)
	if sha != "" && at != "" {
		return feov.Errorf(feov.Validation, "--sha and --anchor each name the source to search — pass one")
	}

	quote, search := Str(cmd, flags.Quote), Str(cmd, flags.Search)
	switch {
	case quote != "" && search != "":
		return feov.Errorf(feov.Validation, "--search and --quote are mutually exclusive: --quote returns every literal match; --search returns a cut ranked list. Pass one")
	case quote == "" && search == "":
		return feov.Errorf(feov.MissingField, "name what to find: --search (ranked hybrid) or --quote (exhaustive literal), and optionally scope to one source with --sha or --anchor")
	}

	var res any
	if sha != "" || at != "" {
		if at != "" {
			ev, err := record.EvidenceAt(run, at)
			if err != nil {
				return feov.Errorf(feov.Validation, "%v", err)
			}
			if ev.Source == nil {
				return feov.Errorf(feov.Validation, "%s names no citation on this record, so it names no source to search", at)
			}
			if ev.Source.Sha256 == "" {
				return feov.Errorf(feov.Validation, "%s cites %s, and the record holds no copy of that source, so the tool holds no text of it to search", at, ev.Source.URL)
			}
			sha = ev.Source.Sha256
		}
		e, ok, err := fetchcache.LookupSha(run, sha)
		if err != nil {
			return err
		}
		if !ok {
			return feov.Errorf(feov.Validation, "no source with sha256 %s is in this run's cache — `fetch --url` reads one into it, and its summary prints the sha256", sha)
		}
		if quote != "" {
			sres, err := fetchcache.SearchSource(run, e, quote)
			if err != nil {
				return handleSearchErr(err, sha, e.URL)
			}
			res = sres
		} else {
			sres, err := fetchcache.SearchSourceHybrid(run, e, search, 10)
			if err != nil {
				return handleSearchErr(err, sha, e.URL)
			}
			res = sres
		}
	} else {
		if quote != "" {
			cres, err := fetchcache.SearchCrossSourcesQuote(run, quote)
			if err != nil {
				return handleSearchErr(err, "", "")
			}
			res = cres
		} else {
			cres, err := fetchcache.SearchCrossSourcesHybrid(run, search, 10)
			if err != nil {
				return handleSearchErr(err, "", "")
			}
			res = cres
		}
	}

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	_, err := cmd.OutOrStdout().Write(b.Bytes())
	return err
}

func handleSearchErr(err error, sha, url string) error {
	switch {
	case errors.Is(err, fetchcache.ErrPatternMatchesEmpty):
		return feov.Errorf(feov.Validation, "%v: it would match at every position of the text — give a pattern that must match at least one character", err)
	case errors.Is(err, fetchcache.ErrNoText):
		return feov.Errorf(feov.Validation, "%v — the tool holds no text for %s, so there is nothing to search; a scan is read by `fetch --url %s`", err, sha, url)
	case errors.Is(err, fetchcache.ErrReadingIncomplete):
		return feov.Errorf(feov.Validation, "%v — `fetch --url %s` re-derives it", err, url)
	default:
		return err
	}
}
