package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// requireOneGapPerClassAndSubject refuses a mint where an OPEN gap of the same class already stands
// on the same subject: two defects of one class on one subject are one gap (ruled by gblock
// 2026-10-28). The second lens's point is a finding about the gap that stands.
//
// THE SUBJECT IS READ FROM FIELDS THE RECORD HOLDS AS IDENTITIES, never from the problem's words:
//   - a quote gap's subject is the sentence its anchor sits in (gapSentenceIn), the sentence the
//     board shows; the new gap's anchor is placed on the report the mint lands on, and two anchors
//     in one sentence span are on one sentence whatever span of it each quoted;
//   - an inquiry gap's subject is the inquiry id it names (about_kind inquiry, about_ref).
//
// A section, a gap or an anchor named by --about is not such a subject: a section holds many
// sentences, and a gap about a gap or an anchor is a dispute the record answers in its own terms. The
// rule does not apply there — so never to the gap a failed check obliges (answersOwnFailedCheck),
// which is about the anchor and, at the CLI, names no quote: refused here, the lens would owe a gap
// the record turns away, and a finding does not answer the check.
//
// MEASURED (universe-m35): the parallel first-wave lenses minted five pairs, each pair 25 s to 2 min
// apart, each second minter having read the board before its sibling minted. Four of the five are
// one class on one subject. The check runs in the mint's own transaction, which takes the write lock
// at BEGIN (`_txlock=immediate`), so the second of two concurrent mints reads the first.
//
// WHAT THE RECORD PROVES IS ANOTHER RELATION IS ADMITTED:
//   - a gap the mint names in --supersedes does not block it: a successor states its ancestor's
//     subject by design;
//   - a mint whose --found-by names the finding that raised an unanswered contradiction is that
//     contradiction's answer, which nothing else gives: refusing it would hold the PASS on a
//     contradiction no act could answer.
//
// A closed gap does not block: a defect that recurs after its gap closed is a new gap.
//
// MIGRATE REPLAYS ARCHIVED RECORDS, which hold such pairs: an archived mint is what a seat did, so
// replay admits it.
func requireOneGapPerClassAndSubject(w writeView, m *recordpb.Mint) error {
	if Migrating {
		return nil
	}
	quote := m.GetLocation()
	inquiry := ""
	if m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_INQUIRY {
		inquiry = m.GetAboutRef()
	}
	if strings.TrimSpace(quote) == "" && inquiry == "" {
		return nil
	}
	held, err := openGapsOfClassAt(w.q, m.GetClass(), inquiry)
	if err != nil {
		return fmt.Errorf("the board's open gaps could not be read, so whether one of class %s already stands on this subject cannot be told: %w", recordpb.Word(m.GetClass()), err)
	}
	held = slices.DeleteFunc(held, func(g heldGap) bool { return slices.Contains(m.GetSupersedes(), g.id) })
	if len(held) == 0 {
		return nil
	}
	if answers, err := answersRaisedContradictionAt(w.q, m.GetFoundBy()); err != nil {
		return fmt.Errorf("the record's contradictions could not be read, so whether this mint answers one cannot be told: %w", err)
	} else if answers {
		return nil
	}
	if inquiry != "" {
		return oneGapRefusal(m, held[0], "inquiry "+inquiry)
	}
	report, err := w.report()
	if err != nil {
		return fmt.Errorf("the report does not render, so the sentence this gap quotes cannot be found to ask whether an open gap of class %s stands on it: %w", recordpb.Word(m.GetClass()), err)
	}
	// A quote the report does not hold once places nothing: the anchor's own placement refuses it,
	// in its own words, after this check.
	placed, err := anchortext.Attach(report, m.GetGapId(), quote)
	if err != nil {
		return nil
	}
	spans, bare := anchor.Sentences(placed), claimcount.BareAnchorIDs(placed)
	at, ok := gapSentenceIn(placed, spans, bare, m.GetGapId())
	if !ok {
		return nil
	}
	for _, g := range held {
		if sp, ok := gapSentenceIn(placed, spans, bare, g.id); ok && sp == at {
			return oneGapRefusal(m, g, "the sentence "+fmt.Sprintf("%q", strings.Join(strings.Fields(claimcount.StripAnchors(placed[at[0]:at[1]])), " ")))
		}
	}
	return nil
}

// heldGap is an open gap as the one-gap rule names it: its id, who minted it and what it says is wrong.
type heldGap struct {
	id, mintedBy, problem string
}

// openGapsOfClassAt is the open gaps of class on the gap view, in mint order: those about inquiry
// where inquiry is not empty, else those with a quote, whose sentence the caller reads.
func openGapsOfClassAt(q recordsql.Querier, class recordpb.GapClass, inquiry string) ([]heldGap, error) {
	if noRecord(q) {
		return nil, nil
	}
	query := `SELECT "gap_id", COALESCE("minted_by", ''), COALESCE("problem", '') FROM "gap"
	           WHERE "open" AND "class" = ? AND COALESCE("location", '') <> '' ORDER BY "minted_seq"`
	args := []any{recordpb.Word(class)}
	if inquiry != "" {
		query = `SELECT "gap_id", COALESCE("minted_by", ''), COALESCE("problem", '') FROM "gap"
		          WHERE "open" AND "class" = ? AND "about_kind" = ? AND "about_ref" = ? ORDER BY "minted_seq"`
		args = append(args, recordpb.Word(recordpb.AboutKind_ABOUT_KIND_INQUIRY), inquiry)
	}
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []heldGap
	for rows.Next() {
		var g heldGap
		if err := rows.Scan(&g.id, &g.mintedBy, &g.problem); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// answersRaisedContradictionAt is whether foundBy names the finding that raised a contradiction no
// gap answers yet: the mint is that contradiction's answer, whoever raised it.
func answersRaisedContradictionAt(q recordsql.Querier, foundBy []string) (bool, error) {
	if len(foundBy) == 0 {
		return false, nil
	}
	evs, win, err := eventsOfAt(q, contradictionActs...)
	if err != nil {
		return false, err
	}
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.raisedBy != "" && c.answeredBy == "" && slices.Contains(foundBy, c.raisedBy) {
			return true, nil
		}
	}
	return false, nil
}

// oneGapRefusal names the open gap that holds the subject and the act the point is owed as.
func oneGapRefusal(m *recordpb.Mint, g heldGap, subject string) error {
	return feov.Errorf(feov.Conflict,
		"gap %s, minted by %s, is open on %s with class %s: %q. Two defects of one class on one subject are one gap, so nothing was written. "+
			"Record your point as a finding about that gap — `finding --about-kind gap --about %s` — where its minter and blue read it beside the gap",
		g.id, g.mintedBy, subject, recordpb.Word(m.GetClass()), g.problem, g.id)
}
