package corpus

import (
	"fmt"
	"sort"
	"strings"
)

// The READMEs, REFERENCES.md and STATUS.md are GENERATED from the records and compared on every
// run. Hand-keeping a second copy of a fact beside the record that holds it is the drift this
// package exists to refuse, so the prose is derived and its staleness is a failing test.

// README renders one case's README.md — the human's answer to "where did this page come from".
func README(p Provenance) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", p.Slug)
	fmt.Fprintf(&b, "%s\n\n", p.Why)
	b.WriteString("## Where it came from\n\n")
	fmt.Fprintf(&b, "- Source: %s\n", p.SourceURL)
	fmt.Fprintf(&b, "- Page %d of that document\n", p.SourcePage)
	fmt.Fprintf(&b, "- Publisher: %s (%s)\n", p.Publisher, p.Date)
	fmt.Fprintf(&b, "- Rights: **%s** — %s\n", p.Rights, p.RightsNote)
	if p.RightsCaveat != "" {
		fmt.Fprintf(&b, "- Caveat: %s\n", p.RightsCaveat)
	}
	fmt.Fprintf(&b, "- sha256 of the page as served: `%s`\n", p.PageSha256)
	fmt.Fprintf(&b, "- sha256 of the downloaded file it was cut from: `%s`\n", p.LocalFileSha256)
	fmt.Fprintf(&b, "- %d text objects were removed, so the reader meets pixels\n", p.TextObjectsRemove)
	fmt.Fprintf(&b, "- Added %s\n\n", p.AddedAt)
	b.WriteString("## What it is a specimen of\n\n")
	for _, d := range p.DefectClass {
		fmt.Fprintf(&b, "- %s\n", d)
	}
	b.WriteString("\n## What a correct reader would do\n\n")
	for _, line := range expectLines(p.Expect) {
		fmt.Fprintf(&b, "- %s\n", line)
	}
	b.WriteString("\nThis file is GENERATED from provenance.json. Edit the record, not this.\n")
	return b.String()
}

func expectLines(e Expect) []string {
	var out []string
	if e.Table != nil {
		out = append(out, fmt.Sprintf("the detector %s call this page a table", yesNo(*e.Table)))
	}
	if e.Tables != nil {
		out = append(out, fmt.Sprintf("tables on the page: %d", *e.Tables))
	}
	if e.Rows != nil {
		out = append(out, fmt.Sprintf("rows: %d", *e.Rows))
	}
	if e.Columns != nil {
		out = append(out, fmt.Sprintf("columns: %d", *e.Columns))
	}
	for _, s := range e.MustContain {
		out = append(out, fmt.Sprintf("the reading contains %q, which is printed on the page", s))
	}
	return out
}

func yesNo(b bool) string {
	if b {
		return "SHOULD"
	}
	return "should NOT"
}

// Result is one page's verdict, as the harness measured it.
type Result struct {
	Slug   string
	Failed []string // the expect claims that failed; empty means the reader gets this page right
}

// Status renders STATUS.md — the meter. It is DERIVED from checking each page's expect block
// against the reading, never from a field somebody set by hand: a hand-set status survives the fix
// that makes it false, and then the number counts pages that already work.
//
// It is written by the tagged harness, which is the only thing that has read the pages. The
// default-build gate can check that it names exactly the corpus's slugs, and cannot re-establish
// the verdicts — that is stated here so the weaker check is not mistaken for the stronger one.
func Status(results []Result) string {
	rs := append([]Result(nil), results...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Slug < rs[j].Slug })
	broken := 0
	for _, r := range rs {
		if len(r.Failed) > 0 {
			broken++
		}
	}
	var b strings.Builder
	b.WriteString("# Corpus status\n\n")
	fmt.Fprintf(&b, "**%d of %d pages do not yet read as they should.**\n\n", broken, len(rs))
	b.WriteString("Derived from each page's `expect` block, checked against the reading by the tagged\n")
	b.WriteString("harness. GENERATED — edit a record or fix the reader, not this file.\n\n")
	b.WriteString("| page | reads correctly | what still fails |\n|---|---|---|\n")
	for _, r := range rs {
		if len(r.Failed) == 0 {
			fmt.Fprintf(&b, "| %s | yes | — |\n", r.Slug)
			continue
		}
		fmt.Fprintf(&b, "| %s | **no** | %s |\n", r.Slug, strings.Join(r.Failed, "; "))
	}
	return b.String()
}

// SlugsIn returns the slugs STATUS.md names, so a default-build gate can compare them with the
// corpus without pretending to know the verdicts.
func SlugsIn(status string) []string {
	var out []string
	for _, line := range strings.Split(status, "\n") {
		if !strings.HasPrefix(line, "| ") || strings.HasPrefix(line, "| page ") || strings.HasPrefix(line, "|---") {
			continue
		}
		f := strings.Split(line, "|")
		if len(f) < 2 {
			continue
		}
		out = append(out, strings.TrimSpace(f[1]))
	}
	sort.Strings(out)
	return out
}

// References renders REFERENCES.md: the sources we want and may not commit.
func References(refs []Reference) string {
	rs := append([]Reference(nil), refs...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })
	var b strings.Builder
	b.WriteString("# Sources not committed\n\n")
	b.WriteString("Pages we would want and may not redistribute, and leads that could not be reached.\n")
	b.WriteString("Kept as a record so the next round does not re-find them and assume.\n")
	b.WriteString("GENERATED from references.json.\n\n")
	for _, r := range rs {
		fmt.Fprintf(&b, "## %s\n\n", r.Name)
		fmt.Fprintf(&b, "- %s\n", r.URL)
		fmt.Fprintf(&b, "- Wanted for: %s\n", r.WhyWanted)
		fmt.Fprintf(&b, "- Rights: %s\n", r.Rights)
		fmt.Fprintf(&b, "- Blocked because: %s\n\n", r.BlockedReason)
	}
	return b.String()
}
