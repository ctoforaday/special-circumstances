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
// to verify the citation, or a source of its own to corroborate one. The chair and the bench are handed no source to read: the
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
var sourceLong = "SEARCH ONE CACHED SOURCE'S TEXT — the text a citation of it is found in: the extraction of a PDF or an HTML page, the bytes of a text type, or the OCR reading of a scan, page by page. It reads; it records nothing and refuses no act.\n\n" +
	"NAME THE SOURCE with --sha (its sha256, as the fetch summary and `show evidence` print it) or --anchor (a citation's id), and WHAT TO FIND with --match (an RE2 regex) or --quote (literal text). Both match case-insensitively, against the text as the citation locator reads it: every run of whitespace, line breaks included, is one space, and curly quotes, en and em dashes, an ellipsis and the ligatures read as their keyboard forms — so a phrase broken across lines is found, and a ' you type matches the page's ’.\n\n" +
	"EACH HIT carries `matched`, the source's bytes; `start` and `end`, their byte offsets in the text whose sha256 is `text_sha` (on an OCR source, the reading of `page`) — the coordinates a citation records; and `sentence`, the source's sentence the match sits in, between `sentence_start` and `sentence_end`. `source_span` is that sentence when blue's `cite --source-span` takes it as printed — the source holds it once and it starts and ends at a boundary — and empty otherwise, with `not_citeable` saying why: quote a whole clause of it instead.\n\n" +
	"EVERY MATCH IS LISTED, in text order; nothing is ranked or cut. `selection` says what was asked (`criterion`) and how many matches the text holds (`matched`) — a count far larger than you can read is the signal to narrow the pattern.\n\n" +
	"An unknown source, a source the tool holds no text for, an empty pattern, and a pattern that matches the empty string are refused."

// addSourceFlags declares what `show source` reads: the source and the pattern.
func addSourceFlags(c *cobra.Command) {
	c.Flags().String(flags.Sha, "", "the cached source's `sha256`, as the fetch summary and the evidence view print it. Pass this or --anchor")
	c.Flags().Var(flags.CitationAnchor(), flags.Anchor, "the citation whose source to search, by its anchor `id`. Pass this or --sha")
	c.Flags().Lookup(flags.Match).Usage = "find every place the source's text matches this `regex` (RE2, case-insensitive; a bounded repeat such as .{0,80} is fine). Every match is listed and the total is stated; nothing is ranked or cut. Pass this or --quote"
	c.Flags().Lookup(flags.Quote).Usage = "find every place the source's text holds this `text` LITERALLY (case-insensitive, in the source's normalisation). Use this rather than --match whenever the text has (), ., *, ? or [] in it. Pass this or --match"
}

// showSource prints a search of one cached source's text.
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
		sha = ev.Source.Sha256
	case sha == "":
		return feov.Errorf(feov.MissingField, "name the source to search: --sha (its sha256, from the fetch summary or `show evidence`) or --anchor (a citation's id)")
	}
	if _, err := SelectorOf(cmd); err != nil {
		return err
	}
	match, quote := Str(cmd, flags.Match), Str(cmd, flags.Quote)
	if match == "" && quote == "" {
		return feov.Errorf(feov.MissingField, "name what to find with --match (a regex) or --quote (literal text); an empty pattern finds nothing")
	}
	e, ok, err := fetchcache.LookupSha(run, sha)
	if err != nil {
		return err
	}
	if !ok {
		return feov.Errorf(feov.Validation, "no source with sha256 %s is in this run's cache — `fetch --url` reads one into it, and its summary prints the sha256", sha)
	}
	res, err := fetchcache.SearchSource(run, e, match, quote)
	switch {
	case errors.Is(err, fetchcache.ErrPatternMatchesEmpty):
		return feov.Errorf(feov.Validation, "%v: it would match at every position of the text — give a pattern that must match at least one character", err)
	case errors.Is(err, fetchcache.ErrNoText):
		return feov.Errorf(feov.Validation, "%v — the tool holds no text for %s, so there is nothing to search; a scan is read by `fetch --url %s`", err, sha, e.URL)
	case errors.Is(err, fetchcache.ErrReadingIncomplete):
		return feov.Errorf(feov.Validation, "%v — `fetch --url %s` re-derives it", err, e.URL)
	case err != nil:
		return err
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
