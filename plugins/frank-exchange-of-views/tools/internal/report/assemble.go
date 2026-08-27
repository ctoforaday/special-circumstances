// Package report assembles the final research report from the RECORD — mechanically, in-tool.
//
// The report is f(event log, blue/report.md). Nothing is authored at assembly; two classes
// of content are combined:
//
//   - BLUE-AUTHORED, RED-AUDITED, LIFTED VERBATIM: the title, the TL;DR, the Catechism, the
//     technical foundations, the analysis, the open questions. These are synthesis surfaces,
//     and a synthesis surface authored at assembly is authored AFTER red's last audit — the
//     run-5 catechism defect (6/7 answers regressed) and, unfixed until now, the TL;DR that
//     nothing ever checked. They live inside blue/report.md, which red re-reads in full every
//     round, so moving them there makes them audited. The assembler copies them; it never
//     writes them. A missing one is FLAGGED, never filled in.
//
//   - TOOL-COMPOSED FROM THE RECORD: the verdict (the terminal `outcome` event), the risk
//     matrix (the board), the three research areas (line-of-inquiry events by fate), the red
//     findings (the board's gaps), and the debate transcript (position/closing/dispute/
//     opinion/petition-rule/halt/certify events). The event log is the source of truth; the
//     rendered projection .md files are in-run artifacts for the seats, NOT read here.
//
// The blue sections are RAW-SLICED, never parsed-and-rendered: round-tripping markdown through
// an AST normalises whitespace and reflows, which is authorship. A fence-aware scan finds the
// true heading boundaries so a "## " line inside a code block is not mistaken for one.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/verify"
)

// Assemble writes <runDir>/report.md and returns its path. It reads the board once (which
// carries both the ordered event log and the replayed gaps) and blue/report.md, and composes
// the report from those two — no --inputs, no intermediate round-trip.
// findingMarker matches the invisible finding-anchor token "<!--fx:f-<id>-->" (slice
// 1b). It is stripped from the FINAL assembled output — not just blue's lifted content
// — because a finding's location/reason text can carry the token into the
// record-derived findings/transcript sections, which a blue-only strip misses.
var findingMarker = regexp.MustCompile(`<!--fx:[^>]*-->`)

// StripFindingMarkers removes every finding-anchor token from report markdown.
func StripFindingMarkers(md string) string { return findingMarker.ReplaceAllString(md, "") }

// citeAnchor matches an invisible citation anchor "<!--cite:c-<id>-->" — the tool-inserted
// immortal marker blue cite splices at a cited sentence. Unlike a finding marker (stripped),
// a citation is RESOLVED at assembly: rewritten to a visible [^N] and listed in the composed
// bibliography.
var citeAnchor = regexp.MustCompile(`<!--cite:(c-[0-9a-f]+)-->`)

// weaveCitations turns the invisible citation layer into a visible one: each "<!--cite:c-…-->"
// anchor becomes a footnote reference [^N] (N in first-appearance order; a label used twice
// shares one N), and a "## Bibliography" of "[^N]: <title>. <url> (accessed <date>)" is
// appended, composed from the cite events. A dangling anchor — one with no source on the
// record (bijection-impossible under the lockdown, but defended) — becomes an explicit
// unresolved-citation line rather than a crash or a silent drop. With no citations the report
// is returned unchanged (no empty bibliography).
func weaveCitations(md string, sources []record.Source) string {
	byLabel := map[string]record.Source{}
	for _, s := range sources {
		byLabel[s.Label] = s
	}
	var order []string
	num := map[string]int{}
	body := citeAnchor.ReplaceAllStringFunc(md, func(tok string) string {
		label := citeAnchor.FindStringSubmatch(tok)[1]
		n, seen := num[label]
		if !seen {
			n = len(order) + 1
			num[label] = n
			order = append(order, label)
		}
		return fmt.Sprintf("[^%d]", n)
	})
	if len(order) == 0 {
		return body
	}
	var b strings.Builder
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n\n## Bibliography\n\n")
	for _, label := range order {
		n := num[label]
		s, ok := byLabel[label]
		if !ok {
			fmt.Fprintf(&b, "[^%d]: _(unresolved citation %s — no source on the record)_\n", n, label)
			continue
		}
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = "_(untitled)_"
		}
		accessed := ""
		if s.AccessDate != "" {
			accessed = fmt.Sprintf(" (accessed %s)", s.AccessDate)
		}
		fmt.Fprintf(&b, "[^%d]: %s. %s%s\n", n, title, s.URL, accessed)
	}
	return b.String()
}

func Assemble(runDir string) (string, error) {
	blue := readOr(filepath.Join(runDir, "blue", "report.md"), "")

	board, err := record.BoardState(runDir)
	if err != nil {
		return "", fmt.Errorf("assemble: board: %w", err)
	}
	bj := record.BoardJSONOf(board)
	evs := board.Events

	var b strings.Builder
	p := func(s string) { b.WriteString(strings.TrimRight(s, "\n")); b.WriteString("\n\n") }

	// Blue-authored head (lifted) + the tool's verdict stamp between the title and the TL;DR,
	// then the reviewer-facing "read this first" composed from the board and the bench's voice.
	p(titleOr(blue))
	p(verdictStamp(outcomeOf(evs)))
	// WHAT ANSWERED, before what was found. A reader deciding how much weight this document
	// carries needs the verdict and the adversary's actual strength in the same breath: a PASS
	// from a tier nobody configured is not the PASS the run was set up to produce.
	if c := conduct(board); c != "" {
		p(c)
	}
	p(orientation(board, evs))
	p(sectionOr(blue, "TL;DR"))
	p(sectionOr(blue, "The Catechism"))
	p(sectionOr(blue, "Technical foundations"))
	p(sectionOr(blue, "Analysis"))

	// Tool-composed from the record.
	p(riskMatrix(bj))
	// THREE DESCRIPTIVE AREAS, and every line of inquiry lands in exactly one. See below.
	p(inquiries(board, "Research areas", accepted))
	p(inquiries(board, "Future research directions", deferred))
	p(inquiries(board, "Alternatives considered", rejected))
	p(sectionOr(blue, "Open questions"))
	// The embed carries ONLY blue content not already composed above — its lifted synthesis
	// surfaces and any tool-owned sections it wrongly authored are dropped (see blueEmbed).
	// If nothing genuinely additional survives, the section is omitted rather than left empty.
	if extra := blueEmbed(blue); extra != "" {
		p("## Blue team report (sections not composed above)\n\n" + extra)
	}
	p(redFindings(board))
	p(debate(board, evs))
	// Every adjudicated exchange, joined on its id (#344).
	if m := motions(board); m != "" {
		p(m)
	}
	if f := frictionLog(evs); f != "" {
		p(f)
	}
	if w := withdrawnClaims(evs); w != "" {
		p(w)
	}
	if r := revisionHistory(evs); r != "" {
		p(r)
	}
	// The record's own invariant check, rendered for the human the report is for. See
	// recordVerification: a section, never a gate.
	p(recordVerification(board))

	// Strip finding-markers from the WHOLE composed report, not just blue's lifted
	// content: a finding's location/reason text can carry a "<!--fx:...-->" token into
	// the record-derived findings/transcript sections, and only a final-output strip
	// catches those. No raw marker ships (the leak fix).
	out := StripFindingMarkers(collapseBlanks(b.String()))

	// Resolve the citation layer: rewrite every "<!--cite:c-…-->" anchor to a visible [^N]
	// and append the composed "## Bibliography" from the cite events. Findings are STRIPPED,
	// citations are RESOLVED — orthogonal passes over the same document, strip first.
	sources, err := record.CitedSources(runDir)
	if err != nil {
		return "", fmt.Errorf("assemble: cited sources: %w", err)
	}
	out = collapseBlanks(weaveCitations(out, sources))

	// Resolve the PROOF layer the same way (#277). Without this pass the anchor shipped RAW
	// into the deliverable and the computation appeared nowhere in it — the evidence existed
	// on the record, in the cache and to the auditor, and was invisible to the reader.
	proofs, err := record.RecordedProofs(runDir)
	if err != nil {
		return "", fmt.Errorf("assemble: recorded proofs: %w", err)
	}
	out = collapseBlanks(weaveProofs(runDir, out, proofs))

	path := filepath.Join(runDir, "report.md")
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return "", fmt.Errorf("assemble: write report.md: %w", err)
	}
	return path, nil
}

func readOr(path, fallback string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return string(b)
}

// blueEmbed returns the parts of blue/report.md NOT already composed elsewhere. Blue's lifted
// synthesis surfaces are dropped (they appear at the top), and the sections blue must never
// author — the risk matrix, red findings, the debate, a verdict, and now the footnotes /
// bibliography — are dropped as fabrication (the tool composes those from the record; blue
// cannot know red's findings or the run's outcome, and citations are tool-managed, woven from
// the cite events at assembly). What survives is genuinely ADDITIONAL blue content. Empty
// output means blue authored only what it should; the caller then omits the embed entirely
// rather than lift-AND-embed the same sections twice.
func blueEmbed(blue string) string {
	drop := map[string]bool{
		// lifted to the top verbatim
		"tl;dr": true, "the catechism": true, "technical foundations": true,
		"analysis": true, "open questions": true,
		// tool-owned — composed from the record, never blue's to author. Footnotes /
		// bibliography are now tool-composed too (woven from the cite events), so a
		// blue-authored one is dropped rather than shipped alongside the composed one.
		"risk matrix": true, "the expansions": true, "expansions": true,
		"alternatives considered": true, "red team findings": true,
		"the debate": true, "blue team report": true, "verdict": true,
		"footnotes": true, "bibliography": true,
		// Composed from the retire events. A blue-authored one would be a SECOND account of
		// what left the report, written by the party that removed it — the record already
		// carries the claim, the reason and the successor, checked at the write.
		"claims withdrawn": true,
		// Composed from register.served_model, and the one section a seat provably cannot
		// author: a seat reads the run's CONFIGURATION and never learns what replied to it.
		// Blue authoring this would ship the unmeasured premise beside the measured one, in
		// the same document, and the reader would have no way to tell which is which — which
		// is nearly what happened (#589), minus the measured half.
		"how this run was conducted": true,
	}
	var out []string
	fence, keep, inPreamble := false, false, true
	for _, ln := range strings.Split(blue, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fence = !fence
			if keep {
				out = append(out, ln)
			}
			continue
		}
		if !fence && strings.HasPrefix(t, "## ") {
			inPreamble = false
			keep = !drop[normalizeHeading(strings.TrimPrefix(t, "## "))]
			if keep {
				out = append(out, ln)
			}
			continue
		}
		if inPreamble {
			// Preamble before the first "## ": drop blue's H1 title (lifted) and any Verdict
			// line (blue cannot author a verdict — #79), keep any genuine prose.
			if strings.HasPrefix(t, "# ") || strings.HasPrefix(strings.ToLower(t), "**verdict:") {
				continue
			}
			out = append(out, ln)
			continue
		}
		if keep {
			out = append(out, ln)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// normalizeHeading folds a heading to a comparison key: lowercase, single-spaced, and with a
// trailing "(in full)" stripped — so "Red Team Findings (in full)" matches "red team findings".
func normalizeHeading(h string) string {
	h = strings.ToLower(strings.Join(strings.Fields(h), " "))
	return strings.TrimSpace(strings.TrimSuffix(h, " (in full)"))
}

// titleOr lifts blue's H1 (the "# <Topic> — research report" line), or flags it missing.
func titleOr(blue string) string {
	for _, ln := range strings.Split(blue, "\n") {
		if t := strings.TrimSpace(ln); strings.HasPrefix(t, "# ") {
			return t
		}
	}
	return "# _(blue/report.md has no title — not authored here)_"
}

// section returns the "## heading" block verbatim (heading included), or "" if absent. It
// tracks fenced code blocks so a "## " line inside ``` or ~~~ is not read as a heading.
func section(md, heading string) string {
	lines := strings.Split(md, "\n")
	start, end := -1, len(lines)
	fence := false
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		if start < 0 {
			// Case-insensitive: blue titled sections "## Technical Foundations" while the
			// template says "## Technical foundations". An exact match declared a present
			// section absent and dropped it — the union invariant lost to a capital letter.
			if strings.EqualFold(t, "## "+heading) {
				start = i
			}
		} else if strings.HasPrefix(t, "## ") {
			end = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	return strings.TrimRight(strings.Join(lines[start:end], "\n"), " \t\n")
}

// sectionOr copies blue's section verbatim, or says in one line that it is missing — NEVER
// authoring a replacement (the union-copy invariant, made structural).
func sectionOr(blue, heading string) string {
	if s := section(blue, heading); s != "" {
		return s
	}
	return fmt.Sprintf("## %s\n\n_(blue/report.md has no \"## %s\" section — not authored here; blue owns this surface and red audits it.)_", heading, heading)
}

// outcomeOf returns the LAST terminal outcome event's body, or nil if none was recorded.
//
// THE TYPE IS THE TEST, NOT THE BODY, and the two differ on exactly one case: an `outcome` event
// carrying no body at all. That must still land as nil — "the bench ran `outcome` and the record
// carries nothing readable" is the absence verdictStamp flags, and matching on the body alone
// would skip such an event and let an EARLIER outcome stand as the terminal one. BodyAs's zero is
// a nil *Outcome, so the assignment says that in one line.
func outcomeOf(evs []*record.Event) *recordpb.Outcome {
	var last *recordpb.Outcome
	for _, e := range evs {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_OUTCOME {
			continue
		}
		last, _ = recordpb.BodyAs[*recordpb.Outcome](e)
	}
	return last
}

// verdictStamp composes the verdict line from the terminal outcome event. A missing outcome
// is flagged, never invented — the same invariant as a missing blue section.
//
// CASING IS PRESENTATION. The record carries the schema's spelling (`verified`, `ceiling`) and
// the stamp wants emphasis, so the UPPER-CASING HAPPENS HERE, at display time. A second
// upper-case vocabulary in the data would be the same fact spelled two ways, which is what this
// migration exists to remove — so the enum is never matched on a shouted word.
func verdictStamp(o *recordpb.Outcome) string {
	if o == nil {
		return "**Verdict:** _(no terminal outcome recorded — `bench outcome` was not run before assembly)_"
	}
	// EVERY branch carries the basis. The first cut appended it only to the default arm, so
	// CEILING and HALTED — which returned early — dropped it; the fuzz failed 35 of 60 runs on
	// exactly that, because a ceiling termination IS derived (rounds against the configured
	// ceiling) and is the most common way a run ends.
	basis := basisNote(o.GetVerdictBasis()) + verdictWhy(o)
	switch o.GetVerdict() {
	case recordpb.RunOutcome_RUN_OUTCOME_CEILING:
		return "**Verdict:** CEILING-TERMINATED — the run hit its round ceiling while still converging. This is NOT a judged failure to verify and must not be read as one: gaps remain open, the final blue revision was never audited by a red pass, and that re-audit debt travels OUT of the run." + basis
	case recordpb.RunOutcome_RUN_OUTCOME_HALTED:
		return "**Verdict:** HALTED — the bench ended this run. The halt opinion is on the record below (Bench disposition) and is relayed to the human verbatim, never smoothed." + basis
	default:
		by := ""
		// `ended` stays a STRING in the schema — it is not one of the enums — so this compare is
		// unchanged, hyphenless words and all.
		switch o.GetEnded() {
		case "deadlock":
			by = " by judged deadlock"
		case "ceiling":
			by = " by safety ceiling"
		}
		// An UNSET verdict spells "" (recordpb.Word maps the zero back to the empty string), so
		// this renders the same blank the absent payload key produced rather than the word
		// `unspecified`, which no seat ever chose.
		return fmt.Sprintf("**Verdict:** %s%s%s", strings.ToUpper(recordpb.Word(o.GetVerdict())), by, basis)
	}
}

// verdictWhy carries the DERIVATION'S OWN REASONING and, on a deadlock, the bench's.
//
// The derivation computed a `why` on every call — "the merge recorded a PASS verdict", "the
// record reaches round 3 against a ceiling of 3" — and used it only to phrase an error, so the
// report could stamp a verdict and never say why it was that one. A judged deadlock is the
// opposite case and had no account at all: it is the ONE terminal verdict the record cannot
// derive (#289), so the bench's --reason is the only evidence it will ever have.
//
// `--reason` LANDS ON `prose`, and recordpb/required.go declares that split in as many words:
// the flag a seat types and the field the record keeps are two vocabularies. Outcome.prose is
// this message's only prose channel, and its own requirement message ("on a judged deadlock it
// is the only evidence that determination will ever have") is this paragraph's other half.
func verdictWhy(o *recordpb.Outcome) string {
	out := ""
	if why := strings.TrimSpace(o.GetVerdictWhy()); why != "" {
		out += " (" + why + ")"
	}
	if r := strings.TrimSpace(o.GetProse()); r != "" {
		out += "\n\n> **The deadlock, in the bench's words:** " + r
	}
	return out
}

// basisNote spells out how a verdict came to be — DERIVED from the record, or ASSERTED by the
// bench and cross-checked against nothing.
//
// The field exists because a seat asked to self-report reports the flattering value, and the
// verdict is the single most consequential such report in the run. It gated the write (`bench
// outcome --as` is refused when it contradicts the record) and then reached the reader as the
// bare word "VERIFIED" — which is exactly the same word an unbacked assertion produces. The
// distinction the field was built to preserve survived every stage except the one that mattered.
func basisNote(basis string) string {
	switch basis {
	case record.VerdictDerived:
		return " — **derived from the record**, not claimed: the events themselves decide this verdict, and `bench outcome` refuses an `--as` that contradicts them."
	case record.VerdictAsserted:
		return " — **asserted by the bench.** The record could not derive this verdict, so it rests on the bench's judgement rather than on the events; read it as an opinion with authority, not as a mechanical result."
	default:
		return ""
	}
}

// orientation is the reviewer-facing "read this first": the open gaps a human should
// re-examine, most severe first, preceded by the bench's terminal ask if it evented one. It
// authors no new judgement — it ORDERS the board (severity, then impact, then likelihood) and
// PROMOTES the bench's already-evented voice (certify/halt), which otherwise sits buried in
// the debate's Bench-disposition line. When the bench never certified/halted (as in a
// ceiling-terminated run), only the ranked gaps show; nothing is invented to fill the space.
func orientation(board *record.Board, evs []*record.Event) string {
	type ranked struct {
		g    *record.Gap
		rank int
	}
	var open []ranked
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil || !g.Open {
			continue
		}
		open = append(open, ranked{g, sevRank(g.Severity)*100 + sevRank(g.Impact)*10 + sevRank(g.Likelihood)})
	}
	sort.SliceStable(open, func(i, j int) bool { return open[i].rank > open[j].rank })

	var b strings.Builder
	b.WriteString("## Read this first\n\n")
	// The bench's terminal ask, if any — promoted from the record, not buried below.
	//
	// ONE FLAG, TWO FIELDS. The seat types `--reason` for both acts and the schema keeps each on
	// the message's own prose channel: a certify's is `statement`, a halt's is `opinion`
	// (recordpb/required.go declares both). Neither is a `reason` field, and inventing one would
	// have been the migration's easiest silent defect.
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Certify](e); ok {
			if s := c.GetStatement(); s != "" {
				b.WriteString("**The bench asks a human to re-examine:** " + s + "\n\n")
			}
			continue
		}
		if h, ok := recordpb.BodyAs[*recordpb.Halt](e); ok {
			if s := h.GetOpinion(); s != "" {
				b.WriteString("**The bench HALTED this run:** " + s + "\n\n")
			}
		}
	}
	if len(open) == 0 {
		b.WriteString("_(no open gaps remain — nothing outstanding to re-examine)_")
		return b.String()
	}
	fmt.Fprintf(&b, "%d open gap(s) remain, most severe first — full statements in **Red team findings** below.\n\n", len(open))
	for i, r := range open {
		fmt.Fprintf(&b, "%d. **[%s]** %s (%s) — %s\n", i+1, gradeWord(r.g.Severity), concise(r.g.Mint.GetProblem()), r.g.ID, concise(r.g.Mint.GetRequiredFix()))
	}
	return strings.TrimRight(b.String(), "\n")
}

// sevRank orders a gap by one grade, using the canonical MASS weight (record.MASS) scaled to an
// int — so all eight real domain grades sort correctly (realized 0 · trivial 1 · low 2 ·
// low-medium 3 · medium 4 · medium-high 5 · high 6 · certain 7), consistent with how the rest of
// the system weights grades. The earlier critical|high|medium|low table matched NONE of the
// domain grades past high/medium/low and sank certain/realized/medium-high/low-medium/trivial to
// 0 — the most severe open gaps sorted below the least. An absent grade is the UNSPECIFIED zero,
// which GradeStr spells "" and MASS has no key for, so it ranks 0 exactly as an absent payload key
// did.
//
// IT TAKES THE TYPED GRADE, and that is not cosmetic. The old signature was `any`, and a
// `recordpb.Grade` passed to it would still COMPILE — the type assertion inside `grade` would
// simply fail, every gap would rank 0, and the ordering this function exists to produce would be
// silently arbitrary with nothing failing. The lower-casing is gone with it: GradeStr returns the
// schema's own spelling, which is what MASS is keyed on (`low_medium`, one separator).
func sevRank(g recordpb.Grade) int {
	return int(record.MASS[record.GradeStr(g)] * 2)
}

// gradeWord renders a typed grade for the document — the schema's word, or an em dash when the
// gap carries none. It is `grade`'s counterpart for the replayed board, whose grades are
// `recordpb.Grade` values rather than the JSON view's `any`.
func gradeWord(g recordpb.Grade) string {
	if w := record.GradeStr(g); w != "" {
		return w
	}
	return "—"
}

func riskMatrix(bj record.BoardJSON) string {
	var b strings.Builder
	b.WriteString("## Risk matrix\n\n")
	b.WriteString("| Risk | Likelihood | Impact | Complexity to mitigate | Mitigation / disposition |\n")
	b.WriteString("|---|---|---|---|---|\n")
	if len(bj.Open) == 0 {
		b.WriteString("| _(no open gaps)_ |  |  |  |  |\n")
	}
	for _, g := range bj.Open {
		risk := g.Problem
		if strings.TrimSpace(risk) == "" {
			risk = g.ID
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			cell(concise(risk)), grade(g.Likelihood), grade(g.Impact), grade(g.ComplexityCost), cell(concise(g.RequiredFix))))
	}
	b.WriteString("\nThe matrix is a scan surface: each row's full problem statement, required fix and acceptance check are in **Red team findings** below.")
	return b.String()
}

// concise returns a scannable one-line version of a long field — its first sentence, or a
// hard-truncated head. A risk matrix is scanned, not read; the tool cannot summarize a
// 60-word gap description, so it takes the lead sentence and leaves the full text in Red
// team findings, where nothing is lost.
func concise(s string) string {
	s = strings.Join(strings.Fields(s), " ") // collapse internal whitespace/newlines
	if i := strings.IndexAny(s, ".!?"); i > 0 && i <= 100 {
		return strings.TrimSpace(s[:i+1])
	}
	if len(s) > 100 {
		return strings.TrimSpace(s[:100]) + "…"
	}
	return s
}

// THE LIFECYCLE OF A RESEARCH TOPIC, in the three descriptive areas a reader gets:
//
//	Research areas               `pursued` — followed, and what it yielded. AND `proposed`: a line
//	                             put forward and not yet resolved is research this run has taken
//	                             on, and its row carries the undecided state so the heading does
//	                             not have to lie about it.
//	Future research directions   `deferred` — "worth taking, and not by THIS run". KEPT, not
//	                             rejected. --reason says what a later run should pick it up FOR,
//	                             and it reaches the report as a proposal a human selects.
//	Alternatives considered      `declined` (weighed, not taken) and `abandoned` (tried, died).
//
// # It rendered TWO, and the missing ones made the report say something false
//
// Taking `rejected` as the complement of `pursued` puts `deferred` AND `proposed` under
// "Alternatives considered" — wrong in a way no reader can detect, since only the `[deferred]`
// tag on the row contradicts the heading above it. Narrowing the pair instead makes them vanish
// from the report entirely, which is the failure the status enum's own Why warns about ("it
// silently vanishes from the section that exists to show the roads not taken").
//
// # `proposed` is a research area, and both alternatives were worse
//
// A first cut excluded `proposed` from every section, arguing a heading announces a fate and an
// undecided line has none. TestFuzzDebate failed six seeds with `avenue prose absent from report`,
// and that invariant wins: a seat's recorded reasoning must reach the reader. A second cut gave it
// a fourth section. Three is the decision — a line blue put forward IS an area this run is
// researching, and red's per-round support verdict is what stops it sitting there undecided,
// rather than a heading that describes the omission.
//
// `rejected` stays a COMPLEMENT so a sixth status cannot silently match nothing.
func accepted(status string) bool { return status == "pursued" || status == "proposed" }
func deferred(status string) bool { return status == "deferred" }

func rejected(status string) bool { return !accepted(status) && !deferred(status) }

// inquiries renders the line of inquiry LIFECYCLE under the given heading — replayed state, one row per
// line of inquiry, not one row per event. Reading raw events double-listed every line of inquiry that MOVED: a
// line pursued at r0 and abandoned at r2 rendered under both headings at once, as an expansion
// and as an alternative to itself.
//
// Each row now carries what the reader needed to judge the choice and could not see: the
// history that produced the status, RED'S RULING on the direction, and — when blue moved a line
// red ruled out-of-scope or too-thin — the fact that it did so against that ruling. Red's ruling
// is an argument, not a command, so blue may pursue anyway; the disagreement is the substance,
// and until now the report showed the line with no trace that anyone had contested it.
func inquiries(board *record.Board, heading string, want func(string) bool) string {
	var rows []string
	for _, a := range record.Inquiries(board) {
		if !want(a.Status) {
			continue
		}
		method := ""
		if a.Method != "" {
			method = fmt.Sprintf(" _(%s)_", a.Method)
		}
		reason := ""
		if a.Reason != "" {
			reason = " — " + a.Reason
		}
		status := ""
		// THE ROW CARRIES THE STATUS WHENEVER THE HEADING DOES NOT SETTLE IT. `pursued` is what
		// "Research areas" means, so it is left off; `proposed` shares that heading and must say
		// so, or the reader cannot tell a line this run FOLLOWED from one it merely named.
		if a.Status != "pursued" {
			status = fmt.Sprintf(" [%s]", a.Status) // abandoned vs declined vs deferred is the shape of the counter
		}
		row := fmt.Sprintf("- **%s**%s%s%s (%s)", a.Line, method, status, reason, a.SeatID)
		if len(a.History) > 1 {
			row += fmt.Sprintf("\n  - history: %s", strings.Join(a.History, " → "))
		}
		// THERE IS NO PER-LINE SUPPORT ROW, AND THE ABSENCE IS A RULING RATHER THAN A DROPPED
		// FEATURE. This rendered red's per-round `supported`/`weakened`/`unsupported`/`absent`
		// verdict on each line — a vocabulary that made PRESENCE the question. Presence is not a
		// question: the lines reach the report on the WORKLIST generated from this projection, so
		// blue cannot cut them, and what remains (did blue's body deliver the research) is an
		// ORDINARY GAP that renders under Red team findings with an id, a grade and a PASS gate.
		// record.Inquiry's own header states the same decision from the projection's side.
		//
		// What replaced it is ONE per-round `InquiryReview`, read by record.InquiryReviewDue.
		// Whether the report should carry a line saying that read happened — and where — is a
		// composition decision, not a conversion, so nothing is invented here.
		if a.Ruling != "" {
			ruled := fmt.Sprintf("\n  - red ruled **%s** (r%d)", a.Ruling, a.RuledRound)
			if a.RulingWhy != "" {
				ruled += " — " + a.RulingWhy
			}
			row += ruled
		}
		if a.Contests != "" {
			row += fmt.Sprintf("\n  - **blue took this line against red's `%s` ruling.** A ruling is an argument, not a command; the disagreement stands on the record.", a.Contests)
		}
		rows = append(rows, row)
	}
	body := "_(none on the record)_"
	if len(rows) > 0 {
		body = strings.Join(rows, "\n")
	}
	return "## " + heading + "\n\n" + body
}

// withdrawnClaims renders the claims blue REMOVED from the report, from the retire events.
//
// Substance leaves the report only through `blue retire`, which names the claim as it stood,
// why it went, and what replaced it — and the reader of the finished report saw none of it. A
// claim that was argued, weighed and then withdrawn is part of what the debate decided; dropping
// it makes the report indistinguishable from one where the claim was never made.
func withdrawnClaims(evs []*record.Event) string {
	var rows []string
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.Retire](e)
		if !ok {
			continue
		}
		claim := strings.TrimSpace(r.GetClaim())
		if claim == "" {
			continue
		}
		row := fmt.Sprintf("- **%s** — %s (%s, r%d)", concise(claim), r.GetReason(), e.GetSeatId(), e.GetRound())
		if s := r.GetSupersededBy(); s != "" {
			row += "\n  - superseded by: " + s
		}
		// A PHANTOM RETIREMENT IS WORSE THAN USELESS, and only the basis distinguishes one.
		// The scorecard's additive-integrity detector computes unrecorded_claim_loss as the
		// drop in claim_count MINUS the retire events, so a retirement of a claim that was
		// never in the report subtracts from the accounted side and CANCELS real loss —
		// blinding the one detector built to catch silent deletion. The field records whether
		// the record can actually show the claim leaving; the reader could not see which.
		switch r.GetRemovalBasis() {
		case record.RemovalVerified:
			row += "\n  - basis: **verified** — the claim appears in the old span of a recorded edit, so the record shows it leaving."
		case record.RemovalAsserted:
			row += "\n  - basis: **asserted** — the claim is absent now, but nothing on the record shows it was ever present. Honest for a round-0 claim written and rewritten in one sitting; indistinguishable, here, from a retirement of something that never existed."
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return ""
	}
	return "## Claims withdrawn\n\n_Substance leaves this report only through the `retire` verb, which records the claim as it stood and why it went. These were argued and then removed; the reasoning is part of what the debate decided._\n\n" + strings.Join(rows, "\n")
}

// provenance renders which lens findings surfaced a gap, IN THE FINDING'S OWN WORDS.
//
// This was `surfaced by: L5-F1, L6-F2` — labels and nothing else. Nothing in the report defines
// those labels: unmintedFindings renders a finding's text only when NO gap claims it, so the
// instant the merge acted on a finding, the leaf-level evidence that produced it left the
// document and the citation dangled.
//
// It is the wrong half to drop. A gap's `problem` is the merge's RESTATEMENT; the finding is what
// red actually observed at the leaf, and the two sitting together is what lets a reader see a
// restatement drift from its evidence. Unresolvable labels are kept as bare labels rather than
// silently dropped — an unresolvable citation is itself worth seeing.
func provenance(m *recordpb.Mint, findings map[string]*recordpb.Finding) string {
	fb := m.GetFoundBy()
	if len(fb) == 0 {
		return ""
	}
	var lines []string
	for _, lbl := range fb {
		f := findings[lbl]
		if f == nil {
			lines = append(lines, "- "+lbl+": (no finding with this label is on the record)")
			continue
		}
		loc := f.GetLocation()
		if loc != "" {
			loc = " (" + loc + ")"
		}
		lines = append(lines, fmt.Sprintf("- %s%s: %s", lbl, loc, f.GetText()))
	}
	return "\nsurfaced by:\n" + strings.Join(lines, "\n")
}

// redFindings composes the full findings from the board's gaps: every open gap with its
// grades and required fix, then the closure index. This is the ledger's content, drawn from
// the replayed board rather than read back from the projection file.
func redFindings(board *record.Board) string {
	// Label -> the finding it names, so a gap can quote the evidence it was minted from.
	findings := map[string]*recordpb.Finding{}
	for _, e := range board.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && f.GetLabel() != "" {
			findings[f.GetLabel()] = f
		}
	}
	var open, closed []string
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		if g.Open {
			regraded := regradeHistory(g)
			// Provenance: which lens findings surfaced this gap, IN THE FINDING'S OWN WORDS.
			//
			// This was `surfaced by: L2-F1, L5-F2` — labels and nothing else. Nothing in the
			// report defines those labels: unmintedFindings renders a finding's text only when
			// NO gap claims it, so the moment the merge acts on a finding, the leaf-level
			// evidence that produced it leaves the document and the citation dangles. The fuzz
			// found runs where EVERY finding was minted and red's words appeared nowhere at all.
			//
			// It is the wrong half to drop. A gap's `problem` is the merge's RESTATEMENT; the
			// finding is what red actually observed at the leaf, and the two sitting together is
			// what lets a reader see a restatement drift from its evidence. Unresolvable labels
			// (a found_by naming no finding on the record) are kept as bare labels rather than
			// silently dropped — an unresolvable citation is itself worth seeing.
			foundBy := provenance(g.Mint, findings)
			// `class` is a registry SLUG, not a grade — it goes through `grade` only for that
			// helper's em-dash-when-empty arm, exactly as it did before.
			open = append(open, fmt.Sprintf("### %s — %s\n%s\nseverity %s | %s x %s | cx %s | class %s%s\nrequired_fix: %s%s\nacceptance_check: %s%s",
				g.ID, g.Mint.GetProblem(),
				g.Mint.GetLocation(),
				gradeWord(g.Severity), gradeWord(g.Likelihood), gradeWord(g.Impact), gradeWord(g.ComplexityCost), grade(g.Mint.GetClass()),
				regraded,
				g.Mint.GetRequiredFix(),
				fixProposal(g.Mint),
				g.Mint.GetAcceptanceCheck(),
				foundBy))
		} else {
			// THE `close` EVENT'S CLASS, and the fallback covers the gap a BENCH opinion closed.
			//
			// `g.Closure` is now the Close body ONLY: replay.go splits a bench closure onto
			// BenchClosure, and its own comment warns that `g.Closure != nil` no longer means
			// "closed by anything". A nil Closure spells the empty word here and falls through to
			// "closed", which is what the old single-payload read produced for a bench closure
			// (an Opinion payload has no `closure_class` key). Behaviour held.
			// THE LAST CLOSER'S WORD, and no flattering default. This read g.Closure alone and
			// defaulted "" to "repaired" — and g.Closure is nil for every BENCH closure, so a gap
			// the bench ruled `defect_accepted` (a live defect, shipping in this very report)
			// reached the one human-facing document as "repaired". That inverts the artifact
			// axis ArtifactStateOf exists to protect, in the terminal artifact. A class-less
			// closure now says the neutral thing instead of the healthiest possible word.
			cc := g.ClosureReason()
			if cc == "" {
				cc = "closed (no recorded class)"
			}
			succ := g.Closure.GetSuccessor()
			if succ == "" {
				succ = "-"
			}
			// THE EVIDENCE TRAVELS WITH THE CLOSED ENTRY TOO. This was scoped out on the argument
			// that a closed gap's open question is how it was settled — but the finding's text
			// then reached NO reader at all, because unmintedFindings skips anything a gap
			// claimed. A run where every finding was minted and every gap closed printed red's
			// leaf-level words nowhere. The closure answers what happened; it does not restate
			// what was observed, and an audit of a closure needs both.
			closed = append(closed, fmt.Sprintf("- %s | %s | %s | successor %s%s%s", g.ID, cc, g.Mint.GetProblem(), succ, regradeHistory(g), provenance(g.Mint, findings)))
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## Red team findings (in full)\n\n### Open gaps (%d)\n\n", len(open))
	if len(open) == 0 {
		b.WriteString("_(none open)_\n")
	} else {
		b.WriteString(strings.Join(open, "\n\n"))
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\n### Closure index (%d)\n\n", len(closed))
	if len(closed) == 0 {
		b.WriteString("_(nothing closed)_")
	} else {
		b.WriteString(strings.Join(closed, "\n"))
	}

	// Lens findings the merge did NOT raise to a gap — red's leaf audit that fell short of a
	// mint but still carries substance (a claimed failure mode shown inapplicable, a resilient
	// result confirmed). A finding is "minted" when its label appears in some gap's found_by
	// credit chain; the rest are dropped on the floor by every report before this one. Surfaced
	// here, subordinate to the gaps, so red's voice is not silently lost (#77).
	if un := unmintedFindings(board); un != "" {
		fmt.Fprintf(&b, "\n\n%s", un)
	}
	if sc := archiveSpotChecks(board); sc != "" {
		fmt.Fprintf(&b, "\n\n%s", sc)
	}
	if cm := correctnessManifest(board); cm != "" {
		fmt.Fprintf(&b, "\n\n%s", cm)
	}
	return b.String()
}

// correctnessManifest renders blue's self-audit receipts — one row per repaired gap, saying what
// blue checked and what checking it showed.
//
// "An unmanifested repair is unchecked by blue's OWN standard, which is a stronger thing to be
// able to say than 'we think it was checked'." That is the verb's own justification, and until
// now nobody could say it: the manifest was scored from a transient envelope and the receipt on
// the record reached no reader. A repair claimed as done and a repair audited by the party that
// made it are different things, and only this section shows which one a closure was.
//
// A repaired gap with NO row is named rather than omitted. An absent receipt is the finding.
func correctnessManifest(board *record.Board) string {
	type row struct {
		gapID, text, seat string
		round             int
	}
	var rows []row
	manifested := map[string]bool{}
	for _, e := range board.Events {
		mr, ok := recordpb.BodyAs[*recordpb.ManifestRow](e)
		if !ok {
			continue
		}
		text := strings.TrimSpace(mr.GetRow())
		if text == "" {
			continue
		}
		id := mr.GetGapId()
		manifested[id] = true
		rows = append(rows, row{id, text, e.GetSeatId(), int(e.GetRound())})
	}
	// Every gap blue actually repaired: one the record shows CLOSED. A closure with no manifest
	// row is a repair nobody audited, including its author.
	var unmanifested []string
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g != nil && g.HasClosed && !manifested[id] {
			unmanifested = append(unmanifested, id)
		}
	}
	if len(rows) == 0 && len(unmanifested) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "### Blue's correctness manifest (%d)\n\n", len(rows))
	b.WriteString("Blue's self-audit of its own repairs — what it checked for each gap and what checking it showed. An unmanifested repair is unchecked by blue's own standard, which is a stronger thing to be able to say than \"we think it was checked\".\n\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "- **%s** (%s, r%d): %s\n", r.gapID, r.seat, r.round, r.text)
	}
	if len(unmanifested) > 0 {
		fmt.Fprintf(&b, "\n**%d closed gap(s) carry no manifest row (%s).** Those repairs were not audited by the party that made them.\n",
			len(unmanifested), strings.Join(unmanifested, ", "))
	}
	return strings.TrimRight(b.String(), "\n")
}

// archiveSpotChecks renders red's re-verification of its own closure archive — which closures it
// re-read each round, what that found, and any round that owed a sample and did not take one.
//
// The closure index is the report's claim that a gap was dealt with, and it is only as good as
// the last time anyone looked. Red looked, recorded which closures it re-read, and the reader was
// never told — the receipt sat on the record with no consumer at all. The DEBT is rendered beside
// the discharges rather than left to the exit code, because a reader deciding how much to trust
// the closure index needs to know which rounds checked it and which did not.
func archiveSpotChecks(board *record.Board) string {
	checks, debt, falseEmpty := record.SpotCheckAudit(board)
	if len(checks) == 0 && len(debt) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "### Archive spot-checks (%d)\n\n", len(checks))
	b.WriteString("Red re-reading its own closure record. A closure index is only as good as the last time anyone looked; these are the rounds that looked.\n\n")
	for _, sc := range checks {
		b.WriteString("- " + sc.Describe() + "\n")
	}
	for _, sc := range falseEmpty {
		fmt.Fprintf(&b, "- **r%d claimed there was nothing to sample, and the board shows %d archived closure(s) at that round's start.** The claim does not survive the record.\n", sc.Round, sc.Archived)
	}
	if len(debt) > 0 {
		fmt.Fprintf(&b, "\n**%d round(s) entered with a non-empty archive and sampled none of it", len(debt))
		var rs []string
		for _, r := range debt {
			rs = append(rs, fmt.Sprintf("r%d", r))
		}
		fmt.Fprintf(&b, " (%s).** Those closures went un-re-examined; weigh the closure index accordingly.\n", strings.Join(rs, ", "))
	}
	return strings.TrimRight(b.String(), "\n")
}

// fixProposal renders how a gap's required_fix was arrived at, and — where red stated one — the
// concrete span-and-replacement that goes with it.
//
// `fix_basis` is DERIVED at the mint, never claimed: it reads `verified` only when red supplied
// both the span and its replacement and the tool validated that span against the live report. That
// validation is a forced re-read, and it exists because all three of one smoke's round-2 gaps
// were contradictions between blue's new text and text red had never re-read before prescribing.
//
// The field gated the estoppel machinery and reached the reader as nothing at all, so a demand
// red had checked against the document read identically to one written from memory of what the
// document probably said. That is the whole difference the axis was built to record.
func fixProposal(mint *recordpb.Mint) string {
	if mint == nil {
		return ""
	}
	// `fix_basis` IS A STRING IN THE SCHEMA, not an enum, so the two literals below are the
	// same compare they always were rather than a vocabulary this file could get wrong.
	switch mint.GetFixBasis() {
	case "verified":
		s := "\nfix_basis: **verified** — red stated an exact replacement and the tool checked the span against the live report, so this demand was written with the text in front of it."
		if old, nw := mint.GetLocation(), mint.GetFixNew(); old != "" && nw != "" {
			s += fmt.Sprintf("\n  - replace: %q\n  - with: %q", old, nw)
		}
		return s
	case "proposed":
		return "\nfix_basis: **proposed** — prose only. Red did not state an exact replacement, so nothing checked this demand against the report's current text."
	default:
		return ""
	}
}

// regradeHistory renders EVERY grade movement on a gap with the reason given for it, or "" if
// the gap was never regraded.
//
// EVERY movement, on open and closed gaps alike. Rendering a count and the latest basis loses a
// stated reason twice over: an earlier regrade's basis is overwritten in the display by a later
// one, and a gap that closes drops its whole history — so a grade argued down over three rounds
// and then closed shows the reader no argument at all. A regrade is red revising its own
// assessment, usually because blue disputed it; the dispute renders, and the reasoning that
// answered it must too.
func regradeHistory(g *record.Gap) string {
	if len(g.Regrades) == 0 {
		return ""
	}
	var rows []string
	for _, r := range g.Regrades {
		// THE AXIS NAMES STAY, and they are already the schema's own spelling — one separator,
		// underscores, so `complexity_cost` needs no translation. Iterating the four typed
		// getters replaces a loop over four payload KEYS; the axis label and the field it reads
		// can no longer disagree.
		var axes []string
		for _, axis := range []struct {
			name  string
			grade recordpb.Grade
		}{
			{"severity", r.GetSeverity()},
			{"likelihood", r.GetLikelihood()},
			{"impact", r.GetImpact()},
			{"complexity_cost", r.GetComplexityCost()},
		} {
			// A MOVEMENT IS A GRADE WITH A WORD. The old test was `Str(axis) != ""` — an axis the
			// regrade did not move had no key. GradeStr spells the UNSPECIFIED zero "" for the
			// same reason, so an untouched axis is skipped exactly as before, and an axis
			// explicitly set to UNSPECIFIED does not render `severity → ` with nothing after it.
			if v := record.GradeStr(axis.grade); v != "" {
				axes = append(axes, axis.name+" → "+v)
			}
		}
		moved := strings.Join(axes, ", ")
		if moved == "" {
			moved = "regraded"
		}
		// `--reason` LANDS ON `basis` (recordpb/required.go: "grade movement is recorded with
		// its reason"). Regrade has no `reason` field, and inventing one would have been a
		// silent blank in the one place this section exists to show.
		rows = append(rows, fmt.Sprintf("\n  - %s — %s", moved, r.GetBasis()))
	}
	return fmt.Sprintf(" · regraded x%d%s", len(g.Regrades), strings.Join(rows, ""))
}

// unmintedFindings renders the lens findings whose label is credited by NO gap's found_by, or
// "" if every finding earned a gap. Ordered by the event log so the section is deterministic.
func unmintedFindings(board *record.Board) string {
	minted := map[string]bool{}
	for _, g := range board.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		for _, lbl := range g.Mint.GetFoundBy() {
			minted[lbl] = true
		}
	}
	var rows []string
	for _, e := range board.Events {
		f, ok := recordpb.BodyAs[*recordpb.Finding](e)
		if !ok {
			continue
		}
		lbl := f.GetLabel()
		if lbl != "" && minted[lbl] {
			continue
		}
		head := lbl
		if head == "" {
			head = f.GetFindingId()
		}
		loc := f.GetLocation()
		if loc != "" {
			loc = " — " + loc
		}
		// A finding is addressed by COALESCENCE and nothing else: its label named in some gap's
		// found_by. So a finding reaching this section means exactly one thing — the merge
		// weighed it and did not mint it — and the section says so.
		//
		// `--reason` LANDS ON `text`: Finding's only prose channel, the same flag/field split
		// recordpb/required.go declares for Verify.text and Close.prose.
		rows = append(rows, fmt.Sprintf("### %s%s\nseverity %s | %s x %s | %s\n%s",
			head, loc,
			gradeWord(f.GetSeverity()), gradeWord(f.GetLikelihood()), gradeWord(f.GetImpact()),
			e.GetSeatId(),
			f.GetText()))
	}
	if len(rows) == 0 {
		return ""
	}
	return fmt.Sprintf("### Lens findings not raised to a gap (%d)\n\nRed's leaf audit that the merge weighed but did not mint — kept for the record, not a gate on the verdict.\n\n%s",
		len(rows), strings.Join(rows, "\n\n"))
}

// debate composes the one transcript from the event log: per round, the parties' positions
// and closings, the grade disputes and their answers, then the bench's opinions and petition
// rulings; then the terminal bench disposition (halt / certify). Everything the parties and
// the bench put on the record, in one place — the seat re-narrated none of it.
// debate takes the BOARD as well as the events, because a petition's ruling cannot be attributed
// to its filing from an event alone: motion-rule carries motion_id, never the filer or subject of
// the ask. record.Motions performs that join.
func debate(board *record.Board, evs []*record.Event) string {
	var order []int
	byRound := map[int][]*record.Event{}
	for _, e := range evs {
		r := int(e.GetRound())
		if _, seen := byRound[r]; !seen {
			order = append(order, r)
		}
		byRound[r] = append(byRound[r], e)
	}

	var parts []string
	for _, r := range order {
		re := byRound[r]
		var round []string
		// THE BODY IS THE TYPE, and the party is still the seat's. `--reason` lands on
		// Position.text and Closing.text — one prose channel each, declared for Closing in
		// recordpb/required.go ("the closing argument for this gap").
		for _, e := range re {
			party := record.PartyOf(e)
			if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
				switch party {
				case "merge":
					round = append(round, "### RED\n"+p.GetText())
				case "blue":
					round = append(round, "### BLUE\n"+p.GetText())
				}
				continue
			}
			if c, ok := recordpb.BodyAs[*recordpb.Closing](e); ok {
				switch party {
				case "merge":
					round = append(round, fmt.Sprintf("### RED CLOSING — %s\n%s", c.GetGapId(), c.GetText()))
				case "blue":
					round = append(round, fmt.Sprintf("### BLUE CLOSING — %s\n%s", c.GetGapId(), c.GetText()))
				}
			}
		}

		// PETITIONS: the filing AND the ruling, in event order, in one block.
		//
		// Rendering the ruling alone — "petition red-merge: granted — <opinion>" — gives the
		// reader the bench's answer with no question attached. A petition is the one channel a
		// seat has for an ethical, safety, integrity or constitutional objection; the relief it
		// sought and the basis it argued are the substance, and the ruling is only meaningful
		// against them.
		//
		// The two are NOT joined into a single row. `petition-rule` carries the petitioner and
		// the class but no petition id (#312), so pairing two filings by the same seat in one
		// round would be a guess. They are rendered in the order they happened, which is a fact,
		// and the run-level count check below says plainly if a filing went unanswered.
		// The bench's in-round acts: opinions on the docket.
		var lead []string
		for _, e := range re {
			o, ok := recordpb.BodyAs[*recordpb.Opinion](e)
			if !ok {
				continue
			}
			// `--reason` LANDS ON `rationale` — Opinion's prose channel, declared in
			// recordpb/required.go ("a disposition with no stated reasoning is
			// indistinguishable from a default"). `disposition` stays a STRING in the schema,
			// so it renders as written.
			lead = append(lead, fmt.Sprintf("- %s: %s — principle: %s; tension: %s; review: %s\n%s",
				// The vocabulary's word, not the generated constant name — see view.go's
				// opinion renderer, where the same substitution was needed for the same reason.
				o.GetGapId(), recordpb.Word(o.GetDisposition()), o.GetPrinciple(),
				o.GetTension(), o.GetReviewFlag(), o.GetRationale()))
		}
		if len(lead) > 0 {
			round = append(round, "### LEAD\n"+strings.Join(lead, "\n"))
		}
		if len(round) > 0 {
			parts = append(parts, fmt.Sprintf("### Round %d\n\n%s", r, strings.Join(round, "\n\n")))
		}
	}

	// Terminal bench disposition: halt and certify are run-level, not a round's.
	var disp []string
	filed, ruled := 0, 0
	//
	// THREE ACTS, THREE PROSE FIELDS, ONE FLAG. The bench types `--reason` for a halt and a
	// certify and the schema keeps each on its own channel — `Halt.opinion` (relayed verbatim, so
	// a halt with no written opinion cannot do its job) and `Certify.statement` (the bench's only
	// continuity between runs). A `Halt.reason` does not exist and must not be invented.
	for _, e := range evs {
		if h, ok := recordpb.BodyAs[*recordpb.Halt](e); ok {
			disp = append(disp, "**HALT** — "+h.GetOpinion())
			continue
		}
		if c, ok := recordpb.BodyAs[*recordpb.Certify](e); ok {
			disp = append(disp, "**Certification** — "+c.GetStatement())
			continue
		}
		// A DECLARATION BINDS HOW THE RECORD IS READ, so it belongs in the artifact the
		// record produces, not only in the transcript view. It moves no gap and names none —
		// which is why it could never have appeared in the per-gap dispositions above, and
		// why the bench that needed one had nowhere to put it (#361).
		if d, ok := recordpb.BodyAs[*recordpb.Declare](e); ok {
			disp = append(disp, "**Declared** — "+d.GetHolding())
		}
	}
	// A PETITION IS A MOTION NOW, AND THAT MAKES THIS COUNT BETTER THAN IT WAS. It read the
	// retired `petition`/`petition-rule` types, so after the collapse it saw zero of each and the
	// unanswered-petition warning below could never fire — silence that read as "no petitions
	// went unanswered".
	//
	// The old note said the pair was COUNTED rather than JOINED because `petition-rule` carried
	// no id, so pairing two filings by one seat in one round would have been a guess. A motion
	// has an id; record.Motions joins the ask to its answer, so this is now an exact count of
	// petitions that were never ruled rather than a difference between two tallies.
	for _, m := range record.Motions(board) {
		if m.Subject != "petition" {
			continue
		}
		filed++
		if m.Ruled() {
			ruled++
		}
	}
	// AN UNANSWERED PETITION IS THE LOUD CASE. A petition is a seat's channel for an ethical,
	// safety, integrity or constitutional objection, and the engine routes it to a bench sitting
	// BEFORE the debate continues — so a filing with no ruling means that sitting did not happen.
	// Reporting nothing would make the failure indistinguishable from a run that had no
	// petitions at all, which is why this counts rather than staying silent.
	if filed > ruled {
		disp = append(disp, fmt.Sprintf("**%d petition(s) received no ruling on the record.** A petition is heard before the debate continues; a filing with no ruling means that sitting is missing, not that the objection was withdrawn.", filed-ruled))
	}

	var b strings.Builder
	b.WriteString("## The debate\n\n")
	if len(parts) == 0 {
		b.WriteString("_(no debate on the record)_")
	} else {
		b.WriteString(strings.Join(parts, "\n\n"))
	}
	if len(disp) > 0 {
		b.WriteString("\n\n### Bench disposition\n\n" + strings.Join(disp, "\n\n"))
	}
	return b.String()
}

// frictionLog surfaces the tooling gaps the seats hit — friction events, recorded through the
// friction verb but rendered by nothing before this (write-only, per the 2026-07-23 audit). A
// missing capability the run hit is a finding about the tooling; surfacing it is how it reaches
// the human who can retool the seat, instead of dying on an unread channel.
func frictionLog(evs []*record.Event) string {
	var rows, attested []string
	// TWO MESSAGES, ONE FIELD NAME. Both carry the seat's words on `text`; the empty-text skip
	// that used to sit above the type test now sits inside each arm, which changes nothing
	// except that the field is read from the message that actually has it.
	for _, e := range evs {
		if f, ok := recordpb.BodyAs[*recordpb.Friction](e); ok {
			if t := strings.TrimSpace(f.GetText()); t != "" {
				rows = append(rows, fmt.Sprintf("- **%s**: %s", e.GetSeatId(), t))
			}
			continue
		}
		if f, ok := recordpb.BodyAs[*recordpb.FrictionNone](e); ok {
			if t := strings.TrimSpace(f.GetText()); t != "" {
				attested = append(attested, fmt.Sprintf("- **%s**: %s", e.GetSeatId(), t))
			}
		}
	}
	if len(rows) == 0 && len(attested) == 0 {
		return ""
	}
	out := "## Friction (tooling gaps the run hit)\n\n"
	if len(rows) > 0 {
		out += strings.Join(rows, "\n") + "\n"
	} else {
		out += "No capability gap was reported.\n"
	}
	// THE ATTESTATIONS BELONG BESIDE THE COMPLAINTS, and that is the whole point of recording
	// them. "No friction this run" is worth nothing alone: it reads identically whether the seats
	// looked and found none or never used the channel, and across eighteen recorded sittings it
	// was the second every time. A named seat saying what it reached for turns the zero into a
	// statement someone can be wrong about.
	if len(attested) > 0 {
		out += "\n### Seats that reported nothing blocked them\n\n" + strings.Join(attested, "\n") + "\n"
	} else if len(rows) > 0 {
		out += "\nNo seat closed the channel explicitly, so the list above is what was volunteered rather than what was met.\n"
	}
	return strings.TrimRight(out, "\n")
}

// revisionHistory is blue's per-round revision record folded into the report as
// bottom-of-document provenance — how the report evolved round by round. Composed from revision
// events; a run with no revisions omits it.
func revisionHistory(evs []*record.Event) string {
	var rows []string
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.Revision](e)
		if !ok {
			continue
		}
		// Revision's one field is `text`; `--reason` is the flag, as everywhere else.
		if t := strings.TrimSpace(r.GetText()); t != "" {
			rows = append(rows, fmt.Sprintf("### Round %d — %s\n\n%s", e.GetRound(), e.GetSeatId(), t))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	return "## Report revision history\n\n" + strings.Join(rows, "\n\n")
}

// grade renders a grade that arrives WITHOUT a type — the JSON board view carries its four
// grades as `any` so an ungraded gap can render `null` rather than a word (see viewjson.go's
// gradeVal). Everything reading the REPLAYED board goes through gradeWord instead, which takes
// the typed value.
//
// It is not a second spelling of gradeWord: the two answer different questions — "is this
// interface value a non-empty string" and "does this enum have a word". Merging them would put
// the `any` assertion back in the path of every typed grade, which is the exact silent zero the
// sevRank note above describes.
func grade(v any) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return "—"
}

// cell keeps a value on one table row: pipes and newlines would break the markdown table.
func cell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func collapseBlanks(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimRight(s, "\n") + "\n"
}

// recordVerification renders the record's own invariant check into the deliverable.
//
// WHY HERE, AND NOT ONLY IN A VERB. `verify` is the operator's instrument and the fuzzer's
// oracle; both are real, and neither is read by the person the report is for. Meanwhile the
// engine declined to fail a lineage gap on the stated grounds that "the record is authoritative
// and already checked there" (#415, #417) — a claim that was not true of any surface a human
// sees. Rendering the result here makes it true in the register the sentence meant: the record
// IS checked, and the check is visible to the reader deciding how much to trust it.
//
// This follows the precedent one section above rather than inventing one. archiveSpotChecks
// already renders a verify-class invariant computed by replay, and says why it renders instead
// of gating: "the DEBT is rendered beside the discharges rather than left to the exit code,
// because a reader deciding how much to trust the closure index needs to know which rounds
// checked it and which did not." Same reasoning, applied to the other six.
//
// IT IS A SECTION, NOT A GATE. Nothing here changes an exit code. `verify` keeps its non-zero
// exit for CI and the fuzzer; assembly reports. A new failure mode in the bench's terminal act
// is a different decision from making a fact visible, and only the second is being made.
//
// THREE STATES (#411). A check that did not apply is not a check that held. Printing both the
// same way is exactly how pass-closes-all-gaps sat inapplicable on every run ever recorded
// while reading as a considered judgement, so `n/a` is its own mark and carries its reason.
func recordVerification(board *record.Board) string {
	checks := verify.Run(board)
	if len(checks) == 0 {
		// Not reachable today, and said out loud rather than rendered as a clean board: an
		// empty check set and a sound record must never be the same output.
		return "## Record verification\n\nNo invariants ran, so this run's record is **unverified**. " +
			"That is not the same as sound."
	}

	var b strings.Builder
	b.WriteString("## Record verification\n\n")
	b.WriteString("Computed by replaying the event record, not reported by any seat. " +
		"These are the invariants that must hold if the record is internally consistent.\n\n")

	held, na, failed := 0, 0, 0
	for _, c := range checks {
		switch {
		case c.NA:
			na++
		case c.OK:
			held++
		default:
			failed++
		}
		fmt.Fprintf(&b, "- **%s** — `%s` — %s\n", c.Status(), c.Name, c.Detail)
		for _, v := range c.Violations {
			fmt.Fprintf(&b, "    - %s\n", v)
		}
	}

	fmt.Fprintf(&b, "\n**%d held · %d did not apply · %d violated.**", held, na, failed)
	if failed > 0 {
		b.WriteString(" A violated invariant is a record-integrity finding: the report's " +
			"content may still be sound, but the record it was assembled from contradicts itself, " +
			"so treat every count derived from it as suspect until the violation is explained.")
	}
	if na > 0 {
		// The whole point of the third state. An invariant that never applies is not evidence
		// of health, and a reader comparing runs is the one who can notice it never does.
		b.WriteString(" An invariant that did not apply checked NOTHING — it is not a pass, and " +
			"one that never applies across runs is a gate worth questioning rather than a clean bill.")
	}
	return b.String()
}
