package tessocr

import "strings"

// The two text transforms between the engine's raw output and the bytes a record hashes.
// They live in this package, not in the read loops that call them, because their output IS
// the reading: Identity() hashes this package's source, and a normalisation or join that
// sat outside it could change every stored text_sha under an unchanged identity.

// NormalizeReading settles line endings and strips trailing whitespace.
//
// THIS IS HYGIENE ON STORED TEXT, NOT INTERPRETATION OF IT. A reading is about to become
// a citable artifact, and CRLF or a trailing run of spaces in it is noise a reader would
// have to look past. Interior line structure is PRESERVED — paragraph breaks are how the
// page reads, and flattening them would be this function editing the content rather than
// tidying its edges.
func NormalizeReading(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// AssembleReading joins normalised page readings, in page order, into the document text:
// one blank line between pages, no trailing newline. A blank page contributes an empty
// string and so still occupies its separator — the assembled text keeps every page's
// place, which is what lets a citation's page be found in it.
func AssembleReading(pages []string) string {
	return strings.TrimRight(strings.Join(pages, "\n\n"), "\n")
}
