package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE REPORT STATES EACH PART OF ITS QUESTION IT DOES NOT ANSWER, AND THE TOOL COMPOSES THE SENTENCE
// (ruled by gblock 2026-11-11, J: a limitation's cause is a field on the coverage record, and the tool
// composes the sentence). The cause is the part's coverage — its sources not obtained, or the part
// excluded from the report's answer — and the sentence is in the subject's voice: it says what the
// report does not answer and why, never how the run went. The coverage's free reason is the logic
// lens's to judge and is printed in run.md's coverage table, never here.

const coverageLine = "**Not answered here:**"

// coverageSentence is report.md's sentence for a question part not covered, "" for a covered part or
// one with no coverage on the record.
func coverageSentence(p record.QuestionPartRow) string {
	text := strings.TrimRight(strings.TrimSpace(p.Text), ".?!")
	switch p.Coverage {
	case recordpb.Word(recordpb.PartCoverage_PART_COVERAGE_NOT_OBTAINED):
		return "*" + text + "* is not answered here: its sources could not be obtained."
	case recordpb.Word(recordpb.PartCoverage_PART_COVERAGE_EXCLUDED):
		return "*" + text + "* is not answered here: it lies outside this report's answer."
	}
	return ""
}

// notAnsweredLine is report.md's line under the outcome naming each part of the question the report
// does not answer, "" where it answers every part the record holds.
func notAnsweredLine(parts []record.QuestionPartRow) string {
	var out []string
	for _, p := range parts {
		if s := coverageSentence(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return coverageLine + " " + strings.Join(out, " ")
}

// coverageTable is run.md's account of the question's coverage: each question part with its coverage,
// the reason recorded for it and whether a check stands — "" where the record holds no part.
func coverageTable(parts []record.QuestionPartRow) string {
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("| Part | Question part | Coverage | Reason | Checked |\n|---|---|---|---|---|\n")
	for _, p := range parts {
		cov, reason, checked := p.Coverage, p.Reason, "no"
		if cov == "" {
			cov = "none recorded"
		}
		if p.State == record.PartSound {
			checked = "yes (" + p.CheckedBy + ")"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", p.ID, cell(p.Text), cov, cell(reason), checked)
	}
	return b.String()
}

// partsAt is the question parts by coverage at a point on the assembly's run.
func (a *assembly) partsAt(point int64) (record.PartCoverages, error) {
	return record.PartCoveragesAt(a.run, point)
}
