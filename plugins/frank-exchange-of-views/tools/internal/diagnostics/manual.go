package diagnostics

import (
	"fmt"
	"regexp"
	"slices"
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
// word on several pages — the Global Flags, the log footer, a flag line several verbs share, the
// sentence that closes several flags' descriptions — is printed ONCE in a SHARED section and
// replaced on each page by a marker naming it. Two lines every page would print the same way — the
// Usage that restates the page's own header, and the help flag — are left off and said once.
// ExpandManual puts every one of them back, so "each page is that command's own help" stays a
// checkable statement rather than a claim about a summary.
//
// WHAT IS LEFT OFF IS ONLY WHAT CAN BE PUT BACK EXACTLY. LeanManualPage hands back a page with a
// line removed only when RestoreManualPage rebuilds the original from it byte for byte, and the
// writer expands its own output and compares it with the help it ran before printing a word. A page
// whose shape the restoration does not know keeps its lines.

// ManualCommand is the command's name on every surface.
const ManualCommand = "manual"

// ManualRule is the separator line that opens every page. It is short because it is printed once
// per page into every seat's constitution; the header under it is what names the page.
const ManualRule = "========"

// ManualSharedHeading opens the section holding every lifted block, between the opening line and
// the first page.
const ManualSharedHeading = "SHARED BY THE COMMANDS BELOW — each block is printed ONCE here. A page carries a marker ending `→ SHARED §n` exactly where the block was: alone on its line the marker stands for the whole block, and at the end of a flag's line it stands for the rest of that flag's description. A block labelled the last block of its pages is not marked on them:"

// ManualLeftOff is the line, printed once before SHARED, that says what every page below omits. It
// is also the reader's licence: ExpandManual restores a Usage block and a help flag only to a
// manual that carries this line, so a text that never left them off is read as printed.
const ManualLeftOff = "LEFT OFF EVERY PAGE BELOW, because each page would say the same: the `Usage:` block wherever it is exactly the command in the page's header followed by `[flags]`, so a page that prints a Usage is saying something more; and the flag `-h, --help`, which every command takes and which prints that command's page, so a page with no `Flags:` section has no flag of its own."

// ManualFailedLead opens the line a page carries when its help did not run. Such a page is printed
// as it came — nothing is left off it and nothing is restored to it.
const ManualFailedLead = "THIS PAGE DID NOT RUN"

// ManualWordsHeading opens the section, printed once after the opening line and before SHARED,
// that defines each concept word this surface uses. It sits before the first page, so the page
// parsers above never read it as part of one.
const ManualWordsHeading = "WORDS THIS SURFACE USES — one word for each concept, and it is the same word on every page, prompt and constitution you read:"

// ManualHeader is the line naming the command a page belongs to: the command as a seat would type
// it, with `--help`, so the page is unmistakably that command's. An empty path is the surface's own
// page.
func ManualHeader(bin string, path []string) string {
	return "$ " + strings.Join(append([]string{bin}, path...), " ") + " --help"
}

// ManualSharedLabel is the line that opens lifted block id in the SHARED section.
func ManualSharedLabel(id, pages int) string {
	return fmt.Sprintf(sharedLabelFormat, id, pages)
}

// The two label formats are the writer's and the reader's: sharedLabel scans a line with the same
// format that printed it and accepts it only where printing what it scanned gives the line back.
const (
	sharedLabelFormat  = "§%d (on %d pages):"
	closingLabelFormat = "§%d (the last block of %d pages — every page below that carries no marker for it):"
)

// ManualClosingLabel opens the one block that CLOSES pages without a marker: it is the last block
// of every page that carries no marker for it, and closes is how many pages that is. A marker costs
// a line on every page, and the block every command inherits is on all of them.
func ManualClosingLabel(id, closes int) string {
	return fmt.Sprintf(closingLabelFormat, id, closes)
}

// ManualEntriesLead opens the marker that stands for a lifted block's lines WITHOUT its heading —
// the entries of a flag section, where a page lists them under a heading of its own.
const ManualEntriesLead = "  (every entry under that block's heading) "

// ManualMarker is the line a page carries where block id was lifted, in one of four shapes.
//
//	a whole block      "(first words of it) → SHARED §n"  — the line stands for the block
//	a block's entries  ManualEntriesLead + "→ SHARED §n" — the block's lines after its heading
//	a flag entry       "      --reason string   → SHARED §n" — the lead is the entry's own flag
//	                   name WITH ITS PADDING, kept on the page, and §n is its description
//	a description tail "      --quote text   <its own sentences> → SHARED §n" — the lead also
//	                   keeps the sentences only this flag says, and §n is the ones several share
//
// The flag shapes exist because pflag pads every flag set to its own column, so the same flag
// described the same way is a different LINE on pages whose other flags are longer. Lifting the
// description and keeping the padded name on the page merges those, and expansion stays exact:
// whatever precedes the arrow stays, and the block follows it.
func ManualMarker(lead string, id int) string {
	return lead + markerArrow + strconv.Itoa(id)
}

const markerArrow = "→ SHARED §"

// markedBlock is the id of the SHARED block a line's marker names, and the lead before its arrow.
// A line is a marker only where ManualMarker, given that lead and id, prints the line back.
func markedBlock(l string) (id int, lead string, ok bool) {
	i := strings.LastIndex(l, markerArrow)
	if i < 0 {
		return 0, "", false
	}
	id, err := strconv.Atoi(l[i+len(markerArrow):])
	if err != nil || ManualMarker(l[:i], id) != l {
		return 0, "", false
	}
	return id, l[:i], true
}

// sharedLabel reads a SHARED section label: the block's id, and whether it is the block that
// closes the pages that do not mark it.
func sharedLabel(l string) (id int, closing, ok bool) {
	for _, f := range []string{sharedLabelFormat, closingLabelFormat} {
		var pages int
		if n, _ := fmt.Sscanf(l, f, &id, &pages); n == 2 && fmt.Sprintf(f, id, pages) == l {
			return id, f == closingLabelFormat, true
		}
	}
	return 0, false, false
}

// ManualPage is one page of a manual: the command path it names and what that command printed.
type ManualPage struct {
	Path []string
	Body string
	// bin is the tool's name as the page's header printed it — what a restored Usage line names.
	bin string
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
			if bin, path, ok := parseManualHeader(lines[i+1], binName); ok {
				flush(true)
				out = append(out, ManualPage{Path: path, bin: bin})
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

// ExpandManual returns each page as its command printed it: every marker replaced by the block it
// stands for, the closing block put back on each page that does not mark it, and — in a manual
// that says it left them off — the Usage block and the help flag restored. A marker naming a block
// the SHARED section does not hold is an error: that page has lost a line, and saying so is the
// point of numbering the blocks.
func ExpandManual(text, binName string) ([]ManualPage, error) {
	units, closing := manualShared(text)
	leftOff := slices.Contains(strings.Split(text, "\n"), ManualLeftOff)
	pages := ManualPages(text, binName)
	for i, p := range pages {
		lines := strings.Split(p.Body, "\n")
		marksClosing := false
		for j, l := range lines {
			id, lead, ok := markedBlock(l)
			if !ok {
				continue
			}
			u, ok := units[id]
			if !ok {
				return nil, fmt.Errorf("page %q marks SHARED §%d, which the SHARED section does not hold", strings.Join(p.Path, " "), id)
			}
			marksClosing = marksClosing || id == closing
			switch {
			// A whole block's marker opens with its parenthesised first words and stands for the
			// block.
			case strings.HasPrefix(l, "("):
				lines[j] = u
			case strings.HasPrefix(l, ManualEntriesLead):
				_, lines[j], _ = strings.Cut(u, "\n")
			// Anything else is a flag entry: what precedes the arrow stays — the padded name, and
			// any sentences of its own — and the block comes back after it.
			default:
				lines[j] = lead + u
			}
		}
		body := strings.Join(lines, "\n")
		if !strings.HasPrefix(body, ManualFailedLead) {
			if closing != 0 && !marksClosing {
				body = strings.TrimRight(body, "\n") + "\n\n" + units[closing] + "\n"
			}
			if leftOff {
				body = RestoreManualPage(p.bin, p.Path, body)
			}
		}
		pages[i].Body = body
	}
	return pages, nil
}

var (
	helpFlagLine  = regexp.MustCompile(`^  -h, --help +help for `)
	flagEntryName = regexp.MustCompile(`^\s+(?:-\S, )?--(\S+)`)
	flagEntryCol  = regexp.MustCompile(`^\s+-\S.*?\S\s{2,}`)
)

// ManualFlagsHeading opens a command's OWN flag section on its page. The writer never lifts that
// section whole — a page must still say which flags its command takes — and the reader restores the
// help flag into it.
const ManualFlagsHeading = "Flags:"

const (
	usageHeading = "Usage:"
	// inheritedHeading opens the section cobra prints after a command's own flags; a restored
	// Flags section goes before it.
	inheritedHeading = "Global Flags:"
	helpFlag         = "  -h, --help"
)

func standardUsage(bin string, path []string) []string {
	return []string{usageHeading, "  " + strings.Join(append([]string{bin}, path...), " ") + " [flags]", ""}
}

// LeanManualPage returns a command's help with the lines every page prints the same way left off
// — the Usage block that only restates the command, and the help flag — and only where
// RestoreManualPage rebuilds help from the result byte for byte. A page whose shape the restoration
// would get wrong keeps the line in question, and one no form of which survives the round trip is
// handed back whole — the writer's own read-back is what refuses a page a manual cannot carry.
func LeanManualPage(bin string, path []string, help string) string {
	noHelp := func(s string) string {
		lines := strings.Split(s, "\n")
		i := slices.IndexFunc(lines, helpFlagLine.MatchString)
		if i < 0 {
			return s
		}
		lines = slices.Delete(lines, i, i+1)
		// A Flags section that held nothing else goes with it, and the blank line that ended it.
		if i > 0 && i < len(lines) && lines[i-1] == ManualFlagsHeading && lines[i] == "" {
			lines = slices.Delete(lines, i-1, i+1)
		}
		return strings.Join(lines, "\n")
	}
	noUsage := func(s string) string {
		usage := strings.Join(standardUsage(bin, path), "\n")
		// Where the Flags section went with its only flag, the Usage block is what ends the page.
		if cut, ok := strings.CutSuffix(s, "\n"+usage); ok {
			return cut
		}
		return strings.Replace(s, usage+"\n", "", 1)
	}
	for _, c := range []string{noUsage(noHelp(help)), noHelp(help), noUsage(help)} {
		if RestoreManualPage(bin, path, c) == help {
			return c
		}
	}
	return help
}

// RestoreManualPage puts back what LeanManualPage left off: the help flag, at its sorted place in
// the page's Flags section and padded to that section's column — or as a section of its own where
// the command has no other flag — and the Usage block naming the command, before the Flags section.
// A page that already carries either is given none.
func RestoreManualPage(bin string, path []string, body string) string {
	lines := strings.Split(body, "\n")
	name := bin
	if len(path) > 0 {
		name = path[len(path)-1]
	}
	if !slices.ContainsFunc(lines, helpFlagLine.MatchString) {
		if f := slices.Index(lines, ManualFlagsHeading); f >= 0 {
			// pflag prints a section's flags sorted by name, each description at one column.
			col, put, end := 0, -1, f+1
			for ; end < len(lines) && lines[end] != ""; end++ {
				if col == 0 {
					col = len(flagEntryCol.FindString(lines[end]))
				}
				if m := flagEntryName.FindStringSubmatch(lines[end]); m != nil && put < 0 && m[1] > "help" {
					put = end
				}
			}
			if put < 0 {
				put = end
			}
			lines = slices.Insert(lines, put, helpFlag+strings.Repeat(" ", max(col-len(helpFlag), 3))+"help for "+name)
		} else {
			section := []string{ManualFlagsHeading, helpFlag + "   help for " + name, ""}
			if g := slices.Index(lines, inheritedHeading); g >= 0 {
				lines = slices.Insert(lines, g, section...)
			} else {
				// Nothing follows a command's own flags: the section closes the page.
				for len(lines) > 0 && lines[len(lines)-1] == "" {
					lines = lines[:len(lines)-1]
				}
				lines = append(append(lines, ""), section...)
			}
		}
	}
	if slices.Index(lines, usageHeading) < 0 {
		if f := slices.Index(lines, ManualFlagsHeading); f >= 0 {
			lines = slices.Insert(lines, f, standardUsage(bin, path)...)
		}
	}
	return strings.Join(lines, "\n")
}

// manualShared reads the SHARED section: each label line, then its block up to the next label or
// the first page. A block may hold a blank line — paragraphs that always travel together are lifted
// as one — so it is the next label that ends a block, and the blank lines before it are separator.
// closing is the id of the block labelled as closing pages unmarked, or 0.
func manualShared(text string) (blocks map[int]string, closing int) {
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
		id, closes, ok := sharedLabel(l)
		if !ok {
			continue
		}
		if closes {
			closing = id
		}
		var block []string
		for i+1 < len(lines) && lines[i+1] != ManualRule && !isSharedLabel(lines[i+1]) {
			i++
			block = append(block, lines[i])
		}
		out[id] = strings.TrimRight(strings.Join(block, "\n"), "\n")
	}
	return out, closing
}

func isSharedLabel(l string) bool {
	_, _, ok := sharedLabel(l)
	return ok
}

func parseManualHeader(line, binName string) (bin string, path []string, ok bool) {
	rest, ok := strings.CutPrefix(line, "$ ")
	if !ok {
		return "", nil, false
	}
	f := strings.Fields(rest)
	if len(f) < 2 || f[len(f)-1] != "--help" || !isBin(f[0], binName) {
		return "", nil, false
	}
	return f[0], f[1 : len(f)-1], true
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
