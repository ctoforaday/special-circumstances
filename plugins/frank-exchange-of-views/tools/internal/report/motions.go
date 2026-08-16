package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ONE SECTION FOR EVERY ADJUDICATED EXCHANGE, joined on the motion id (#344).
//
// It replaces three renderers that had to be written, and fixed, separately. #315 found the
// petition FILING unrendered while the avenue RULING was found unrendered in the same sweep,
// because nothing said they were one mechanism. And #320 had to render filings and rulings SIDE
// BY SIDE rather than joined, because `petition-rule` carried no id and pairing two filings by
// one seat in one round would have been a guess. A motion has an id, so an ask and its answer are
// one row.
//
// It reads through record.AllMotions, which is the DUAL-READ: a pre-collapse record renders here
// too, from its dispute/petition/avenue-rule events. A consumer calling record.Motions directly
// would render NOTHING for an old record and look exactly like a run that had no disputes and no
// petitions — the plausible zero this stage exists to remove.
func motions(board *record.Board) string {
	ms := record.Motions(board)
	if len(ms) == 0 {
		return ""
	}
	var rows []string
	unruled := 0
	for _, m := range ms {
		rows = append(rows, motionRow(m, &unruled))
	}
	out := "## Motions\n\nEvery contested question and how it was answered: grade disputes, petitions, and rulings on proposed directions. One mechanism, one id — an ask and its answer are one row.\n\n" +
		strings.Join(rows, "\n")
	if unruled > 0 {
		out += fmt.Sprintf("\n\n**%d motion(s) received no ruling.** A motion is answered before the debate moves on; a filing with no ruling means that sitting is missing, not that the ask was withdrawn.", unruled)
	}
	return out
}

func motionRow(m *record.Motion, unruled *int) string {
	var b strings.Builder
	filer := m.Filer
	if filer == "" {
		filer = "the record does not say"
	}
	fmt.Fprintf(&b, "- **%s** · %s · filed by %s (r%d)", m.ID, motionHead(m), filer, m.Round)
	if m.Basis != "" {
		fmt.Fprintf(&b, "\n  - asked: %s", m.Basis)
	}
	if m.Relief != "" {
		// The OPERATIVE half of a granted ruling, distinct from the reasoning (#330).
		fmt.Fprintf(&b, "\n  - relief sought: %s", m.Relief)
	}
	if m.Ruled() {
		fmt.Fprintf(&b, "\n  - **ruled %s** by %s (r%d)", m.Ruling, m.RulingBy, m.RulingRound)
		if m.Opinion != "" {
			fmt.Fprintf(&b, " — %s", m.Opinion)
		}
	} else {
		// THE ABSENCE IS THE INTERESTING CASE and must not fold into the empty set: an unruled
		// motion means the sitting did not happen, which reads identically to no motion at all
		// unless it is said.
		*unruled++
		b.WriteString("\n  - **NOT RULED.** The ask is on the record and no answer is.")
	}
	if m.Appealed {
		fmt.Fprintf(&b, "\n  - **appealed** — the filer pressed on after the ruling: %s", m.AppealReason)
	}
	if m.Fields["legacy"] == "true" {
		b.WriteString("\n  - _(id assigned at read time: this run predates the motion collapse, and its exchanges carried no identity of their own)_")
	}
	return b.String()
}

// motionHead names what the motion is ABOUT, from the subject's own fields.
func motionHead(m *record.Motion) string {
	switch m.Subject {
	case "grade":
		return fmt.Sprintf("grade of %s (%s to %s)", m.Fields["gap_id"], m.Fields["dimension"], m.Fields["proposed"])
	case "petition":
		return "petition (" + m.Fields["class"] + ")"
	case "direction":
		return "direction " + m.Fields["avenue_id"]
	default:
		return m.Subject
	}
}
