package fetchcache

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hybridsearch"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// A RANKED SEARCH READS THE TEXT A SEAT READS AND CITES. Its chunks are cut from the bytes
// SourceTextPath names — the extraction, the body read as-is, or the assembled OCR reading — and
// its vectors are cached under the sha256 of exactly those bytes, so a re-read that changes the text
// changes the address and no stale vector is ever read. Prewarming at a fetch or a reading and the
// search itself reach the text through the one function (vectorTextOf), so the address one writes
// is the address the other reads.
//
// A HIT IS STATED IN ONE TEXT. Lines and offsets are counted in the text whose hash a hit carries —
// the extraction, or one page's reading — because that is what `cite --source-lines` resolves. A
// chunk of an OCR reading that crosses a page break is stated on the page its best sentence sits on:
// its matched/start/end/line/last_line are the chunk's part on that page.

// vectorText is a source's text as the ranked search reads it: the spanTexts a citation is located
// in, and the whole text SourceTextPath holds, with its sha256 and, for an OCR reading, the offset in
// it where each page's text starts.
type vectorText struct {
	texts   []spanText
	origin  recordpb.SourceTextOrigin
	full    string
	sha     string
	pageMap []int
}

// vectorTextOf is e's text for ranking. An OCR reading whose assembled file is not its page texts
// joined is refused: its chunk offsets would place hits on the wrong page.
func vectorTextOf(run record.Run, e Entry) (vectorText, error) {
	texts, origin, _, err := sourceTexts(run, e)
	if err != nil {
		return vectorText{}, err
	}
	vt := vectorText{texts: texts, origin: origin}
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		path := SourceTextPath(run, e, origin)
		b, err := os.ReadFile(path)
		if err != nil {
			return vectorText{}, fmt.Errorf("%w: the assembled reading is missing (%v)", ErrNoText, err)
		}
		pages := make([]string, len(texts))
		offset := 0
		for i, t := range texts {
			pages[i] = t.text
			vt.pageMap = append(vt.pageMap, offset)
			offset += len(t.text) + len("\n\n")
		}
		if tessocr.AssembleReading(pages) != string(b) {
			return vectorText{}, fmt.Errorf("%w: the assembled reading at %s is not its pages' texts joined; read the source again", ErrReadingIncomplete, path)
		}
		vt.full, vt.sha = string(b), Sha(b)
	default:
		// sourceTexts holds one embedded text: the bytes SourceTextPath names, and their sha256.
		vt.full, vt.sha = texts[0].text, texts[0].sha
	}
	return vt, nil
}

// VectorCachePath is where the chunk vectors of the text whose sha256 is textSha are cached.
func VectorCachePath(run record.Run, textSha string) string {
	return filepath.Join(Dir(run), textSha+".vec")
}

// ReportVecCachePath is where the report's paragraph vectors are cached.
func ReportVecCachePath(run record.Run) string {
	return filepath.Join(Dir(run), "report_vecs.bin")
}

// chunkVectors is vt's chunks with their vectors: the cache's, when it was written by this model,
// helper, target and chunker, and otherwise embedded now and cached.
func chunkVectors(run record.Run, emb vecembed.Embedder, vt vectorText) ([]hybridsearch.CachedChunk, error) {
	path := VectorCachePath(run, vt.sha)
	cached, err := hybridsearch.ReadVecCache(path, emb.ModelID(), hybridsearch.SourceChunkSpec)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, hybridsearch.ErrCacheMiss) {
		return nil, err
	}
	chunks := hybridsearch.ChunkSource(vt.full, vt.pageMap)
	if len(chunks) == 0 {
		return nil, nil
	}
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = ch.Text
	}
	vecs, err := emb.Embed(texts)
	if err != nil {
		return nil, err
	}
	if len(vecs) != len(chunks) {
		return nil, fmt.Errorf("the embedder returned %d vectors for %d chunks", len(vecs), len(chunks))
	}
	cached = make([]hybridsearch.CachedChunk, len(chunks))
	for i, ch := range chunks {
		cached[i] = hybridsearch.CachedChunk{Index: ch.Index, StartByte: ch.StartByte, EndByte: ch.EndByte, Page: ch.Page, Vector: vecs[i]}
	}
	// The cache saves the next search the embedding; a write that fails costs only that.
	_ = writeReplacing(path, hybridsearch.EncodeVecCache(emb.ModelID(), hybridsearch.SourceChunkSpec, cached))
	return cached, nil
}

// prewarmSourceVectors embeds e's text and caches its chunk vectors, so the first ranked search of
// it does not pay for them. Best-effort: with no embedder, or no text, the search embeds lazily.
func prewarmSourceVectors(run record.Run, e Entry) {
	vt, err := vectorTextOf(run, e)
	if err != nil {
		return
	}
	emb, err := vecembed.Default()
	if err != nil {
		return
	}
	defer emb.Close()
	_, _ = chunkVectors(run, emb, vt)
}

// prewarmReadingVectors is prewarmSourceVectors for the cached document a reading was just written for.
func prewarmReadingVectors(run record.Run, sha string) {
	if e, ok, err := LookupSha(run, sha); err == nil && ok {
		prewarmSourceVectors(run, e)
	}
}

// SearchSourceHybrid ranks one cached source's chunks against query by BM25 and vector similarity
// fused by RRF, and returns the top topK, cut.
func SearchSourceHybrid(run record.Run, e Entry, query string, topK int) (RankedSourceSearch, error) {
	if strings.TrimSpace(query) == "" {
		return RankedSourceSearch{}, errors.New("the search query is blank")
	}
	vt, err := vectorTextOf(run, e)
	if err != nil {
		return RankedSourceSearch{}, err
	}
	emb, err := vecembed.Default()
	if err != nil {
		return RankedSourceSearch{}, err
	}
	defer emb.Close()
	chunks, err := chunkVectors(run, emb, vt)
	if err != nil {
		return RankedSourceSearch{}, err
	}
	qv, err := queryVector(emb, query)
	if err != nil {
		return RankedSourceSearch{}, err
	}
	items := make([]hybridsearch.Item, len(chunks))
	for i, c := range chunks {
		items[i] = hybridsearch.Item{Index: i, Text: vt.full[c.StartByte:c.EndByte], Vector: c.Vector}
	}
	scored, err := hybridsearch.FuseRRF(items, query, qv, topK)
	if err != nil {
		return RankedSourceSearch{}, err
	}
	out := RankedSourceSearch{
		Sha: e.Sha, URL: e.URL, Origin: recordpb.Word(vt.origin), TextPath: SourceTextPath(run, e, vt.origin),
		Hits: make([]RankedSourceHit, 0, len(scored)),
		SearchScope: hybridsearch.SearchScope{RankedAndCut: true, Returned: len(scored), SearchedChunks: len(chunks),
			Notice: hybridsearch.SearchNotice},
	}
	for _, sc := range scored {
		out.Hits = append(out.Hits, RankedSourceHit{SourceHit: chunkHit(vt, chunks[sc.Index], query), Score: sc.Score})
	}
	return out, nil
}

// SearchCrossSourcesHybrid is SearchSourceHybrid over every cached source, ranked as one list, with
// each source that holds no text to rank named in UnsearchedSources.
func SearchCrossSourcesHybrid(run record.Run, query string, topK int) (RankedCrossSourceSearch, error) {
	if strings.TrimSpace(query) == "" {
		return RankedCrossSourceSearch{}, errors.New("the search query is blank")
	}
	entries, err := AllEntries(run)
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}
	emb, err := vecembed.Default()
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}
	defer emb.Close()

	type ref struct {
		e  Entry
		vt vectorText
		c  hybridsearch.CachedChunk
	}
	var refs []ref
	var items []hybridsearch.Item
	scope := CrossSourceScope{SearchScope: hybridsearch.SearchScope{RankedAndCut: true, Notice: hybridsearch.SearchNotice}}
	for _, e := range entries {
		vt, err := vectorTextOf(run, e)
		if err != nil {
			scope.UnsearchedSources = append(scope.UnsearchedSources, unsearched(e, err))
			continue
		}
		// An embedder that fails is the tool's failure, not the source's: the search stops.
		chunks, err := chunkVectors(run, emb, vt)
		if err != nil {
			return RankedCrossSourceSearch{}, err
		}
		scope.SearchedSources++
		for _, c := range chunks {
			refs = append(refs, ref{e: e, vt: vt, c: c})
			items = append(items, hybridsearch.Item{Index: len(items), Text: vt.full[c.StartByte:c.EndByte], Vector: c.Vector})
		}
	}
	qv, err := queryVector(emb, query)
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}
	scored, err := hybridsearch.FuseRRF(items, query, qv, topK)
	if err != nil {
		return RankedCrossSourceSearch{}, err
	}
	hits := make([]RankedCrossSourceHit, 0, len(scored))
	for _, sc := range scored {
		r := refs[sc.Index]
		hits = append(hits, RankedCrossSourceHit{Score: sc.Score, CrossSourceHit: crossSourceHit(r.e, r.vt.origin, chunkHit(r.vt, r.c, query))})
	}
	scope.Returned, scope.SearchedChunks = len(hits), len(items)
	return RankedCrossSourceSearch{Hits: hits, SearchScope: scope}, nil
}

// queryVector is query's embedding.
func queryVector(emb vecembed.Embedder, query string) ([]float32, error) {
	v, err := emb.Embed([]string{query})
	if err != nil {
		return nil, err
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("the embedder returned %d vectors for one query", len(v))
	}
	return v[0], nil
}

// crossSourceHit is h, found in e's text of origin, as a cross-source hit.
func crossSourceHit(e Entry, origin recordpb.SourceTextOrigin, h SourceHit) CrossSourceHit {
	return CrossSourceHit{
		Sha: e.Sha, URL: e.URL, Origin: recordpb.Word(origin), TextSha: h.TextSha, Page: h.Page,
		Matched: h.Matched, Line: h.Line, LastLine: h.LastLine, Sentence: h.Sentence,
		SentenceStart: h.SentenceStart, SentenceEnd: h.SentenceEnd, SourceSpan: h.SourceSpan,
		NotCiteable: h.NotCiteable, Start: h.Start, End: h.End,
	}
}

// unsearched names a source whose text could not be read for a search, and why. A PDF the extractor
// opened, counted pages in and found no text on, with no reading of its pages yet, is image_only:
// what it needs is a reading.
func unsearched(e Entry, err error) hybridsearch.UnsearchedSource {
	u := hybridsearch.UnsearchedSource{Sha: e.Sha, URL: e.URL, Reason: err.Error()}
	switch {
	case !errors.Is(err, ErrNoText):
	case e.TextExtracted != nil && !*e.TextExtracted && MediaType(e.ContentType) == "application/pdf" && e.Pages > 0 &&
		(e.NotRenderable == nil || !*e.NotRenderable):
		u.Reason = "image_only: no OCR reading yet; run fetch --url " + e.URL
	case e.TextReason != "":
		u.Reason = e.TextReason
	}
	return u
}

// segment is the part [a, z) of one text a chunk covers.
type segment struct {
	t    spanText
	a, z int
}

// segmentsOf is the chunk's part of each text it covers, in order: the whole chunk in an embedded
// text, or its part on each page of a reading.
func segmentsOf(vt vectorText, c hybridsearch.CachedChunk) []segment {
	if vt.pageMap == nil {
		return []segment{{t: vt.texts[0], a: c.StartByte, z: c.EndByte}}
	}
	var out []segment
	for i, t := range vt.texts {
		start := vt.pageMap[i]
		a, z := max(c.StartByte, start), min(c.EndByte, start+len(t.text))
		if a < z {
			out = append(out, segment{t: t, a: a - start, z: z - start})
		}
	}
	return out
}

// chunkHit is the hit a ranked chunk makes: matched/start/end/line/last_line are the chunk — its
// part on the page its best sentence sits on — and the sentence fields are that sentence's.
func chunkHit(vt vectorText, c hybridsearch.CachedChunk, query string) SourceHit {
	seg, sa, sz := bestChunkSentence(segmentsOf(vt, c), query)
	h := sourceHit(vt.texts, seg.t, sa, sz)
	h.Start, h.End, h.Matched = seg.a, seg.z, seg.t.text[seg.a:seg.z]
	h.Line, h.LastLine = lineAt(seg.t.text, seg.a), lineAt(seg.t.text, max(seg.a, seg.z-1))
	return h
}

// bestChunkSentence is the sentence of the chunk that shares the most of the query's words — after
// lowercasing, dropping stopwords and stemming — the earliest on a tie, with the segment it sits in.
// It is what a ranked hit offers to cite: cite takes one sentence, and a chunk holds several.
func bestChunkSentence(segs []segment, query string) (segment, int, int) {
	want := stems(query)
	best, bestA, bestZ, bestScore := segs[0], segs[0].a, segs[0].z, -1
	for _, s := range segs {
		for p := nextWord(s.t.text, s.a); p < s.z; {
			a, z, _, _ := sentenceBounds(s.t.text, p, p+utf8.RuneLen(runeAt(s.t.text, p)))
			score := 0
			for w := range stems(s.t.text[a:z]) {
				if want[w] {
					score++
				}
			}
			if score > bestScore {
				best, bestA, bestZ, bestScore = s, a, z, score
			}
			p = nextWord(s.t.text, max(z, p+1))
		}
	}
	return best, bestA, bestZ
}

// nextWord is the offset of the first character at or after i that is neither whitespace nor in a
// table rule (ruleTokenAt), or len(text).
func nextWord(text string, i int) int {
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		if z, ok := ruleTokenAt(text, i); ok {
			i = z
			continue
		}
		if !unicode.IsSpace(r) {
			return i
		}
		i += n
	}
	return len(text)
}

// stopwords are the words too common to say which sentence a query is about.
var stopwords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
	"for": true, "from": true, "has": true, "have": true, "in": true, "is": true, "it": true, "its": true,
	"of": true, "on": true, "or": true, "that": true, "the": true, "this": true, "to": true, "was": true,
	"were": true, "what": true, "which": true, "with": true,
}

// stems is the set of s's words, lowercased, stopwords dropped, each stemmed.
func stems(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if !stopwords[w] {
			out[stem(w)] = true
		}
	}
	return out
}

// stem strips the inflectional suffix porter strips first, so "costs", "costing" and "cost" agree.
func stem(w string) string {
	for _, suf := range []string{"ing", "ed", "es", "s"} {
		if base, ok := strings.CutSuffix(w, suf); ok && len(base) >= 3 {
			return base
		}
	}
	return w
}

// SearchReport ranks the report's paragraphs against query and returns the top topK, cut, and how
// many paragraphs it ranked.
//
// A PARAGRAPH'S VECTOR IS CACHED UNDER THE HASH OF ITS VISIBLE TEXT in ReportVecCachePath, so a
// search after an edit embeds the query and the paragraphs the edit changed, never the report again.
// The cache keeps the paragraphs the report now holds.
func SearchReport(run record.Run, reportMD, query string, emb vecembed.Embedder, topK int) ([]hybridsearch.ReportChunk, int, error) {
	chunks := hybridsearch.ChunkReport(reportMD)
	if len(chunks) == 0 {
		return []hybridsearch.ReportChunk{}, 0, nil
	}
	path := ReportVecCachePath(run)
	cached, err := hybridsearch.ReadReportVecCache(path, emb.ModelID(), hybridsearch.ReportChunkSpec)
	if errors.Is(err, hybridsearch.ErrCacheMiss) {
		cached = map[hybridsearch.ParaSha][]float32{}
	} else if err != nil {
		return nil, 0, err
	}
	shas := make([]hybridsearch.ParaSha, len(chunks))
	texts := []string{query}
	var fresh []hybridsearch.ParaSha
	queued := map[hybridsearch.ParaSha]bool{}
	for i, ch := range chunks {
		shas[i] = sha256.Sum256([]byte(ch.Text))
		if _, ok := cached[shas[i]]; !ok && !queued[shas[i]] {
			queued[shas[i]] = true
			fresh = append(fresh, shas[i])
			texts = append(texts, ch.Text)
		}
	}
	vecs, err := emb.Embed(texts)
	if err != nil {
		return nil, 0, err
	}
	if len(vecs) != len(texts) {
		return nil, 0, fmt.Errorf("the embedder returned %d vectors for %d texts", len(vecs), len(texts))
	}
	kept := make(map[hybridsearch.ParaSha][]float32, len(chunks))
	for i, sha := range fresh {
		kept[sha] = vecs[i+1]
	}
	for _, sha := range shas {
		if v, ok := cached[sha]; ok {
			kept[sha] = v
		}
	}
	if len(fresh) > 0 || len(kept) != len(cached) {
		// The cache saves the next search the embedding; a write that fails costs only that.
		_ = writeReplacing(path, hybridsearch.EncodeReportVecCache(emb.ModelID(), hybridsearch.ReportChunkSpec, kept))
	}
	items := make([]hybridsearch.Item, len(chunks))
	for i, ch := range chunks {
		items[i] = hybridsearch.Item{Index: i, Text: ch.Text, Vector: kept[shas[i]]}
	}
	scored, err := hybridsearch.FuseRRF(items, query, vecs[0], topK)
	if err != nil {
		return nil, 0, err
	}
	out := make([]hybridsearch.ReportChunk, 0, len(scored))
	for _, sc := range scored {
		ch := chunks[sc.Index]
		ch.Score = sc.Score
		out = append(out, ch)
	}
	return out, len(chunks), nil
}
