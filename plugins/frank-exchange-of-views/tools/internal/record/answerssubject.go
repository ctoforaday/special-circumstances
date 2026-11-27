package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE AUTHOR'S ANSWER NAMES A GAP WHOSE SUBJECT THE ACT CHANGES, AND ONLY THE AUTHOR GIVES IT
// (plans/findings-fold.md §III.B1, "The author's answer"). An inquiry fate and a question part carry
// `answers` the way an edit does, and the turn fold reads an unnamed one as the raiser's turn
// (lensTurnActsOf): a name the act's subject does not hold would answer a gap the act never touched,
// and an answer by a seat that is not the gap's author would spend the author's exchange for it.
//
// What each act changes is what fateActsOf reads: a fate changes its inquiry and the inquiry sections
// that print it — the one its previous record put it under and the one this record puts it under; a
// coverage changes its part; a split changes the question for a gap that names no part.

// requireInquiryAnswers is an inquiry record's `answers` (validate): a gap on the record, about the
// inquiry or a section that prints it before or after this record, whose author is seatID. A
// proposal changes nothing a gap is about, and answers none.
func requireInquiryAnswers(q recordsql.Querier, seatID string, b *recordpb.Inquiry) error {
	g := b.GetAnswers()
	if g == "" {
		return nil
	}
	if b.GetSupersedesStatus() == "" {
		return feov.Errorf(feov.Validation, "--answers names the gap a recorded fate answers — a proposal changes nothing a gap is about, so it answers none")
	}
	kind, ref, err := answeredSubject(q, seatID, g)
	if err != nil {
		return err
	}
	if kind == recordpb.AboutKind_ABOUT_KIND_INQUIRY && ref == b.GetInquiryId() {
		return nil
	}
	if kind == recordpb.AboutKind_ABOUT_KIND_SECTION {
		var prior string
		if _, err := queryRowAt(q, []any{&prior},
			`SELECT "status" FROM "inquiry" WHERE "inquiry_id" = ? ORDER BY "event_id" DESC LIMIT 1`, b.GetInquiryId()); err != nil {
			return err
		}
		key := reportheading.HeadingKey(ref)
		for _, st := range []string{prior, recordpb.Word(b.GetStatus())} {
			if st != "" && reportheading.HeadingKey(reportheading.InquirySectionOf(st)) == key {
				return nil
			}
		}
	}
	return feov.Errorf(feov.Validation, "--answers names %s, which is about %s — a fate of %s answers a gap about that inquiry, or about the inquiry section that prints it before or after this record; it changes nothing %s is about", g, subjectWords(kind, ref), b.GetInquiryId(), g)
}

// requireQuestionPartAnswers is a question part's `answers` (validate): a gap on the record whose
// author is seatID, about the part a coverage is recorded on, or about the question with no part for a
// part split out.
func requireQuestionPartAnswers(q recordsql.Querier, seatID string, b *recordpb.QuestionPart) error {
	g := b.GetAnswers()
	if g == "" {
		return nil
	}
	kind, ref, err := answeredSubject(q, seatID, g)
	if err != nil {
		return err
	}
	want, what := "", "a part split out answers a gap about the question that names no part"
	if b.GetSupersedesCoverage() != "" {
		want, what = b.GetPartId(), "a coverage of "+b.GetPartId()+" answers a gap about that part"
	}
	if kind == recordpb.AboutKind_ABOUT_KIND_QUESTION && ref == want {
		return nil
	}
	return feov.Errorf(feov.Validation, "--answers names %s, which is about %s — %s; it changes nothing %s is about", g, subjectWords(kind, ref), what, g)
}

// requireAnswersAuthor is an edit's, an evidence-anchor retire's, a citation's or a computation's
// `answers` (validate): a gap on the record whose author is seatID — the bench's application in its
// ruling's transaction aside. Which gaps' subjects such an act changes is not asked here.
func requireAnswersAuthor(q recordsql.Querier, seatID, g string) error {
	if g == "" {
		return nil
	}
	_, _, err := answeredSubject(q, seatID, g)
	return err
}

// answeredSubject is gap g's subject, read off its mint, once g is on the record and seatID is its
// author (authorOf) — the bench's application in its ruling's transaction aside.
func answeredSubject(q recordsql.Querier, seatID, g string) (recordpb.AboutKind, string, error) {
	if err := requireGap(q, g, "--answers"); err != nil {
		return 0, "", err
	}
	author, _, err := authorOfGap(q, g)
	if err != nil {
		return 0, "", err
	}
	if seatID != author && seatclass.RoleOfSeat(seatID) != "bench" {
		return 0, "", feov.Errorf(feov.RoleViolation, "--answers names %s, whose subject's author is %s — only a gap's author answers it, and this seat is %s", g, author, seatID)
	}
	var kind, ref string
	if _, err := queryRowAt(q, []any{&kind, &ref},
		`SELECT COALESCE("about_kind", ''), COALESCE("about_ref", '') FROM "mint" WHERE "gap_id" = ?`, g); err != nil {
		return 0, "", err
	}
	k, _ := AboutKindOf(kind)
	return k, ref, nil
}

// subjectWords is a gap's subject as a refusal names it.
func subjectWords(kind recordpb.AboutKind, ref string) string {
	switch {
	case kind == recordpb.AboutKind_ABOUT_KIND_UNSPECIFIED:
		return "a sentence of the report"
	case ref == "":
		return "the " + recordpb.Word(kind)
	}
	return fmt.Sprintf("the %s %q", recordpb.Word(kind), ref)
}
