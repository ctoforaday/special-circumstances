package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func TestSectionCopiesVerbatimAndIsFenceAware(t *testing.T) {
	md := strings.Join([]string{
		"# report", "",
		"## The Catechism", "", "Q1: kept  as-is.  ", "", // trailing spaces preserved inside
		"```", "## Technical foundations", "not a heading — inside a fence", "```", "",
		"## Analysis", "", "the analysis.", "",
	}, "\n")

	cat := section(md, "The Catechism")
	if !strings.Contains(cat, "Q1: kept  as-is.") {
		t.Errorf("catechism not copied verbatim: %q", cat)
	}
	// The fenced "## Technical foundations" must NOT end the Catechism section.
	if !strings.Contains(cat, "not a heading — inside a fence") {
		t.Errorf("a ## inside a code fence wrongly ended the section: %q", cat)
	}
	// A real "## Technical foundations" section does not exist (only the fenced one), so it
	// is reported missing rather than authored.
	if got := section(md, "Technical foundations"); got != "" {
		t.Errorf("a fenced heading was mistaken for a real one: %q", got)
	}
	if !strings.Contains(sectionOr(md, "Technical foundations"), "not authored here") {
		t.Error("a missing section must be flagged, never authored")
	}
}

func TestRiskMatrixFromBoard(t *testing.T) {
	bj := record.BoardJSON{Open: []record.GapJSON{
		{ID: "R1-1", Problem: "overclaims capture", Likelihood: "high", Impact: "medium", RequiredFix: "grep the sites"},
		{ID: "R1-2", Problem: "cost model rough", Likelihood: "low", Impact: "low", Class: "risk_accepted", RequiredFix: "accepted: low blast radius"},
	}}
	m := riskMatrix(bj)
	if !strings.Contains(m, "| overclaims capture | high | medium | — | grep the sites |") {
		t.Errorf("open gap row wrong:\n%s", m)
	}
	if !strings.Contains(m, "risk_accepted — accepted: low blast radius") {
		t.Errorf("risk_accepted disposition not marked:\n%s", m)
	}
	// An absent grade is a dash, not the string "undefined" or empty.
	if !strings.Contains(m, "| — |") {
		t.Errorf("absent complexity grade should render as a dash:\n%s", m)
	}
	empty := riskMatrix(record.BoardJSON{})
	if !strings.Contains(empty, "no open gaps") {
		t.Errorf("empty board should say so:\n%s", empty)
	}
}

func TestStampVariesByVerdict(t *testing.T) {
	base := Inputs{Topic: "t", TLDR: "the answer."}
	base.Verdict = "UNVERIFIED"
	if s := stamp("/run", base); !strings.Contains(s, "**Verdict:** UNVERIFIED") || !strings.Contains(s, "**TL;DR:** the answer.") {
		t.Errorf("UNVERIFIED stamp: %q", s)
	}
	base.Verdict = "CEILING"
	if s := stamp("/run", base); !strings.Contains(s, "CEILING-TERMINATED") || !strings.Contains(s, "judged failure") {
		t.Errorf("CEILING stamp must not read as a failure: %q", s)
	}
	base.Verdict = "HALTED"
	if s := stamp("/run", base); !strings.Contains(s, "HALTED") {
		t.Errorf("HALTED stamp: %q", s)
	}
}

// cell keeps a value on one table row: a pipe or newline would break the table.
func TestCellEscapesTableBreakers(t *testing.T) {
	if got := cell("a | b\nc"); strings.ContainsAny(got, "\n") || strings.Contains(got, " | ") {
		t.Errorf("cell did not neutralise a pipe/newline: %q", got)
	}
	if cell("   ") != "—" {
		t.Error("a blank cell should be a dash")
	}
}
