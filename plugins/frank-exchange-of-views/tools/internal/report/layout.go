package report

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE REPORT SET'S SECTIONS ARE ONE TABLE, and everything that needs to know a section reads it.
//
// Assembly composes each document by walking `layout` in order, so a section ships because it has
// a row and in the place its row has. The same rows are what blue's ingest and edit refuse in the
// report's text (ComposedIn), what `show report --assembled` prints (Layout), what the help pages
// of those verbs list (RefusedHeadings), and what the lists in blue's dispatch prompts are held
// to. A section added to the report is added here and nowhere else.
//
// THE REPORT'S TEXT CARRIES NO SECTION THE TOOL COMPOSES. Assembly composes
// "## Alternatives considered" from the inquiries and would leave a hand-written copy out in
// silence, so the write refuses that heading, and each refusal carries its row's Path: how content
// does reach the section, or that no act of the refused seat's does and who fills it.
//
// ONE FIELD SAYS WHICH OF THE TOOL'S HEADINGS IS NOT REFUSED (SubjectHeading), and why. A heading
// composed into report.md is refused because the report would carry two sections of one name. A
// heading composed into another document is refused when what stands under it is an account of
// the run its author cannot know — the board, the debate, the verdict's basis. "Proofs" and
// "Motions" are neither: a section of that name in the report's text meets no section of the
// tool's in report.md, ships whole and once among the report's own sections, and is what a
// report on a theorem or on a legislature calls a section of its own.
//
// THE REPORT'S STRUCTURE IS ITS AUTHOR'S, AND ASSEMBLY MOVES NONE OF IT. The report's text ships
// in the order written, each section under the heading its author gave it: the body, and after
// the tool's sections of report.md the section that closes the text. Which of its sections open
// it and which closes it is the table's (first, and the row after the tool's), and the write
// holds the text to that order (OutOfOrderIn), so the order a seat reads is the order that ships.
// A row for a heading the report need not have is a slot that reads as unfilled: universe-m21's
// sections named the reader's questions, and it shipped a line under two fixed headings saying
// neither was authored, with its argument after the tool's sections under a heading of the tool's.

// Who composes a section.
const (
	Blue = "blue" // the report's own text, placed where its row stands
	Tool = "tool" // composed by assembly from the record
)

// Section is one part of the report set.
type Section struct {
	// File is the document of the set the section ships in.
	File string
	// Heading names the section. It ships as "## <Heading>" unless Line is set: assembly writes
	// that line above a section the tool composes (block), and the report's own carries its
	// author's. The body's row (embeds) names what it holds and ships no heading of that name.
	Heading string
	// Line is how a section with no heading opens: it ships as one line that starts with this.
	Line string
	// Composer is Blue or Tool.
	Composer string
	// From finishes "assembly composes it from …" for a Tool section, and "it is …" for a Blue one.
	From string
	// Path is how content reaches the section, in whole sentences: the act and the seat whose act
	// it is, or that no seat's act does.
	Path string
	// SubjectHeading marks a section the tool composes in ANOTHER document under a name a report's
	// own subject uses for a section. The report's text may carry a heading of that name: it is
	// the report's own section, placed where every section of the report's own is. Every other
	// heading the tool composes is refused in the report's text.
	SubjectHeading bool

	// embeds marks the body: the report's own sections, each under the heading its author gave it.
	embeds bool
	// first, on the body, is the sections that open it when the text carries them, in order.
	first []string
	// fate, on an inquiry section, is which inquiry statuses it holds.
	fate func(status string) bool
	// none, on an inquiry section, is the sentence it holds when no inquiry has its fate: what that
	// absence says about the subject, true for that fate alone.
	none string
	// compose renders the section's body; nil for one a later pass over the whole document writes.
	compose func(a *assembly, s Section) string
}

// block is the section as assembly places it, or "" for one this run has nothing under.
func (s Section) block(a *assembly) string {
	text := s.compose(a, s)
	if s.Composer == Tool {
		// WHAT ASSEMBLY COMPOSES OPENS NO FOOTNOTE (notes.go): the weave runs over the document
		// after this, and is the one writer of a note in it.
		text = EscapeNotes(text)
	}
	if s.Composer == Tool && s.Line == "" && strings.TrimSpace(text) != "" {
		return "## " + s.Heading + "\n\n" + text
	}
	return text
}

// assembly is what one assembly of the set reads: the record once, and the report's text parted
// into its title, its body and the section that closes it.
type assembly struct {
	run     record.Run
	fam     record.Family
	bj      record.BoardJSON
	outcome *recordpb.Outcome
	blue    blueText
}

const (
	outcomeLine  = "**Outcome:**"
	questionLine = "**Question:**"

	blueTextPath = "The report's text: the opening synthesis writes it and blue's `edit` amends it."
	bodyHeading  = "The report's own sections"
	noSeatPath   = "No seat's act reaches it."

	// The two sections a pass over the whole document writes, after the rows are composed: the
	// weave (weaveCitations) and the evidence document (evidenceDoc) write these headings.
	bibliographyHeading = "Bibliography"
	evidenceHeading     = "Proofs"
)

var layout = []Section{
	{File: FileReport, Heading: "Title", Line: "# ", Composer: Blue,
		From: "the report's `# ` line",
		Path: blueTextPath + fmt.Sprintf(" Past %d characters assembly cuts it at the first colon, dash or sentence end its author wrote, and the whole of it ships beneath as the Question line.", titleLimit)},
	{File: FileReport, Heading: "Outcome", Line: outcomeLine, Composer: Tool,
		From:    "the outcome on the record and its basis",
		Path:    "No act of blue's or of red's reaches it. The bench records the run's outcome with `outcome`, and this line is that outcome. What the research found is the TL;DR's to say.",
		compose: func(a *assembly, _ Section) string { return outcomeStamp(a.outcome) }},
	{File: FileReport, Heading: "Question", Line: questionLine, Composer: Blue,
		From:    fmt.Sprintf("the Title in full, when it runs past %d characters", titleLimit),
		Path:    blueTextPath,
		compose: func(a *assembly, _ Section) string { _, q := heading(a.blue.title); return q }},
	// THE BODY: before every section the tool composes, so the reader meets the argument before
	// the record's account of it. The two surfaces that open the report are its first sections.
	{File: FileReport, Heading: bodyHeading, Composer: Blue, embeds: true, first: []string{"TL;DR", "The Catechism"},
		From:    "the report's text up to the section that closes it, whole and in the order written, each section under the heading its author gave it",
		Path:    blueTextPath + " Each section stands under a heading of its author's choosing, and assembly writes none above it.",
		compose: func(a *assembly, _ Section) string { return a.blue.body }},
	{File: FileReport, Heading: "Risk matrix", Composer: Tool,
		From:    "the open gaps on the board, one row for each: its problem, its grades and its required fix",
		Path:    "No act of blue's adds a row. A lens's `mint` opens a gap and with it a row, and the lens's `close` takes the row out; blue's `edit` answering the gap is what the lens closes on. A risk the report itself argues for is prose in the section that makes the claim.",
		compose: func(a *assembly, _ Section) string { return riskMatrix(a.bj) }},
	// THREE DESCRIPTIVE AREAS, and every inquiry lands in exactly one. See taken, deferred and
	// setAside in assemble.go for why three and why setAside is a complement.
	{File: FileReport, Heading: "Research areas", Composer: Tool, fate: taken, none: "No approach was put forward or pursued.", compose: inquiries},
	{File: FileReport, Heading: "Future research directions", Composer: Tool, fate: deferred, none: "No approach was kept for later research.", compose: inquiries},
	{File: FileReport, Heading: "Alternatives considered", Composer: Tool, fate: setAside, none: "No approach was weighed and set aside.", compose: inquiries},
	{File: FileReport, Heading: "Open questions", Composer: Blue,
		From:    "the section of this name in the report's text, which closes it",
		Path:    blueTextPath + " The write holds it last in the text, and a report with no section of this name ships without one.",
		compose: func(a *assembly, _ Section) string { return a.blue.close }},
	{File: FileReport, Heading: bibliographyHeading, Composer: Tool,
		From: "the citations and proofs anchored in the text: each anchor becomes a numbered note where it stands, the notes follow the body, and each cited source is listed once",
		Path: "Blue's `cite` places a citation on the sentence its source backs, and blue's `prove` a proof on the sentence it establishes. No note is typed by hand under a number, or under P and a number: a footnote of the report's own — an aside that is neither a source nor a computation — takes a label that is a word."},

	{File: FileDocket, Heading: "The board", Composer: Tool,
		From:    "the gaps and findings on the record, blue's correctness manifest and the chair's spot-checks",
		Path:    "No act writes it as text. A lens's `finding` and `mint` open its rows, and the lens's `close` and `regrade` move them; blue's `manifest-row` adds a manifest row and the chair's `spot-check` a spot-check.",
		compose: func(a *assembly, _ Section) string { return boardSection(a.fam) }},
	{File: FileDebate, Heading: "The debate", Composer: Tool,
		From:    "the positions, closings, verdicts, spot-checks and rulings on the record, sitting by sitting",
		Path:    "The chair's `position`, `verdict` and `spot-check`, the lens's and blue's `closing` on a docketed gap, a lens's `regrade` that keeps a gap's grade, and the bench's rulings and `certify`.",
		compose: func(a *assembly, _ Section) string { return debate(a.fam) }},
	{File: FileJudgments, Heading: "Motions", Composer: Tool, SubjectHeading: true,
		From:    "the motions on the record, each with its ruling",
		Path:    "A seat files one with `motion grade file`, `motion docket file` or `motion petition file`, and the seat that hears it answers with `motion grade rule`, `motion docket rule` or `motion petition rule`.",
		compose: func(a *assembly, _ Section) string { return motions(a.fam) }},
	// inquiries.md is not a section of this table: it is the rendering `show inquiries` prints
	// (view.InquiryBody), whole, under headings that carry each fate's count.
	{File: FileEvidence, Heading: evidenceHeading, Composer: Tool, SubjectHeading: true,
		From: "the proofs on the record: each with its script, its output and red's re-run",
		Path: "Blue's `prove` records a computation and a lens's `reproduce` its re-run."},
	{File: FileEvidence, Heading: citationsHeading, Composer: Tool,
		From: "the citations on the record and red's corroborations: each with the sentence it backs and every reading of it, its outcome, its confidence and what was read; then red's readings of sources the report does not cite",
		Path: "Blue's `cite` records a citation, a lens's `verify` a reading of it, and a lens's `corroborate` a source of its own."},
	// FIRST in run.md, because it is the fact about the run a reader of the verdict most needs:
	// what answered each seat, measured from the seat's own trajectory (#589).
	{File: FileRun, Heading: "How this run was conducted", Composer: Tool,
		From: "what answered each seat, measured from the seat's own trajectory, and the build, hook and route to the run directory each seat registered with, beside the hooks setup found installed",
		Path: "No seat writes it. A seat reads the run's configuration and never what replied to it, so no seat holds the model; " +
			"the build, the hook and the route are what the engine observed at each seat's `register`, and the seat types none of them.",
		compose: func(a *assembly, _ Section) string { return conduct(a.run, a.fam) }},
	// THE ONLY PLACE ANY VERDICT'S GLOSS LIVES, for every outcome and for none: report.md carries
	// the stamp and its basis as state, and every sentence explaining either is here.
	{File: FileRun, Heading: "The verdict's basis", Composer: Tool,
		From:    "the outcome on the record",
		Path:    noSeatPath + " It explains the outcome the bench records with `outcome`.",
		compose: func(a *assembly, _ Section) string { return verdictGloss(a.outcome) }},
	{File: FileRun, Heading: "Log (what the run told the operator)", Composer: Tool,
		From:    "the log entries on the record",
		Path:    "Any seat's `log`: a fact about the run or its tooling goes there.",
		compose: func(a *assembly, _ Section) string { return logSection(a.fam) }},
	// The record's own invariant check, rendered for the human the report is for: a section, never
	// a gate.
	{File: FileRun, Heading: "Record verification", Composer: Tool,
		From:    "the record's own invariant check",
		Path:    noSeatPath,
		compose: func(a *assembly, _ Section) string { return recordVerification(a.fam) }},
	{File: FileChangelog, Heading: "Claims retired", Composer: Tool,
		From:    "the retirements on the record",
		Path:    "Blue's `retire`, which records the claim as it stood and why it went.",
		compose: func(a *assembly, _ Section) string { return withdrawnClaims(a.fam) }},
	{File: FileChangelog, Heading: "Superseded bench statements", Composer: Tool,
		From:    "the bench's statements that a later one replaced",
		Path:    "The bench's `certify`: its last statement is the one the debate carries, and each earlier one is listed here.",
		compose: func(a *assembly, _ Section) string { return supersededAsks(a.fam.Events) }},
	{File: FileChangelog, Heading: "Edits to the report", Composer: Tool,
		From:    "the edits on the record, grouped by the sitting that made each: the gap it answers, blue's reason, and the span it replaced beside the text that replaced it",
		Path:    "Blue's `edit`, which records the span it replaces, the new text and why.",
		compose: func(a *assembly, _ Section) string { return editHistory(a.fam) }},
}

// The inquiry sections' From and Path are written from each row's own fate, so the sentence a seat
// reads and the predicate the composer applies are one fact; and the body's Path names the
// sections the write holds first from the row's own list.
func init() {
	for i := range layout {
		s := &layout[i]
		if len(s.first) > 0 {
			s.Path += " The write holds " + "`## " + strings.Join(s.first, "` and `## ") + "` first among them, in that order, where the text carries them."
		}
		if s.fate == nil {
			continue
		}
		fates := fatesOf(*s)
		s.From = "the inquiries on the record that are " + orList(fates) + ", one row for each, labelled: the question it asks, its hypothesis, its fate with the method it states, and what was found or why it stands there, in their author's words"
		s.Path = "Blue's `inquiry propose` puts an approach on the record as `proposed`, and blue's `inquiry record` gives it a fate with `--as`. An inquiry that is " +
			orList(ticked(fates)) + " is listed under this heading; the heading is not written into the report's text, and neither is a list of approaches under another one."
	}
}

// fatesOf is the inquiry statuses a section holds, in the record's order.
func fatesOf(s Section) []string {
	var out []string
	for _, status := range record.Names(record.MustEnum("inquiry", "status").Values) {
		if s.fate(status) {
			out = append(out, status)
		}
	}
	return out
}

func ticked(words []string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = "`" + w + "`"
	}
	return out
}

// orList writes "a", "a or b", "a, b or c".
func orList(words []string) string {
	if len(words) < 2 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " or " + words[len(words)-1]
}

// InquirySections says which fate lands under which heading, written from the inquiry rows.
var InquirySections = func() string {
	var parts []string
	for _, s := range layout {
		if s.fate == nil {
			continue
		}
		verb := " each one "
		if len(parts) == 0 {
			verb = " holds each inquiry that is "
		}
		parts = append(parts, s.Heading+verb+orList(fatesOf(s)))
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}()

// InquirySectionsRule says whose those sections are, on the pages of the acts that write an inquiry.
// A seat told only that its words are "printed in the report under" them writes the section too.
var InquirySectionsRule = "THE TOOL COMPOSES THOSE SECTIONS, WHOLE, from the inquiries on the record: " + InquirySections +
	". They are not the report's text to write: a report that carries one of those headings is refused, and a list of approaches and their fates kept by hand anywhere in it is a second account that nothing checks against the record."

// AssembledNote is what the page that prints the report's text says about what ships, and the
// way to the assembled form: a seat reading that text reads the authored part of report.md, and
// audits a hand copy of a composed section as the reader's text unless it is told the section is
// the tool's.
var AssembledNote = "THIS IS THE REPORT'S TEXT, NOT THE ASSEMBLED REPORT. Assembly places this text among sections it composes from the record, and none of those is in what this prints: " +
	strings.Join(RefusedHeadings(), ", ") + ". Blue's `ingest` and `edit` refuse report text that carries one. " +
	"--assembled reads the set as the run ships it: every section in shipping order, who composes it and from what, and how content reaches it. " +
	"With --assembled --json each section that ships carries its `text` as assembly composes it — the record's text as written, unescaped, with anchors resolved and notes numbered — the form in which a composed section is read or audited: pipe it to `jq` and select on `heading`. " +
	"The shipped .md differs from that text in one way: a backslash stands before each `<` a markdown reader would take for HTML a seat wrote."

// Sections is the table's rows, in shipping order.
func Sections() []Section { return append([]Section(nil), layout...) }

// refused is whether the report's text may not claim the section: the tool composes it, and its
// name is not one the report's subject shares (SubjectHeading).
func (s Section) refused() bool { return s.Composer == Tool && !s.SubjectHeading }

// RefusedHeadings is the headings the report's text may not carry, in the table's order.
func RefusedHeadings() []string {
	var out []string
	for _, s := range layout {
		if s.refused() {
			out = append(out, s.Heading)
		}
	}
	return out
}

// OwnInTheReport finishes a sentence about a SubjectHeading section, wherever one is described:
// what a section of that name in the report's text is.
const OwnInTheReport = "a section of that name in the report's text is the report's own, and ships whole in " + FileReport + " among the report's own sections"

// SubjectHeadingsNote names the sections the tool composes in other documents under a name the
// report's text may use, on the pages that list the refused ones.
var SubjectHeadingsNote = func() string {
	var parts []string
	for _, s := range layout {
		if s.Composer == Tool && s.SubjectHeading {
			parts = append(parts, s.Heading+" in "+s.File)
		}
	}
	return "Assembly also composes " + strings.Join(parts, " and ") + ", and those names are not refused: " + OwnInTheReport + "."
}()

// HEADINGS ARE READ ONE WAY, by every reader here: the composer that parts the report's text, the
// refusals at the write, and the view that slices the assembled document.
//
// A line opens a section when report.html's reading of the text (pageReader, eachLine) opens an ATX
// heading on it and, its surrounding whitespace trimmed, it starts with "## ". Two headings name the
// same section when they agree after their anchors and the note references assembly weaves from them
// are taken out, case is folded, runs of whitespace are single spaces and a trailing "(in full)" is
// dropped.
func headingKey(l mdLine) (string, bool) {
	if !l.heading || !strings.HasPrefix(l.trimmed, "## ") {
		return "", false
	}
	return normalizeHeading(strings.TrimPrefix(l.trimmed, "## ")), true
}

// normalizeHeading folds a heading to its comparison key. The note references are read as text
// (noteRefsIn): a heading is read one line at a time, with no definition in reach to resolve them.
func normalizeHeading(h string) string {
	h = anchor.Replace(h, func(string, string) string { return "" })
	for _, sp := range slices.Backward(noteRefsIn(h)) {
		h = h[:sp[0]] + h[sp[1]:]
	}
	h = strings.ToLower(strings.Join(strings.Fields(h), " "))
	return strings.TrimSpace(strings.TrimSuffix(h, " (in full)"))
}

// mdLine is one line of a text, as report.html's reading of the whole text has it.
type mdLine struct {
	text, trimmed string // the line, and the line with its surrounding whitespace trimmed
	code          bool   // the line is code: a fenced block's line or delimiter, or an indented block's line
	heading       bool   // an ATX heading opens on the line
}

// eachLine visits md's lines in order, each as report.html's reading of md has it (pageReader): what
// is code and what is a heading is the parser's answer, the one both shipped readings share, so a
// line indented into code, or a fence's info string holding backticks, is what the parser says.
func eachLine(md string, visit func(l mdLine)) {
	src := []byte(md)
	lines := strings.Split(md, "\n")
	starts := make([]int, len(lines))
	for i, at := 1, 0; i < len(lines); i++ {
		at += len(lines[i-1]) + 1
		starts[i] = at
	}
	lineOf := func(off int) int {
		i, found := slices.BinarySearch(starts, off)
		if !found {
			i--
		}
		return i
	}
	code := make([]bool, len(lines))
	heading := make([]bool, len(lines))
	_ = ast.Walk(parsePage(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.Heading:
			if x.Pos() >= 0 && src[x.Pos()] == '#' {
				heading[lineOf(x.Pos())] = true
			}
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			last := -1
			if x.Pos() >= 0 {
				last = lineOf(x.Pos())
				code[last] = true
			}
			for i := 0; i < x.Lines().Len(); i++ {
				last = lineOf(x.Lines().At(i).Start)
				code[last] = true
			}
			if f, ok := x.(*ast.FencedCodeBlock); ok && f.Pos() >= 0 && last+1 < len(lines) && closesFence(src[f.Pos():], lines[last+1]) {
				code[last+1] = true
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	for i, ln := range lines {
		visit(mdLine{text: ln, trimmed: strings.TrimSpace(ln), code: code[i], heading: heading[i]})
	}
}

// closesFence reports whether ln is the delimiter that closes the fence opened at the start of open:
// past its indentation and any quote markers, a run of the fence's character at least as long as the
// one that opened it, and nothing after the run but whitespace.
func closesFence(open []byte, ln string) bool {
	n := 0
	for n < len(open) && open[n] == open[0] {
		n++
	}
	t := strings.TrimSpace(strings.TrimLeft(ln, " \t>"))
	return len(t) >= n && strings.Trim(t, string(open[0])) == ""
}

// A claim is one place where text takes what is the tool's, or stands out of the order the text
// ships in: the line or the label as the text writes it, the row it belongs to, and the key two
// claims of the same thing share. A claim of order (orderClaims) also says what the line stands
// on the wrong side of, and which side.
type claim struct {
	key, text string
	row       Section
	beside    string
	after     bool
}

// added is the first claim now makes that was does not, counted by key.
//
// A WRITE IS REFUSED FOR WHAT IT ADDS, NOT FOR WHAT THE RECORD HOLDS. A record ingested before a
// refusal existed, or migrated, holds claims no live write could make. Asked of the whole planned
// document, the refusal fired on every edit of such a report, the one that takes a claim out
// included, because a second claim still stood: no edit could be recorded at all. The text as it
// stood is the allowance; an edit that leaves the report with no more of a claim than it had is
// not the one that made it.
func added(was, now []claim) (c claim, found bool) {
	held := map[string]int{}
	for _, c := range was {
		held[c.key]++
	}
	for _, c := range now {
		if held[c.key] == 0 {
			return c, true
		}
		held[c.key]--
	}
	return claim{}, false
}

// ComposedIn returns the first line by which text claims a section the report's text may not
// carry (refused) and was, the text as it stood, did not: a heading that names one, or, before the
// first heading, a line opening as one that ships without a heading does. It is returned as the
// text writes it, with that section's row.
func ComposedIn(was, text string) (line string, s Section, found bool) {
	c, found := added(composedClaims(was), composedClaims(text))
	return c.text, c.row, found
}

func composedClaims(text string) (out []claim) {
	preamble := true
	eachLine(text, func(l mdLine) {
		if l.code {
			return
		}
		t := l.trimmed
		key, isHeading := headingKey(l)
		if isHeading {
			preamble = false
		}
		for _, row := range layout {
			if !row.refused() {
				continue
			}
			if isHeading && key == normalizeHeading(row.Heading) ||
				preamble && row.Line != "" && strings.HasPrefix(strings.ToLower(t), strings.ToLower(row.Line)) {
				out = append(out, claim{key: row.Heading, text: t, row: row})
				return
			}
		}
	})
	return out
}

// THE REPORT'S TEXT STANDS IN THE ORDER IT SHIPS IN. Every seat reads the text in the order its
// author wrote it, and a reader gets it in the table's. universe-m21's Catechism was written after
// the argument it answers from ("These four sources…") and shipped ahead of the sections that
// introduce them, where no seat had read it. The write refuses a text out of the table's order
// (OutOfOrderIn), so the order a seat audits is the order that ships.
//
// THE ORDER IS THE ROWS THAT PLACE THE TEXT (textRows): the title line, each section a row names,
// and the body, which is every other section in the order its author chooses. A line of the text
// stands at its row's rank. A row before the body is out of order when something of its own rank
// or a later one stands before it; a row after the body, when a section of the body stands after
// it or a section of its own name before it. So a section the text lacks leaves no place open; a
// second section under a row's name is out of order, the first standing before it; text before the
// first heading is the body's, so a row's section after it is out of order, as the title line is
// after any text at all; and where a section that opens the text stands after the one that closes
// it, the one named is the one that opens it, which is the one to move.

// textRows is the rows that place the report's text, in the table's order.
func textRows() (rows []Section, body int) {
	for _, s := range layout {
		// The Question line is the title again, composed; the text holds no line of its own for it.
		if s.File == FileReport && s.Composer == Blue && (s.Line == "" || s.compose == nil) {
			for _, h := range s.first {
				rows = append(rows, Section{Heading: h})
			}
			if s.embeds {
				body = len(rows)
			}
			rows = append(rows, s)
		}
	}
	return rows, body
}

// orderName is how a row of the text's order is named to a seat.
func orderName(s Section) string {
	switch {
	case s.embeds:
		return "the report's own sections, in the order their author chooses"
	case s.Line != "":
		return "the " + strings.ToLower(s.Heading) + " line (`" + s.Line + "…`)"
	}
	return "`## " + s.Heading + "`"
}

// TextOrder is the order the report's text stands in, as a seat reads it.
var TextOrder = func() string {
	rows, _ := textRows()
	names := make([]string, len(rows))
	for i, s := range rows {
		names[i] = orderName(s)
	}
	return strings.Join(names, "; ")
}()

// TextOrderRule is the rule on the pages of the verbs that write report text.
var TextOrderRule = "ONE ORDER, THE ONE THE REPORT SHIPS IN. The report's text ships in the order it is written, and it stands in this order: " + TextOrder +
	". A text that holds one of those out of its place is refused and nothing is recorded; the refusal names the line, what it stands on the wrong side of, and where it belongs. " +
	"A section the report does not carry leaves no place open: what is refused is a section out of place, never one that is absent. " +
	"A second section under one of those names is out of place, and so is one of them after text that stands before the first heading: that text is one of the report's own sections."

// Misplaced is one line or section the report's text holds out of the order it ships in.
type Misplaced struct {
	Line    string // the title line or the section's heading, as the text writes it
	Beside  string // what it stands on the wrong side of, as the text writes it
	After   bool   // it stands after Beside; before it otherwise
	Belongs string // where the order holds it
}

// OutOfOrderIn returns the first line text holds out of the order it ships in (TextOrder) that
// was, the text as it stood, did not (added).
func OutOfOrderIn(was, text string) (Misplaced, bool) {
	c, found := added(orderClaims(was), orderClaims(text))
	if !found {
		return Misplaced{}, false
	}
	rows, body := textRows()
	at := slices.IndexFunc(rows, func(s Section) bool { return s.Heading == c.key })
	others, side := rows[:at], "before"
	if at > body {
		others, side = rows[at+1:], "after"
	}
	belongs := "nothing in the report's text stands " + side + " it"
	if len(others) > 0 {
		names := make([]string, len(others))
		for i, s := range others {
			names[i] = orderName(s)
		}
		belongs = "nothing in the report's text but " + strings.Join(names, " and ") + " stands " + side + " it"
	}
	return Misplaced{Line: c.text, Beside: c.beside, After: c.after, Belongs: belongs}, true
}

// A stand is one thing the order places, where the text holds it: the title line, a section's
// heading, or the first line of text before the first heading, which opens the body.
type stand struct {
	rank int    // its row's place in textRows
	line string // as the text writes it
	at   int    // the line's index in the text
}

func stands(text string) (out []stand, rows []Section, body int) {
	rows, body = textRows()
	preamble, titled, at := true, false, -1
	eachLine(text, func(l mdLine) {
		at++
		t, fenced := l.trimmed, l.code
		key, isHeading := "", false
		if !fenced {
			key, isHeading = headingKey(l)
		}
		switch {
		case isHeading:
			rank := body
			for i, s := range rows {
				if s.Line == "" && !s.embeds && normalizeHeading(s.Heading) == key {
					rank = i
				}
			}
			preamble = false
			out = append(out, stand{rank, t, at})
		// The title is the first such line wherever it stands, as assembly reads it (part), and
		// its row is the table's first.
		case !fenced && !titled && strings.HasPrefix(t, rows[0].Line):
			titled = true
			out = append(out, stand{0, t, at})
		case preamble && t != "":
			preamble = false
			out = append(out, stand{body, t, at})
		}
	})
	return out, rows, body
}

// orderClaims is each line of a row's that text holds out of order, with the first thing that
// stands on the wrong side of it.
func orderClaims(text string) (out []claim) {
	all, rows, body := stands(text)
	for i, s := range all {
		if s.rank == body {
			continue
		}
		for j, o := range all {
			if j < i && (o.rank == s.rank || s.rank < body && o.rank > s.rank) || j > i && s.rank > body && o.rank >= body && o.rank < s.rank {
				out = append(out, claim{key: rows[s.rank].Heading, text: s.line, row: rows[s.rank], beside: o.line, after: j < i})
				break
			}
		}
	}
	return out
}

// THE NOTES' LABELS ARE THE TOOL'S. The weave writes a citation's note as [^1], [^2] … in order of
// first appearance (weaveCitations) and a proof's as [^P1], [^P2] … in record order
// (weaveProofRefs), into the text as it stands and beside whatever that text already holds. A
// footnote typed by hand under a label of that shape ships next to the tool's note of the same
// label: markdown resolves a label to one definition and report.html lists one entry for it, so
// a reader gets one of the two notes and no sign of the other. A reference typed with no
// definition reads as the tool's note for whichever source or proof comes to hold that number.
// The write refuses the label (WovenNoteIn). A footnote under a label of any other shape is the
// report's own: it ships as written and resolves to its own definition in both tiers.

// citationNote and proofNote are the labels the weave writes.
func citationNote(n int) string { return fmt.Sprintf("[^%d]", n) }
func proofNote(n int) string    { return fmt.Sprintf("[^P%d]", n) }

// wovenNote reads a footnote label of the weave's shape, in a reference or in a definition. Case
// is folded, as a markdown renderer folds a label's.
var wovenNote = regexp.MustCompile(`(?i)\[\^p?[0-9]+\]`)

// WovenNoteLabels names the weave's labels to a seat.
var WovenNoteLabels = "`" + citationNote(1) + "`, `" + citationNote(2) + "` … for citations and `" + proofNote(1) + "`, `" + proofNote(2) + "` … for proofs"

// WovenNotesRule is the rule on the pages of the verbs that write report text. The way in is the
// Bibliography row's Path: the notes are that section's. Why the label is refused: a hand-typed note
// would ship beside the tool's note of the same label, and a reader is shown one of the two — the
// refusal says so (blue/ingest.go); the page states the rule.
var WovenNotesRule = "NO NOTE THE TOOL NUMBERS. Assembly writes the report's notes and numbers them: " + WovenNoteLabels +
	". A footnote reference or definition typed into the report's text under a label of that shape — digits, or P and digits — is refused and nothing is recorded. " + notesRow().Path + " " + NoteInCode

// NoteInCode says how text that is not a footnote carries such a label.
const NoteInCode = "Where such a label is the subject's own text — a character class in a regular expression — it goes in a code span or a fenced block, and is no footnote there."

// notesRow is the row of the section the woven notes belong to.
func notesRow() Section {
	for _, s := range layout {
		if s.Heading == bibliographyHeading {
			return s
		}
	}
	panic("report: the layout has no " + bibliographyHeading + " row")
}

// WovenNoteIn returns the first footnote label of the weave's shape that text carries and was, the
// text as it stood, did not (added), as the text writes it, with the row of the section the notes
// belong to. Code is not read — a fenced block, or a span in backticks: a label there is no
// footnote.
func WovenNoteIn(was, text string) (label string, s Section, found bool) {
	c, found := added(wovenNotes(was), wovenNotes(text))
	return c.text, c.row, found
}

func wovenNotes(text string) (out []claim) {
	src := []byte(text)
	prose := func(s string) {
		for _, m := range wovenNote.FindAllString(s, -1) {
			out = append(out, claim{key: strings.ToLower(m), text: m, row: notesRow()})
		}
	}
	last := 0
	for _, c := range codeIn(parsePage(src), src) {
		prose(text[last:c[0]])
		last = c[1]
	}
	prose(text[last:])
	return out
}

// blueText is the report's text, parted: every line of it is in exactly one of these, and the two
// that hold sections hold them in the order written.
type blueText struct {
	title string // the first "# " line
	body  string // everything else, up to the section that closes the text
	close string // that section, and whatever the text holds after it
}

// part takes the title line out of the report's text and cuts the rest at the heading of the
// section that closes it. ASSEMBLY MOVES NO LINE OF THE TEXT PAST ANOTHER: the tool's sections of
// report.md stand between the two parts, and each part ships as written. The write holds the
// closing section last (OutOfOrderIn); a record migrated with its text as written can hold more
// after it, a second section of its name included, and that ships after it with every anchor it
// carries.
func part(blue string) (out blueText) {
	var parts [2][]string
	closed := 0
	eachLine(blue, func(l mdLine) {
		ln, t, fenced := l.text, l.trimmed, l.code
		if key, ok := headingKey(l); ok && !fenced {
			for _, s := range layout {
				if s.Composer == Blue && s.Line == "" && !s.embeds && key == normalizeHeading(s.Heading) {
					closed = 1
				}
			}
		} else if !fenced && out.title == "" && strings.HasPrefix(t, "# ") {
			out.title = t
			return
		}
		parts[closed] = append(parts[closed], ln)
	})
	out.body, out.close = strings.TrimSpace(strings.Join(parts[0], "\n")), strings.TrimSpace(strings.Join(parts[1], "\n"))
	return out
}
