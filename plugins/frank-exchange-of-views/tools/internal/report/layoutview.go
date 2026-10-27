package report

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE ASSEMBLED LAYOUT: what the shipped set looks like, asked of the assembler.
//
// A seat reads the report's TEXT with `show report`, and the report a reader gets is that text
// placed among sections the tool composes. A seat needs to see which sections the tool composes.
// Layout assembles the set from the record as it stands and slices what assembly made by the rows
// of `layout`, so the view cannot describe a report other than the one that ships. It shows the
// record's text as written, because a seat quotes and edits it: the shipped .md is that text with a
// backslash before each `<` a CommonMark reader would take for a seat's HTML (Ship), and report.html
// shows it as typed.

// LayoutSection is one row of the assembled layout, for the run it was asked of.
type LayoutSection struct {
	File    string `json:"file"`
	Heading string `json:"heading"`
	// Line is set for a section that ships with no heading: the line it ships as opens with this.
	Line     string `json:"line,omitempty"`
	Composer string `json:"composer"`
	From     string `json:"from"`
	Path     string `json:"path"`
	// Ships is whether the set assembled from the record as it stands carries the section.
	Ships bool `json:"ships"`
	// OwnInReport is set for a section the tool composes in another document under a name the
	// report's text may use: a section of that name there is the report's own.
	OwnInReport bool `json:"own_in_report,omitempty"`
	// AuthoredHeadings is the headings of the report's own sections, on the row that is the body.
	AuthoredHeadings []string `json:"authored_headings,omitempty"`
	// Text is the section as assembly composed it: anchors resolved, notes numbered, the record's
	// text as written — not the shipped .md, which escapes a seat's HTML in it. It is always a string, "" on a
	// row that does not ship, so a search over every row's text is a search, not a type error.
	Text string `json:"text"`
	// NoText is why a row carries no text: set exactly on a row that does not ship. An empty `text`
	// alone reads as a section that ships empty, or as data the view failed to read.
	NoText string `json:"no_text,omitempty"`
}

// NotInTheSet is a row's `no_text` when the set assembled from the record as it stands does not
// carry the section.
const NotInTheSet = "not in the set as the record stands: assembly composed nothing for this section from the record, so no text of it ships"

// LayoutJSON is `show report --assembled --json`.
type LayoutJSON struct {
	Sections []LayoutSection `json:"sections"`
}

// Layout is the report set as assembly ships it from the record as it stands: every row of the
// table in shipping order, and what each holds. It assembles the set and slices the composed
// documents, so what it prints is what assembly composed, before the .md tier escapes it.
//
// A ROW SHIPS WHEN ASSEMBLY COMPOSED IT, AND STANDS WHERE ASSEMBLY PUT IT, never where a line of
// the document reads as its opening: a line of the report's own text can open as a section's
// does, the body opens under no heading of the table's, and a section in it can carry the heading
// of a row after it (a record from before the refusal holds such). Assembly writes a mark line
// above each block it composes here, in the table's order, and the row's text runs to the next
// row that ships. The mark is drawn for the call, so no line of the report's text is one.
func Layout(run record.Run) (LayoutJSON, error) {
	mark := "\x00" + rand.Text()
	docs, composed, title, err := assemble(run, mark)
	if err != nil {
		return LayoutJSON{}, err
	}
	// Each document is read once, its mark lines out: the k-th mark stands above the block of the
	// k-th row of that document assembly composed, and the line under it is where that row starts.
	rendered := map[string]ParsedText{}
	marks := map[string][]int{}
	for _, d := range docs {
		var kept []string
		for _, ln := range strings.Split(composedText(d, title, docs), "\n") {
			if ln == mark {
				marks[d.File] = append(marks[d.File], len(kept))
				continue
			}
			kept = append(kept, ln)
		}
		rendered[d.File] = ParseText(strings.Join(kept, "\n"))
	}
	starts := make([]int, len(layout))
	for i, s := range layout {
		starts[i] = -1
		doc, ok := rendered[s.File]
		switch {
		case !ok:
			continue
		case s.compose != nil && composed[s.Heading]:
			if m := marks[s.File]; len(m) > 0 {
				starts[i], marks[s.File] = m[0], m[1:]
			}
		case s.compose == nil && (composed[s.Heading] || s.Line != ""):
			starts[i] = opens(doc, s)
		default:
			continue
		}
		if starts[i] < 0 {
			return LayoutJSON{}, fmt.Errorf("assembly composed %q for %s and the document does not open it", s.Heading, s.File)
		}
	}
	var out LayoutJSON
	for i, s := range layout {
		row := LayoutSection{File: s.File, Heading: s.Heading, Line: strings.TrimSpace(s.Line), Composer: s.Composer, From: s.From, Path: s.Path, OwnInReport: s.SubjectHeading, NoText: NotInTheSet}
		if start := starts[i]; start >= 0 {
			lines := rendered[s.File].lines
			end := len(lines)
			for j := i + 1; j < len(layout) && end == len(lines); j++ {
				if layout[j].File == s.File && starts[j] >= 0 {
					end = starts[j]
				}
			}
			if s.Line != "" {
				end = start + 1
			}
			row.Ships, row.NoText = true, ""
			var text []string
			for _, l := range lines[start:end] {
				text = append(text, l.text)
				if _, ok := headingKey(l); ok && s.embeds && !l.code {
					row.AuthoredHeadings = append(row.AuthoredHeadings, strings.TrimPrefix(l.trimmed, "## "))
				}
			}
			row.Text = strings.TrimSpace(strings.Join(text, "\n"))
		}
		out.Sections = append(out.Sections, row)
	}
	return out, nil
}

// opens is the line a row no block of assembly's places opens on: the title, the first line that
// starts as it does, which is the head (head) composedText puts on every document; and a section a pass over
// the whole document writes after the rows, the last heading of its name. -1 when none does.
func opens(doc ParsedText, s Section) int {
	found := -1
	for i, l := range doc.lines {
		if l.code {
			continue
		}
		t := l.trimmed
		if s.Line != "" {
			if found < 0 && strings.HasPrefix(t, s.Line) {
				found = i
			}
		} else if key, ok := headingKey(l); ok && key == normalizeHeading(s.Heading) {
			found = i
		}
	}
	return found
}

// LayoutHasNoText is what the bare layout says of itself on standard error, where it survives a
// pipe. The bare form lists the sections and carries none of their text, so a search piped over it
// finds nothing and reads as a clean report: on universe-m22 a gap's acceptance check was
// `show report --assembled | grep -c <phrase>`, twice, and zero was the answer either way. A line
// at the head of the listing is the first thing such a pipe discards.
const LayoutHasNoText = "feov-record: `show report --assembled` LISTS THE SECTIONS WITHOUT THEIR TEXT — a search of this output finds no sentence of the report, and a count of zero here says nothing about what ships. `show report --assembled --json` carries each section's `text`."

// Markdown is the layout as a seat reads it: one entry per section, in shipping order.
func (l LayoutJSON) Markdown() string {
	var b strings.Builder
	b.WriteString("THE REPORT SET AS ASSEMBLY SHIPS IT, from the record as it stands — each section in shipping order, who composes it, and how content reaches it. " +
		"The report's text carries no section the tool composes: blue's `ingest` and `edit` refuse one. The exception is a section of another document whose entry says that a section of that name in the report's text is the report's own.\n")
	file := ""
	for _, s := range l.Sections {
		if s.File != file {
			file = s.File
			b.WriteString("\n" + file + "\n")
		}
		name := "## " + s.Heading
		switch {
		case s.Line != "":
			name = s.Heading + " (no heading: one line, opening `" + s.Line + "`)"
		case s.Heading == bodyHeading:
			name = s.Heading + " (no heading of the tool's: each section ships under the heading its author gave it)"
		}
		ships := ""
		if !s.Ships {
			ships = " — NOT IN THE SET as the record stands"
		}
		fmt.Fprintf(&b, "\n  %s%s\n    composed by: %s, from %s\n    how content reaches it: %s\n", name, ships, s.Composer, s.From, s.Path)
		if s.OwnInReport {
			fmt.Fprintf(&b, "    in the report's text: this entry is the tool's section, in %s; %s\n", s.File, OwnInTheReport)
		}
		if len(s.AuthoredHeadings) > 0 {
			fmt.Fprintf(&b, "    as the record stands: %s\n", strings.Join(s.AuthoredHeadings, "; "))
		}
	}
	return b.String()
}
