// Package hybridsearch implements chunking, vector caching, and hybrid (BM25 + dense vector RRF) retrieval.
package hybridsearch

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// chunkWords and chunkStride are ChunkSource's window: 200 words, each chunk starting 160 words
// after the last, so neighbours share 40.
const (
	chunkWords  = 200
	chunkStride = 160
)

// SourceChunkSpec names ChunkSource's chunking in a source vector cache's header, so a file of
// another chunker's offsets is a miss. It follows the window; any other change to how ChunkSource
// cuts text changes it by hand.
var SourceChunkSpec = fmt.Sprintf("w%d_s%d", chunkWords, chunkStride)

// ReportChunkSpec names ChunkReport's paragraphs in the report vector cache's header.
const ReportChunkSpec = "para_v1"

// SourceChunk is one chunk of a source text.
type SourceChunk struct {
	Index     int
	Text      string
	StartByte int
	EndByte   int
	Page      int
}

// WordSpan is a word's byte boundaries in text.
type WordSpan struct {
	Start int
	End   int
}

// FindWordSpans returns the byte offset spans of whitespace-delimited words in text.
func FindWordSpans(text string) []WordSpan {
	var spans []WordSpan
	inWord := false
	start := 0
	for i, r := range text {
		if unicode.IsSpace(r) {
			if inWord {
				spans = append(spans, WordSpan{Start: start, End: i})
				inWord = false
			}
		} else {
			if !inWord {
				start = i
				inWord = true
			}
		}
	}
	if inWord {
		spans = append(spans, WordSpan{Start: start, End: len(text)})
	}
	return spans
}

// ChunkSource splits text into 200-word chunks with 40-word overlap (stride 160).
// pageMap is an optional slice of page boundary start byte offsets. If empty, page is 0.
func ChunkSource(text string, pageMap []int) []SourceChunk {
	spans := FindWordSpans(text)
	if len(spans) == 0 {
		return nil
	}

	if len(spans) <= chunkWords {
		startByte := spans[0].Start
		endByte := spans[len(spans)-1].End
		return []SourceChunk{
			{
				Index:     0,
				Text:      text[startByte:endByte],
				StartByte: startByte,
				EndByte:   endByte,
				Page:      pageForOffset(startByte, pageMap),
			},
		}
	}

	var chunks []SourceChunk
	idx := 0
	for start := 0; start < len(spans); start += chunkStride {
		end := start + chunkWords
		if end > len(spans) {
			end = len(spans)
		}
		startByte := spans[start].Start
		endByte := spans[end-1].End
		chunks = append(chunks, SourceChunk{
			Index:     idx,
			Text:      text[startByte:endByte],
			StartByte: startByte,
			EndByte:   endByte,
			Page:      pageForOffset(startByte, pageMap),
		})
		idx++
		if end == len(spans) {
			break
		}
	}
	return chunks
}

func pageForOffset(offset int, pageMap []int) int {
	if len(pageMap) == 0 {
		return 0
	}
	page := 0
	for p, start := range pageMap {
		if offset >= start {
			page = p + 1
		} else {
			break
		}
	}
	if page == 0 && len(pageMap) > 0 {
		page = 1
	}
	return page
}

// ReportChunk is one paragraph or heading-bounded block of the report: its visible text, every
// anchor on it, and the lines it spans.
type ReportChunk struct {
	Index     int     `json:"index,omitempty"`
	Heading   string  `json:"heading"`
	StartLine int     `json:"start_line"`
	EndLine   int     `json:"end_line"`
	Text      string  `json:"text"`
	Score     float32 `json:"score,omitempty"`
	// Anchor is the first anchor id on the paragraph, and Anchors every one, in order.
	Anchor  string   `json:"anchor,omitempty"`
	Anchors []string `json:"anchors,omitempty"`
}

// ChunkReport splits a markdown report into paragraphs, each under the heading it sits beneath.
//
// A PARAGRAPH IS RANKED BY THE WORDS A READER SEES. Its anchor tokens — every kind the anchor table
// holds — come out of Text before BM25 tokenises it and before it is embedded, so a tag or a hex id
// is never a term; the ids they carry are kept in Anchors. A paragraph that is nothing but anchors
// has no words to rank and is not a chunk.
func ChunkReport(reportMD string) []ReportChunk {
	lines := strings.Split(strings.ReplaceAll(reportMD, "\r\n", "\n"), "\n")
	var chunks []ReportChunk
	curHeading := ""
	var curLines []string
	startLine := 1

	flush := func(endLine int) {
		if len(curLines) == 0 {
			return
		}
		raw := strings.Join(curLines, "\n")
		text := strings.TrimSpace(anchor.Replace(raw, func(string, string) string { return "" }))
		if text != "" {
			ch := ReportChunk{
				Index:     len(chunks),
				Heading:   curHeading,
				StartLine: startLine,
				EndLine:   endLine,
				Text:      text,
				Anchors:   anchor.IDs(raw),
			}
			if len(ch.Anchors) > 0 {
				ch.Anchor = ch.Anchors[0]
			}
			chunks = append(chunks, ch)
		}
		curLines = nil
	}

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			flush(lineNum - 1)
			curHeading = trimmed
			startLine = lineNum
			continue
		}
		if trimmed == "" {
			flush(lineNum - 1)
			startLine = lineNum + 1
			continue
		}
		if len(curLines) == 0 {
			startLine = lineNum
		}
		curLines = append(curLines, line)
	}
	flush(len(lines))
	return chunks
}
