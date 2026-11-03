package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ONE SECTION FOR EVERY ADJUDICATED EXCHANGE, joined on the motion id (#344).
//
// A motion has an id, so an ask and its answer are one row.
//
// It reads through record.Motions, the one read of the one motion vocabulary.
func motions(fam record.Family) string {
	ms := record.MotionsOf(fam.Events, fam.At)
	if len(ms) == 0 {
		return ""
	}
	var rows []string
	unruled := 0
	for _, m := range ms {
		rows = append(rows, motionRow(m, &unruled))
	}
	out := "Every contested question and how it was answered: grade motions, petitions, rulings on proposed inquiries, and the chair's positions on a PASS it withheld. One mechanism, one id — an ask and its answer are one row.\n\n" +
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
	fmt.Fprintf(&b, "- **%s** · %s · filed by %s #%d", m.ID, motionHead(m), filer, m.Sitting)
	if m.Basis != "" {
		fmt.Fprintf(&b, "\n  - asked: %s", inItem(m.Basis, "    "))
	}
	if m.Relief != "" {
		// The OPERATIVE half of a granted ruling, distinct from the reasoning (#330).
		fmt.Fprintf(&b, "\n  - relief sought: %s", inItem(m.Relief, "    "))
	}
	// A RECONSIDERATION NAMES ITS NEW GROUNDS, as the filer located them.
	if g := m.Fields["grounds"]; g != "" {
		fmt.Fprintf(&b, "\n  - on new grounds: %s", g)
	}
	if q := m.Fields["grounds_quote"]; q != "" {
		fmt.Fprintf(&b, "\n  - on new grounds, the report's text: %s", inItem(q, "    "))
	}
	if m.Ruled() {
		fmt.Fprintf(&b, "\n  - **ruled %s** by %s #%d", m.Ruling, m.RulingBy, m.RulingSitting)
		if m.Opinion != "" {
			fmt.Fprintf(&b, " — %s", inItem(m.Opinion, "    "))
		}
		// THE ORDER A GRANT BINDS THE SEATS TO, beside the reasoning it follows from.
		if m.RulingRelief != "" {
			fmt.Fprintf(&b, "\n  - **relief ordered**, binding %s: %s", m.RulingBinds, inItem(m.RulingRelief, "    "))
		}
		// The work an upheld position names, and the direction it owes.
		switch {
		case m.Work.Lens != "":
			fmt.Fprintf(&b, "\n  - work named: lens %s — %s", m.Work.Lens, inItem(m.Work.Direction, "    "))
		case len(m.Work.GapIDs) > 0:
			fmt.Fprintf(&b, "\n  - work named: gaps %s, each sent back for one more exchange — %s", strings.Join(m.Work.GapIDs, ", "), inItem(m.Work.Direction, "    "))
		}
	} else {
		// THE ABSENCE IS THE INTERESTING CASE and must not fold into the empty set: an unruled
		// motion means the sitting did not happen, which reads identically to no motion at all
		// unless it is said.
		*unruled++
		b.WriteString("\n  - **NOT RULED.** The ask is on the record and no answer is.")
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
	case "inquiry":
		return "inquiry " + m.Fields["inquiry_id"]
	case "position":
		return "the chair's position on a PASS it withheld"
	case "docket":
		// THE SECOND SUBJECT THAT IS ABOUT A GAP, and the row is unreadable without it. Left to
		// the default arm this renders as the bare word "docket" — the bench's disposition of a
		// specific gap, printed as a row naming no gap at all, in the section whose whole promise
		// is "an ask and its answer are one row".
		return "docket " + m.Fields["gap_id"]
	default:
		return m.Subject
	}
}
