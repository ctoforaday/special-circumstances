package seat

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/jsonshape"
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
var sourceLong = "SEARCH THE CACHED SOURCES' TEXT — the text a citation of one is found in: the extraction of a PDF or an HTML page, the bytes of a text type, or the OCR reading of a scan, page by page. It reads; it records nothing and refuses no act.\n\n" +
	"NAME ONE SOURCE with --sha (its sha256, as the fetch summary and `show evidence` print it) or --anchor (a citation's id), or neither to search every source in the run's cache. Say WHAT TO FIND with --match (an RE2 regex), --quote (literal text) or --search (a query to rank by). --match and --quote match case-insensitively, against the text as the citation locator reads it: every run of whitespace, line breaks included, is one space, and curly quotes, en and em dashes, an ellipsis and the ligatures read as their keyboard forms — so a phrase broken across lines is found, and a ' you type matches the page's ’. A page's text is what you search and quote: " + fetchcache.TextMarkup + ". To --match a caret, write \\^.\n\n" +
	"EACH HIT carries `matched`, the source's bytes; `start` and `end`, their byte offsets in the text whose sha256 is `text_sha` (on an OCR source, the reading of `page`) — the coordinates a citation records; `line` and `last_line`, the lines of that text it starts and ends on, which `cite --source-lines` takes with that `text_sha`; and `sentence`, the source's sentence the match sits in, between `sentence_start` and `sentence_end`. `source_span` is that sentence when blue's `cite --source-span` takes it as printed — the source holds it once and it starts and ends at a boundary — and empty otherwise, with `not_citeable` saying why: quote a whole clause of it instead. Across sources, each hit also carries the source's `sha256` and `url`.\n\n" +
	"UNDER --match OR --quote, EVERY MATCH IS LISTED, in text order; nothing is ranked or cut. `selection` says what was asked (`criterion`) and how many matches the text holds (`matched`) — a count far larger than you can read is the signal to narrow the pattern; across sources it also says how many sources were searched and which held no text to search. UNDER --search, the hits are the top few passages RANKED by hybrid lexical (BM25) and vector relevance to the query, for when you do not know the source's wording: the list is cut, and `search_scope` says so — a passage absent from it proves nothing.\n\n" +
	"An unknown source, a source the tool holds no text for, an empty pattern, a pattern that matches the empty string, and --search beside --match or --quote are refused.\n\n" +
	"WITH NEITHER --sha NOR --anchor, --match and --quote print across sources (JSON): " + jsonshape.Tree(fetchcache.CrossSourceSearch{}) + "; with one, the OUTPUT below.\n\n" +
	"UNDER --search (JSON), on one source: " + jsonshape.Tree(fetchcache.RankedSourceSearch{}) + "; across sources: " + jsonshape.Tree(fetchcache.RankedCrossSourceSearch{})

// sourceSearchTopK is how many passages a ranked source search returns.
const sourceSearchTopK = 10

// addSourceFlags declares what `show source` reads: the source, if one, and the pattern or query.
func addSourceFlags(c *cobra.Command) {
	c.Flags().String(flags.Sha, "", "the cached source's `sha256`, as the fetch summary and the evidence view print it. Pass this or --anchor, or neither to search every cached source")
	c.Flags().Var(flags.CitationAnchor(), flags.Anchor, "the citation whose source to search, by its anchor `id`. Pass this or --sha, or neither to search every cached source")
	c.Flags().Lookup(flags.Match).Usage = "find every place the source's text matches this `regex` (RE2, case-insensitive; a bounded repeat such as .{0,80} is fine). Every match is listed and the total is stated; nothing is ranked or cut. Pass this, --quote or --search"
	c.Flags().Lookup(flags.Quote).Usage = "find every place the source's text holds this `text` LITERALLY (case-insensitive, in the source's normalisation). Use this rather than --match whenever the text has (), ., *, ?, ^ or [] in it. Pass this, --match or --search"
	flags.Text(c, flags.Search, "RANK the source's passages by hybrid lexical (BM25) and vector relevance to this `query` and return the top few — for when you do not know the source's wording. The list is cut and a passage absent from it proves nothing; to prove text absent, or to see every place it occurs, use --quote or --match. Refused beside --match and --quote")
}

// showSource prints a search of one cached source's text, or of every cached source's.
func showSource(cmd *cobra.Command, run record.Run) error {
	sha, at := Str(cmd, flags.Sha), Str(cmd, flags.Anchor)
	switch {
	case sha != "" && at != "":
		return feov.Errorf(feov.Validation, "--sha and --anchor each name the source to search — pass one")
	case at != "":
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
	query, err := searchOf(cmd)
	if err != nil {
		return err
	}
	match, quote := Str(cmd, flags.Match), Str(cmd, flags.Quote)
	if match == "" && quote == "" && query == "" {
		return feov.Errorf(feov.MissingField, "name what to find: --search (ranked hybrid), --quote (exhaustive literal), or --match (exhaustive regex), and optionally scope to one source with --sha or --anchor")
	}
	var res any
	if sha == "" {
		if query != "" {
			res, err = fetchcache.SearchCrossSourcesHybrid(run, query, sourceSearchTopK)
		} else {
			res, err = fetchcache.SearchCrossSources(run, match, quote)
		}
		if err != nil {
			return sourceSearchRefusal(err, "", "")
		}
	} else {
		e, ok, lerr := fetchcache.LookupSha(run, sha)
		if lerr != nil {
			return lerr
		}
		if !ok {
			return feov.Errorf(feov.Validation, "no source with sha256 %s is in this run's cache — `fetch --url` reads one into it, and its summary prints the sha256", sha)
		}
		if query != "" {
			res, err = fetchcache.SearchSourceHybrid(run, e, query, sourceSearchTopK)
		} else {
			res, err = fetchcache.SearchSource(run, e, match, quote)
		}
		if err != nil {
			return sourceSearchRefusal(err, sha, e.URL)
		}
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(b.Bytes())
	return err
}

// sourceSearchRefusal says why a search of a source found nothing to search, and what reads it.
func sourceSearchRefusal(err error, sha, url string) error {
	switch {
	case errors.Is(err, fetchcache.ErrPatternMatchesEmpty):
		return feov.Errorf(feov.Validation, "%v: it would match at every position of the text — give a pattern that must match at least one character", err)
	case errors.Is(err, fetchcache.ErrNoText):
		return feov.Errorf(feov.Validation, "%v — the tool holds no text for %s, so there is nothing to search; a scan is read by `fetch --url %s`", err, sha, url)
	case errors.Is(err, fetchcache.ErrReadingIncomplete):
		return feov.Errorf(feov.Validation, "%v — `fetch --url %s` re-derives it", err, url)
	}
	return err
}
