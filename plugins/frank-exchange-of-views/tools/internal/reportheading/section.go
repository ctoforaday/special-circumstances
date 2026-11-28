package reportheading

import "strings"

// A REPORT SECTION runs from its heading line to the next heading at its level or shallower, or the
// end: the bound a gap's passage is read in (record.PassageAround), and the bound a section gap's
// subject covers (#1695 §III.B0 "Sections").

// SectionStart is the offset of the heading line governing at, and that heading's level. Level 0
// with offset 0 means at sits before any heading — the report's preamble is its section.
func SectionStart(report string, at int) (int, int) {
	for i := lineStart(report, at); ; {
		if lvl := HeadingLevel(report, i); lvl > 0 {
			return i, lvl
		}
		if i == 0 {
			return 0, 0
		}
		i = lineStart(report, i-1)
	}
}

// SectionEnd is the offset of the next heading after from at level or shallower, or the end.
func SectionEnd(report string, from, level int) int {
	// LEVEL 0 IS "NO HEADING GOVERNS THIS", which is text before the document's first heading. It
	// ends at the NEXT heading of any level — defaulting it to 1 made a `##` fail to stop it, so a
	// quote in an unheaded preamble swallowed the first real section.
	if level == 0 {
		level = 6
	}
	for i := from; i < len(report); {
		nl := strings.IndexByte(report[i:], '\n')
		if nl < 0 {
			return len(report)
		}
		i += nl + 1
		if lvl := HeadingLevel(report, i); lvl > 0 && lvl <= level {
			return i
		}
	}
	return len(report)
}

// HeadingLevel is the number of leading '#' on the line beginning at i, or 0 if it is not an ATX
// heading. A '#' not followed by a space is not one — `#1091` in prose is not a section.
func HeadingLevel(report string, i int) int {
	n := 0
	for i+n < len(report) && report[i+n] == '#' {
		n++
	}
	if n == 0 || n > 6 || i+n >= len(report) || report[i+n] != ' ' {
		return 0
	}
	return n
}

func lineStart(s string, at int) int {
	if at > len(s) {
		at = len(s)
	}
	if i := strings.LastIndexByte(s[:at], '\n'); i >= 0 {
		return i + 1
	}
	return 0
}
