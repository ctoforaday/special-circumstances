package tessocr

import (
	"strings"
	"unicode"
)

// LINES THE RECOGNIZER READ OUT OF SOMETHING THAT IS NOT PRINT (#1030).
//
// A dashed rule, a row of leader dots, or a word in a script the language data does not hold is
// still handed to the recognizer, and it answers with characters: "ee ee Fe ee! ee ee", "9S SS SSS
// 9539", "a |". Those lines are quotable, they inflate every word count taken from the reading, and
// no threshold on the TEXT finds them — they are mostly letters. Their BOXES give them away. A
// letter or digit is never much wider than it is tall, so a word whose box spreads several heights
// per character it claims was read from something else: a dash rule read as "ee" sits in a box 225
// px wide and 7 tall.
//
// Measured over the 13 corpus pages, both segmentation passes, 7286 words read at confidence 85 or
// more: the widest such word spreads 1.81 heights per character ("on" on a page read sideways). Every
// line removed on nbs602-dashed-matrix is a rule, and nothing on that page that holds a value is.
// The ratio is the same at any resolution, so it is judged on the boxes as tesseract returned them.
const (
	// spreadAspect is the width per character, in multiples of the box's height, at which a word
	// holding a letter or digit is SPREAD — wider than what it claims to read can fill.
	spreadAspect = 2.5
	// spreadLineShare is the share of a line's letters and digits that spread words must carry for
	// the line to go. A real line with one broken word keeps it.
	spreadLineShare = 0.5
	// spreadSureConf is the confidence at which a word that is not spread vouches for its line: a
	// line holding any such word is kept unless spread words carry spreadLineShare of it. A rule
	// crossed by short vertical dashes reads as "ce ce ee" in narrow boxes that are not spread, at
	// confidence 22-45. Over the corpus, every line this clause removes is a rule or an unread
	// script; a line of print holds at least one word read surely.
	spreadSureConf = 60
)

// alnumRunes counts a word's letters and digits — what a spread box is measured against.
func alnumRunes(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			n++
		}
	}
	return n
}

// spread reports a word that holds a letter or digit in a box wider per character than
// spreadAspect times its height. A word of punctuation is never spread: a dash is meant to be wide.
func (wd tsvWord) spread() bool {
	n := len([]rune(wd.text))
	if wd.h <= 0 || n == 0 || alnumRunes(wd.text) == 0 {
		return false
	}
	return float64(wd.w)/float64(n) >= spreadAspect*float64(wd.h)
}

// spreadLine decides one layout line. A line with no letter or digit is not judged at all: there is
// nothing in it a box could fail to fill.
func spreadLine(ws []tsvWord) bool {
	total, spread, sure := 0, 0, 0
	for _, w := range ws {
		n := alnumRunes(w.text)
		total += n
		switch {
		case w.spread():
			spread += n
		case w.conf >= spreadSureConf:
			sure += n
		}
	}
	if total == 0 || spread == 0 {
		return false
	}
	return float64(spread) >= spreadLineShare*float64(total) || sure == 0
}

// layoutLines groups words, in the order tesseract wrote them, into its layout lines.
func layoutLines(words []tsvWord) [][]tsvWord {
	var out [][]tsvWord
	for i, w := range words {
		if i == 0 || w.line != words[i-1].line {
			out = append(out, nil)
		}
		out[len(out)-1] = append(out[len(out)-1], w)
	}
	return out
}

// removeSpreadLines returns the words of every line that is not spread, and the text of every line
// that is, in reading order. The removed text is returned rather than discarded: what the reader
// took out of a page has to stay where a reader of the record can see it.
func removeSpreadLines(words []tsvWord) (kept []tsvWord, removed []string) {
	for _, ln := range layoutLines(words) {
		if spreadLine(ln) {
			removed = append(removed, lineText(ln))
			continue
		}
		kept = append(kept, ln...)
	}
	return kept, removed
}

func lineText(ws []tsvWord) string {
	parts := make([]string, len(ws))
	for i, w := range ws {
		parts[i] = w.text
	}
	return strings.Join(parts, " ")
}

// proseText writes words as the plain reading: one line per layout line, a blank line between
// paragraphs. It is what tesseract's own text output says for the same recognition — measured
// identical, after NormalizeReading, on all 13 corpus pages — written from the TSV so that the
// lines removeSpreadLines removed are absent from it.
func proseText(words []tsvWord) string {
	var b strings.Builder
	lines := layoutLines(words)
	for i, ln := range lines {
		if i > 0 {
			b.WriteString("\n")
			prev := lines[i-1][0].line
			if prev.block != ln[0].line.block || prev.par != ln[0].line.par {
				b.WriteString("\n")
			}
		}
		b.WriteString(lineText(ln))
	}
	return b.String()
}
