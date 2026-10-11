package report

import (
	"fmt"
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

// AssembledNote is ComposedNote and the way to the assembled form, on the page that prints the
// report's text.
var AssembledNote = ComposedNote + " --assembled reads the set as the run ships it: every section in shipping order, who composes it and from what, and how content reaches it. " +
	"With --assembled --json each section that ships carries its `text` as the reader gets it, anchors resolved — the form in which a composed section is read or audited: pipe it to `jq` and select on `heading`."

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
	// AuthoredHeadings is the headings of the report's own sections that ship inside this one.
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
func Layout(run record.Run) (LayoutJSON, error) {
	docs, err := AssembleAll(run)
	if err != nil {
		return LayoutJSON{}, err
	}
	title := Title(run)
	rendered := map[string][]string{}
	for _, d := range docs {
		rendered[d.File] = strings.Split(render(d, title, docs), "\n")
	}
	var out LayoutJSON
	at := map[string]int{} // per file: the line the previous row's section opened on
	for i, s := range layout {
		row := LayoutSection{File: s.File, Heading: s.Heading, Line: strings.TrimSpace(s.Line), Composer: s.Composer, From: s.From, Path: s.Path}
		lines := rendered[s.File]
		if start := opens(lines, at[s.File], s); start >= 0 {
			end := len(lines)
			for _, next := range layout[i+1:] {
				if next.File != s.File {
					continue
				}
				if n := opens(lines, start+1, next); n >= 0 {
					end = n
					break
				}
			}
			if s.Line != "" {
				end = start + 1
			}
			at[s.File] = start + 1
			row.Ships = true
			row.Text = strings.TrimSpace(strings.Join(lines[start:end], "\n"))
			if s.embeds {
				eachLine(strings.Join(lines[start+1:end], "\n"), func(_, t string, fenced bool) {
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

// opens is the first line at or after from on which s opens, or -1.
func opens(lines []string, from int, s Section) int {
	found := -1
	n := 0
	eachLine(strings.Join(lines, "\n"), func(_, t string, fenced bool) {
		i := n
		n++
		if found >= 0 || fenced || i < from {
			return
		}
		if s.Line != "" {
			if strings.HasPrefix(t, s.Line) {
				found = i
			}
			return
		}
		if key, ok := headingKey(t); ok && key == normalizeHeading(s.Heading) {
			found = i
		}
	})
	return found
}

// Markdown is the layout as a seat reads it: one entry per section, in shipping order.
func (l LayoutJSON) Markdown() string {
	var b strings.Builder
	b.WriteString("THE REPORT SET AS ASSEMBLY SHIPS IT, from the record as it stands — each section in shipping order, who composes it, and how content reaches it. " +
		"The report's text carries no section the tool composes: blue's `ingest` and `edit` refuse one.\n")
	file := ""
	for _, s := range l.Sections {
		if s.File != file {
			file = s.File
			b.WriteString("\n" + file + "\n")
		}
		name := "## " + s.Heading
		if s.Line != "" {
			name = s.Heading + " (no heading: one line, opening `" + s.Line + "`)"
		}
		ships := ""
		if !s.Ships {
			ships = " — NOT IN THE SET as the record stands"
		}
		fmt.Fprintf(&b, "\n  %s%s\n    composed by: %s, from %s\n    how content reaches it: %s\n", name, ships, s.Composer, s.From, s.Path)
		if len(s.AuthoredHeadings) > 0 {
			fmt.Fprintf(&b, "    holds the report's own sections: %s\n", strings.Join(s.AuthoredHeadings, "; "))
		}
	}
	return b.String()
}
