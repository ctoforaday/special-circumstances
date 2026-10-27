package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE CITATIONS AS THEY WERE CHECKED. The report's notes print a source's title, URL and access
// date, and red's reading of each source against the sentence it backs reached no reader: across
// three archived runs, 82 readings and not one of their texts in any document (the reach audit,
// 2026-10-20). A reader could not tell a citation red found `refutes` from one nobody read. This
// section is the evidence view's sources (record.EvidenceJSONOf) as a document: each source with
// every reading of it, its outcome, its confidence and what the reader read, and red's readings of
// sources the report does not cite after them.

// citationsHeading is the section of evidence.md this file composes.
const citationsHeading = "Citations"

// citationsSection renders evidence.md's citations, or "" for a run that cited nothing and read nothing.
func citationsSection(fam record.Family) string {
	ev := record.EvidenceJSONOf(fam.Events, fam.At)
	// A labelled corroboration is listed among the sources, with its own reading first; it is under
	// `independent` as well, where it would be listed twice.
	var others []record.EvidenceVerificationJSON
	for _, v := range ev.Independent {
		if v.Label == "" || v.Claim == "" {
			others = append(others, v)
		}
	}
	if len(ev.Sources) == 0 && len(others) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## " + citationsHeading + "\n\n")
	b.WriteString("Each source the report cites, with the sentence it backs and every reading of the source against that sentence. ")
	b.WriteString("A reading's outcome is what the source does for the sentence — `supports`, `supports_with_bridge` (it supports the sentence through a step the source does not state), `weak`, `refutes`, `absent` (the source says nothing on it) or `unreachable` — and its confidence is how sure the reading is.\n\n")
	for _, s := range ev.Sources {
		fmt.Fprintf(&b, "### %s\n\n", firstNonEmpty(plain(s.Title), plain(s.URL)))
		fmt.Fprintf(&b, "- **source**: %s", plain(s.URL))
		if s.AccessDate != "" {
			fmt.Fprintf(&b, ", accessed %s", s.AccessDate)
		}
		b.WriteString("\n")
		if loc := plain(s.Location); loc != "" {
			fmt.Fprintf(&b, "- **backs**: %s\n", loc)
		}
		if s.CorroboratedBy != "" {
			fmt.Fprintf(&b, "- **red's corroboration**, found by `%s`: blue did not cite it\n", s.CorroboratedBy)
		}
		if len(s.Verified) == 0 {
			b.WriteString("- **not read against its sentence** — no reading of this source is on the record\n")
		}
		b.WriteString(readings(s.Verified, plain(s.Location)))
		b.WriteString("\n")
	}
	if len(others) > 0 {
		fmt.Fprintf(&b, "### Sources the report does not cite (%d)\n\n", len(others))
		b.WriteString("Sources red found and read against a sentence of the report that place no citation in it: a reading that does not support its sentence places none.\n\n")
		for _, v := range others {
			fmt.Fprintf(&b, "- **%s**", firstNonEmpty(plain(v.Title), plain(v.URL), "(no source named)"))
			if v.URL != "" && plain(v.Title) != "" {
				fmt.Fprintf(&b, " — %s", plain(v.URL))
			}
			b.WriteString("\n")
			b.WriteString(indent(readings([]record.EvidenceVerificationJSON{v}, ""), "  "))
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// readings renders each reading of one source as a list item: who read it and in which epoch, the
// outcome and the confidence, the page when it read one, the sentence it read against when that is
// not the sentence the source now backs, and what it read. A text repeated from the reading above
// it is said to be, rather than printed twice.
func readings(vs []record.EvidenceVerificationJSON, backs string) string {
	var b strings.Builder
	prev, prevClaim := "", ""
	for _, v := range vs {
		fmt.Fprintf(&b, "- **read by `%s`** (epoch %d): **%s**", v.SeatID, v.Epoch, firstNonEmpty(v.Outcome, "no outcome"))
		if v.Confidence != "" {
			fmt.Fprintf(&b, ", confidence %s", v.Confidence)
		}
		if v.Page > 0 {
			fmt.Fprintf(&b, ", page %d", v.Page)
		}
		text := plain(v.Text)
		switch {
		case text == "":
		case text == prev:
			b.WriteString(" — the same reading as the one above")
		default:
			b.WriteString(" — " + text)
		}
		prev = text
		b.WriteString("\n")
		switch claim := plain(v.Claim); {
		case claim == "" || claim == backs:
		case claim == prevClaim:
			b.WriteString("  - of the same sentence as the reading above\n")
		default:
			fmt.Fprintf(&b, "  - of the sentence: %s\n", claim)
		}
		prevClaim = plain(v.Claim)
	}
	return b.String()
}

// plain is a record field as the tool writes it into a line it composes after the weave — a field of
// evidence.md, a note's text, a Bibliography entry — on one line, with no anchor token in it
// (stripTokens): no weave follows to take one, and no strip.
func plain(s string) string { return oneLine(stripTokens(s, false)) }

// plainInLine is plain for a field a line holds after the notes of its document are escaped — a
// note's text, a Bibliography entry — with its own escaped (escapeNotesInLine).
func plainInLine(s string) string { return escapeNotesInLine(plain(s)) }

// indent prefixes every non-empty line of s.
func indent(s, by string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln != "" {
			lines[i] = by + ln
		}
	}
	return strings.Join(lines, "\n")
}
