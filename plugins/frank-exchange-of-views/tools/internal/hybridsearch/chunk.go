// Package hybridsearch implements chunking, vector caching, and hybrid (BM25 + dense vector RRF) retrieval.
package hybridsearch

import (
	"strings"
	"unicode"
)

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

	const chunkSize = 200
	const stride = 160

	if len(spans) <= chunkSize {
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
	for start := 0; start < len(spans); start += stride {
		end := start + chunkSize
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

// ReportChunk is one paragraph or logical section of the report.
type ReportChunk struct {
	Index     int     `json:"index,omitempty"`
	Heading   string  `json:"heading"`
	StartLine int     `json:"start_line"`
	EndLine   int     `json:"end_line"`
	Text      string  `json:"text"`
	Score     float32 `json:"score,omitempty"`
	Anchor    string  `json:"anchor,omitempty"`
}

// ChunkReport splits markdown report into paragraphs / heading-bounded blocks.
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
		text := strings.TrimSpace(strings.Join(curLines, "\n"))
		if text != "" {
			anchor := extractAnchor(text)
			chunks = append(chunks, ReportChunk{
				Index:     len(chunks),
				Heading:   curHeading,
				StartLine: startLine,
				EndLine:   endLine,
				Text:      text,
				Anchor:    anchor,
			})
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

func extractAnchor(text string) string {
	idx := strings.Index(text, "C-")
	if idx < 0 {
		return ""
	}
	end := idx + 2
	for end < len(text) && ((text[end] >= '0' && text[end] <= '9') || (text[end] >= 'a' && text[end] <= 'f') || (text[end] >= 'A' && text[end] <= 'F')) {
		end++
	}
	if end-idx >= 8 {
		return text[idx:end]
	}
	return ""
}
