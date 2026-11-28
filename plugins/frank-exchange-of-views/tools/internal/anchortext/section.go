package anchortext

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
)

// A SECTION GAP STANDS ON ITS HEADING (plans/findings-fold.md §III.B0 "Sections", picked 2026-11-25
// "Anchor headings"): its token sits at the end of the heading its --about keys to, the one anchor a
// heading holds. The write places it (AttachToHeading) and replay places it at the same byte
// (InsertAnchor over the location HeadingLocation returns): the location is the heading's line as the
// report prints it, and the write refuses where that line's first match in the report is not the
// heading itself, which is where replay would put the token.

// Errors a heading placement refuses with.
var (
	ErrNoHeading      = errors.New("no heading of the report keys to the section")
	ErrHeadingTwice   = errors.New("two headings of the report key to the section")
	ErrHeadingEarlier = errors.New("the heading's line occurs earlier in the report than the heading")
)

// Heading is one ATX heading of a document: its line [Start, Stop), its level, and its text — the line
// without its leading `#` run, anchors kept.
type Heading struct {
	Start, Stop, Level int
	Text               string
}

// Headings is every ATX heading of doc, in order, as anchor.Blocks reads its blocks: a `#` line in a
// fence is no heading.
func Headings(doc string) []Heading {
	var out []Heading
	for _, b := range anchor.Blocks(doc) {
		if b.Kind != anchor.Heading {
			continue
		}
		line := doc[b.Start:b.End]
		trimmed := strings.TrimLeft(line, " ")
		level := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if level == 0 || level > 6 {
			continue // a setext heading: no section gap stands on one
		}
		out = append(out, Heading{Start: b.Start, Stop: b.End, Level: level, Text: strings.TrimSpace(trimmed[level:])})
	}
	return out
}

// HeadingsKeyed is the headings of doc whose key (reportheading.HeadingKey) is ref's.
func HeadingsKeyed(doc, ref string) []Heading {
	key := reportheading.HeadingKey(strings.TrimLeft(ref, "# "))
	var out []Heading
	for _, h := range Headings(doc) {
		if reportheading.HeadingKey(h.Text) == key {
			out = append(out, h)
		}
	}
	return out
}

// End is where h's section ends in doc: the start of the next heading (Headings) at h's level or
// above, or the end of doc. A `#` line in a fence is no heading and ends nothing.
func (h Heading) End(doc string) int {
	for _, n := range Headings(doc) {
		if n.Start > h.Start && n.Level <= h.Level {
			return n.Start
		}
	}
	return len(doc)
}

// HeadingLocation is where a section gap about ref places its token in doc: the one heading keyed to
// ref, as the location replay places it at — the heading's line, its anchors stripped. It refuses
// ErrNoHeading, ErrHeadingTwice, and ErrHeadingEarlier where that line's first match is not the
// heading.
func HeadingLocation(doc, ref string) (string, Heading, error) {
	hs := HeadingsKeyed(doc, ref)
	switch len(hs) {
	case 0:
		return "", Heading{}, ErrNoHeading
	case 1:
	default:
		return "", Heading{}, ErrHeadingTwice
	}
	h := hs[0]
	loc := strings.TrimSpace(claimcount.StripAnchors(doc[h.Start:h.Stop]))
	if _, end := locate(doc, loc, StopAtParagraph); end <= h.Start || end > h.Stop {
		return "", Heading{}, fmt.Errorf("%w: %q", ErrHeadingEarlier, loc)
	}
	return loc, h, nil
}

// AttachToHeading is doc with gap id's token at the end of the heading HeadingLocation finds for ref,
// and the location replay places it at.
func AttachToHeading(doc, id, ref string) (string, string, error) {
	loc, _, err := HeadingLocation(doc, ref)
	if err != nil {
		return "", "", err
	}
	next, err := InsertAnchor([]byte(doc), loc, anchor.Token(id))
	if err != nil {
		return "", "", err
	}
	return string(next), loc, nil
}
