package diagnostics

import "strings"

// THE MANUAL'S PAGE GRAMMAR LIVES WITH ITS READER.
//
// `manual` prints every command on a seat's surface with that command's own help, each page opened
// by a rule and a header naming the command. The survey below has to know which pages a seat was
// shown — a seat that read its manual has read every command's help, and a survey that could not
// see that would report the seat as having run everything blind, the flattering zero turned the
// other way. So the header is not a format two packages each describe: the tool WRITES it with
// ManualHeader and the survey READS it with ManualPages, and the round trip is tested from the
// writer's side (internal/cli) against real output.

// ManualCommand is the command's name on every surface.
const ManualCommand = "manual"

// ManualRule is the separator line that opens every page.
const ManualRule = "=============================================================================="

// ManualHeader is the line naming the command a page belongs to: the command as a seat would type
// it, with `--help`, so the page is unmistakably that command's. An empty path is the surface's own
// page.
func ManualHeader(bin string, path []string) string {
	return "$ " + strings.Join(append([]string{bin}, path...), " ") + " --help"
}

// ManualPage is one page of a manual: the command path it names and what that command printed.
type ManualPage struct {
	Path []string
	Body string
}

// ManualPages splits one manual's output into its pages, by the rule-and-header pair that opens
// each. A rule not followed by a header for this binary is body text, not a page boundary.
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
