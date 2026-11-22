package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
// WHAT THE RECORD PROVES IS ANOTHER RELATION IS ADMITTED: a gap the mint names in --supersedes does
// not block it, since a successor states its ancestor's subject by design.
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
	inquiry, about, part := "", "", ""
	switch m.GetAboutKind() {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
		inquiry = m.GetAboutRef()
	// A GAP ABOUT A QUESTION PART IS ONE PER CLASS AND PART; one about the question as a whole names no
	// part, and two of them may be about two parts the record does not hold, so the rule reads none.
	case recordpb.AboutKind_ABOUT_KIND_QUESTION:
		part = m.GetAboutRef()
	case recordpb.AboutKind_ABOUT_KIND_ANCHOR:
		about = m.GetAboutRef()
	}
	if quote == "" && inquiry == "" && about == "" && part == "" {
		return nil
	}
	held, err := openGapsOfClassAt(w.q, m.GetClass())
	if err != nil {
		return fmt.Errorf("the board's open gaps could not be read, so whether one of class %s already stands on this subject cannot be told: %w", recordpb.Word(m.GetClass()), err)
	}
	held = slices.DeleteFunc(held, func(g heldGap) bool { return g.id == m.GetSupersedes() })
	if len(held) == 0 {
		return nil
	}
	if inquiry != "" {
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_INQUIRY, inquiry) {
				return oneGapRefusal(w, seatID, m, g, "inquiry "+inquiry)
			}
		}
	}
	if part != "" {
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_QUESTION, part) {
				return oneGapRefusal(w, seatID, m, g, "question part "+part)
			}
		}
	}
	// The exception, read once and only where an anchor-side arm finds a gap standing.
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
			spans, bare := anchor.Sentences(report), claimcount.BareAnchorIDs(report)
			if at, ok := gapSentenceIn(report, spans, bare, about); ok {
				for _, g := range held {
					if g.location == "" {
						continue
					}
					if sp, ok := gapSentenceIn(report, spans, bare, g.id); ok && sp == at {
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
		if sp, ok := gapSentenceIn(ms.placed, ms.spans, ms.bare, g.id); ok && sp == ms.at {
			return oneGapRefusal(w, seatID, m, g, "the sentence "+sentenceWords(ms.placed, ms.at))
		}
	}
	// The cross-arm, quote over about: an open gap of the class about a citation the sentence holds.
	for _, c := range anchor.IDs(ms.placed[ms.at[0]:ms.at[1]]) {
		if anchor.Kind(c) != "citation" {
			continue
		}
		for _, g := range held {
			if g.about(recordpb.AboutKind_ABOUT_KIND_ANCHOR, c) {
				return oneGapRefusal(w, seatID, m, g, "citation "+c+", which the sentence "+sentenceWords(ms.placed, ms.at)+" holds")
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
	rows, err := q.Query(`SELECT "gap_id", COALESCE("assignee", ''), COALESCE("problem", ''),
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

// oneGapRefusal names the open gap g that holds the subject, and the refused seat's road from there.
// The road turns on whose gap g is: the seat's own gap is restated by superseding it and reweighed by
// regrading it; another seat's gap already carries a point that agrees with it, and a dispute of it
// is the act the seat proposes on it, which g's assignee may accept and the bench otherwise rules.
// While g has had no exchange, the refusal says so, since a proposal is admitted only once it has.
func oneGapRefusal(w writeView, seatID string, m *recordpb.Mint, g heldGap, subject string) error {
	if g.assignee == seatID {
		return feov.Errorf(feov.Conflict,
			"gap %s is yours and is open on %s with class %s: %q. Two defects of one class on one subject are one gap. "+
				"To restate its defect, supersede it (`--supersedes %s`); to reweigh it, regrade it (`regrade --id %s`). Nothing was written",
			g.id, subject, recordpb.Word(m.GetClass()), g.problem, g.id, g.id)
	}
	exchanged, err := exchangesSentence(w.q, g.id)
	if err != nil {
		return err
	}
	return feov.Errorf(feov.Conflict,
		"gap %s, assigned to %s, is open on %s with class %s: %q. Two defects of one class on one subject are one gap. "+
			"If your point is %s's defect as %s states it, it is on the board: nothing more is recorded. "+
			"If you dispute %s — its problem, fix, check or weight — propose the act you want on it (`motion docket file supersede --id %s`, or regrade or close in place of supersede), argued in --reason; %s may accept it, else the bench rules.%s Nothing was written",
		g.id, g.assignee, subject, recordpb.Word(m.GetClass()), g.problem, g.id, g.id, g.id, g.id, g.assignee, exchanged)
}

// exchangesSentence is the sentence a road to a proposed act on gap id carries while id has had no
// exchange — " " and the sentence, "" once it has had one — read in the transaction q. The count is
// worded by GapExchanges.Counted, so a sitting the record cannot close reads as NOT MEASURED rather
// than as a zero.
func exchangesSentence(q recordsql.Querier, id string) (string, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return "", err
	}
	remands, err := remandRulingsOf(q)
	if err != nil {
		return "", err
	}
	x := exchangeFold(evs, win.IDs(evs), win, WhileRunning, remands)[id]
	if x == nil {
		x = &GapExchanges{GapID: id}
	}
	if x.Exchanges >= 1 {
		return "", nil
	}
	return " Gap " + id + " stands at " + x.Counted() + ": a proposal on it is admitted once it has had one exchange.", nil
}
