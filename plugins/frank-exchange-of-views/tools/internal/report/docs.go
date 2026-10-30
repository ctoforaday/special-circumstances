package report

// THE REPORT IS A SET, NOT A FILE.
//
// Measured on the two archived runs, 70–76% of the single assembled report.md was process
// record — the debate transcript, the board in full, a friction log LARGER THAN THE ENTIRE
// RESEARCH ARGUMENT (60–65 KB against 15 KB of analysis), a change history, cost. The research
// the run was commissioned to produce was 24–30% of the document it was delivered in. Six
// audiences were unioned into one artifact, so no reader could be addressed and none could be
// revised, linked or archived without the other five.
//
// Splitting costs nothing in provenance, and that is not an opinion about markdown: the run
// archive carries records/ and proofs/ ONLY, and Assemble re-derives the document from the
// event log. The report was always a projection. This file makes it one projection per audience
// instead of one for all of them, from the same composers, in the same pass.
//
// WHAT IS NOT DONE HERE: nothing is authored, nothing is summarized, and nothing is dropped.
// Every section that shipped in the single file still ships, in exactly one document, with a
// link bar that names the others. A reader who wants the union still has it — it is a
// directory now, with an index.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Doc is one document in a run's report set: the file it lands in, how the link bar and the
// index name it, and the composed markdown body (no link bar, no title — Render adds those).
type Doc struct {
	File  string
	Nav   string // the word in the link bar
	Title string // the document's own H1, after the run's short title
	Blurb string // one line in the index: what a reader comes here for
	Body  string
}

// The canonical file names. They are constants because other carriers — the capture screens,
// the audit skills, the archive — name these files, and a name assembled at the call site is
// the defect this suite has a rule about.
const (
	FileReport    = "report.md"
	FileDocket    = "docket.md"
	FileDebate    = "debate.md"
	FileJudgments = "judgments.md"
	FileInquiries = "inquiries.md"
	FileEvidence  = "evidence.md"
	FileRun       = "run.md"
	FileChangelog = "CHANGELOG.md"
	FileIndex     = "README.md"
)

// docOrder is the reading order, and the link bar's order. Research first, process behind it.
var docOrder = []string{FileReport, FileDocket, FileDebate, FileJudgments, FileInquiries, FileEvidence, FileRun, FileChangelog}

// Files is the document set's names, for the readers OUTSIDE this package — the capture
// screens, the audits, the archive. They exist because a check that reads report.md alone now
// measures one document of the deliverable and reports the others as clean.
func Files() []string { return append([]string(nil), docOrder...) }

// AssemblyWritten is whether the bench's assembly sitting assembled the run's documents: run.md,
// the run document, is on disk. Capture re-assembles the set at the close only for a run where it
// is, and `engine-stopped` sends a stopped run to capture only where it is — one fact, two
// readers, so the two cannot disagree about whether a run reached assembly.
//
// WHAT run.md PROVES. The bench's `assemble` is the one writer of the set outside capture, and the
// record refuses it in every sitting but one registered for the assemble occasion, on a run not yet
// captured (record.RequireAssemblySitting), which is the one sitting the outcome may be stamped in
// as well. So run.md on disk says an assembly sitting assembled the run. It does not say the
// outcome came first — the prompt orders that, and the record holds the outcome's own event — nor
// which assembly sitting: a resumed one assembles over the same files.
//
// IT IS A FILE BECAUSE THE RECORD HOLDS NO SUCH FACT. Assembly writes documents and no event, and
// the assembly sitting's own close is no substitute: the sitting's span closes when its seat stops,
// whether or not `assemble` ran or succeeded.
//
// IT IS THE WHOLE OF WHAT ASSEMBLY OWES. Assemble composes every document from the record before
// it writes any, so run.md on disk means the record assembled; and capture re-assembles the set
// from the record as it stands at the close, so a declaration, a certification or a log entry
// recorded after it reaches the documents that ship all the same.
//
// A FILE THAT CANNOT BE READ IS THE ERROR, never false: an unreadable run directory is not a run
// that never assembled, and a caller that folded the two would resume a finished run or capture
// an unfinished one on a guess.
func AssemblyWritten(run record.Run) (bool, error) {
	switch _, err := os.Stat(filepath.Join(run.Dir(), FileRun)); {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}

// AssembleShipped is the run's report set as it ships (shipSet), under the run's title. It writes
// nothing; Write does that. A document with an empty body is not in the set — a run with no motions
// has no judgments.md, and no empty heading standing in for one.
func AssembleShipped(run record.Run) (Shipped, error) {
	rd := stdReaders()
	docs, _, title, err := rd.assemble(run, "")
	if err != nil {
		return Shipped{}, err
	}
	return rd.shipSet(title, docs), nil
}

// assemble composes the run's report set from the record and the report's text, before it ships:
// each document's body, with the proofs' marks in evidence.md where their anchors go (proofanchor.go)
// and the record's text as written, no `<` escaped. It returns the headings of the rows it composed a
// section for too: what the assembled layout reads to tell a section the table placed from a line of
// the report's text that opens as one does. The layout passes a mark line to stand above each block
// composed, which is how it finds where each stands; the set that ships is assembled with none. And
// it returns the run's short title — the H1 every document in the set opens with — read from the same
// parse of the report's text as the rest.
func (rd readers) assemble(run record.Run, mark string) ([]Doc, map[string]bool, string, error) {
	blue, err := blueReport(run)
	if err != nil {
		return nil, nil, "", err
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		return nil, nil, "", fmt.Errorf("reading the board: %w", err)
	}
	bj, err := record.BoardJSONOfRun(run)
	if err != nil {
		return nil, nil, "", fmt.Errorf("encoding the board: %w", err)
	}
	evs := fam.Events

	// EVERY SECTION SHIPS BECAUSE THE TABLE HAS A ROW FOR IT, in the row's document and the row's
	// place (layout.go). report.md is THE RESEARCH: what the run was asked for, and of how it was
	// run only the stamp — the outcome's word and its basis as state. Every sentence explaining a
	// verdict, what answered the seats, the board in full and the transcript are envelope, and
	// their rows put them in the documents behind it.
	a := &assembly{run: run, fam: fam, bj: bj, outcome: outcomeOf(evs), blue: part(rd.text(blue)), rd: rd}
	// The title enters composition here, so its tokens leave here (stripTokens), by the weave's rule:
	// the Question line composes the title into a section the weave reads, so a citation's or a
	// proof's token stays for the weave to make a note of. The H1 is read by no weave, so it is cut
	// from the title with no token in it: a token's bytes are no part of the length a reader sees,
	// and the cut is at a boundary in what ships.
	a.blue.title = rd.stripTokens(a.blue.title, true)
	title, _ := heading(rd.stripTokens(a.blue.title, false))
	body := map[string]*sections{}
	composed := map[string]bool{}
	for _, s := range layout {
		if body[s.File] == nil {
			body[s.File] = &sections{}
		}
		if s.compose != nil {
			block := s.block(a)
			if mark != "" && strings.TrimSpace(block) != "" {
				block = mark + "\n" + block
			}
			composed[s.Heading] = body[s.File].add(block)
		}
	}
	// inquiries.md — THE INQUIRIES, and the one home of each inquiry's PATH. report.md files an
	// inquiry by its final fate and carries neither who recorded it, the path it took nor the ruling
	// beside it, because those are the debate. It is the SAME rendering `show inquiries` prints
	// (view.InquiryBody), whole, with each superseded record of an inquiry's path under it and each
	// value's paragraphs kept inside its item (inItem) — not a section of the table and not a second
	// account — followed by the chair's reviews of the report against those inquiries (inquiryReviews).
	inq := view.InquiryBody(evs, fam.At, true, inItem)
	if reviews := inquiryReviews(evs, fam.At); reviews != "" {
		if inq != "" {
			inq = strings.TrimRight(inq, "\n") + "\n\n"
		}
		inq += reviews
	}
	inq = rd.escapeNotesRead(rd.stripTokensCode(inq, true))

	docs := []Doc{
		{File: FileReport, Nav: "Report", Title: "", Body: body[FileReport].String(),
			Blurb: "the research: the verdict and what it means, the Catechism, the argument, the risks, the open questions"},
		// THE FILE KEEPS ITS NAME AND THE DOCUMENT GETS A TRUE ONE. `docket.md` is a published
		// URL — the shipped READMEs, the skill, the bench's constitution and the site's link
		// rewriter all name it — so renaming the file buys consistency in a string no machine
		// reads. The nav, title and blurb are what a human reads beside the heading, and all
		// three said this was red's when three parties write into it.
		{File: FileDocket, Nav: "Board", Title: "the board",
			Blurb: "the board all three parties wrote: every gap red minted, with red's argument for it, the fix it required and the check that would settle it, and how each was closed, with the closing lens's attestation; blue's correctness manifest for the repairs it made, and red's spot-checks", Body: body[FileDocket].String()},
		{File: FileDebate, Nav: "Debate", Title: "the debate",
			Blurb: "the transcript, epoch by epoch — red's audits, blue's answers, the closings, red's spot-checks, and the bench's terminal disposition", Body: body[FileDebate].String()},
		{File: FileJudgments, Nav: "Judgments", Title: "judgments",
			Blurb: "every contested question and how it was answered: grade motions, petitions, and the bench's opinions", Body: body[FileJudgments].String()},
		{File: FileInquiries, Nav: "Inquiries", Title: "inquiries",
			Blurb: "every inquiry the research proposed, took, deferred, declined or abandoned: the path each one took, the seat that last recorded it and red's ruling, then each of the chair's reads of the report against them, with its reason", Body: inq},
		{File: FileEvidence, Nav: "Evidence", Title: "evidence",
			Blurb: "the computations this run ran, with the exact script, the output, the sha256, and red's independent re-run; then each citation with the sentence it backs and every reading of it — its outcome, its confidence and what was read", Body: ""},
		{File: FileRun, Nav: "Run", Title: "the run",
			Blurb: "how the machinery behaved: what the seats logged, the record's own invariant check, and what the run cost", Body: body[FileRun].String()},
		{File: FileChangelog, Nav: "Changelog", Title: "changelog",
			Blurb: "the provenance of this report: every claim blue retired from it, every ask a later certification superseded, and every edit blue made to it, with blue's reason and the span it replaced", Body: body[FileChangelog].String()},
	}

	// THE EVIDENCE LAYER IS RESOLVED ACROSS THE SET, NOT WITHIN A DOCUMENT.
	//
	// A footnote definition cannot cross a file boundary — that is a fact about markdown, not a
	// preference — so weaving globally and splitting afterwards would ship dangling references
	// in every document but one. Each document therefore numbers and defines the citations
	// it actually contains, and the proof layer is numbered ONCE for the whole run (record
	// order) so that a P3 in the debate and a P3 in the report are the same computation.
	sources, err := record.CitedSources(run)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cited sources: %w", err)
	}
	proofs, err := record.RecordedProofs(run)
	if err != nil {
		return nil, nil, "", fmt.Errorf("recorded proofs: %w", err)
	}

	used := map[string]bool{}
	for i := range docs {
		if docs[i].Body == "" {
			continue
		}
		// Every token the weave does not take left the text as it entered composition (stripTokens).
		body := collapseBlanks(docs[i].Body)
		woven := weaveCitations(body, sources)
		// The weave writes the notes and the bibliography when the document cites, and returns
		// it as it came when it does not.
		if docs[i].File == FileReport {
			composed[bibliographyHeading] = woven != body
		}
		body, hit := weaveComputationRefs(collapseBlanks(woven), proofs)
		for _, label := range hit {
			used[label] = true
		}
		docs[i].Body = collapseBlanks(body)
	}
	// evidence.md is composed LAST, from the proofs the set actually anchors, so a proof no
	// document references is still shown (it is on the record) and a reference to a proof that
	// is not on the record is still stated (it is a defect, and a silent drop reads as clean).
	// The citations follow the proofs, each with every reading of it.
	for i := range docs {
		if docs[i].File == FileEvidence {
			proofSec, citeSec := evidenceDoc(run, proofs, used), citationsSection(fam)
			composed[evidenceHeading], composed[citationsHeading] = proofSec != "", citeSec != ""
			docs[i].Body = rd.escapeNotes(strings.TrimRight(strings.Join(slices.DeleteFunc([]string{proofSec, citeSec}, func(s string) bool { return s == "" }), "\n"), "\n"))
		}
	}

	out := slices.DeleteFunc(docs, func(d Doc) bool { return strings.TrimSpace(d.Body) == "" })
	return out, composed, title, nil
}

// NO TOKEN SHIPS IN PROSE, AND IT IS MADE SO WHERE TEXT ENTERS COMPOSITION. A seat copies a token into
// record prose — a finding's location, a proof's reason, a citation's title — and the report's text
// carries the ones the tool placed. Every such text passes stripTokens as it enters a document: each
// section a row composes, before its notes are escaped (Section.block); the inquiries document; and
// each field the tool writes into a line after the weave — a note's text, a Bibliography entry, a
// field of evidence.md (plain). Stripped later, a token's removal joined the bytes either side of it
// into syntax nothing had escaped: `[<!--gap:G-00000001-->^1]` became a live `[^1]`. A token in a
// seat's code ships as typed: code is the bytes a seat quotes, a script as it ran, and the proofs'
// scripts and outputs in evidence.md are no field at all.

// stripTokens is md with every anchor token as written (anchor.StripWritten) taken out but those in
// code and, where weave is set, a token of a real id the weave turns into a note (a citation's or a
// proof's): the text a woven document is composed from keeps those for the weave, and every other
// text keeps none. Code is what both tiers read as code (bothTierCode), the reading the note escape takes:
// a token one tier reads as prose would ship there as prose.
//
// A LINE THE STRIP LEAVES A HUSK GOES WHOLE, AND ONLY THAT LINE. A bullet an edit cut down to its
// finding anchors, which no retire took, is tokens and a terminator; stripped, it printed "- ." in
// universe-m19's verified report. The line that goes is that husk (anchor.Husk): a list marker at
// most, sentence punctuation, and the brackets or emphasis that held the text. A line of symbols that
// held a token is content — a table's delimiter row, a row of marks, a rule — and ships without the
// token; a line of code keeps its token, so the strip leaves it no husk.
func (rd readers) stripTokens(md string, weave bool) string {
	out, _, _ := rd.stripTokensCode(md, weave)
	return out
}

// stripTokensCode is stripTokens, with the stretches of code it read md's tokens in when it took none
// out: the text is md then, and its code is the code escapeNotesRead asks for. read is false where
// nothing was read, or where the strip changed the text and with it what is code.
func (rd readers) stripTokensCode(md string, weave bool) (out string, code [][2]int, read bool) {
	if !holdsToken(md) {
		return md, nil, false
	}
	code = rd.bothTierCode([]byte(md))
	out = stripOutside(md, code, weave)
	return out, code, out == md
}

// stripTokensInLine is stripTokens for a fragment the tool writes on one line after its own text
// (plain), with no token woven: in place the fragment is inline text of that line, so the code in it
// is what each tier reads as code in one line of a paragraph (lineReader, codeLineReader), as
// escapeNotesInLine reads it. Read as a document of its own, "```x <!--gap:G-00000001-->" is a fence
// whose info string holds the token, which in place is prose.
func stripTokensInLine(s string) string { return lineReaders.stripTokens(s, false) }

// stripTokensAfter is stripTokens, with no token woven, for text the tool writes after its own on a
// line, which may run on to lines of its own: in place its first line is inline text behind the
// tool's, so its code is what both tiers read as code in the text behind a character of prose — a
// first line opening "```" opens no fence there, and a token in what alone would be its info string is
// prose. The husk rule reads the text itself.
func stripTokensAfter(s string) string {
	if !holdsToken(s) {
		return s
	}
	code := stdReaders().bothTierCode([]byte("x" + s))
	for i := range code {
		code[i] = [2]int{code[i][0] - 1, code[i][1] - 1}
	}
	return stripOutside(s, code, false)
}

// holdsToken reports whether md holds an anchor token as written.
func holdsToken(md string) bool {
	held := false
	anchor.EachWritten(md, func(int, int, string) { held = true })
	return held
}

// stripOutside is md with each token as written taken out but those inside code, the stretches in
// order, and where weave is set the woven ones.
func stripOutside(md string, code [][2]int, weave bool) string {
	return anchor.StripWritten(md, func(start, end int) bool {
		return inSpan(code, start, end) || weave && wovenToken(md[start:end])
	})
}

// StripTokens is a record field with its tokens taken out as assembly takes them from a field it
// writes: the release gate's prose oracle reads a seat's field with it.
func StripTokens(s string) string { return stdReaders().stripTokens(s, false) }

// wovenToken reports whether tok, one token as written, is a token of a real id the weave turns into a
// note: a citation's or a proof's.
func wovenToken(tok string) bool {
	woven := false
	anchor.Each(tok, func(start, end int, id string) {
		woven = start == 0 && end == len(tok) && anchor.AssemblyOf(id) != anchor.Strip
	})
	return woven
}

// blueReport renders the report's text from the record. A run with no base ingested has none, and
// assembles with none of blue's sections. A report the record HOLDS and cannot replay is an error:
// assembled as though blue wrote nothing, the set would ship without a word of the report.
func blueReport(run record.Run) (string, error) {
	have, err := reportproj.BaseIngested(run)
	if err == nil && have {
		var md string
		if md, err = reportproj.RenderFromRecord(run); err == nil {
			return md, nil
		}
	}
	if err != nil {
		err = fmt.Errorf("the report's text: %w", err)
	}
	return "", err
}

// sections accumulates the composed blocks of one document, dropping the empties. It is the
// invariant "no heading ships without a body" expressed as the only way to add one.
type sections struct{ parts []string }

func (s *sections) add(block string) bool {
	if strings.TrimSpace(block) == "" {
		return false
	}
	s.parts = append(s.parts, strings.TrimRight(block, "\n"))
	return true
}

func (s *sections) String() string { return strings.Join(s.parts, "\n\n") }

// masthead is what the document opens with: the run's short title, this document's name, the link bar
// across the set, and a rule. Every file in the set opens identically, so a reader who lands in the
// middle of one knows what run they are in and how to reach the rest.
func masthead(d Doc, title string, set []Doc) string {
	h1 := title
	if d.Title != "" {
		// The siblings drop the template's "research report" suffix rather than stack a second
		// noun on it: "<topic> — research report — the docket" names two documents.
		h1 = strings.TrimSuffix(title, " — research report") + " — " + d.Title
	}
	return h1 + "\n\n" + navBar(d.File, set) + "\n\n---\n\n"
}

// composedText is the document as assembly composed it, its masthead on it and its proofs' anchors in
// place, with the record's text as written: what a seat reads of the set (Layout).
func composedText(d Doc, title string, set []Doc) string {
	return masthead(d, title, set) + resolveMarks(strings.TrimRight(d.Body, "\n")+"\n")
}

// Shipped is a report set as it ships, each document read once per tier: its .md text, masthead and
// body shipped whole (ship), and its body as report.html reads it, which the page renders under its
// own header. It is made only by shipSet, so neither tier can be handed a text the decision was not
// taken on.
type Shipped struct {
	title string
	docs  []Doc
	md    []string     // each document as report.md and its siblings ship it
	html  []ParsedText // and its body as report.html renders it: the tree alone (tree)
}

// shipSet takes each tier's reading of each document of the set, under the run's title, reading with rd.
func (rd readers) shipSet(title string, docs []Doc) Shipped {
	set := Shipped{title: title, docs: docs}
	for _, d := range docs {
		body := strings.TrimRight(d.Body, "\n") + "\n"
		set.md = append(set.md, rd.ship(masthead(d, title, docs)+body))
		// The reading tier shows a marker byte as the replacement character; the durable tier ships
		// the text as written.
		set.html = append(set.html, rd.tree(unmarked(body)))
	}
	return set
}

// navBar is the closest thing markdown has to a tab strip: a row of sibling links with the
// current document named but not linked. There are no tabs in CommonMark or GitHub-flavored
// markdown and there is no honest way to fake them — <details> is an accordion, and a folded
// 60 KB section is still 60 KB in the file, the diff and the grep. report.html has real tabs
// over the same set; this is the durable tier.
func navBar(current string, set []Doc) string {
	var cells []string
	for _, d := range set {
		if d.File == current {
			cells = append(cells, "**"+d.Nav+"**")
			continue
		}
		cells = append(cells, "["+d.Nav+"]("+d.File+")")
	}
	return strings.Join(cells, " · ")
}

// indexDoc is the run directory's front door: what this run asked, what it answered, and which
// document holds what. It is written for a human opening an archived run months later with no
// memory of it, which is the only reader a run directory reliably gets.
func indexDoc(run record.Run, title string, set []Doc, fam record.Family) string {
	var b strings.Builder
	b.WriteString(title + "\n\n")
	b.WriteString(navBar("", set) + "\n\n---\n\n")
	b.WriteString(factBox(fam) + "\n\n")
	b.WriteString("## The documents\n\n")
	for _, d := range set {
		fmt.Fprintf(&b, "- **[%s](%s)** — %s\n", d.Nav, d.File, d.Blurb)
	}
	if _, err := os.Stat(filepath.Join(run.Dir(), "report.html")); err == nil {
		b.WriteString("\n[**report.html**](report.html) is the same set with real tabs and cross-document links — one self-contained file, no server and no network.\n")
	}
	b.WriteString("\nThe record these were rendered from is `records/`; the computations are cached under `proofs/`. Both survive the run; every document here can be rebuilt from them with `feov-record --seat-id <bench seat id> assemble --run <run dir>` — the bench's seat id is the `judge` one this run recorded.\n")
	return b.String()
}

// factBox answers "what is this run and how much do I trust it" in one glance. Every cell is a
// field off the record — nothing here is derived from the prose it sits above.
func factBox(fam record.Family) string {
	open, closed := 0, 0
	// EPOCHS: how many times the chair sat — the largest epoch the record holds an act in, floored
	// by the epochs the gaps were minted and closed in, for a caller holding a board and no log.
	epochs := 0
	for _, g := range fam.Gaps {
		if g == nil {
			continue
		}
		if g.Open {
			open++
		} else {
			closed++
		}
		if g.Epoch > epochs {
			epochs = g.Epoch
		}
		if g.ClosedEpoch > epochs {
			epochs = g.ClosedEpoch
		}
	}
	for _, e := range fam.Events {
		if w := fam.At.Of(e); w.Epoch > epochs {
			epochs = w.Epoch
		}
	}
	outcome := "_(none recorded)_"
	if o := outcomeOf(fam.Events); o != nil {
		outcome = outcomeWord(o)
	}
	var b strings.Builder
	b.WriteString("| | |\n|---|---|\n")
	fmt.Fprintf(&b, "| **Outcome** | %s |\n", outcome)
	fmt.Fprintf(&b, "| **Epochs** | %d |\n", epochs)
	fmt.Fprintf(&b, "| **Gaps** | %d open · %d closed |\n", open, closed)
	return b.String()
}

// supersededAsks is the changelog's account of a bench that spoke more than once. "Read this
// first" carries the terminal statement alone; the ones it replaced are history, and history
// belongs in a document rather than stacked under a heading that reads as parallel asks.
func supersededAsks(evs []*record.Event) string {
	// The statement the report carries is the LAST one; every other, replaced by a later
	// certification, is listed here, in the order it was made, never dropped.
	var all []string
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Certify](e); ok {
			if s := strings.TrimSpace(c.GetStatement()); s != "" {
				all = append(all, s)
			}
		}
	}
	var rows []string
	if len(all) > 1 {
		rows = all[:len(all)-1]
	}
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("The LAST statement is the one the report carries; these are the ones it replaced, in the order they were made.\n\n")
	for i, s := range rows {
		fmt.Fprintf(&b, "%d. %s\n\n", i+1, oneLine(s))
	}
	return strings.TrimRight(b.String(), "\n")
}

// Write renders the set to disk and returns report.md's path. Files in the set that this run
// produced no content for are REMOVED rather than left stale: a run re-assembled after its
// motions were withdrawn must not keep yesterday's judgments.md next to today's report.
func Write(run record.Run, set Shipped, index string) (string, error) {
	want := map[string]bool{FileIndex: true}
	for i, d := range set.docs {
		want[d.File] = true
		path := filepath.Join(run.Dir(), d.File)
		// Each document's link bar is over the set shipSet was given, which is the set written here,
		// so it can never link a file nobody wrote.
		if err := os.WriteFile(path, []byte(set.md[i]), 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", d.File, err)
		}
	}
	if err := os.WriteFile(filepath.Join(run.Dir(), FileIndex), []byte(ship(index)), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", FileIndex, err)
	}
	for _, name := range docOrder {
		if want[name] {
			continue
		}
		if err := os.Remove(filepath.Join(run.Dir(), name)); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("remove stale %s: %w", name, err)
		}
	}
	return filepath.Join(run.Dir(), FileReport), nil
}
