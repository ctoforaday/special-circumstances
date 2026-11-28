package record

import (
	"database/sql"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A DOCKETING PUTS A GAP BEFORE THE BENCH (#1695, §III.B13). The tool writes one — at impasse, for the
// board's grade movement, or for a proposal its assignee did not accept in its window — and no verb
// does. The bench's ruling names it (DocketRuling.docketing_id), once; a non-carry Close of the gap
// written after it ends it unanswered. A docketing holds a proposal exactly when its cause is PROPOSAL.
//
// TWO STATEMENTS OF ONE RULE: gap_state.standing_docketing (recordsql.GapRefreshViewDDL) and
// docketingsAt below. Every duty reader of a standing docketing reads one of them, and
// TestTheStoredDocketingIsTheFold holds the two level after every step.

// Docketing is one docketing after replay: the gap before the bench, why, the proposal it holds, and
// how it ended.
type Docketing struct {
	ID       string
	GapID    string
	Cause    recordpb.DocketingCause
	MotionID string // the proposal a PROPOSAL docketing holds; "" for every other cause
	Basis    string

	// Ruling is the bench's ruling naming the docketing, nil until ruled; RulingSeat its seat.
	Ruling     *recordpb.MotionRule
	RulingSeat string

	// at, ruled and ended are places on the stream (seq, or the stream position where none is given):
	// the docketing's row, its ruling's, and the non-carry Close of its gap written after it that ended
	// it unruled (zero for none).
	at, ruled, ended int64
}

// Ruled reports whether the bench has ruled the docketing.
func (d *Docketing) Ruled() bool { return d.Ruling != nil }

// Standing reports whether the docketing is before the bench: unruled, and no Close of its gap ended
// it.
func (d *Docketing) Standing() bool { return d.Ruling == nil && d.ended == 0 }

// Word is the ruling's word (rulingWord): its closing word, else its side or grade answer; "" unruled.
func (d *Docketing) Word() string {
	if d.Ruling == nil {
		return ""
	}
	return rulingWord(d.Ruling)
}

// appliesProposal reports whether the ruling sided with the proposal the docketing holds: the
// author's, whose statement is the proposal's act.
func (d *Docketing) appliesProposal() bool {
	return d.MotionID != "" && d.Docket().GetSidedWith() == recordpb.SidedWith_SIDED_WITH_AUTHOR
}

// movesGap reports whether the ruling's transaction moved its gap: a closing word or a remand, the
// restored grades, or the act of the proposal it sided with. A ruling for the side whose position is
// the gap as it stands, and grades kept, move nothing.
func (d *Docketing) movesGap() bool {
	r := d.Docket()
	if r == nil {
		return false
	}
	return r.Disposition != nil || r.GetMovedGrades() == recordpb.MovedGrades_MOVED_GRADES_RESTORED || d.appliesProposal()
}

// Docket is the ruling's docket arm, nil unruled.
func (d *Docketing) Docket() *recordpb.DocketRuling { return d.Ruling.GetDocket() }

// DocketingsOf is docketingsAt over the events themselves, places by stream position.
func DocketingsOf(evs []*Event) []*Docketing { return docketingsAt(evs, nil) }

// concession is what one act ends with no ruling (recordsql.concededSQL's twin): the docketing and
// the proposal standing on gap, ended by the gap's non-carry close, the author's act answering it with
// accepted, or an amend of it (which ends no grade-movement docketing); and the docketing holding
// motion, which the raiser's act takes (accepts).
type concession struct {
	gap, motion string
	amend       bool
}

// concessions reads each act's concession off the stream, in order: f is called with the act's place
// in evs and what it ends. The author of a gap's subject is authorOf's, read off the stream's mints.
func concessions(evs []*Event, f func(i int, c concession)) {
	// A gap is minted before any act on it, so the authors read off the whole stream are the ones in
	// force at each act.
	authorSeat := authorsOn(evs)
	authorAccept := func(e *Event, answers string, accepted bool) bool {
		return accepted && answers != "" && e.GetSeatId() == authorSeat(answers)
	}
	for i, e := range evs {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Close:
			if b.GetCarriedFrom() == "" {
				f(i, concession{gap: b.GetGapId()})
			}
		case *recordpb.Regrade:
			if b.GetAccepts() != "" {
				f(i, concession{motion: b.GetAccepts()})
			}
			if authorAccept(e, b.GetAnswers(), b.GetAccepted()) {
				f(i, concession{gap: b.GetAnswers()})
			}
		case *recordpb.Amend:
			f(i, concession{gap: b.GetGapId(), amend: true})
			if authorAccept(e, b.GetAnswers(), b.GetAccepted()) {
				f(i, concession{gap: b.GetAnswers()})
			}
		case *recordpb.BlueEdit:
			if authorAccept(e, b.GetAnswers(), b.GetAccepted()) {
				f(i, concession{gap: b.GetAnswers()})
			}
		case *recordpb.Retire:
			if authorAccept(e, b.GetAnswers(), b.GetAccepted()) {
				f(i, concession{gap: b.GetAnswers()})
			}
		}
	}
}

// docketingsAt is every docketing on the stream, in the order written, with the bench's ruling naming
// each and the act that ended each unruled. seq is evs's places (events.id), aligned with it; nil
// places by stream position. A ruling names its docketing by id (one per docketing: the column is
// UNIQUE), so the first is the one that stands. An act ends a docketing when it is written after the
// docketing's row (concessions): a Close when it is no carry (L42: a carry restates an earlier
// epoch's closure and answers nothing); a party's concession; an amend of its gap, but at grade
// movement, whose question is the grades.
func docketingsAt(evs []*Event, seq []int64) []*Docketing {
	byID := map[string]*Docketing{}
	var order []*Docketing
	byGap := map[string][]*Docketing{}
	byMotion := map[string][]*Docketing{}
	place := func(i int) int64 {
		if seq != nil {
			return seq[i]
		}
		return int64(i + 1)
	}
	for i, e := range evs {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Docketing:
			if b.GetDocketingId() == "" || byID[b.GetDocketingId()] != nil {
				continue
			}
			d := &Docketing{ID: b.GetDocketingId(), GapID: b.GetGapId(), Cause: b.GetCause(), MotionID: b.GetMotionId(),
				Basis: b.GetBasis(), at: place(i)}
			byID[d.ID] = d
			order = append(order, d)
			byGap[d.GapID] = append(byGap[d.GapID], d)
			if d.MotionID != "" {
				byMotion[d.MotionID] = append(byMotion[d.MotionID], d)
			}
		case *recordpb.MotionRule:
			id := b.GetDocket().GetDocketingId()
			if d := byID[id]; d != nil && d.Ruling == nil {
				d.Ruling, d.RulingSeat, d.ruled = b, e.GetSeatId(), place(i)
			}
		}
	}
	concessions(evs, func(i int, c concession) {
		at := place(i)
		end := func(ds []*Docketing) {
			for _, d := range ds {
				if (d.Ruling == nil || d.ruled > at) && d.ended == 0 && d.at < at &&
					!(c.amend && d.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT) {
					d.ended = at
				}
			}
		}
		if c.gap != "" {
			end(byGap[c.gap])
		}
		if c.motion != "" {
			end(byMotion[c.motion])
		}
	})
	return order
}

// standingDocketings is each gap's standing docketing (Docketing.Standing): the newest, where a gap
// holds more than one — gap_state.standing_docketing's rule.
func standingDocketings(ds []*Docketing) map[string]*Docketing {
	out := map[string]*Docketing{}
	for _, d := range ds {
		if d.Standing() {
			out[d.GapID] = d
		}
	}
	return out
}

// proposalStates is, per proposal (a docket motion) on the stream: held, where a standing docketing
// holds it — the docketed arm's, not its own; and ended, where a non-carry Close of its gap was written
// after its filing — answered by nobody, holding nothing (gap_state.unruled_docket_filed's rule). ids is
// evs's places, aligned with it.
func proposalStates(evs []*Event, ids []int64, win WindowIndex) (held, ended map[string]bool) {
	held, ended = map[string]bool{}, map[string]bool{}
	for _, d := range docketingsAt(evs, ids) {
		if d.MotionID != "" && d.Standing() {
			held[d.MotionID] = true
		}
	}
	byGap, byMotion := map[string][]int64{}, map[string][]int64{}
	concessions(evs, func(i int, c concession) {
		byGap[c.gap] = append(byGap[c.gap], ids[i])
		byMotion[c.motion] = append(byMotion[c.motion], ids[i])
	})
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != "docket" {
			continue
		}
		for _, at := range append(byGap[m.GapID], byMotion[m.ID]...) {
			if at > m.filed {
				ended[m.ID] = true
			}
		}
	}
	return held, ended
}

// docketStandingProposals writes Docketing{PROPOSAL} for each proposal standing with no docketing holding
// it (L29), through the write path's validation, as seatID — the bench's register at the terminal
// sitting or the assembly, where no plan follows to end the proposal's window.
func docketStandingProposals(tx *sql.Tx, files runFiles, seatID string) error {
	evs, win, err := eventsAt(tx)
	if err != nil {
		return err
	}
	ids := win.IDs(evs)
	held, ended := proposalStates(evs, ids, win)
	assignees := assigneeOf(evs)
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != "docket" || m.Ruled() || held[m.ID] || ended[m.ID] {
			continue
		}
		body := &recordpb.Docketing{
			DocketingId: proto.String(NewID("docketing")),
			GapId:       proto.String(m.GapID),
			Cause:       recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL.Enum(),
			MotionId:    proto.String(m.ID),
			Basis:       proto.String(proposalBasis(m, m.GapID, assignees[m.GapID])),
		}
		ev := &Event{}
		typ, err := recordpb.SetBody(ev, body)
		if err != nil {
			return err
		}
		if _, err := writeIn(tx, ev, seatID, typ, body, func(*sql.Tx) error {
			return validateAct(writeView{q: tx, files: files}, seatID, typ, body)
		}, stamp); err != nil {
			return fmt.Errorf("docketing proposal %s at the exit: %w", m.ID, err)
		}
	}
	return nil
}

// DocketRulingGap is the gap of a docket ruling: its docketing's gap where it names one (every bench
// ruling), else the gap its motion's filing names (the assignee's accept). ONE FUNCTION for every
// reader of a docket ruling's gap — the supersession fold, the views, the report, capture — over the
// stream the caller holds; a stream without the DOCKETING and MOTION events resolves every ruling to
// "", so every caller's typed read asks for both alongside MOTION_RULE.
func DocketRulingGap(evs []*Event) func(*recordpb.MotionRule) string {
	byMotion, byDocketing := map[string]string{}, map[string]string{}
	for _, e := range evs {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Motion:
			if d, isDocket := b.GetFiling().(*recordpb.Motion_Docket); isDocket && b.GetMotionId() != "" {
				byMotion[b.GetMotionId()] = d.Docket.GetGapId()
			}
		case *recordpb.Docketing:
			if b.GetDocketingId() != "" {
				byDocketing[b.GetDocketingId()] = b.GetGapId()
			}
		}
	}
	return func(r *recordpb.MotionRule) string {
		if id := r.GetDocket().GetDocketingId(); id != "" {
			return byDocketing[id]
		}
		return byMotion[r.GetMotionId()]
	}
}

// docketingRow is a docketing as the write checks read it.
type docketingRow struct {
	ID, GapID, MotionID string
	Cause               recordpb.DocketingCause
	EventID             int64
}

// docketingAt is the docketing id on q's record.
func docketingAt(q recordsql.Querier, id string) (docketingRow, bool, error) {
	var r docketingRow
	var cause string
	var motion sql.NullString
	found, err := queryRowAt(q, []any{&r.EventID, &r.GapID, &cause, &motion},
		`SELECT "event_id", "gap_id", "cause", "motion_id" FROM "docketing" WHERE "docketing_id" = ?`, id)
	if err != nil || !found {
		return r, found, err
	}
	r.ID, r.MotionID = id, motion.String
	if vd, ok := recordpb.BySpelling(recordpb.DocketingCause(0).Descriptor(), cause); ok {
		r.Cause = recordpb.DocketingCause(vd.Number())
	}
	return r, true, nil
}

// standingOnGap is the gap's standing docketing and standing proposal, as gap_state holds them ("" for
// none).
func standingOnGap(q recordsql.Querier, gapID string) (docketing, proposal string, err error) {
	var d, p sql.NullString
	if _, err := queryRowAt(q, []any{&d, &p},
		`SELECT "standing_docketing", "unruled_docket_motion" FROM "gap_state" WHERE "gap_id" = ?`, gapID); err != nil {
		return "", "", err
	}
	return d.String, p.String, nil
}

// requireDocketingShape is validate's arm for a Docketing: its shape and L28, and nothing more. A
// docketing is written by the tool, and by no verb; the arm cannot tell the dispatch's row from another
// writer's, and does not try. The CHECK holds the cause and the proposal together; the refusal here is
// the one in words. The gap is a gap; a PROPOSAL docketing holds the proposal standing on it; and no
// proposal stands on the gap of any other docketing (L28: a proposal in its assignee's window
// outranks impasse and grade movement).
func requireDocketingShape(q recordsql.Querier, d *recordpb.Docketing) error {
	if err := requireGap(q, d.GetGapId(), "gap_id"); err != nil {
		return err
	}
	proposalCause := d.GetCause() == recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL
	if proposalCause != (d.GetMotionId() != "") {
		return fmt.Errorf("a docketing holds a proposal exactly when its cause is the proposal its assignee did not accept: cause %s with motion %q", recordpb.Word(d.GetCause()), d.GetMotionId())
	}
	_, proposal, err := standingOnGap(q, d.GetGapId())
	if err != nil {
		return err
	}
	switch {
	case proposalCause && d.GetMotionId() != proposal:
		return fmt.Errorf("a proposal docketing on %s holds the proposal standing on it; motion %s is no docket motion standing unanswered on that gap", d.GetGapId(), d.GetMotionId())
	case !proposalCause && proposal != "":
		return fmt.Errorf("proposal %s stands on %s in its assignee's window, and it outranks the %s: the gap is docketed for the proposal once the window ends, and for nothing else while it stands", proposal, d.GetGapId(), recordpb.Word(d.GetCause()))
	}
	return nil
}

// docketingPair is the two sides a docketing's ruling may side with: the raiser's and the author's at
// impasse and on a proposal; none at grade movement, whose answer is a grade answer.
func docketingPair(q recordsql.Querier, k docketingRow) ([2]recordpb.SidedWith, error) {
	pair := [2]recordpb.SidedWith{recordpb.SidedWith_SIDED_WITH_RAISER, recordpb.SidedWith_SIDED_WITH_AUTHOR}
	if k.Cause != recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL {
		return pair, nil
	}
	if _, found, err := docketProposalAt(q, k.MotionID); err != nil || !found {
		return [2]recordpb.SidedWith{}, err
	}
	return pair, nil
}

// requireDocketAnswer is a docket ruling against what it answers (validate): the bench's alone, naming
// the docketing it answers, with exactly one of a side, a grade answer and a remand (the CHECK's rule,
// in words), from the docketing's pair.
func requireDocketAnswer(q recordsql.Querier, seatID string, r *recordpb.MotionRule) error {
	d := r.GetDocket()
	if d == nil {
		return nil
	}
	n := 0
	if d.SidedWith != nil {
		n++
	}
	if d.MovedGrades != nil {
		n++
	}
	if d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
		n++
	}
	if n != 1 {
		return feov.Errorf(feov.Validation, "a docket ruling answers with exactly one of a side of its docketing's pair, a grade answer (--keep or --restore) at grade movement, or a remand. Nothing was recorded")
	}
	if seatclass.RoleOfSeat(seatID) == "bench" {
		return requireBenchDocketAnswer(q, r)
	}
	// A DOCKET RULING IS THE BENCH'S. The raiser takes the author's proposal by its own act — the
	// regrade or close it proposes, carrying accepts — never by a ruling.
	return feov.Errorf(feov.RoleViolation,
		"a docket ruling is the bench's: the raiser of motion %s's gap takes the proposal by its own act — the regrade or close it proposes, accepting it — and a docketing is answered by the bench alone. Nothing was recorded",
		r.GetMotionId())
}

// requireBenchDocketAnswer is the bench's docket ruling against its docketing: it names a docketing on
// the record, the proposal the docketing holds and no other, and an answer from the docketing's cause —
// a grade answer at grade movement and nowhere else, no side and no remand there (L40), a side of the
// docketing's pair elsewhere.
func requireBenchDocketAnswer(q recordsql.Querier, r *recordpb.MotionRule) error {
	d := r.GetDocket()
	if d.GetDocketingId() == "" {
		return feov.Errorf(feov.MissingField, "a bench docket ruling names the docketing it answers. Nothing was recorded")
	}
	k, found, err := docketingAt(q, d.GetDocketingId())
	if err != nil {
		return err
	}
	if !found {
		return feov.Errorf(feov.Validation, "names docketing %s, which no docketing on the record carries. Nothing was recorded", d.GetDocketingId())
	}
	// A DOCKETING IS RULED ONCE: its second answer is refused by the first-wins guard (requireUnanswered),
	// under the write lock, before this check is reached.
	if k.MotionID != r.GetMotionId() {
		if k.MotionID == "" {
			return feov.Errorf(feov.Validation, "docketing %s holds no proposal, so its ruling names no motion; it names %s. Nothing was recorded", k.ID, r.GetMotionId())
		}
		return feov.Errorf(feov.Validation, "docketing %s holds proposal %s, and its ruling names it. Nothing was recorded", k.ID, k.MotionID)
	}
	movement := k.Cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT
	switch {
	case movement && d.MovedGrades == nil:
		return feov.Errorf(feov.Validation, "docketing %s is the board's grade movement: its ruling keeps or restores the moved grades (--keep | --restore), and neither sides with a party nor remands. Nothing was recorded", k.ID)
	case !movement && d.MovedGrades != nil:
		return feov.Errorf(feov.Validation, "docketing %s is no grade movement: --keep and --restore answer a grade-movement docketing only. Nothing was recorded", k.ID)
	case d.SidedWith != nil:
		pair, err := docketingPair(q, k)
		if err != nil {
			return err
		}
		if s := d.GetSidedWith(); s != pair[0] && s != pair[1] {
			return feov.Errorf(feov.Validation, "docketing %s is between %s and %s, and its ruling sides with one of them (--for-%s | --for-%s), not %s. Nothing was recorded",
				k.ID, recordpb.Word(pair[0]), recordpb.Word(pair[1]), recordpb.Word(pair[0]), recordpb.Word(pair[1]), recordpb.Word(s))
		}
	}
	// THE ANSWER'S BYTES ARE ITS SIDE'S STATEMENT, and nothing the bench wrote of its own.
	return requireBenchAnswerApplied(q, r)
}

// docketRulingGapAt is the gap of the docketing a bench docket ruling answers on q's record
// (DocketRulingGap's rule for the bench); "" for none.
func docketRulingGapAt(q recordsql.Querier, r *recordpb.MotionRule) (string, error) {
	k, _, err := docketingAt(q, r.GetDocket().GetDocketingId())
	return k.GapID, err
}

// ClosingTakesText is closingTakesText asked of the run's record: whether red's closing on gapID
// states a road or that it states none.
func ClosingTakesText(run Run, gapID string) (takes bool, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		takes, _, err = closingTakesText(q, gapID)
		return err
	})
	return takes, err
}

// closingTakesText reports whether red's closing on gapID states text (L37, L49): the gap is open and
// about a sentence, a citation or a computation, and before the bench at impasse, or for a close or a
// supersede a party proposed on it — standing in its assignee's window, or held by its docketing. On a
// regrade proposal red's statement is the gap's grades, at grade movement the recorded grades, and a
// gap about an inquiry or a section has no road (F-T2 (a)); a closed gap takes none (L46).
func closingTakesText(q recordsql.Querier, gapID string) (takes bool, aboutKind string, err error) {
	var open bool
	var about, docketing, proposal sql.NullString
	found, err := queryRowAt(q, []any{&open, &about, &docketing, &proposal},
		`SELECT "open", "about_kind", "standing_docketing", "unruled_docket_motion" FROM "gap" WHERE "gap_id" = ?`, gapID)
	if err != nil || !found || !open {
		return false, about.String, err
	}
	switch about.String {
	case recordpb.Word(recordpb.AboutKind_ABOUT_KIND_INQUIRY), recordpb.Word(recordpb.AboutKind_ABOUT_KIND_SECTION):
		return false, about.String, nil
	}
	motion := proposal.String
	if docketing.String != "" {
		k, _, err := docketingAt(q, docketing.String)
		if err != nil {
			return false, about.String, err
		}
		switch k.Cause {
		case recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE:
			return true, about.String, nil
		case recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT:
			return false, about.String, nil
		}
		motion = k.MotionID
	}
	if motion == "" {
		return false, about.String, nil
	}
	var act sql.NullString
	if _, err := queryRowAt(q, []any{&act}, `SELECT "act_case" FROM "motion" WHERE "motion_id" = ?`, motion); err != nil {
		return false, about.String, err
	}
	return act.String == "close" || act.String == "supersede", about.String, nil
}

// requireClosingStatement is validate's arm for what a closing states (§III.B14): red's text, blue's
// word, each on its own side's closing and nowhere else. Blue's closing carries no road, and states a
// word only at an impasse docketing — one of the five words a for-author answer writes. Red's closing
// carries no word; where its docketing's question takes text (closingTakesText) it carries exactly one
// of a road — the pair, or the anchor's retire on a gap about a citation or computation — and
// no_text, so an omitted road is never recorded as a statement of none; elsewhere it carries none.
// Whether the pair plans on the report, and whether a standing estoppel row refuses it, is the closing
// verb's and the for-raiser answer's to ask.
func requireClosingStatement(q recordsql.Querier, seatID string, c *recordpb.Closing) error {
	pair := c.Location != nil || c.FixNew != nil
	road := pair || c.GetRetire()
	author, _, err := authorOfGap(q, c.GetGapId())
	if err != nil {
		return err
	}
	if seatID == author {
		if road || c.GetNoText() {
			return feov.Errorf(feov.Validation, "blue's closing states its word, never text: --quote, --new, --retire and --no-text are red's closing's road. Nothing was recorded")
		}
		if c.ClosureClass == nil {
			return nil
		}
		docketing, _, err := standingOnGap(q, c.GetGapId())
		if err != nil {
			return err
		}
		k, _, err := docketingAt(q, docketing)
		if err != nil {
			return err
		}
		if k.Cause != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
			return feov.Errorf(feov.Validation, "--as states blue's word on a gap docketed at impasse, and %s is not: on a proposal blue's statement is the proposal, and at grade movement the recorded grades. Nothing was recorded", c.GetGapId())
		}
		return nil
	}
	if c.ClosureClass != nil {
		return feov.Errorf(feov.Validation, "a closing word is blue's statement; red's closing states its text. Nothing was recorded")
	}
	if pair && (c.Location == nil || c.FixNew == nil) {
		return feov.Errorf(feov.Validation, "--quote and --new are one road: the span of the report as it stands, ending at the gap's marker, and what it becomes. Pass both. Nothing was recorded")
	}
	takes, about, err := closingTakesText(q, c.GetGapId())
	if err != nil {
		return err
	}
	n := 0
	for _, set := range []bool{pair, c.GetRetire(), c.GetNoText()} {
		if set {
			n++
		}
	}
	switch {
	case !takes && n > 0:
		return feov.Errorf(feov.Validation, "red's closing on %s states no text: the question before the bench there takes none (a regrade, grade movement, a closed gap, or a gap about an inquiry or a section). Nothing was recorded", c.GetGapId())
	case takes && n != 1:
		return feov.Errorf(feov.Validation, "red's closing on %s states its text: exactly one of a road — --quote with --new, or --retire on a gap about a citation or a computation — and --no-text, where every road red can state is refused. Nothing was recorded", c.GetGapId())
	case c.GetRetire() && about != recordpb.Word(recordpb.AboutKind_ABOUT_KIND_ANCHOR):
		return feov.Errorf(feov.Validation, "--retire retires the citation or computation a gap is about, and %s is about none. Nothing was recorded", c.GetGapId())
	}
	// A RETIRE ROAD IS HELD TO L6 AS ITS APPLICATION IS: the bench's retire for red, or blue's
	// `retire --accept`, of a citation a computation still in the report applies is refused, so a
	// statement of it would be a road no ruling and no concession can take.
	if c.GetRetire() {
		var kind, ref sql.NullString
		if _, err := queryRowAt(q, []any{&kind, &ref}, `SELECT "about_kind", "about_ref" FROM "mint" WHERE "gap_id" = ?`, c.GetGapId()); err != nil {
			return err
		}
		live, err := computationsApplying(q, ref.String, nil)
		if err != nil {
			return err
		}
		if len(live) > 0 {
			return feov.Errorf(feov.Conflict, "red's closing on %s states the retire of citation %s, and computation %s, still in the report, applies its method: that retire is refused while it does (L6), so neither the bench's answer for red nor blue's accept could take this road. "+
				"State a pair over the citation's sentence, or file --no-text where every road you can state is refused. Nothing was recorded", c.GetGapId(), ref.String, strings.Join(live, ", "))
		}
	}
	// THE ROAD IS ESTOPPED AS A RED MINT IS (L43): text red could not mint is never applied through the
	// bench.
	if pair {
		fam, err := familyAt(q)
		if err != nil {
			return err
		}
		return roadEstoppel(fam, fam.Events, c.GetGapId(), c.GetLocation())
	}
	return nil
}

// StandingOnGap is gap_state's standing proposal and standing docketing on gap g in run's record (""
// for none): the readers' twin a test holds the stream folds to.
func StandingOnGap(run Run, g string) (proposal, docketing string, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		docketing, proposal, err = standingOnGap(q, g)
		return err
	})
	return proposal, docketing, err
}

// RaiserRulingSince is raiserRulingSinceAt asked of the record the write lands on.
func (l *Landing) RaiserRulingSince(id string) (string, error) {
	return raiserRulingSinceAt(l.w.q, id)
}

// raiserRulingSinceAt is the latest docketing other than id ruled for its raiser after id's row on q's
// record — the earlier ruling in the docket order whose applied statement can have broken a statement
// stated on id — "" for none.
func raiserRulingSinceAt(q recordsql.Querier, id string) (string, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return "", err
	}
	ds := docketingsAt(evs, win.IDs(evs))
	var at int64
	for _, d := range ds {
		if d.ID == id {
			at = d.at
		}
	}
	broke, latest := "", int64(0)
	for _, d := range ds {
		if d.ID != id && d.ruled > at && d.ruled > latest && d.Docket().GetSidedWith() == recordpb.SidedWith_SIDED_WITH_RAISER {
			broke, latest = d.ID, d.ruled
		}
	}
	return broke, nil
}
