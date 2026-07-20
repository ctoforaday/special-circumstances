// Package report assembles the final research report — MECHANICALLY, inside the tool.
//
// Assembly was the single longest seat (505s in smoke-c), running last and alone, and almost
// none of it is judgment: the catechism, technical foundations and analysis are COPIED verbatim
// from blue's report; blue's report and red's board embed in full; the risk matrix is every open
// gap's grades in a table; the verdict stamp is templated. A generating agent regenerated all of
// that. Here the copy is done by code — which also means it CANNOT be authored, and an
// assembly-authored catechism (6/7 answers regressed to pre-repair phrasings) is exactly the
// run-5 defect this makes structurally impossible. The agent shrinks to the TL;DR and the
// per-round synopsis, handed in through Inputs.
//
// The sections are RAW-SLICED, never parsed-and-rendered: round-tripping markdown through an AST
// would normalise whitespace and reflow, which is authorship. A fence-aware scan finds the true
// heading boundaries so a "## " line inside a code block is not mistaken for one.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Inputs are the values the assemble seat supplies: the run-level facts it holds and the two
// things that are genuinely its judgment — the TL;DR and the debate synopsis.
type Inputs struct {
	Verdict          string   `json:"verdict"`
	Topic            string   `json:"topic"`
	Rounds           int      `json:"rounds"`
	TLDR             string   `json:"tldr"`
	Synopsis         string   `json:"synopsis"`
	OpenQuestions    []string `json:"open_questions"`
	HaltOpinion      string   `json:"halt_opinion"`
	Deadlocked       bool     `json:"deadlocked"`
	Exhausted        bool     `json:"exhausted"`
	GradeAdjustments []string `json:"grade_adjustments"`
	PetitionLog      []string `json:"petition_log"`
	InfraDebts       []string `json:"infra_debts"`
}

// Assemble writes <runDir>/report.md and returns its path. It renders the projections once
// (in-process) so the embedded ledger/archive and the board risk matrix reflect the final state.
func Assemble(runDir string, in Inputs) (string, error) {
	blue := readOr(filepath.Join(runDir, "blue", "report.md"), "")

	if _, err := record.Render(runDir, ""); err != nil {
		return "", fmt.Errorf("assemble: render before embed: %w", err)
	}
	shadow := filepath.Join(recordsDir(runDir), "render-shadow")
	ledger := readOr(filepath.Join(shadow, "ledger.md"), "_(ledger unavailable)_")
	archive := readOr(filepath.Join(shadow, "archive.md"), "_(archive unavailable)_")
	loi := readOr(filepath.Join(shadow, "lines-of-inquiry.md"), "")

	board, err := record.BoardState(runDir)
	if err != nil {
		return "", fmt.Errorf("assemble: board: %w", err)
	}
	bj := record.BoardJSONOf(board)

	var b strings.Builder
	p := func(s string) { b.WriteString(s); b.WriteString("\n\n") }

	p(stamp(runDir, in))
	p(sectionOr(blue, "The Catechism"))
	p(sectionOr(blue, "Technical foundations"))
	p(sectionOr(blue, "Analysis"))
	p(riskMatrix(bj))
	if in.Verdict == "HALTED" && strings.TrimSpace(in.HaltOpinion) != "" {
		p("## Halt opinion (verbatim)\n\n" + in.HaltOpinion)
	}
	p("## Blue team report (in full)\n\n" + orMissing(blue, "blue/report.md"))
	p("## Red team findings (in full)\n\n" + ledger + "\n\n" + archive)
	syn := in.Synopsis
	if strings.TrimSpace(syn) == "" {
		syn = "_(per-round synopsis not supplied)_ — the literal transcript is at " + filepath.Join(runDir, "debate.md")
	}
	p("## Debate record\n\n" + syn)
	if strings.TrimSpace(loi) != "" {
		p(strings.TrimRight(loi, " \t\n"))
	}
	p("## Open questions carried past this run\n\n" + list(in.OpenQuestions, "_(none)_"))
	for _, s := range []struct {
		title string
		items []string
	}{
		{"Judicial record — petitions", in.PetitionLog},
		{"Terminal grade adjustments", in.GradeAdjustments},
		{"Infrastructure debts", in.InfraDebts},
	} {
		if len(s.items) > 0 {
			p("## " + s.title + "\n\n" + list(s.items, ""))
		}
	}

	out := collapseBlanks(b.String())
	path := filepath.Join(runDir, "report.md")
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return "", fmt.Errorf("assemble: write report.md: %w", err)
	}
	return path, nil
}

// recordsDir mirrors the record package's layout without exporting it.
func recordsDir(runDir string) string { return filepath.Join(runDir, "records") }

func readOr(path, fallback string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return string(b)
}

func orMissing(md, name string) string {
	if strings.TrimSpace(md) == "" {
		return "_(" + name + " is missing)_"
	}
	return md
}

// section returns the "## heading" block verbatim (heading included), or "" if absent. It tracks
// fenced code blocks so a "## " line inside ``` or ~~~ is not read as a heading.
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
			if t == "## "+heading {
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
	return fmt.Sprintf("## %s\n\n_(blue/report.md has no \"## %s\" section — not authored here; see the full blue report below.)_", heading, heading)
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
		disp := g.RequiredFix
		if g.Class == "risk_accepted" || g.Class == "risk_argued" {
			disp = "risk_accepted — " + disp
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			cell(risk), grade(g.Likelihood), grade(g.Impact), grade(g.ComplexityCost), cell(disp)))
	}
	b.WriteString("\nRisk-accepted items are recorded here with their rationale — elevated, never dropped.")
	return b.String()
}

func stamp(runDir string, in Inputs) string {
	var head string
	switch in.Verdict {
	case "CEILING":
		head = "**Verdict:** CEILING-TERMINATED — the run hit its round ceiling while still converging. This is NOT a judged failure to verify and must not be read as one: gaps remain open, the final blue revision was never audited by a red pass, and that re-audit debt travels OUT of the run."
	case "HALTED":
		head = "**Verdict:** HALTED — the bench ended this run. The halt opinion is quoted verbatim below and relayed to the human, never smoothed."
	default:
		by := ""
		if in.Deadlocked {
			by = " by judged deadlock"
		} else if in.Exhausted {
			by = " by safety ceiling"
		}
		head = fmt.Sprintf("**Verdict:** %s%s · **Run:** `%s`", in.Verdict, by, runDir)
	}
	tldr := in.TLDR
	if strings.TrimSpace(tldr) == "" {
		tldr = "_(none supplied)_"
	}
	return fmt.Sprintf("# %s — research report\n\n%s\n\n**TL;DR:** %s", in.Topic, head, tldr)
}

func list(items []string, empty string) string {
	if len(items) == 0 {
		return empty
	}
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = "- " + s
	}
	return strings.Join(out, "\n")
}

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
