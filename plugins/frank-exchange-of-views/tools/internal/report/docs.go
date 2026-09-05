package report

// THE REPORT IS A SET, NOT A FILE.
//
// Measured on the two archived runs, 70–76% of the single assembled report.md was process
// record — the debate transcript, the board in full, a friction log LARGER THAN THE ENTIRE
// RESEARCH ARGUMENT (60–65 KB against 15 KB of analysis), revision history, cost. The research
// the run was commissioned to produce was 24–30% of the document it was delivered in. Six
// audiences were unioned into one artifact, so no reader could be addressed and none could be
// revised, linked or archived without the other five.
//
// Splitting costs nothing in provenance, and that is not an opinion about markdown: the run
// archive carries records/ and proofs/ ONLY, and Assemble re-derives the document from the
// event log. The report was always a projection. This file makes it seven projections instead
// of one, from the same composers, in the same pass.
//
// WHAT IS NOT DONE HERE: nothing is authored, nothing is summarized, and nothing is dropped.
// Every section that shipped in the single file still ships, in exactly one document, with a
// link bar that names the others. A reader who wants the union still has it — it is a
// directory now, with an index.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
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
	FileEvidence  = "evidence.md"
	FileRun       = "run.md"
	FileChangelog = "CHANGELOG.md"
	FileIndex     = "README.md"
)

// docOrder is the reading order, and the link bar's order. Research first, process behind it.
var docOrder = []string{FileReport, FileDocket, FileDebate, FileJudgments, FileEvidence, FileRun, FileChangelog}

// Files is the document set's names, for the readers OUTSIDE this package — the capture
// screens, the audits, the archive. They exist because a check that reads report.md alone now
// measures a seventh of the deliverable and reports the other six as clean.
func Files() []string { return append([]string(nil), docOrder...) }

// AssembleAll composes the run's whole report set from the record and blue's audited report.
// It writes nothing; Assemble does that. Docs with an empty body are omitted entirely — a run
// with no motions has no judgments.md, and no empty heading standing in for one.
func AssembleAll(run record.Run) ([]Doc, error) {
	blue := blueReport(run)

	board, err := record.BoardState(run)
	if err != nil {
		return nil, fmt.Errorf("assemble: board: %w", err)
	}
	bj := record.BoardJSONOf(board)
	evs := board.Events
	outcome := outcomeOf(evs)

	_, question := heading(blue)

	// report.md — THE RESEARCH. What the run was asked for, and nothing about how it was run
	// except the two facts that bear on trusting it (the verdict's basis, and what actually
	// answered the seats).
	var r sections
	r.add(verdictStamp(outcome))
	if question != "" {
		r.add(question)
	}
	// WHAT ANSWERED, before what was found. A reader deciding how much weight this document
	// carries needs the verdict and the adversary's actual strength in the same breath: a PASS
	// from a tier nobody configured is not the PASS the run was set up to produce.
	r.add(conduct(board))
	// The gloss opens "Read this first" — EXCEPT when there is no outcome, where the stamp
	// already says exactly that and repeating it is the duplication this whole pass removes.
	gloss := ""
	if outcome != nil {
		gloss = verdictGloss(outcome)
	}
	r.add(orientation(board, evs, gloss))
	r.add(sectionOr(blue, "TL;DR"))
	r.add(sectionOr(blue, "The Catechism"))
	r.add(sectionOr(blue, "Technical foundations"))
	r.add(sectionOr(blue, "Analysis"))
	r.add(riskMatrix(bj))
	// THREE DESCRIPTIVE AREAS, and every line of inquiry lands in exactly one.
	r.add(inquiries(board, "Research areas", accepted))
	r.add(inquiries(board, "Future research directions", deferred))
	r.add(inquiries(board, "Alternatives considered", rejected))
	r.add(sectionOr(blue, "Open questions"))
	// The embed carries ONLY blue content not already composed above — its lifted synthesis
	// surfaces and any tool-owned sections it wrongly authored are dropped (see blueEmbed).
	// If nothing genuinely additional survives, the section is omitted rather than left empty.
	if extra := blueEmbed(blue); extra != "" {
		r.add("## Blue team report (sections not composed above)\n\n" + extra)
	}

	var docket sections
	docket.add(boardSection(board))

	var deb sections
	deb.add(debate(board, evs))

	var jud sections
	jud.add(motions(board))

	var runsec sections
	runsec.add(frictionLog(evs))
	// The record's own invariant check, rendered for the human the report is for. See
	// recordVerification: a section, never a gate.
	runsec.add(recordVerification(board))

	var chg sections
	chg.add(revisionHistory(evs))
	chg.add(withdrawnClaims(evs))
	chg.add(supersededAsks(evs))

	docs := []Doc{
		{File: FileReport, Nav: "Report", Title: "", Body: r.String(),
			Blurb: "the research: the verdict and what it means, the Catechism, the foundations, the analysis, the risks, the open questions"},
		// THE FILE KEEPS ITS NAME AND THE DOCUMENT GETS A TRUE ONE. `docket.md` is a published
		// URL — the shipped READMEs, the skill, the bench's constitution and the site's link
		// rewriter all name it — so renaming the file buys consistency in a string no machine
		// reads. The nav, title and blurb are what a human reads beside the heading, and all
		// three said this was red's when three parties write into it.
		{File: FileDocket, Nav: "Board", Title: "the board",
			Blurb: "the board all three parties wrote: every gap red minted and how each was closed, blue's correctness manifest for the repairs it made, and red's archive spot-checks", Body: docket.String()},
		{File: FileDebate, Nav: "Debate", Title: "the debate",
			Blurb: "the adversarial record round by round — red's audits, blue's answers, the closings, and the bench's terminal disposition", Body: deb.String()},
		{File: FileJudgments, Nav: "Judgments", Title: "judgments",
			Blurb: "every contested question and how it was answered: grade disputes, petitions, and the bench's opinions", Body: jud.String()},
		{File: FileEvidence, Nav: "Evidence", Title: "evidence",
			Blurb: "the computations this run ran, with the exact script, the output, the sha256, and red's independent re-run", Body: ""},
		{File: FileRun, Nav: "Run", Title: "the run",
			Blurb: "how the machinery behaved: the friction the seats hit, the record's own invariant check, and what the run cost", Body: runsec.String()},
		{File: FileChangelog, Nav: "Changelog", Title: "changelog",
			Blurb: "the provenance of this report: every revision, every claim withdrawn, and any post-run repair", Body: chg.String()},
	}

	// THE EVIDENCE LAYER IS RESOLVED ACROSS THE SET, NOT WITHIN A DOCUMENT.
	//
	// A footnote definition cannot cross a file boundary — that is a fact about markdown, not a
	// preference — so weaving globally and splitting afterwards would ship dangling references
	// in six of the seven documents. Each document therefore numbers and defines the citations
	// it actually contains, and the proof layer is numbered ONCE for the whole run (record
	// order) so that a P3 in the debate and a P3 in the report are the same computation.
	sources, err := record.CitedSources(run)
	if err != nil {
		return nil, fmt.Errorf("assemble: cited sources: %w", err)
	}
	proofs, err := record.RecordedProofs(run)
	if err != nil {
		return nil, fmt.Errorf("assemble: recorded proofs: %w", err)
	}

	used := map[string]bool{}
	for i := range docs {
		if docs[i].Body == "" {
			continue
		}
		// Strip finding-markers from the WHOLE composed document, not just blue's lifted
		// content: a finding's location/reason text can carry a "<!--fx:...-->" token into
		// the record-derived findings/transcript sections, and only a final-output strip
		// catches those. No raw marker ships (the leak fix).
		body := StripFindingMarkers(collapseBlanks(docs[i].Body))
		body = collapseBlanks(weaveCitations(body, sources))
		body, hit := weaveProofRefs(body, proofs)
		for _, label := range hit {
			used[label] = true
		}
		docs[i].Body = collapseBlanks(body)
	}
	// evidence.md is composed LAST, from the proofs the set actually anchors, so a proof no
	// document references is still shown (it is on the record) and a reference to a proof that
	// is not on the record is still stated (it is a defect, and a silent drop reads as clean).
	for i := range docs {
		if docs[i].File == FileEvidence {
			docs[i].Body = evidenceDoc(run, proofs, used)
		}
	}

	var out []Doc
	for _, d := range docs {
		if strings.TrimSpace(d.Body) == "" {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

// Title returns the run's short title — the H1 every document in the set opens with.
func Title(run record.Run) string {
	t, _ := heading(blueReport(run))
	return t
}

// blueReport renders the report projection (#709), or "" on any error — no base ingested yet, an
// unreadable record. The callers here compose the whole document set best-effort; a missing report
// yields an empty blue section, never a failed assembly.
func blueReport(run record.Run) string {
	md, err := reportproj.RenderFromRecord(run)
	if err != nil {
		return ""
	}
	return md
}

// sections accumulates the composed blocks of one document, dropping the empties. It is the
// invariant "no heading ships without a body" expressed as the only way to add one.
type sections struct{ parts []string }

func (s *sections) add(block string) {
	if strings.TrimSpace(block) == "" {
		return
	}
	s.parts = append(s.parts, strings.TrimRight(block, "\n"))
}

func (s *sections) String() string { return strings.Join(s.parts, "\n\n") }

// render puts the document's own head on it: the run's short title, this document's name, the
// link bar across the set, and a rule. Every file in the set opens identically, so a reader who
// lands in the middle of one knows what run they are in and how to reach the rest.
func render(d Doc, title string, set []Doc) string {
	h1 := title
	if d.Title != "" {
		// The siblings drop the template's "research report" suffix rather than stack a second
		// noun on it: "<topic> — research report — the docket" names two documents.
		h1 = strings.TrimSuffix(title, " — research report") + " — " + d.Title
	}
	return h1 + "\n\n" + navBar(d.File, set) + "\n\n---\n\n" + strings.TrimRight(d.Body, "\n") + "\n"
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
func indexDoc(run record.Run, title string, set []Doc, board *record.Board, evs []*record.Event) string {
	var b strings.Builder
	b.WriteString(title + "\n\n")
	b.WriteString(navBar("", set) + "\n\n---\n\n")
	b.WriteString(factBox(board, evs) + "\n\n")
	b.WriteString("## The documents\n\n")
	for _, d := range set {
		fmt.Fprintf(&b, "- **[%s](%s)** — %s\n", d.Nav, d.File, d.Blurb)
	}
	if _, err := os.Stat(filepath.Join(run.Dir(), "report.html")); err == nil {
		b.WriteString("\n[**report.html**](report.html) is the same set with real tabs and cross-document links — one self-contained file, no server and no network.\n")
	}
	b.WriteString("\nThe record these were rendered from is `records/`; the computations are cached under `proofs/`. Both survive the run; every document here can be rebuilt from them with `feov-record bench assemble`.\n")
	return b.String()
}

// factBox answers "what is this run and how much do I trust it" in one glance. Every cell is a
// field off the record — nothing here is derived from the prose it sits above.
func factBox(board *record.Board, evs []*record.Event) string {
	open, closed := 0, 0
	rounds := 0
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		if g.Open {
			open++
		} else {
			closed++
		}
		if g.Round > rounds {
			rounds = g.Round
		}
		if g.ClosedRound > rounds {
			rounds = g.ClosedRound
		}
	}
	for _, e := range evs {
		if int(e.GetRound()) > rounds {
			rounds = int(e.GetRound())
		}
	}
	verdict := "_(none recorded)_"
	if o := outcomeOf(evs); o != nil {
		verdict = verdictWord(o)
	}
	var b strings.Builder
	b.WriteString("| | |\n|---|---|\n")
	fmt.Fprintf(&b, "| **Verdict** | %s |\n", verdict)
	fmt.Fprintf(&b, "| **Rounds** | %d |\n", rounds)
	fmt.Fprintf(&b, "| **Gaps** | %d open · %d closed |\n", open, closed)
	return b.String()
}

// supersededAsks is the changelog's account of a bench that spoke more than once. "Read this
// first" carries the terminal statement alone; the ones it replaced are history, and history
// belongs in a document rather than stacked under a heading that reads as parallel asks.
func supersededAsks(evs []*record.Event) string {
	var all []string
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Certify](e); ok {
			if s := strings.TrimSpace(c.GetStatement()); s != "" {
				all = append(all, s)
			}
		}
	}
	if len(all) < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Superseded bench statements\n\n")
	b.WriteString("The bench certified more than once. The LAST statement is the one the report carries; these are the ones it replaced, in the order they were made.\n\n")
	for i, s := range all[:len(all)-1] {
		fmt.Fprintf(&b, "%d. %s\n\n", i+1, s)
	}
	return strings.TrimRight(b.String(), "\n")
}

// Write renders the set to disk and returns report.md's path. Files in the set that this run
// produced no content for are REMOVED rather than left stale: a run re-assembled after its
// motions were withdrawn must not keep yesterday's judgments.md next to today's report.
func Write(run record.Run, title string, docs []Doc, index string) (string, error) {
	want := map[string]bool{FileIndex: true}
	for _, d := range docs {
		want[d.File] = true
		path := filepath.Join(run.Dir(), d.File)
		// The link bar is applied HERE, over the set that actually exists, so it can never
		// link a file nobody wrote.
		if err := os.WriteFile(path, []byte(render(d, title, docs)), 0o644); err != nil {
			return "", fmt.Errorf("assemble: write %s: %w", d.File, err)
		}
	}
	if err := os.WriteFile(filepath.Join(run.Dir(), FileIndex), []byte(index), 0o644); err != nil {
		return "", fmt.Errorf("assemble: write %s: %w", FileIndex, err)
	}
	for _, name := range docOrder {
		if want[name] {
			continue
		}
		if err := os.Remove(filepath.Join(run.Dir(), name)); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("assemble: remove stale %s: %w", name, err)
		}
	}
	return filepath.Join(run.Dir(), FileReport), nil
}
