package surface

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// NO SEAT IS TOLD THAT A REPORT OWES A HEADING.
//
// The report's structure is its author's: the sections of its argument ship as the body under the
// headings their author gave them. The template once held "Technical foundations" and "Analysis"
// as slots, blue's prompts listed them, the logic lens's constitution named "template compliance"
// as a thing it audits, and assembly wrote "Additional findings" above whatever fitted neither.
// On universe-m21 the report's sections named the reader's questions, as the synthesis prompt
// asks, and the logic lens minted a gap for the two headings it lacked. A surface a seat reads
// that names one of them again is that reading coming back.
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
		for _, told := range []string{"template compliance", "echnical foundations", "## Analysis", "foundations, analysis", "Additional findings"} {
			if strings.Contains(string(b), told) {
				t.Errorf("%s says %q: a heading required of a report, or an audit of the report against a template", pluginRel(t, path), told)
			}
		}
	}
}
