package compute

import (
	"strconv"
	"unicode/utf8"
)

// A PROGRAM MAY WRITE ANY BYTE, AND A RECORD STRING FIELD HOLDS ONLY UTF-8. A drift carries what a
// program wrote — its output, the names of the files it left under out/ — and a body whose string
// field holds invalid UTF-8 is written without complaint and then refused by every read that
// projects it. Every program byte bound for the record passes through Printable or Excerpt.
//
// Text that is valid UTF-8 is carried as the program wrote it. Text that is not is carried
// Go-quoted, whole: every byte kept, an invalid one as \xNN, and the quotes say the rendering is not
// the raw text. Replacing an invalid byte with U+FFFD would also be valid, and would lose which byte
// the program wrote — a drift that no longer says what moved.

// Printable is s as a record string field carries it: s itself when it is valid UTF-8, else s
// Go-quoted.
func Printable(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strconv.Quote(s)
}

// Excerpt is at most limit bytes of s, as Printable carries them, with "…" where s was cut. A cut
// that falls inside a rune of otherwise-valid text ends on the rune before it, so the cut never
// makes valid text read as invalid.
func Excerpt(s string, limit int) string {
	if len(s) <= limit {
		return Printable(s)
	}
	n := limit
	for i := 0; i < utf8.UTFMax-1 && n > 0 && !utf8.RuneStart(s[n]); i++ {
		n--
	}
	return Printable(s[:n]) + "…"
}

// Short is a sha's leading twelve characters, the form every message about a computation prints.
func Short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
