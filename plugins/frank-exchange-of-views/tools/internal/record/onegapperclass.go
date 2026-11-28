package record

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
)

// requireOneGapPerClassAndSubject refuses a mint where an OPEN gap of the same class already stands
// on the same subject: two defects of one class on one subject are one gap (ruled by gblock
// 2026-10-28). The refusal names the refused seat's road from the gap that stands (oneGapRefusal).
//
// THE SUBJECT IS READ FROM FIELDS THE RECORD HOLDS AS IDENTITIES, never from the problem's words:
//   - a quote gap's subject is the sentence its anchor sits in (gapSentenceIn), the sentence the
//     board shows; the new gap's anchor is placed on the report the mint lands on, and two anchors
//     in one sentence span are on one sentence whatever span of it each quoted;
//   - an inquiry gap's subject is the inquiry id it names (about_kind inquiry, about_ref);
//   - an anchor gap's subject is the anchor it names (about_kind anchor, about_ref), for EVERY class
//     (ruled by gblock 2026-11-06, F5 (a)) — with one exception, the gap a seat's own owed unread or
//     unsound check obliges (answersOwnFailedCheckAt): refused here, the lens would owe a gap the
//     record turns away. The exception admits that gap over
//     any open gap of its class, in either form below.
//
// A GAP ABOUT CITATION C AND A GAP QUOTING THE SENTENCE THAT HOLDS C ARE ONE SUBJECT per class (ruled
// by gblock 2026-11-06/10, F9: "if they were truly different we'd assign a different class"), so the
// rule has a cross-arm in both directions: a `--quote` mint is refused while an open gap of its class
// is about a citation whose token its sentence holds, and an `--about-kind anchor` mint is refused
// while an open quote gap of its class sits in the citation's sentence. Each refusal names the standing
// gap in the form it has. A sentence holding two citations: a quote gap on it is refused by an open
// gap about either; a gap about one citation is not refused by an open gap about the other. A
// computation's token pairs with nothing: a gap about a computation is the computation ruling's.
//
// A section is not such a subject: a section holds many sentences. The rule does not apply there.
//
// MEASURED (universe-m35): the parallel first-wave lenses minted five pairs, each pair 25 s to 2 min
// apart, each second minter having read the board before its sibling minted. Four of the five are
// one class on one subject. The check runs in the mint's own transaction, which takes the write lock
// at BEGIN (`_txlock=immediate`), so the second of two concurrent mints reads the first.
//
// A closed gap does not block: a defect that recurs after its gap closed is a new gap.
//
// ONE QUERY FEEDS EVERY ARM (openGapsOfClassAt): the open gaps of the class with their subject
// columns, and the arms fold over that one set.
//
// AN ACT JUDGEMENT (judgeAct): archived records hold such pairs, and an archived mint is what a seat
// did.
func requireOneGapPerClassAndSubject(w writeView, seatID string, m *recordpb.Mint) error {
	quote := strings.TrimSpace(m.GetLocation())
	inquiry, about, part, section, onGap := "", "", "", "", ""
	switch m.GetAboutKind() {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
		inquiry = m.GetAboutRef()
	// A SECTION AND A GAP ARE SUBJECTS ONLY OF A RULING'S TOMBSTONE (§III.B6): a section holds many
	// sentences, so two open gaps of a class on it are two defects; a ruling on one bars its class on
	// the section, keyed by the heading every reader of headings uses, to every minter. A section
	// assembly composes takes no bar (#1942). A gap on a gap is barred by a ruling on a gap of its
	// class about the same gap, keyed on that gap's id.
	case recordpb.AboutKind_ABOUT_KIND_SECTION:
		if !composedSection(m.GetAboutRef()) {
			section = reportheading.HeadingKey(m.GetAboutRef())
		}
	case recordpb.AboutKind_ABOUT_KIND_GAP:
		onGap = m.GetAboutRef()
	// A GAP ABOUT A QUESTION PART IS ONE PER CLASS AND PART; one about the question as a whole names no
	// part, and two of them may be about two parts the record does not hold, so the rule reads none.
	case recordpb.AboutKind_ABOUT_KIND_QUESTION:
		part = m.GetAboutRef()
	case recordpb.AboutKind_ABOUT_KIND_ANCHOR:
		about = m.GetAboutRef()
	}
	if quote == "" && inquiry == "" && about == "" && part == "" && section == "" && onGap == "" {
		return nil
	}
	held, err := openGapsOfClassAt(w.q, m.GetClass())
	if err != nil {
		return fmt.Errorf("the board's open gaps could not be read, so whether one of class %s already stands on this subject cannot be told: %w", recordpb.Word(m.GetClass()), err)
	}
	if held, err = withRuledGapsOfClassAt(w.q, held, m.GetClass()); err != nil {
		return fmt.Errorf("the record's ruling tombstones could not be read, so whether a ruling on a gap of class %s stands on this subject cannot be told: %w", recordpb.Word(m.GetClass()), err)
	}
	if len(held) == 0 {
		return nil
	}
	// The exception, read once and only where an arm finds a gap standing that it admits over.
	exempt, asked := false, false
	obliged := func() (bool, error) {
		if !asked {
			own, err := answersOwnFailedCheckAt(w.q, seatID, m)
			if err != nil {
				return false, feov.Wrap(feov.Conflict, err, "the record's checks could not be read, so whether this mint answers a failed check %s made cannot be told: %v", seatID, err)
			}
			exempt, asked = own != "", true
		}
		return exempt, nil
	}
	// refuse is the refusal of g on subject, but where g stands only by a ruling's tombstone and the
	// mint is the seat's owed failed check, which no tombstone refuses.
	refuse := func(g heldGap, subject string) error {
		if g.ruling != nil {
			if ok, err := obliged(); err != nil || ok {
				return err
			}
		}
		return oneGapRefusal(w, seatID, m, g, subject)
	}
	if inquiry != "" {
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_INQUIRY, inquiry) {
				return refuse(g, "inquiry "+inquiry)
			}
		}
	}
	if section != "" {
		for _, g := range held {
			if g.ruling != nil && g.aboutKind == recordpb.Word(recordpb.AboutKind_ABOUT_KIND_SECTION) && reportheading.HeadingKey(g.aboutRef) == section {
				return refuse(g, "section "+strconv.Quote(m.GetAboutRef()))
			}
		}
	}
	if onGap != "" {
		for _, g := range held {
			if g.ruling != nil && g.about(recordpb.AboutKind_ABOUT_KIND_GAP, onGap) {
				return refuse(g, "gap "+onGap)
			}
		}
	}
	if part != "" {
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_QUESTION, part) {
				return refuse(g, "question part "+part)
			}
		}
	}
	if about != "" {
		// The anchor arm.
		for _, g := range held {
			if !g.about(recordpb.AboutKind_ABOUT_KIND_ANCHOR, about) {
				continue
			}
			if ok, err := obliged(); err != nil || ok {
				return err
			}
			return oneGapRefusal(w, seatID, m, g, EvidenceNoun(about)+" "+about)
		}
		// The cross-arm, about over quote: an open quote gap of the class in the citation's sentence.
		if anchor.Kind(about) == "citation" {
			report, err := w.report()
			if err != nil {
				return fmt.Errorf("the report does not render, so the sentence citation %s sits in cannot be found to ask whether an open gap of class %s quotes it: %w", about, recordpb.Word(m.GetClass()), err)
			}
			spans := anchor.Sentences(report)
			if at, ok := gapSentenceIn(report, spans, about); ok {
				for _, g := range held {
					if g.location == "" {
						continue
					}
					if sp, ok := gapSentenceIn(report, spans, g.id); ok && sp == at {
						if ok, err := obliged(); err != nil || ok {
							return err
						}
						return oneGapRefusal(w, seatID, m, g, "the sentence "+sentenceWords(report, at)+", which holds citation "+about)
					}
				}
			}
		}
	}
	if quote == "" {
		return nil
	}
	ms, err := mintSentenceAt(w, m)
	if err != nil {
		return err
	}
	if !ms.ok {
		return nil
	}
	for _, g := range held {
		if g.location == "" {
			continue
		}
		if sp, ok := gapSentenceIn(ms.placed, ms.spans, g.id); ok && sp == ms.at {
			return refuse(g, "the sentence "+sentenceWords(ms.placed, ms.at))
		}
	}
	// The cross-arm, quote over about: an open gap of the class about a citation the sentence holds.
	for _, c := range anchor.IDs(ms.placed[ms.at[0]:ms.at[1]]) {
		if anchor.Kind(c) != "citation" {
			continue
		}
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_ANCHOR, c) {
				return refuse(g, "citation "+c+", which the sentence "+sentenceWords(ms.placed, ms.at)+" holds")
			}
		}
	}
	return nil
}

// heldGap is an open gap as the one-gap rule names it: its id, who minted it, what it says is wrong,
// and its subject columns — what it is about, and the quote it was placed by.
type heldGap struct {
	id, assignee, problem string
	aboutKind, aboutRef   string
	location              string
	// ruling is the standing ruling tombstone on the gap, nil for none: a closed gap of the class is
	// held by it alone, an open one by its openness and it (#1695 §III.B12, double jeopardy).
	ruling *RulingTombstone
}

// about reports whether g is about ref of kind.
func (g heldGap) about(kind recordpb.AboutKind, ref string) bool {
	return g.aboutKind == recordpb.Word(kind) && g.aboutRef == ref
}

// openGapsOfClassAt is the open gaps of class on the gap view, in mint order, with the subject
// columns every arm of the one-gap rule folds over.
func openGapsOfClassAt(q recordsql.Querier, class recordpb.GapClass) ([]heldGap, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(`SELECT "gap_id", COALESCE("minted_by", ''), COALESCE("current_problem", ''),
	           COALESCE("about_kind", ''), COALESCE("about_ref", ''), COALESCE("location", '')
	           FROM "gap" WHERE "open" AND "class" = ? ORDER BY "minted_seq"`, recordpb.Word(class))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []heldGap
	for rows.Next() {
		var g heldGap
		if err := rows.Scan(&g.id, &g.assignee, &g.problem, &g.aboutKind, &g.aboutRef, &g.location); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// withRuledGapsOfClassAt is held with every gap of class a standing ruling tombstone stands on: an
// open one already held is marked with it, a closed one joins the set by its own subject columns.
func withRuledGapsOfClassAt(q recordsql.Querier, held []heldGap, class recordpb.GapClass) ([]heldGap, error) {
	ruled, err := rulingTombstonesAt(q, ``)
	if err != nil {
		return nil, err
	}
	for i := range ruled {
		t := &ruled[i]
		if k := slices.IndexFunc(held, func(g heldGap) bool { return g.id == t.Gap }); k >= 0 {
			held[k].ruling = t
			continue
		}
		g := heldGap{id: t.Gap, ruling: t}
		found, err := queryRowAt(q, []any{&g.problem, &g.aboutKind, &g.aboutRef, &g.location},
			`SELECT COALESCE((SELECT a."result_problem" FROM "amend" a WHERE a."gap_id" = m."gap_id" ORDER BY a."event_id" DESC LIMIT 1), m."problem"),
			        COALESCE(m."about_kind", ''), COALESCE(m."about_ref", ''), COALESCE(m."location", '') FROM "mint" m WHERE m."gap_id" = ? AND m."class" = ?`,
			t.Gap, recordpb.Word(class))
		if err != nil {
			return nil, err
		}
		if found {
			held = append(held, g)
		}
	}
	return held, nil
}

// oneGapRefusal names the open gap g that holds the subject, and the refused seat's road from there.
// The road turns on whose gap g is: the seat's own gap is restated by amending it and reweighed by
// regrading it; another seat's gap already carries a point that agrees with it, and what the seat sees
// wrong in it — its printed wording or its grades — is the seat's gap on g, which g's raiser answers
// (plans/findings-fold.md §III.A2).
func oneGapRefusal(w writeView, seatID string, m *recordpb.Mint, g heldGap, subject string) error {
	if t := g.ruling; t != nil {
		return feov.Errorf(feov.Conflict,
			"the bench ruled gap %s %s on docketing %s (event %d); its tombstone stands on %s, so a gap of class %s there relitigates it. "+
				"%s ends it; blue's evidence belongs in the exchange before the bench rules. Nothing was written",
			t.Gap, t.Word, t.Docketing, t.Event, subject, recordpb.Word(m.GetClass()), capitalized(tombstoneEnds(g.aboutKind, g.aboutRef)))
	}
	if g.assignee == seatID {
		return feov.Errorf(feov.Conflict,
			"gap %s is yours and is open on %s with class %s: %q. Two defects of one class on one subject are one gap. "+
				"To restate it, amend it (`amend --id %s`); to reweigh it, regrade it (`regrade --id %s`). Nothing was written",
			g.id, subject, recordpb.Word(m.GetClass()), g.problem, g.id, g.id)
	}
	return feov.Errorf(feov.Conflict,
		"gap %s, raised by %s, is open on %s with class %s: %q. Two defects of one class on one subject are one gap. "+
			"If your point is %s's defect as %s states it, it is on the board: nothing more is recorded. "+
			"What you see wrong in %s — its problem, fix, check or weight — is a gap of yours about %s, which %s answers. "+
			"A defect that stands in the report after %s closes is a new gap of yours. Nothing was written",
		g.id, g.assignee, subject, recordpb.Word(m.GetClass()), g.problem, g.id, g.id, g.id, g.id, g.assignee, g.id)
}
