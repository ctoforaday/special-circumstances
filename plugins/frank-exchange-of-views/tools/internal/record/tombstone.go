package record

import (
	"database/sql"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// TOMBSTONES — a bench ruling is final for the ruled gap's text, and red is bound by the fix it
// prescribed (#1695 §III.B12; gblock 2026-11-14: "make them real markers with locators of real
// concrete types").
//
// A tombstone stands on a gap's own token: a RULING row on G, the gap a bench ruling but a remand
// decided, and an ESTOPPEL row on P, the gap whose prescription blue applied verbatim. The tool writes
// each row in the transaction of the act that owes it, under that act's identity and stamp, and an end
// row in the transaction of the act that ends it; no verb writes one. A tombstone stands while the
// latest row on its (gap, kind) places it (the tombstone_standing view).
//
// WHAT PLACES ONE, and what ends one, is tombstoneArms: one row per arm, so a test can remove each in
// turn (TestEveryTombstoneArmIsNeeded). An end is written only where a tombstone stands, one row per
// standing (gap, kind); a placement is written only on a token standing after the act — the ruling
// follows the edit that applies its winner's text, and the verbatim edit carries its gap's token in
// the text it applies — so no act writes a row it then ends. A cut ends nothing: an edit that leaves a
// token bare lists nothing in reopened (ReopenedAnchors), and a ruled or estopped gap whose sentence is
// cut keeps its tombstone for the run, as one whose token was bare or gone at the ruling does.

// tombstoneArm is one way an act owes tombstone rows: the kind it places or ends, and the gaps (with
// the docketing, for a ruling placement) the act names for it.
type tombstoneArm struct {
	name string
	kind recordpb.TombstoneKind
	ends bool
	owed func(q recordsql.Querier, seatID string, body proto.Message) ([]owedTombstone, error)
}

// owedTombstone is one gap an arm names, and the docketing a ruling placement names.
type owedTombstone struct{ gap, docketing string }

// tombstoneArms is every placing and ending arm, the ends first: within one act's rows the end rows
// are written before the placements, so a verbatim edit answering P1 whose reopened names P2 ends P2's
// tombstone and places P1's, and a placement where one stands writes a new row.
var tombstoneArms = []tombstoneArm{
	{name: "a reword of a ruled gap's sentence", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, ends: true, owed: rewordedGaps},
	{name: "a reword of the anchor a ruled gap is about", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, ends: true, owed: rewordedAnchorGaps},
	{name: "a retire of the anchor a ruled gap is about", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, ends: true, owed: retiredAnchorGaps},
	{name: "the next fate of the inquiry a ruled gap is about", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, ends: true, owed: inquiryFateGaps},
	{name: "a reword of an estopped gap's sentence", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL, ends: true, owed: rewordedUnplacedGaps},
	{name: "a bench ruling of a docketing", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING, owed: rulingPlacement},
	{name: "blue's verbatim application of a gap's prescription", kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL, owed: verbatimPlacement},
}

// writeTombstones writes the rows the act just inserted owes (tombstonesOwed), through writeIn under
// the act's identity, stamped as the act is, in its transaction. Each row meets validate's shape arm
// and nothing else, as a replay's write does. A Tombstone owes none.
func writeTombstones(tx *sql.Tx, act *Event, seatID string, body proto.Message, at func() string) error {
	if _, ok := body.(*recordpb.Tombstone); ok {
		return nil
	}
	rows, err := tombstonesOwed(tx, seatID, body)
	if err != nil {
		return fmt.Errorf("reading the tombstones this act owes: %w", err)
	}
	ts := act.GetTs()
	for _, r := range rows {
		ev := &Event{}
		typ, err := recordpb.SetBody(ev, r)
		if err != nil {
			return err
		}
		if _, err := writeIn(tx, ev, seatID, typ, r, func(*sql.Tx) error {
			return validate(writeView{q: tx}, seatID, typ, r)
		}, func() string { return ts }); err != nil {
			return fmt.Errorf("the tombstone row this act owes on gap %s could not be written: %w", r.GetGapId(), err)
		}
	}
	return nil
}

// tombstonesOwed is the rows body owes, in tombstoneArms' order: each ending arm's gaps whose tombstone
// of its kind stands, once per (gap, kind), then each placing arm's.
func tombstonesOwed(q recordsql.Querier, seatID string, body proto.Message) ([]*recordpb.Tombstone, error) {
	switch body.(type) {
	case *recordpb.MotionRule, *recordpb.BlueEdit, *recordpb.Retire, *recordpb.Inquiry:
	default:
		return nil, nil
	}
	var out []*recordpb.Tombstone
	type pair struct {
		gap  string
		kind recordpb.TombstoneKind
	}
	ended := map[pair]bool{}
	for _, a := range tombstoneArms {
		gaps, err := a.owed(q, seatID, body)
		if err != nil {
			return nil, err
		}
		for _, g := range gaps {
			row := &recordpb.Tombstone{GapId: proto.String(g.gap), Kind: a.kind.Enum()}
			if a.ends {
				k := pair{g.gap, a.kind}
				if ended[k] {
					continue
				}
				standing, err := tombstoneStands(q, a.kind, g.gap)
				if err != nil {
					return nil, err
				}
				if !standing {
					continue
				}
				ended[k] = true
				row.Ends = proto.Bool(true)
			} else if g.docketing != "" {
				row.DocketingId = proto.String(g.docketing)
			}
			out = append(out, row)
		}
	}
	return out, nil
}

// tombstoneStands is whether a tombstone of kind stands on gap.
func tombstoneStands(q recordsql.Querier, kind recordpb.TombstoneKind, gap string) (bool, error) {
	var one int
	return queryRowAt(q, []any{&one}, `SELECT 1 FROM "tombstone_standing" WHERE "kind" = ? AND "gap_id" = ?`, recordpb.Word(kind), gap)
}

// rewordedGaps is the gaps whose token sat in a sentence the edit reworded: its reopened.
func rewordedGaps(_ recordsql.Querier, _ string, body proto.Message) ([]owedTombstone, error) {
	ed, ok := body.(*recordpb.BlueEdit)
	if !ok {
		return nil, nil
	}
	return owedOn(ed.GetReopened()), nil
}

// rewordedUnplacedGaps is rewordedGaps but the gap the edit itself places an estoppel on: the placing
// edit's own reopened lists its gap, and is not a later reword.
func rewordedUnplacedGaps(q recordsql.Querier, seatID string, body proto.Message) ([]owedTombstone, error) {
	ed, ok := body.(*recordpb.BlueEdit)
	if !ok {
		return nil, nil
	}
	placed, err := verbatimPlacement(q, seatID, body)
	if err != nil {
		return nil, err
	}
	var out []owedTombstone
	for _, id := range ed.GetReopened() {
		if len(placed) > 0 && placed[0].gap == id {
			continue
		}
		out = append(out, owedTombstone{gap: id})
	}
	return out, nil
}

// rewordedAnchorGaps is the gaps about a citation or computation whose sentence the edit reworded.
func rewordedAnchorGaps(q recordsql.Querier, _ string, body proto.Message) ([]owedTombstone, error) {
	ed, ok := body.(*recordpb.BlueEdit)
	if !ok {
		return nil, nil
	}
	return gapsAbout(q, recordpb.AboutKind_ABOUT_KIND_ANCHOR, ed.GetReopened())
}

// retiredAnchorGaps is the gaps about a citation or computation the retire takes.
func retiredAnchorGaps(q recordsql.Querier, _ string, body proto.Message) ([]owedTombstone, error) {
	rt, ok := body.(*recordpb.Retire)
	if !ok {
		return nil, nil
	}
	return gapsAbout(q, recordpb.AboutKind_ABOUT_KIND_ANCHOR, rt.GetAnchors())
}

// inquiryFateGaps is the gaps about the inquiry whose fate the row records.
func inquiryFateGaps(q recordsql.Querier, _ string, body proto.Message) ([]owedTombstone, error) {
	in, ok := body.(*recordpb.Inquiry)
	if !ok || in.GetInquiryId() == "" {
		return nil, nil
	}
	return gapsAbout(q, recordpb.AboutKind_ABOUT_KIND_INQUIRY, []string{in.GetInquiryId()})
}

// rulingPlacement is the gap of the docketing a bench ruling but a remand decides, with that
// docketing. The assignee's accept names no docketing and is no bench ruling; a ruling by the gap's
// assignee places none.
func rulingPlacement(q recordsql.Querier, seatID string, body proto.Message) ([]owedTombstone, error) {
	r, ok := body.(*recordpb.MotionRule)
	if !ok || r.GetDocket() == nil || r.GetDocket().GetDocketingId() == "" ||
		r.GetDocket().GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
		return nil, nil
	}
	d := r.GetDocket().GetDocketingId()
	k, found, err := docketingAt(q, d)
	if err != nil || !found {
		return nil, err
	}
	var assignee sql.NullString
	if _, err := queryRowAt(q, []any{&assignee}, `SELECT "minted_by" FROM "gap" WHERE "gap_id" = ?`, k.GapID); err != nil {
		return nil, err
	}
	if assignee.String == seatID {
		return nil, nil
	}
	return []owedTombstone{{gap: k.GapID, docketing: d}}, nil
}

// verbatimPlacement is the gap whose prescription blue applied verbatim, where its mint carries one.
// The bench's application of red's closing text on a for-red (for-assignee) answer places none,
// whatever the gap's mint prescribed: the bar on what the bench decided is the ruling's own tombstone,
// for either side (gblock 2026-11-22: "Estoppel applies to the judges ruling for either side").
func verbatimPlacement(q recordsql.Querier, seatID string, body proto.Message) ([]owedTombstone, error) {
	ed, ok := body.(*recordpb.BlueEdit)
	if !ok || !ed.GetAppliedVerbatim() || ed.GetAnswers() == "" || seatclass.RoleOfSeat(seatID) == "bench" {
		return nil, nil
	}
	var fix sql.NullString
	found, err := queryRowAt(q, []any{&fix}, `SELECT "fix_new" FROM "mint" WHERE "gap_id" = ?`, ed.GetAnswers())
	if err != nil || !found || fix.String == "" {
		return nil, err
	}
	return []owedTombstone{{gap: ed.GetAnswers()}}, nil
}

// owedOn is ids as owed gaps.
func owedOn(ids []string) []owedTombstone {
	out := make([]owedTombstone, 0, len(ids))
	for _, id := range ids {
		out = append(out, owedTombstone{gap: id})
	}
	return out
}

// gapsAbout is the gaps about any of refs of kind, through mint_about.
func gapsAbout(q recordsql.Querier, kind recordpb.AboutKind, refs []string) ([]owedTombstone, error) {
	var out []owedTombstone
	for _, ref := range refs {
		rows, err := q.Query(`SELECT "gap_id" FROM "mint" WHERE "about_kind" = ? AND "about_ref" = ? ORDER BY "event_id"`, recordpb.Word(kind), ref)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var g string
			if err := rows.Scan(&g); err != nil {
				rows.Close()
				return nil, err
			}
			out = append(out, owedTombstone{gap: g})
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// requireTombstoneShape is validate's arm for a Tombstone: its ordinary shape and nothing more. A
// tombstone is written by the tool in the transaction of the act that owes it, and by no verb; the
// arm cannot tell the write path's row from another writer's, and does not try. Its gap is a gap;
// a ruling placement names a docketing on that gap — the one whose ruling placed it; an estoppel row
// and an end row name no docketing (the table's CHECK states the last too, so the refusal here is the
// one in words).
func requireTombstoneShape(q recordsql.Querier, t *recordpb.Tombstone) error {
	if err := requireGap(q, t.GetGapId(), "gap_id"); err != nil {
		return err
	}
	ruling := t.GetKind() == recordpb.TombstoneKind_TOMBSTONE_KIND_RULING && !t.GetEnds()
	switch {
	case ruling && t.GetDocketingId() == "":
		return fmt.Errorf("a ruling tombstone names the docketing whose ruling placed it")
	case !ruling && t.DocketingId != nil:
		return fmt.Errorf("an estoppel tombstone and an end row name no docketing")
	case !ruling:
		return nil
	}
	k, found, err := docketingAt(q, t.GetDocketingId())
	if err != nil {
		return err
	}
	if !found || k.GapID != t.GetGapId() {
		return fmt.Errorf("a ruling tombstone on %s names docketing %s, which is no docketing on that gap", t.GetGapId(), t.GetDocketingId())
	}
	return nil
}

// RulingTombstone is a ruling tombstone standing as a reader sees it: the gap it stands on, the
// docketing whose ruling placed it, and that ruling's event and word.
type RulingTombstone struct {
	Gap       string `json:"-"`
	Docketing string `json:"docketing"`
	Event     int64  `json:"event"`
	Word      string `json:"word"`
	// settled, reopensOn and final are the ruling's own answers, for the estopped row.
	settled, reopensOn string
	final              bool
}

// rulingTombstoneSQL selects every standing ruling row with its ruling's event and word columns; the
// caller adds a WHERE over t.
const rulingTombstoneSQL = `SELECT t."gap_id", t."docketing_id", md."event_id",
       md."disposition", md."sided_with", md."moved_grades",
       COALESCE(md."settled", ''), COALESCE(md."reopens_on", ''), COALESCE(md."final", 0)
  FROM "tombstone_standing" t
  JOIN "motion_rule_docket" md ON md."docketing_id" = t."docketing_id"
 WHERE t."kind" = 'ruling'`

// rulingTombstonesAt is the standing ruling rows the WHERE clause cond (over t) selects.
func rulingTombstonesAt(q recordsql.Querier, cond string, args ...any) ([]RulingTombstone, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(rulingTombstoneSQL+cond, args...)
	if err != nil {
		return nil, fmt.Errorf("reading the standing ruling tombstones: %w", err)
	}
	defer rows.Close()
	var out []RulingTombstone
	for rows.Next() {
		var t RulingTombstone
		var disp, side, moved sql.NullString
		if err := rows.Scan(&t.Gap, &t.Docketing, &t.Event, &disp, &side, &moved, &t.settled, &t.reopensOn, &t.final); err != nil {
			return nil, err
		}
		switch {
		case disp.Valid:
			t.Word = disp.String
		case side.Valid:
			t.Word = "for " + side.String
		case moved.Valid:
			t.Word = "grades " + moved.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// rulingProtecting is the standing ruling tombstone on gap.
func rulingProtecting(q recordsql.Querier, gap string) (RulingTombstone, bool, error) {
	own, err := rulingTombstonesAt(q, ` AND t."gap_id" = ?`, gap)
	if err != nil || len(own) == 0 {
		return RulingTombstone{}, false, err
	}
	return own[0], true, nil
}

// rulingsProtectingAll is every gap a standing ruling row stands on, keyed by gap.
func rulingsProtectingAll(q recordsql.Querier) (map[string]RulingTombstone, error) {
	rows, err := rulingTombstonesAt(q, ``)
	if err != nil {
		return nil, err
	}
	out := map[string]RulingTombstone{}
	for _, t := range rows {
		out[t.Gap] = t
	}
	return out, nil
}

// estoppelsStandingAt is every gap an estoppel tombstone stands on, with the row's event.
func estoppelsStandingAt(q recordsql.Querier) (map[string]int64, error) {
	out := map[string]int64{}
	if noRecord(q) {
		return out, nil
	}
	rows, err := q.Query(`SELECT "gap_id", "event_id" FROM "tombstone_standing" WHERE "kind" = 'estoppel'`)
	if err != nil {
		return nil, fmt.Errorf("reading the standing estoppel tombstones: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var g string
		var at int64
		if err := rows.Scan(&g, &at); err != nil {
			return nil, err
		}
		out[g] = at
	}
	return out, rows.Err()
}

// standingTombstonesIn is the tombstones standing at the end of evs, folded from its Tombstone rows:
// per (gap, kind), the latest row's, where it places one. The folds that read a stream prefix —
// estoppel against the report before an act, the lock — read the rows this way.
func standingTombstonesIn(evs []*Event) map[recordpb.TombstoneKind]map[string]bool {
	out := map[recordpb.TombstoneKind]map[string]bool{
		recordpb.TombstoneKind_TOMBSTONE_KIND_RULING:   {},
		recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL: {},
	}
	for _, e := range evs {
		t, ok := recordpb.BodyAs[*recordpb.Tombstone](e)
		if !ok {
			continue
		}
		if t.GetEnds() {
			delete(out[t.GetKind()], t.GetGapId())
		} else {
			out[t.GetKind()][t.GetGapId()] = true
		}
	}
	return out
}

// tombstoneEnds is what ends a ruling tombstone on a gap about kind (its about_kind word, "" for a
// quote gap) ref, in words a refusal and the work list's estopped row both say.
func tombstoneEnds(aboutKind, ref string) string {
	switch aboutKind {
	case "":
		return "a reword of its sentence, not a cut,"
	case recordpb.Word(recordpb.AboutKind_ABOUT_KIND_ANCHOR):
		return "a reword of " + EvidenceNoun(ref) + " " + ref + "'s sentence, or its retire,"
	case recordpb.Word(recordpb.AboutKind_ABOUT_KIND_INQUIRY):
		return "inquiry " + ref + "'s next fate"
	}
	return "nothing — it stands for the run —"
}

// capitalized is s with its first letter upper-cased, for a clause that opens a sentence.
func capitalized(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// EstoppelTombstone is a standing estoppel tombstone as a reader sees it: the row's event, written in
// the verbatim edit's transaction.
type EstoppelTombstone struct {
	Event int64 `json:"event"`
}

// tombstonesOn is gap's standing ruling and estoppel tombstones out of the read maps, nil where none.
func tombstonesOn(gap string, rulings map[string]RulingTombstone, estoppels map[string]int64) (*RulingTombstone, *EstoppelTombstone) {
	var r *RulingTombstone
	var e *EstoppelTombstone
	if t, ok := rulings[gap]; ok {
		r = &t
	}
	if at, ok := estoppels[gap]; ok {
		e = &EstoppelTombstone{Event: at}
	}
	return r, e
}
