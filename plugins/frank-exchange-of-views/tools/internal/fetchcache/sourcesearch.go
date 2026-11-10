package fetchcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hybridsearch"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// SEARCHING A SOURCE IS A READ OF THE TEXT A CITATION IS FOUND IN, so it reads that text the way
// LocateSourceSpan does: the same texts (each page of an OCR reading, or the embedded text), in the
// same normalisation (whitespace runs as one space, typographic characters as their keyboard
// forms). A seat that found a passage with its shell's grep searched the raw bytes, line by line,
// and then had to work out what `cite --source-span` would accept; m36's seats wrote keyword-in-
// context patterns (`.{0,80}word.{0,80}`) that the seats' grep refuses outright, nine calls in seven
// seats. A hit here carries the source's sentence it sits in and whether cite takes that sentence,
// answered by the locator cite runs, so the read and the act cannot disagree.
//
// IT GATES NOTHING. It writes no event and refuses no act; a hit is information, and the seat still
// chooses the words it cites.

// ErrPatternMatchesEmpty is a pattern that matches the empty string: it would match at every
// position of the text, which is no question about it.
var ErrPatternMatchesEmpty = errors.New("the pattern matches the empty string")

// SourceSearch is a search of one cached source's text.
type SourceSearch struct {
	// Sha is the cached document's sha256, and URL the address it was fetched from.
	Sha string `json:"sha256"`
	URL string `json:"url"`
	// Origin is where the text searched came from — embedded, or OCR-derived — and TextPath the file
	// that holds it.
	Origin   string `json:"source_text_origin"`
	TextPath string `json:"text_path"`
	// Hits are every match, in text order.
	Hits []SourceHit `json:"hits"`
	// Selection says what was asked and how many matches the text holds.
	Selection SourceSelection `json:"selection"`
}

// RankedSourceSearch is a ranked search of one source: the top passages for a query, cut, with
// SearchScope saying so. It carries no Selection, which would claim every match is listed.
type RankedSourceSearch struct {
	Sha         string                   `json:"sha256"`
	URL         string                   `json:"url"`
	Origin      string                   `json:"source_text_origin"`
	TextPath    string                   `json:"text_path"`
	Hits        []RankedSourceHit        `json:"hits"`
	SearchScope hybridsearch.SearchScope `json:"search_scope"`
}

// RankedSourceHit is a SourceHit with the fused rank score that placed it.
type RankedSourceHit struct {
	SourceHit
	Score float32 `json:"score"`
}

// SourceSelection is what a search asked and how many matches the text holds.
type SourceSelection struct {
	Criterion string `json:"criterion"`
	Matched   int    `json:"matched"`
}

// SourceHit is one place a search matched.
type SourceHit struct {
	// Page is the PDF page of an OCR reading the match sits on; absent for embedded text.
	Page int `json:"page,omitempty"`
	// Start and End are the match's byte offsets in the text it was found in — the source's text, or
	// the page's reading — whose sha256 is TextSha: the coordinates a citation records its span in.
	Start   int    `json:"start"`
	End     int    `json:"end"`
	TextSha string `json:"text_sha"`
	// Matched is the source's bytes the pattern matched.
	Matched string `json:"matched"`
	// Line and LastLine are the lines of that text the match starts and ends on, numbered from 1 —
	// what `cite --source-lines` counts, in the text TextSha names.
	Line     int `json:"line"`
	LastLine int `json:"last_line"`
	// Sentence is the source's sentence the match sits in, its whitespace folded, between
	// SentenceStart and SentenceEnd in the same text; an ellipsis marks an edge the search stopped at
	// without finding the sentence's.
	Sentence      string `json:"sentence"`
	SentenceStart int    `json:"sentence_start"`
	SentenceEnd   int    `json:"sentence_end"`
	// SourceSpan is Sentence when `cite --source-span` takes it as typed — the source holds it once and
	// it starts and ends at a boundary — and empty when it does not, with NotCiteable saying why.
	SourceSpan  string `json:"source_span"`
	NotCiteable string `json:"not_citeable,omitempty"`
}

// CrossSourceSearch is a search across all cached sources in a run.
type CrossSourceSearch struct {
	Hits      []CrossSourceHit     `json:"hits"`
	Selection CrossSourceSelection `json:"selection"`
}

// RankedCrossSourceSearch is a ranked search across every cached source: the top passages, cut.
type RankedCrossSourceSearch struct {
	Hits        []RankedCrossSourceHit   `json:"hits"`
	SearchScope hybridsearch.SearchScope `json:"search_scope"`
}

// RankedCrossSourceHit is a CrossSourceHit with the fused rank score that placed it.
type RankedCrossSourceHit struct {
	CrossSourceHit
	Score float32 `json:"score"`
}

// CrossSourceSelection is what a cross-source literal search asked and how many sources/matches were found.
type CrossSourceSelection struct {
	Criterion         string                          `json:"criterion"`
	Matched           int                             `json:"matched"`
	SearchedSources   int                             `json:"searched_sources"`
	UnsearchedSources []hybridsearch.UnsearchedSource `json:"unsearched_sources,omitempty"`
}

// CrossSourceHit is one hit from a cross-source search.
type CrossSourceHit struct {
	Sha           string `json:"sha256"`
	URL           string `json:"url"`
	Origin        string `json:"source_text_origin"`
	TextSha       string `json:"text_sha"`
	Page          int    `json:"page,omitempty"`
	Matched       string `json:"matched"`
	Line          int    `json:"line"`
	LastLine      int    `json:"last_line"`
	Sentence      string `json:"sentence"`
	SentenceStart int    `json:"sentence_start"`
	SentenceEnd   int    `json:"sentence_end"`
	SourceSpan    string `json:"source_span"`
	NotCiteable   string `json:"not_citeable,omitempty"`
	Start         int    `json:"start"`
	End           int    `json:"end"`
}

// SearchSource finds every match of a pattern in a cached source's text, case-insensitively. match
// is an RE2 regex; quote, when given instead, is literal text and is read in the source's
// normalisation, so a ’ typed from the page matches the ’ there. Exactly one of the two is given;
// an empty one, or a pattern that matches the empty string, is refused.
func SearchSource(run record.Run, e Entry, match, quote string) (SourceSearch, error) {
	re, criterion, err := searchPattern(match, quote)
	if err != nil {
		return SourceSearch{}, err
	}
	texts, origin, _, err := sourceTexts(run, e)
	if err != nil {
		return SourceSearch{}, err
	}
	out := SourceSearch{Sha: e.Sha, URL: e.URL, Origin: recordpb.Word(origin), TextPath: SourceTextPath(run, e, origin), Hits: []SourceHit{}}
	for _, t := range texts {
		f := fold(t.text, hyphenAsIs)
		for _, m := range re.FindAllStringIndex(f.s, -1) {
			out.Hits = append(out.Hits, sourceHit(texts, t, f.at[m[0]], f.end[m[1]-1]))
		}
	}
	out.Selection = SourceSelection{Criterion: criterion, Matched: len(out.Hits)}
	return out, nil
}

// SearchCrossSources is SearchSource over every cached source in the run: every match of match (an
// RE2 regex) or quote (literal text) in each source's text, in cache order and then text order,
// with the sources that hold no text to search named rather than skipped in silence.
func SearchCrossSources(run record.Run, match, quote string) (CrossSourceSearch, error) {
	re, criterion, err := searchPattern(match, quote)
	if err != nil {
		return CrossSourceSearch{}, err
	}

	entries, err := AllEntries(run)
	if err != nil {
		return CrossSourceSearch{}, err
	}

	hits := []CrossSourceHit{}
	searchedSources := 0
	var unsearchedSources []hybridsearch.UnsearchedSource

	for _, e := range entries {
		texts, origin, _, err := sourceTexts(run, e)
		if err != nil {
			unsearchedSources = append(unsearchedSources, hybridsearch.UnsearchedSource{
				Sha:    e.Sha,
				URL:    e.URL,
				Reason: unsearchedReason(e),
			})
			continue
		}
		searchedSources++
		originWord := recordpb.Word(origin)
		for _, t := range texts {
			f := fold(t.text, hyphenAsIs)
			for _, m := range re.FindAllStringIndex(f.s, -1) {
				sh := sourceHit(texts, t, f.at[m[0]], f.end[m[1]-1])
				hits = append(hits, CrossSourceHit{
					Sha:           e.Sha,
					URL:           e.URL,
					Origin:        originWord,
					TextSha:       sh.TextSha,
					Page:          sh.Page,
					Matched:       sh.Matched,
					Line:          sh.Line,
					LastLine:      sh.LastLine,
					Sentence:      sh.Sentence,
					SentenceStart: sh.SentenceStart,
					SentenceEnd:   sh.SentenceEnd,
					SourceSpan:    sh.SourceSpan,
					NotCiteable:   sh.NotCiteable,
					Start:         sh.Start,
					End:           sh.End,
				})
			}
		}
	}

	return CrossSourceSearch{
		Hits: hits,
		Selection: CrossSourceSelection{
			Criterion:         criterion,
			Matched:           len(hits),
			SearchedSources:   searchedSources,
			UnsearchedSources: unsearchedSources,
		},
	}, nil
}

// SearchSourceHybrid executes ranked top-K hybrid (BM25 + dense vector RRF) retrieval over one cached source.
func SearchSourceHybrid(run record.Run, e Entry, query string, topK int) (RankedSourceSearch, error) {
	if strings.TrimSpace(query) == "" {
		return RankedSourceSearch{}, errors.New("the search query is blank")
	}
	if topK <= 0 {
		topK = 10
	}
	texts, origin, rec, err := sourceTexts(run, e)
	if err != nil {
		return RankedSourceSearch{}, err
	}
	emb, err := vecembed.Default()
	if err != nil {
		return RankedSourceSearch{}, err
	}
	defer emb.Close()

	fullText, pageMap, textSha, err := fullSourceText(run, e, origin, rec, texts)
	if err != nil {
		return RankedSourceSearch{}, err
	}

	cachedChunks, err := loadOrComputeChunks(run, emb, textSha, fullText, pageMap)
	if err != nil {
		return RankedSourceSearch{}, err
	}

	qVecs, err := emb.Embed([]string{query})
	if err != nil {
		return RankedSourceSearch{}, err
	}

	items := make([]hybridsearch.Item, len(cachedChunks))
	for i, c := range cachedChunks {
		items[i] = hybridsearch.Item{
			Index:  i,
			Text:   fullText[c.StartByte:c.EndByte],
			Vector: c.Vector,
		}
	}

	scored, err := hybridsearch.FuseRRF(items, query, qVecs[0], topK)
	if err != nil {
		return RankedSourceSearch{}, err
	}

	out := RankedSourceSearch{
		Sha:      e.Sha,
		URL:      e.URL,
		Origin:   recordpb.Word(origin),
		TextPath: SourceTextPath(run, e, origin),
		Hits:     make([]RankedSourceHit, 0, len(scored)),
		SearchScope: hybridsearch.SearchScope{
			RankedAndCut:   true,
			Returned:       len(scored),
			SearchedChunks: len(cachedChunks),
			Notice:         hybridsearch.SearchNotice,
		},
	}

	for _, sc := range scored {
		c := cachedChunks[sc.Index]
		chunkText := fullText[c.StartByte:c.EndByte]
		hit := hitForChunk(texts, origin, rec, pageMap, fullText, c, chunkText, query)
		out.Hits = append(out.Hits, RankedSourceHit{SourceHit: hit, Score: sc.Score})
	}

	return out, nil
}

// SearchCrossSourcesHybrid executes ranked top-K hybrid retrieval across all cached sources in the run.
func SearchCrossSourcesHybrid(run record.Run, query string, topK int) (RankedCrossSourceSearch, error) {
	if strings.TrimSpace(query) == "" {
		return RankedCrossSourceSearch{}, errors.New("the search query is blank")
	}
	if topK <= 0 {
		topK = 10
	}
	emb, err := vecembed.Default()
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}
	defer emb.Close()

	qVecs, err := emb.Embed([]string{query})
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}

	entries, err := AllEntries(run)
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}

	type sourceChunkRef struct {
		entry    Entry
		texts    []spanText
		origin   recordpb.SourceTextOrigin
		rec      ReadingRecord
		pageMap  []int
		fullText string
		chunk    hybridsearch.CachedChunk
	}

	var allRefs []sourceChunkRef
	var items []hybridsearch.Item
	searchedSources := 0
	var unsearchedSources []hybridsearch.UnsearchedSource

	for _, e := range entries {
		texts, origin, rec, err := sourceTexts(run, e)
		if err != nil {
			unsearchedSources = append(unsearchedSources, hybridsearch.UnsearchedSource{
				Sha:    e.Sha,
				URL:    e.URL,
				Reason: unsearchedReason(e),
			})
			continue
		}
		fullText, pageMap, textSha, err := fullSourceText(run, e, origin, rec, texts)
		if err != nil {
			unsearchedSources = append(unsearchedSources, hybridsearch.UnsearchedSource{
				Sha:    e.Sha,
				URL:    e.URL,
				Reason: err.Error(),
			})
			continue
		}
		cachedChunks, err := loadOrComputeChunks(run, emb, textSha, fullText, pageMap)
		if err != nil {
			unsearchedSources = append(unsearchedSources, hybridsearch.UnsearchedSource{
				Sha:    e.Sha,
				URL:    e.URL,
				Reason: err.Error(),
			})
			continue
		}
		searchedSources++
		for _, c := range cachedChunks {
			allRefs = append(allRefs, sourceChunkRef{
				entry:    e,
				texts:    texts,
				origin:   origin,
				rec:      rec,
				pageMap:  pageMap,
				fullText: fullText,
				chunk:    c,
			})
			items = append(items, hybridsearch.Item{
				Index:  len(items),
				Text:   fullText[c.StartByte:c.EndByte],
				Vector: c.Vector,
			})
		}
	}

	scored, err := hybridsearch.FuseRRF(items, query, qVecs[0], topK)
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}

	hits := make([]RankedCrossSourceHit, 0, len(scored))
	for _, sc := range scored {
		ref := allRefs[sc.Index]
		chunkText := ref.fullText[ref.chunk.StartByte:ref.chunk.EndByte]
		sh := hitForChunk(ref.texts, ref.origin, ref.rec, ref.pageMap, ref.fullText, ref.chunk, chunkText, query)
		hits = append(hits, RankedCrossSourceHit{Score: sc.Score, CrossSourceHit: CrossSourceHit{
			Sha:           ref.entry.Sha,
			URL:           ref.entry.URL,
			Origin:        recordpb.Word(ref.origin),
			TextSha:       sh.TextSha,
			Page:          sh.Page,
			Matched:       sh.Matched,
			Line:          sh.Line,
			LastLine:      sh.LastLine,
			Sentence:      sh.Sentence,
			SentenceStart: sh.SentenceStart,
			SentenceEnd:   sh.SentenceEnd,
			SourceSpan:    sh.SourceSpan,
			NotCiteable:   sh.NotCiteable,
			Start:         sh.Start,
			End:           sh.End,
		}})
	}

	return RankedCrossSourceSearch{
		Hits: hits,
		SearchScope: hybridsearch.SearchScope{
			RankedAndCut:      true,
			Returned:          len(hits),
			SearchedChunks:    len(items),
			SearchedSources:   searchedSources,
			UnsearchedSources: unsearchedSources,
			Notice:            hybridsearch.SearchNotice,
		},
	}, nil
}

func fullSourceText(run record.Run, e Entry, origin recordpb.SourceTextOrigin, rec ReadingRecord, texts []spanText) (string, []int, string, error) {
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		path := OCRTextPath(run, e.Sha)
		b, err := os.ReadFile(path)
		if err != nil {
			return "", nil, "", err
		}
		fullText := string(b)
		pageMap := make([]int, len(texts))
		offset := 0
		for i, t := range texts {
			pageMap[i] = offset
			offset += len(t.text) + 2 // "\n\n" separator in AssembleReading
		}
		return fullText, pageMap, rec.TextSha, nil
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED:
		if len(texts) == 0 {
			return "", nil, "", ErrNoText
		}
		return texts[0].text, nil, texts[0].sha, nil
	default:
		return "", nil, "", fmt.Errorf("%w: %s", ErrNoText, recordpb.Word(origin))
	}
}

func loadOrComputeChunks(run record.Run, emb vecembed.Embedder, textSha, fullText string, pageMap []int) ([]hybridsearch.CachedChunk, error) {
	cachePath := VectorCachePath(run, textSha)
	cached, err := hybridsearch.ReadCache(cachePath, emb.ModelID())
	if err == nil {
		return cached, nil
	}
	chunks := hybridsearch.ChunkSource(fullText, pageMap)
	if len(chunks) == 0 {
		return nil, nil
	}
	chunkTexts := make([]string, len(chunks))
	for i, ch := range chunks {
		chunkTexts[i] = ch.Text
	}
	vecs, err := emb.Embed(chunkTexts)
	if err != nil {
		return nil, err
	}
	cached = make([]hybridsearch.CachedChunk, len(chunks))
	for i, ch := range chunks {
		cached[i] = hybridsearch.CachedChunk{
			Index:     ch.Index,
			StartByte: ch.StartByte,
			EndByte:   ch.EndByte,
			Page:      ch.Page,
			Vector:    vecs[i],
		}
	}
	_ = hybridsearch.WriteCache(cachePath, emb.ModelID(), cached)
	return cached, nil
}

func hitForChunk(texts []spanText, origin recordpb.SourceTextOrigin, rec ReadingRecord, pageMap []int, fullText string, c hybridsearch.CachedChunk, chunkText, query string) SourceHit {
	var t spanText
	localChunkStart := c.StartByte
	localChunkEnd := c.EndByte
	if origin == recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR && len(texts) > 0 {
		pageIdx := c.Page - 1
		if pageIdx < 0 || pageIdx >= len(texts) {
			pageIdx = 0
		}
		t = texts[pageIdx]
		pageOffset := 0
		if pageIdx < len(pageMap) {
			pageOffset = pageMap[pageIdx]
		}
		localChunkStart = max(0, c.StartByte-pageOffset)
		localChunkEnd = min(len(t.text), c.EndByte-pageOffset)
	} else if len(texts) > 0 {
		t = texts[0]
	}

	wordStart, wordEnd := findBestWordInChunk(t.text, localChunkStart, localChunkEnd, query)
	return sourceHit(texts, t, wordStart, wordEnd)
}

func findBestWordInChunk(text string, chunkStart, chunkEnd int, query string) (int, int) {
	if chunkStart < 0 {
		chunkStart = 0
	}
	if chunkEnd > len(text) {
		chunkEnd = len(text)
	}
	if chunkStart >= chunkEnd {
		return chunkStart, chunkEnd
	}
	slice := text[chunkStart:chunkEnd]
	words := strings.Fields(query)
	for _, w := range words {
		wClean := strings.TrimFunc(w, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		if len(wClean) < 2 {
			continue
		}
		idx := strings.Index(strings.ToLower(slice), strings.ToLower(wClean))
		if idx >= 0 {
			start := chunkStart + idx
			end := start + len(wClean)
			return start, end
		}
	}
	inWord := false
	wStart := chunkStart
	for i := chunkStart; i < chunkEnd; {
		r, size := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			if inWord {
				return wStart, i
			}
		} else {
			if !inWord {
				wStart = i
				inWord = true
			}
		}
		i += size
	}
	if inWord {
		return wStart, chunkEnd
	}
	return chunkStart, chunkEnd
}

// sourceHit is the hit at [start, end) of t, with the sentence it sits in and whether cite takes it.
func sourceHit(texts []spanText, t spanText, start, end int) SourceHit {
	a, z, cutA, cutZ := sentenceBounds(t.text, start, end)
	// The bounds stop at the text's own edges, which can be whitespace; the sentence is its words.
	for a < z && unicode.IsSpace(runeAt(t.text, a)) {
		a += utf8.RuneLen(runeAt(t.text, a))
	}
	for z > a && unicode.IsSpace(runeBefore(t.text, z)) {
		z -= utf8.RuneLen(runeBefore(t.text, z))
	}
	h := SourceHit{Page: t.page, Start: start, End: end, TextSha: t.sha, Matched: t.text[start:end],
		Line: lineAt(t.text, start), LastLine: lineAt(t.text, max(start, end-1)),
		Sentence: sentenceAround(t.text, start, end), SentenceStart: a, SentenceEnd: z}
	if cutA || cutZ {
		h.NotCiteable = fmt.Sprintf("the source marks no sentence edge within %d bytes of the match; cite its lines instead", spanReach)
		return h
	}
	sentence := strings.Join(strings.Fields(t.text[a:z]), " ")
	// The sentence is this text's own words at [a, z), so the locator finds it here: once, and then
	// it is cited at these offsets, or more than once. Found once, it can still end where cite reads
	// no sentence end — "e.g." before a lowercase word.
	_, _, err := locateOnce(texts, sentence)
	switch {
	case errors.Is(err, ErrSpanRepeated):
		h.NotCiteable = "the source holds this sentence more than once"
	case err != nil:
		h.NotCiteable = "the sentence does not start and end at a boundary cite takes"
	default:
		h.SourceSpan = sentence
	}
	return h
}

func unsearchedReason(e Entry) string {
	if e.TextReason != "" {
		if strings.Contains(e.TextReason, "scanned or image-only") {
			return "image_only: no OCR reading yet; run fetch --url " + e.URL
		}
		return e.TextReason
	}
	if e.TextExtracted != nil && !*e.TextExtracted {
		if e.ContentType == "application/pdf" {
			return "image_only: no OCR reading yet; run fetch --url " + e.URL
		}
		return "no text extracted from document"
	}
	return "the source has no extracted text"
}

// VectorCachePath is where dense vector embeddings of textSha's chunks are cached.
func VectorCachePath(run record.Run, textSha string) string {
	return filepath.Join(Dir(run), textSha+".vec")
}

// searchPattern compiles a search's pattern, case-insensitive, and names it.
func searchPattern(match, quote string) (*regexp.Regexp, string, error) {
	var re *regexp.Regexp
	var criterion string
	switch {
	case match != "" && quote != "":
		return nil, "", errors.New("a search takes a regex or a literal, not both")
	case quote != "":
		want := fold(quote, hyphenAsIs).s
		if want == "" {
			return nil, "", fmt.Errorf("%w: the quote is blank", ErrPatternMatchesEmpty)
		}
		re, criterion = regexp.MustCompile("(?i)"+regexp.QuoteMeta(want)), `the quote "`+quote+`"`
	case match != "":
		var err error
		if re, err = regexp.Compile("(?i)" + match); err != nil {
			return nil, "", err
		}
		criterion = "/" + match + "/"
	default:
		return nil, "", fmt.Errorf("%w: no pattern was given", ErrPatternMatchesEmpty)
	}
	if re.MatchString("") {
		return nil, "", ErrPatternMatchesEmpty
	}
	return re, criterion, nil
}
