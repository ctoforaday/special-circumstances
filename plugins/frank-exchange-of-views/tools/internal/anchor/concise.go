package anchor

import (
	"strings"
	"unicode/utf8"
)

// ConciseLimit is the scan surface's width, and it is named where the reader meets it: the risk
// matrix's own footer states it, so a cut row reads as a stated limit rather than as the whole of
// what the seat wrote.
const ConciseLimit = 100

// Concise returns a scannable one-line version of a long field — its first sentence, or a head cut
// at a word boundary. A risk matrix is scanned, not read; the tool cannot summarize a 60-word gap
// description, so it takes the lead sentence and leaves the full text on the board, where nothing
// is lost.
//
// IT LIVES BESIDE Sentences BECAUSE TWO PACKAGES PRINT ITS CUT: the report's risk matrix cuts each
// open gap's problem and fix to it, and the voice lens's work list quotes those cells as the text it
// audits (record.composedSourcesOf). One function is one cut, so the lens reads what ships.
//
// TRUNCATION IS NOT COMPRESSION, and two ways of forgetting that both shipped here. The sentence
// split took the FIRST `.` anywhere, so "e.g. the parser" became "e." — a cell saying nothing, and
// saying it in a tool's voice, as though the seat had written one letter. The fallback cut at byte
// 100 regardless of what was there, producing "a piece of int…" mid-word and splitting any
// multi-byte rune that straddled the boundary. A sentence is therefore the one every reader of a
// sentence reads (Sentences), with a floor under the head so a leading initial ("A. B. Turing")
// does not make the cell "A."; the fallback backs up to the last space, exactly as the citation
// scent in the report's md.go already does. Where neither finds a boundary the cell is cut at the
// limit and marked — cut, and visibly so, is the honest report of a scan surface.
func Concise(s string) string {
	s = strings.Join(strings.Fields(s), " ") // collapse internal whitespace/newlines
	for _, sp := range Sentences(s) {
		if sp[1] > ConciseLimit {
			break
		}
		if sp[1] >= minSentence {
			return s[:sp[1]]
		}
	}
	if len(s) <= ConciseLimit {
		return s
	}
	if cut := strings.LastIndexByte(s[:ConciseLimit], ' '); cut > ConciseLimit/2 {
		return strings.TrimSpace(s[:cut]) + "…"
	}
	return strings.TrimSpace(truncateRunes(s, ConciseLimit)) + "…"
}

// minSentence is the floor under a cell's lead sentence.
const minSentence = 25

// truncateRunes cuts to at most n bytes WITHOUT splitting a rune, so a cell that has no space to
// back up to still renders as text rather than as a replacement character.
func truncateRunes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
