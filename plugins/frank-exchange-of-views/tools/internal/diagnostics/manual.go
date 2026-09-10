package diagnostics

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// THE MANUAL'S GRAMMAR LIVES WITH ITS READER.
//
// `manual` prints every command on a seat's surface with that command's own help, each page opened
// by a rule and a header naming the command. The survey has to know which pages a seat was shown —
// a seat that read its manual has read every command's help, and a survey that could not see that
// would report the seat as having run everything blind. So the format is not described twice: the
// tool WRITES it with these functions and the survey READS it with the ones beside them, and the
// round trip is tested from the writer's side (internal/cli) against real output.
//
// THE MANUAL IS LEAN, AND THE LEANNESS IS REVERSIBLE BY CONSTRUCTION. A block that repeats word for
// word on several pages — the Global Flags, the friction footer, a flag line several verbs share —
// is printed ONCE in a SHARED section and replaced on each page by a marker line naming it.
// ExpandManual puts every marker back, so "each page is that command's own help" stays a checkable
// statement rather than a claim about a summary.

// ManualCommand is the command's name on every surface.
const ManualCommand = "manual"

// ManualRule is the separator line that opens every page.
const ManualRule = "=============================================================================="

// ManualSharedHeading opens the section holding every lifted block, between the opening line and
// the first page.
const ManualSharedHeading = "SHARED BY THE COMMANDS BELOW — each block is printed ONCE here, and every page it was lifted from carries a marker line ending `→ SHARED §n` exactly where it was:"

// ManualHeader is the line naming the command a page belongs to: the command as a seat would type
// it, with `--help`, so the page is unmistakably that command's. An empty path is the surface's own
// page.
func ManualHeader(bin string, path []string) string {
	return "$ " + strings.Join(append([]string{bin}, path...), " ") + " --help"
}

// ManualSharedLabel is the line that opens lifted block id in the SHARED section.
func ManualSharedLabel(id, pages int) string {
	return fmt.Sprintf("§%d (on %d pages):", id, pages)
}

// ManualMarker is the line a page carries where block id was lifted, in one of two shapes.
//
//	a whole block      "(first words of it) → SHARED §n"  — the line stands for the block
//	a flag entry       "      --reason string   → SHARED §n" — the lead is the entry's own flag
//	                   name WITH ITS PADDING, kept on the page, and §n is its description
//
// The second shape exists because pflag pads every flag set to its own column, so the same flag
// described the same way is a different LINE on pages whose other flags are longer. Lifting the
// description and keeping the padded name on the page merges those, and expansion stays exact.
func ManualMarker(lead string, id int) string {
	return lead + markerArrow + strconv.Itoa(id)
}

const markerArrow = "→ SHARED §"

var (
	manualMarkerRe = regexp.MustCompile(`→ SHARED §(\d+)$`)
	manualLabelRe  = regexp.MustCompile(`^§(\d+) \(on \d+ pages\):$`)
)

// ManualPage is one page of a manual: the command path it names and what that command printed.
type ManualPage struct {
	Path []string
	Body string
}

// ManualPages splits one manual's output into its pages, by the rule-and-header pair that opens
// each. Bodies are as printed — a lifted block is still its marker. ExpandManual resolves them.
//
// A text with no page in it returns none, and the caller has to say so rather than read it as a
// manual that listed nothing — see Survey.ManualUnread.
func ManualPages(text, binName string) []ManualPage {
	lines := strings.Split(text, "\n")
	var out []ManualPage
	var body []string
	// closed says whether a following rule ended the page. The newline that ended its last line
	// is then the one Split consumed as the boundary, and it belongs to the page: without it every
	// page but the last came back one byte short of what the command printed.
	flush := func(closed bool) {
		if len(out) > 0 {
			b := strings.Join(body, "\n")
			if closed && len(body) > 0 {
				b += "\n"
			}
			out[len(out)-1].Body = b
		}
		body = nil
	}
	for i := 0; i < len(lines); i++ {
		if lines[i] == ManualRule && i+1 < len(lines) {
			if path, ok := parseManualHeader(lines[i+1], binName); ok {
				flush(true)
				out = append(out, ManualPage{Path: path})
				i++
				continue
			}
		}
		if len(out) > 0 {
			body = append(body, lines[i])
		}
	}
	flush(false)
	return out
}

// ExpandManual returns the pages with every marker replaced by the block it stands for — each page
// as its command printed it. A marker naming a block the SHARED section does not hold is an error:
// that page has lost a line, and saying so is the point of numbering the blocks.
func ExpandManual(text, binName string) ([]ManualPage, error) {
	units := manualShared(text)
	pages := ManualPages(text, binName)
	for i, p := range pages {
		lines := strings.Split(p.Body, "\n")
		for j, l := range lines {
			m := manualMarkerRe.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			id, _ := strconv.Atoi(m[1])
			u, ok := units[id]
			if !ok {
				return nil, fmt.Errorf("page %q marks SHARED §%d, which the SHARED section does not hold", strings.Join(p.Path, " "), id)
			}
			// A whole block's marker opens with its parenthesised first words and stands for the
			// block; anything else is a flag entry whose padded name stays and whose description
			// comes back after it.
			if strings.HasPrefix(l, "(") {
				lines[j] = u
				continue
			}
			lines[j] = l[:strings.LastIndex(l, markerArrow)] + u
		}
		pages[i].Body = strings.Join(lines, "\n")
	}
	return pages, nil
}

// manualShared reads the SHARED section: each label line, then its block up to the blank line that
// ends it. A lifted block never contains a blank line — it is one paragraph or one flag entry.
func manualShared(text string) map[int]string {
	out := map[int]string{}
	lines := strings.Split(text, "\n")
	in := false
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if l == ManualRule {
			break
		}
		if l == ManualSharedHeading {
			in = true
			continue
		}
		if !in {
			continue
		}
		m := manualLabelRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		id, _ := strconv.Atoi(m[1])
		var block []string
		for i+1 < len(lines) && lines[i+1] != "" && lines[i+1] != ManualRule {
			i++
			block = append(block, lines[i])
		}
		out[id] = strings.Join(block, "\n")
	}
	return out
}

func parseManualHeader(line, binName string) ([]string, bool) {
	rest, ok := strings.CutPrefix(line, "$ ")
	if !ok {
		return nil, false
	}
	f := strings.Fields(rest)
	if len(f) < 2 || f[len(f)-1] != "--help" || !isBin(f[0], binName) {
		return nil, false
	}
	return f[1 : len(f)-1], true
}

// readLineNumber is the prefix the Read tool puts on every line it returns: the line number,
// right-aligned, and a tab.
var readLineNumber = regexp.MustCompile(`(?m)^ *\d+\t`)

// redirectTarget is the file a shell command sends its stdout to, or "" for none — the path after a
// `>` (or `>>`, `1>`, `&>`), the form the prompt prescribes for a manual too long to read inline.
// `2>` is stderr and does not count.
func redirectTarget(command string) string {
	f := strings.Fields(strings.NewReplacer("\"", " ", "'", " ").Replace(command))
	for i, t := range f {
		switch {
		case t == ">" || t == ">>" || t == "1>" || t == "&>":
			if i+1 < len(f) {
				return f[i+1]
			}
		case strings.HasPrefix(t, ">") && !strings.HasPrefix(t, ">&"):
			if p := strings.TrimLeft(t, ">"); p != "" {
				return p
			}
		}
	}
	return ""
}
