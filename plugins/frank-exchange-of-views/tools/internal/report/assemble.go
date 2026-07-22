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
//     matrix (the board), the expansions and alternatives (avenue events by fate), the red
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
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Assemble writes <runDir>/report.md and returns its path. It reads the board once (which
// carries both the ordered event log and the replayed gaps) and blue/report.md, and composes
// the report from those two — no --inputs, no render-shadow round-trip.
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
	p(orientation(board, evs))
	p(sectionOr(blue, "TL;DR"))
	p(sectionOr(blue, "The Catechism"))
	p(sectionOr(blue, "Technical foundations"))
	p(sectionOr(blue, "Analysis"))

	// Tool-composed from the record.
	p(riskMatrix(bj))
	p(avenues(evs, "The expansions", accepted))
	p(avenues(evs, "Alternatives considered", rejected))
	p(sectionOr(blue, "Open questions"))
	// The embed carries ONLY blue content not already composed above — its lifted synthesis
	// surfaces and any tool-owned sections it wrongly authored are dropped (see blueEmbed).
	// If nothing genuinely additional survives, the section is omitted rather than left empty.
	if extra := blueEmbed(blue); extra != "" {
		p("## Blue team report (sections not composed above)\n\n" + extra)
	}
	p(redFindings(board))
	p(debate(evs))

	out := collapseBlanks(b.String())
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
// author — the risk matrix, red findings, the debate, a verdict — are dropped as fabrication
// (the tool composes those from the record; blue cannot know red's findings or the run's
// outcome). What survives is genuinely ADDITIONAL blue content — including blue's Footnotes,
// its citation apparatus, which the tool does not yet compose a bibliography for, so they are
// KEPT here rather than lost. Empty output means blue authored only what it should; the caller
// then omits the embed entirely rather than lift-AND-embed the same sections twice.
func blueEmbed(blue string) string {
	drop := map[string]bool{
		// lifted to the top verbatim
		"tl;dr": true, "the catechism": true, "technical foundations": true,
		"analysis": true, "open questions": true,
		// tool-owned — composed from the record, never blue's to author. NOTE: footnotes are
		// NOT here — they are blue's own citations and nothing else composes them yet.
		"risk matrix": true, "the expansions": true, "expansions": true,
		"alternatives considered": true, "red team findings": true,
		"the debate": true, "blue team report": true, "verdict": true,
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

// outcomeOf returns the LAST terminal outcome event's payload, or nil if none was recorded.
func outcomeOf(evs []record.Event) *record.Payload {
	var last *record.Payload
	for _, e := range evs {
		if e.Type == "outcome" {
			last = e.Payload
		}
	}
	return last
}

// verdictStamp composes the verdict line from the terminal outcome event. A missing outcome
// is flagged, never invented — the same invariant as a missing blue section.
func verdictStamp(o *record.Payload) string {
	if o == nil {
		return "**Verdict:** _(no terminal outcome recorded — `bench outcome` was not run before assembly)_"
	}
	switch o.Str("verdict") {
	case "CEILING":
		return "**Verdict:** CEILING-TERMINATED — the run hit its round ceiling while still converging. This is NOT a judged failure to verify and must not be read as one: gaps remain open, the final blue revision was never audited by a red pass, and that re-audit debt travels OUT of the run."
	case "HALTED":
		return "**Verdict:** HALTED — the bench ended this run. The halt opinion is on the record below (Bench disposition) and is relayed to the human verbatim, never smoothed."
	default:
		by := ""
		switch {
		case payloadBool(o, "deadlocked"):
			by = " by judged deadlock"
		case payloadBool(o, "exhausted"):
			by = " by safety ceiling"
		}
		return fmt.Sprintf("**Verdict:** %s%s", o.Str("verdict"), by)
	}
}

// orientation is the reviewer-facing "read this first": the open gaps a human should
// re-examine, most severe first, preceded by the bench's terminal ask if it evented one. It
// authors no new judgement — it ORDERS the board (severity, then impact, then likelihood) and
// PROMOTES the bench's already-evented voice (certify/halt), which otherwise sits buried in
// the debate's Bench-disposition line. When the bench never certified/halted (as in a
// ceiling-terminated run), only the ranked gaps show; nothing is invented to fill the space.
func orientation(board *record.Board, evs []record.Event) string {
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
	for _, e := range evs {
		switch e.Type {
		case "certify":
			if s := e.Payload.Str("statement"); s != "" {
				b.WriteString("**The bench asks a human to re-examine:** " + s + "\n\n")
			}
		case "halt":
			if s := e.Payload.Str("opinion"); s != "" {
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
		fmt.Fprintf(&b, "%d. **[%s]** %s (%s) — %s\n", i+1, grade(r.g.Severity), concise(r.g.Mint.Str("problem")), r.g.ID, concise(r.g.Mint.Str("required_fix")))
	}
	return strings.TrimRight(b.String(), "\n")
}

// sevRank maps a grade (critical..low) to 4..1, unknown to 0 — the ordering key for the
// orientation ranking. It reads any because a gap's grades arrive as interface values.
func sevRank(v any) int {
	switch strings.ToLower(grade(v)) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	}
	return 0
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

// avenue fate: the user's mapping — an avenue PURSUED is a concept expansion accepted; an
// avenue abandoned or declined is an alternative considered, its reason the counter.
func accepted(status string) bool { return status == "pursued" }
func rejected(status string) bool { return status == "abandoned" || status == "declined" }

// avenues renders the avenue events whose status matches want, under the given heading.
func avenues(evs []record.Event, heading string, want func(string) bool) string {
	var rows []string
	for _, e := range evs {
		if e.Type != "avenue" || !want(e.Payload.Str("status")) {
			continue
		}
		method := ""
		if m := e.Payload.Str("method"); m != "" {
			method = fmt.Sprintf(" _(%s)_", m)
		}
		reason := ""
		if r := e.Payload.Str("reason"); r != "" {
			reason = " — " + r
		}
		status := ""
		if s := e.Payload.Str("status"); s != "pursued" {
			status = fmt.Sprintf(" [%s]", s) // abandoned vs declined is the shape of the counter
		}
		rows = append(rows, fmt.Sprintf("- **%s**%s%s%s (%s)", e.Payload.Str("line"), method, status, reason, e.SeatID))
	}
	body := "_(none on the record)_"
	if len(rows) > 0 {
		body = strings.Join(rows, "\n")
	}
	return "## " + heading + "\n\n" + body
}

// redFindings composes the full findings from the board's gaps: every open gap with its
// grades and required fix, then the closure index. This is the ledger's content, drawn from
// the replayed board rather than read back from the projection file.
func redFindings(board *record.Board) string {
	var open, closed []string
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		if g.Open {
			regraded := ""
			if n := len(g.Regrades); n > 0 {
				regraded = fmt.Sprintf(" · regraded x%d (latest basis: %s)", n, g.Regrades[n-1].Str("basis"))
			}
			open = append(open, fmt.Sprintf("### %s — %s\n%s\nseverity %s | %s x %s | cx %s | class %s%s\nrequired_fix: %s\nacceptance_check: %s",
				g.ID, g.Mint.Str("problem"),
				g.Mint.Str("location"),
				grade(g.Severity), grade(g.Likelihood), grade(g.Impact), grade(g.ComplexityCost), grade(g.Mint.Str("class")),
				regraded,
				g.Mint.Str("required_fix"),
				g.Mint.Str("acceptance_check")))
		} else {
			cc := g.Closure.Str("closure_class")
			if cc == "" {
				cc = "closed"
			}
			succ := g.Closure.Str("successor")
			if succ == "" {
				succ = "-"
			}
			closed = append(closed, fmt.Sprintf("- %s | %s | %s | successor %s", g.ID, cc, g.Mint.Str("problem"), succ))
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
	return b.String()
}

// unmintedFindings renders the lens findings whose label is credited by NO gap's found_by, or
// "" if every finding earned a gap. Ordered by the event log so the section is deterministic.
func unmintedFindings(board *record.Board) string {
	minted := map[string]bool{}
	for _, g := range board.Gaps {
		if g == nil || g.Mint == nil {
			continue
		}
		for _, lbl := range g.Mint.StrList("found_by") {
			minted[lbl] = true
		}
	}
	var rows []string
	for _, e := range board.Events {
		if e.Type != "finding" {
			continue
		}
		lbl := e.Payload.Str("label")
		if lbl != "" && minted[lbl] {
			continue
		}
		head := lbl
		if head == "" {
			head = e.Payload.Str("finding_id")
		}
		loc := e.Payload.Str("location")
		if loc != "" {
			loc = " — " + loc
		}
		rows = append(rows, fmt.Sprintf("### %s%s\nseverity %s | %s x %s | %s\n%s",
			head, loc,
			grade(e.Payload.Str("severity")), grade(e.Payload.Str("likelihood")), grade(e.Payload.Str("impact")),
			e.SeatID,
			e.Payload.Str("text")))
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
func debate(evs []record.Event) string {
	var order []int
	byRound := map[int][]record.Event{}
	for _, e := range evs {
		if _, seen := byRound[e.Round]; !seen {
			order = append(order, e.Round)
		}
		byRound[e.Round] = append(byRound[e.Round], e)
	}

	var parts []string
	for _, r := range order {
		re := byRound[r]
		var round []string
		for _, e := range re {
			switch {
			case e.Type == "position" && strings.HasPrefix(e.SeatID, "red-merge"):
				round = append(round, "### RED\n"+e.Payload.Str("text"))
			case e.Type == "closing" && strings.HasPrefix(e.SeatID, "red-merge"):
				round = append(round, fmt.Sprintf("### RED CLOSING — %s\n%s", e.Payload.Str("gap_id"), e.Payload.Str("text")))
			case e.Type == "position" && strings.HasPrefix(e.SeatID, "blue"):
				round = append(round, "### BLUE\n"+e.Payload.Str("text"))
			case e.Type == "closing" && strings.HasPrefix(e.SeatID, "blue"):
				round = append(round, fmt.Sprintf("### BLUE CLOSING — %s\n%s", e.Payload.Str("gap_id"), e.Payload.Str("text")))
			}
		}
		// Grade disputes and their answers — the claim-level alternative and its counter.
		var disp []string
		for _, e := range re {
			switch e.Type {
			case "dispute":
				disp = append(disp, fmt.Sprintf("- **%s** disputes %s/%s → %s: %s", e.SeatID, e.Payload.Str("gap_id"), e.Payload.Str("dimension"), e.Payload.Str("proposed"), e.Payload.Str("basis")))
			case "dispute-respond":
				disp = append(disp, fmt.Sprintf("  - answered (%s): %s", e.Payload.Str("as"), e.Payload.Str("basis")))
			}
		}
		if len(disp) > 0 {
			round = append(round, "### Grade disputes\n"+strings.Join(disp, "\n"))
		}
		// The bench's in-round acts: opinions and petition rulings.
		var lead []string
		for _, e := range re {
			switch e.Type {
			case "opinion":
				lead = append(lead, fmt.Sprintf("- %s: %s — principle: %s; tension: %s; review: %s\n%s",
					e.Payload.Str("gap_id"), e.Payload.Str("disposition"), e.Payload.Str("principle"),
					e.Payload.Str("tension"), e.Payload.Str("review_flag"), e.Payload.Str("rationale")))
			case "petition-rule":
				lead = append(lead, fmt.Sprintf("- petition %s: %s — %s", e.Payload.Str("petitioner"), e.Payload.Str("ruling"), e.Payload.Str("rationale")))
			}
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
	for _, e := range evs {
		switch e.Type {
		case "halt":
			disp = append(disp, "**HALT** — "+e.Payload.Str("opinion"))
		case "certify":
			disp = append(disp, "**Certification** — "+e.Payload.Str("statement"))
		}
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

func grade(v any) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return "—"
}

func payloadBool(p *record.Payload, k string) bool {
	if v, ok := p.Get(k); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
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
