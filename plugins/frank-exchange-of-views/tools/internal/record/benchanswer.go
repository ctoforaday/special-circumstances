package record

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE BENCH ANSWERS A DOCKETING BY NAMING A SIDE, AND THE TOOL APPLIES THAT SIDE'S STATEMENT (#1695,
// §III.B14). Each docketing puts two positions before the bench, and each side states its own in the
// closing it owes there: red's (G's assignee's) closing its replacement text, blue's closing its word.
// A held proposal is its filer's statement; G's grades are the statement against a regrade proposal;
// G's recorded closure the statement against a supersede of a closed G. The bench writes no word, grade
// or text of its own: its answer names a side — `--for-raiser | --for-author` — or `--keep | --restore` at grade movement, or remands; the tool
// writes the side's statement in the ruling's transaction, before the ruling, and validate holds the
// transaction to exactly those bytes (requireBenchAnswerApplied). A side wins only by what it stated:
// siding with the side whose statement is text, where its closing states none, is refused.

// Positions is what a docketing puts before the bench, as the record holds it at the ruling.
type Positions struct {
	ID, GapID, MotionID string
	Cause               recordpb.DocketingCause
	// Pair is the two sides a ruling may side with; zero at grade movement, whose answer is a grade
	// answer.
	Pair [2]recordpb.SidedWith
	// Act is the held proposal's stored act (Regrade, Close or Mint); nil with none.
	Act proto.Message
	// GapOpen is whether G is open; AboutKind what it is about ("" for a sentence).
	GapOpen   bool
	AboutKind string
	// AboutRef is the anchor G is about, on a gap about a citation or a computation: the road's retire
	// retires it.
	AboutRef string
	// TextSide is the side whose statement is text — red's at impasse and against a close or a
	// supersede on open G, the assignee's against another red seat's — zero where neither side's is.
	TextSide recordpb.SidedWith
	// Road is the closing of G's assignee filed for this docketing, nil where it filed none: its text,
	// or its no_text.
	Road *recordpb.Closing
	// BlueWord is the word blue's closing states at an impasse docketing, nil where it states none.
	BlueWord *recordpb.Disposition
	// Restoring is the regrade `--restore` writes at grade movement: each axis the window moved set back
	// to the grade it measured from; nil where no axis moved net.
	Restoring *recordpb.Regrade
	// Ruled is whether the bench has ruled the docketing.
	Ruled bool

	at int64 // the docketing's row
	// bench is the bench's acts on G written since the docketing's row: the statement a ruling applies,
	// written in its transaction before it.
	bench []benchAct
}

// HeldSide is the side a held proposal is: its author's, the one party that files on a gap; zero
// where the docketing holds none.
func (p Positions) HeldSide() recordpb.SidedWith {
	if p.Act == nil {
		return 0
	}
	return recordpb.SidedWith_SIDED_WITH_AUTHOR
}

// benchAct is one act of the bench's on a docketed gap since its docketing.
type benchAct struct {
	at   int64
	body proto.Message
}

// Stated reports whether side stated a position the tool can apply: every side but the text side
// states one by the record itself; the text side states one by a road in its closing — the pair, or the
// anchor's retire — on a gap about a sentence, a citation or a computation.
func (p Positions) Stated(side recordpb.SidedWith) bool {
	if side != p.TextSide || side == 0 {
		return true
	}
	return p.missingText() == ""
}

// missingText is why the text side states no text, "" where it does.
func (p Positions) missingText() string {
	switch p.AboutKind {
	case recordpb.Word(recordpb.AboutKind_ABOUT_KIND_INQUIRY), recordpb.Word(recordpb.AboutKind_ABOUT_KIND_SECTION):
		return fmt.Sprintf("gap %s is about %s %s, where no road of red's text exists", p.GapID, article(p.AboutKind), p.AboutKind)
	}
	switch {
	case p.Road == nil:
		return fmt.Sprintf("gap %s's assignee filed no closing on docketing %s, so it states no text", p.GapID, p.ID)
	case p.Road.GetNoText():
		return fmt.Sprintf("gap %s's assignee's closing on docketing %s states no text (--no-text)", p.GapID, p.ID)
	case p.Road.Location == nil && !p.Road.GetRetire():
		return fmt.Sprintf("gap %s's assignee's closing on docketing %s carries no road", p.GapID, p.ID)
	}
	return ""
}

// article is "an" before a word opening with a vowel, else "a".
func article(w string) string {
	if w != "" && (w[0] == 'a' || w[0] == 'e' || w[0] == 'i' || w[0] == 'o' || w[0] == 'u') {
		return "an"
	}
	return "a"
}

// PairWords is the docketing's answers in the flags' words: `--for-raiser | --for-author`, or
// `--keep | --restore` at grade movement.
func (p Positions) PairWords() string {
	if p.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT {
		return "--keep | --restore"
	}
	return "--for-" + recordpb.Word(p.Pair[0]) + " | --for-" + recordpb.Word(p.Pair[1])
}

// PositionsOf is positionsAt asked of run's record.
func PositionsOf(run Run, docketingID string) (p Positions, found bool, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		p, found, err = positionsAt(q, docketingID)
		return err
	})
	return p, found, err
}

// Positions is positionsAt asked of the record the write lands on.
func (l *Landing) Positions(docketingID string) (Positions, bool, error) {
	return positionsAt(l.w.q, docketingID)
}

// positionsAt is docketing id's positions on q's record.
func positionsAt(q recordsql.Querier, id string) (Positions, bool, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return Positions{}, false, err
	}
	ids := win.IDs(evs)
	var d *Docketing
	for _, k := range docketingsAt(evs, ids) {
		if k.ID == id {
			d = k
		}
	}
	if d == nil {
		return Positions{}, false, nil
	}
	p := Positions{ID: d.ID, GapID: d.GapID, MotionID: d.MotionID, Cause: d.Cause, Ruled: d.Ruled(), at: d.at}
	var about sql.NullString
	if _, err := queryRowAt(q, []any{&p.GapOpen, &about}, `SELECT "open", "about_kind" FROM "gap" WHERE "gap_id" = ?`, d.GapID); err != nil {
		return p, false, err
	}
	p.AboutKind = about.String
	var ref sql.NullString
	if _, err := queryRowAt(q, []any{&ref}, `SELECT "about_ref" FROM "mint" WHERE "gap_id" = ?`, d.GapID); err != nil {
		return p, false, err
	}
	p.AboutRef = ref.String
	from := d.at
	if d.MotionID != "" {
		dp, found, err := docketProposalAt(q, d.MotionID)
		if err != nil {
			return p, false, err
		}
		if found {
			p.Act = StoredAct(dp.Motion)
		}
		for _, m := range motionsAt(evs, ids, win) {
			if m != nil && m.ID == d.MotionID {
				from = m.filed
			}
		}
	}
	red, blue := recordpb.SidedWith_SIDED_WITH_RAISER, recordpb.SidedWith_SIDED_WITH_AUTHOR
	switch {
	case d.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT:
	default:
		p.Pair = [2]recordpb.SidedWith{red, blue}
	}
	// THE TEXT SIDE: red's (the assignee's, lens against lens) at impasse and against a close or a
	// supersede on open G. Against a regrade its statement is G's grades; on a closed G, G's closure.
	if p.GapOpen && p.Pair[0] != 0 {
		textual := d.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE
		switch p.Act.(type) {
		case *recordpb.Close, *recordpb.Mint:
			textual = true
		}
		if textual {
			p.TextSide = red
		}
	}
	// EACH PARTY'S STATEMENT IS ITS CLOSING FOR THIS DOCKETING: filed after its span opened — its row,
	// or the held proposal's filing (L30) — and before its ruling.
	assignee, author := assigneeOf(evs)[d.GapID], authorsOn(evs)(d.GapID)
	restoreTo := len(evs)
	for i, e := range evs {
		if ids[i] <= from || (d.ruled != 0 && ids[i] >= d.ruled) {
			continue
		}
		switch b := mustBody(e).(type) {
		case *recordpb.Closing:
			if b.GetGapId() != d.GapID {
				continue
			}
			switch e.GetSeatId() {
			case assignee:
				if p.Road == nil && seatclass.RoleOfSeat(assignee) == "lens" {
					p.Road = b
				}
			case author:
				if b.ClosureClass != nil && p.BlueWord == nil && d.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
					p.BlueWord = b.ClosureClass
				}
			}
		case *recordpb.BlueEdit, *recordpb.Retire, *recordpb.Regrade:
			if ids[i] <= d.at || seatclass.RoleOfSeat(e.GetSeatId()) != "bench" || benchActGap(b) != d.GapID {
				continue
			}
			p.bench = append(p.bench, benchAct{at: ids[i], body: b})
			if _, ok := b.(*recordpb.Regrade); ok && restoreTo == len(evs) {
				restoreTo = i
			}
		}
	}
	// THE RESTORE IS READ FROM THE WINDOW AS IT STOOD BEFORE THE BENCH'S OWN REGRADE: that regrade ends
	// the window on the axes it names (L41′), so a window read after it holds nothing to restore.
	if d.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT {
		gm := gradeMovementOf(evs[:restoreTo], ids[:restoreTo], win, map[string]bool{d.GapID: true})
		p.Restoring = gm.restoring(d.GapID, "")
	}
	return p, true, nil
}

// benchActGap is the gap a bench edit, retire or regrade is on: the gap an edit or a retire answers,
// the gap a regrade grades.
func benchActGap(b proto.Message) string {
	switch b := b.(type) {
	case *recordpb.BlueEdit:
		return b.GetAnswers()
	case *recordpb.Retire:
		return b.GetAnswers()
	case *recordpb.Regrade:
		return b.GetGapId()
	}
	return ""
}

// AnswerWord is the closing word a ruling for side writes, nil where it writes none: `repaired` for the
// text side, whose text the tool applies; blue's closing's word at impasse, or `not_a_defect` where it
// states none — blue's text stands; the stored word of a held close for the filer's side. A side whose
// statement is a regrade, a successor, or G's grades or closure writes none.
func (p Positions) AnswerWord(side recordpb.SidedWith) *recordpb.Disposition {
	switch {
	case side == 0:
		return nil
	case side == p.TextSide:
		return recordpb.Disposition_DISPOSITION_REPAIRED.Enum()
	case side == recordpb.SidedWith_SIDED_WITH_AUTHOR && p.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE:
		if p.BlueWord != nil {
			return p.BlueWord.Enum()
		}
		return recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
	case side == p.HeldSide():
		if c, ok := p.Act.(*recordpb.Close); ok {
			return c.GetClosureClass().Enum()
		}
	}
	return nil
}

// requireBenchAnswerApplied is validate's check of a bench ruling's bytes against its docketing's
// positions (§III.B14): a side that stated nothing is refused, naming the missing statement; the
// closing word is the one the side's statement gives (AnswerWord); and the bench's acts on G since the
// docketing's row are exactly the statement the answer applies — the edit or retire of red's road for
// the text side, re-planned on the report before it; the stored regrade for the filer's side of a
// regrade proposal; the restoring regrade on --restore; none for every other answer. So a bench
// regrade, edit or retire written on G with no ruling applying it is refused at the ruling.
func requireBenchAnswerApplied(q recordsql.Querier, r *recordpb.MotionRule) error {
	d := r.GetDocket()
	p, found, err := positionsAt(q, d.GetDocketingId())
	if err != nil || !found {
		return err
	}
	side := d.GetSidedWith()
	if side != 0 && !p.Stated(side) {
		return feov.Errorf(feov.Validation,
			"the bench sides with a party by what it stated, and %s: --for-%s is refused. Answer for the other side, or remand. Nothing was recorded",
			p.missingText(), recordpb.Word(side))
	}
	remand := d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED
	if !remand {
		if want := p.AnswerWord(side); (want == nil) != (d.Disposition == nil) || (want != nil && *want != d.GetDisposition()) {
			return feov.Errorf(feov.Validation, "%s. Nothing was recorded", p.wordRefusal(side, want, d.Disposition))
		}
	}
	var wantRegrade *recordpb.Regrade
	wantText := false
	switch {
	case remand, d.GetMovedGrades() == recordpb.MovedGrades_MOVED_GRADES_KEPT:
	case d.GetMovedGrades() == recordpb.MovedGrades_MOVED_GRADES_RESTORED:
		wantRegrade = p.Restoring
	case side == p.TextSide:
		wantText = true
	case side == p.HeldSide():
		if rg, ok := p.Act.(*recordpb.Regrade); ok {
			wantRegrade = rg
		}
	}
	var regrades, texts []benchAct
	for _, a := range p.bench {
		if _, ok := a.body.(*recordpb.Regrade); ok {
			regrades = append(regrades, a)
		} else {
			texts = append(texts, a)
		}
	}
	// WHO SITS AFTER: a raiser's statement applied in this transaction, and a remand, each need the
	// sitting its reason names (lensSitsAfter, partiesSitAfter).
	if err := requireSitsAfter(q, p, d, len(texts) > 0); err != nil {
		return err
	}
	switch {
	case wantRegrade == nil && len(regrades) > 0:
		return feov.Errorf(feov.Validation, "the bench writes no grade of its own: a bench regrade of %s at event %d stands since docketing %s, and this answer applies none. Nothing was recorded", p.GapID, regrades[0].at, p.ID)
	case wantRegrade != nil && (len(regrades) != 1 || !sameAxes(regrades[0].body.(*recordpb.Regrade), wantRegrade)):
		return feov.Errorf(feov.Validation, "this answer applies %s's grades as %s — exactly those axes at those grades, written as the bench's regrade before the ruling, and nothing else. Nothing was recorded", p.GapID, regradeWords(wantRegrade))
	case !wantText && len(texts) > 0:
		return feov.Errorf(feov.Validation, "the bench writes no text of its own: a bench edit or retire answering %s at event %d stands since docketing %s, and this answer applies none. Nothing was recorded", p.GapID, texts[0].at, p.ID)
	case wantText:
		return requireRoadApplied(q, p, texts)
	}
	return nil
}

// requireRoadApplied is the for-raiser answer's check that the bench's act is red's road:
// the anchor's retire where the road is a retire, else the edit blue's `edit --accept` makes of the
// pair, its span the one record.ProposalOver plans on the report as it stood before the edit. Where
// red's text already stands — the anchor retired, or the report holding the pair's replacement at the
// gap's place — the answer applies no act.
func requireRoadApplied(q recordsql.Querier, p Positions, texts []benchAct) error {
	// A RETIRE ROAD ALREADY DONE: the anchor's retire matches red's text, and the freeze admits it, so
	// the answer for red applies nothing more.
	if len(texts) == 0 && p.Road.GetRetire() {
		_, retired, err := anchorRetiredAt(q, p.AboutRef)
		if err != nil || retired {
			return err
		}
	}
	// A PAIR ROAD ALREADY DONE: the report holds red's text at the gap's place, as an edit of blue's that
	// matched it left it, so the answer for red applies nothing more.
	if len(texts) == 0 && !p.Road.GetRetire() && p.Road.GetFixNew() != "" {
		report, err := RenderAt(q)
		if err != nil {
			return err
		}
		if done, err := redTextAppliedAt(q, report, p.GapID, p.Road.GetLocation(), p.Road.GetFixNew()); err != nil || done {
			return err
		}
	}
	if len(texts) != 1 {
		return feov.Errorf(feov.Validation, "--for-%s applies red's text on %s as one bench act before the ruling — the edit or the retire its closing states — and the record holds %d. Nothing was recorded", recordpb.Word(p.TextSide), p.GapID, len(texts))
	}
	a := texts[0]
	if p.Road.GetRetire() {
		rt, ok := a.body.(*recordpb.Retire)
		if !ok || !rt.GetClaimKept() || len(rt.GetAnchors()) != 1 {
			return feov.Errorf(feov.Validation, "red's closing on %s states the anchor's retire, and the bench's act at event %d is not that retire. Nothing was recorded", p.GapID, a.at)
		}
		if rt.GetAnchors()[0] != p.AboutRef {
			return feov.Errorf(feov.Validation, "red's closing on %s retires %s, the anchor the gap is about; the bench's retire at event %d takes %s. Nothing was recorded", p.GapID, p.AboutRef, a.at, rt.GetAnchors()[0])
		}
		return nil
	}
	ed, ok := a.body.(*recordpb.BlueEdit)
	if !ok {
		return feov.Errorf(feov.Validation, "red's closing on %s states a replacement, and the bench's act at event %d is no edit. Nothing was recorded", p.GapID, a.at)
	}
	// THE ROAD IS RE-CHECKED AGAINST THE ESTOPPELS STANDING AT THE RULING (L43), on the stream before the
	// bench's own edit, which applies the road verbatim.
	fam, err := familyAt(q)
	if err != nil {
		return err
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	ids := win.IDs(evs)
	n := 0
	for n < len(evs) && ids[n] < a.at {
		n++
	}
	if err := roadEstoppel(fam, evs[:n], p.GapID, p.Road.GetLocation()); err != nil {
		return err
	}
	// THE TEXT BEFORE THE EDIT IS ITS OWN STAMPED PARAGRAPHS (stampedParagraphs): the edit's write
	// records the paragraphs it changed, so nothing is rebuilt as the report stood before it.
	before, _, err := stampedParagraphs(editAt{id: a.at, ed: ed})
	if err != nil {
		return err
	}
	old, _ := ProposalOver(before, p.GapID, p.Road.GetLocation(), p.Road.GetFixNew())
	if ed.GetOld() != old || !ed.GetAppliedVerbatim() {
		return feov.Errorf(feov.Validation, "the bench's edit at event %d is not red's text on %s as blue's `edit --accept` applies it: the span red's closing quotes, replaced by its text. Nothing was recorded", a.at, p.GapID)
	}
	return nil
}

// sameAxes reports whether a and b name the same axes at the same grades.
func sameAxes(a, b *recordpb.Regrade) bool {
	for k, ga := range regradeAxes(a) {
		gb := regradeAxes(b)[k]
		if (ga == nil) != (gb == nil) || (ga != nil && *ga != *gb) {
			return false
		}
	}
	return true
}

// regradeWords is a regrade's axes in words.
func regradeWords(r *recordpb.Regrade) string {
	s := ""
	for k, g := range regradeAxes(r) {
		if g == nil {
			continue
		}
		if s != "" {
			s += ", "
		}
		s += recordpb.Word(gradeAxes[k]) + " " + recordpb.Word(*g)
	}
	return s
}

// wordRefusal is the refusal of a ruling whose closing word is not the one its side's statement gives.
func (p Positions) wordRefusal(side recordpb.SidedWith, want, got *recordpb.Disposition) string {
	if want == nil {
		return fmt.Sprintf("--for-%s on docketing %s writes no closing word: that side's statement is a grade, a successor, or %s as it stands, and the ruling carries %s", recordpb.Word(side), p.ID, p.GapID, wordOr(got))
	}
	return fmt.Sprintf("--for-%s on docketing %s writes %s, the word that side stated; the ruling carries %q", recordpb.Word(side), p.ID, wordOr(want), wordOr(got))
}

// wordOr is a closing word's spelling, "" for none.
func wordOr(d *recordpb.Disposition) string {
	if d == nil {
		return ""
	}
	return recordpb.Word(d)
}

// AboutAnchor is the anchor gapID is about on the record the write lands on: a citation or a
// computation, "" for a gap about anything else.
func (l *Landing) AboutAnchor(gapID string) (string, error) {
	var kind, ref sql.NullString
	if _, err := queryRowAt(l.w.q, []any{&kind, &ref}, `SELECT "about_kind", "about_ref" FROM "mint" WHERE "gap_id" = ?`, gapID); err != nil {
		return "", err
	}
	if kind.String != recordpb.Word(recordpb.AboutKind_ABOUT_KIND_ANCHOR) {
		return "", nil
	}
	return ref.String, nil
}
