package report

import (
	"fmt"
	"regexp"
	"strings"

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
// THE REPORT'S TEXT CARRIES NO SECTION THE TOOL COMPOSES. universe-m20's synthesizer wrote a
// "## Alternatives considered" by hand; assembly composed its own from the avenues and left the
// hand copy out in silence, with a proof, two gaps and thirteen findings anchored in it. The
// write refuses that heading now, and each refusal carries its row's Path: how content does
// reach the section, or that no act of the refused seat's does and who fills it.
//
// ONE FIELD SAYS WHICH OF THE TOOL'S HEADINGS IS NOT REFUSED (SubjectHeading), and why. A heading
// composed into report.md is refused because the report would carry two sections of one name. A
// heading composed into another document is refused when what stands under it is an account of
// the run its author cannot know — the board, the debate, the verdict's basis. "Proofs" and
// "Motions" are neither: a section of that name in the report's text meets no section of the
// tool's in report.md, ships whole and once among the report's own sections, and is what a
// report on a theorem or on a legislature calls a section of its own.
//
// THE REPORT'S STRUCTURE IS ITS AUTHOR'S. The table places the three sections a report is asked
// to carry, and every other section of its text ships as the body, in the order written, under
// the heading its author gave it. A row for a heading the report need not have is a slot that
// reads as unfilled: universe-m21's sections named the reader's questions, and it shipped a line
// under two fixed headings saying neither was authored, with its argument after the tool's
// sections under a heading of the tool's.

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
	// fate, on an avenue section, is which avenue statuses it holds.
	fate func(status string) bool
	// compose renders the section's body; nil for one a later pass over the whole document writes.
	compose func(a *assembly, s Section) string
}

// block is the section as assembly places it, or "" for one this run has nothing under.
func (s Section) block(a *assembly) string {
	text := s.compose(a, s)
	if s.Composer == Tool && s.Line == "" && strings.TrimSpace(text) != "" {
		return "## " + s.Heading + "\n\n" + text
	}
	return text
}

// assembly is what one assembly of the set reads: the record once, and the report's text parted
// into its title, its lifted sections and the rest.
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
	liftedFrom   = "the section of this name in the report's text, lifted whole"
	liftedPath   = blueTextPath + " A report with no section of this name ships without one."
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
	{File: FileReport, Heading: "TL;DR", Composer: Blue, From: liftedFrom, Path: liftedPath, compose: lift},
	{File: FileReport, Heading: "The Catechism", Composer: Blue, From: liftedFrom, Path: liftedPath, compose: lift},
	// THE BODY: after the two surfaces that open the report and before every section the tool
	// composes, so the reader meets the argument before the record's account of it.
	{File: FileReport, Heading: bodyHeading, Composer: Blue, embeds: true,
		From:    "every part of the report's text that no other row places, whole and in the order written, each section under the heading its author gave it",
		Path:    blueTextPath + " Each section stands under a heading of its author's choosing, and assembly writes none above it.",
		compose: func(a *assembly, _ Section) string { return a.blue.rest }},
	{File: FileReport, Heading: "Risk matrix", Composer: Tool,
		From:    "the open gaps on the board, one row for each: its problem, its grades and its required fix",
		Path:    "No act of blue's adds a row. A lens's `mint` opens a gap and with it a row, and the lens's `close` takes the row out; blue's `edit` answering the gap is what the lens closes on. A risk the report itself argues for is prose in the section that makes the claim.",
		compose: func(a *assembly, _ Section) string { return riskMatrix(a.bj) }},
	// THREE DESCRIPTIVE AREAS, and every avenue lands in exactly one. See accepted, deferred and
	// rejected in assemble.go for why three and why `rejected` is a complement.
	{File: FileReport, Heading: "Research areas", Composer: Tool, fate: accepted, compose: avenues},
	{File: FileReport, Heading: "Future research directions", Composer: Tool, fate: deferred, compose: avenues},
	{File: FileReport, Heading: "Alternatives considered", Composer: Tool, fate: rejected, compose: avenues},
	{File: FileReport, Heading: "Open questions", Composer: Blue, From: liftedFrom, Path: liftedPath, compose: lift},
	{File: FileReport, Heading: bibliographyHeading, Composer: Tool,
		From: "the citations and proofs anchored in the text: each anchor becomes a numbered note where it stands, the notes follow the body, and each cited source is listed once",
		Path: "Blue's `cite` places a citation on the sentence its source backs, and blue's `prove` a proof on the sentence it establishes. No note is typed by hand under a number, or under P and a number: a footnote of the report's own — an aside that is neither a source nor a computation — takes a label that is a word."},

	{File: FileDocket, Heading: "The board", Composer: Tool,
		From:    "the gaps and findings on the record, blue's correctness manifest and the chair's spot-checks",
		Path:    "No act writes it as text. A lens's `finding` and `mint` open its rows, and the lens's `close` and `regrade` move them; blue's `manifest-row` adds a manifest row and the chair's `spot-check` a spot-check.",
		compose: func(a *assembly, _ Section) string { return boardSection(a.fam) }},
	{File: FileDebate, Heading: "The debate", Composer: Tool,
		From:    "the positions, closings, verdicts and rulings on the record, sitting by sitting",
		Path:    "Blue's and the chair's `position` and `closing`, the chair's `verdict`, and the bench's rulings and `certify`.",
		compose: func(a *assembly, _ Section) string { return debate(a.fam) }},
	{File: FileJudgments, Heading: "Motions", Composer: Tool, SubjectHeading: true,
		From:    "the motions on the record, each with its ruling",
		Path:    "A seat files one with `motion grade file`, `motion docket file` or `motion petition file`, and the seat that hears it answers with `motion grade rule`, `motion docket rule` or `motion petition rule`.",
		compose: func(a *assembly, _ Section) string { return motions(a.fam) }},
	// avenues.md is not a section of this table: it is the rendering `show avenues` prints
	// (view.AvenueBody), whole, under headings that carry each fate's count.
	{File: FileEvidence, Heading: evidenceHeading, Composer: Tool, SubjectHeading: true,
		From: "the proofs on the record: each with its script, its output and red's re-run",
		Path: "Blue's `prove` records a computation and a lens's `reproduce` its re-run."},
	// FIRST in run.md, because it is the fact about the run a reader of the verdict most needs:
	// what answered each seat, measured from the seat's own trajectory (#589).
	{File: FileRun, Heading: "How this run was conducted", Composer: Tool,
		From:    "what answered each seat, measured from the seat's own trajectory",
		Path:    noSeatPath + " A seat reads the run's configuration and never what replied to it, so no seat holds this fact.",
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
		From:    "the bench's statements that a later one replaced or the bench struck",
		Path:    "The bench's `certify`: its last standing statement is the one the debate carries, and each earlier one is listed here.",
		compose: func(a *assembly, _ Section) string { return supersededAsks(a.fam.Events) }},
}

// The avenue sections' From and Path are written from each row's own fate, so the sentence a seat
// reads and the predicate the composer applies are one fact.
func init() {
	for i := range layout {
		s := &layout[i]
		if s.fate == nil {
			continue
		}
		fates := fatesOf(*s)
		s.From = "the avenues on the record that are " + orList(fates) + ", one row for each: the line, its fate and the reason, in their author's words"
		s.Path = "Blue's `avenue propose` puts an approach on the record as `proposed`, and blue's `avenue move` gives it a fate with `--as`. An avenue that is " +
			orList(ticked(fates)) + " is listed under this heading; the heading is not written into the report's text, and neither is a list of approaches under another one."
	}
}

// fatesOf is the avenue statuses a section holds, in the record's order.
func fatesOf(s Section) []string {
	var out []string
	for _, status := range record.Names(record.MustEnum("avenue", "status").Values) {
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

// AvenueSections says which fate lands under which heading, written from the avenue rows.
var AvenueSections = func() string {
	var parts []string
	for _, s := range layout {
		if s.fate == nil {
			continue
		}
		verb := " each one "
		if len(parts) == 0 {
			verb = " holds each avenue that is "
		}
		parts = append(parts, s.Heading+verb+orList(fatesOf(s)))
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}()

// AvenueSectionsRule says whose those sections are, on the pages of the acts that write an avenue.
// A seat told only that its words are "printed in the report under" them writes the section too.
var AvenueSectionsRule = "THE TOOL COMPOSES THOSE SECTIONS, WHOLE, from the avenues on the record: " + AvenueSections +
	". They are not the report's text to write: a report that carries one of those headings is refused, and a list of approaches and their fates kept by hand anywhere in it is a second account that nothing checks against the record."

// AssembledNote is what the page that prints the report's text says about what ships, and the
// way to the assembled form: a seat reading that text reads the authored part of report.md, and
// audits a hand copy of a composed section as the reader's text unless it is told the section is
// the tool's.
var AssembledNote = "THIS IS THE REPORT'S TEXT, NOT THE ASSEMBLED REPORT. Assembly places this text among sections it composes from the record, and none of those is in what this prints: " +
	strings.Join(RefusedHeadings(), ", ") + ". Blue's `ingest` and `edit` refuse report text that carries one. " +
	"--assembled reads the set as the run ships it: every section in shipping order, who composes it and from what, and how content reaches it. " +
	"With --assembled --json each section that ships carries its `text` as the reader gets it, anchors resolved — the form in which a composed section is read or audited: pipe it to `jq` and select on `heading`."

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

// HEADINGS ARE READ ONE WAY, by every reader here: the composer that lifts blue's sections, the
// refusal at the write, and the view that slices the assembled document.
//
// A line opens a section when, outside a code fence, it starts with "## " once its surrounding
// whitespace is trimmed. Two headings name the same section when they agree after their anchors
// and the note references assembly weaves from them are taken out, case is folded, runs of
// whitespace are single spaces and a trailing "(in full)" is dropped. A fence opens and closes on a line starting ``` or ~~~.
func headingKey(trimmed string) (string, bool) {
	if !strings.HasPrefix(trimmed, "## ") {
		return "", false
	}
	return normalizeHeading(strings.TrimPrefix(trimmed, "## ")), true
}

// noteRef is a footnote reference, as the weave writes one where a citation or proof anchor stood.
var noteRef = regexp.MustCompile(`\[\^[^\]]+\]`)

// normalizeHeading folds a heading to its comparison key.
func normalizeHeading(h string) string {
	h = noteRef.ReplaceAllString(anchor.Replace(h, func(string, string) string { return "" }), "")
	h = strings.ToLower(strings.Join(strings.Fields(h), " "))
	return strings.TrimSpace(strings.TrimSuffix(h, " (in full)"))
}

// eachLine visits md's lines with each one's trimmed form and whether it is inside a code fence
// or is one of the fence's own delimiters.
func eachLine(md string, visit func(ln, trimmed string, fenced bool)) {
	fence := false
	for _, ln := range strings.Split(md, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fence = !fence
			visit(ln, t, true)
			continue
		}
		visit(ln, t, fence)
	}
}

// A claim is one place where text takes what is the tool's: the line or the label as the text
// writes it, the row it belongs to, and the key two claims of the same thing share.
type claim struct {
	key, text string
	row       Section
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
	eachLine(text, func(_, t string, fenced bool) {
		if fenced {
			return
		}
		key, isHeading := headingKey(t)
		if isHeading {
			preamble = false
		}
		for _, row := range layout {
			if !row.refused() {
				continue
			}
			if isHeading && key == normalizeHeading(row.Heading) ||
				preamble && row.Line != "" && strings.HasPrefix(strings.ToLower(t), strings.ToLower(row.Line)) {
				out = append(out, claim{row.Heading, t, row})
				return
			}
		}
	})
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
// Bibliography row's Path: the notes are that section's.
var WovenNotesRule = "NO NOTE THE TOOL NUMBERS. Assembly writes the report's notes and numbers them: " + WovenNoteLabels +
	". A footnote reference or definition typed into the report's text under a label of that shape — digits, or P and digits — is refused and nothing is recorded: " +
	"it would ship beside the tool's note of the same label, and a reader is shown one of the two. " + notesRow().Path + " " + NoteInCode

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
	eachLine(text, func(_, t string, fenced bool) {
		if fenced {
			return
		}
		for _, m := range wovenNote.FindAllString(reCodeSpan.ReplaceAllString(t, ""), -1) {
			out = append(out, claim{strings.ToLower(m), m, notesRow()})
		}
	})
	return out
}

// blueText is the report's text, parted: every line of it is in exactly one of these.
type blueText struct {
	title  string            // the first "# " line
	lifted map[string]string // heading key → the first section under each heading a Blue row lifts
	rest   string            // everything else, in the order written
}

// part splits the report's text among the rows that place it. NOTHING IS LEFT OVER: the title
// line, the first section under each lifted heading, and the rest, which ships whole as the body
// — a second section under a lifted heading, text before the first heading and a section under a
// heading the tool composes included. Of the last, a live seat writes only one under a
// SubjectHeading; a record migrated from before the refusal can hold any, and ships it there with
// every anchor it carries.
func part(blue string) blueText {
	lifts := map[string]bool{}
	for _, s := range layout {
		if s.Composer == Blue && s.Line == "" && !s.embeds {
			lifts[normalizeHeading(s.Heading)] = true
		}
	}
	out := blueText{lifted: map[string]string{}}
	var rest, cur []string
	curKey := ""
	flush := func() {
		if curKey != "" {
			out.lifted[curKey] = strings.TrimRight(strings.Join(cur, "\n"), " \t\n")
		}
		cur, curKey = nil, ""
	}
	eachLine(blue, func(ln, t string, fenced bool) {
		if !fenced {
			if key, ok := headingKey(t); ok {
				flush()
				if _, done := out.lifted[key]; lifts[key] && !done {
					curKey = key
				}
			} else if out.title == "" && strings.HasPrefix(t, "# ") {
				out.title = t
				return
			}
		}
		if curKey != "" {
			cur = append(cur, ln)
		} else {
			rest = append(rest, ln)
		}
	})
	flush()
	out.rest = strings.TrimSpace(strings.Join(rest, "\n"))
	return out
}

// lift copies blue's section verbatim. A report with no section of that name ships none, and the
// assembled layout is where a seat reads that the row is not in the set.
func lift(a *assembly, s Section) string { return a.blue.lifted[normalizeHeading(s.Heading)] }
