package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE ASSEMBLED LAYOUT: what the shipped set looks like, asked of the assembler.
//
// A seat reads the report's TEXT with `show report`, and the report a reader gets is that text
// placed among sections the tool composes. Nothing showed a seat the second thing: universe-m20's
// synthesizer wrote three avenue sections it could not see were the tool's, and red audited them
// as the reader's text for five sittings. Layout assembles the set from the record as it stands
// and slices what assembly made by the rows of `layout`, so the view cannot describe a report
// other than the one that ships.

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
	// Text is the section as it ships: anchors resolved, notes numbered.
	Text string `json:"text,omitempty"`
}

// LayoutJSON is `show report --assembled --json`.
type LayoutJSON struct {
	Sections []LayoutSection `json:"sections"`
}

// Layout is the report set as assembly ships it from the record as it stands: every row of the
// table in shipping order, and what each holds. It assembles the set and slices the result, so
// what it prints is what ships.
//
// A ROW SHIPS WHEN ASSEMBLY COMPOSED IT, never because a line of the document reads as its
// opening: a line of the report's own text can open as a section's does. The document is read
// only for where a composed row's text begins — the first line that opens it after the row
// before it — and a row's text runs to the next row that ships.
//
// THE BODY OPENS UNDER NO HEADING OF THE TABLE'S, and a section in it can carry the heading of a
// row after it (a record from before the refusal holds such). Assembly writes a mark line before
// and after it here; the marks say where it stands, and the rows after it are looked for past it.
func Layout(run record.Run) (LayoutJSON, error) {
	const mark = "\x00"
	docs, composed, err := assemble(run, mark)
	if err != nil {
		return LayoutJSON{}, err
	}
	title := Title(run)
	rendered := map[string][]string{}
	for _, d := range docs {
		rendered[d.File] = strings.Split(render(d, title, docs), "\n")
	}
	starts := make([]int, len(layout))
	at := map[string]int{} // per file: the line after the one the previous row opened on
	for i, s := range layout {
		starts[i] = -1
		if s.embeds {
			if composed[s.Heading] {
				lines := rendered[s.File]
				open := slices.Index(lines, mark)
				end := slices.Index(lines[open+1:], mark) + open // where the closing mark stands once the opening one is out
				if open < 0 || end < open {
					return LayoutJSON{}, fmt.Errorf("report: assembly composed the report's own sections for %s and the document does not hold them between its marks", s.File)
				}
				rendered[s.File] = slices.Delete(slices.Delete(lines, end+1, end+2), open, open+1)
				starts[i], at[s.File] = open, end
			}
			continue
		}
		// The title is the head render writes on every document.
		if lines := rendered[s.File]; lines != nil && (composed[s.Heading] || s.compose == nil && s.Line != "") {
			if starts[i] = opens(lines, at[s.File], s); starts[i] < 0 {
				return LayoutJSON{}, fmt.Errorf("report: assembly composed %q for %s and the document does not open it", s.Heading, s.File)
			}
			at[s.File] = starts[i] + 1
		}
	}
	var out LayoutJSON
	for i, s := range layout {
		row := LayoutSection{File: s.File, Heading: s.Heading, Line: strings.TrimSpace(s.Line), Composer: s.Composer, From: s.From, Path: s.Path, OwnInReport: s.SubjectHeading}
		if start := starts[i]; start >= 0 {
			lines := rendered[s.File]
			end := len(lines)
			for j := i + 1; j < len(layout) && end == len(lines); j++ {
				if layout[j].File == s.File && starts[j] >= 0 {
					end = starts[j]
				}
			}
			if s.Line != "" {
				end = start + 1
			}
			row.Ships = true
			row.Text = strings.TrimSpace(strings.Join(lines[start:end], "\n"))
			if s.embeds {
				eachLine(strings.Join(lines[start:end], "\n"), func(_, t string, fenced bool) {
					if _, ok := headingKey(t); ok && !fenced {
						row.AuthoredHeadings = append(row.AuthoredHeadings, strings.TrimPrefix(t, "## "))
					}
				})
			}
		}
		out.Sections = append(out.Sections, row)
	}
	return out, nil
}

// opens is the line at or after from on which s opens: the first, or for a section a pass over
// the whole document writes after the rows, the last; -1 when none does.
func opens(lines []string, from int, s Section) int {
	found := -1
	n := 0
	eachLine(strings.Join(lines, "\n"), func(_, t string, fenced bool) {
		i := n
		n++
		if fenced || i < from || found >= 0 && s.compose != nil {
			return
		}
		if s.Line != "" {
			if found < 0 && strings.HasPrefix(t, s.Line) {
				found = i
			}
		} else if key, ok := headingKey(t); ok && key == normalizeHeading(s.Heading) {
			found = i
		}
	})
	return found
}

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
