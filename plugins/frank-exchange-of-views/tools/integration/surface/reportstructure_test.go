package surface

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// NO SEAT IS TOLD THAT A REPORT'S ARGUMENT OWES A HEADING.
//
// The report's structure is its author's: the sections of its argument ship as the body under the
// headings their author gave them. The three surfaces the synthesis prompt asks blue to write —
// TL;DR, The Catechism, Open questions — are that seat's duty and no more: no write refuses a
// report for lacking one and no lens audits one by name. The template once held "Technical foundations" and "Analysis"
// as slots, blue's prompts listed them, the logic lens's constitution named "template compliance"
// as a thing it audits, and assembly wrote "Additional findings" above whatever fitted neither.
// On universe-m21 the report's sections named the reader's questions, as the synthesis prompt
// asks, and the logic lens minted a gap for the two headings it lacked. A surface a seat reads
// that names one of them again is that reading coming back.
//
// AN ORDER HELD IS NOT A HEADING OWED. The write holds those three surfaces to their places where
// the text carries them — TL;DR first, the Catechism second, Open questions last — and the prompts
// say so; it takes a report with none of them, and the surfaces a seat reads state the order of
// what a report has, never that it must have it (report.TestATextInOrderIsNotRefused, and the last
// case of cli.TestIngestRefusesASectionOutOfOrder, which freezes a report with none of the three).
func TestNoSeatIsToldAReportOwesAHeading(t *testing.T) {
	audit, err := repotree.Plugin("skills", "adversarial-audit", "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(agentFacingFiles(t), audit) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, told := range []string{"template compliance", "echnical foundations", "## Analysis", "foundations, analysis", "foundations / analysis", "Additional findings"} {
			if strings.Contains(string(b), told) {
				t.Errorf("%s says %q: a heading required of a report, or an audit of the report against a template", pluginRel(t, path), told)
			}
		}
	}
}

// NO SEAT IS TOLD THAT ASSEMBLY MOVES THE REPORT'S TEXT.
//
// The report ships in the order its text is written, and the write holds that order. While
// assembly placed three of blue's sections by name, the surfaces said so — "Assembly LIFTS these
// verbatim", "lifted whole" — and a seat told that reads the order of the text as one assembly
// will repair: universe-m21's Catechism was written after the argument it answers from and
// shipped ahead of it. A surface that says it again describes a second order that does not exist.
func TestNoSeatIsToldAssemblyMovesTheReportsText(t *testing.T) {
	audit, err := repotree.Plugin("skills", "adversarial-audit", "SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(agentFacingFiles(t), audit) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ToLower(string(b))
		for _, told := range []string{"lifts these", "lifted verbatim", "lifted whole", "lifted surface", "lifted section", "lifted from it by assembly", "assembly lifts", "assembly moves the", "moved to the top"} {
			if strings.Contains(text, told) {
				t.Errorf("%s says %q: the report's text ships where its author wrote it", pluginRel(t, path), told)
			}
		}
	}
}
